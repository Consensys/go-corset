package corset

import (
	"fmt"

	"github.com/consensys/go-corset/pkg/schema"
	"github.com/consensys/go-corset/pkg/sexp"
	tr "github.com/consensys/go-corset/pkg/trace"
)

// ResolveCircuit resolves all symbols declared and used within a circuit,
// producing an environment which can subsequently be used to look up the
// relevant module or column identifiers.  This process can fail, of course, it
// a symbol (e.g. a column) is referred to which doesn't exist.  Likewise, if
// two modules or columns with identical names are declared in the same scope,
// etc.
func ResolveCircuit(srcmap *sexp.SourceMaps[Node], circuit *Circuit) (*Environment, []SyntaxError) {
	r := resolver{EmptyEnvironment(), srcmap}
	// Allocate declared modules
	r.resolveModules(circuit)
	// Allocate declared input columns
	errs := r.resolveColumns(circuit)
	// Check expressions
	errs = append(errs, r.resolveConstraints(circuit)...)
	// Done
	return r.env, errs
}

// Resolver packages up information necessary for resolving a circuit and
// checking that everything makes sense.
type resolver struct {
	// Environment determines module and column indices, as needed for
	// translating the various constructs found in a circuit.
	env *Environment
	// Source maps nodes in the circuit back to the spans in their original
	// source files.  This is needed when reporting syntax errors to generate
	// highlights of the relevant source line(s) in question.
	srcmap *sexp.SourceMaps[Node]
}

// Process all module declarations, and allocating them into the environment.
// If any duplicates are found, one or more errors will be reported.  Note: it
// is important that this traverses the modules in an identical order to the
// translator.  This is to ensure that the relevant module identifiers line up.
func (r *resolver) resolveModules(circuit *Circuit) {
	// Register the root module (which should always exist)
	r.env.RegisterModule("")
	//
	for _, m := range circuit.Modules {
		r.env.RegisterModule(m.Name)
	}
}

// Process all input column or column assignment declarations.
func (r *resolver) resolveColumns(circuit *Circuit) []SyntaxError {
	// Allocate input columns first.  These must all be done before any
	// assignments are allocated, since the hir.Schema separates these out.
	ierrs := r.resolveInputColumns(circuit)
	// Now we can resolve any assignments.
	aerrs := r.resolveAssignments(circuit)
	//
	return append(ierrs, aerrs...)
}

// Process all input column declarations.
func (r *resolver) resolveInputColumns(circuit *Circuit) []SyntaxError {
	// Input columns must be allocated before assignemts, since the hir.Schema
	// separates these out.
	errs := r.resolveInputColumnsInModule("", circuit.Declarations)
	//
	for _, m := range circuit.Modules {
		// Process all declarations in the module
		merrs := r.resolveInputColumnsInModule(m.Name, m.Declarations)
		// Package up all errors
		errs = append(errs, merrs...)
	}
	//
	return errs
}

