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
	"github.com/consensys/go-corset/pkg/schema"
	"github.com/consensys/go-corset/pkg/schema/constraint"
	"github.com/consensys/go-corset/pkg/schema/constraint/interleaving"
	"github.com/consensys/go-corset/pkg/schema/constraint/lookup"
	"github.com/consensys/go-corset/pkg/schema/constraint/permutation"
	"github.com/consensys/go-corset/pkg/schema/constraint/ranged"
	"github.com/consensys/go-corset/pkg/schema/constraint/sorted"
	"github.com/consensys/go-corset/pkg/schema/constraint/vanishing"
	"github.com/consensys/go-corset/pkg/util/field"
)

// Following types capture top-level abstractions at the MIR level.
type (
	// SchemaBuilder is used for building the MIR schemas
	SchemaBuilder[F field.Element[F]] = ir.SchemaBuilder[F, Constraint[F], Term[F]]
	// ModuleBuilder is used for building various MIR modules.
	ModuleBuilder[F field.Element[F]] = ir.ModuleBuilder[F, Constraint[F], Term[F]]
	// Module captures the essence of a module at the MIR level.  Specifically, it
	// is limited to only those constraint forms permitted at the MIR level.
	Module[F field.Element[F]] = *schema.Table[F, Constraint[F]]
	// Schema captures the notion of an MIR schema which is uniform and consists of
	// MIR modules only.
	Schema[F field.Element[F]] = schema.UniformSchema[F, Module[F]]
	// Term represents the fundamental for arithmetic expressions in the MIR
	// representation.
	Term[F any] interface {
		ir.Term[F, Term[F]]
	}
	// LogicalTerm represents the fundamental for logical expressions in the MIR
	// representation.
	LogicalTerm[F any] interface {
		ir.LogicalTerm[F, LogicalTerm[F]]
	}
)

// Following types capture key assignment forms at the MIR level.
type (
// ComputedRegister captures one form of computation permitted at the MIR level.
// ComputedRegister[F field.Element[F]] = assignment.ComputedRegister[F, Term[F]]
)

// Following types capture permitted constraint forms at the MIR level.
type (
	// Assertion captures the notion of an arbitrary property which should hold for
	// all acceptable traces.  However, such a property is not enforced by the
	// prover.
	Assertion[F field.Element[F]] = constraint.Assertion[F, LogicalTerm[F]]
	// InterleavingConstraint captures the essence of an interleaving constraint
	// at the MIR level.
	InterleavingConstraint[F field.Element[F]] = interleaving.Constraint[F, Term[F]]
	// LookupConstraint captures the essence of a lookup constraint at the MIR
	// level.
	LookupConstraint[F field.Element[F]] = lookup.Constraint[F, Term[F]]
	// PermutationConstraint captures the essence of a permutation constraint at the
	// MIR level.
	PermutationConstraint[F field.Element[F]] = permutation.Constraint[F]
	// RangeConstraint captures the essence of a range constraints at the MIR level.
	RangeConstraint[F field.Element[F]] = ranged.Constraint[F, *RegisterAccess[F]]
	// SortedConstraint captures the essence of a sorted constraint at the MIR
	// level.
	SortedConstraint[F field.Element[F]] = sorted.Constraint[F, Term[F]]
	// VanishingConstraint captures the essence of a vanishing constraint at the MIR
	// level. A vanishing constraint is a row constraint which must evaluate to
	// zero.
	VanishingConstraint[F field.Element[F]] = vanishing.Constraint[F, LogicalTerm[F]]
)

// Following types capture permitted expression forms at the MIR level.
type (
	// Add represents the addition of zero or more expressions.
	Add[F field.Element[F]] = ir.Add[F, Term[F]]
	// Constant represents a constant value within an expression.
	Constant[F field.Element[F]] = ir.Constant[F, Term[F]]
	// RegisterAccess represents reading the value held at a given column in the
	// tabular context.  Furthermore, the current row maybe shifted up (or down) by
	// a given amount.
	RegisterAccess[F field.Element[F]] = ir.RegisterAccess[F, Term[F]]
	// Mul represents the product over zero or more expressions.
	Mul[F field.Element[F]] = ir.Mul[F, Term[F]]
	// Sub represents the subtraction over zero or more expressions.
	Sub[F field.Element[F]] = ir.Sub[F, Term[F]]
	// VectorAccess represents a compound variable
	VectorAccess[F field.Element[F]] = ir.VectorAccess[F, Term[F]]
)

// Following types capture permitted logical forms at the MIR level.
type (
	// Conjunct represents a logical conjunction at the MIR level.
	Conjunct[F field.Element[F]] = ir.Conjunct[F, LogicalTerm[F]]
	// Disjunct represents a logical conjunction at the MIR level.
	Disjunct[F field.Element[F]] = ir.Disjunct[F, LogicalTerm[F]]
	// Equal represents an equality comparator between two arithmetic terms
	// at the MIR level.
	Equal[F field.Element[F]] = ir.Equal[F, LogicalTerm[F], Term[F]]
	// Ite represents an If-Then-Else expression where either branch is optional
	// (though we must have at least one).
	Ite[F field.Element[F]] = ir.Ite[F, LogicalTerm[F]]
	// Negate represents a logical negation at the MIR level.
	Negate[F field.Element[F]] = ir.Negate[F, LogicalTerm[F]]
	// NotEqual represents a non-equality comparator between two arithmetic terms
	// at the MIR level.
	NotEqual[F field.Element[F]] = ir.NotEqual[F, LogicalTerm[F], Term[F]]
	// Inequality an inequality comparator (e.g. X < Y or X <= Y) between two arithmetic terms
	// at the MIR level.
	Inequality[F field.Element[F]] = ir.Inequality[F, LogicalTerm[F], Term[F]]
)

// SubstituteConstants substitutes the value of matching labelled constants for
// all expressions used within the schema.
func SubstituteConstants[F field.Element[F]](schema schema.AnySchema[F], mapping map[string]F) {
	// Constraints
	for iter := schema.Modules(); iter.HasNext(); {
		module := iter.Next()
		module.Substitute(mapping)
	}
}
