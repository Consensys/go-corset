package corset

import (
	"fmt"

	"github.com/consensys/go-corset/pkg/sexp"
)

// ResolveCircuit resolves all symbols declared and used within a circuit,
// producing an environment which can subsequently be used to look up the
// relevant module or column identifiers.  This process can fail, of course, it
// a symbol (e.g. a column) is referred to which doesn't exist.  Likewise, if
// two modules or columns with identical names are declared in the same scope,
// etc.
func ResolveCircuit(srcmap *sexp.SourceMaps[Node], circuit *Circuit) (*ModuleScope, []SyntaxError) {
	// Construct top-level scope
	scope := NewModuleScope()
	// Define intrinsics
	for _, i := range INTRINSICS {
		scope.Define(&i)
	}
	// Register modules
	for _, m := range circuit.Modules {
		scope.Declare(m.Name, false)
	}
	// Construct resolver
	r := resolver{srcmap}
	// Initialise all columns
	errs1 := r.initialiseDeclarations(scope, circuit)
	// Finalise all columns / declarations
	errs2 := r.resolveDeclarations(scope, circuit)
	//
	if len(errs1)+len(errs2) > 0 {
		return nil, append(errs1, errs2...)
	}
	// Done
	return scope, nil
}

// Resolver packages up information necessary for resolving a circuit and
// checking that everything makes sense.
type resolver struct {
	// Source maps nodes in the circuit back to the spans in their original
	// source files.  This is needed when reporting syntax errors to generate
	// highlights of the relevant source line(s) in question.
	srcmap *sexp.SourceMaps[Node]
}

// Initialise all columns from their declaring constructs.
func (r *resolver) initialiseDeclarations(scope *ModuleScope, circuit *Circuit) []SyntaxError {
	// Input columns must be allocated before assignemts, since the hir.Schema
	// separates these out.
	errs := r.initialiseDeclarationsInModule(scope, circuit.Declarations)
	//
	for _, m := range circuit.Modules {
		// Process all declarations in the module
		merrs := r.initialiseDeclarationsInModule(scope.Enter(m.Name), m.Declarations)
		// Package up all errors
		errs = append(errs, merrs...)
	}
	//
	return errs
}

// Initialise all declarations in the given module scope.  That means allocating
// all bindings into the scope, whilst also ensuring that we never have two
// bindings for the same symbol, etc.  The key is that, at this stage, all
// bindings are potentially "non-finalised".  That means they may be missing key
// information which is yet to be determined (e.g. information about types, or
// contexts, etc).
func (r *resolver) initialiseDeclarationsInModule(scope *ModuleScope, decls []Declaration) []SyntaxError {
	errors := make([]SyntaxError, 0)
	// First, initialise any perspectives as submodules of the given scope.  Its
	// slightly frustrating that we have to do this separately, but the
	// non-lexical nature of perspectives forces our hand.
	for _, d := range decls {
		if def, ok := d.(*DefPerspective); ok {
			// Attempt to declare the perspective.  Note, we don't need to check
			// whether or not this succeeds here as, if it fails, this will be
			// caught below.
			scope.Declare(def.Name(), true)
		}
	}
	// Second, initialise all symbol (e.g. column) definitions.
	for _, d := range decls {
		for iter := d.Definitions(); iter.HasNext(); {
			def := iter.Next()
			// Attempt to declare symbol
			if !scope.Define(def) {
				msg := fmt.Sprintf("symbol %s already declared", def.Path())
				err := r.srcmap.SyntaxError(def, msg)
				errors = append(errors, *err)
			}
		}
	}
	// Third, intialise aliases
	if errors := r.initialiseAliasesInModule(scope, decls); len(errors) > 0 {
		return errors
	}
	//
	return errors
}

