package corset

import (
	"fmt"

	"github.com/consensys/go-corset/pkg/sexp"
)

// TypeCheckCircuit performs a type checking pass over the circuit to ensure
// types are used correctly.  Additionally, this resolves some ambiguities
// arising from the possibility of overloading function calls, etc.
func TypeCheckCircuit(srcmap *sexp.SourceMaps[Node],
	circuit *Circuit) []SyntaxError {
	// Construct fresh typeCheckor
	p := typeChecker{srcmap}
	// typeCheck all declarations
	return p.typeCheckDeclarations(circuit)
}

// typeCheckor performs typeChecking prior to final translation. Specifically,
// it expands all invocations, reductions and for loops.  Thus, final
// translation is greatly simplified after this step.
type typeChecker struct {
	// Source maps nodes in the circuit back to the spans in their original
	// source files.  This is needed when reporting syntax errors to generate
	// highlights of the relevant source line(s) in question.
	srcmap *sexp.SourceMaps[Node]
}

// typeCheck all assignment or constraint declarations in the circuit.
func (p *typeChecker) typeCheckDeclarations(circuit *Circuit) []SyntaxError {
	errors := p.typeCheckDeclarationsInModule(circuit.Declarations)
	// typeCheck each module
	for _, m := range circuit.Modules {
		errs := p.typeCheckDeclarationsInModule(m.Declarations)
		errors = append(errors, errs...)
	}
	// Done
	return errors
}

// typeCheck all assignment or constraint declarations in a given module within
// the circuit.
func (p *typeChecker) typeCheckDeclarationsInModule(decls []Declaration) []SyntaxError {
	var errors []SyntaxError
	//
	for _, d := range decls {
		errs := p.typeCheckDeclaration(d)
		errors = append(errors, errs...)
	}
	// Done
	return errors
}

// typeCheck an assignment or constraint declarartion which occurs within a
// given module.
func (p *typeChecker) typeCheckDeclaration(decl Declaration) []SyntaxError {
	var errors []SyntaxError
	//
	switch d := decl.(type) {
	case *DefAliases:
		// ignore
	case *DefColumns:
		// ignore
	case *DefConst:
		errors = p.typeCheckDefConstInModule(d)
	case *DefConstraint:
		errors = p.typeCheckDefConstraint(d)
	case *DefFun:
		errors = p.typeCheckDefFunInModule(d)
	case *DefInRange:
		errors = p.typeCheckDefInRange(d)
	case *DefInterleaved:
		// ignore
	case *DefLookup:
		errors = p.typeCheckDefLookup(d)
	case *DefPermutation:
		// ignore
	case *DefPerspective:
		errors = p.typeCheckDefPerspective(d)
	case *DefProperty:
		errors = p.typeCheckDefProperty(d)
	default:
		// Error handling
		panic("unknown declaration")
	}
	//
	return errors
}

// Type check one or more constant definitions within a given module.
func (p *typeChecker) typeCheckDefConstInModule(decl *DefConst) []SyntaxError {
	var errors []SyntaxError
	//
	for _, c := range decl.constants {
		// Resolve constant body
		_, errs := p.typeCheckExpressionInModule(c.binding.value)
		// Accumulate errors
		errors = append(errors, errs...)
	}
	//
	return errors
}

// typeCheck a "defconstraint" declaration.
func (p *typeChecker) typeCheckDefConstraint(decl *DefConstraint) []SyntaxError {
	// typeCheck (optional) guard
	guard_t, guard_errors := p.typeCheckOptionalExpressionInModule(decl.Guard)
	// typeCheck constraint body
	constraint_t, constraint_errors := p.typeCheckExpressionInModule(decl.Constraint)
	// Check guard type
	if guard_t != nil && guard_t.HasLoobeanSemantics() {
		err := p.srcmap.SyntaxError(decl.Guard, "unexpected loobean guard")
		guard_errors = append(guard_errors, *err)
	}
	// Check constraint type
	if constraint_t != nil && !constraint_t.HasLoobeanSemantics() {
		msg := fmt.Sprintf("expected loobean constraint (found %s)", constraint_t.String())
		err := p.srcmap.SyntaxError(decl.Constraint, msg)
		constraint_errors = append(constraint_errors, *err)
	}
	// Combine errors
	return append(constraint_errors, guard_errors...)
}

// Type check the body of a function.
func (p *typeChecker) typeCheckDefFunInModule(decl *DefFun) []SyntaxError {
	// Resolve property body
	_, errors := p.typeCheckExpressionInModule(decl.Body())
	// FIXME: type check return?
	// Done
	return errors
}

// typeCheck a "deflookup" declaration.
//
//nolint:staticcheck
func (p *typeChecker) typeCheckDefLookup(decl *DefLookup) []SyntaxError {
	// typeCheck source expressions
	_, source_errs := p.typeCheckExpressionsInModule(decl.Sources)
	_, target_errs := p.typeCheckExpressionsInModule(decl.Targets)
	// Combine errors
	return append(source_errs, target_errs...)
}

