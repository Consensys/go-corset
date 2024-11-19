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
	errs := r.resolveModules(circuit)
	// Allocate declared input columns
	errs = append(errs, r.resolveColumns(circuit)...)
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
func (r *resolver) resolveModules(circuit *Circuit) []SyntaxError {
	// Register the root module (which should always exist)
	r.env.RegisterModule("")
	//
	for _, m := range circuit.Modules {
		r.env.RegisterModule(m.Name)
	}
	// Done
	return nil
}

// Process all input column or column assignment declarations.
func (r *resolver) resolveColumns(circuit *Circuit) []SyntaxError {
	// Input columns must be allocated before assignemts, since the hir.Schema
	// separates these out.
	errs := r.resolveColumnsInModule("", circuit.Declarations)
	//
	for _, m := range circuit.Modules {
		// Process all declarations in the module
		merrs := r.resolveColumnsInModule(m.Name, m.Declarations)
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
func (r *resolver) resolveColumnsInModule(module string, decls []Declaration) []SyntaxError {
	errors := make([]SyntaxError, 0)
	mid := r.env.Module(module)
	//
	for _, d := range decls {
		// Look for defcolumns decalarations only
		if dcols, ok := d.(*DefColumns); ok {
			// Found one.
			for _, col := range dcols.Columns {
				if r.env.HasColumn(mid, col.Name) {
					err := r.srcmap.SyntaxError(col, fmt.Sprintf("column %s already declared in module %s", col.Name, module))
					errors = append(errors, *err)
				} else {
					context := tr.NewContext(mid, col.LengthMultiplier)
					r.env.RegisterColumn(context, col.Name, col.DataType)
				}
			}
		}
	}
	//
	return errors
}

type colInfo struct {
	length_multiplier uint
	datatype          schema.Type
}

// Initialise the column allocation from the definitions.
func (r *resolver) initialiseColumnAllocation(module string, decls []Declaration) (map[string]colInfo, []SyntaxError) {
	panic("TODO")
}

// Finalising the columns in the module is important to ensure that they are
// registered in the correct order.  This is because they must be registered in
// the order of occurence.  We can assume that, once we get here, then there are
// no errors with the column declarations.
func (r *resolver) finaliseColumnsAllocation(module string, decls []Declaration, alloc map[string]colInfo) []SyntaxError {
	mid := r.env.Module(module)
	// (1) register input columns.
	for _, d := range decls {
		// Look for defcolumns decalarations only
		if dcols, ok := d.(*DefColumns); ok {
			// Found one.
			for _, col := range dcols.Columns {
				context := tr.NewContext(mid, col.LengthMultiplier)
				r.env.RegisterColumn(context, col.Name, col.DataType)
			}
		}
	}
	// (2) register assignments.
	for _, d := range decls {
		if dInterleave, ok := d.(*DefInterleaved); ok {
			info := alloc[dInterleave.Target]
			context := tr.NewContext(mid, info.length_multiplier)
			r.env.RegisterColumn(context, dInterleave.Target, info.datatype)
		}
	}
}

// Resolve all assignment declarations.   Managing these is slightly more
// complex than for input columns, since they can depend upon each other.
// Furthermore, information about their form is not always clear from the
// declaration itself (e.g. the type of an interleaved column is determined by
// the types of its source columns, etc).
func (r *resolver) resolveAssignmentsInModule(module string, decls []Declaration) []SyntaxError {
	changed := true
	errors := make([]SyntaxError, 0)
	done := make(map[Declaration]bool, 0)
	// Keep going until no new assignments can be resolved.
	for changed {
		changed = false
		// Discard any previously generated errors and repeat.
		errors = make([]SyntaxError, 0)
		// Reconsider all outstanding assignments.
		for _, d := range decls {
			if dInterleave, ok := d.(*DefInterleaved); ok && !done[d] {
				if errs := r.resolveInterleavedAssignment(module, dInterleave); errs == nil {
					// Mark assignment as done, so we never visit it again.
					done[d] = true
					changed = true
				} else {
					// Combine errors
					errors = append(errors, errs...)
				}
			}
		}
	}
	//
	return errors
}

// Resolve an interleaving assignment.  This means: (1) checking that the
// relevant column was not already defined; (2) that the source columns have
// been defined.  If there are no problems here, then we register it after
// determining its type and length multiplier.
func (r *resolver) resolveInterleavedAssignment(module string, decl *DefInterleaved) []SyntaxError {
	var (
		length_multiplier uint
		datatype          schema.Type
		errors            []SyntaxError
	)
	// Determine enclosing module identifier
	mid := r.env.Module(module)
	// Check target column does not exist
	if _, ok := r.env.LookupColumn(mid, decl.Target); ok {
		errors = r.srcmap.SyntaxErrors(decl, fmt.Sprintf("column %s already declared in module %s", decl.Target, module))
	}
	// Check source columns do exist, whilst determining the type and length multiplier
	for i, source := range decl.Sources {
		// Check whether column exists or not.
		if info, ok := r.env.LookupColumn(mid, source); ok {
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
		} else {
			// Column does not exist!
			err := r.srcmap.SyntaxError(decl, fmt.Sprintf("unknown column %s in module %s", decl.Target, module))
			errors = append(errors, *err)
		}
	}
	//
	if errors != nil {
		return errors
	}
	// Determine actual length multiplier
	length_multiplier *= uint(len(decl.Sources))
	// Construct context for this column
	context := tr.NewContext(mid, length_multiplier)
	// Register new column
	r.env.RegisterColumn(context, decl.Target, datatype)
	// Done
	return nil
}

// Examine each constraint and attempt to resolve any variables used within
// them.  For example, a vanishing constraint may refer to some variable "X".
// Prior to this function being called, its not clear what "X" refers to --- it
// could refer to a column a constant, or even an alias.  The purpose of this
// pass is to: firstly, check that every variable refers to something which was
// declared; secondly, to determine what each variable represents (i.e. column
// access, a constant, etc).
func (r *resolver) resolveConstraints(circuit *Circuit) []SyntaxError {
	errs := r.resolveConstraintsInModule("", circuit.Declarations)
	//
	for _, m := range circuit.Modules {
		// Process all declarations in the module
		merrs := r.resolveConstraintsInModule(m.Name, m.Declarations)
		// Package up all errors
		errs = append(errs, merrs...)
	}
	//
	return errs
}

// Helper for resolve constraints which considers those constraints declared in
// a particular module.
func (r *resolver) resolveConstraintsInModule(module string, decls []Declaration) []SyntaxError {
	var errors []SyntaxError

	for _, d := range decls {
		// Look for defcolumns decalarations only
		if _, ok := d.(*DefColumns); ok {
			// Safe to ignore.
		} else if c, ok := d.(*DefConstraint); ok {
			errors = append(errors, r.resolveDefConstraintInModule(module, c)...)
		} else if c, ok := d.(*DefInRange); ok {
			errors = append(errors, r.resolveDefInRangeInModule(module, c)...)
		} else if _, ok := d.(*DefInterleaved); ok {
			// Nothing to do here, since this assignment form contains no
			// expressions to be resolved.
		} else if c, ok := d.(*DefLookup); ok {
			errors = append(errors, r.resolveDefLookupInModule(module, c)...)
		} else if c, ok := d.(*DefProperty); ok {
			errors = append(errors, r.resolveDefPropertyInModule(module, c)...)
		} else {
			errors = append(errors, *r.srcmap.SyntaxError(d, fmt.Sprintf("unknown declaration in module %s", module)))
		}
	}
	//
	return errors
}

// Resolve those variables appearing in either the guard or the body of this constraint.
func (r *resolver) resolveDefConstraintInModule(module string, decl *DefConstraint) []SyntaxError {
	var errors []SyntaxError
	if decl.Guard != nil {
		errors = r.resolveExpressionInModule(module, false, decl.Constraint)
	}
	// Resolve constraint body
	errors = append(errors, r.resolveExpressionInModule(module, false, decl.Constraint)...)
	// Done
	return errors
}

// Resolve those variables appearing in the body of this range constraint.
func (r *resolver) resolveDefInRangeInModule(module string, decl *DefInRange) []SyntaxError {
	var errors []SyntaxError
	// Resolve property body
	errors = append(errors, r.resolveExpressionInModule(module, false, decl.Expr)...)
	// Done
	return errors
}

// Resolve those variables appearing in the body of this lookup constraint.
func (r *resolver) resolveDefLookupInModule(module string, decl *DefLookup) []SyntaxError {
	var errors []SyntaxError
	// Resolve source expressions
	errors = append(errors, r.resolveExpressionsInModule(module, true, decl.Sources)...)
	// Resolve target expressions
	errors = append(errors, r.resolveExpressionsInModule(module, true, decl.Targets)...)
	// Done
	return errors
}

// Resolve those variables appearing in the body of this property assertion.
func (r *resolver) resolveDefPropertyInModule(module string, decl *DefProperty) []SyntaxError {
	var errors []SyntaxError
	// Resolve property body
	errors = append(errors, r.resolveExpressionInModule(module, false, decl.Assertion)...)
	// Done
	return errors
}

// Resolve a sequence of zero or more expressions within a given module.  This
// simply resolves each of the arguments in turn, collecting any errors arising.
func (r *resolver) resolveExpressionsInModule(module string, global bool, args []Expr) []SyntaxError {
	var errors []SyntaxError
	// Visit each argument
	for _, arg := range args {
		if arg != nil {
			errs := r.resolveExpressionInModule(module, global, arg)
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
func (r *resolver) resolveExpressionInModule(module string, global bool, expr Expr) []SyntaxError {
	if _, ok := expr.(*Constant); ok {
		return nil
	} else if v, ok := expr.(*Add); ok {
		return r.resolveExpressionsInModule(module, global, v.Args)
	} else if v, ok := expr.(*Exp); ok {
		return r.resolveExpressionInModule(module, global, v.Arg)
	} else if v, ok := expr.(*IfZero); ok {
		return r.resolveExpressionsInModule(module, global, []Expr{v.Condition, v.TrueBranch, v.FalseBranch})
	} else if v, ok := expr.(*List); ok {
		return r.resolveExpressionsInModule(module, global, v.Args)
	} else if v, ok := expr.(*Mul); ok {
		return r.resolveExpressionsInModule(module, global, v.Args)
	} else if v, ok := expr.(*Normalise); ok {
		return r.resolveExpressionInModule(module, global, v.Arg)
	} else if v, ok := expr.(*Sub); ok {
		return r.resolveExpressionsInModule(module, global, v.Args)
	} else if v, ok := expr.(*VariableAccess); ok {
		return r.resolveVariableInModule(module, global, v)
	} else {
		return r.srcmap.SyntaxErrors(expr, "unknown expression")
	}
}

// Resolve a specific variable access contained within some expression which, in
// turn, is contained within some module.  Note, qualified accesses are only
// permitted in a global context.
func (r *resolver) resolveVariableInModule(module string, global bool, expr *VariableAccess) []SyntaxError {
	// Check whether this is a qualified access, or not.
	if global && expr.Module != nil {
		module = *expr.Module
	} else if expr.Module != nil {
		return r.srcmap.SyntaxErrors(expr, "qualified access not permitted here")
	}
	// FIXME: handle qualified variable accesses
	mid := r.env.Module(module)
	// Attempt resolve as a column access in enclosing module
	if cinfo, ok := r.env.LookupColumn(mid, expr.Name); ok {
		ctx := tr.NewContext(mid, cinfo.multiplier)
		// Register the binding to complete resolution.
		expr.Binding = &Binder{true, ctx, cinfo.cid}
		// Done
		return nil
	}
	// Unable to resolve variable
	return r.srcmap.SyntaxErrors(expr, fmt.Sprintf("unknown symbol in module %s", module))
}
