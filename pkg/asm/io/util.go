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
	"fmt"
	"math/big"
	"strings"

	"github.com/consensys/go-corset/pkg/util"
)

// Alias for big integer representation of 0.
var zero big.Int = *big.NewInt(0)

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

// NumberOfLimbs determines the number of register limbs required for a given
// bitwidth. For example, a 64bit register splits into two limbs for a maximum
// register width of 32bits. Observe that an e.g. 60bit register also splits
// into two limbs here as well, where the most significant limb is 28bits wide
// and the least significant is 32bits width.
func NumberOfLimbs(maxWidth uint, regWidth uint) uint {
	n := regWidth / maxWidth
	m := regWidth % maxWidth
	// Check for uneven split
	if m != 0 {
		return n + 1
	}
	//
	return n
}

// MaxNumberOfLimbs returns the maximum number of limbs required for any
// register in the given target registers.  For example, given registers r0 and
// r1 of bitwidths 16bits and 8bits (respectively), then 2 is maximum number of
// limbs for an 8bit maximum register width.
func MaxNumberOfLimbs(maxWidth uint, regs []Register, targets []uint) uint {
	var n = uint(0)
	//
	for _, target := range targets {
		regWidth := regs[target].Width
		n = max(n, NumberOfLimbs(maxWidth, regWidth))
	}
	//
	return n
}

// SplitRegister splits a register into a number of limbs with the given maximum
// bitwidth.  For the resulting array, the least significant register is first.
// Since registers are always split to the maximum width as much as possible, it
// is only the most significant register which may (in some cases) have fewer
// bits than the maximum allowed.
func SplitRegister(maxWidth uint, r Register) []Register {
	var (
		nlimbs = NumberOfLimbs(maxWidth, r.Width)
		limbs  = make([]Register, nlimbs)
		width  = r.Width
	)
	//
	for i := uint(0); i < nlimbs; i++ {
		ith_name := fmt.Sprintf("%s'%d", r.Name, i)
		ith_width := min(maxWidth, width)
		limbs[i] = Register{Name: ith_name, Kind: r.Kind, Width: ith_width}
		width -= maxWidth
	}
	//
	return limbs
}

// SplitValueAcrossRegisters splits a given value across a set of registers,
// where the least significant register comes first.  This will panic if the
// value does not fit into the given registers.
func SplitValueAcrossRegisters(constant *big.Int, registers ...Register) []big.Int {
	var (
		limb  big.Int
		limbs []big.Int = make([]big.Int, len(registers))
		acc   big.Int
	)
	//
	acc.Set(constant)
	//
	for i := 0; acc.Cmp(&zero) != 0; i++ {
		var (
			bound     = big.NewInt(2)
			limbWidth = registers[i].Width
		)
		// Determine upper bound
		bound.Exp(bound, big.NewInt(int64(limbWidth)), nil)
		//
		limb.Mod(&acc, bound)
		limbs[i] = limb

		acc.Rsh(&acc, limbWidth)
	}
	// Sanity check nothing is left.
	if acc.Cmp(&zero) != 0 {
		panic("value overflows given registers")
	}
	//
	return limbs
}

// RegistersToString returns a string representation for zero or more registers
// separated by a comma.
func RegistersToString(rids []uint, regs []Register) string {
	var builder strings.Builder
	//
	for i := 0; i < len(rids); i++ {
		var rid = rids[i]
		//
		if i != 0 {
			builder.WriteString(", ")
		}
		//
		if i < len(regs) {
			builder.WriteString(regs[rid].Name)
		} else {
			builder.WriteString(fmt.Sprintf("?%d", rid))
		}
	}
	//
	return builder.String()
}

// RegistersReversedToString returns a string representation for zero or more
// registers in reverse order, separated by a comma.  This is useful, for
// example, when printing the left-hand side of an assignment.
func RegistersReversedToString(rids []uint, regs []Register) string {
	return RegistersToString(util.Reverse(rids), regs)
}
