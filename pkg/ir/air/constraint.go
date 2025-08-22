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
	"github.com/consensys/go-corset/pkg/schema/constraint/interleaving"
	"github.com/consensys/go-corset/pkg/schema/constraint/lookup"
	"github.com/consensys/go-corset/pkg/schema/constraint/permutation"
	"github.com/consensys/go-corset/pkg/schema/constraint/ranged"
	"github.com/consensys/go-corset/pkg/schema/constraint/vanishing"
	"github.com/consensys/go-corset/pkg/trace"
	"github.com/consensys/go-corset/pkg/util"
	"github.com/consensys/go-corset/pkg/util/collection/bit"
	"github.com/consensys/go-corset/pkg/util/field"
	"github.com/consensys/go-corset/pkg/util/source/sexp"
)

// ============================================================================
// Helpers
// ============================================================================

// ConstraintBound limits the permitted set of underlying constraints.  This
// should never change, unless the underlying prover changes in some way to
// offer different or more fundamental primitives.
type ConstraintBound[F field.Element[F]] interface {
	schema.Constraint[F]

	constraint.Assertion[F, ir.Testable[F]] |
		interleaving.Constraint[F, *ColumnAccess[F]] |
		lookup.Constraint[F, *ColumnAccess[F]] |
		permutation.Constraint[F] |
		ranged.Constraint[F, *ColumnAccess[F]] |
		vanishing.Constraint[F, LogicalTerm[F]]
}

// Air attempts to encapsulate the notion of a valid constraint at the AIR
// level.  Since this is the fundamental level, only certain constraint forms
// are permitted.  As such, we want to try and ensure that arbitrary constraints
// are not found at the Air level.
type Air[F field.Element[F], C ConstraintBound[F]] struct {
	constraint C
}

// newAir is a helper method for the various constraint constructors, basically
// to avoid lots of generic types.
func newAir[F field.Element[F], C ConstraintBound[F]](constraint C) Air[F, C] {
	return Air[F, C]{constraint}
}

// NewAssertion constructs a new AIR assertion
func NewAssertion[F field.Element[F]](handle string, ctx schema.ModuleId, term ir.Testable[F]) Assertion[F] {
	//
	return newAir(constraint.NewAssertion(handle, ctx, term))
}

// NewInterleavingConstraint creates a new interleaving constraint with a given handle.
func NewInterleavingConstraint[F field.Element[F]](handle string, targetContext schema.ModuleId,
	sourceContext schema.ModuleId, target ColumnAccess[F], sources []*ColumnAccess[F]) Constraint[F] {
	return newAir(interleaving.NewConstraint(handle, targetContext, sourceContext, &target, sources))
}

// NewLookupConstraint constructs a new AIR lookup constraint
func NewLookupConstraint[F field.Element[F]](handle string, targets []lookup.Vector[F, *ColumnAccess[F]],
	sources []lookup.Vector[F, *ColumnAccess[F]]) LookupConstraint[F] {
	//
	return newAir(lookup.NewConstraint(handle, targets, sources))
}

// NewPermutationConstraint creates a new permutation
func NewPermutationConstraint[F field.Element[F]](handle string, context schema.ModuleId, targets []schema.RegisterId,
	sources []schema.RegisterId) Constraint[F] {
	return newAir(permutation.NewConstraint[F](handle, context, targets, sources))
}

// NewRangeConstraint constructs a new AIR range constraint
func NewRangeConstraint[F field.Element[F]](handle string, ctx schema.ModuleId, expr ColumnAccess[F],
	bitwidth uint) RangeConstraint[F] {
	//
	return newAir(ranged.NewConstraint(handle, ctx, &expr, bitwidth))
}

// NewVanishingConstraint constructs a new AIR vanishing constraint
func NewVanishingConstraint[F field.Element[F]](handle string, ctx schema.ModuleId, domain util.Option[int],
	term Term[F]) VanishingConstraint[F] {
	//
	return newAir(vanishing.NewConstraint(handle, ctx, domain, LogicalTerm[F]{term}))
}

// Air marks the constraint as being valid for the AIR representation.
func (p Air[F, C]) Air() {
	// nothing as just a marker.
}

// Accepts determines whether a given constraint accepts a given trace or
// not.  If not, a failure is produced.  Otherwise, a bitset indicating
// branch coverage is returned.
func (p Air[F, C]) Accepts(trace trace.Trace[F], schema schema.AnySchema[F],
) (bit.Set, schema.Failure) {
	return p.constraint.Accepts(trace, schema)
}

// Bounds determines the well-definedness bounds for this constraint in both the
// negative (left) or positive (right) directions.  For example, consider an
// expression such as "(shift X -1)".  This is technically undefined for the
// first row of any trace and, by association, any constraint evaluating this
// expression on that first row is also undefined (and hence must pass)
func (p Air[F, C]) Bounds(module uint) util.Bounds {
	return p.constraint.Bounds(module)
}

// Consistent applies a number of internal consistency checks.  Whilst not
// strictly necessary, these can highlight otherwise hidden problems as an aid
// to debugging.
func (p Air[F, C]) Consistent(schema schema.AnySchema[F]) []error {
	return p.constraint.Consistent(schema)
}

// Contexts returns the evaluation contexts (i.e. enclosing module + length
// multiplier) for this constraint.  Most constraints have only a single
// evaluation context, though some (e.g. lookups) have more.  Note that all
// constraints have at least one context (which we can call the "primary"
// context).
func (p Air[F, C]) Contexts() []schema.ModuleId {
	return p.constraint.Contexts()
}

// Name returns a unique name and case number for a given constraint.  This
// is useful purely for identifying constraints in reports, etc.
func (p Air[F, C]) Name() string {
	return p.constraint.Name()
}

// Lisp converts this schema element into a simple S-Expression, for example
// so it can be printed.
//
//nolint:revive
func (p Air[F, C]) Lisp(schema schema.AnySchema[F]) sexp.SExp {
	return p.constraint.Lisp(schema)
}

// Substitute any matchined labelled constants within this constraint
func (p Air[F, C]) Substitute(map[string]F) {
	// This should never be called since AIR expressions cannot contain labelled
	// constants.
	panic("unreachable")
}

// Unwrap provides access to the underlying constraint.
func (p Air[F, C]) Unwrap() C {
	return p.constraint
}
