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
package io

import (
	"bytes"
	"encoding/gob"
	"math"
	"math/big"

	"github.com/consensys/go-corset/pkg/schema"
	"github.com/consensys/go-corset/pkg/util/collection/iter"
)

const (
	// PC_NAME gives the name used for the program counter in traces.
	PC_NAME = "$pc"
	// PC_INDEX gives the register index used for the program counter (which is
	// currently always be 0).
	PC_INDEX = uint(0)
)

// Register defines the notion of a register within a function.
type Register = schema.Register

// RegisterId abstracts the notion of a register id.
type RegisterId = schema.RegisterId

const (
	// UNUSED_REGISTER provides a simple way to distinguish registers and
	// constants in certain instructions.
	UNUSED_REGISTER = math.MaxUint
)

// FunctionInstance represents a specific instance of a function.  That is, a
// mapping from input values to expected output values.
type FunctionInstance struct {
	// Identifies corresponding function.
	Function uint
	// Inputs identifies the input arguments
	Inputs map[string]big.Int
	// Outputs identifies the outputs
	Outputs map[string]big.Int
}

// Function defines a distinct functional entity within the system.  Functions
// accepts zero or more inputs and produce zero or more outputs.  Functions
// declare zero or more internal registers for use, and their interpretation is
// given by a sequence of zero or more instructions.
type Function[T Instruction[T]] struct {
	// unique module identifier
	id schema.ModuleId
	// Unique name of this function.
	name string
	// Registers describes zero or more registers of a given width.  Each
	// register can be designated as an input / output or temporary.
	registers []Register
	// Code defines the body of this function.
	code []T
}

// NewFunction constructs a new function with the given components.
func NewFunction[T Instruction[T]](id schema.ModuleId, name string, registers []Register, code []T) Function[T] {
	return Function[T]{id, name, registers, code}
}

// Assignments returns an iterator over the assignments of this schema.
// These are the computations used to assign values to all computed columns
// in this module.
func (p *Function[T]) Assignments() iter.Iterator[schema.Assignment] {
	var assignment Assignment[T] = Assignment[T]{p.id, p.name, p.registers, p.code}
	//
	return iter.NewUnitIterator[schema.Assignment](assignment)
}

// CodeAt returns the ith instruction making up the body of this function.
func (p *Function[T]) CodeAt(i uint) T {
	return p.code[i]
}

// Code returns the instructions making up the body of this function.
func (p *Function[T]) Code() []T {
	return p.code
}

// Constraints provides access to those constraints associated with this
// function.
func (p *Function[T]) Constraints() iter.Iterator[schema.Constraint] {
	var constraint Constraint[T] = Constraint[T]{p.id, p.name, p.registers, p.code}
	//
	return iter.NewUnitIterator[schema.Constraint](constraint)
}

// Consistent applies a number of internal consistency checks.  Whilst not
// strictly necessary, these can highlight otherwise hidden problems as an aid
// to debugging.
func (p *Function[T]) Consistent(schema.Schema[schema.Constraint]) []error {
	// TODO: add checks?
	return nil
}

// Id returns the unique module identifier for this function.
func (p *Function[T]) Id() schema.ModuleId {
	return p.id
}

// HasRegister checks whether a register with the given name exists and, if
// so, returns its register identifier.  Otherwise, it returns false.
func (p *Function[T]) HasRegister(name string) (RegisterId, bool) {
	for i, r := range p.registers {
		if r.Name == name {
			return schema.NewRegisterId(uint(i)), true
		}
	}
	// Failed
	return schema.NewUnusedRegisterId(), false
}

// Inputs returns the set of input registers for this function.
func (p *Function[T]) Inputs() []Register {
	var inputs []Register
	//
	for _, r := range p.registers {
		if r.IsInput() {
			inputs = append(inputs, r)
		}
	}
	//
	return inputs
}

// LengthMultiplier identifies the length multiplier for this module.  For every
// trace, the height of the corresponding module must be a multiple of this.
// This is used specifically to support interleaving constraints.
func (p *Function[T]) LengthMultiplier() uint {
	return 1
}

// Name returns the name of this function.
func (p *Function[T]) Name() string {
	return p.name
}

// Outputs returns the set of output registers for this function.
func (p *Function[T]) Outputs() []Register {
	var outputs []Register
	//
	for _, r := range p.registers {
		if r.IsOutput() {
			outputs = append(outputs, r)
		}
	}
	//
	return outputs
}

// Register returns the ith register used in this function.
func (p *Function[T]) Register(id schema.RegisterId) Register {
	return p.registers[id.Unwrap()]
}

// Registers returns the set of all registers used during execution of this
// function.
func (p *Function[T]) Registers() []Register {
	return p.registers
}

// Subdivide implementation for the FieldAgnosticModule interface.
func (p *Function[T]) Subdivide(mapping schema.RegisterMappings) *Function[T] {
	var (
		registers []schema.Register
		// Determine mapping information for this function.
		modmap = NewSplittingEnvironment(mapping.ModuleOf(p.name))
		// Updated instruction sequence
		ninsns []T
	)
	// Append mapping registers
	for i := range p.registers {
		rid := schema.NewRegisterId(uint(i))
		for _, limb := range modmap.LimbIds(rid) {
			registers = append(registers, modmap.Limb(limb))
		}
	}
	// Split instructions
	for _, insn := range p.Code() {
		var ith Instruction[T] = insn
		//nolint
		if i, ok := ith.(SplittableInstruction[T]); ok {
			ninsns = append(ninsns, i.SplitRegisters(modmap))
		} else {
			panic("non-field agnostic instruction encountered")
		}
	}
	// Done
	nf := NewFunction(p.Id(), p.Name(), registers, ninsns)
	//
	return &nf
}

// Width identifiers the number of registers in this function.
func (p *Function[T]) Width() uint {
	return uint(len(p.registers))
}

// AllocateRegister allocates a new register of the given kind, name and width
// into this function.
func (p *Function[T]) AllocateRegister(kind schema.RegisterType, name string, width uint) RegisterId {
	index := uint(len(p.registers))
	p.registers = append(p.registers, schema.NewRegister(kind, name, width))
	// Done
	return schema.NewRegisterId(index)
}

// ============================================================================
// Encoding / Decoding
// ============================================================================

// nolint
func (p *Function[T]) GobEncode() ([]byte, error) {
	var buffer bytes.Buffer
	gobEncoder := gob.NewEncoder(&buffer)
	//
	if err := gobEncoder.Encode(p.name); err != nil {
		return nil, err
	}
	//
	if err := gobEncoder.Encode(p.registers); err != nil {
		return nil, err
	}
	//
	if err := gobEncoder.Encode(p.code); err != nil {
		return nil, err
	}
	//
	return buffer.Bytes(), nil
}

// nolint
func (p *Function[T]) GobDecode(data []byte) error {
	var (
		buffer     = bytes.NewBuffer(data)
		gobDecoder = gob.NewDecoder(buffer)
	)
	//
	if err := gobDecoder.Decode(&p.name); err != nil {
		return err
	}
	//
	if err := gobDecoder.Decode(&p.registers); err != nil {
		return err
	}
	//
	if err := gobDecoder.Decode(&p.code); err != nil {
		return err
	}
	//
	return nil
}
