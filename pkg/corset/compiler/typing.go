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
	"reflect"

	"github.com/consensys/go-corset/pkg/corset/ast"
	"github.com/consensys/go-corset/pkg/util/source"
)

// SyntaxError defines the kind of errors that can be reported by this compiler.
// Syntax errors are always associated with some line in one of the original
// source files.  For simplicity, we reuse existing notion of syntax error from
// the S-Expression library.
type SyntaxError = source.SyntaxError

// TypeCheckCircuit performs a type checking pass over the circuit to ensure
// types are used correctly.  Additionally, this resolves some ambiguities
// arising from the possibility of overloading function calls, etc.
func TypeCheckCircuit(srcmap *source.Maps[ast.Node],
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
	srcmap *source.Maps[ast.Node]
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
	case *ast.DefSorted:
		errors = p.typeCheckDefSorted(d)
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
		_, errs := p.typeCheckExpressionInModule(ast.INT_TYPE, c.ConstBinding.Value)
		// Accumulate errors
		errors = append(errors, errs...)
	}
	//
	return errors
}

// typeCheck a "defconstraint" declaration.
func (p *typeChecker) typeCheckDefConstraint(decl *ast.DefConstraint) []SyntaxError {
	// FIXME: eventually, the guard should be a BOOLEAN_TYPE in order to
	// force a suitable interpetation.
	//
	// typeCheck (optional) guard
	_, guard_errors := p.typeCheckOptionalExpressionInModule(ast.INT_TYPE, decl.Guard)
	// typeCheck constraint body
	_, constraint_errors := p.typeCheckExpressionInModule(ast.BOOLEAN_TYPE, decl.Constraint)
	// Combine errors
	return append(constraint_errors, guard_errors...)
}

// ast.Type check the body of a function.
func (p *typeChecker) typeCheckDefFunInModule(decl *ast.DefFun) []SyntaxError {
	// Resolve property body
	_, errors := p.typeCheckExpressionInModule(nil, decl.Body())
	// FIXME: type check return?
	// Done
	return errors
}

// typeCheck a "deflookup" declaration.
//
//nolint:staticcheck
func (p *typeChecker) typeCheckDefLookup(decl *ast.DefLookup) []SyntaxError {
	// typeCheck source expressions
	_, source_errs := p.typeCheckExpressionsInModule(ast.INT_TYPE, decl.Sources)
	_, target_errs := p.typeCheckExpressionsInModule(ast.INT_TYPE, decl.Targets)
	// Combine errors
	return append(source_errs, target_errs...)
}

// typeCheck a "definrange" declaration.
func (p *typeChecker) typeCheckDefInRange(decl *ast.DefInRange) []SyntaxError {
	// typeCheck constraint body
	_, errors := p.typeCheckExpressionInModule(ast.INT_TYPE, decl.Expr)
	// Done
	return errors
}

// typeCheck a "defperspective" declaration.
func (p *typeChecker) typeCheckDefPerspective(decl *ast.DefPerspective) []SyntaxError {
	// FIXME: eventually, the selector should be a BOOLEAN_TYPE in order to
	// force a suitable interpetation.
	//
	// typeCheck selector expression
	_, errors := p.typeCheckExpressionInModule(ast.INT_TYPE, decl.Selector)
	// Combine errors
	return errors
}

// typeCheck a "defproperty" declaration.
func (p *typeChecker) typeCheckDefProperty(decl *ast.DefProperty) []SyntaxError {
	// type check constraint body
	_, errors := p.typeCheckExpressionInModule(ast.BOOLEAN_TYPE, decl.Assertion)
	// Done
	return errors
}

// typeCheck a "defproperty" declaration.
func (p *typeChecker) typeCheckDefSorted(decl *ast.DefSorted) []SyntaxError {
	var errors []SyntaxError
	//
	if decl.Selector.HasValue() {
		// FIXME: eventually, the selector should be a BOOLEAN_TYPE in order to
		// force a suitable interpetation.
		_, errors = p.typeCheckExpressionInModule(ast.INT_TYPE, decl.Selector.Unwrap())
	}
	//
	return errors
}

// typeCheck an optional expression in a given context.  That is an expression
// which maybe nil (i.e. doesn't exist).  In such case, nil is returned (i.e.
// without any errors).
func (p *typeChecker) typeCheckOptionalExpressionInModule(expected ast.Type, expr ast.Expr) (ast.Type, []SyntaxError) {
	//
	if expr != nil {
		return p.typeCheckExpressionInModule(expected, expr)
	}
	//
	return nil, nil
}

