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
package micro

import (
	"fmt"
	"math/big"

	"github.com/consensys/go-corset/pkg/asm/insn"
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

// Sequential indicates whether or not this microinstruction can execute
// sequentially onto the ne
func (p *Mul) Sequential() bool {
	return true
}

// Terminal indicates whether or not this microinstruction terminates the
// enclosing function.
func (p *Mul) Terminal() bool {
	return false
}

// Execute a given instruction at a given program counter position, using a
// given set of register values.  This may update the register values, and
// returns the next program counter position.  If the program counter is
// math.MaxUint then a return is signaled.
func (p *Mul) Execute(state []big.Int, regs []Register) uint {
	var value big.Int
	// Assign first value
	value.Set(&state[p.Sources[0]])
	// Multiply register values
	for _, src := range p.Sources[1:] {
		value.Mul(&value, &state[src])
	}
	// Multiply constant
	value.Mul(&value, &p.Constant)
	// Write value
	insn.WriteTargetRegisters(p.Targets, state, regs, value)
	//
	return insn.FALL_THRU
}

// Lower this instruction into a exactly one more micro instruction.
func (p *Mul) Lower() Instruction {
	// Lowering here produces an instruction containing a single microcode.
	return Instruction{[]Code{p}}
}

// Registers returns the set of registers read/written by this instruction.
func (p *Mul) Registers() []uint {
	return append(p.Targets, p.Sources...)
}

// RegistersRead returns the set of registers read by this instruction.
func (p *Mul) RegistersRead() []uint {
	return p.Sources
}

// RegistersWritten returns the set of registers written by this instruction.
func (p *Mul) RegistersWritten() []uint {
	return p.Targets
}

func (p *Mul) String(regs []Register) string {
	return assignmentToString(p.Targets, p.Sources, p.Constant, regs, one, " * ")
}

// Validate checks whether or not this instruction is correctly balanced.
func (p *Mul) Validate(regs []Register) error {
	var (
		lhs_bits = sum_bits(p.Targets, regs)
		rhs      big.Int
	)
	//
	rhs.Set(&one)
	//
	for _, target := range p.Sources {
		rhs.Mul(&rhs, regs[target].MaxValue())
	}
	// Include constant (if relevant)
	rhs.Mul(&rhs, &p.Constant)
	//
	rhs_bits := uint(rhs.BitLen())
	// check
	if lhs_bits < rhs_bits {
		return fmt.Errorf("bit overflow (%d bits into %d bits)", rhs_bits, lhs_bits)
	}
	//
	return insn.CheckTargetRegisters(p.Targets, regs)
}

/*
// Translate this instruction into low-level constraints.
func (p *Mul) Translate(st *StateTranslator) {
	// build rhs
	rhs := st.ReadRegisters(p.Sources)
	// build lhs (must be after rhs)
	lhs := st.WriteRegisters(p.Targets)
	// include constant if this makes sense
	if p.Constant.Cmp(&one) != 0 {
		var elem fr.Element
		//
		elem.SetBigInt(&p.Constant)
		rhs = append(rhs, hir.NewConst(elem))
	}
	// construct equation
	eqn := hir.Equals(hir.Sum(lhs...), hir.Product(rhs...))
	// construct constraint
	st.Constrain("mul", eqn)
}
*/
