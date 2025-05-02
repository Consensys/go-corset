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
package insn

import (
	"math"
	"math/big"
)

// FALL_THRU is used to signal that a given instruction will fall through to the next
const FALL_THRU = math.MaxUint - 1

// RETURN is used to signal that a given instruction returns from the enclosing
// function.
const RETURN = math.MaxUint

// UNUSED_REGISTER is used to signal that a given register operand is unused.
const UNUSED_REGISTER = math.MaxUint

const (
	// INPUT_REGISTER signals a register used for holding the input values of a
	// function.
	INPUT_REGISTER = uint8(0)
	// OUTPUT_REGISTER signals a register used for holding the output values of
	// a function.
	OUTPUT_REGISTER = uint8(1)
	// TEMP_REGISTER signals a register used for holding temporary values during
	// computation.
	TEMP_REGISTER = uint8(2)
)

// Instruction provides an abstract notion of an executable "machine instruction".
type Instruction interface {
	// Execute a given instruction at a given program counter position, using a
	// given set of register values.  This may update the register values, and
	// returns the next program counter position.  If the program counter is
	// math.MaxUint then a return is signaled.
	//
	// Executing an instruction produces an updated program counter position, or
	// signals a fall through to the next instruction (MaxUint-1) or signals a
	// return from the enclosing function.
	Execute(state []big.Int, regs []Register) uint
	// Registers returns the set of registers read/written by this micro instruction.
	Registers() []uint
	// Registers returns the set of registers read this micro instruction.
	RegistersRead() []uint
	// Registers returns the set of registers written by this micro instruction.
	RegistersWritten() []uint
	// Validate that this instruction is well-formed.  For example, that it is
	// balanced, that there are no conflicting writes, that all temporaries have
	// been allocated, etc.
	Validate(regs []Register) error
	// Produce a suitable string representation of this instruction.  This is
	// primarily used for debugging.
	String(regs []Register) string
}

// Register describes a single register within a function.
type Register struct {
	// Kind of register (input / output)
	Kind uint8
	// Given name of this register.
	Name string
	// Width (in bits) of this register
	Width uint
}

// NewRegister creates a new register of a given kind with a given width.
func NewRegister(kind uint8, name string, width uint) Register {
	return Register{kind, name, width}
}

// IsInput determines whether or not this is an input register
func (p *Register) IsInput() bool {
	return p.Kind == INPUT_REGISTER
}

// IsOutput determines whether or not this is an output register
func (p *Register) IsOutput() bool {
	return p.Kind == OUTPUT_REGISTER
}

// Bound returns the first value which cannot be represented by the given
// bitwidth.  For example, the bound of an 8bit register is 256.
func (p *Register) Bound() *big.Int {
	var (
		bound = big.NewInt(2)
		width = big.NewInt(int64(p.Width))
	)
	// Compute 2^n
	return bound.Exp(bound, width, nil)
}

// MaxValue returns the largest value expressible in this register (i.e. Bound() -
// 1).  For example, the max value of an 8bit register is 255.
func (p *Register) MaxValue() *big.Int {
	max := p.Bound()
	max.Sub(max, &one)
	//
	return max
}

var one = *big.NewInt(1)
