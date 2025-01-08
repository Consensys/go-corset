package corset

import (
	"math/big"

	"github.com/consensys/go-corset/pkg/sexp"
)

// PreprocessCircuit performs preprocessing prior to final translation.
// Specifically, it expands all invocations, reductions and for loops.  Thus,
// final translation is greatly simplified after this step.
func PreprocessCircuit(debug bool, srcmap *sexp.SourceMaps[Node],
	circuit *Circuit) []SyntaxError {
	// Construct fresh preprocessor
	p := preprocessor{debug, srcmap}
	// Preprocess all declarations
	return p.preprocessDeclarations(circuit)
}

// Preprocessor performs preprocessing prior to final translation. Specifically,
// it expands all invocations, reductions and for loops.  Thus, final
// translation is greatly simplified after this step.
type preprocessor struct {
	// Debug enables the use of debug constraints.
	debug bool
	// Source maps nodes in the circuit back to the spans in their original
	// source files.  This is needed when reporting syntax errors to generate
	// highlights of the relevant source line(s) in question.
	srcmap *sexp.SourceMaps[Node]
}

// preprocess all assignment or constraint declarations in the circuit.
func (p *preprocessor) preprocessDeclarations(circuit *Circuit) []SyntaxError {
	errors := p.preprocessDeclarationsInModule(circuit.Declarations)
	// preprocess each module
	for _, m := range circuit.Modules {
		errs := p.preprocessDeclarationsInModule(m.Declarations)
		errors = append(errors, errs...)
	}
	// Done
	return errors
}

// preprocess all assignment or constraint declarations in a given module within
// the circuit.
func (p *preprocessor) preprocessDeclarationsInModule(decls []Declaration) []SyntaxError {
	var errors []SyntaxError
	//
	for _, d := range decls {
		errs := p.preprocessDeclaration(d)
		errors = append(errors, errs...)
	}
	// Done
	return errors
}

// preprocess an assignment or constraint declarartion which occurs within a
// given module.
func (p *preprocessor) preprocessDeclaration(decl Declaration) []SyntaxError {
	var errors []SyntaxError
	//
	switch d := decl.(type) {
	case *DefAliases:
		// ignore
	case *DefColumns:
		// ignore
	case *DefConst:
		// ignore
	case *DefConstraint:
		errors = p.preprocessDefConstraint(d)
	case *DefFun:
		// ignore
	case *DefInRange:
		errors = p.preprocessDefInRange(d)
	case *DefInterleaved:
		// ignore
	case *DefLookup:
		errors = p.preprocessDefLookup(d)
	case *DefPermutation:
		// ignore
	case *DefPerspective:
		errors = p.preprocessDefPerspective(d)
	case *DefProperty:
		errors = p.preprocessDefProperty(d)
	default:
		// Error handling
		panic("unknown declaration")
	}
	//
	return errors
}

// preprocess a "defconstraint" declaration.
func (p *preprocessor) preprocessDefConstraint(decl *DefConstraint) []SyntaxError {
	var (
		constraint_errors []SyntaxError
		guard_errors      []SyntaxError
	)
	// preprocess constraint body
	decl.Constraint, constraint_errors = p.preprocessExpressionInModule(decl.Constraint)
	// preprocess (optional) guard
	decl.Guard, guard_errors = p.preprocessOptionalExpressionInModule(decl.Guard)
	// Combine errors
	return append(constraint_errors, guard_errors...)
}

// preprocess a "deflookup" declaration.
//
//nolint:staticcheck
func (p *preprocessor) preprocessDefLookup(decl *DefLookup) []SyntaxError {
	var (
		source_errs []SyntaxError
		target_errs []SyntaxError
	)
	// preprocess source expressions
	decl.Sources, source_errs = p.preprocessExpressionsInModule(decl.Sources)
	decl.Targets, target_errs = p.preprocessExpressionsInModule(decl.Targets)
	// Combine errors
	return append(source_errs, target_errs...)
}

