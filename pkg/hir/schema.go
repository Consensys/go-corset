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
	"bytes"
	"encoding/gob"
	"fmt"

	"github.com/consensys/gnark-crypto/ecc/bls12-377/fr"
	"github.com/consensys/go-corset/pkg/schema"
	sc "github.com/consensys/go-corset/pkg/schema"
	"github.com/consensys/go-corset/pkg/schema/assignment"
	"github.com/consensys/go-corset/pkg/schema/constraint"
	"github.com/consensys/go-corset/pkg/trace"
	"github.com/consensys/go-corset/pkg/util"
	"github.com/consensys/go-corset/pkg/util/collection/iter"
)

// DataColumn captures the essence of a data column at AIR level.
type DataColumn = *assignment.DataColumn

// VanishingConstraint captures the essence of a vanishing constraint at the HIR
// level. A vanishing constraint is a row constraint which must evaluate to
// zero.
type VanishingConstraint = *constraint.VanishingConstraint[ZeroArrayTest]

// LookupConstraint captures the essence of a lookup constraint at the HIR
// level.  To make this work, the UnitExpr adaptor is required, and this means
// certain expression forms cannot be permitted (e.g. the use of lists).
type LookupConstraint = *constraint.LookupConstraint[UnitExpr]

// LookupVector captures the essence of either the source or target for a
// lookup.
type LookupVector = constraint.LookupVector[UnitExpr]

// RangeConstraint captures the essence of a range constraints at the HIR level.
type RangeConstraint = *constraint.RangeConstraint[MaxExpr]

// SortedConstraint captures the essence of a sorted constraints at the HIR level.
type SortedConstraint = *constraint.SortedConstraint[UnitExpr]

// PropertyAssertion captures the notion of an arbitrary property which should
// hold for all acceptable traces.  However, such a property is not enforced by
// the prover.
type PropertyAssertion = *sc.PropertyAssertion[ZeroArrayTest]

// Permutation captures the notion of a (sorted) permutation at the HIR level.
type Permutation = *assignment.SortedPermutation

// Schema for HIR constraints and columns.
type Schema struct {
	// The modules of the schema
	modules []sc.Module
	// The data columns of this schema.
	inputs []sc.Declaration
	// The sorted permutations of this schema.
	assignments []sc.Assignment
	// Constraints of this schema, which are either vanishing, lookup or type
	// constraints.
	constraints []sc.Constraint
	// Property assertions for this schema.
	assertions []PropertyAssertion
	// Cache list of columns declared in inputs and assignments.
	column_cache []sc.Column
}

