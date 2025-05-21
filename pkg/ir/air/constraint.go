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
	"github.com/consensys/go-corset/pkg/ir/schema"
	"github.com/consensys/go-corset/pkg/ir/schema/constraint"
	"github.com/consensys/go-corset/pkg/ir/schema/expr"
	"github.com/consensys/go-corset/pkg/ir/schema/term"
	"github.com/consensys/go-corset/pkg/trace"
	"github.com/consensys/go-corset/pkg/util"
	"github.com/consensys/go-corset/pkg/util/collection/bit"
	"github.com/consensys/go-corset/pkg/util/source/sexp"
)

// LookupConstraint captures the essence of a lookup constraint at the AIR
// level.  At the AIR level, lookup constraints are only permitted between
// columns (i.e. not arbitrary expressions).
type LookupConstraint = Air[constraint.LookupConstraint[*term.ColumnAccess[Term]]]

// PermutationConstraint captures the essence of a permutation constraint at the
// AIR level. Specifically, it represents a constraint that one (or more)
// columns are a permutation of another.
type PermutationConstraint = Air[constraint.PermutationConstraint]

// RangeConstraint captures the essence of a range constraints at the AIR level.
type RangeConstraint = Air[constraint.RangeConstraint[*term.ColumnAccess[Term]]]

// VanishingConstraint captures the essence of a vanishing constraint at the AIR level.
type VanishingConstraint = Air[constraint.VanishingConstraint[expr.Expr[Term]]]

// ============================================================================
// Helpers
// ============================================================================

// AirConstraint limits the permitted set of underlying constraints.  This
// should never change, unless the underlying prover changes in some way to
// offer different or more fundamental primitives.
type AirConstraint interface {
	schema.Constraint

	constraint.LookupConstraint[*term.ColumnAccess[Term]] |
		constraint.PermutationConstraint |
		constraint.RangeConstraint[*term.ColumnAccess[Term]] |
		constraint.VanishingConstraint[expr.Expr[Term]]
}

// Air attempts to encapsulate the notion of a valid constraint at the AIR
// level.  Since this is the fundamental level, only certain constraint forms
// are permitted.  As such, we want to try and ensure that arbitrary constraints
// are not found at the Air level.
type Air[C AirConstraint] struct {
	constraint C
}

// NewConstraint constructs a new Air constraint.
func NewConstraint[C AirConstraint](constraint C) Air[C] {
	return Air[C]{constraint}
}

// Air marks the constraint as being valid for the AIR representation.
func (p *Air[C]) Air() {
	// nothing as just a marker.
}

// Accepts determines whether a given constraint accepts a given trace or
// not.  If not, a failure is produced.  Otherwise, a bitset indicating
// branch coverage is returned.
func (p *Air[C]) Accepts(trace trace.Trace) (bit.Set, schema.Failure) {
	return p.constraint.Accepts(trace)
}

// Determine the well-definedness bounds for this constraint in both the
// negative (left) or positive (right) directions.  For example, consider an
// expression such as "(shift X -1)".  This is technically undefined for the
// first row of any trace and, by association, any constraint evaluating
// this expression on that first row is also undefined (and hence must pass)
func (p *Air[C]) Bounds(module uint) util.Bounds {
	return p.constraint.Bounds(module)
}

// Return the total number of logical branches this constraint can take
// during evaluation.
func (p *Air[C]) Branches() uint {
	return p.constraint.Branches()
}

// Contexts returns the evaluation contexts (i.e. enclosing module + length
// multiplier) for this constraint.  Most constraints have only a single
// evaluation context, though some (e.g. lookups) have more.  Note that all
// constraints have at least one context (which we can call the "primary"
// context).
func (p *Air[C]) Contexts() []trace.Context {
	return p.constraint.Contexts()
}

// Name returns a unique name and case number for a given constraint.  This
// is useful purely for identifying constraints in reports, etc.  The case
// number is used to differentiate different low-level constraints which are
// extracted from the same high-level constraint.
func (p *Air[C]) Name() (string, uint) {
	return p.constraint.Name()
}

// Lisp converts this schema element into a simple S-Expression, for example
// so it can be printed.
//
//nolint:revive
func (p Air[C]) Lisp(module schema.Module) sexp.SExp {
	return p.constraint.Lisp(module)
}

// ============================================================================

// ============================================================================

// Assertion captures the notion of an arbitrary property which should hold for
// all acceptable traces.  However, such a property is not enforced by the
// prover.
type Assertion = *schema.Assertion[schema.Testable]

var _ schema.Constraint = &LookupConstraint{}
var _ schema.Constraint = &VanishingConstraint{}
var _ schema.Constraint = &RangeConstraint{}
var _ schema.Constraint = &PermutationConstraint{}
var _ Constraint = &LookupConstraint{}
var _ Constraint = &VanishingConstraint{}
var _ Constraint = &RangeConstraint{}
var _ Constraint = &PermutationConstraint{}
