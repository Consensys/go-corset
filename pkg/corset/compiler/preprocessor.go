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
	"math/big"

	"github.com/consensys/go-corset/pkg/corset/ast"
	"github.com/consensys/go-corset/pkg/util"
	"github.com/consensys/go-corset/pkg/util/sexp"
)

// PreprocessCircuit performs preprocessing prior to final translation.
// Specifically, it expands all invocations, reductions and for loops.  Thus,
// final translation is greatly simplified after this step.
func PreprocessCircuit(debug bool, srcmap *sexp.SourceMaps[ast.Node],
	circuit *ast.Circuit) []SyntaxError {
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
	srcmap *sexp.SourceMaps[ast.Node]
}

// preprocess all assignment or constraint declarations in the circuit.
func (p *preprocessor) preprocessDeclarations(circuit *ast.Circuit) []SyntaxError {
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
func (p *preprocessor) preprocessDeclarationsInModule(decls []ast.Declaration) []SyntaxError {
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
func (p *preprocessor) preprocessDeclaration(decl ast.Declaration) []SyntaxError {
	var errors []SyntaxError
	//
	switch d := decl.(type) {
	case *ast.DefAliases:
		// ignore
	case *ast.DefColumns:
		// ignore
	case *ast.DefComputed:
		// ignore
	case *ast.DefConst:
		// ignore
	case *ast.DefConstraint:
		errors = p.preprocessDefConstraint(d)
	case *ast.DefFun:
		// ignore
	case *ast.DefInRange:
		errors = p.preprocessDefInRange(d)
	case *ast.DefInterleaved:
		// ignore
	case *ast.DefLookup:
		errors = p.preprocessDefLookup(d)
	case *ast.DefPermutation:
		// ignore
	case *ast.DefPerspective:
		errors = p.preprocessDefPerspective(d)
	case *ast.DefProperty:
		errors = p.preprocessDefProperty(d)
	case *ast.DefSorted:
		errors = p.preprocessDefSorted(d)
	default:
		// Error handling
		panic("unknown declaration")
	}
	//
	return errors
}

// preprocess a "defconstraint" declaration.
func (p *preprocessor) preprocessDefConstraint(decl *ast.DefConstraint) []SyntaxError {
	var (
		constraint_errors []SyntaxError
		guard_errors      []SyntaxError
	)
	// preprocess constraint body
	decl.Constraint, constraint_errors = p.preprocessExpressionInModule(decl.Constraint)
	// preprocess (optional) guard
	decl.Guard, guard_errors = p.preprocessOptionalExpressionInModule(decl.Guard)
	// sanity check
	if decl.Constraint == nil {
		// this case is possible when the constraint expression consists
		// entirely of debug constraints, and debug mode is not enabled.
		decl.Constraint = &ast.List{Args: nil}
	}
	// Combine errors
	return append(constraint_errors, guard_errors...)
}

// preprocess a "deflookup" declaration.
//
//nolint:staticcheck
func (p *preprocessor) preprocessDefLookup(decl *ast.DefLookup) []SyntaxError {
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
func (p *preprocessor) preprocessDefInRange(decl *ast.DefInRange) []SyntaxError {
	var errors []SyntaxError
	// preprocess constraint body
	decl.Expr, errors = p.preprocessExpressionInModule(decl.Expr)
	// Done
	return errors
}

// preprocess a "defperspective" declaration.
func (p *preprocessor) preprocessDefPerspective(decl *ast.DefPerspective) []SyntaxError {
	var errors []SyntaxError
	// preprocess selector expression
	decl.Selector, errors = p.preprocessExpressionInModule(decl.Selector)
	// Combine errors
	return errors
}

// preprocess a "defproperty" declaration.
func (p *preprocessor) preprocessDefProperty(decl *ast.DefProperty) []SyntaxError {
	var errors []SyntaxError
	// preprocess constraint body
	decl.Assertion, errors = p.preprocessExpressionInModule(decl.Assertion)
	// Done
	return errors
}

func (p *preprocessor) preprocessDefSorted(decl *ast.DefSorted) []SyntaxError {
	if decl.Selector.HasValue() {
		selector, errors := p.preprocessExpressionInModule(decl.Selector.Unwrap())
		//
		decl.Selector = util.Some(selector)
		//
		return errors
	}
	//
	return nil
}

// preprocess an optional expression in a given context.  That is an expression
// which maybe nil (i.e. doesn't exist).  In such case, nil is returned (i.e.
// without any errors).
func (p *preprocessor) preprocessOptionalExpressionInModule(expr ast.Expr) (ast.Expr, []SyntaxError) {
	//
	if expr != nil {
		return p.preprocessExpressionInModule(expr)
	}

	return nil, nil
}

// preprocess a sequence of zero or more expressions enclosed in a given module.
// All expressions are expected to be non-voidable (see below for more on
// voidability).
func (p *preprocessor) preprocessExpressionsInModule(exprs []ast.Expr) ([]ast.Expr, []SyntaxError) {
	//
	errors := []SyntaxError{}
	hirExprs := make([]ast.Expr, len(exprs))
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
func (p *preprocessor) preprocessVoidableExpressionsInModule(exprs []ast.Expr) ([]ast.Expr, []SyntaxError) {
	//
	errors := []SyntaxError{}
	hirExprs := make([]ast.Expr, len(exprs))
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
	nHirExprs := make([]ast.Expr, len(exprs)-nils)
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
func (p *preprocessor) preprocessExpressionInModule(expr ast.Expr) (ast.Expr, []SyntaxError) {
	var (
		nexpr  ast.Expr
		errors []SyntaxError
	)
	//
	switch e := expr.(type) {
	case *ast.ArrayAccess:
		arg, errs := p.preprocessExpressionInModule(e.Arg)
		nexpr, errors = &ast.ArrayAccess{Name: e.Name, Arg: arg, ArrayBinding: e.ArrayBinding}, errs
	case *ast.Add:
		args, errs := p.preprocessExpressionsInModule(e.Args)
		nexpr, errors = &ast.Add{Args: args}, errs
	case *ast.Cast:
		arg, errs := p.preprocessExpressionInModule(e.Arg)
		nexpr, errors = &ast.Cast{Arg: arg, BitWidth: e.BitWidth}, errs
	case *ast.Constant:
		return e, nil
	case *ast.Debug:
		if p.debug {
			return p.preprocessExpressionInModule(e.Arg)
		}
		// When debug is not enabled, return "void".
		return nil, nil
	case *ast.Exp:
		arg, errs1 := p.preprocessExpressionInModule(e.Arg)
		pow, errs2 := p.preprocessExpressionInModule(e.Pow)
		// Done
		nexpr, errors = &ast.Exp{Arg: arg, Pow: pow}, append(errs1, errs2...)
	case *ast.For:
		return p.preprocessForInModule(e)
	case *ast.If:
		args, errs := p.preprocessExpressionsInModule([]ast.Expr{e.Condition, e.TrueBranch, e.FalseBranch})
		// Construct appropriate if form
		nexpr, errors = &ast.If{Kind: e.Kind, Condition: args[0], TrueBranch: args[1], FalseBranch: args[2]}, errs
	case *ast.Invoke:
		return p.preprocessInvokeInModule(e)
	case *ast.Let:
		return p.preprocessLetInModule(e)
	case *ast.List:
		args, errs := p.preprocessVoidableExpressionsInModule(e.Args)
		nexpr, errors = &ast.List{Args: args}, errs
	case *ast.Mul:
		args, errs := p.preprocessExpressionsInModule(e.Args)
		nexpr, errors = &ast.Mul{Args: args}, errs
	case *ast.Normalise:
		arg, errs := p.preprocessExpressionInModule(e.Arg)
		nexpr, errors = &ast.Normalise{Arg: arg}, errs
	case *ast.Reduce:
		return p.preprocessReduceInModule(e)
	case *ast.Sub:
		args, errs := p.preprocessExpressionsInModule(e.Args)
		nexpr, errors = &ast.Sub{Args: args}, errs
	case *ast.Shift:
		arg, errs := p.preprocessExpressionInModule(e.Arg)
		nexpr, errors = &ast.Shift{Arg: arg, Shift: e.Shift}, errs
	case *ast.VariableAccess:
		return e, nil
	default:
		return nil, p.srcmap.SyntaxErrors(expr, "unknown expression encountered during translation")
	}
	// Copy over source information
	p.srcmap.Copy(expr, nexpr)
	// Done
	return nexpr, errors
}

func (p *preprocessor) preprocessForInModule(expr *ast.For) (ast.Expr, []SyntaxError) {
	var (
		errors  []SyntaxError
		mapping map[uint]ast.Expr = make(map[uint]ast.Expr)
	)
	// Determine range for index variable
	n := expr.End - expr.Start + 1
	args := make([]ast.Expr, n)
	// Expand body n times
	for i := uint(0); i < n; i++ {
		var errs []SyntaxError
		// Substitute through for i
		mapping[expr.Binding.Index] = &ast.Constant{Val: *big.NewInt(int64(i + expr.Start))}
		ith := ast.Substitute(expr.Body, mapping, p.srcmap)
		// preprocess subsituted expression
		args[i], errs = p.preprocessExpressionInModule(ith)
		errors = append(errors, errs...)
	}
	// Error check
	if len(errors) != 0 {
		return nil, errors
	}
	// Done
	return &ast.List{Args: args}, nil
}

func (p *preprocessor) preprocessLetInModule(expr *ast.Let) (ast.Expr, []SyntaxError) {
	var (
		mapping map[uint]ast.Expr = make(map[uint]ast.Expr)
		errors  []SyntaxError
		errs    []SyntaxError
	)
	// Construct variable mapping and preprocess
	for i, v := range expr.Vars {
		mapping[v.Index], errs = p.preprocessExpressionInModule(expr.Args[i])
		errors = append(errors, errs...)
	}
	// Apply substituteion
	body := ast.Substitute(expr.Body, mapping, p.srcmap)
	// Constinue preprocessing
	body, errs = p.preprocessExpressionInModule(body)
	// Done
	return body, append(errors, errs...)
}

func (p *preprocessor) preprocessInvokeInModule(expr *ast.Invoke) (ast.Expr, []SyntaxError) {
	if expr.Signature != nil {
		var (
			args   []ast.Expr = make([]ast.Expr, len(expr.Args))
			errors []SyntaxError
			errs   []SyntaxError
		)
		// Preprocess arguments prior to subsitution.
		for i, e := range expr.Args {
			args[i], errs = p.preprocessExpressionInModule(e)
			errors = append(errors, errs...)
		}
		// Substitute through body
		body := expr.Signature.Apply(args, p.srcmap)
		// Preprocess body
		body, errs = p.preprocessExpressionInModule(body)
		// Done
		return body, append(errors, errs...)
	}
	//
	return nil, p.srcmap.SyntaxErrors(expr, "unbound function")
}

func (p *preprocessor) preprocessReduceInModule(expr *ast.Reduce) (ast.Expr, []SyntaxError) {
	body, errors := p.preprocessExpressionInModule(expr.Arg)
	//
	if list, ok := body.(*ast.List); !ok {
		return nil, append(errors, *p.srcmap.SyntaxError(expr.Arg, "expected list"))
	} else if sig := expr.Signature; sig == nil {
		return nil, append(errors, *p.srcmap.SyntaxError(expr.Arg, "unbound function"))
	} else {
		reduction := list.Args[0]
		// Build reduction
		for i := 1; i < len(list.Args); i++ {
			reduction = sig.Apply([]ast.Expr{reduction, list.Args[i]}, p.srcmap)
		}
		// done
		return reduction, errors
	}
}