// EmptySchema is used to construct a fresh schema onto which new columns and
// constraints will be added.
func EmptySchema() *Schema {
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

// AddDataColumn appends a new data column with a given type.  Furthermore, the
// type is enforced by the system when checking is enabled.
func (p *Schema) AddDataColumn(context trace.Context, name string, base sc.Type) uint {
	if context.Module() >= uint(len(p.modules)) {
		panic(fmt.Sprintf("invalid module index (%d)", context.Module()))
	}

	cid := uint(len(p.inputs))
	col := assignment.NewDataColumn(context, name, base)
	p.inputs = append(p.inputs, col)
	// Update column cache
	for c := col.Columns(); c.HasNext(); {
		p.column_cache = append(p.column_cache, c.Next())
	}

	return cid
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
	// Done
	return index
}

// AddLookupConstraint appends a new lookup constraint.
func (p *Schema) AddLookupConstraint(handle string, source LookupVector, target LookupVector) {
	if target.Len() != source.Len() {
		panic("differeng number of target / source lookup columns")
	}
	// TODO: sanity source columns are in the source module, and likewise target
	// columns are in the target module (though source != target is permitted).

	// Finally add constraint
	p.constraints = append(p.constraints,
		constraint.NewLookupConstraint(handle, source, target))
}

// AddVanishingConstraint appends a new vanishing constraint.
func (p *Schema) AddVanishingConstraint(handle string, context trace.Context, domain util.Option[int], expr Expr) {
	if context.Module() >= uint(len(p.modules)) {
		panic(fmt.Sprintf("invalid module index (%d)", context.Module()))
	}

	p.constraints = append(p.constraints,
		constraint.NewVanishingConstraint(handle, 0, context, domain, ZeroArrayTest{expr}))
}

// AddRangeConstraint appends a new range constraint with a raw bound.
func (p *Schema) AddRangeConstraint(handle string, context trace.Context, expr Expr, bitwidth uint) {
	// Check whether is a field type, as these can actually be ignored.
	maxExpr := MaxExpr{expr}
	p.constraints = append(p.constraints,
		constraint.NewRangeConstraint[MaxExpr](handle, 0, context, maxExpr, bitwidth))
}

// AddSortedConstraint appends a new sorted constraint.
func (p *Schema) AddSortedConstraint(handle string, context trace.Context, bitwidth uint,
	selector util.Option[UnitExpr], sources []UnitExpr, signs []bool, strict bool) {
	// Finally add constraint
	p.constraints = append(p.constraints,
		constraint.NewSortedConstraint(handle, context, bitwidth, selector, sources, signs, strict))
}

// AddPropertyAssertion appends a new property assertion.
func (p *Schema) AddPropertyAssertion(handle string, context trace.Context, property Expr) {
	p.assertions = append(p.assertions, sc.NewPropertyAssertion[ZeroArrayTest](handle, context, ZeroArrayTest{property}))
}

// SubstituteConstants substitutes the value of matching labelled constants for
// all expressions used within the schema.
func (p *Schema) SubstituteConstants(mapping map[string]fr.Element) {
	// Constraints
	for _, a := range p.constraints {
		substituteConstraint(mapping, a)
	}
	// Assertions
	for _, a := range p.assertions {
		substituteConstraint(mapping, a)
	}
}

// ============================================================================
// Consistency Check
// ============================================================================

// CheckConsistency performs some simple checks that the given schema is
// consistent.  This provides a double check of certain key properties, such as
// that registers used for assignments are large enough, etc.
func (p *Schema) CheckConsistency() error {
	// For now, the only consistency check is for assignments.  More could be
	// done here.
	for _, a := range p.assignments {
		if err := a.CheckConsistency(p); err != nil {
			return err
		}
	}
	//
	return nil
}

// ============================================================================
// Schema Interface
// ============================================================================

// InputColumns returns an array over the input columns of this schema.  That
// is, the subset of columns whose trace values must be provided by the
// user.
func (p *Schema) InputColumns() iter.Iterator[sc.Column] {
	inputs := iter.NewArrayIterator(p.inputs)
	return iter.NewFlattenIterator[schema.Declaration, schema.Column](inputs,
		func(d schema.Declaration) iter.Iterator[schema.Column] { return d.Columns() })
}

// Assertions returns an iterator over the property assertions of this
// schema.  These are properties which should hold true for any valid trace
// (though, of course, may not hold true for an invalid trace).
func (p *Schema) Assertions() iter.Iterator[schema.Constraint] {
	properties := iter.NewArrayIterator(p.assertions)
	return iter.NewCastIterator[PropertyAssertion, schema.Constraint](properties)
}

// Assignments returns an array over the assignments of this sc.  That
// is, the subset of declarations whose trace values can be computed from
// the inputs.
func (p *Schema) Assignments() iter.Iterator[sc.Assignment] {
	return iter.NewArrayIterator(p.assignments)
}

// Columns returns an array over the underlying columns of this sc.
// Specifically, the index of a column in this array is its column index.
func (p *Schema) Columns() iter.Iterator[sc.Column] {
	return iter.NewArrayIterator(p.column_cache)
}

// Constraints returns an array over the underlying constraints of this
// sc.
func (p *Schema) Constraints() iter.Iterator[sc.Constraint] {
	return iter.NewArrayIterator(p.constraints)
}

// Declarations returns an array over the column declarations of this
// sc.
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

// ============================================================================
// Encoding / Decoding
// ============================================================================

// GobEncode an HIR schema.  This allows it to be marshalled into a binary form.
func (p *Schema) GobEncode() (data []byte, err error) {
	var buffer bytes.Buffer
	gobEncoder := gob.NewEncoder(&buffer)
	// Modules
	if err := gobEncoder.Encode(p.modules); err != nil {
		return nil, err
	}
	// Inputs
	if err := gobEncoder.Encode(p.inputs); err != nil {
		return nil, err
	}
	// Assignments
	if err := gobEncoder.Encode(p.assignments); err != nil {
		return nil, err
	}
	// Constraints
	if err := gobEncoder.Encode(p.constraints); err != nil {
		return nil, err
	}
	// Assertions
	if err := gobEncoder.Encode(p.assertions); err != nil {
		return nil, err
	}
	// Success
	return buffer.Bytes(), nil
}

// GobDecode a previously encoded schema
func (p *Schema) GobDecode(data []byte) error {
	buffer := bytes.NewBuffer(data)
	gobDecoder := gob.NewDecoder(buffer)
	// Modules
	if err := gobDecoder.Decode(&p.modules); err != nil {
		return err
	}
	// Inputs
	if err := gobDecoder.Decode(&p.inputs); err != nil {
		return err
	}
	// Assignments
	if err := gobDecoder.Decode(&p.assignments); err != nil {
		return err
	}
	// Constraints
	if err := gobDecoder.Decode(&p.constraints); err != nil {
		return err
	}
	// Assertions
	if err := gobDecoder.Decode(&p.assertions); err != nil {
		return err
	}
	// Rebuild column cache
	p.rebuildCaches()
	// Success
	return nil
}

func (p *Schema) rebuildCaches() {
	// Add all inputs
	for _, col := range p.inputs {
		for c := col.Columns(); c.HasNext(); {
			p.column_cache = append(p.column_cache, c.Next())
		}
	}
	// Add all assignments
	for _, col := range p.assignments {
		for c := col.Columns(); c.HasNext(); {
			p.column_cache = append(p.column_cache, c.Next())
		}
	}
}

// ============================================================================
// Encoding / Decoding
// ============================================================================

func init() {
	gob.Register(sc.Constraint(&constraint.VanishingConstraint[ZeroArrayTest]{}))
	gob.Register(sc.Constraint(&constraint.RangeConstraint[MaxExpr]{}))
	gob.Register(sc.Constraint(&constraint.PermutationConstraint{}))
	gob.Register(sc.Constraint(&constraint.LookupConstraint[UnitExpr]{}))
	gob.Register(sc.Constraint(&constraint.SortedConstraint[UnitExpr]{}))
}