// typeCheck a sequence of zero or more expressions enclosed in a given module.
// All expressions are expected to be non-voidable (see below for more on
// voidability).
func (p *typeChecker) typeCheckExpressionsInModule(expected ast.Type, exprs []ast.Expr) ([]ast.Type, []SyntaxError) {
	errors := []SyntaxError{}
	types := make([]ast.Type, len(exprs))
	// Iterate each expression in turn
	for i, e := range exprs {
		if e == nil {
			continue
		}
		//
		var errs []SyntaxError
		types[i], errs = p.typeCheckExpressionInModule(expected, e)
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
func (p *typeChecker) typeCheckExpressionInModule(expected ast.Type, expr ast.Expr) (ast.Type, []SyntaxError) {
	var (
		result ast.Type
		errors []SyntaxError
	)
	//
	switch e := expr.(type) {
	case *ast.ArrayAccess:
		result, errors = p.typeCheckArrayAccessInModule(e)
	case *ast.Add:
		_, errors = p.typeCheckExpressionsInModule(ast.INT_TYPE, e.Args)
		result = ast.INT_TYPE
	case *ast.Cast:
		result, errors = p.typeCheckExpressionInModule(e.Type, e.Arg)
	case *ast.Constant:
		nbits := e.Val.BitLen()
		result = ast.NewUintType(uint(nbits))
	case *ast.Debug:
		result, errors = p.typeCheckExpressionInModule(nil, e.Arg)
	case *ast.Equals:
		_, errs1 := p.typeCheckExpressionInModule(ast.INT_TYPE, e.Lhs)
		_, errs2 := p.typeCheckExpressionInModule(ast.INT_TYPE, e.Rhs)
		// Done
		result, errors = ast.BOOLEAN_TYPE, append(errs1, errs2...)
	case *ast.Exp:
		_, errs1 := p.typeCheckExpressionInModule(ast.INT_TYPE, e.Arg)
		_, errs2 := p.typeCheckExpressionInModule(ast.INT_TYPE, e.Pow)
		// Done
		result, errors = ast.INT_TYPE, append(errs1, errs2...)
	case *ast.For:
		// TODO: update environment with type of index variable.
		result, errors = p.typeCheckExpressionInModule(nil, e.Body)
	case *ast.If:
		result, errors = p.typeCheckIfInModule(e)
	case *ast.List:
		types, errs := p.typeCheckExpressionsInModule(nil, e.Args)
		result, errors = ast.LeastUpperBound(types...), errs
	case *ast.Mul:
		_, errors = p.typeCheckExpressionsInModule(ast.INT_TYPE, e.Args)
		result = ast.INT_TYPE
	case *ast.Normalise:
		_, errors = p.typeCheckExpressionInModule(ast.INT_TYPE, e.Arg)
		// Normalise guaranteed to return either 0 or 1.
		result = ast.NewUintType(1)
	case *ast.Shift:
		lhs_t, arg_errs := p.typeCheckExpressionInModule(ast.INT_TYPE, e.Arg)
		_, shf_errs := p.typeCheckExpressionInModule(ast.INT_TYPE, e.Shift)
		// combine errors
		result, errors = lhs_t, append(arg_errs, shf_errs...)
	case *ast.Sub:
		_, errors = p.typeCheckExpressionsInModule(ast.INT_TYPE, e.Args)
		result = ast.INT_TYPE
	case *ast.VariableAccess:
		result, errors = p.typeCheckVariableInModule(e)
	default:
		msg := fmt.Sprintf("unknown expression encountered during typing (%s)", reflect.TypeOf(expr).String())
		return nil, p.srcmap.SyntaxErrors(expr, msg)
	}
	// Error check
	if result == nil && len(errors) == 0 {
		return nil, p.srcmap.SyntaxErrors(expr, "internal failure")
	} else if expected != nil && result != nil && !result.SubtypeOf(expected) {
		msg := fmt.Sprintf("expected %s, found %s", expected.String(), result.String())
		return nil, p.srcmap.SyntaxErrors(expr, msg)
	}
	//
	return result, errors
}

// ast.Type check an array access expression.  The main thing is to check that the
// column being accessed was originally defined as an array column.
func (p *typeChecker) typeCheckArrayAccessInModule(expr *ast.ArrayAccess) (ast.Type, []SyntaxError) {
	// ast.Type check index expression
	_, errs := p.typeCheckExpressionInModule(ast.INT_TYPE, expr.Arg)
	// NOTE: following cast safe because resolver already checked them.
	if binding, ok := expr.Binding().(*ast.ColumnBinding); !ok || !expr.IsResolved() {
		// NOTE: we don't return an error here, since this case would have already
		// been caught by the resolver and we don't want to double up on errors.
		return nil, nil
	} else if arr_t, ok := binding.DataType.(*ast.ArrayType); !ok {
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
	// Check condition
	_, errors := p.typeCheckExpressionInModule(ast.BOOLEAN_TYPE, expr.Condition)
	// Check true branch
	res_t, errs := p.typeCheckExpressionInModule(nil, expr.TrueBranch)
	errors = append(errors, errs...)
	//
	if expr.FalseBranch != nil {
		rhs_t, errs2 := p.typeCheckExpressionInModule(nil, expr.FalseBranch)
		errors = append(errors, errs2...)
		// Join result types
		res_t = ast.LeastUpperBound(res_t, rhs_t)
	}
	// sanity check
	if len(errors) > 0 {
		return nil, errors
	}
	// success
	return res_t, nil
}

func (p *typeChecker) typeCheckVariableInModule(expr *ast.VariableAccess) (ast.Type, []SyntaxError) {
	// Check what we've got.
	if !expr.IsResolved() {
		//
	} else if binding, ok := expr.Binding().(*ast.ColumnBinding); ok {
		return binding.DataType, nil
	} else if binding, ok := expr.Binding().(*ast.ConstantBinding); ok {
		// Constant
		return p.typeCheckExpressionInModule(binding.DataType, binding.Value)
	} else if binding, ok := expr.Binding().(*ast.LocalVariableBinding); ok {
		// Parameter, for or let variable
		return binding.DataType, nil
	}
	// NOTE: we don't return an error here, since this case would have already
	// been caught by the resolver and we don't want to double up on errors.
	return nil, nil
}