// Initialise all alias declarations in the given module scope.  This means
// declaring them within the module scope, whilst also supporting aliases of
// aliases, etc.  Since the order of aliases is unspecified, this means we have
// to iterate the alias declarations until a fixed point is reached.  Once that
// is done, if there are any aliases left unallocated then they indicate errors.
func (r *resolver) initialiseAliasesInModule(scope *ModuleScope, decls []Declaration) []SyntaxError {
	// Apply any aliases
	errors := make([]SyntaxError, 0)
	visited := make(map[string]Declaration)
	changed := true
	// Iterate aliases to fixed point (i.e. until no new aliases discovered)
	for changed {
		changed = false
		// Look for all aliases
		for _, d := range decls {
			if a, ok := d.(*DefAliases); ok {
				for i, alias := range a.aliases {
					symbol := a.symbols[i]
					if _, ok := visited[alias.name]; !ok {
						// Attempt to make the alias
						if change := scope.Alias(alias.name, symbol); change {
							visited[alias.name] = d
							changed = true
						}
					}
				}
			}
		}
	}
	// Check for any aliases which remain incomplete
	for _, decl := range decls {
		if a, ok := decl.(*DefAliases); ok {
			for i, alias := range a.aliases {
				symbol := a.symbols[i]
				// Check whether it already exists (or not)
				if d, ok := visited[alias.name]; ok && d == decl {
					continue
				} else if scope.Binding(alias.name, symbol.IsFunction()) != nil {
					err := r.srcmap.SyntaxError(alias, "symbol already exists")
					errors = append(errors, *err)
				} else {
					err := r.srcmap.SyntaxError(symbol, "unknown symbol")
					errors = append(errors, *err)
				}
			}
		}
	}
	// Done
	return errors
}

// Process all assignment column declarations.  These are more complex than for
// input columns, since there can be dependencies between them.  Thus, we cannot
// simply resolve them in one linear scan.
func (r *resolver) resolveDeclarations(scope *ModuleScope, circuit *Circuit) []SyntaxError {
	// Input columns must be allocated before assignemts, since the hir.Schema
	// separates these out.
	errs := r.finaliseDeclarationsInModule(scope, circuit.Declarations)
	//
	for _, m := range circuit.Modules {
		// Process all declarations in the module
		merrs := r.finaliseDeclarationsInModule(scope.Enter(m.Name), m.Declarations)
		// Package up all errors
		errs = append(errs, merrs...)
	}
	//
	return errs
}

// Finalise all declarations given in a module.  This requires an iterative
// process as we cannot finalise a declaration until all of its dependencies
// have been themselves finalised.  For example, a function which depends upon
// an interleaved column.  Until the interleaved column is finalised, its type
// won't be available and, hence, we cannot type the function.
func (r *resolver) finaliseDeclarationsInModule(scope *ModuleScope, decls []Declaration) []SyntaxError {
	// Changed indicates whether or not a new assignment was finalised during a
	// given iteration.  This is important to know since, if the assignment is
	// not complete and we didn't finalise any more assignments --- then, we've
	// reached a fixed point where the final assignment is incomplete (i.e.
	// there is some error somewhere).
	changed := true
	// Complete tells us whether or not the assignment is complete.  The
	// assignment is not complete if there it at least one declaration which is
	// not yet finalised.
	complete := false
	// For an incomplete assignment, this identifies the last declaration that
	// could not be finalised (i.e. as an example so we have at least one for
	// error reporting).
	var (
		incomplete Node = nil
		counter    uint = 32
	)
	//
	for changed && !complete && counter > 0 {
		errors := make([]SyntaxError, 0)
		changed = false
		complete = true
		//
		for _, d := range decls {
			// Check whether already finalised
			if !d.IsFinalised() {
				// No, so attempt to finalise
				ready, errs := r.declarationDependenciesAreFinalised(scope, d)
				// Check what we found
				if errs != nil {
					errors = append(errors, errs...)
				} else if ready {
					// Finalise declaration and handle errors
					errs := r.finaliseDeclaration(scope, d)
					errors = append(errors, errs...)
					// Record that a new assignment is available.
					changed = changed || len(errs) == 0
				} else {
					// Declaration not ready yet
					complete = false
					incomplete = d
				}
			}
		}
		// Sanity check for any errors caught during this iteration.
		if len(errors) > 0 {
			return errors
		}
		// Decrement counter
		counter--
	}
	// Check whether we actually finished the allocation.
	if counter == 0 {
		err := r.srcmap.SyntaxError(incomplete, "unable to complete resolution")
		return []SyntaxError{*err}
	} else if !complete {
		// No, we didn't.  So, something is wrong --- assume it must be a cyclic
		// definition for now.
		err := r.srcmap.SyntaxError(incomplete, "cyclic declaration")
		return []SyntaxError{*err}
	}
	// Done
	return nil
}

