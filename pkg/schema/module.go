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

	"github.com/consensys/go-corset/pkg/schema/register"
	"github.com/consensys/go-corset/pkg/util/collection/iter"
	"github.com/consensys/go-corset/pkg/util/field"
)

// ModuleMap provides a mapping from module identifiers (or names) to register
// maps.
type ModuleMap[T register.Map] interface {
	fmt.Stringer
	// Field returns the underlying field configuration used for this mapping.
	// This includes the field bandwidth (i.e. number of bits available in
	// underlying field) and the maximum register width (i.e. width at which
	// registers are capped).
	Field() FieldConfig
	// Module returns register mapping information for the given module.
	Module(ModuleId) T
	// ModuleOf returns register mapping information for the given module.
	ModuleOf(string) T
	// Returns number of modules in this map
	Width() uint
}

// ModuleId abstracts the notion of a "module identifier"
type ModuleId = uint

// ModuleView provides access to certain structural information about a module.
type ModuleView interface {
	register.Map
	// Module name
	Name() string
	// IsPublic indicates whether or not this module is externally visible.
	IsPublic() bool
	// IsSynthetic modules are generated during compilation, rather than being
	// provided by the user.
	IsSynthetic() bool
	// Returns the number of registers in this module.
	Width() uint
}

// Module represents a "table" within a schema which contains zero or more rows
// for a given set of registers.
type Module[F any] interface {
	ModuleView
	// Assignments returns an iterator over the assignments of this module.
	// These are the computations used to assign values to all computed columns
	// in this module.
	Assignments() iter.Iterator[Assignment[F]]
	// Constraints provides access to those constraints associated with this
	// module.
	Constraints() iter.Iterator[Constraint[F]]
	// Consistent applies a number of internal consistency checks.  Whilst not
	// strictly necessary, these can highlight otherwise hidden problems as an aid
	// to debugging.
	Consistent(fieldWidth uint, schema AnySchema[F]) []error
	// AllowPadding determines the amount of initial padding a module expects.
	AllowPadding() bool
	// Identifies the length multiplier for this module.  For every trace, the
	// height of the corresponding module must be a multiple of this.  This is
	// used specifically to support interleaving constraints.
	LengthMultiplier() uint
	// Substitute any matchined labelled constants within this module
	Substitute(map[string]F)
}

