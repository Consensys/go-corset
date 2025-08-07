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
	"encoding/gob"

	"github.com/consensys/go-corset/pkg/ir"
	"github.com/consensys/go-corset/pkg/schema"
	"github.com/consensys/go-corset/pkg/schema/constraint"
	"github.com/consensys/go-corset/pkg/schema/constraint/interleaving"
	"github.com/consensys/go-corset/pkg/schema/constraint/lookup"
	"github.com/consensys/go-corset/pkg/schema/constraint/permutation"
	"github.com/consensys/go-corset/pkg/schema/constraint/ranged"
	"github.com/consensys/go-corset/pkg/schema/constraint/sorted"
	"github.com/consensys/go-corset/pkg/schema/constraint/vanishing"
	"github.com/consensys/go-corset/pkg/util/field/bls12_377"
)

// Following types capture top-level abstractions at the MIR level.
type (
	// Module captures the essence of a module at the MIR level.  Specifically, it
	// is limited to only those constraint forms permitted at the MIR level.
	Module = *schema.Table[Constraint]
	// Schema captures the notion of an MIR schema which is uniform and consists of
	// MIR modules only.
	Schema = schema.UniformSchema[Module]
	// Term represents the fundamental for arithmetic expressions in the MIR
	// representation.
	Term interface {
		ir.Term[bls12_377.Element, Term]
	}
	// LogicalTerm represents the fundamental for logical expressions in the MIR
	// representation.
	LogicalTerm interface {
		ir.LogicalTerm[bls12_377.Element, LogicalTerm]
	}
)

// Following types capture permitted constraint forms at the MIR level.
type (
	// Assertion captures the notion of an arbitrary property which should hold for
	// all acceptable traces.  However, such a property is not enforced by the
	// prover.
	Assertion = constraint.Assertion[LogicalTerm]
	// InterleavingConstraint captures the essence of an interleaving constraint
	// at the MIR level.
	InterleavingConstraint = interleaving.Constraint[Term]
	// LookupConstraint captures the essence of a lookup constraint at the MIR
	// level.
	LookupConstraint = lookup.Constraint[Term]
	// PermutationConstraint captures the essence of a permutation constraint at the
	// MIR level.
	PermutationConstraint = permutation.Constraint
	// RangeConstraint captures the essence of a range constraints at the MIR level.
	RangeConstraint = ranged.Constraint[Term]
	// SortedConstraint captures the essence of a sorted constraint at the MIR
	// level.
	SortedConstraint = sorted.Constraint[Term]
	// VanishingConstraint captures the essence of a vanishing constraint at the MIR
	// level. A vanishing constraint is a row constraint which must evaluate to
	// zero.
	VanishingConstraint = vanishing.Constraint[LogicalTerm]
)

// Following types capture permitted expression forms at the MIR level.
type (
	// Add represents the addition of zero or more expressions.
	Add = ir.Add[bls12_377.Element, Term]
	// Cast attempts to narrow the width a given expression.
	Cast = ir.Cast[bls12_377.Element, Term]
	// Constant represents a constant value within an expression.
	Constant = ir.Constant[bls12_377.Element, Term]
	// IfZero represents a conditional branch at the MIR level.
	IfZero = ir.IfZero[bls12_377.Element, LogicalTerm, Term]
	// LabelledConst represents a labelled constant at the MIR level.
	LabelledConst = ir.LabelledConst[bls12_377.Element, Term]
	// RegisterAccess represents reading the value held at a given column in the
	// tabular context.  Furthermore, the current row maybe shifted up (or down) by
	// a given amount.
	RegisterAccess = ir.RegisterAccess[bls12_377.Element, Term]
	// Exp represents the a given value taken to a power.
	Exp = ir.Exp[bls12_377.Element, Term]
	// Mul represents the product over zero or more expressions.
	Mul = ir.Mul[bls12_377.Element, Term]
	// Norm reduces the value of an expression to either zero (if it was zero)
	// or one (otherwise).
	Norm = ir.Norm[bls12_377.Element, Term]
	// Sub represents the subtraction over zero or more expressions.
	Sub = ir.Sub[bls12_377.Element, Term]
	// VectorAccess represents a compound variable
	VectorAccess = ir.VectorAccess[bls12_377.Element, Term]
)

// Following types capture permitted logical forms at the MIR level.
type (
	// Conjunct represents a logical conjunction at the MIR level.
	Conjunct = ir.Conjunct[bls12_377.Element, LogicalTerm]
	// Disjunct represents a logical conjunction at the MIR level.
	Disjunct = ir.Disjunct[bls12_377.Element, LogicalTerm]
	// Equal represents an equality comparator between two arithmetic terms
	// at the MIR level.
	Equal = ir.Equal[bls12_377.Element, LogicalTerm, Term]
	// Ite represents an If-Then-Else expression where either branch is optional
	// (though we must have at least one).
	Ite = ir.Ite[bls12_377.Element, LogicalTerm]
	// Negate represents a logical negation at the MIR level.
	Negate = ir.Negate[bls12_377.Element, LogicalTerm]
	// NotEqual represents a non-equality comparator between two arithmetic terms
	// at the MIR level.
	NotEqual = ir.NotEqual[bls12_377.Element, LogicalTerm, Term]
	// Inequality an inequality comparator (e.g. X < Y or X <= Y) between two arithmetic terms
	// at the MIR level.
	Inequality = ir.Inequality[bls12_377.Element, LogicalTerm, Term]
)

// SubstituteConstants substitutes the value of matching labelled constants for
// all expressions used within the schema.
func SubstituteConstants[M schema.Module](schema schema.MixedSchema[M, Module], mapping map[string]bls12_377.Element) {
	// Constraints
	for iter := schema.Constraints(); iter.HasNext(); {
		constraint := iter.Next()
		constraint.Substitute(mapping)
	}
}

func init() {
	gob.Register(schema.Constraint(&VanishingConstraint{}))
	gob.Register(schema.Constraint(&RangeConstraint{}))
	gob.Register(schema.Constraint(&PermutationConstraint{}))
	gob.Register(schema.Constraint(&LookupConstraint{}))
	gob.Register(schema.Constraint(&SortedConstraint{}))
	//
	gob.Register(Term(&Add{}))
	gob.Register(Term(&Mul{}))
	gob.Register(Term(&Sub{}))
	gob.Register(Term(&Cast{}))
	gob.Register(Term(&Exp{}))
	gob.Register(Term(&IfZero{}))
	gob.Register(Term(&Constant{}))
	gob.Register(Term(&LabelledConst{}))
	gob.Register(Term(&Norm{}))
	gob.Register(Term(&RegisterAccess{}))
	//
	gob.Register(LogicalTerm(&Conjunct{}))
	gob.Register(LogicalTerm(&Disjunct{}))
	gob.Register(LogicalTerm(&Equal{}))
	gob.Register(LogicalTerm(&NotEqual{}))
}
