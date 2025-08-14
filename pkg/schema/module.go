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
package schema

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"reflect"

	"github.com/consensys/go-corset/pkg/util/collection/iter"
	"github.com/consensys/go-corset/pkg/util/field/bls12_377"
)

// ModuleMap provides a mapping from module identifiers (or names) to register
// maps.
type ModuleMap[T RegisterMap] interface {
	// Field returns the underlying field configuration used for this mapping.
	// This includes the field bandwidth (i.e. number of bits available in
	// underlying field) and the maximum register width (i.e. width at which
	// registers are capped).
	Field() FieldConfig
	// Module returns register mapping information for the given module.
	Module(ModuleId) T
	// ModuleOf returns register mapping information for the given module.
	ModuleOf(string) T
}

// ModuleId abstracts the notion of a "module identifier"
type ModuleId = uint

// Module represents a "table" within a schema which contains zero or more rows
// for a given set of registers.
type Module interface {
	RegisterMap
	// Assignments returns an iterator over the assignments of this module.
	// These are the computations used to assign values to all computed columns
	// in this module.
	Assignments() iter.Iterator[Assignment[bls12_377.Element]]
	// Constraints provides access to those constraints associated with this
	// module.
	Constraints() iter.Iterator[Constraint[bls12_377.Element]]
	// Consistent applies a number of internal consistency checks.  Whilst not
	// strictly necessary, these can highlight otherwise hidden problems as an aid
	// to debugging.
	Consistent(AnySchema[bls12_377.Element]) []error
	// Identifies the length multiplier for this module.  For every trace, the
	// height of the corresponding module must be a multiple of this.  This is
	// used specifically to support interleaving constraints.
	LengthMultiplier() uint
	// AllowPadding determines the amount of initial padding a module expects.
	AllowPadding() bool
	// Module name
	Name() string
	// Returns the number of registers in this module.
	Width() uint
}

// FieldAgnosticModule captures the notion of a module which is agnostic to the
// underlying field being used.
type FieldAgnosticModule[M Module] interface {
	Module
	FieldAgnostic[M]
}

// ============================================================================
//
// ============================================================================

// Table provides a straightforward, reusable module implementation.  There is
// nothing fancy here: we simply have a set of registers, constraints and
// assignments.  A table is a field agnostic module with a simple strategy of
// subdividing registers "in place".  For example, suppose we have registers X
// and Y (in that order) where both are to be halfed.  Then, the result is X'0,
// X'1, Y'0. Y'1 (in that order).  Hence, predicting the new register indices is
// relatively straightforward.
type Table[F any, C Constraint[F]] struct {
	name        string
	multiplier  uint
	padding     bool
	registers   []Register
	constraints []C
	assignments []Assignment[F]
}

// NewTable constructs a table module with the given registers and constraints.
func NewTable[F any, C Constraint[F]](name string, multiplier uint, padding bool) *Table[F, C] {
	return &Table[F, C]{name, multiplier, padding, nil, nil, nil}
}

// Assignments provides access to those assignments defined as part of this
// table.
func (p *Table[F, C]) Assignments() iter.Iterator[Assignment[F]] {
	return iter.NewArrayIterator(p.assignments)
}

// Constraints provides access to those constraints associated with this
// module.
func (p *Table[F, C]) Constraints() iter.Iterator[Constraint[F]] {
	arrIter := iter.NewArrayIterator(p.constraints)
	return iter.NewCastIterator[C, Constraint[F]](arrIter)
}

// Consistent applies a number of internal consistency checks.  Whilst not
// strictly necessary, these can highlight otherwise hidden problems as an aid
// to debugging.
func (p *Table[F, C]) Consistent(schema AnySchema[F]) []error {
	var errors []error
	// Check constraints
	for _, c := range p.constraints {
		errors = append(errors, c.Consistent(schema)...)
	}
	// Check assignments
	for _, a := range p.assignments {
		errors = append(errors, a.Consistent(schema)...)
	}
	// Done
	return errors
}

// HasRegister checks whether a register with the given name exists and, if
// so, returns its register identifier.  Otherwise, it returns false.
func (p *Table[F, C]) HasRegister(name string) (RegisterId, bool) {
	for i := range p.Width() {
		if p.registers[i].Name == name {
			return NewRegisterId(i), true
		}
	}
	// Fail
	return NewUnusedRegisterId(), false
}

// Name returns the module name.
func (p *Table[F, C]) Name() string {
	return p.name
}

// LengthMultiplier identifies the length multiplier for this module.  For every
// trace, the height of the corresponding module must be a multiple of this.
// This is used specifically to support interleaving constraints.
func (p *Table[F, C]) LengthMultiplier() uint {
	return p.multiplier
}

