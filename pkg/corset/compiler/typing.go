// Copyright Consensys Software Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License"); you may not use this file except in compliance with
// the License. You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
// an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
// specific language governing permissions and limitations under the License.
//
// SPDX-License-Identifier: Apache-2.0
package compiler

import (
	"fmt"

	"github.com/consensys/go-corset/pkg/corset/ast"
	"github.com/consensys/go-corset/pkg/util/sexp"
)

// SyntaxError defines the kind of errors that can be reported by this compiler.
// Syntax errors are always associated with some line in one of the original
// source files.  For simplicity, we reuse existing notion of syntax error from
// the S-Expression library.
type SyntaxError = sexp.SyntaxError

// TypeCheckCircuit performs a type checking pass over the circuit to ensure
// types are used correctly.  Additionally, this resolves some ambiguities
// arising from the possibility of overloading function calls, etc.
func TypeCheckCircuit(srcmap *sexp.SourceMaps[ast.Node],
	circuit *ast.Circuit) []SyntaxError {
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
	srcmap *sexp.SourceMaps[ast.Node]
}

// typeCheck all assignment or constraint declarations in the circuit.
func (p *typeChecker) typeCheckDeclarations(circuit *ast.Circuit) []SyntaxError {
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
func (p *typeChecker) typeCheckDeclarationsInModule(decls []ast.Declaration) []SyntaxError {
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
func (p *typeChecker) typeCheckDeclaration(decl ast.Declaration) []SyntaxError {
	var errors []SyntaxError
	//
	switch d := decl.(type) {
	case *ast.DefAliases:
		// ignore
	case *ast.DefColumns:
		// ignore
	case *ast.DefComputed:
		// ignore (for now)
	case *ast.DefConst:
		errors = p.typeCheckDefConstInModule(d)
	case *ast.DefConstraint:
		errors = p.typeCheckDefConstraint(d)
	case *ast.DefFun:
		errors = p.typeCheckDefFunInModule(d)
	case *ast.DefInRange:
		errors = p.typeCheckDefInRange(d)
	case *ast.DefInterleaved:
		// ignore
	case *ast.DefLookup:
		errors = p.typeCheckDefLookup(d)
	case *ast.DefPermutation:
		// ignore
	case *ast.DefPerspective:
		errors = p.typeCheckDefPerspective(d)
	case *ast.DefProperty:
		errors = p.typeCheckDefProperty(d)
	default:
		// Error handling
		panic("unknown declaration")
	}
	//
	return errors
}

// ast.Type check one or more constant definitions within a given module.
func (p *typeChecker) typeCheckDefConstInModule(decl *ast.DefConst) []SyntaxError {
	var errors []SyntaxError
	//
	for _, c := range decl.Constants {
		// Resolve constant body
		_, errs := p.typeCheckExpressionInModule(c.ConstBinding.Value)
		// Accumulate errors
		errors = append(errors, errs...)
	}
	//
	return errors
}

// typeCheck a "defconstraint" declaration.
func (p *typeChecker) typeCheckDefConstraint(decl *ast.DefConstraint) []SyntaxError {
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

// ast.Type check the body of a function.
func (p *typeChecker) typeCheckDefFunInModule(decl *ast.DefFun) []SyntaxError {
	// Resolve property body
	_, errors := p.typeCheckExpressionInModule(decl.Body())
	// FIXME: type check return?
	// Done
	return errors
}

// typeCheck a "deflookup" declaration.
//
//nolint:staticcheck
func (p *typeChecker) typeCheckDefLookup(decl *ast.DefLookup) []SyntaxError {
	// typeCheck source expressions
	_, source_errs := p.typeCheckExpressionsInModule(decl.Sources)
	_, target_errs := p.typeCheckExpressionsInModule(decl.Targets)
	// Combine errors
	return append(source_errs, target_errs...)
}

// typeCheck a "definrange" declaration.
func (p *typeChecker) typeCheckDefInRange(decl *ast.DefInRange) []SyntaxError {
	// typeCheck constraint body
	_, errors := p.typeCheckExpressionInModule(decl.Expr)
	// Done
	return errors
}

// typeCheck a "defperspective" declaration.
func (p *typeChecker) typeCheckDefPerspective(decl *ast.DefPerspective) []SyntaxError {
	// typeCheck selector expression
	_, errors := p.typeCheckExpressionInModule(decl.Selector)
	// Combine errors
	return errors
}

// typeCheck a "defproperty" declaration.
func (p *typeChecker) typeCheckDefProperty(decl *ast.DefProperty) []SyntaxError {
	// type check constraint body
	_, errors := p.typeCheckExpressionInModule(decl.Assertion)
	// Done
	return errors
}

// typeCheck an optional expression in a given context.  That is an expression
// which maybe nil (i.e. doesn't exist).  In such case, nil is returned (i.e.
// without any errors).
func (p *typeChecker) typeCheckOptionalExpressionInModule(expr ast.Expr) (ast.Type, []SyntaxError) {
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
func (p *typeChecker) typeCheckExpressionsInModule(exprs []ast.Expr) ([]ast.Type, []SyntaxError) {
	errors := []SyntaxError{}
	types := make([]ast.Type, len(exprs))
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
func (p *typeChecker) typeCheckExpressionInModule(expr ast.Expr) (ast.Type, []SyntaxError) {
	switch e := expr.(type) {
	case *ast.ArrayAccess:
		return p.typeCheckArrayAccessInModule(e)
	case *ast.Add:
		types, errs := p.typeCheckExpressionsInModule(e.Args)
		return ast.LeastUpperBoundAll(types), errs
	case *ast.Constant:
		nbits := e.Val.BitLen()
		return ast.NewUintType(uint(nbits)), nil
	case *ast.Debug:
		return p.typeCheckExpressionInModule(e.Arg)
	case *ast.Exp:
		arg_t, errs1 := p.typeCheckExpressionInModule(e.Arg)
		_, errs2 := p.typeCheckExpressionInModule(e.Pow)
		// Done
		return arg_t, append(errs1, errs2...)
	case *ast.For:
		// TODO: update environment with type of index variable.
		return p.typeCheckExpressionInModule(e.Body)
	case *ast.If:
		return p.typeCheckIfInModule(e)
	case *ast.Invoke:
		return p.typeCheckInvokeInModule(e)
	case *ast.Let:
		return p.typeCheckLetInModule(e)
	case *ast.List:
		types, errs := p.typeCheckExpressionsInModule(e.Args)
		return ast.LeastUpperBoundAll(types), errs
	case *ast.Mul:
		types, errs := p.typeCheckExpressionsInModule(e.Args)
		return ast.GreatestLowerBoundAll(types), errs
	case *ast.Normalise:
		_, errs := p.typeCheckExpressionInModule(e.Arg)
		// Normalise guaranteed to return either 0 or 1.
		return ast.NewUintType(1), errs
	case *ast.Reduce:
		return p.typeCheckReduceInModule(e)
	case *ast.Shift:
		arg_t, arg_errs := p.typeCheckExpressionInModule(e.Arg)
		_, shf_errs := p.typeCheckExpressionInModule(e.Shift)
		// combine errors
		return arg_t, append(arg_errs, shf_errs...)
	case *ast.Sub:
		types, errs := p.typeCheckExpressionsInModule(e.Args)
		return ast.LeastUpperBoundAll(types), errs
	case *ast.VariableAccess:
		return p.typeCheckVariableInModule(e)
	default:
		return nil, p.srcmap.SyntaxErrors(expr, "unknown expression encountered during translation")
	}
}

// ast.Type check an array access expression.  The main thing is to check that the
// column being accessed was originally defined as an array column.
func (p *typeChecker) typeCheckArrayAccessInModule(expr *ast.ArrayAccess) (ast.Type, []SyntaxError) {
	// ast.Type check index expression
	_, errs := p.typeCheckExpressionInModule(expr.Arg)
	// NOTE: following cast safe because resolver already checked them.
	binding := expr.Binding().(*ast.ColumnBinding)
	if arr_t, ok := binding.DataType.(*ast.ArrayType); !ok {
		return nil, append(errs, *p.srcmap.SyntaxError(expr, "expected array column"))
	} else {
		return arr_t.Element(), errs
	}
}

// ast.Type an if condition contained within some expression which, in turn, is
// contained within some module.  An important step occurrs here where, based on
// the semantics of the condition, this is inferred as an "if-zero" or an
// "if-notzero".
func (p *typeChecker) typeCheckIfInModule(expr *ast.If) (ast.Type, []SyntaxError) {
	types, errs := p.typeCheckExpressionsInModule([]ast.Expr{expr.Condition, expr.TrueBranch, expr.FalseBranch})
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
	return ast.GreatestLowerBoundAll(types[1:]), errs
}

func (p *typeChecker) typeCheckInvokeInModule(expr *ast.Invoke) (ast.Type, []SyntaxError) {
	arity := uint(len(expr.Args))
	//
	if binding, ok := expr.Name.Binding().(ast.FunctionBinding); !ok {
		// We don't return an error here, since one would already have been
		// generated during resolution.
		return nil, nil
	} else if argTypes, errors := p.typeCheckExpressionsInModule(expr.Args); len(errors) > 0 {
		return nil, errors
	} else if argTypes == nil {
		// An upstream expression could not because of a resolution error.
		return nil, nil
	} else if signature := binding.Select(arity); signature != nil {
		// Check arguments are accepted, based on their type.
		for i := 0; i < len(argTypes); i++ {
			expected := signature.Parameter(uint(i))
			actual := argTypes[i]
			// subtype check
			if actual != nil && !actual.SubtypeOf(expected) {
				msg := fmt.Sprintf("expected type %s (found %s)", expected, actual)
				errors = append(errors, *p.srcmap.SyntaxError(expr.Args[i], msg))
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
		body := signature.Apply(expr.Args, nil)
		// Dig out the type
		return p.typeCheckExpressionInModule(body)
	}
	// ambiguous invocation
	return nil, p.srcmap.SyntaxErrors(expr.Name, "ambiguous invocation")
}

func (p *typeChecker) typeCheckLetInModule(expr *ast.Let) (ast.Type, []SyntaxError) {
	// NOTE: there is a limitation here since we are using the type of the
	// assigned expressions.  It would be nice to retain this, but it would
	// require a more flexible notion of environment than we currently have.
	if types, arg_errors := p.typeCheckExpressionsInModule(expr.Args); types != nil {
		// Update type for let-bound variables.
		for i := range expr.Vars {
			if types[i] != nil {
				expr.Vars[i].DataType = types[i]
			}
		}
		// ast.Type check body
		body_t, body_errors := p.typeCheckExpressionInModule(expr.Body)
		//
		return body_t, append(arg_errors, body_errors...)
	} else {
		return nil, arg_errors
	}
}

func (p *typeChecker) typeCheckReduceInModule(expr *ast.Reduce) (ast.Type, []SyntaxError) {
	var signature *ast.FunctionSignature
	// ast.Type check body of reduction
	body_t, errors := p.typeCheckExpressionInModule(expr.Arg)
	// Following safe as resolver checked this already.
	if binding, ok := expr.Name.Binding().(ast.FunctionBinding); ok && body_t != nil {
		//
		if signature = binding.Select(2); signature != nil {
			// Check left parameter type
			if !body_t.SubtypeOf(signature.Parameter(0)) {
				msg := fmt.Sprintf("expected type %s (found %s)", signature.Parameter(0), body_t)
				errors = append(errors, *p.srcmap.SyntaxError(expr.Arg, msg))
			}
			// Check right parameter type
			if !body_t.SubtypeOf(signature.Parameter(1)) {
				msg := fmt.Sprintf("expected type %s (found %s)", signature.Parameter(1), body_t)
				errors = append(errors, *p.srcmap.SyntaxError(expr.Arg, msg))
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

func (p *typeChecker) typeCheckVariableInModule(expr *ast.VariableAccess) (ast.Type, []SyntaxError) {
	// Check what we've got.
	if !expr.IsResolved() {
		//
	} else if binding, ok := expr.Binding().(*ast.ColumnBinding); ok {
		return binding.DataType, nil
	} else if binding, ok := expr.Binding().(*ast.ConstantBinding); ok {
		// Constant
		return p.typeCheckExpressionInModule(binding.Value)
	} else if binding, ok := expr.Binding().(*ast.LocalVariableBinding); ok {
		// Parameter, for or let variable
		return binding.DataType, nil
	}
	// NOTE: we don't return an error here, since this case would have already
	// been caught by the resolver and we don't want to double up on errors.
	return nil, nil
}