// Check that a given set of symbols have been finalised.  This is important,
// since we cannot finalise a declaration until all of its dependencies have
// themselves been finalised.
func (r *resolver) declarationDependenciesAreFinalised(scope *ModuleScope,
	decl Declaration) (bool, []SyntaxError) {
	var (
		errors    []SyntaxError
		finalised bool = true
	)
	// DefConstraints require special handling because they can be associated
	// with a perspective.  Perspectives are challenging here because they are
	// effectively non-lexical scopes, which is not a good fit for the module
	// tree structure used.
	if dc, ok := decl.(*DefConstraint); ok && dc.Perspective != nil {
		if dc.Perspective.IsResolved() || scope.Bind(dc.Perspective) {
			// Temporarily enter the perspective for the purposes of resolving
			// symbols within this declaration.
			scope = scope.Enter(dc.Perspective.Name())
		}
	}
	//
	for iter := decl.Dependencies(); iter.HasNext(); {
		symbol := iter.Next()
		// Attempt to resolve
		if !symbol.IsResolved() && !scope.Bind(symbol) {
			errors = append(errors, *r.srcmap.SyntaxError(symbol, "unknown symbol"))
			// not finalised yet
			finalised = false
		} else {
			// Check whether this declaration defines this symbol (because if it
			// does, we cannot expect it to be finalised yet :)
			selfdefinition := decl.Defines(symbol)
			// Check whether this symbol is already finalised.
			symbol_finalised := symbol.Binding().IsFinalised()
			// Final check
			if !selfdefinition && !symbol_finalised {
				// Ok, not ready for finalisation yet.
				finalised = false
			}
		}
	}
	//
	return finalised, errors
}

// Finalise a declaration.
func (r *resolver) finaliseDeclaration(scope *ModuleScope, decl Declaration) []SyntaxError {
	switch d := decl.(type) {
	case *DefConst:
		return r.finaliseDefConstInModule(scope, d)
	case *DefConstraint:
		return r.finaliseDefConstraintInModule(scope, d)
	case *DefFun:
		return r.finaliseDefFunInModule(scope, d)
	case *DefInRange:
		return r.finaliseDefInRangeInModule(scope, d)
	case *DefInterleaved:
		return r.finaliseDefInterleavedInModule(d)
	case *DefLookup:
		return r.finaliseDefLookupInModule(scope, d)
	case *DefPermutation:
		return r.finaliseDefPermutationInModule(d)
	case *DefPerspective:
		return r.finaliseDefPerspectiveInModule(scope, d)
	case *DefProperty:
		return r.finaliseDefPropertyInModule(scope, d)
	}
	//
	return nil
}

// Finalise one or more constant definitions within a given module.
// Specifically, we need to check that the constant values provided are indeed
// constants.
func (r *resolver) finaliseDefConstInModule(enclosing Scope, decl *DefConst) []SyntaxError {
	var errors []SyntaxError
	//
	for _, c := range decl.constants {
		scope := NewLocalScope(enclosing, false, true)
		// Resolve constant body
		errs := r.finaliseExpressionInModule(scope, c.binding.value)
		// Accumulate errors
		errors = append(errors, errs...)
		if len(errors) == 0 {
			// Check it is indeed constant!
			if constant := c.binding.value.AsConstant(); constant != nil {
				// Finalise constant binding.  Note, no need to register a syntax
				// error for the error case, because it would have already been
				// accounted for during resolution.
				c.binding.Finalise()
			}
		}
	}
	//
	return errors
}

// Finalise a vanishing constraint declaration after all symbols have been
// resolved. This involves: (a) checking the context is valid; (b) checking the
// expressions are well-typed.
func (r *resolver) finaliseDefConstraintInModule(enclosing *ModuleScope, decl *DefConstraint) []SyntaxError {
	var guard_errors []SyntaxError
	// Identifiery enclosing perspective (if applicable)
	if decl.Perspective != nil {
		// As before, we must temporarily enter the perspective here.
		perspective := decl.Perspective.Name()
		enclosing = enclosing.Enter(perspective)
	}
	// Construct scope in which to resolve constraint
	scope := NewLocalScope(enclosing, false, false)
	// Resolve guard
	if decl.Guard != nil {
		guard_errors = r.finaliseExpressionInModule(scope, decl.Guard)
	}
	// Resolve constraint body
	constraint_errors := r.finaliseExpressionInModule(scope, decl.Constraint)
	//
	if len(guard_errors) == 0 && len(constraint_errors) == 0 {
		// Finalise declaration.
		decl.Finalise()
	}
	// Done
	return append(guard_errors, constraint_errors...)
}

