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
	"slices"

	"github.com/consensys/go-corset/pkg/asm/insn"
	"github.com/consensys/go-corset/pkg/hir"
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

// Clone this micro code.
func (p *Sub) Clone() Code {
	var constant big.Int
	//
	constant.Set(&p.Constant)
	//
	return &Sub{
		slices.Clone(p.Targets),
		slices.Clone(p.Sources),
		constant,
	}
}

// Sequential indicates whether or not this microinstruction can execute
// sequentially onto the next.
func (p *Sub) Sequential() bool {
	return true
}

// Terminal indicates whether or not this microinstruction terminates the
// enclosing function.
func (p *Sub) Terminal() bool {
	return false
}

// Execute a given instruction at a given program counter position, using a
// given set of register values.  This may update the register values, and
// returns the next program counter position.  If the program counter is
// math.MaxUint then a return is signaled.
func (p *Sub) Execute(state []big.Int, regs []Register) uint {
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
	insn.WriteTargetRegisters(p.Targets, state, regs, value)
	//
	return insn.FALL_THRU
}

// Lower this instruction into a exactly one more micro instruction.
func (p *Sub) Lower(pc uint) Instruction {
	// Lowering here produces an instruction containing a single microcode.
	return Instruction{[]Code{p, &Jmp{Target: pc + 1}}}
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

func (p *Sub) String(regs []Register) string {
	return assignmentToString(p.Targets, p.Sources, p.Constant, regs, zero, " - ")
}

// Validate checks whether or not this instruction is correctly balanced.  The
// algorithm here may seem a little odd at first.  It counts the number of
// *unique values* required to hold both the positive and negative components of
// the right-hand side.  This gives the minimum bitwidth required.
func (p *Sub) Validate(regs []Register) error {
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
	return insn.CheckTargetRegisters(p.Targets, regs)
}

/*
// Translate this instruction into low-level constraints.
func (p *Sub) Translate(st *StateTranslator) {
	// build rhs
	rhs := st.ReadRegisters(p.Sources)
	// build lhs (must be after rhs)
	lhs := st.WriteRegisters(p.Targets)
	// include constant if this makes sense
	if p.Constant.Cmp(&zero) != 0 {
		var elem fr.Element
		//
		elem.SetBigInt(&p.Constant)
		rhs = append(rhs, hir.NewConst(elem))
	}
	// Rebalance the subtraction
	lhs, rhs = rebalanceSubtraction(lhs, rhs, st.mapping.Registers, p)
	// construct (balanced) equation
	eqn := hir.Equals(hir.Sum(lhs...), hir.Sum(rhs...))
	// construct constraint
	st.Constrain("sub", eqn)
}
*/
// Consider an assignment b, X := Y - 1.  This should be translated into the
// constraint: X + 1 == Y - 256.b (assuming b is u1, and X/Y are u8).
func rebalanceSubtraction(lhs []hir.Expr, rhs []hir.Expr, regs []Register, insn *Sub) ([]hir.Expr, []hir.Expr) {
	//
	pivot := 0
	width := int(regs[insn.Sources[0]].Width)
	//
	for width > 0 {
		reg := regs[insn.Targets[pivot]]
		//
		pivot++
		width -= int(reg.Width)
	}
	// Sanity check
	if width < 0 {
		// Should be caught earlier, hence unreachable.
		panic("failed rebalancing subtraction")
	}
	//
	nlhs := slices.Clone(lhs[:pivot])
	nrhs := []hir.Expr{rhs[0]}
	// rebalance
	nlhs = append(nlhs, rhs[1:]...)
	nrhs = append(nrhs, lhs[pivot:]...)
	// done
	return nlhs, nrhs
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