// typeCheck a "definrange" declaration.
func (p *typeChecker) typeCheckDefInRange(decl *DefInRange) []SyntaxError {
	// typeCheck constraint body
	_, errors := p.typeCheckExpressionInModule(decl.Expr)
	// Done
	return errors
}

// typeCheck a "defperspective" declaration.
func (p *typeChecker) typeCheckDefPerspective(decl *DefPerspective) []SyntaxError {
	// typeCheck selector expression
	_, errors := p.typeCheckExpressionInModule(decl.Selector)
	// Combine errors
	return errors
}

// typeCheck a "defproperty" declaration.
func (p *typeChecker) typeCheckDefProperty(decl *DefProperty) []SyntaxError {
	// type check constraint body
	_, errors := p.typeCheckExpressionInModule(decl.Assertion)
	// Done
	return errors
}

// typeCheck an optional expression in a given context.  That is an expression
// which maybe nil (i.e. doesn't exist).  In such case, nil is returned (i.e.
// without any errors).
func (p *typeChecker) typeCheckOptionalExpressionInModule(expr Expr) (Type, []SyntaxError) {
	//
	if expr != nil {
		return p.typeCheckExpressionInModule(expr)
	}
	//
	return nil, nil
}

// typeCheck a sequence of zero or more expressions enclosed in a given module.
// All expressions are expected to be non-voidable (see below for more on
// voidability).
func (p *typeChecker) typeCheckExpressionsInModule(exprs []Expr) ([]Type, []SyntaxError) {
	errors := []SyntaxError{}
	types := make([]Type, len(exprs))
	// Iterate each expression in turn
	for i, e := range exprs {
		if e == nil {
			continue
		}
		//
		var errs []SyntaxError
		types[i], errs = p.typeCheckExpressionInModule(e)
		errors = append(errors, errs...)
		// Sanity check what we got back
		if types[i] == nil {
			return nil, errors
		}
	}
	//
	return types, errors
}

// typeCheck an expression situated in a given context.  The context is
// necessary to resolve unqualified names (e.g. for column access, function
// invocations, etc).
func (p *typeChecker) typeCheckExpressionInModule(expr Expr) (Type, []SyntaxError) {
	switch e := expr.(type) {
	case *ArrayAccess:
		return p.typeCheckArrayAccessInModule(e)
	case *Add:
		types, errs := p.typeCheckExpressionsInModule(e.Args)
		return LeastUpperBoundAll(types), errs
	case *Constant:
		nbits := e.Val.BitLen()
		return NewUintType(uint(nbits)), nil
	case *Debug:
		return p.typeCheckExpressionInModule(e.Arg)
	case *Exp:
		arg_t, errs1 := p.typeCheckExpressionInModule(e.Arg)
		_, errs2 := p.typeCheckExpressionInModule(e.Pow)
		// Done
		return arg_t, append(errs1, errs2...)
	case *For:
		// TODO: update environment with type of index variable.
		return p.typeCheckExpressionInModule(e.Body)
	case *If:
		return p.typeCheckIfInModule(e)
	case *Invoke:
		return p.typeCheckInvokeInModule(e)
	case *Let:
		return p.typeCheckLetInModule(e)
	case *List:
		types, errs := p.typeCheckExpressionsInModule(e.Args)
		return LeastUpperBoundAll(types), errs
	case *Mul:
		types, errs := p.typeCheckExpressionsInModule(e.Args)
		return GreatestLowerBoundAll(types), errs
	case *Normalise:
		_, errs := p.typeCheckExpressionInModule(e.Arg)
		// Normalise guaranteed to return either 0 or 1.
		return NewUintType(1), errs
	case *Reduce:
		return p.typeCheckReduceInModule(e)
	case *Shift:
		arg_t, arg_errs := p.typeCheckExpressionInModule(e.Arg)
		_, shf_errs := p.typeCheckExpressionInModule(e.Shift)
		// combine errors
		return arg_t, append(arg_errs, shf_errs...)
	case *Sub:
		types, errs := p.typeCheckExpressionsInModule(e.Args)
		return LeastUpperBoundAll(types), errs
	case *VariableAccess:
		return p.typeCheckVariableInModule(e)
	default:
		return nil, p.srcmap.SyntaxErrors(expr, "unknown expression encountered during translation")
	}
}

// Type check an array access expression.  The main thing is to check that the
// column being accessed was originally defined as an array column.
func (p *typeChecker) typeCheckArrayAccessInModule(expr *ArrayAccess) (Type, []SyntaxError) {
	// Type check index expression
	_, errs := p.typeCheckExpressionInModule(expr.arg)
	// NOTE: following cast safe because resolver already checked them.
	binding := expr.Binding().(*ColumnBinding)
	if arr_t, ok := binding.dataType.(*ArrayType); !ok {
		return nil, append(errs, *p.srcmap.SyntaxError(expr, "expected array column"))
	} else {
		return arr_t.element, errs
	}
}