// Finalise an interleaving assignment.  Since the assignment would already been
// initialised, all we need to do is determine the appropriate type and length
// multiplier for the interleaved column.  This can still result in an error,
// for example, if the multipliers between interleaved columns are incompatible,
// etc.
func (r *resolver) finaliseDefInterleavedInModule(decl *DefInterleaved) []SyntaxError {
	var (
		// Length multiplier being determined
		length_multiplier uint
		// Column type being determined
		datatype Type
		// Errors discovered
		errors []SyntaxError
	)
	// Determine type and length multiplier
	for i, source := range decl.Sources {
		// Lookup binding of column being interleaved.
		binding := source.Binding().(*ColumnBinding)
		//
		if i == 0 {
			length_multiplier = binding.multiplier
			datatype = binding.dataType
		} else if binding.multiplier != length_multiplier {
			// Columns to be interleaved must have the same length multiplier.
			err := r.srcmap.SyntaxError(decl, fmt.Sprintf("source column %s has incompatible length multiplier", source.Path()))
			errors = append(errors, *err)
		}
		// Combine datatypes.
		datatype = GreatestLowerBound(datatype, binding.dataType)
	}
	// Finalise details only if no errors
	if len(errors) == 0 {
		// Determine actual length multiplier
		length_multiplier *= uint(len(decl.Sources))
		// Lookup existing declaration
		binding := decl.Target.Binding().(*ColumnBinding)
		// Finalise column binding
		binding.Finalise(length_multiplier, datatype)
	}
	// Done
	return errors
}

// Finalise a permutation assignment after all symbols have been resolved.  This
// requires checking the contexts of all columns is consistent.
func (r *resolver) finaliseDefPermutationInModule(decl *DefPermutation) []SyntaxError {
	var (
		multiplier uint = 0
		errors     []SyntaxError
	)
	// Finalise each column in turn
	for i := 0; i < len(decl.Sources); i++ {
		ith := decl.Sources[i]
		// Lookup source of column being permuted
		source := ith.Binding().(*ColumnBinding)
		// Sanity check length multiplier
		if i == 0 && source.dataType.AsUnderlying().AsUint() == nil {
			errors = append(errors, *r.srcmap.SyntaxError(ith, "fixed-width type required"))
		} else if i == 0 {
			multiplier = source.multiplier
		} else if multiplier != source.multiplier {
			// Problem
			errors = append(errors, *r.srcmap.SyntaxError(ith, "incompatible length multiplier"))
		}
		// All good, finalise target column
		target := decl.Targets[i].Binding().(*ColumnBinding)
		// Update with completed information
		target.multiplier = source.multiplier
		target.dataType = source.dataType
	}
	// Done
	return errors
}

// Resolve those variables appearing in the body of this property assertion.
func (r *resolver) finaliseDefPerspectiveInModule(enclosing Scope, decl *DefPerspective) []SyntaxError {
	scope := NewLocalScope(enclosing, false, false)
	// Resolve assertion
	errors := r.finaliseExpressionInModule(scope, decl.Selector)
	// Error check
	if len(errors) == 0 {
		decl.Finalise()
	}
	// Done
	return errors
}

// Finalise a range constraint declaration after all symbols have been
// resolved. This involves: (a) checking the context is valid; (b) checking the
// expressions are well-typed.
func (r *resolver) finaliseDefInRangeInModule(enclosing Scope, decl *DefInRange) []SyntaxError {
	var scope = NewLocalScope(enclosing, false, false)
	// Resolve property body
	errors := r.finaliseExpressionInModule(scope, decl.Expr)
	// Error check
	if len(errors) == 0 {
		decl.Finalise()
	}
	// Done
	return errors
}

// Finalise a function definition after all symbols have been resolved. This
// involves: (a) checking the context is valid for the body; (b) checking the
// body is well-typed; (c) for pure functions checking that no columns are
// accessed; (d) finally, resolving any parameters used within the body of this
// function.
func (r *resolver) finaliseDefFunInModule(enclosing Scope, decl *DefFun) []SyntaxError {
	var scope = NewLocalScope(enclosing, true, decl.IsPure())
	// Declare parameters in local scope
	for _, p := range decl.Parameters() {
		scope.DeclareLocal(p.Binding.name, &p.Binding)
	}
	// Resolve property body
	errors := r.finaliseExpressionInModule(scope, decl.Body())
	// Finalise declaration
	if len(errors) == 0 {
		decl.Finalise()
	}
	// Done
	return errors
}