// preprocess a "definrange" declaration.
func (p *preprocessor) preprocessDefInRange(decl *DefInRange) []SyntaxError {
	var errors []SyntaxError
	// preprocess constraint body
	decl.Expr, errors = p.preprocessExpressionInModule(decl.Expr)
	// Done
	return errors
}

// preprocess a "defperspective" declaration.
func (p *preprocessor) preprocessDefPerspective(decl *DefPerspective) []SyntaxError {
	var errors []SyntaxError
	// preprocess selector expression
	decl.Selector, errors = p.preprocessExpressionInModule(decl.Selector)
	// Combine errors
	return errors
}

// preprocess a "defproperty" declaration.
func (p *preprocessor) preprocessDefProperty(decl *DefProperty) []SyntaxError {
	var errors []SyntaxError
	// preprocess constraint body
	decl.Assertion, errors = p.preprocessExpressionInModule(decl.Assertion)
	// Done
	return errors
}

// preprocess an optional expression in a given context.  That is an expression
// which maybe nil (i.e. doesn't exist).  In such case, nil is returned (i.e.
// without any errors).
func (p *preprocessor) preprocessOptionalExpressionInModule(expr Expr) (Expr, []SyntaxError) {
	//
	if expr != nil {
		return p.preprocessExpressionInModule(expr)
	}

	return nil, nil
}

// preprocess a sequence of zero or more expressions enclosed in a given module.
// All expressions are expected to be non-voidable (see below for more on
// voidability).
func (p *preprocessor) preprocessExpressionsInModule(exprs []Expr) ([]Expr, []SyntaxError) {
	//
	errors := []SyntaxError{}
	hirExprs := make([]Expr, len(exprs))
	// Iterate each expression in turn
	for i, e := range exprs {
		if e != nil {
			var errs []SyntaxError
			hirExprs[i], errs = p.preprocessExpressionInModule(e)
			errors = append(errors, errs...)
			// Check for non-voidability
			if hirExprs[i] == nil {
				errors = append(errors, *p.srcmap.SyntaxError(e, "void expression not permitted here"))
			}
		}
	}
	//
	return hirExprs, errors
}

// preprocess a sequence of zero or more expressions enclosed in a given module.
// A key aspect of this function is that it additionally accounts for "voidable"
// expressions.  That is, essentially, to account for debug constraints which
// only exist in debug mode.  Hence, when debug mode is not enabled, then a
// debug constraint is "void".
func (p *preprocessor) preprocessVoidableExpressionsInModule(exprs []Expr) ([]Expr, []SyntaxError) {
	//
	errors := []SyntaxError{}
	hirExprs := make([]Expr, len(exprs))
	nils := 0
	// Iterate each expression in turn
	for i, e := range exprs {
		if e != nil {
			var errs []SyntaxError
			hirExprs[i], errs = p.preprocessExpressionInModule(e)
			errors = append(errors, errs...)
			// Update dirty flag
			if hirExprs[i] == nil {
				nils++
			}
		}
	}
	// Nil check.
	if nils == 0 {
		// Done
		return hirExprs, errors
	}
	// Stip nils. Recall that nils can arise legitimately when we have debug
	// constraints, but debug mode is not enabled.  In such case, we want to
	// strip them out.  Since this is a rare occurrence, we try to keep the happy
	// path efficient.
	nHirExprs := make([]Expr, len(exprs)-nils)
	i := 0
	// Strip out nils
	for _, e := range hirExprs {
		if e != nil {
			nHirExprs[i] = e
			i++
		}
	}
	//
	return nHirExprs, errors
}

