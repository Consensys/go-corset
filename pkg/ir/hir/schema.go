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
package hir

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
	"github.com/consensys/go-corset/pkg/util/word"
)

// Following types capture top-level abstractions at the HIR level.
type (
	// Module captures the essence of a module at the HIR level.  Specifically, it
	// is limited to only those constraint forms permitted at the HIR level.
	Module = *schema.Table[word.BigEndian, Constraint]
	// Schema captures the notion of an HIR schema which is uniform and consists of
	// HIR modules only.
	Schema = schema.UniformSchema[word.BigEndian, Module]
	// Term represents the fundamental for arithmetic expressions in the HIR
	// representation.
	Term interface {
		ir.Term[word.BigEndian, Term]
	}
	// LogicalTerm represents the fundamental for logical expressions in the HIR
	// representation.
	LogicalTerm interface {
		ir.LogicalTerm[word.BigEndian, LogicalTerm]
	}
)

// Following types capture permitted constraint forms at the HIR level.
type (
	// Assertion captures the notion of an arbitrary property which should hold for
	// all acceptable traces.  However, such a property is not enforced by the
	// prover.
	Assertion = constraint.Assertion[word.BigEndian, LogicalTerm]
	// InterleavingConstraint captures the essence of an interleaving constraint
	// at the HIR level.
	InterleavingConstraint = interleaving.Constraint[word.BigEndian, Term]
	// LookupConstraint captures the essence of a lookup constraint at the HIR
	// level.
	LookupConstraint = lookup.Constraint[word.BigEndian, Term]
	// PermutationConstraint captures the essence of a permutation constraint at the
	// HIR level.
	PermutationConstraint = permutation.Constraint[word.BigEndian]
	// RangeConstraint captures the essence of a range constraints at the HIR level.
	RangeConstraint = ranged.Constraint[word.BigEndian, Term]
	// SortedConstraint captures the essence of a sorted constraint at the HIR
	// level.
	SortedConstraint = sorted.Constraint[word.BigEndian, Term]
	// VanishingConstraint captures the essence of a vanishing constraint at the HIR
	// level. A vanishing constraint is a row constraint which must evaluate to
	// zero.
	VanishingConstraint = vanishing.Constraint[word.BigEndian, LogicalTerm]
)

// Following types capture permitted expression forms at the HIR level.
type (
	// Add represents the addition of zero or more expressions.
	Add = ir.Add[word.BigEndian, Term]
	// Cast attempts to narrow the width a given expression.
	Cast = ir.Cast[word.BigEndian, Term]
	// Constant represents a constant value within an expression.
	Constant = ir.Constant[word.BigEndian, Term]
	// IfZero represents a conditional branch at the HIR level.
	IfZero = ir.IfZero[word.BigEndian, LogicalTerm, Term]
	// LabelledConst represents a labelled constant at the HIR level.
	LabelledConst = ir.LabelledConst[word.BigEndian, Term]
	// RegisterAccess represents reading the value held at a given column in the
	// tabular context.  Furthermore, the current row maybe shifted up (or down) by
	// a given amount.
	RegisterAccess = ir.RegisterAccess[word.BigEndian, Term]
	// Exp represents the a given value taken to a power.
	Exp = ir.Exp[word.BigEndian, Term]
	// Mul represents the product over zero or more expressions.
	Mul = ir.Mul[word.BigEndian, Term]
	// Norm reduces the value of an expression to either zero (if it was zero)
	// or one (otherwise).
	Norm = ir.Norm[word.BigEndian, Term]
	// Sub represents the subtraction over zero or more expressions.
	Sub = ir.Sub[word.BigEndian, Term]
	// VectorAccess represents a compound variable
	VectorAccess = ir.VectorAccess[word.BigEndian, Term]
)

// Following types capture permitted logical forms at the HIR level.
type (
	// Conjunct represents a logical conjunction at the HIR level.
	Conjunct = ir.Conjunct[word.BigEndian, LogicalTerm]
	// Disjunct represents a logical conjunction at the HIR level.
	Disjunct = ir.Disjunct[word.BigEndian, LogicalTerm]
	// Equal represents an equality comparator between two arithmetic terms
	// at the HIR level.
	Equal = ir.Equal[word.BigEndian, LogicalTerm, Term]
	// Ite represents an If-Then-Else expression where either branch is optional
	// (though we must have at least one).
	Ite = ir.Ite[word.BigEndian, LogicalTerm]
	// Negate represents a logical negation at the HIR level.
	Negate = ir.Negate[word.BigEndian, LogicalTerm]
	// NotEqual represents a non-equality comparator between two arithmetic terms
	// at the HIR level.
	NotEqual = ir.NotEqual[word.BigEndian, LogicalTerm, Term]
)

// SubstituteConstants substitutes the value of matching labelled constants for
// all expressions used within the schema.
func SubstituteConstants(schema schema.AnySchema[word.BigEndian], mapping map[string]word.BigEndian) {
	// Constraints
	for iter := schema.Modules(); iter.HasNext(); {
		module := iter.Next()
		module.Substitute(mapping)
	}
}

func init() {
	gob.Register(schema.Constraint[word.BigEndian](&VanishingConstraint{}))
	gob.Register(schema.Constraint[word.BigEndian](&RangeConstraint{}))
	gob.Register(schema.Constraint[word.BigEndian](&PermutationConstraint{}))
	gob.Register(schema.Constraint[word.BigEndian](&LookupConstraint{}))
	gob.Register(schema.Constraint[word.BigEndian](&SortedConstraint{}))
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
