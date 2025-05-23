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
package mir

import (
	"github.com/consensys/gnark-crypto/ecc/bls12-377/fr"
	"github.com/consensys/go-corset/pkg/ir/schema"
	"github.com/consensys/go-corset/pkg/ir/schema/expr"
	"github.com/consensys/go-corset/pkg/ir/schema/term"
)

// Term represents the fundamental for arithmetic expressions in the MIR
// representation.
type Term interface {
	schema.Term[Term]
}

type Expr = expr.Expr[Term]

// Add represents the addition of zero or more expressions.
type Add = term.Add[Term]

// Cast attempts to narrow the width a given expression.
type Cast = term.Cast[Term]

// Constant represents a constant value within an expression.
type Constant = term.Constant[Term]

// ColumnAccess represents reading the value held at a given column in the
// tabular context.  Furthermore, the current row maybe shifted up (or down) by
// a given amount.
type ColumnAccess = term.ColumnAccess[Term]

// Exp represents the a given value taken to a power.
type Exp = term.Exp[Term]

// Mul represents the product over zero or more expressions.
type Mul = term.Mul[Term]

// Norm reduces the value of an expression to either zero (if it was zero)
// or one (otherwise).
type Norm = term.Norm[Term]

// Sub represents the subtraction over zero or more expressions.
type Sub = term.Sub[Term]

// Void represents the empty expression.
var VOID Expr

// NewColumnAccess constructs an AIR expression representing the value of a given
// column on the current row.
func NewColumnAccess(column uint, shift int) Expr {
	term := &ColumnAccess{Column: column, Shift: shift}
	return Expr{Term: term}
}

// NewConst construct an AIR expression representing a given constant.
func NewConst(val fr.Element) Expr {
	term := &Constant{Value: val}
	return Expr{Term: term}
}

// NewConst64 construct an AIR expression representing a given constant from a
// uint64.
func NewConst64(val uint64) Expr {
	element := fr.NewElement(val)
	term := &Constant{Value: element}
	return Expr{Term: term}
}

// Sum zero or more expressions together.
func Sum(exprs ...Expr) Expr {
	panic("todo")
}

// Product returns the product of zero or more multiplications.
func Product(exprs ...Expr) Expr {
	panic("todo")
}

// Subtract returns the subtraction of the subsequent expressions from the
// first.
func Subtract(exprs ...Expr) Expr {
	panic("todo")
}
