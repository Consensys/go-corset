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
	"math"
	"math/big"
)

// RETURN is used to signal that a given instruction returns from the enclosing
// function.
const RETURN = math.MaxUint

// State collects together local state necessary for executing a given
// instruction.  This contrasts with an I/O Map which represents the non-local
// state.
type State struct {
	// Program Counter position.
	Pc uint
	// Values for each register in this state.
	State []big.Int
	// Registers referenced in this state.  This is necessary to determine
	// appropriate bitwidths for copying data, and also for zadebugging.
	Registers []Register
}

// Read value of a given register from this state.
func (p *State) Read(reg uint) *big.Int {
	return &p.State[reg]
}

// ReadN reads the values of zero or more registers from this state.
func (p *State) ReadN(registers []uint) []big.Int {
	values := make([]big.Int, len(registers))
	//
	for i, src := range registers {
		values[i] = p.State[src]
	}
	//
	return values
}

// Next returns the program counter for the following instruction.
func (p *State) Next() uint {
	return p.Pc + 1
}

// Write a given value across a set of registers, splitting its bits as
// necessary.  The target registers are given with the least significant first.
// For example, consider writing 01100010 to registers [R1, R2] of type u4.
// Then, after the write, we have R1=0010 and R2=0110.
func (p *State) Write(value big.Int, registers ...uint) {
	var offset uint = 0
	//
	for _, reg := range registers {
		width := p.Registers[reg].Width
		p.State[reg] = ReadBitSlice(offset, width, value)
		offset += width
	}
}

// WriteN writes a set of zero or more values to a corresponding set of
// registers in this state.
func (p *State) WriteN(registers []uint, values []big.Int) {
	for i, dst := range registers {
		p.State[dst] = values[i]
	}
}

// ReadBitSlice reads a slice of bits starting at a given offset in a give
// value.  For example, consider the value is 10111000 and we have offset=1 and
// width=4, then the result is 1100.
func ReadBitSlice(offset uint, width uint, value big.Int) big.Int {
	var slice big.Int
	//
	for i := 0; uint(i) < width; i++ {
		// Read appropriate bit
		bit := value.Bit(i + int(offset))
		// set appropriate bit
		slice.SetBit(&slice, i, bit)
	}
	//
	return slice
}
