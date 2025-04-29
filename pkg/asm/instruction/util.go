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
package instruction

import (
	"fmt"
	"math/big"
)

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

var zero = *big.NewInt(0)
var one = *big.NewInt(1)

// Write the value to a given set of target registers, splitting its bits as
// necessary.  The target registers are given with the least significant first.
func writeTargetRegisters(targets []uint, state []big.Int, regs []Register, value big.Int) {
	var (
		offset uint = 0
	)
	//
	for _, reg := range targets {
		width := regs[reg].Width
		state[reg] = readBitSlice(offset, width, value)
		offset += width
	}
}

func readBitSlice(offset uint, width uint, value big.Int) big.Int {
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

// Ensure a given
func checkUniqueTargets(targets []uint, regs []Register) error {
	for i := range targets {
		for j := i + 1; j < len(targets); j++ {
			if targets[i] == targets[j] {
				return fmt.Errorf("conflicting write to %s", regs[targets[i]].Name)
			}
		}
	}
	//
	return nil
}