// AllowPadding determines whether the given module supports padding at the
// beginning of the module.  This is necessary because legacy modules expect an
// initial padding row, and allow defensive padding as well.
func (p *Table[F, C]) AllowPadding() bool {
	return p.padding
}

// Register returns the given register in this table.
func (p *Table[F, C]) Register(id RegisterId) Register {
	return p.registers[id.Unwrap()]
}

// Registers returns an iterator over the underlying registers of this schema.
// Specifically, the index of a register in this array is its register index.
func (p *Table[F, C]) Registers() []Register {
	return p.registers
}

// Subdivide implementation for the FieldAgnosticModule interface.
func (p *Table[F, C]) Subdivide(mapping LimbsMap) *Table[F, C] {
	var (
		modmap      = mapping.ModuleOf(p.name)
		registers   []Register
		constraints []C
		assignments []Assignment[F]
	)
	// Append mapping registers
	for i := range p.registers {
		rid := NewRegisterId(uint(i))
		//
		for _, limb := range modmap.LimbIds(rid) {
			registers = append(registers, modmap.Limb(limb))
		}
	}
	// Subdivide assignments
	for _, c := range p.assignments {
		var a any = c
		//nolint
		if fc, ok := a.(FieldAgnostic[Assignment[F]]); ok {
			assignments = append(assignments, fc.Subdivide(mapping))
		} else {
			panic(fmt.Sprintf("non-field agnostic assignment (%s)", reflect.TypeOf(a).String()))
		}
	}
	// Subdivide constraints
	for _, c := range p.constraints {
		var a any = c
		//nolint
		if fc, ok := a.(FieldAgnostic[C]); ok {
			constraints = append(constraints, fc.Subdivide(mapping))
		} else {
			panic(fmt.Sprintf("non-field agnostic constraint (%s)", reflect.TypeOf(a).String()))
		}
	}
	//
	return &Table[F, C]{p.name, p.multiplier, p.padding, registers, constraints, assignments}
}

// Width returns the number of registers in this Table.
func (p *Table[F, C]) Width() uint {
	return uint(len(p.registers))
}

// ============================================================================
// Mutators
// ============================================================================

// AddAssignments adds a new assignments to this table.
func (p *Table[F, C]) AddAssignments(assignments ...Assignment[F]) {
	p.assignments = append(p.assignments, assignments...)
}

// AddConstraints adds new constraints to this table.
func (p *Table[F, C]) AddConstraints(constraints ...C) {
	p.constraints = append(p.constraints, constraints...)
}

// AddRegisters adds new registers to this table.
func (p *Table[F, C]) AddRegisters(registers ...Register) {
	// Add registers
	p.registers = append(p.registers, registers...)
}

// ============================================================================
// Encoding / Decoding
// ============================================================================

// GobEncode an option.  This allows it to be marshalled into a binary form.
func (p *Table[F, M]) GobEncode() (data []byte, err error) {
	var buffer bytes.Buffer
	gobEncoder := gob.NewEncoder(&buffer)
	// Name
	if err := gobEncoder.Encode(p.name); err != nil {
		return nil, err
	}
	// Multiplier
	if err := gobEncoder.Encode(p.multiplier); err != nil {
		return nil, err
	}
	// Padding
	if err := gobEncoder.Encode(p.padding); err != nil {
		return nil, err
	}
	// registers
	if err := gobEncoder.Encode(p.registers); err != nil {
		return nil, err
	}
	// constraints
	if err := gobEncoder.Encode(p.constraints); err != nil {
		return nil, err
	}
	// assignments
	if err := gobEncoder.Encode(p.assignments); err != nil {
		return nil, err
	}
	// Done
	return buffer.Bytes(), nil
}

// GobDecode a previously encoded option
func (p *Table[F, M]) GobDecode(data []byte) error {
	buffer := bytes.NewBuffer(data)
	gobDecoder := gob.NewDecoder(buffer)
	// Name
	if err := gobDecoder.Decode(&p.name); err != nil {
		return err
	}
	// Multiplier
	if err := gobDecoder.Decode(&p.multiplier); err != nil {
		return err
	}
	// Padding
	if err := gobDecoder.Decode(&p.padding); err != nil {
		return err
	}
	// Registers
	if err := gobDecoder.Decode(&p.registers); err != nil {
		return err
	}
	// Constraints
	if err := gobDecoder.Decode(&p.constraints); err != nil {
		return err
	}
	// Assignments
	if err := gobDecoder.Decode(&p.assignments); err != nil {
		return err
	}
	// Success!
	return nil
}
