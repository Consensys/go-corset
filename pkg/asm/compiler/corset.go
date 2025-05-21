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
	"math/big"
	"strings"

	corset_ast "github.com/consensys/go-corset/pkg/corset/ast"
	"github.com/consensys/go-corset/pkg/util"
	"github.com/consensys/go-corset/pkg/util/source/sexp"
)

// Compile2Circuit compiles a given set of micro functions into a corset
// circuit.
func Compile2Circuit(functions []MicroFunction) corset_ast.Circuit {
	var (
		circuit corset_ast.Circuit
		// Construct compiler
		compiler = NewCompiler[util.Path, CorsetExpr, CorsetModule]()
	)
	// Compile all functions
	compiler.Compile(functions...)
	// Construct modules
	for _, m := range compiler.Modules() {
		circuit.Modules = append(circuit.Modules, *m.module)
	}
	//
	return circuit
}

// CorsetModule provides a wrapper around a corset-level module declaration.
type CorsetModule struct {
	module *corset_ast.Module
}

// Initialise this module
func (p CorsetModule) Initialise(name string) CorsetModule {
	p = CorsetModule{&corset_ast.Module{}}
	p.module.Name = name

	return p
}

// NewColumn constructs a new column of the given name and bitwidth within
// this module.
func (p CorsetModule) NewColumn(name string, bitwidth uint, internal bool) util.Path {
	path := util.NewAbsolutePath(p.module.Name)
	name_path := *path.Extend(name)
	datatype := corset_ast.NewUintType(bitwidth)
	columns := []*corset_ast.DefColumn{
		corset_ast.NewDefColumn(path, name_path, datatype, true, 1, false, internal, "hex"),
	}
	//
	p.module.Add(corset_ast.NewDefColumns(columns))
	//
	return name_path
}

// NewConstraint constructs a new vanishing constraint with the given name
// within this module.
func (p CorsetModule) NewConstraint(name string, domain util.Option[int], constraint CorsetExpr) {
	p.module.Add(corset_ast.NewDefConstraint(name, domain, nil, nil, constraint.expr))
}

// NewLookup constructs a new lookup constraint
func (p CorsetModule) NewLookup(name string, from []CorsetExpr, to []CorsetExpr) {
	sources := unwrapCorsetExprs(from...)
	targets := unwrapCorsetExprs(to...)
	p.module.Add(corset_ast.NewDefLookup(name, sources, targets))
}

// String returns an appropriately formatted representation of the module.
func (p CorsetModule) String() string {
	var builder strings.Builder
	//
	formatter := sexp.NewFormatter(120)
	formatter.Add(&sexp.SFormatter{Head: "if", Priority: 0})
	formatter.Add(&sexp.SFormatter{Head: "ifnot", Priority: 0})
	formatter.Add(&sexp.LFormatter{Head: "begin", Priority: 1})
	formatter.Add(&sexp.LFormatter{Head: "∧", Priority: 1})
	formatter.Add(&sexp.LFormatter{Head: "∨", Priority: 1})
	formatter.Add(&sexp.LFormatter{Head: "+", Priority: 2})
	formatter.Add(&sexp.LFormatter{Head: "*", Priority: 3})
	//
	builder.WriteString(fmt.Sprintf("(module %s)\n", p.module.Name))
	//
	for _, decl := range p.module.Declarations {
		text := formatter.Format(decl.Lisp())
		builder.WriteString(text)
	}
	//
	return builder.String()
}

// CorsetExpr is a wrapper around a corset expression which provides the
// necessary interface.
type CorsetExpr struct {
	expr corset_ast.Expr
}

// Add constructs a sum between this expression and zero or more
func (p CorsetExpr) Add(exprs ...CorsetExpr) CorsetExpr {
	args := unwrapCorsetSplitExprs(p, exprs...)
	return CorsetExpr{&corset_ast.Add{Args: args}}
}

// And constructs a conjunction between this expression and zero or more
// expressions.
func (p CorsetExpr) And(exprs ...CorsetExpr) CorsetExpr {
	args := unwrapCorsetSplitExprs(p, exprs...)
	return CorsetExpr{&corset_ast.Connective{Sign: false, Args: args}}
}

// Equals constructs an equality between two expressions.
func (p CorsetExpr) Equals(rhs CorsetExpr) CorsetExpr {
	return CorsetExpr{&corset_ast.Equation{
		Kind: corset_ast.EQUALS,
		Lhs:  p.expr,
		Rhs:  rhs.expr,
	}}
}

// Then constructs an implication between two expressions.
func (p CorsetExpr) Then(trueBranch CorsetExpr) CorsetExpr {
	if trueBranch.expr == nil {
		panic("got here")
	}

	return CorsetExpr{&corset_ast.If{
		Condition:   p.expr,
		TrueBranch:  trueBranch.expr,
		FalseBranch: nil,
	}}
}

// ThenElse constructs an if-then-else expression with this expression
// acting as the condition.
func (p CorsetExpr) ThenElse(trueBranch CorsetExpr, falseBranch CorsetExpr) CorsetExpr {
	if trueBranch.expr == nil || falseBranch.expr == nil {
		panic("got here")
	}

	return CorsetExpr{&corset_ast.If{
		Condition:   p.expr,
		TrueBranch:  trueBranch.expr,
		FalseBranch: falseBranch.expr,
	}}
}

// Multiply constructs a product between this expression and zero or more
// expressions.
func (p CorsetExpr) Multiply(exprs ...CorsetExpr) CorsetExpr {
	args := unwrapCorsetSplitExprs(p, exprs...)
	return CorsetExpr{&corset_ast.Mul{Args: args}}
}

// NotEquals constructs a non-equality between two expressions.
func (p CorsetExpr) NotEquals(rhs CorsetExpr) CorsetExpr {
	return CorsetExpr{&corset_ast.Equation{
		Kind: corset_ast.NOT_EQUALS,
		Lhs:  p.expr,
		Rhs:  rhs.expr,
	}}
}

// BigInt constructs a constant expression from a big integer.
func (p CorsetExpr) BigInt(number big.Int) CorsetExpr {
	return CorsetExpr{
		&corset_ast.Constant{Val: number},
	}
}

// Or constructs a disjunction between this expression and zero or more
// expressions.
func (p CorsetExpr) Or(exprs ...CorsetExpr) CorsetExpr {
	args := unwrapCorsetSplitExprs(p, exprs...)
	return CorsetExpr{&corset_ast.Connective{Sign: true, Args: args}}
}

// Variable constructs a variable with a given shift.
func (p CorsetExpr) Variable(path util.Path, shift int) CorsetExpr {
	var v corset_ast.Expr = corset_ast.NewVariableAccess(path, corset_ast.NON_FUNCTION, nil)
	//
	if shift != 0 {
		number := big.NewInt(int64(shift))
		c := &corset_ast.Constant{Val: *number}
		v = &corset_ast.Shift{
			Arg:   v,
			Shift: c,
		}
	}
	//
	return CorsetExpr{v}
}

func unwrapCorsetSplitExprs(head CorsetExpr, tail ...CorsetExpr) []corset_ast.Expr {
	cexprs := make([]corset_ast.Expr, len(tail)+1)
	//
	cexprs[0] = head.expr
	//
	for i, e := range tail {
		cexprs[i+1] = e.expr
	}
	//
	return cexprs
}

func unwrapCorsetExprs(exprs ...CorsetExpr) []corset_ast.Expr {
	cexprs := make([]corset_ast.Expr, len(exprs))
	//
	for i, e := range exprs {
		cexprs[i] = e.expr
	}
	//
	return cexprs
}
