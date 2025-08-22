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

	sc "github.com/consensys/go-corset/pkg/schema"
	"github.com/consensys/go-corset/pkg/util/collection/iter"
	"github.com/consensys/go-corset/pkg/util/field"
)

const (
	// PC_NAME gives the name used for the program counter in traces.
	PC_NAME = "$pc"
	// RET_NAME gives the name used for the return line in traces.
	RET_NAME = "$ret"
)

// Register defines the notion of a register within a function.
type Register = sc.Register

// RegisterId abstracts the notion of a register id.
type RegisterId = sc.RegisterId

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
type Function[F field.Element[F], T Instruction[T]] struct {
	// unique module identifier
	id sc.ModuleId
	// Unique name of this function.
	name string
	// Registers describes zero or more registers of a given width.  Each
	// register can be designated as an input / output or temporary.
	registers []Register
	// Code defines the body of this function.
	code []T
}

// NewFunction constructs a new function with the given components.
func NewFunction[F field.Element[F], T Instruction[T]](id sc.ModuleId, name string, registers []Register, code []T,
) Function[F, T] {
	return Function[F, T]{id, name, registers, code}
}

// Assignments returns an iterator over the assignments of this schema.
// These are the computations used to assign values to all computed columns
// in this module.
func (p *Function[F, T]) Assignments() iter.Iterator[sc.Assignment[F]] {
	var assignment = Assignment[F, T]{p.id, p.name, p.registers, p.code}
	//
	return iter.NewUnitIterator[sc.Assignment[F]](assignment)
}

// CodeAt returns the ith instruction making up the body of this function.
func (p *Function[F, T]) CodeAt(i uint) T {
	return p.code[i]
}

// Code returns the instructions making up the body of this function.
func (p *Function[F, T]) Code() []T {
	return p.code
}

// Constraints provides access to those constraints associated with this
// function.
func (p *Function[F, T]) Constraints() iter.Iterator[sc.Constraint[F]] {
	var constraint Constraint[F, T] = Constraint[F, T]{p.id, p.name, p.registers, p.code}
	//
	return iter.NewUnitIterator[sc.Constraint[F]](constraint)
}

// Consistent applies a number of internal consistency checks.  Whilst not
// strictly necessary, these can highlight otherwise hidden problems as an aid
// to debugging.
func (p *Function[F, T]) Consistent(sc.AnySchema[F]) []error {
	// TODO: add checks?
	return nil
}

// Id returns the unique module identifier for this function.
func (p *Function[F, T]) Id() sc.ModuleId {
	return p.id
}

// IsAtomic determines whether or not this is a "one line function".  That is,
// where every instance of this function occupies exactly one line in the
// corresponding trace.  This is useful to know, as certain optimisations can be
// applied for one line functions (e.g. no PC register is required).
func (p *Function[F, T]) IsAtomic() bool {
	return len(p.code) == 1
}

// HasRegister checks whether a register with the given name exists and, if
// so, returns its register identifier.  Otherwise, it returns false.
func (p *Function[F, T]) HasRegister(name string) (RegisterId, bool) {
	for i, r := range p.registers {
		if r.Name == name {
			return sc.NewRegisterId(uint(i)), true
		}
	}
	// Failed
	return sc.NewUnusedRegisterId(), false
}

// Inputs returns the set of input registers for this function.
func (p *Function[F, T]) Inputs() []Register {
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
func (p *Function[F, T]) LengthMultiplier() uint {
	return 1
}

// AllowPadding determines whether the given module supports padding at the
// beginning of the module.  Assembly modules do not support padding, as this
// causes various problems of its own.
func (p *Function[F, T]) AllowPadding() bool {
	return false
}

// Name returns the name of this function.
func (p *Function[F, T]) Name() string {
	return p.name
}

// Outputs returns the set of output registers for this function.
func (p *Function[F, T]) Outputs() []Register {
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
func (p *Function[F, T]) Register(id sc.RegisterId) Register {
	return p.registers[id.Unwrap()]
}

// Registers returns the set of all registers used during execution of this
// function.
func (p *Function[F, T]) Registers() []Register {
	return p.registers
}

// Width identifiers the number of registers in this function.
func (p *Function[F, T]) Width() uint {
	return uint(len(p.registers))
}

// AllocateRegister allocates a new register of the given kind, name and width
// into this function.
func (p *Function[F, T]) AllocateRegister(kind sc.RegisterType, name string, width uint) RegisterId {
	var (
		index = uint(len(p.registers))
		// Default padding (for now)
		padding big.Int
	)

	p.registers = append(p.registers, sc.NewRegister(kind, name, width, padding))
	// Done
	return sc.NewRegisterId(index)
}

// ============================================================================
// Encoding / Decoding
// ============================================================================

// nolint
func (p *Function[F, T]) GobEncode() ([]byte, error) {
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
func (p *Function[F, T]) GobDecode(data []byte) error {
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
