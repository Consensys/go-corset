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
	"math"

	"github.com/consensys/go-corset/pkg/corset/ast"
	"github.com/consensys/go-corset/pkg/util"
	"github.com/consensys/go-corset/pkg/util/source/sexp"
)

// IntrinsicDefinition is a SymbolDefinition for an intrinsic (i.e. built-in)
// operation, such as "+", "-", etc.  These are needed for two reasons: firstly,
// so we can alias them; secondly, so they can be used in reductions.
type IntrinsicDefinition struct {
	// Name of the intrinsic (e.g. "+")
	name string
	// Minimum number of arguments this intrinsic can accept.
	min_arity uint
	// Maximum number of arguments this intrinsic can accept.
	max_arity uint
	// Construct an instance of this intrinsic for a given arity (i.e. number of
	// arguments).
	constructor func(uint) ast.Expr
}

var _ ast.FunctionBinding = &IntrinsicDefinition{}

// Name returns the name of the intrinsic being defined.
func (p *IntrinsicDefinition) Name() string {
	return p.name
}

// Path returns the qualified name (i.e. absolute path) of this symbol.  For
// example, "m1.X" for a column X defined in module m1.
func (p *IntrinsicDefinition) Path() *util.Path {
	path := util.NewAbsolutePath(p.name)
	return &path
}

// IsPure checks whether this pure (which intrinsics always are).
func (p *IntrinsicDefinition) IsPure() bool {
	return true
}

// IsNative checks whether this function binding is native (or not).
func (p *IntrinsicDefinition) IsNative() bool {
	return false
}

// IsFunction identifies whether or not the intrinsic being defined is a
// function.  At this time, all intrinsics are functions.
func (p *IntrinsicDefinition) IsFunction() bool {
	return true
}

// IsFinalised checks whether this binding has been finalised yet or not.
func (p *IntrinsicDefinition) IsFinalised() bool {
	return true
}

// Binding returns the binding associated with this intrinsic.
func (p *IntrinsicDefinition) Binding() ast.Binding {
	return p
}

// Lisp returns a lisp representation of this intrinsic.
func (p *IntrinsicDefinition) Lisp() sexp.SExp {
	panic("unreachable")
}

// HasArity checks whether this function accepts a given number of arguments (or
// not).
func (p *IntrinsicDefinition) HasArity(arity uint) bool {
	return arity >= p.min_arity && arity <= p.max_arity
}

// Select corresponding signature based on arity.  If no matching signature
// exists then this will return nil.
func (p *IntrinsicDefinition) Select(arity uint) *ast.FunctionSignature {
	// construct the body
	body := p.constructor(arity)
	types := make([]ast.Type, arity)
	//
	for i := 0; i < len(types); i++ {
		types[i] = ast.NewFieldType()
	}
	// Allow return type to be inferred.
	return ast.NewFunctionSignature(true, types, nil, body)
}

// Overload (a.k.a specialise) this function binding to incorporate another
// function signature.  This can fail for a few reasons: (1) some bindings
// (e.g. intrinsics) cannot be overloaded; (2) duplicate overloadings are
// not permitted; (3) combinding pure and impure overloadings is also not
// permitted.
func (p *IntrinsicDefinition) Overload(*ast.DefunBinding) (ast.FunctionBinding, bool) {
	// Easy case, as intrinsics cannot be overloaded.
	return nil, false
}

// ============================================================================
// Intrinsic Definitions
// ============================================================================

// INTRINSICS identifies all of the built-in functions used within the corset
// language, such as "+", "-", etc.  This is needed for two reasons: firstly, so
// we can alias them; secondly, so they can be used in reductions.
var INTRINSICS []IntrinsicDefinition = []IntrinsicDefinition{
	// Addition
	{"+", 1, math.MaxUint, intrinsicAdd},
	// Subtraction
	{"-", 1, math.MaxUint, intrinsicSub},
	// Multiplication
	{"*", 1, math.MaxUint, intrinsicMul},
}

func intrinsicAdd(arity uint) ast.Expr {
	return &ast.Add{Args: intrinsicNaryBody(arity)}
}

func intrinsicSub(arity uint) ast.Expr {
	return &ast.Sub{Args: intrinsicNaryBody(arity)}
}

func intrinsicMul(arity uint) ast.Expr {
	return &ast.Mul{Args: intrinsicNaryBody(arity)}
}

func intrinsicNaryBody(arity uint) []ast.Expr {
	args := make([]ast.Expr, arity)
	//
	for i := uint(0); i != arity; i++ {
		name := fmt.Sprintf("$%d", i)
		path := util.NewAbsolutePath(name)
		binding := &ast.LocalVariableBinding{Name: name, DataType: nil, Index: i}
		args[i] = ast.NewVariableAccess(path, true, binding)
	}
	//
	return args
}
