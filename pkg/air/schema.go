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
	"fmt"

	"github.com/consensys/gnark-crypto/ecc/bls12-377/fr"

	sc "github.com/consensys/go-corset/pkg/schema"
	"github.com/consensys/go-corset/pkg/schema/assignment"
	"github.com/consensys/go-corset/pkg/schema/constraint"
	"github.com/consensys/go-corset/pkg/trace"
	"github.com/consensys/go-corset/pkg/util"
	"github.com/consensys/go-corset/pkg/util/collection/iter"
)

// DataColumn captures the essence of a data column at AIR level.
type DataColumn = *assignment.DataColumn

// LookupConstraint captures the essence of a lookup constraint at the AIR
// level.  At the AIR level, lookup constraints are only permitted between
// columns (i.e. not arbitrary expressions).
type LookupConstraint = *constraint.LookupConstraint[*ColumnAccess]

// VanishingConstraint captures the essence of a vanishing constraint at the AIR level.
type VanishingConstraint = *constraint.VanishingConstraint[Expr]

// RangeConstraint captures the essence of a range constraints at the AIR level.
type RangeConstraint = *constraint.RangeConstraint[*ColumnAccess]

// PermutationConstraint captures the essence of a permutation constraint at the AIR level.
// Specifically, it represents a constraint that one (or more) columns are a permutation of another.
type PermutationConstraint = *constraint.PermutationConstraint

// PropertyAssertion captures the notion of an arbitrary property which should
// hold for all acceptable traces.  However, such a property is not enforced by
// the prover.
type PropertyAssertion = *sc.PropertyAssertion[sc.Testable]

// Schema for AIR traces which is parameterised on a notion of computation as
// permissible in computed columns.
type Schema struct {
	// The modules of the schema
	modules []sc.Module
	// The set of data columns corresponding to the inputs of this schema.
	inputs []sc.Declaration
	// Assignments defines the set of column declarations whose trace values are
	// computed from the inputs.
	assignments []sc.Assignment
	// The constraints of this schema.  A constraint is either a vanishing
	// constraint, a permutation constraint, a lookup constraint or a range
	// constraint.
	constraints []sc.Constraint
	// Property assertions.
	assertions []PropertyAssertion
	// Cache list of columns declared in inputs and assignments.
	column_cache []sc.Column
}

// EmptySchema is used to construct a fresh schema onto which new columns and
// constraints will be added.
func EmptySchema[C sc.Evaluable]() *Schema {
	p := new(Schema)
	p.modules = make([]sc.Module, 0)
	p.inputs = make([]sc.Declaration, 0)
	p.assignments = make([]sc.Assignment, 0)
	p.constraints = make([]sc.Constraint, 0)
	p.assertions = make([]PropertyAssertion, 0)
	p.column_cache = make([]sc.Column, 0)
	// Done
	return p
}

// AddModule adds a new module to this schema, returning its module index.
func (p *Schema) AddModule(name string) uint {
	mid := uint(len(p.modules))
	p.modules = append(p.modules, sc.NewModule(name))

	return mid
}

// AddColumn appends a new data column whose values must be provided by the
// user.
func (p *Schema) AddColumn(context trace.Context, name string, datatype sc.Type) uint {
	if context.Module() >= uint(len(p.modules)) {
		panic(fmt.Sprintf("invalid module index (%d)", context.Module()))
	}

	col := assignment.NewDataColumn(context, name, datatype)
	// NOTE: the air level has no ability to enforce the type specified for a
	// given column.
	p.inputs = append(p.inputs, col)
	// Update column cache
	for c := col.Columns(); c.HasNext(); {
		p.column_cache = append(p.column_cache, c.Next())
	}
	// Calculate column index
	return uint(len(p.inputs) - 1)
}

// AddAssignment appends a new assignment (i.e. set of computed columns) to be
// used during trace expansion for this schema.  Computed columns are introduced
// by the process of lowering from HIR / MIR to AIR.
func (p *Schema) AddAssignment(c sc.Assignment) uint {
	index := p.Columns().Count()
	p.assignments = append(p.assignments, c)
	// Update column cache
	for c := c.Columns(); c.HasNext(); {
		p.column_cache = append(p.column_cache, c.Next())
	}

	return index
}

