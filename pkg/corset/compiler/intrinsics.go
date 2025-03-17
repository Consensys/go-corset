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
	"github.com/consensys/go-corset/pkg/util"
	"github.com/consensys/go-corset/pkg/util/source/sexp"
)

// IntrinsicDefinition is a SymbolDefinition for an intrinsic (i.e. built-in)
// operation, such as "+", "-", etc.  These are needed for two reasons: firstly,
// so we can alias them; secondly, so they can be used in reductions.
type IntrinsicDefinition struct {
	// Name of the intrinsic (e.g. "+")
	name string
	// Number of arguments this intrinsic can accept.
	arity uint
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

// Arity indicates whether or not this is a function and, if so, what arity
// (i.e. how many arguments) the function has.
func (p *IntrinsicDefinition) Arity() util.Option[uint] {
	return util.Some(p.arity)
}

// IsPure checks whether this pure (which intrinsics always are).
func (p *IntrinsicDefinition) IsPure() bool {
	return true
}

// IsNative checks whether this function binding is native (or not).
func (p *IntrinsicDefinition) IsNative() bool {
	return false
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

// Signature returns the function signature for this binding.
func (p *IntrinsicDefinition) Signature() *ast.FunctionSignature {
	// construct the body
	body := p.constructor(p.arity)
	types := make([]ast.Type, p.arity)
	//
	for i := 0; i < len(types); i++ {
		types[i] = ast.INT_TYPE
	}
	// Allow return type to be inferred.
	return ast.NewFunctionSignature(true, types, nil, body)
}

// ============================================================================
// Intrinsic Definitions
// ============================================================================

// INTRINSICS identifies all of the built-in functions used within the corset
// language, such as "+", "-", etc.  This is needed for two reasons: firstly, so
// we can alias them; secondly, so they can be used in reductions.
var INTRINSICS []IntrinsicDefinition = []IntrinsicDefinition{
	// Addition
	{"+", 1, intrinsicAdd},
	// Subtraction
	{"-", 1, intrinsicSub},
	// Multiplication
	{"*", 1, intrinsicMul},
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
		args[i] = ast.NewVariableAccess(path, util.Some(arity), binding)
	}
	//
	return args
}
