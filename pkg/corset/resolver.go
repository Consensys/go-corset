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
func ResolveCircuit(srcmap *sexp.SourceMaps[Node], circuit *Circuit) (*GlobalScope, []SyntaxError) {
	// Construct top-level scope
	scope := NewGlobalScope()
	// Register the root module (which should always exist)
	scope.DeclareModule("")
	// Register other modules
	for _, m := range circuit.Modules {
		scope.DeclareModule(m.Name)
	}
	// Construct resolver
	r := resolver{srcmap}
	// Allocate declared input columns
	errs := r.resolveDeclarations(scope, circuit)
	//
	if len(errs) > 0 {
		return nil, errs
	}
	// Done
	return scope, errs
}

// Resolver packages up information necessary for resolving a circuit and
// checking that everything makes sense.
type resolver struct {
	// Source maps nodes in the circuit back to the spans in their original
	// source files.  This is needed when reporting syntax errors to generate
	// highlights of the relevant source line(s) in question.
	srcmap *sexp.SourceMaps[Node]
}

// Process all assignment column declarations.  These are more complex than for
// input columns, since there can be dependencies between them.  Thus, we cannot
// simply resolve them in one linear scan.
func (r *resolver) resolveDeclarations(scope *GlobalScope, circuit *Circuit) []SyntaxError {
	// Input columns must be allocated before assignemts, since the hir.Schema
	// separates these out.
	errs := r.resolveDeclarationsInModule(scope.Module(""), circuit.Declarations)
	//
	for _, m := range circuit.Modules {
		// Process all declarations in the module
		merrs := r.resolveDeclarationsInModule(scope.Module(m.Name), m.Declarations)
		// Package up all errors
		errs = append(errs, merrs...)
	}
	//
	return errs
}

// Resolve all columns declared in a given module.  This is tricky because
// assignments can depend on the declaration of other columns.  Hence, we have
// to process all columns before we can sure that they are all declared
// correctly.
func (r *resolver) resolveDeclarationsInModule(scope *ModuleScope, decls []Declaration) []SyntaxError {
	// Columns & Assignments
	if errors := r.initialiseDeclarationsInModule(scope, decls); len(errors) > 0 {
		return errors
	}
	// Aliases
	if errors := r.initialiseAliasesInModule(scope, decls); len(errors) > 0 {
		return errors
	}
	// Finalise everything
	return r.finaliseDeclarationsInModule(scope, decls)
}

