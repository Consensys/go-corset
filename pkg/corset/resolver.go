package corset

import (
	"fmt"

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
	errs = append(errs, r.resolveInputColumns(circuit)...)
	// TODO: Allocate declared assignments
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

// Process all input (column) declarations.  These must be allocated before
// assignemts, since the hir.Schema separates these out.  Again, if any
// duplicates are found then one or more errors will be reported.
func (r *resolver) resolveInputColumns(circuit *Circuit) []SyntaxError {
	errs := r.resolveInputColumnsInModule(r.env.Module(""), circuit.Declarations)
	//
	for _, m := range circuit.Modules {
		// The module must exist given after resolveModules.
		ctx := r.env.Module(m.Name)
		// Process all declarations in the module
		merrs := r.resolveInputColumnsInModule(ctx, m.Declarations)
		// Package up all errors
		errs = append(errs, merrs...)
	}
	//
	return errs
}

func (r *resolver) resolveInputColumnsInModule(module uint, decls []Declaration) []SyntaxError {
	var errors []SyntaxError
	//
	for _, d := range decls {
		// Look for defcolumns decalarations only
		if dcols, ok := d.(*DefColumns); ok {
			// Found one.
			for _, col := range dcols.Columns {
				if r.env.HasColumn(module, col.Name) {
					errors = append(errors, *r.srcmap.SyntaxError(col, "duplicate declaration"))
				} else {
					context := tr.NewContext(module, col.LengthMultiplier)
					r.env.RegisterColumn(context, col.Name, col.DataType)
				}
			}
		}
	}
	//
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
	if _, ok := r.env.LookupColumn(mid, expr.Name); ok {
		return nil
	}
	// Unable to resolve variable
	return r.srcmap.SyntaxErrors(expr, fmt.Sprintf("unknown symbol in module %s", module))
}
