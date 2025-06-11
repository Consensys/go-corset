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

	"github.com/consensys/gnark-crypto/ecc/bls12-377/fr"
	"github.com/consensys/go-corset/pkg/ir"
	"github.com/consensys/go-corset/pkg/ir/mir"
	"github.com/consensys/go-corset/pkg/schema"
	"github.com/consensys/go-corset/pkg/util"
)

// ModuleBuilder is used within this translator for building the various modules
// which are contained within the mixed MIR schema.
type ModuleBuilder = ir.ModuleBuilder[mir.Constraint, mir.Term]

// MirModule provides a wrapper around a corset-level module declaration.
type MirModule struct {
	Module *ModuleBuilder
}

// Initialise this module
func (p MirModule) Initialise(name string, mid uint) MirModule {
	builder := ir.NewModuleBuilder[mir.Constraint, mir.Term](name, mid, 1)
	p.Module = &builder

	return p
}

// NewColumn constructs a new column of the given name and bitwidth within
// this module.
func (p MirModule) NewColumn(kind schema.RegisterType, name string, bitwidth uint) schema.RegisterId {
	return p.Module.NewRegister(schema.NewRegister(kind, name, bitwidth))
}

// NewConstraint constructs a new vanishing constraint with the given name
// within this module.
func (p MirModule) NewConstraint(name string, domain util.Option[int], constraint MirExpr) {
	p.Module.AddConstraint(
		mir.NewVanishingConstraint(name, p.Module.Id(), domain, constraint.logical))
}

// NewLookup constructs a new lookup constraint
func (p MirModule) NewLookup(name string, from []MirExpr, target uint, to []MirExpr) {
	var (
		sources = unwrapMirExprs(from...)
		targets = unwrapMirExprs(to...)
	)

	p.Module.AddConstraint(mir.NewLookupConstraint(name, target, targets, p.Module.Id(), sources))
}

// String returns an appropriately formatted representation of the module.
func (p MirModule) String() string {
	panic("todo")
}

// MirExpr is a wrapper around a corset expression which provides the
// necessary interface.
type MirExpr struct {
	expr    mir.Term
	logical mir.LogicalTerm
}

// Add constructs a sum between this expression and zero or more
func (p MirExpr) Add(exprs ...MirExpr) MirExpr {
	args := unwrapSplitMirExpr(p, exprs...)
	return MirExpr{ir.Sum(args...), nil}
}

// And constructs a conjunction between this expression and zero or more
// expressions.
func (p MirExpr) And(exprs ...MirExpr) MirExpr {
	args := unwrapSplitMirLogicals(p, exprs...)
	return MirExpr{nil, ir.Conjunction(args...)}
}

// Equals constructs an equality between two expressions.
func (p MirExpr) Equals(rhs MirExpr) MirExpr {
	logical := ir.Equals[mir.LogicalTerm](p.expr, rhs.expr)
	return MirExpr{nil, logical}
}

// Then constructs an implication between two expressions.
func (p MirExpr) Then(trueBranch MirExpr) MirExpr {
	logical := ir.IfThenElse(p.logical, trueBranch.logical, nil)
	return MirExpr{nil, logical}
}

// ThenElse constructs an if-then-else expression with this expression
// acting as the condition.
func (p MirExpr) ThenElse(trueBranch MirExpr, falseBranch MirExpr) MirExpr {
	logical := ir.IfThenElse(p.logical, trueBranch.logical, falseBranch.logical)
	return MirExpr{nil, logical}
}

// Multiply constructs a product between this expression and zero or more
// expressions.
func (p MirExpr) Multiply(exprs ...MirExpr) MirExpr {
	args := unwrapSplitMirExpr(p, exprs...)
	return MirExpr{ir.Product(args...), nil}
}

// NotEquals constructs a non-equality between two expressions.
func (p MirExpr) NotEquals(rhs MirExpr) MirExpr {
	logical := ir.Equals[mir.LogicalTerm](p.expr, rhs.expr)
	return MirExpr{nil, logical}
}

// BigInt constructs a constant expression from a big integer.
func (p MirExpr) BigInt(number big.Int) MirExpr {
	var frNum fr.Element
	//
	frNum.BigInt(&number)
	//
	return MirExpr{ir.Const[mir.Term](frNum), nil}
}

// Or constructs a disjunction between this expression and zero or more
// expressions.
func (p MirExpr) Or(exprs ...MirExpr) MirExpr {
	args := unwrapSplitMirLogicals(p, exprs...)
	return MirExpr{nil, ir.Disjunction(args...)}
}

// Variable constructs a variable with a given shift.
func (p MirExpr) Variable(index schema.RegisterId, shift int) MirExpr {
	return MirExpr{ir.NewRegisterAccess[mir.Term](index, shift), nil}
}

func unwrapSplitMirExpr(head MirExpr, tail ...MirExpr) []mir.Term {
	cexprs := make([]mir.Term, len(tail)+1)
	//
	cexprs[0] = head.expr
	//
	for i, e := range tail {
		cexprs[i+1] = e.expr
		//
		if e.logical != nil {
			panic("logical expression encountered")
		}
	}
	//
	return cexprs
}

func unwrapSplitMirLogicals(head MirExpr, tail ...MirExpr) []mir.LogicalTerm {
	cexprs := make([]mir.LogicalTerm, len(tail)+1)
	//
	cexprs[0] = head.logical
	//
	for i, e := range tail {
		cexprs[i+1] = e.logical
		//
		if e.expr != nil {
			panic("arithmetic expression encountered")
		}
	}
	//
	return cexprs
}

func unwrapMirExprs(exprs ...MirExpr) []mir.Term {
	cexprs := make([]mir.Term, len(exprs))
	//
	for i, e := range exprs {
		cexprs[i] = e.expr
		//
		if e.logical != nil {
			panic("logical expression encountered")
		}
	}
	//
	return cexprs
}