// Initialise all declarations in the given module scope.  That means allocating
// all bindings into the scope, whilst also ensuring that we never have two
// bindings for the same symbol, etc.  The key is that, at this stage, all
// bindings are potentially "non-finalised".  That means they may be missing key
// information which is yet to be determined (e.g. information about types, or
// contexts, etc).
func (r *resolver) initialiseDeclarationsInModule(scope *ModuleScope, decls []Declaration) []SyntaxError {
	errors := make([]SyntaxError, 0)
	// Initialise all columns
	for _, d := range decls {
		for iter := d.Definitions(); iter.HasNext(); {
			def := iter.Next()
			// Attempt to declare symbol
			if !scope.Declare(def) {
				msg := fmt.Sprintf("symbol %s already declared", def.Name())
				err := r.srcmap.SyntaxError(def, msg)
				errors = append(errors, *err)
			}
		}
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
					err := r.srcmap.SyntaxError(symbol, "unknown symbol encountered during resolution")
					errors = append(errors, *err)
				}
			}
		}
	}
	// Done
	return errors
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
		counter    uint = 4
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
	//
	for iter := decl.Dependencies(); iter.HasNext(); {
		symbol := iter.Next()
		// Attempt to resolve
		if !symbol.IsResolved() && !scope.Bind(symbol) {
			errors = append(errors, *r.srcmap.SyntaxError(symbol, "unknown symbol encountered during resolution"))
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
	if d, ok := decl.(*DefConst); ok {
		return r.finaliseDefConstInModule(scope, d)
	} else if d, ok := decl.(*DefConstraint); ok {
		return r.finaliseDefConstraintInModule(scope, d)
	} else if d, ok := decl.(*DefFun); ok {
		return r.finaliseDefFunInModule(scope, d)
	} else if d, ok := decl.(*DefInRange); ok {
		return r.finaliseDefInRangeInModule(scope, d)
	} else if d, ok := decl.(*DefInterleaved); ok {
		return r.finaliseDefInterleavedInModule(d)
	} else if d, ok := decl.(*DefLookup); ok {
		return r.finaliseDefLookupInModule(scope, d)
	} else if d, ok := decl.(*DefPermutation); ok {
		return r.finaliseDefPermutationInModule(d)
	} else if d, ok := decl.(*DefProperty); ok {
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
		datatype, errs := r.finaliseExpressionInModule(scope, c.binding.value)
		// Accumulate errors
		errors = append(errors, errs...)
		// Check it is indeed constant!
		if constant := c.binding.value.AsConstant(); constant != nil {
			// Finalise constant binding.  Note, no need to register a syntax
			// error for the error case, because it would have already been
			// accounted for during resolution.
			c.binding.Finalise(datatype)
		}
	}
	//
	return errors
}

// Finalise a vanishing constraint declaration after all symbols have been
// resolved. This involves: (a) checking the context is valid; (b) checking the
// expressions are well-typed.
func (r *resolver) finaliseDefConstraintInModule(enclosing Scope, decl *DefConstraint) []SyntaxError {
	var (
		guard_errors []SyntaxError
		guard_t      Type
		scope        = NewLocalScope(enclosing, false, false)
	)
	// Resolve guard
	if decl.Guard != nil {
		guard_t, guard_errors = r.finaliseExpressionInModule(scope, decl.Guard)
		//
		if guard_t != nil && guard_t.HasLoobeanSemantics() {
			err := r.srcmap.SyntaxError(decl.Guard, "unexpected loobean guard")
			guard_errors = append(guard_errors, *err)
		}
	}
	// Resolve constraint body
	constraint_t, errors := r.finaliseExpressionInModule(scope, decl.Constraint)
	//
	if constraint_t != nil && !constraint_t.HasLoobeanSemantics() {
		msg := fmt.Sprintf("expected loobean constraint (found %s)", constraint_t.String())
		err := r.srcmap.SyntaxError(decl.Constraint, msg)
		errors = append(errors, *err)
	} else if len(errors) == 0 {
		// Finalise declaration.
		decl.Finalise()
	}
	// Done
	return append(guard_errors, errors...)
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
			err := r.srcmap.SyntaxError(decl, fmt.Sprintf("source column %s has incompatible length multiplier", source.Name()))
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

// Finalise a range constraint declaration after all symbols have been
// resolved. This involves: (a) checking the context is valid; (b) checking the
// expressions are well-typed.
func (r *resolver) finaliseDefInRangeInModule(enclosing Scope, decl *DefInRange) []SyntaxError {
	var (
		scope = NewLocalScope(enclosing, false, false)
	)
	// Resolve property body
	_, errors := r.finaliseExpressionInModule(scope, decl.Expr)
	// Done
	return errors
}

// Finalise a function definition after all symbols have been resolved. This
// involves: (a) checking the context is valid for the body; (b) checking the
// body is well-typed; (c) for pure functions checking that no columns are
// accessed; (d) finally, resolving any parameters used within the body of this
// function.
func (r *resolver) finaliseDefFunInModule(enclosing Scope, decl *DefFun) []SyntaxError {
	var (
		scope = NewLocalScope(enclosing, false, decl.IsPure())
	)
	// Declare parameters in local scope
	for _, p := range decl.Parameters() {
		scope.DeclareLocal(p.Binding.name, &p.Binding)
	}
	// Resolve property body
	datatype, errors := r.finaliseExpressionInModule(scope, decl.Body())
	// Finalise declaration
	if len(errors) == 0 {
		decl.binding.Finalise(datatype)
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
	_, source_errors := r.finaliseExpressionsInModule(sourceScope, decl.Sources)
	// Resolve target expressions
	_, target_errors := r.finaliseExpressionsInModule(targetScope, decl.Targets)
	//
	return append(source_errors, target_errors...)
}

// Resolve those variables appearing in the body of this property assertion.
func (r *resolver) finaliseDefPropertyInModule(enclosing Scope, decl *DefProperty) []SyntaxError {
	var (
		scope = NewLocalScope(enclosing, false, false)
	)
	// Resolve assertion
	_, errors := r.finaliseExpressionInModule(scope, decl.Assertion)
	// Done
	return errors
}

// Resolve a sequence of zero or more expressions within a given module.  This
// simply resolves each of the arguments in turn, collecting any errors arising.
func (r *resolver) finaliseExpressionsInModule(scope LocalScope, args []Expr) ([]Type, []SyntaxError) {
	var (
		errs   []SyntaxError
		errors []SyntaxError
		types  []Type = make([]Type, len(args))
	)
	// Visit each argument
	for i, arg := range args {
		if arg != nil {
			types[i], errs = r.finaliseExpressionInModule(scope, arg)
			errors = append(errors, errs...)
		}
	}
	// Done
	return types, errors
}

// Resolve any variable accesses with this expression (which is declared in a
// given module).  The enclosing module is required to resolve unqualified
// variable accesses.  As above, the goal is ensure variable refers to something
// that was declared and, more specifically, what kind of access it is (e.g.
// column access, constant access, etc).
//
//nolint:staticcheck
func (r *resolver) finaliseExpressionInModule(scope LocalScope, expr Expr) (Type, []SyntaxError) {
	if v, ok := expr.(*ArrayAccess); ok {
		return r.finaliseArrayAccessInModule(scope, v)
	} else if v, ok := expr.(*Add); ok {
		types, errs := r.finaliseExpressionsInModule(scope, v.Args)
		return LeastUpperBoundAll(types), errs
	} else if v, ok := expr.(*Constant); ok {
		nbits := v.Val.BitLen()
		return NewUintType(uint(nbits)), nil
	} else if v, ok := expr.(*Debug); ok {
		return r.finaliseExpressionInModule(scope, v.Arg)
	} else if v, ok := expr.(*Exp); ok {
		purescope := scope.NestedPureScope()
		arg_types, arg_errs := r.finaliseExpressionInModule(scope, v.Arg)
		_, pow_errs := r.finaliseExpressionInModule(purescope, v.Pow)
		// combine errors
		return arg_types, append(arg_errs, pow_errs...)
	} else if v, ok := expr.(*For); ok {
		nestedscope := scope.NestedScope()
		// Declare local variable
		nestedscope.DeclareLocal(v.Binding.name, &v.Binding)
		// Continue resolution
		return r.finaliseExpressionInModule(nestedscope, v.Body)
	} else if v, ok := expr.(*If); ok {
		return r.finaliseIfInModule(scope, v)
	} else if v, ok := expr.(*Invoke); ok {
		return r.finaliseInvokeInModule(scope, v)
	} else if v, ok := expr.(*List); ok {
		types, errs := r.finaliseExpressionsInModule(scope, v.Args)
		return LeastUpperBoundAll(types), errs
	} else if v, ok := expr.(*Mul); ok {
		types, errs := r.finaliseExpressionsInModule(scope, v.Args)
		return GreatestLowerBoundAll(types), errs
	} else if v, ok := expr.(*Normalise); ok {
		_, errs := r.finaliseExpressionInModule(scope, v.Arg)
		// Normalise guaranteed to return either 0 or 1.
		return NewUintType(1), errs
	} else if v, ok := expr.(*Reduce); ok {
		return r.finaliseReduceInModule(scope, v)
	} else if v, ok := expr.(*Shift); ok {
		purescope := scope.NestedPureScope()
		arg_types, arg_errs := r.finaliseExpressionInModule(scope, v.Arg)
		_, shf_errs := r.finaliseExpressionInModule(purescope, v.Shift)
		// combine errors
		return arg_types, append(arg_errs, shf_errs...)
	} else if v, ok := expr.(*Sub); ok {
		types, errs := r.finaliseExpressionsInModule(scope, v.Args)
		return LeastUpperBoundAll(types), errs
	} else if v, ok := expr.(*VariableAccess); ok {
		return r.finaliseVariableInModule(scope, v)
	} else {
		return nil, r.srcmap.SyntaxErrors(expr, "unknown expression encountered during resolution")
	}
}

// Resolve a specific array access contained within some expression which, in
// turn, is contained within some module.
func (r *resolver) finaliseArrayAccessInModule(scope LocalScope, expr *ArrayAccess) (Type, []SyntaxError) {
	// Resolve argument
	if _, errors := r.finaliseExpressionInModule(scope, expr.arg); errors != nil {
		return nil, errors
	}
	//
	if !expr.IsResolved() && !scope.Bind(expr) {
		return nil, r.srcmap.SyntaxErrors(expr, "unknown array column")
	} else if binding, ok := expr.Binding().(*ColumnBinding); !ok {
		return nil, r.srcmap.SyntaxErrors(expr, "unknown array column")
	} else if arr_t, ok := binding.dataType.(*ArrayType); !ok {
		return nil, r.srcmap.SyntaxErrors(expr, "expected array column")
	} else {
		// All good
		return arr_t.element, nil
	}
}

// Resolve an if condition contained within some expression which, in turn, is
// contained within some module.  An important step occurrs here where, based on
// the semantics of the condition, this is inferred as an "if-zero" or an
// "if-notzero".
func (r *resolver) finaliseIfInModule(scope LocalScope, expr *If) (Type, []SyntaxError) {
	types, errs := r.finaliseExpressionsInModule(scope, []Expr{expr.Condition, expr.TrueBranch, expr.FalseBranch})
	// Sanity check
	if len(errs) != 0 {
		return nil, errs
	}
	// Check & Resolve Condition
	if types[0].HasLoobeanSemantics() {
		// if-zero
		expr.FixSemantics(true)
	} else if types[0].HasBooleanSemantics() {
		// if-notzero
		expr.FixSemantics(false)
	} else {
		return nil, r.srcmap.SyntaxErrors(expr.Condition, "invalid condition (neither loobean nor boolean)")
	}
	// Join result types
	return GreatestLowerBoundAll(types[1:]), errs
}

// Resolve a specific invocation contained within some expression which, in
// turn, is contained within some module.  Note, qualified accesses are only
// permitted in a global context.
func (r *resolver) finaliseInvokeInModule(scope LocalScope, expr *Invoke) (Type, []SyntaxError) {
	var (
		errors   []SyntaxError
		argTypes []Type
	)
	// Resolve arguments
	argTypes, errors = r.finaliseExpressionsInModule(scope, expr.Args())
	// Lookup the corresponding function definition.
	if !expr.fn.IsResolved() && !scope.Bind(expr.fn) {
		return nil, append(errors, *r.srcmap.SyntaxError(expr, "unknown function"))
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
	// Select best overloaded signature
	if signature := binding.Select(argTypes); signature != nil {
		// Check arguments are accepted, based on their type.
		for i := 0; i < len(argTypes); i++ {
			expected := signature.Parameter(uint(i))
			actual := argTypes[i]
			// subtype check
			if actual != nil && !actual.SubtypeOf(expected) {
				msg := fmt.Sprintf("expected type %s (found %s)", expected, actual)
				errors = append(errors, *r.srcmap.SyntaxError(expr.args[i], msg))
			}
		}
		//
		expr.Finalise(signature)
		//
		if len(errors) != 0 {
			return nil, errors
		} else if signature.Return() != nil {
			// no need, it was provided
			return signature.Return(), nil
		}
		// TODO: this is potentially expensive, and it would likely be good if we
		// could avoid it.  Realistically, this is just about determining the right
		// type information.  Potentially, we could adjust the local scope to
		// provide the required type information.  Or we could have a separate pass
		// which just determines the type.
		body := signature.Apply(expr.Args(), nil)
		// Dig out the type
		return r.finaliseExpressionInModule(scope, body)
	}
	// ambiguous invocation
	return nil, append(errors, *r.srcmap.SyntaxError(expr, "ambiguous invocation"))
}

// Resolve a specific invocation contained within some expression which, in
// turn, is contained within some module.  Note, qualified accesses are only
// permitted in a global context.
func (r *resolver) finaliseReduceInModule(scope LocalScope, expr *Reduce) (Type, []SyntaxError) {
	var signature *FunctionSignature
	// Resolve arguments
	body_t, errors := r.finaliseExpressionInModule(scope, expr.arg)
	// Lookup the corresponding function definition.
	if !expr.fn.IsResolved() && !scope.Bind(expr.fn) {
		errors = append(errors, *r.srcmap.SyntaxError(expr, "unknown function"))
	} else {
		// Following must be true if we get here.
		binding := expr.fn.binding.(FunctionBinding)

		if scope.IsPure() && !binding.IsPure() {
			errors = append(errors, *r.srcmap.SyntaxError(expr, "not permitted in pure context"))
		} else if !binding.HasArity(2) {
			msg := "incorrect number of arguments (expected 2)"
			errors = append(errors, *r.srcmap.SyntaxError(expr, msg))
		} else if signature = binding.Select([]Type{body_t, body_t}); signature != nil {
			// Check left parameter type
			if !body_t.SubtypeOf(signature.Parameter(0)) {
				msg := fmt.Sprintf("expected type %s (found %s)", signature.Parameter(0), body_t)
				errors = append(errors, *r.srcmap.SyntaxError(expr.arg, msg))
			}
			// Check right parameter type
			if !body_t.SubtypeOf(signature.Parameter(1)) {
				msg := fmt.Sprintf("expected type %s (found %s)", signature.Parameter(1), body_t)
				errors = append(errors, *r.srcmap.SyntaxError(expr.arg, msg))
			}
		} else {
			msg := "ambiguous reduction"
			errors = append(errors, *r.srcmap.SyntaxError(expr, msg))
		}
	}
	// Error check
	if len(errors) > 0 {
		return nil, errors
	}
	// Lock in signature
	expr.Finalise(signature)
	//
	return body_t, nil
}

// Resolve a specific variable access contained within some expression which, in
// turn, is contained within some module.  Note, qualified accesses are only
// permitted in a global context.
func (r *resolver) finaliseVariableInModule(scope LocalScope,
	expr *VariableAccess) (Type, []SyntaxError) {
	// Check whether this is a qualified access, or not.
	if !scope.IsGlobal() && expr.IsQualified() {
		return nil, r.srcmap.SyntaxErrors(expr, "qualified access not permitted here")
	} else if expr.IsQualified() && !scope.HasModule(expr.Module()) {
		return nil, r.srcmap.SyntaxErrors(expr, fmt.Sprintf("unknown module %s", expr.Module()))
	}
	// Symbol should be resolved at this point, but we'd better sanity check this.
	if !expr.IsResolved() && !scope.Bind(expr) {
		// Unable to resolve variable
		return nil, r.srcmap.SyntaxErrors(expr, "unresolved symbol")
	}
	//
	if binding, ok := expr.Binding().(*ColumnBinding); ok {
		// For column bindings, we still need to sanity check the context is
		// compatible.
		if !scope.FixContext(binding.Context()) {
			return nil, r.srcmap.SyntaxErrors(expr, "conflicting context")
		} else if scope.IsPure() {
			return nil, r.srcmap.SyntaxErrors(expr, "not permitted in pure context")
		}
		// Use column's datatype
		return binding.dataType, nil
	} else if binding, ok := expr.Binding().(*ConstantBinding); ok {
		// Constant
		return binding.datatype, nil
	} else if binding, ok := expr.Binding().(*LocalVariableBinding); ok {
		// Parameter or other local variable
		return binding.datatype, nil
	} else if _, ok := expr.Binding().(FunctionBinding); ok {
		// Function doesn't makes sense here.
		return nil, r.srcmap.SyntaxErrors(expr, "refers to a function")
	}
	// Should be unreachable.
	return nil, r.srcmap.SyntaxErrors(expr, "unknown symbol kind")
}