// Resolve those variables appearing in the body of this lookup constraint.
func (r *resolver) finaliseDefLookupInModule(enclosing Scope, decl *DefLookup) []SyntaxError {
	var (
		sourceScope = NewLocalScope(enclosing, true, false)
		targetScope = NewLocalScope(enclosing, true, false)
	)
	// Resolve source expressions
	source_errors := r.finaliseExpressionsInModule(sourceScope, decl.Sources)
	// Resolve target expressions
	target_errors := r.finaliseExpressionsInModule(targetScope, decl.Targets)
	//
	return append(source_errors, target_errors...)
}

// Resolve those variables appearing in the body of this property assertion.
func (r *resolver) finaliseDefPropertyInModule(enclosing Scope, decl *DefProperty) []SyntaxError {
	scope := NewLocalScope(enclosing, false, false)
	// Resolve assertion
	return r.finaliseExpressionInModule(scope, decl.Assertion)
}

// Resolve a sequence of zero or more expressions within a given module.  This
// simply resolves each of the arguments in turn, collecting any errors arising.
func (r *resolver) finaliseExpressionsInModule(scope LocalScope, args []Expr) []SyntaxError {
	var errors []SyntaxError
	// Visit each argument
	for _, arg := range args {
		if arg != nil {
			errs := r.finaliseExpressionInModule(scope, arg)
			errors = append(errors, errs...)
		}
	}
	// Done
	return errors
}

// Resolve any variable accesses with this expression (which is declared in a
// given module).  The enclosing module is required to resolve unqualified
// variable accesses.  As above, the goal is ensure variable refers to something
// that was declared and, more specifically, what kind of access it is (e.g.
// column access, constant access, etc).
//
//nolint:staticcheck
func (r *resolver) finaliseExpressionInModule(scope LocalScope, expr Expr) []SyntaxError {
	switch v := expr.(type) {
	case *ArrayAccess:
		return r.finaliseArrayAccessInModule(scope, v)
	case *Add:
		return r.finaliseExpressionsInModule(scope, v.Args)
	case *Constant:
		return nil
	case *Debug:
		return r.finaliseExpressionInModule(scope, v.Arg)
	case *Exp:
		purescope := scope.NestedPureScope()
		arg_errs := r.finaliseExpressionInModule(scope, v.Arg)
		pow_errs := r.finaliseExpressionInModule(purescope, v.Pow)
		// combine errors
		return append(arg_errs, pow_errs...)
	case *For:
		nestedscope := scope.NestedScope()
		// Declare local variable
		nestedscope.DeclareLocal(v.Binding.name, &v.Binding)
		// Continue resolution
		return r.finaliseExpressionInModule(nestedscope, v.Body)
	case *If:
		return r.finaliseExpressionsInModule(scope, []Expr{v.Condition, v.TrueBranch, v.FalseBranch})
	case *Invoke:
		return r.finaliseInvokeInModule(scope, v)
	case *Let:
		return r.finaliseLetInModule(scope, v)
	case *List:
		return r.finaliseExpressionsInModule(scope, v.Args)
	case *Mul:
		return r.finaliseExpressionsInModule(scope, v.Args)
	case *Normalise:
		return r.finaliseExpressionInModule(scope, v.Arg)
	case *Reduce:
		return r.finaliseReduceInModule(scope, v)
	case *Shift:
		purescope := scope.NestedPureScope()
		arg_errs := r.finaliseExpressionInModule(scope, v.Arg)
		shf_errs := r.finaliseExpressionInModule(purescope, v.Shift)
		// combine errors
		return append(arg_errs, shf_errs...)
	case *Sub:
		return r.finaliseExpressionsInModule(scope, v.Args)
	case *VariableAccess:
		return r.finaliseVariableInModule(scope, v)
	default:
		return r.srcmap.SyntaxErrors(expr, "unknown expression encountered during resolution")
	}
}

// Resolve a specific array access contained within some expression which, in
// turn, is contained within some module.
func (r *resolver) finaliseArrayAccessInModule(scope LocalScope, expr *ArrayAccess) []SyntaxError {
	// Resolve argument
	errors := r.finaliseExpressionInModule(scope, expr.arg)
	//
	if !expr.IsResolved() && !scope.Bind(expr) {
		errors = append(errors, *r.srcmap.SyntaxError(expr, "unknown array column"))
	} else if _, ok := expr.Binding().(*ColumnBinding); !ok {
		errors = append(errors, *r.srcmap.SyntaxError(expr, "unknown array column"))
	}
	// All good
	return errors
}

