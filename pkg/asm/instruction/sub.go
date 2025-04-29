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

// Sub represents a generic operation of the following form:
//
// tn, .., t0 := s0 - ... - sm - c
//
// Here, t0 .. tn are the *target registers*, of which tn is the *most
// significant*.  These must be disjoint as we cannot assign simultaneously to
// the same register.  Likewise, s0 ... sm are the source registers, and c is a
// given (non-negative) constant. Observe the n == m is not required, meaning
// one can assign multiple registers.  For example, consider this case:
//
// b, r0 := r1 - 1
//
// Suppose that r0 and r1 are 16bit registers, whilst c is a 1bit register. The
// result of r1 - 2 occupies 17bits, of which the first 16 are written to r0
// with the most significant (i.e. 16th) bit written to b.  Thus, in this
// particular example, b represents a borrow flag.
type Sub struct {
	// Target registers for addition
	Targets []uint
	// Source register for addition
	Sources []uint
	// Constant value (if applicable)
	Constant big.Int
}

// Bind any labels contained within this instruction using the given label map.
func (p *Sub) Bind(labels []uint) {
	// no-op
}

// Execute a given instruction at a given program counter position, using a
// given set of register values.  This may update the register values, and
// returns the next program counter position.  If the program counter is
// math.MaxUint then a return is signaled.
func (p *Sub) Execute(pc uint, state []big.Int, regs []Register) uint {
	var value big.Int
	// Clone initial value
	value.Set(&state[p.Sources[0]])
	// Subtract register values
	for _, src := range p.Sources[1:] {
		value.Sub(&value, &state[src])
	}
	// Subtract constant
	value.Sub(&value, &p.Constant)
	// Write value
	writeTargetRegisters(p.Targets, state, regs, value)
	//
	return pc + 1
}

// IsBalanced checks whether or not this instruction is correctly balanced.  The
// algorithm here may seem a little odd at first.  It counts the number of
// *unique values* required to hold both the positive and negative components of
// the right-hand side.  This gives the minimum bitwidth required.
func (p *Sub) IsBalanced(regs []Register) error {
	var (
		lhs_bits = sum_bits(p.Targets, regs)
		// Initially, include positive component of rhs.
		rhs = *regs[p.Sources[0]].MaxValue()
	)
	// Now, add negative components
	for _, target := range p.Sources[1:] {
		rhs.Add(&rhs, regs[target].MaxValue())
	}
	// Include constant (if relevant)
	rhs.Add(&rhs, &p.Constant)
	// lhs must be able to hold both.
	rhs_bits := uint(rhs.BitLen())
	// check
	if lhs_bits < rhs_bits {
		return fmt.Errorf("bit overflow (%d bits into %d bits)", rhs_bits, lhs_bits)
	}
	// Run the pivot check
	if err := checkPivot(p.Sources[0], p.Targets, regs); err != nil {
		return err
	}
	// Finally, ensure unique targets
	return checkUniqueTargets(p.Targets, regs)
}

// Registers returns the set of registers read/written by this instruction.
func (p *Sub) Registers() []uint {
	return append(p.Targets, p.Sources...)
}

// RegistersRead returns the set of registers read by this instruction.
func (p *Sub) RegistersRead() []uint {
	return p.Sources
}

// RegistersWritten returns the set of registers written by this instruction.
func (p *Sub) RegistersWritten() []uint {
	return p.Targets
}

// the pivot check is necessary to ensure we can properly rebalance a
// subtraction.  Consider "c,x = y-z" which is rebalanced to "x+z = y+256*c".
// The issue is that, for example, x cannot be split across both sides.  Thus,
// we need x to align with y.  For a case like "c,y,x = a-b" then we need either
// x to align with a, or y,x to align with a.
func checkPivot(source uint, targets []uint, regs []Register) error {
	var (
		rhs_width = regs[source].Width
		lhs_width = uint(0)
		pivot     = 0
	)
	// Consume source bits
	for lhs_width < rhs_width {
		lhs_width += regs[targets[pivot]].Width
		pivot = pivot + 1
	}
	// Check for alignment
	if lhs_width == rhs_width {
		// Yes, aligned.
		return nil
	}
	// Problem, no alignment.
	return fmt.Errorf("incorrect alignment (%d bits versus %d bits)", lhs_width, rhs_width)
}
