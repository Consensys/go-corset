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
	"github.com/consensys/gnark-crypto/ecc/bls12-377/fr"
	"github.com/consensys/go-corset/pkg/schema"
	"github.com/consensys/go-corset/pkg/schema/constraint"
	"github.com/consensys/go-corset/pkg/trace"
	"github.com/consensys/go-corset/pkg/util"
	"github.com/consensys/go-corset/pkg/util/collection/bit"
	"github.com/consensys/go-corset/pkg/util/source/sexp"
)

// Constraint attempts to encapsulate the notion of a valid constraint at the MIR
// level.  Since this is the fundamental level, only certain constraint forms
// are permitted.  As such, we want to try and ensure that arbitrary constraints
// are not found at the Constraint level.
type Constraint struct {
	constraint schema.Constraint
}

// NewAssertion constructs a new assertion
func NewAssertion(handle string, ctx schema.ModuleId, term LogicalTerm) Constraint {
	//
	return Constraint{constraint.NewAssertion(handle, ctx, term)}
}

// NewVanishingConstraint constructs a new vanishing constraint
func NewVanishingConstraint(handle string, ctx schema.ModuleId, domain util.Option[int],
	term LogicalTerm) Constraint {
	//
	return Constraint{constraint.NewVanishingConstraint(handle, ctx, domain, term)}
}

// NewInterleavingConstraint creates a new interleaving constraint with a given handle.
func NewInterleavingConstraint(handle string, targetContext schema.ModuleId,
	sourceContext schema.ModuleId, target Term, sources []Term) Constraint {
	return Constraint{constraint.NewInterleavingConstraint(handle, targetContext, sourceContext, target, sources)}
}

// NewLookupConstraint creates a new lookup constraint with a given handle.
func NewLookupConstraint(handle string, targetContext schema.ModuleId, targets []Term,
	sourceContext schema.ModuleId, sources []Term) Constraint {
	//
	if len(targets) != len(sources) {
		panic("differeng number of targetContext / source lookup columns")
	}

	return Constraint{constraint.NewLookupConstraint(handle, targetContext, targets, sourceContext, sources)}
}

// NewPermutationConstraint creates a new permutation
func NewPermutationConstraint(handle string, context schema.ModuleId, targets []schema.RegisterId,
	sources []schema.RegisterId) Constraint {
	return Constraint{constraint.NewPermutationConstraint(handle, context, targets, sources)}
}

// NewRangeConstraint constructs a new Range constraint!
func NewRangeConstraint(handle string, ctx schema.ModuleId, expr Term, bitwidth uint) Constraint {
	return Constraint{constraint.NewRangeConstraint(handle, ctx, expr, bitwidth)}
}

// NewSortedConstraint creates a new Sorted
func NewSortedConstraint(handle string, context schema.ModuleId, bitwidth uint, selector util.Option[Term],
	sources []Term, signs []bool, strict bool) Constraint {
	//
	return Constraint{constraint.NewSortedConstraint(handle, context, bitwidth, selector, sources, signs, strict)}
}

// Accepts determines whether a given constraint accepts a given trace or
// not.  If not, a failure is produced.  Otherwise, a bitset indicating
// branch coverage is returned.
func (p Constraint) Accepts(trace trace.Trace) (bit.Set, schema.Failure) {
	return p.constraint.Accepts(trace)
}

// Bounds determines the well-definedness bounds for this constraint in both the
// negative (left) or positive (right) directions.  For example, consider an
// expression such as "(shift X -1)".  This is technically undefined for the
// first row of any trace and, by association, any constraint evaluating this
// expression on that first row is also undefined (and hence must pass)
func (p Constraint) Bounds(module uint) util.Bounds {
	return p.constraint.Bounds(module)
}

// Consistent applies a number of internal consistency checks.  Whilst not
// strictly necessary, these can highlight otherwise hidden problems as an aid
// to debugging.
func (p Constraint) Consistent(schema schema.AnySchema) []error {
	return p.constraint.Consistent(schema)
}

// Contexts returns the evaluation contexts (i.e. enclosing module + length
// multiplier) for this constraint.  Most constraints have only a single
// evaluation context, though some (e.g. lookups) have more.  Note that all
// constraints have at least one context (which we can call the "primary"
// context).
func (p Constraint) Contexts() []schema.ModuleId {
	return p.constraint.Contexts()
}

// Name returns a unique name and case number for a given constraint.  This
// is useful purely for identifying constraints in reports, etc.
func (p Constraint) Name() string {
	return p.constraint.Name()
}

// Lisp converts this schema element into a simple S-Expression, for example
// so it can be printed.
//
//nolint:revive
func (p Constraint) Lisp(schema schema.AnySchema) sexp.SExp {
	return p.constraint.Lisp(schema)
}

// Subdivide implementation for the FieldAgnosticModule interface.
func (p Constraint) Subdivide(mapping schema.RegisterMappings) Constraint {
	var constraint schema.Constraint
	//
	switch c := p.constraint.(type) {
	case Assertion:
		constraint = subdivideAssertion(c, mapping)
	case InterleavingConstraint:
		constraint = subdivideInterleaving(c, mapping)
	case LookupConstraint:
		constraint = subdivideLookup(c, mapping)
	case PermutationConstraint:
		constraint = subdividePermutation(c, mapping)
	case RangeConstraint:
		constraint = subdivideRange(c, mapping)
	case SortedConstraint:
		constraint = subdivideSorted(c, mapping)
	case VanishingConstraint:
		modmap := mapping.Module(c.Context)
		constraint = subdivideVanishing(c, modmap)
	default:
		panic("unreachable")
	}
	//
	return Constraint{constraint}
}

// Substitute any matchined labelled constants within this constraint
func (p Constraint) Substitute(mapping map[string]fr.Element) {
	p.constraint.Substitute(mapping)
}

// ============================================================================
// Encoding / Decoding
// ============================================================================

// GobEncode an option.  This allows it to be marshalled into a binary form.
func (p Constraint) GobEncode() (data []byte, err error) {
	return encode_constraint(p.constraint)
}

// GobDecode a previously encoded option
func (p *Constraint) GobDecode(data []byte) error {
	var error error
	p.constraint, error = decode_constraint(data)
	//
	return error
}
