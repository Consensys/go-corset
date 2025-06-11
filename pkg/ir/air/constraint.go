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
	"github.com/consensys/go-corset/pkg/trace"
	"github.com/consensys/go-corset/pkg/util"
	"github.com/consensys/go-corset/pkg/util/collection/bit"
	"github.com/consensys/go-corset/pkg/util/source/sexp"
)

// ============================================================================
// Helpers
// ============================================================================

// ConstraintBound limits the permitted set of underlying constraints.  This
// should never change, unless the underlying prover changes in some way to
// offer different or more fundamental primitives.
type ConstraintBound interface {
	schema.Constraint

	constraint.Assertion[ir.Testable] |
		constraint.LookupConstraint[*ir.RegisterAccess[Term]] |
		constraint.PermutationConstraint |
		constraint.RangeConstraint[*ir.RegisterAccess[Term]] |
		constraint.VanishingConstraint[LogicalTerm]
}

// Air attempts to encapsulate the notion of a valid constraint at the AIR
// level.  Since this is the fundamental level, only certain constraint forms
// are permitted.  As such, we want to try and ensure that arbitrary constraints
// are not found at the Air level.
type Air[C ConstraintBound] struct {
	constraint C
}

// newAir is a helper method for the various constraint constructors, basically
// to avoid lots of generic types.
func newAir[C ConstraintBound](constraint C) Air[C] {
	return Air[C]{constraint}
}

// NewAssertion constructs a new AIR assertion
func NewAssertion(handle string, ctx schema.ModuleId, term ir.Testable) Assertion {
	//
	return newAir(constraint.NewAssertion(handle, ctx, term))
}

// NewLookupConstraint constructs a new AIR lookup constraint
func NewLookupConstraint(handle string, source schema.ModuleId,
	target schema.ModuleId, sources []*ColumnAccess, targets []*ColumnAccess) LookupConstraint {
	return newAir(constraint.NewLookupConstraint(handle, source, target, sources, targets))
}

// NewPermutationConstraint creates a new permutation
func NewPermutationConstraint(handle string, context schema.ModuleId, targets []schema.RegisterId,
	sources []schema.RegisterId) Constraint {
	return newAir(constraint.NewPermutationConstraint(handle, context, targets, sources))
}

// NewRangeConstraint constructs a new AIR range constraint
func NewRangeConstraint(handle string, ctx schema.ModuleId, expr ColumnAccess, bitwidth uint) RangeConstraint {
	return newAir(constraint.NewRangeConstraint(handle, ctx, &expr, bitwidth))
}

// NewVanishingConstraint constructs a new AIR vanishing constraint
func NewVanishingConstraint(handle string, ctx schema.ModuleId, domain util.Option[int],
	term Term) VanishingConstraint {
	//
	return newAir(constraint.NewVanishingConstraint(handle, ctx, domain, LogicalTerm{term}))
}

// Air marks the constraint as being valid for the AIR representation.
func (p Air[C]) Air() {
	// nothing as just a marker.
}

// Accepts determines whether a given constraint accepts a given trace or
// not.  If not, a failure is produced.  Otherwise, a bitset indicating
// branch coverage is returned.
func (p Air[C]) Accepts(trace trace.Trace) (bit.Set, schema.Failure) {
	return p.constraint.Accepts(trace)
}

// Bounds determines the well-definedness bounds for this constraint in both the
// negative (left) or positive (right) directions.  For example, consider an
// expression such as "(shift X -1)".  This is technically undefined for the
// first row of any trace and, by association, any constraint evaluating this
// expression on that first row is also undefined (and hence must pass)
func (p Air[C]) Bounds(module uint) util.Bounds {
	return p.constraint.Bounds(module)
}

// Consistent applies a number of internal consistency checks.  Whilst not
// strictly necessary, these can highlight otherwise hidden problems as an aid
// to debugging.
func (p Air[C]) Consistent(schema schema.AnySchema) []error {
	return p.constraint.Consistent(schema)
}

// Contexts returns the evaluation contexts (i.e. enclosing module + length
// multiplier) for this constraint.  Most constraints have only a single
// evaluation context, though some (e.g. lookups) have more.  Note that all
// constraints have at least one context (which we can call the "primary"
// context).
func (p Air[C]) Contexts() []schema.ModuleId {
	return p.constraint.Contexts()
}

// Name returns a unique name and case number for a given constraint.  This
// is useful purely for identifying constraints in reports, etc.
func (p Air[C]) Name() string {
	return p.constraint.Name()
}

// Lisp converts this schema element into a simple S-Expression, for example
// so it can be printed.
//
//nolint:revive
func (p Air[C]) Lisp(schema schema.AnySchema) sexp.SExp {
	return p.constraint.Lisp(schema)
}