// FieldAgnosticModule captures the notion of a module which is agnostic to the
// underlying field being used.
type FieldAgnosticModule[F any, M Module[F]] interface {
	Module[F]
	// Subdivide for a given bandwidth and maximum register width. This will
	// split all registers wider than the maximum permitted width into two or
	// more "limbs" (i.e. subregisters which do not exceeded the permitted
	// width).  For example, consider a register "r" of width u32. Subdividing
	// this register into registers of at most 8bits will result in four limbs:
	// r'0, r'1, r'2 and r'3 where (by convention) r'0 is the least significant.
	//
	// As part of the subdivision process, constraints may also need to be
	// divided when they exceed the maximum permitted bandwidth.  For example,
	// consider a simple constraint such as "x = y + 1" using 16bit registers
	// x,y.  Subdividing for a bandwidth of 10bits and a maximum register width
	// of 8bits means splitting each register into two limbs, and transforming
	// our constraint into:
	//
	// 256*x'1 + x'0 = 256*y'1 + y'0 + 1
	//
	// However, as it stands, this constraint exceeds our bandwidth requirement
	// since it requires at least 17bits of information to safely evaluate each
	// side.  Thus, the constraint itself must be subdivided into two parts:
	//
	// 256*c + x'0 = y'0 + 1  // lower
	//
	//         x'1 = y'1 + c  // upper
	//
	// Here, c is a 1bit register introduced as part of the transformation to
	// act as a "carry" between the two constraints.
	Subdivide(RegisterLimbsMap) M
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
type Table[F field.Element[F], C Constraint[F]] struct {
	name        string
	multiplier  uint
	padding     bool
	public      bool
	synthetic   bool
	registers   []register.Register
	constraints []C
	assignments []Assignment[F]
}

// NewTable constructs a table module with the given registers and constraints.
func NewTable[F field.Element[F], C Constraint[F]](name string, multiplier uint,
	padding, public, synthetic bool) *Table[F, C] {
	//
	return &Table[F, C]{name, multiplier, padding, public, synthetic, nil, nil, nil}
}

// Init implementation for ir.InitModule interface.
func (p *Table[F, C]) Init(name string, multiplier uint, padding, public, synthetic bool) *Table[F, C] {
	return &Table[F, C]{name, multiplier, padding, public, synthetic, nil, nil, nil}
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
func (p *Table[F, C]) Consistent(fieldWidth uint, schema AnySchema[F]) []error {
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
func (p *Table[F, C]) HasRegister(name string) (register.Id, bool) {
	for i := range p.Width() {
		if p.registers[i].Name == name {
			return register.NewId(i), true
		}
	}
	// Fail
	return register.UnusedId(), false
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

// IsPublic identifies whether or not this module is externally visible.
func (p *Table[F, C]) IsPublic() bool {
	return p.public
}

// IsSynthetic modules are generated during compilation, rather than being
// provided by the user.
func (p *Table[F, C]) IsSynthetic() bool {
	return p.synthetic
}

// RawAssignments provides raw access to those assignments defined as part of this
// table.
func (p *Table[F, C]) RawAssignments() []Assignment[F] {
	return p.assignments
}

// RawConstraints provides raw access to those constraints associated with this
// module.
func (p *Table[F, C]) RawConstraints() []C {
	return p.constraints
}

// Register returns the given register in this table.
func (p *Table[F, C]) Register(id register.Id) register.Register {
	return p.registers[id.Unwrap()]
}

// Registers returns an iterator over the underlying registers of this schema.
// Specifically, the index of a register in this array is its register index.
func (p *Table[F, C]) Registers() []register.Register {
	return p.registers
}

// Substitute any matchined labelled constants within this module
func (p *Table[F, C]) Substitute(mapping map[string]F) {
	for _, c := range p.assignments {
		c.Substitute(mapping)
	}
	//
	for _, c := range p.constraints {
		c.Substitute(mapping)
	}
}

// Width returns the number of registers in this Table.
func (p *Table[F, C]) Width() uint {
	return uint(len(p.registers))
}

func (p *Table[F, C]) String() string {
	return register.MapToString(p)
}

// Subdivide implementation for the FieldAgnosticModule interface.
func (p *Table[F, C]) Subdivide(mid ModuleId, mapping LimbsMap,
	assigner func(register.CarryAssignment) Assignment[F]) *Table[F, C] {
	//
	var (
		constraints []C
		assignments []Assignment[F]
		env         = register.NewAllocator(mapping.Module(mid).LimbsMap())
	)
	// Subdivide assignments
	for _, c := range p.assignments {
		var a any = c
		//nolint
		if fc, ok := a.(FieldAgnostic[Assignment[F]]); ok {
			assignments = append(assignments, fc.Subdivide(env, mapping))
		} else {
			panic(fmt.Sprintf("non-field agnostic assignment (%s)", reflect.TypeOf(a).String()))
		}
	}
	// Subdivide constraints
	for _, c := range p.constraints {
		var a any = c
		//nolint
		if fc, ok := a.(FieldAgnostic[C]); ok {
			constraints = append(constraints, fc.Subdivide(env, mapping))
		} else {
			panic(fmt.Sprintf("non-field agnostic constraint (%s)", reflect.TypeOf(a).String()))
		}
	}
	// Include any additional assignments required for carry lines
	for _, a := range env.Assignments() {
		assignments = append(assignments, assigner(a))
	}
	//
	return &Table[F, C]{p.name, p.multiplier, p.padding, p.public, p.synthetic, env.Registers(), constraints, assignments}
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
func (p *Table[F, C]) AddRegisters(registers ...register.Register) {
	// Add registers
	p.registers = append(p.registers, registers...)
}

// ============================================================================
// Encoding / Decoding
// ============================================================================

// GobEncode an option.  This allows it to be marshalled into a binary form.
func (p *Table[F, M]) GobEncode() (data []byte, err error) {
	var buffer bytes.Buffer
	//
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
