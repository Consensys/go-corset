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
	"github.com/consensys/go-corset/pkg/schema"
	"github.com/consensys/go-corset/pkg/schema/constraint"
)

// Following types capture top-level abstractions at the AIR level.
type (
	// Schema captures the essence of an arithmetisation at the AIR level.
	// Specifically, it is limited to only those constraint forms permitted at the
	// AIR level.
	Schema = schema.UniformSchema[Module]
	// Module captures the essence of a module at the AIR level.  Specifically, it
	// is limited to only those constraint forms permitted at the AIR level.
	Module = schema.Table[Constraint]
	// Constraint captures the essence of a constraint at the AIR level.
	Constraint interface {
		schema.Constraint
		// Air marks the constraints as been valid for the AIR representation.
		Air()
	}
	// Term represents the fundamental for arithmetic expressions in the AIR
	// representation.  This should only support addition, subtraction and
	// multiplication of constants and column accesses.  No other terms are
	// permitted at this, the lowest, layer of the stack.
	Term interface {
		ir.Term[Term]
		// Air marks terms which are valid for the AIR representation.
		Air()
	}
	// LogicalTerm represents the fundamental for logical expressions in the AIR
	// representation.
	LogicalTerm interface {
		ir.LogicalTerm[LogicalTerm]
		// Air marks terms which are valid for the AIR representation.
		Air()
	}
)

// Following types capture permitted constraint forms at the AIR level.
type (
	// Assertion captures the notion of an arbitrary property which should hold for
	// all acceptable traces.  However, such a property is not enforced by the
	// prover.
	Assertion = *constraint.Assertion[ir.Testable]
	// LookupConstraint captures the essence of a lookup constraint at the AIR
	// level.  At the AIR level, lookup constraints are only permitted between
	// columns (i.e. not arbitrary expressions).
	LookupConstraint = Air[constraint.LookupConstraint[*ir.RegisterAccess[Term]]]
	// PermutationConstraint captures the essence of a permutation constraint at the
	// AIR level. Specifically, it represents a constraint that one (or more)
	// columns are a permutation of another.
	PermutationConstraint = Air[constraint.PermutationConstraint]
	// RangeConstraint captures the essence of a range constraints at the AIR level.
	RangeConstraint = Air[constraint.RangeConstraint[*ir.RegisterAccess[Term]]]
	// VanishingConstraint captures the essence of a vanishing constraint at the AIR level.
	VanishingConstraint = Air[constraint.VanishingConstraint[LogicalTerm]]
)

// Following types capture permitted expression forms at the AIR level.
type (
	// Add represents the addition of zero or more AIR expressio
	Add = ir.Add[Term]
	// Constant represents a constant value within AIR an expression.
	Constant = ir.Constant[Term]
	// ColumnAccess represents reading the value held at a given column in the
	// tabular context.  Furthermore, the current row maybe shifted up (or down) by
	// a given amount.
	ColumnAccess = ir.RegisterAccess[Term]
	// Mul represents the product over zero or more expressions.
	Mul = ir.Mul[Term]
	// Sub represents the subtraction over zero or more expressions.
	Sub = ir.Sub[Term]
)

// Following types capture permitted logical forms at the AIR level.
type (
	// Conjunct represents a logical conjunction at the AIR level.
	Conjunct = ir.Conjunct[LogicalTerm]
	// Disjunct represents a logical conjunction at the AIR level.
	Disjunct = ir.Disjunct[LogicalTerm]
	// Equal captures the notion of an equality at the AIR level which,
	// practically speaking, reduces to a subtraction.
	Equal = ir.Equal[Term]
)