// Resolve a specific invocation contained within some expression which, in
// turn, is contained within some module.  Note, qualified accesses are only
// permitted in a global context.
func (r *resolver) finaliseInvokeInModule(scope LocalScope, expr *Invoke) []SyntaxError {
	// Resolve arguments
	errors := r.finaliseExpressionsInModule(scope, expr.Args())
	// Lookup the corresponding function definition.
	if !expr.fn.IsResolved() && !scope.Bind(expr.fn) {
		return append(errors, *r.srcmap.SyntaxError(expr, "unknown function"))
	}
	// Following must be true if we get here.
	binding := expr.fn.binding.(FunctionBinding)
	// Check purity
	if scope.IsPure() && !binding.IsPure() {
		errors = append(errors, *r.srcmap.SyntaxError(expr, "not permitted in pure context"))
	}
	// Check provide correct number of arguments
	if !binding.HasArity(uint(len(expr.Args()))) {
		msg := fmt.Sprintf("incorrect number of arguments (found %d)", len(expr.Args()))
		errors = append(errors, *r.srcmap.SyntaxError(expr, msg))
	}
	//
	return errors
}

func (r *resolver) finaliseLetInModule(scope LocalScope, expr *Let) []SyntaxError {
	nestedscope := scope.NestedScope()
	// Declare assigned variable(s)
	for i, letvar := range expr.Vars {
		nestedscope.DeclareLocal(letvar.name, &expr.Vars[i])
	}
	// Finalise assigned expressions
	args_errs := r.finaliseExpressionsInModule(scope, expr.Args)
	// Finalise body
	body_errs := r.finaliseExpressionInModule(nestedscope, expr.Body)
	//
	return append(args_errs, body_errs...)
}

// Resolve a specific invocation contained within some expression which, in
// turn, is contained within some module.  Note, qualified accesses are only
// permitted in a global context.
func (r *resolver) finaliseReduceInModule(scope LocalScope, expr *Reduce) []SyntaxError {
	// Resolve arguments
	errors := r.finaliseExpressionInModule(scope, expr.arg)
	// Lookup the corresponding function definition.
	if !expr.fn.IsResolved() && !scope.Bind(expr.fn) {
		errors = append(errors, *r.srcmap.SyntaxError(expr, "unknown function"))
	} else {
		// Following must be true if we get here.
		binding := expr.fn.binding.(FunctionBinding)

		if scope.IsPure() && !binding.IsPure() {
			errors = append(errors, *r.srcmap.SyntaxError(expr, "not permitted in pure context"))
		}
	}
	// Done
	return errors
}

// Resolve a specific variable access contained within some expression which, in
// turn, is contained within some module.  Note, qualified accesses are only
// permitted in a global context.
func (r *resolver) finaliseVariableInModule(scope LocalScope, expr *VariableAccess) []SyntaxError {
	// Check whether this is a qualified access, or not.
	if !scope.IsGlobal() && expr.Path().IsAbsolute() {
		return r.srcmap.SyntaxErrors(expr, "qualified access not permitted here")
	}
	// Symbol should be resolved at this point, but we'd better sanity check this.
	if !expr.IsResolved() && !scope.Bind(expr) {
		// Unable to resolve variable
		return r.srcmap.SyntaxErrors(expr, "unresolved symbol")
	}
	// Check what we've got.
	if binding, ok := expr.Binding().(*ColumnBinding); ok {
		// For column bindings, we still need to sanity check the context is
		// compatible.
		if !scope.FixContext(binding.Context()) {
			return r.srcmap.SyntaxErrors(expr, "conflicting context")
		} else if scope.IsPure() {
			return r.srcmap.SyntaxErrors(expr, "not permitted in pure context")
		}
		//
		return nil
	} else if _, ok := expr.Binding().(*ConstantBinding); ok {
		// Constant
		return nil
	} else if _, ok := expr.Binding().(*LocalVariableBinding); ok {
		// Parameter, for or let variable
		return nil
	} else if _, ok := expr.Binding().(FunctionBinding); ok {
		// Function doesn't makes sense here.
		return r.srcmap.SyntaxErrors(expr, "refers to a function")
	}
	// Should be unreachable.
	return r.srcmap.SyntaxErrors(expr, "unknown symbol kind")
}