// Resolve all input columns in a given module.
func (r *resolver) resolveInputColumnsInModule(module string, decls []Declaration) []SyntaxError {
	errors := make([]SyntaxError, 0)
	mid := r.env.Module(module)
	//
	for _, d := range decls {
		if dcols, ok := d.(*DefColumns); ok {
			// Found one.
			for _, col := range dcols.Columns {
				// Check whether column already exists
				if _, ok := r.env.LookupColumn(mid, col.Name); ok {
					err := r.srcmap.SyntaxError(col, fmt.Sprintf("column %s already declared in module %s", col.Name, module))
					errors = append(errors, *err)
				} else {
					context := tr.NewContext(mid, col.LengthMultiplier)
					r.env.RegisterColumn(context, col.Name, col.DataType)
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
func (r *resolver) resolveAssignments(circuit *Circuit) []SyntaxError {
	// Input columns must be allocated before assignemts, since the hir.Schema
	// separates these out.
	errs := r.resolveAssignmentsInModule("", circuit.Declarations)
	//
	for _, m := range circuit.Modules {
		// Process all declarations in the module
		merrs := r.resolveAssignmentsInModule(m.Name, m.Declarations)
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
func (r *resolver) resolveAssignmentsInModule(module string, decls []Declaration) []SyntaxError {
	if errors := r.initialiseAssignmentsInModule(module, decls); len(errors) > 0 {
		return errors
	}
	// Check assignments
	if errors := r.checkAssignmentsInModule(module, decls); len(errors) > 0 {
		return errors
	}
	// Iterate until all columns finalised
	return r.finaliseAssignmentsInModule(module, decls)
}

// Initialise the column allocation from the available declarations, whilst
// identifying any duplicate declarations.  Observe that, for some declarations,
// the initial assignment is incomplete because information about dependent
// columns may not be available.  So, the goal of the subsequent phase is to
// flesh out this missing information.
func (r *resolver) initialiseAssignmentsInModule(module string, decls []Declaration) []SyntaxError {
	errors := make([]SyntaxError, 0)
	mid := r.env.Module(module)
	//
	for _, d := range decls {
		if col, ok := d.(*DefInterleaved); ok {
			if _, ok := r.env.LookupColumn(mid, col.Target); ok {
				err := r.srcmap.SyntaxError(col, fmt.Sprintf("column %s already declared in module %s", col.Target, module))
				errors = append(errors, *err)
			} else {
				// Register incomplete (assignment) column.
				r.env.PreRegisterColumn(mid, col.Target)
			}
		} else if col, ok := d.(*DefPermutation); ok {
			for _, c := range col.Targets {
				if _, ok := r.env.LookupColumn(mid, c.Name); ok {
					err := r.srcmap.SyntaxError(col, fmt.Sprintf("column %s already declared in module %s", c.Name, module))
					errors = append(errors, *err)
				} else {
					// Register incomplete (assignment) column.
					r.env.PreRegisterColumn(mid, c.Name)
				}
			}
		}
	}
	// Done
	return errors
}

func (r *resolver) checkAssignmentsInModule(module string, decls []Declaration) []SyntaxError {
	errors := make([]SyntaxError, 0)
	mid := r.env.Module(module)
	//
	for _, d := range decls {
		if col, ok := d.(*DefInterleaved); ok {
			for _, c := range col.Sources {
				if !r.env.HasColumn(mid, c.Name) {
					errors = append(errors, *r.srcmap.SyntaxError(c, "unknown source column"))
				}
			}
		} else if col, ok := d.(*DefPermutation); ok {
			for _, c := range col.Sources {
				if !r.env.HasColumn(mid, c.Name) {
					errors = append(errors, *r.srcmap.SyntaxError(c, "unknown source column"))
				}
			}
		}
	}
	// Done
	return errors
}

// Iterate the column allocation to a fix point by iteratively fleshing out column information.
func (r *resolver) finaliseAssignmentsInModule(module string, decls []Declaration) []SyntaxError {
	mid := r.env.Module(module)
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
	var incomplete Node = nil
	//
	for changed && !complete {
		errors := make([]SyntaxError, 0)
		changed = false
		complete = true
		//
		for _, d := range decls {
			if col, ok := d.(*DefInterleaved); ok {
				// Check whether dependencies are resolved or not.
				if col.CanFinalise(mid, r.env) {
					// Finalise assignment and handle any errors
					errs := r.finaliseInterleavedAssignment(mid, col)
					errors = append(errors, errs...)
					// Record that a new assignment is available.
					changed = changed || len(errs) == 0
				} else {
					complete = false
					incomplete = d
				}
			} else if col, ok := d.(*DefPermutation); ok {
				// Check whether dependencies are resolved or not.
				if col.CanFinalise(mid, r.env) {
					// Finalise assignment and handle any errors
					errs := r.finalisePermutationAssignment(mid, col)
					errors = append(errors, errs...)
					// Record that a new assignment is available.
					changed = changed || len(errs) == 0
				} else {
					complete = false
					incomplete = d
				}
			}
		}
		// Sanity check for any errors caught during this iteration.
		if len(errors) > 0 {
			return errors
		}
	}
	// Check whether we actually finished the allocation.
	if !complete {
		// No, we didn't.  So, something is wrong --- assume it must be a cyclic
		// definition for now.
		err := r.srcmap.SyntaxError(incomplete, "cyclic declaration")
		return []SyntaxError{*err}
	}
	// Done
	return nil
}

// Finalise an interleaving assignment.  Since the assignment would already been
// initialised, all we need to do is determine the appropriate type and length
// multiplier for the interleaved column.  This can still result in an error,
// for example, if the multipliers between interleaved columns are incompatible,
// etc.
func (r *resolver) finaliseInterleavedAssignment(module uint, decl *DefInterleaved) []SyntaxError {
	var (
		// Length multiplier being determined
		length_multiplier uint
		// Column type being determined
		datatype schema.Type
		// Errors discovered
		errors []SyntaxError
	)
	// Determine type and length multiplier
	for i, source := range decl.Sources {
		// Lookup info of column being interleaved.
		info := r.env.Column(module, source.Name)
		if i == 0 {
			length_multiplier = info.multiplier
			datatype = info.datatype
		} else if info.multiplier != length_multiplier {
			// Columns to be interleaved must have the same length multiplier.
			err := r.srcmap.SyntaxError(decl, fmt.Sprintf("source column %s has incompatible length multiplier", source))
			errors = append(errors, *err)
		}
		// Combine datatypes.
		datatype = schema.Join(datatype, info.datatype)
	}
	// Finalise details only if no errors
	if len(errors) == 0 {
		// Determine actual length multiplier
		length_multiplier *= uint(len(decl.Sources))
		// Construct context for this column
		context := tr.NewContext(module, length_multiplier)
		// Finalise column registration
		r.env.FinaliseColumn(context, decl.Target, datatype)
	}
	// Done
	return errors
}

// Finalise a permutation assignment.  Since the assignment would already been
// initialised, this is actually quite easy to do.
func (r *resolver) finalisePermutationAssignment(module uint, decl *DefPermutation) []SyntaxError {
	var (
		multiplier uint = 0
		errors     []SyntaxError
	)
	// Finalise each column in turn
	for i := 0; i < len(decl.Sources); i++ {
		ith := decl.Sources[i]
		src := r.env.Column(module, ith.Name)
		// Sanity check length multiplier
		if i == 0 && src.datatype.AsUint() == nil {
			errors = append(errors, *r.srcmap.SyntaxError(ith, "fixed-width type required"))
		} else if i == 0 {
			multiplier = src.multiplier
		} else if multiplier != src.multiplier {
			// Problem
			errors = append(errors, *r.srcmap.SyntaxError(ith, "incompatible length multiplier"))
		}
		// All good, finalise column
		context := tr.NewContext(module, src.multiplier)
		r.env.FinaliseColumn(context, decl.Targets[i].Name, src.datatype)
	}
	// Done
	return errors
}

// Examine each constraint and attempt to resolve any variables used within
// them.  For example, a vanishing constraint may refer to some variable "X".
// Prior to this function being called, its not clear what "X" refers to --- it
// could refer to a column a constant, or even an alias.  The purpose of this
// pass is to: firstly, check that every variable refers to something which was
// declared; secondly, to determine what each variable represents (i.e. column
// access, a constant, etc).
func (r *resolver) resolveConstraints(circuit *Circuit) []SyntaxError {
	root := r.buildModuleScope("", circuit.Declarations)
	errs := r.resolveConstraintsInModule(root, circuit.Declarations)
	//
	for _, m := range circuit.Modules {
		module := r.buildModuleScope(m.Name, circuit.Declarations)
		// Process all declarations in the module
		merrs := r.resolveConstraintsInModule(module, m.Declarations)
		// Package up all errors
		errs = append(errs, merrs...)
	}
	//
	return errs
}

func (r *resolver) buildModuleScope(name string, decls []Declaration) Scope {
	var (
		scope *ModuleScope = r.env.NewModuleScope(name)
	)
	//
	for _, d := range decls {
		// Look for defcolumns decalarations only
		if c, ok := d.(*DefFun); ok {
			// TODO: sanity check if function already declared.
			scope.DeclareFunction(c.Name.Name, uint(len(c.Parameters)), c.Body)
		}
	}
	//
	return scope
}

// Helper for resolve constraints which considers those constraints declared in
// a particular module.
func (r *resolver) resolveConstraintsInModule(enclosing Scope, decls []Declaration) []SyntaxError {
	var errors []SyntaxError
	//
	for _, d := range decls {
		// Look for defcolumns decalarations only
		if _, ok := d.(*DefColumns); ok {
			// Safe to ignore.
		} else if c, ok := d.(*DefConstraint); ok {
			errors = append(errors, r.resolveDefConstraintInModule(enclosing, c)...)
		} else if c, ok := d.(*DefInRange); ok {
			errors = append(errors, r.resolveDefInRangeInModule(enclosing, c)...)
		} else if _, ok := d.(*DefInterleaved); ok {
			// Nothing to do here, since this assignment form contains no
			// expressions to be resolved.
		} else if c, ok := d.(*DefLookup); ok {
			errors = append(errors, r.resolveDefLookupInModule(enclosing, c)...)
		} else if _, ok := d.(*DefPermutation); ok {
			// Nothing to do here, since this assignment form contains no
			// expressions to be resolved.
		} else if c, ok := d.(*DefFun); ok {
			errors = append(errors, r.resolveDefFunInModule(enclosing, c)...)
		} else if c, ok := d.(*DefProperty); ok {
			errors = append(errors, r.resolveDefPropertyInModule(enclosing, c)...)
		} else {
			errors = append(errors, *r.srcmap.SyntaxError(d, "unknown declaration"))
		}
	}
	//
	return errors
}

// Resolve those variables appearing in either the guard or the body of this constraint.
func (r *resolver) resolveDefConstraintInModule(enclosing Scope, decl *DefConstraint) []SyntaxError {
	var (
		errors []SyntaxError
		scope  = NewLocalScope(enclosing, false)
	)
	// Resolve guard
	if decl.Guard != nil {
		errors = r.resolveExpressionInModule(scope, decl.Guard)
	}
	// Resolve constraint body
	errors = append(errors, r.resolveExpressionInModule(scope, decl.Constraint)...)
	// Done
	return errors
}

// Resolve those variables appearing in the body of this range constraint.
func (r *resolver) resolveDefInRangeInModule(enclosing Scope, decl *DefInRange) []SyntaxError {
	var (
		errors []SyntaxError
		scope  = NewLocalScope(enclosing, false)
	)
	// Resolve property body
	errors = append(errors, r.resolveExpressionInModule(scope, decl.Expr)...)
	// Done
	return errors
}

// Resolve those variables appearing in the body of this function.
func (r *resolver) resolveDefFunInModule(enclosing Scope, decl *DefFun) []SyntaxError {
	var (
		errors []SyntaxError
		scope  = NewLocalScope(enclosing, false)
	)
	// Declare parameters in local scope
	for _, p := range decl.Parameters {
		scope.DeclareLocal(p.Name)
	}
	// Resolve property body
	errors = append(errors, r.resolveExpressionInModule(scope, decl.Body)...)
	// Remove parameters from enclosing environment
	// Done
	return errors
}

// Resolve those variables appearing in the body of this lookup constraint.
func (r *resolver) resolveDefLookupInModule(enclosing Scope, decl *DefLookup) []SyntaxError {
	var (
		errors      []SyntaxError
		sourceScope = NewLocalScope(enclosing, true)
		targetScope = NewLocalScope(enclosing, true)
	)

	// Resolve source expressions
	errors = append(errors, r.resolveExpressionsInModule(sourceScope, decl.Sources)...)
	// Resolve target expressions
	errors = append(errors, r.resolveExpressionsInModule(targetScope, decl.Targets)...)
	// Done
	return errors
}

// Resolve those variables appearing in the body of this property assertion.
func (r *resolver) resolveDefPropertyInModule(enclosing Scope, decl *DefProperty) []SyntaxError {
	var (
		errors []SyntaxError
		scope  = NewLocalScope(enclosing, false)
	)
	// Resolve property body
	errors = append(errors, r.resolveExpressionInModule(scope, decl.Assertion)...)
	// Done
	return errors
}

// Resolve a sequence of zero or more expressions within a given module.  This
// simply resolves each of the arguments in turn, collecting any errors arising.
func (r *resolver) resolveExpressionsInModule(scope LocalScope, args []Expr) []SyntaxError {
	var errors []SyntaxError
	// Visit each argument
	for _, arg := range args {
		if arg != nil {
			errs := r.resolveExpressionInModule(scope, arg)
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
func (r *resolver) resolveExpressionInModule(scope LocalScope, expr Expr) []SyntaxError {
	if _, ok := expr.(*Constant); ok {
		return nil
	} else if v, ok := expr.(*Add); ok {
		return r.resolveExpressionsInModule(scope, v.Args)
	} else if v, ok := expr.(*Exp); ok {
		return r.resolveExpressionInModule(scope, v.Arg)
	} else if v, ok := expr.(*IfZero); ok {
		return r.resolveExpressionsInModule(scope, []Expr{v.Condition, v.TrueBranch, v.FalseBranch})
	} else if v, ok := expr.(*Invoke); ok {
		return r.resolveInvokeInModule(scope, v)
	} else if v, ok := expr.(*List); ok {
		return r.resolveExpressionsInModule(scope, v.Args)
	} else if v, ok := expr.(*Mul); ok {
		return r.resolveExpressionsInModule(scope, v.Args)
	} else if v, ok := expr.(*Normalise); ok {
		return r.resolveExpressionInModule(scope, v.Arg)
	} else if v, ok := expr.(*Sub); ok {
		return r.resolveExpressionsInModule(scope, v.Args)
	} else if v, ok := expr.(*VariableAccess); ok {
		return r.resolveVariableInModule(scope, v)
	} else {
		return r.srcmap.SyntaxErrors(expr, "unknown expression")
	}
}

// Resolve a specific invocation contained within some expression which, in
// turn, is contained within some module.  Note, qualified accesses are only
// permitted in a global context.
func (r *resolver) resolveInvokeInModule(scope LocalScope, expr *Invoke) []SyntaxError {
	// Resolve arguments
	if errors := r.resolveExpressionsInModule(scope, expr.Args); errors != nil {
		return errors
	}
	// Lookup the corresponding function definition.
	binding := scope.Bind(nil, expr.Name, true)
	// Check what we got
	if fnBinding, ok := binding.(*FunctionBinding); ok {
		expr.Binding = fnBinding
		return nil
	}
	//
	return r.srcmap.SyntaxErrors(expr, "unknown function")
}

// Resolve a specific variable access contained within some expression which, in
// turn, is contained within some module.  Note, qualified accesses are only
// permitted in a global context.
func (r *resolver) resolveVariableInModule(scope LocalScope,
	expr *VariableAccess) []SyntaxError {
	// Will identify module of variable
	//var module string = scope.EnclosingModule()
	var mid *uint
	// Check whether this is a qualified access, or not.
	if !scope.IsGlobal() && expr.Module != nil {
		return r.srcmap.SyntaxErrors(expr, "qualified access not permitted here")
	} else if expr.Module != nil && !scope.HasModule(*expr.Module) {
		return r.srcmap.SyntaxErrors(expr, fmt.Sprintf("unknown module %s", *expr.Module))
	} else if expr.Module != nil {
		tmp := scope.Module(*expr.Module)
		mid = &tmp
	}
	// Attempt resolve as a column access in enclosing module
	if expr.Binding = scope.Bind(mid, expr.Name, false); expr.Binding != nil {
		// Update context
		if !scope.FixContext(expr.Binding.Context()) {
			return r.srcmap.SyntaxErrors(expr, "conflicting context")
		}
		// Done
		return nil
	}
	// Unable to resolve variable
	return r.srcmap.SyntaxErrors(expr, "unknown symbol")
}