// AddLookupConstraint appends a new lookup constraint.
func (p *Schema) AddLookupConstraint(handle string, source trace.Context,
	target trace.Context, sources []uint, targets []uint) {
	if len(targets) != len(sources) {
		panic("differeng number of target / source lookup columns")
	}
	// TODO: sanity source columns are in the same module, and likewise target
	// columns (though they don't have to be in the same column together).
	from := make([]*ColumnAccess, len(sources))
	into := make([]*ColumnAccess, len(targets))
	// Construct column accesses from column indices.
	for i := 0; i < len(from); i++ {
		from[i] = &ColumnAccess{Column: sources[i], Shift: 0}
		into[i] = &ColumnAccess{Column: targets[i], Shift: 0}
	}
	// Construct lookup constraint
	var lookup LookupConstraint = constraint.NewLookupConstraint(handle, source, target, from, into)
	// Add
	p.constraints = append(p.constraints, lookup)
}

// AddPermutationConstraint appends a new permutation constraint which
// ensures that one column is a permutation of another.
func (p *Schema) AddPermutationConstraint(handle string, context trace.Context, targets []uint, sources []uint) {
	// TODO: sanity target and source columns are in the same module.
	p.constraints = append(p.constraints, constraint.NewPermutationConstraint(handle, context, targets, sources))
}

// AddPropertyAssertion appends a new property assertion.
func (p *Schema) AddPropertyAssertion(handle string, context trace.Context, assertion sc.Testable) {
	p.assertions = append(p.assertions, sc.NewPropertyAssertion(handle, context, assertion))
}

// AddVanishingConstraint appends a new vanishing constraint.
func (p *Schema) AddVanishingConstraint(handle string, context trace.Context, domain util.Option[int], expr Expr) {
	if context.Module() >= uint(len(p.modules)) {
		panic(fmt.Sprintf("invalid module index (%d)", context.Module()))
	}
	// TODO: sanity check expression enclosed by module
	p.constraints = append(p.constraints,
		constraint.NewVanishingConstraint(handle, context, domain, expr))
}

// AddRangeConstraint appends a new range constraint.
func (p *Schema) AddRangeConstraint(column uint, bound fr.Element) {
	col := p.Columns().Nth(column)
	handle := col.QualifiedName(p)
	tc := constraint.NewRangeConstraint(handle, col.Context, &ColumnAccess{Column: column, Shift: 0}, bound)
	p.constraints = append(p.constraints, tc)
}

// ============================================================================
// Schema Interface
// ============================================================================

// InputColumns returns an array over the input columns of this schema.  That
// is, the subset of columns whose trace values must be provided by the
// user.
func (p *Schema) InputColumns() iter.Iterator[sc.Column] {
	inputs := iter.NewArrayIterator(p.inputs)
	return iter.NewFlattenIterator[sc.Declaration, sc.Column](inputs,
		func(d sc.Declaration) iter.Iterator[sc.Column] { return d.Columns() })
}

// Assertions returns an iterator over the property assertions of this
// schema.  These are properties which should hold true for any valid trace
// (though, of course, may not hold true for an invalid trace).
func (p *Schema) Assertions() iter.Iterator[sc.Constraint] {
	properties := iter.NewArrayIterator(p.assertions)
	return iter.NewCastIterator[PropertyAssertion, sc.Constraint](properties)
}

// Assignments returns an array over the assignments of this schema.  That
// is, the subset of declarations whose trace values can be computed from
// the inputs.
func (p *Schema) Assignments() iter.Iterator[sc.Assignment] {
	return iter.NewArrayIterator(p.assignments)
}

// Columns returns an array over the underlying columns of this schema.
// Specifically, the index of a column in this array is its column index.
func (p *Schema) Columns() iter.Iterator[sc.Column] {
	return iter.NewArrayIterator(p.column_cache)
}

// Constraints returns an array over the underlying constraints of this
// schema.
func (p *Schema) Constraints() iter.Iterator[sc.Constraint] {
	return iter.NewArrayIterator(p.constraints)
}

// Declarations returns an array over the column declarations of this
// schema.
func (p *Schema) Declarations() iter.Iterator[sc.Declaration] {
	inputs := iter.NewArrayIterator(p.inputs)
	ps := iter.NewCastIterator[sc.Assignment, sc.Declaration](p.Assignments())

	return inputs.Append(ps)
}

// Modules returns an iterator over the declared set of modules within this
// schema.
func (p *Schema) Modules() iter.Iterator[sc.Module] {
	return iter.NewArrayIterator(p.modules)
}
