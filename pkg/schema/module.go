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
)

// ModuleId abstracts the notion of a "module identifier"
type ModuleId = uint

// Module represents a "table" within a schema which contains zero or more rows
// for a given set of registers.
type Module interface {
	// Assignments returns an iterator over the assignments of this module.
	// These are the computations used to assign values to all computed columns
	// in this module.
	Assignments() iter.Iterator[Assignment]
	// Constraints provides access to those constraints associated with this
	// module.
	Constraints() iter.Iterator[Constraint]
	// Consistent applies a number of internal consistency checks.  Whilst not
	// strictly necessary, these can highlight otherwise hidden problems as an aid
	// to debugging.
	Consistent(Schema[Constraint]) []error
	// HasRegister checks whether a register with the given name exists and, if
	// so, returns its register identifier.  Otherwise, it returns false.
	HasRegister(name string) (RegisterId, bool)
	// Identifies the length multiplier for this module.  For every trace, the
	// height of the corresponding module must be a multiple of this.  This is
	// used specifically to support interleaving constraints.
	LengthMultiplier() uint
	// Module name
	Name() string
	// Access a given register in this module.
	Register(RegisterId) Register
	// Registers providers access to the underlying registers of this schema.
	Registers() []Register
	// Returns the number of registers in this module.
	Width() uint
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
type Table[C Constraint] struct {
	name        string
	multiplier  uint
	registers   []Register
	constraints []C
	assignments []Assignment
}

// NewTable constructs a table module with the given registers and constraints.
func NewTable[C Constraint](name string, multiplier uint) Table[C] {
	return Table[C]{name, multiplier, nil, nil, nil}
}

// Assignments provides access to those assignments defined as part of this
// table.
func (p Table[C]) Assignments() iter.Iterator[Assignment] {
	return iter.NewArrayIterator(p.assignments)
}

// Constraints provides access to those constraints associated with this
// module.
func (p Table[C]) Constraints() iter.Iterator[Constraint] {
	arrIter := iter.NewArrayIterator(p.constraints)
	return iter.NewCastIterator[C, Constraint](arrIter)
}

// Consistent applies a number of internal consistency checks.  Whilst not
// strictly necessary, these can highlight otherwise hidden problems as an aid
// to debugging.
func (p Table[C]) Consistent(schema Schema[Constraint]) []error {
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
func (p Table[C]) HasRegister(name string) (RegisterId, bool) {
	for i := range p.Width() {
		if p.registers[i].Name == name {
			return NewRegisterId(i), true
		}
	}
	// Fail
	return NewUnusedRegisterId(), false
}

// Name returns the module name.
func (p Table[C]) Name() string {
	return p.name
}

// LengthMultiplier identifies the length multiplier for this module.  For every
// trace, the height of the corresponding module must be a multiple of this.
// This is used specifically to support interleaving constraints.
func (p Table[C]) LengthMultiplier() uint {
	return p.multiplier
}

// Register returns the given register in this table.
func (p Table[C]) Register(id RegisterId) Register {
	return p.registers[id.Unwrap()]
}

// Registers returns an iterator over the underlying registers of this schema.
// Specifically, the index of a register in this array is its register index.
func (p Table[C]) Registers() []Register {
	return p.registers
}

// Subdivide implementation for the FieldAgnosticModule interface.
func (p Table[C]) Subdivide(bandwidth uint, maxRegisterWidth uint) Table[C] {
	var (
		registers   []Register
		constraints []C
		assignments []Assignment
	)
	// Check registers
	for _, r := range p.registers {
		if r.Width > maxRegisterWidth {
			panic(fmt.Sprintf("maximum register width exceeded (%d > %d)", r.Width, maxRegisterWidth))
		}
		//
		registers = append(registers, r)
	}
	// Subdivide assignments
	for _, c := range p.assignments {
		var a any = c
		//nolint
		if fc, ok := a.(FieldAgnostic[Assignment]); ok {
			assignments = append(assignments, fc.Subdivide(bandwidth, maxRegisterWidth))
		} else {
			panic(fmt.Sprintf("non-field agnostic assignment (%s)", reflect.TypeOf(a).String()))
		}
	}
	// Subdivide constraints
	for _, c := range p.constraints {
		var a any = c
		//nolint
		if fc, ok := a.(FieldAgnostic[C]); ok {
			constraints = append(constraints, fc.Subdivide(bandwidth, maxRegisterWidth))
		} else {
			panic(fmt.Sprintf("non-field agnostic constraint (%s)", reflect.TypeOf(a).String()))
		}
	}
	//
	return Table[C]{p.name, p.multiplier, registers, constraints, assignments}
}

// Width returns the number of registers in this Table.
func (p Table[C]) Width() uint {
	return uint(len(p.registers))
}

// ============================================================================
// Mutators
// ============================================================================

// AddAssignments adds a new assignments to this table.
func (p *Table[C]) AddAssignments(assignments ...Assignment) {
	p.assignments = append(p.assignments, assignments...)
}

// AddConstraints adds new constraints to this table.
func (p *Table[C]) AddConstraints(constraints ...C) {
	p.constraints = append(p.constraints, constraints...)
}

// AddRegisters adds new registers to this table.
func (p *Table[C]) AddRegisters(registers ...Register) {
	// Add registers
	p.registers = append(p.registers, registers...)
}

// ============================================================================
// Encoding / Decoding
// ============================================================================

// GobEncode an option.  This allows it to be marshalled into a binary form.
func (p Table[M]) GobEncode() (data []byte, err error) {
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
func (p *Table[M]) GobDecode(data []byte) error {
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
