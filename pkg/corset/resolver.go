package corset

import (
	"fmt"

	"github.com/consensys/go-corset/pkg/schema"
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
	errs := r.resolveColumns(scope, circuit)
	// Check expressions
	errs = append(errs, r.resolveConstraints(scope, circuit)...)
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

// Process all input column or column assignment declarations.
func (r *resolver) resolveColumns(scope *GlobalScope, circuit *Circuit) []SyntaxError {
	// Allocate input columns first.  These must all be done before any
	// assignments are allocated, since the hir.Schema separates these out.
	ierrs := r.resolveInputColumns(scope, circuit)
	// Now we can resolve any assignments.
	aerrs := r.resolveAssignments(scope, circuit)
	//
	return append(ierrs, aerrs...)
}

// Process all input column declarations.
func (r *resolver) resolveInputColumns(scope *GlobalScope, circuit *Circuit) []SyntaxError {
	// Input columns must be allocated before assignemts, since the hir.Schema
	// separates these out.
	errs := r.resolveInputColumnsInModule(scope.Module(""), circuit.Declarations)
	//
	for _, m := range circuit.Modules {
		// Process all declarations in the module
		merrs := r.resolveInputColumnsInModule(scope.Module(m.Name), m.Declarations)
		// Package up all errors
		errs = append(errs, merrs...)
	}
	//
	return errs
}

// Resolve all input columns in a given module.
func (r *resolver) resolveInputColumnsInModule(scope *ModuleScope, decls []Declaration) []SyntaxError {
	errors := make([]SyntaxError, 0)
	//
	for _, d := range decls {
		if dcols, ok := d.(*DefColumns); ok {
			// Found one.
			for _, col := range dcols.Columns {
				// Check whether column already exists
				if scope.Bind(nil, col.Name, false) != nil {
					msg := fmt.Sprintf("symbol %s already declared in %s", col.Name, scope.EnclosingModule())
					err := r.srcmap.SyntaxError(col, msg)
					errors = append(errors, *err)
				} else {
					// Declare new column
					scope.Declare(col.Name, false, NewColumnBinding(scope.EnclosingModule(),
						false, col.LengthMultiplier, col.DataType))
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
func (r *resolver) resolveAssignments(scope *GlobalScope, circuit *Circuit) []SyntaxError {
	// Input columns must be allocated before assignemts, since the hir.Schema
	// separates these out.
	errs := r.resolveAssignmentsInModule(scope.Module(""), circuit.Declarations)
	//
	for _, m := range circuit.Modules {
		// Process all declarations in the module
		merrs := r.resolveAssignmentsInModule(scope.Module(m.Name), m.Declarations)
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
func (r *resolver) resolveAssignmentsInModule(scope *ModuleScope, decls []Declaration) []SyntaxError {
	if errors := r.initialiseAssignmentsInModule(scope, decls); len(errors) > 0 {
		return errors
	}
	// Check assignments
	if errors := r.checkAssignmentsInModule(scope, decls); len(errors) > 0 {
		return errors
	}
	// Iterate until all columns finalised
	return r.finaliseAssignmentsInModule(scope, decls)
}

// Initialise the column allocation from the available declarations, whilst
// identifying any duplicate declarations.  Observe that, for some declarations,
// the initial assignment is incomplete because information about dependent
// columns may not be available.  So, the goal of the subsequent phase is to
// flesh out this missing information.
func (r *resolver) initialiseAssignmentsInModule(scope *ModuleScope, decls []Declaration) []SyntaxError {
	module := scope.EnclosingModule()
	errors := make([]SyntaxError, 0)
	//
	for _, d := range decls {
		if col, ok := d.(*DefInterleaved); ok {
			if binding := scope.Bind(nil, col.Target, false); binding != nil {
				err := r.srcmap.SyntaxError(col, fmt.Sprintf("symbol %s already declared in %s", col.Target, module))
				errors = append(errors, *err)
			} else {
				// Register incomplete (assignment) column.
				scope.Declare(col.Target, false, NewColumnBinding(module, true, 0, nil))
			}
		} else if col, ok := d.(*DefPermutation); ok {
			for _, c := range col.Targets {
				if binding := scope.Bind(nil, c.Name, false); binding != nil {
					err := r.srcmap.SyntaxError(col, fmt.Sprintf("symbol %s already declared in %s", c.Name, module))
					errors = append(errors, *err)
				} else {
					// Register incomplete (assignment) column.
					scope.Declare(c.Name, false, NewColumnBinding(scope.EnclosingModule(), true, 0, nil))
				}
			}
		}
	}
	// Done
	return errors
}

func (r *resolver) checkAssignmentsInModule(scope *ModuleScope, decls []Declaration) []SyntaxError {
	errors := make([]SyntaxError, 0)
	//
	for _, d := range decls {
		if col, ok := d.(*DefInterleaved); ok {
			for _, c := range col.Sources {
				if scope.Bind(nil, c.Name, false) == nil {
					errors = append(errors, *r.srcmap.SyntaxError(c, "unknown source column"))
				}
			}
		} else if col, ok := d.(*DefPermutation); ok {
			for _, c := range col.Sources {
				if scope.Bind(nil, c.Name, false) == nil {
					errors = append(errors, *r.srcmap.SyntaxError(c, "unknown source column"))
				}
			}
		}
	}
	// Done
	return errors
}

// Iterate the column allocation to a fix point by iteratively fleshing out column information.
func (r *resolver) finaliseAssignmentsInModule(scope *ModuleScope, decls []Declaration) []SyntaxError {
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
				if r.columnsAreFinalised(scope, col.Sources) {
					// Finalise assignment and handle any errors
					errs := r.finaliseInterleavedAssignment(scope, col)
					errors = append(errors, errs...)
					// Record that a new assignment is available.
					changed = changed || len(errs) == 0
				} else {
					complete = false
					incomplete = d
				}
			} else if col, ok := d.(*DefPermutation); ok {
				// Check whether dependencies are resolved or not.
				if r.columnsAreFinalised(scope, col.Sources) {
					// Finalise assignment and handle any errors
					errs := r.finalisePermutationAssignment(scope, col)
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

// Check that a given set of source columns have been finalised.  This is
// important, since we cannot finalise an assignment until all of its
// dependencies have themselves been finalised.
func (r *resolver) columnsAreFinalised(scope *ModuleScope, columns []*DefName) bool {
	for _, col := range columns {
		// Look up information
		info := scope.Bind(nil, col.Name, false).(*ColumnBinding)
		// Check whether its finalised
		if info.multiplier == 0 {
			// Nope, not yet.
			return false
		}
	}
	//
	return true
}

// Finalise an interleaving assignment.  Since the assignment would already been
// initialised, all we need to do is determine the appropriate type and length
// multiplier for the interleaved column.  This can still result in an error,
// for example, if the multipliers between interleaved columns are incompatible,
// etc.
func (r *resolver) finaliseInterleavedAssignment(scope *ModuleScope, decl *DefInterleaved) []SyntaxError {
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
		info := scope.Bind(nil, source.Name, false).(*ColumnBinding)
		//
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
		// Lookup existing declaration
		info := scope.Bind(nil, decl.Target, false).(*ColumnBinding)
		// Update with completed information
		info.multiplier = length_multiplier
		info.datatype = datatype
	}
	// Done
	return errors
}

// Finalise a permutation assignment.  Since the assignment would already been
// initialised, this is actually quite easy to do.
func (r *resolver) finalisePermutationAssignment(scope *ModuleScope, decl *DefPermutation) []SyntaxError {
	var (
		multiplier uint = 0
		errors     []SyntaxError
	)
	// Finalise each column in turn
	for i := 0; i < len(decl.Sources); i++ {
		ith := decl.Sources[i]
		// Lookup source of column being permuted
		source := scope.Bind(nil, ith.Name, false).(*ColumnBinding)
		// Sanity check length multiplier
		if i == 0 && source.datatype.AsUint() == nil {
			errors = append(errors, *r.srcmap.SyntaxError(ith, "fixed-width type required"))
		} else if i == 0 {
			multiplier = source.multiplier
		} else if multiplier != source.multiplier {
			// Problem
			errors = append(errors, *r.srcmap.SyntaxError(ith, "incompatible length multiplier"))
		}
		// All good, finalise target column
		target := scope.Bind(nil, decl.Targets[i].Name, false).(*ColumnBinding)
		// Update with completed information
		target.multiplier = source.multiplier
		target.datatype = source.datatype
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
func (r *resolver) resolveConstraints(scope *GlobalScope, circuit *Circuit) []SyntaxError {
	errs := r.resolveConstraintsInModule(scope.Module(""), circuit.Declarations)
	//
	for _, m := range circuit.Modules {
		// Process all declarations in the module
		merrs := r.resolveConstraintsInModule(scope.Module(m.Name), m.Declarations)
		// Package up all errors
		errs = append(errs, merrs...)
	}
	//
	return errs
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
	// Check whether this is a qualified access, or not.
	if !scope.IsGlobal() && expr.Module != nil {
		return r.srcmap.SyntaxErrors(expr, "qualified access not permitted here")
	} else if expr.Module != nil && !scope.HasModule(*expr.Module) {
		return r.srcmap.SyntaxErrors(expr, fmt.Sprintf("unknown module %s", *expr.Module))
	}
	// Attempt resolve this variable access, noting that it definitely does not
	// refer to a function.
	if expr.Binding = scope.Bind(expr.Module, expr.Name, false); expr.Binding != nil {
		// Update context
		binding, ok := expr.Binding.(*ColumnBinding)
		if ok && !scope.FixContext(binding.Context()) {
			return r.srcmap.SyntaxErrors(expr, "conflicting context")
		}
		// Done
		return nil
	}
	// Unable to resolve variable
	return r.srcmap.SyntaxErrors(expr, "unknown symbol")
}
