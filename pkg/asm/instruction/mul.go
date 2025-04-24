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
	"math/big"
)

// Mul represents a generic operation of the following form:
//
// tn, .., t0 := s0 * ... * sm * c
//
// Here, t0 .. tn are the *target registers*, of which tn is the *most
// significant*.  These must be disjoint as we cannot assign simultaneously to
// the same register.  Likewise, s0 ... sm are the source registers, and c is a
// given (non-negative) constant. Observe the n == m is not required, meaning
// one can assign multiple registers.  For example, consider this case:
//
// c, r0 := r1 * 2
//
// Suppose that r0 and r1 are 16bit registers, whilst c is a 1bit register. The
// result of r1 * 2 occupies 17bits, of which the first 16 are written to r0
// with the most significant (i.e. 16th) bit written to c.  Thus, in this
// particular example, c represents a carry flag.
type Mul struct {
	// Target registers for addition
	Targets []uint
	// Source register for addition
	Sources []uint
	// Constant value (if applicable)
	Constant big.Int
}

// Bind any labels contained within this instruction using the given label map.
func (p *Mul) Bind(labels []uint) {
	// no-op
}

// Execute a given instruction at a given program counter position, using a
// given set of register values.  This may update the register values, and
// returns the next program counter position.  If the program counter is
// math.MaxUint then a return is signaled.
func (p *Mul) Execute(pc uint, regs []big.Int, widths []uint) uint {
	var value big.Int = one
	// Multiply register values
	for _, src := range p.Sources {
		value.Mul(&value, &regs[src])
	}
	// Multiply constant
	value.Mul(&value, &p.Constant)
	// Write value
	writeTargetRegisters(p.Targets, regs, widths, value)
	//
	return pc + 1
}

// Registers returns the set of registers read/written by this instruction.
func (p *Mul) Registers() []uint {
	return append(p.Targets, p.Sources...)
}
