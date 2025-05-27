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
package air

import (
	"github.com/consensys/go-corset/pkg/ir"
)

// Expr captures the notion of an expression" at the AIR level.  This is really
// just for convenience more than anything.
type Expr = ir.Expr[Term]

// Term represents the fundamental for arithmetic expressions in the AIR
// representation.  This should only support addition, subtraction and
// multiplication of constants and column accesses.  No other terms are
// permitted at this, the lowest, layer of the stack.
type Term interface {
	ir.Term[Term]
	// Air marks terms which are valid for the AIR representation.
	Air()
}

// Add represents the addition of zero or more AIR expressions.
type Add = ir.Add[Term]

// Constant represents a constant value within AIR an expression.
type Constant = ir.Constant[Term]

// ColumnAccess represents reading the value held at a given column in the
// tabular context.  Furthermore, the current row maybe shifted up (or down) by
// a given amount.
type ColumnAccess = ir.ColumnAccess[Term]

// Mul represents the product over zero or more expressions.
type Mul = ir.Mul[Term]

// Sub represents the subtraction over zero or more expressions.
type Sub = ir.Sub[Term]