// Type an if condition contained within some expression which, in turn, is
// contained within some module.  An important step occurrs here where, based on
// the semantics of the condition, this is inferred as an "if-zero" or an
// "if-notzero".
func (p *typeChecker) typeCheckIfInModule(expr *If) (Type, []SyntaxError) {
	types, errs := p.typeCheckExpressionsInModule([]Expr{expr.Condition, expr.TrueBranch, expr.FalseBranch})
	// Sanity check
	if len(errs) != 0 || types == nil {
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
		return nil, p.srcmap.SyntaxErrors(expr.Condition, "invalid condition (neither loobean nor boolean)")
	}
	// Join result types
	return GreatestLowerBoundAll(types[1:]), errs
}

func (p *typeChecker) typeCheckInvokeInModule(expr *Invoke) (Type, []SyntaxError) {
	if binding, ok := expr.fn.binding.(FunctionBinding); !ok {
		// We don't return an error here, since one would already have been
		// generated during resolution.
		return nil, nil
	} else if argTypes, errors := p.typeCheckExpressionsInModule(expr.args); len(errors) > 0 {
		return nil, errors
	} else if argTypes == nil {
		// An upstream expression could not because of a resolution error.
		return nil, nil
	} else if signature := binding.Select(argTypes); signature != nil {
		// Check arguments are accepted, based on their type.
		for i := 0; i < len(argTypes); i++ {
			expected := signature.Parameter(uint(i))
			actual := argTypes[i]
			// subtype check
			if actual != nil && !actual.SubtypeOf(expected) {
				msg := fmt.Sprintf("expected type %s (found %s)", expected, actual)
				errors = append(errors, *p.srcmap.SyntaxError(expr.args[i], msg))
			}
		}
		// Finalise the selected signature for future reference.
		expr.Finalise(signature)
		//
		if len(errors) != 0 {
			return nil, errors
		} else if signature.Return() != nil {
			// no need, it was provided
			return signature.Return(), nil
		}
		// TODO: this is potentially expensive, and it would likely be good if we
		// could avoid it.
		body := signature.Apply(expr.Args(), nil)
		// Dig out the type
		return p.typeCheckExpressionInModule(body)
	}
	// ambiguous invocation
	return nil, p.srcmap.SyntaxErrors(expr.fn, "ambiguous invocation")
}

func (p *typeChecker) typeCheckLetInModule(expr *Let) (Type, []SyntaxError) {
	// NOTE: there is a limitation here since we are using the type of the
	// assigned expressions.  It would be nice to retain this, but it would
	// require a more flexible notion of environment than we currently have.
	if types, arg_errors := p.typeCheckExpressionsInModule(expr.Args); types != nil {
		// Update type for let-bound variables.
		for i := range expr.Vars {
			if types[i] != nil {
				expr.Vars[i].datatype = types[i]
			}
		}
		// Type check body
		body_t, body_errors := p.typeCheckExpressionInModule(expr.Body)
		//
		return body_t, append(arg_errors, body_errors...)
	} else {
		return nil, arg_errors
	}
}

func (p *typeChecker) typeCheckReduceInModule(expr *Reduce) (Type, []SyntaxError) {
	var signature *FunctionSignature
	// Type check body of reduction
	body_t, errors := p.typeCheckExpressionInModule(expr.arg)
	// Following safe as resolver checked this already.
	if binding, ok := expr.fn.binding.(FunctionBinding); ok && body_t != nil {
		//
		if signature = binding.Select([]Type{body_t, body_t}); signature != nil {
			// Check left parameter type
			if !body_t.SubtypeOf(signature.Parameter(0)) {
				msg := fmt.Sprintf("expected type %s (found %s)", signature.Parameter(0), body_t)
				errors = append(errors, *p.srcmap.SyntaxError(expr.arg, msg))
			}
			// Check right parameter type
			if !body_t.SubtypeOf(signature.Parameter(1)) {
				msg := fmt.Sprintf("expected type %s (found %s)", signature.Parameter(1), body_t)
				errors = append(errors, *p.srcmap.SyntaxError(expr.arg, msg))
			}
		} else if !binding.HasArity(2) {
			msg := "incorrect number of arguments (expected 2)"
			errors = append(errors, *p.srcmap.SyntaxError(expr, msg))
		} else {
			msg := "ambiguous reduction"
			errors = append(errors, *p.srcmap.SyntaxError(expr, msg))
		}
		// Error check
		if len(errors) > 0 {
			return nil, errors
		}
		// Lock in signature
		expr.Finalise(signature)
	}
	//
	return body_t, nil
}

func (p *typeChecker) typeCheckVariableInModule(expr *VariableAccess) (Type, []SyntaxError) {
	// Check what we've got.
	if !expr.IsResolved() {
		//
	} else if binding, ok := expr.Binding().(*ColumnBinding); ok {
		return binding.dataType, nil
	} else if binding, ok := expr.Binding().(*ConstantBinding); ok {
		// Constant
		return p.typeCheckExpressionInModule(binding.value)
	} else if binding, ok := expr.Binding().(*LocalVariableBinding); ok {
		// Parameter, for or let variable
		return binding.datatype, nil
	}
	// NOTE: we don't return an error here, since this case would have already
	// been caught by the resolver and we don't want to double up on errors.
	return nil, nil
}
