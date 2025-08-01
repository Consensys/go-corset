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
	"github.com/consensys/gnark-crypto/ecc/bls12-377/fr"
	"github.com/consensys/go-corset/pkg/ir"
	"github.com/consensys/go-corset/pkg/schema"
	"github.com/consensys/go-corset/pkg/schema/constraint"
	"github.com/consensys/go-corset/pkg/schema/constraint/interleaving"
	"github.com/consensys/go-corset/pkg/schema/constraint/lookup"
	"github.com/consensys/go-corset/pkg/schema/constraint/permutation"
	"github.com/consensys/go-corset/pkg/schema/constraint/ranged"
	"github.com/consensys/go-corset/pkg/schema/constraint/vanishing"
	"github.com/consensys/go-corset/pkg/trace"
	"github.com/consensys/go-corset/pkg/util"
	"github.com/consensys/go-corset/pkg/util/collection/set"
	"github.com/consensys/go-corset/pkg/util/source/sexp"
)

// Following types capture top-level abstractions at the AIR level.
type (
	// Schema captures the essence of an arithmetisation at the AIR level.
	// Specifically, it is limited to only those constraint forms permitted at the
	// AIR level.
	Schema = schema.UniformSchema[Module]
	// Module captures the essence of a module at the AIR level.  Specifically, it
	// is limited to only those constraint forms permitted at the AIR level.
	Module = *schema.Table[Constraint]
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
)

type (
	// SchemaBuilder is used for building the AIR schemas
	SchemaBuilder = ir.SchemaBuilder[Constraint, Term]
	// ModuleBuilder is used for building various AIR modules.
	ModuleBuilder = ir.ModuleBuilder[Constraint, Term]
)

var _ schema.Module = &ModuleBuilder{}

// Following types capture permitted constraint forms at the AIR level.
type (
	// Assertion captures the notion of an arbitrary property which should hold for
	// all acceptable traces.  However, such a property is not enforced by the
	// prover.
	Assertion = Air[constraint.Assertion[ir.Testable]]
	// InterleavingConstraint captures the essence of an interleaving constraint
	// at the MIR level.
	InterleavingConstraint = Air[interleaving.Constraint[*ColumnAccess]]
	// LookupConstraint captures the essence of a lookup constraint at the AIR
	// level.  At the AIR level, lookup constraints are only permitted between
	// columns (i.e. not arbitrary expressions).
	LookupConstraint = Air[lookup.Constraint[*ColumnAccess]]
	// PermutationConstraint captures the essence of a permutation constraint at the
	// AIR level. Specifically, it represents a constraint that one (or more)
	// columns are a permutation of another.
	PermutationConstraint = Air[permutation.Constraint]
	// RangeConstraint captures the essence of a range constraints at the AIR level.
	RangeConstraint = Air[ranged.Constraint[*ColumnAccess]]
	// VanishingConstraint captures the essence of a vanishing constraint at the AIR level.
	VanishingConstraint = Air[vanishing.Constraint[LogicalTerm]]
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

var zero fr.Element = fr.NewElement(0)

// LogicalTerm provides a wrapper around a given term allowing to be "testable".
// That is, it provides a default TestAt implementation.
type LogicalTerm struct {
	Term Term
}

// Bounds implementation for Boundable interface.
func (p LogicalTerm) Bounds() util.Bounds {
	return p.Term.Bounds()
}

// TestAt implementation for Testable interface.
func (p LogicalTerm) TestAt(k int, tr trace.Module, sc schema.Module) (bool, uint, error) {
	var val, err = p.Term.EvalAt(k, tr, sc)
	//
	if err != nil {
		return false, 0, err
	}
	//
	return val.Cmp(&zero) == 0, 0, nil
}

// Lisp returns a lisp representation of this NotEqual, which is useful for
// debugging.
func (p LogicalTerm) Lisp(global bool, mapping schema.RegisterMap) sexp.SExp {
	return p.Term.Lisp(global, mapping)
}

// RequiredRegisters implementation for Contextual interface.
func (p LogicalTerm) RequiredRegisters() *set.SortedSet[uint] {
	return p.Term.RequiredRegisters()
}

// RequiredCells implementation for Contextual interface
func (p LogicalTerm) RequiredCells(row int, mid trace.ModuleId) *set.AnySortedSet[trace.CellRef] {
	return p.Term.RequiredCells(row, mid)
}

// Substitute implementation for Substitutable interface.
func (p LogicalTerm) Substitute(mapping map[string]fr.Element) {
	p.Term.Substitute(mapping)
}
