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
	"fmt"
	"math/big"
)

// WriteTargetRegisters writes a given value to a given set of registers,
// splitting its bits as necessary.  The target registers are given with the
// least significant first.  For example, consider writing 01100010 to registers
// [R1, R2] of type u4.  Then, after the write, we have R1=0010 and R2=0110.
func WriteTargetRegisters(targets []uint, state []big.Int, regs []Register, value big.Int) {
	var offset uint = 0
	//
	for _, reg := range targets {
		width := regs[reg].Width
		state[reg] = ReadBitSlice(offset, width, value)
		offset += width
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

// CheckTargetRegisters performs some simple checks on a set of target registers
// being written.  Firstly, they cannot be input registers (as this are always
// constant).  Secondly, we cannot write to the same register more than once
// (i.e. a conflicting write).
func CheckTargetRegisters(targets []uint, regs []Register) error {
	for i := range targets {
		//
		if regs[targets[i]].IsInput() {
			return fmt.Errorf("cannot write input %s", regs[targets[i]].Name)
		}
		//
		for j := i + 1; j < len(targets); j++ {
			if targets[i] == targets[j] {
				return fmt.Errorf("conflicting write to %s", regs[targets[i]].Name)
			}
		}
	}
	//
	return nil
}