// preprocess an expression situated in a given context.  The context is
// necessary to resolve unqualified names (e.g. for column access, function
// invocations, etc).
func (p *preprocessor) preprocessExpressionInModule(expr Expr) (Expr, []SyntaxError) {
	var (
		nexpr  Expr
		errors []SyntaxError
	)
	//
	switch e := expr.(type) {
	case *ArrayAccess:
		arg, errs := p.preprocessExpressionInModule(e.arg)
		nexpr, errors = &ArrayAccess{e.path, arg, e.binding}, errs
	case *Add:
		args, errs := p.preprocessExpressionsInModule(e.Args)
		nexpr, errors = &Add{args}, errs
	case *Constant:
		return e, nil
	case *Debug:
		if p.debug {
			return p.preprocessExpressionInModule(e.Arg)
		}
		// When debug is not enabled, return "void".
		return nil, nil
	case *Exp:
		arg, errs1 := p.preprocessExpressionInModule(e.Arg)
		pow, errs2 := p.preprocessExpressionInModule(e.Pow)
		// Done
		nexpr, errors = &Exp{arg, pow}, append(errs1, errs2...)
	case *For:
		return p.preprocessForInModule(e)
	case *If:
		args, errs := p.preprocessExpressionsInModule([]Expr{e.Condition, e.TrueBranch, e.FalseBranch})
		// Construct appropriate if form
		nexpr, errors = &If{e.kind, args[0], args[1], args[2]}, errs
	case *Invoke:
		return p.preprocessInvokeInModule(e)
	case *List:
		args, errs := p.preprocessVoidableExpressionsInModule(e.Args)
		nexpr, errors = &List{args}, errs
	case *Mul:
		args, errs := p.preprocessExpressionsInModule(e.Args)
		nexpr, errors = &Mul{args}, errs
	case *Normalise:
		arg, errs := p.preprocessExpressionInModule(e.Arg)
		nexpr, errors = &Normalise{arg}, errs
	case *Reduce:
		return p.preprocessReduceInModule(e)
	case *Sub:
		args, errs := p.preprocessExpressionsInModule(e.Args)
		nexpr, errors = &Sub{args}, errs
	case *Shift:
		arg, errs := p.preprocessExpressionInModule(e.Arg)
		nexpr, errors = &Shift{arg, e.Shift}, errs
	case *VariableAccess:
		return e, nil
	default:
		return nil, p.srcmap.SyntaxErrors(expr, "unknown expression encountered during translation")
	}
	// Copy over source information
	p.srcmap.Copy(expr, nexpr)
	// Done
	return nexpr, errors
}

func (p *preprocessor) preprocessForInModule(expr *For) (Expr, []SyntaxError) {
	var (
		errors  []SyntaxError
		mapping map[uint]Expr = make(map[uint]Expr)
	)
	// Determine range for index variable
	n := expr.End - expr.Start + 1
	args := make([]Expr, n)
	// Expand body n times
	for i := uint(0); i < n; i++ {
		var errs []SyntaxError
		// Substitute through for i
		mapping[expr.Binding.index] = &Constant{*big.NewInt(int64(i + expr.Start))}
		ith := Substitute(expr.Body, mapping, p.srcmap)
		// preprocess subsituted expression
		args[i], errs = p.preprocessExpressionInModule(ith)
		errors = append(errors, errs...)
	}
	// Error check
	if len(errors) != 0 {
		return nil, errors
	}
	// Done
	return &List{args}, nil
}

func (p *preprocessor) preprocessInvokeInModule(expr *Invoke) (Expr, []SyntaxError) {
	if expr.signature != nil {
		body := expr.signature.Apply(expr.Args(), p.srcmap)
		return p.preprocessExpressionInModule(body)
	}
	//
	return nil, p.srcmap.SyntaxErrors(expr, "unbound function")
}

func (p *preprocessor) preprocessReduceInModule(expr *Reduce) (Expr, []SyntaxError) {
	body, errors := p.preprocessExpressionInModule(expr.arg)
	//
	if list, ok := body.(*List); !ok {
		return nil, append(errors, *p.srcmap.SyntaxError(expr.arg, "expected list"))
	} else if sig := expr.signature; sig == nil {
		return nil, append(errors, *p.srcmap.SyntaxError(expr.arg, "unbound function"))
	} else {
		reduction := list.Args[0]
		// Build reduction
		for i := 1; i < len(list.Args); i++ {
			reduction = sig.Apply([]Expr{reduction, list.Args[i]}, p.srcmap)
		}
		// done
		return reduction, errors
	}
}
