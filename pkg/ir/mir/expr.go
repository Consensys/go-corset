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
	"github.com/consensys/go-corset/pkg/ir"
)

// Term represents the fundamental for arithmetic expressions in the MIR
// representation.
type Term interface {
	ir.Term[Term]
}

type Expr = ir.Expr[Term]

// Add represents the addition of zero or more expressions.
type Add = ir.Add[Term]

// Cast attempts to narrow the width a given expression.
type Cast = ir.Cast[Term]

// Constant represents a constant value within an expression.
type Constant = ir.Constant[Term]

// ColumnAccess represents reading the value held at a given column in the
// tabular context.  Furthermore, the current row maybe shifted up (or down) by
// a given amount.
type ColumnAccess = ir.ColumnAccess[Term]

// Exp represents the a given value taken to a power.
type Exp = ir.Exp[Term]

// Mul represents the product over zero or more expressions.
type Mul = ir.Mul[Term]

// Norm reduces the value of an expression to either zero (if it was zero)
// or one (otherwise).
type Norm = ir.Norm[Term]

// Sub represents the subtraction over zero or more expressions.
type Sub = ir.Sub[Term]

// Void represents the empty expression.
var VOID Expr
