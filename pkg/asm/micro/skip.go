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

// Skip microcode performs a conditional skip over a given number of codes. The
// condition is either that two registers are equal, or that they are not equal.
// This has two variants: register-register; and, register-constant.  The latter
// is indiciated when the right register is marked as UNUSED.
type Skip struct {
	// Left and right comparisons
	Left, Right uint
	//
	Constant big.Int
	// Skip
	Skip uint
}

// Clone this micro code.
func (p *Skip) Clone() Code {
	var constant big.Int
	//
	constant.Set(&p.Constant)
	//
	return &Skip{
		Left:     p.Left,
		Right:    p.Right,
		Constant: constant,
		Skip:     p.Skip,
	}
}

// Sequential indicates whether or not this microinstruction can execute
// sequentially onto the next.
func (p *Skip) Sequential() bool {
	return false
}

// Terminal indicates whether or not this microinstruction terminates the
// enclosing function.
func (p *Skip) Terminal() bool {
	return false
}

// Execute an unconditional branch instruction by returning the destination
// program counter.
func (p *Skip) Execute(state []big.Int, regs []Register) uint {
	panic("goto")
}

// Registers returns the set of registers read/written by this instruction.
func (p *Skip) Registers() []uint {
	return p.RegistersRead()
}

// RegistersRead returns the set of registers read by this instruction.
func (p *Skip) RegistersRead() []uint {
	if p.Right != insn.UNUSED_REGISTER {
		return []uint{p.Left}
	}
	//
	return []uint{p.Left, p.Right}
}

// RegistersWritten returns the set of registers written by this instruction.
func (p *Skip) RegistersWritten() []uint {
	return nil
}

// Split this micro code using registers of arbirary width into one or more
// micro codes using registers of a fixed maximum width.
func (p *Skip) Split(env *RegisterSplittingEnvironment) []Code {
	// NOTE: we can assume left and right have matching bitwidths
	var (
		lhsLimbs = env.SplitTargetRegisters(p.Left)
		ncodes   []Code
		n        = uint(len(lhsLimbs))
		skip     = p.Skip + n - 1
	)
	//
	if p.Right != insn.UNUSED_REGISTER {
		rhsLimbs := env.SplitTargetRegisters(p.Right)
		for i := uint(0); i < n; i++ {
			ncode := &Skip{lhsLimbs[i], rhsLimbs[i], p.Constant, skip - i}
			ncodes = append(ncodes, ncode)
		}
	} else {
		constantLimbs := env.SplitConstant(p.Constant, n)
		for i := uint(0); i < n; i++ {
			ncode := &Skip{lhsLimbs[i], insn.UNUSED_REGISTER, constantLimbs[i], skip - i}
			ncodes = append(ncodes, ncode)
		}
	}
	//
	return ncodes
}

func (p *Skip) String(regs []Register) string {
	var l = regs[p.Left].Name
	//
	if p.Right != insn.UNUSED_REGISTER {
		return fmt.Sprintf("skip %s!=%s %d", l, regs[p.Right].Name, p.Skip)
	}
	//
	return fmt.Sprintf("skip %s!=%s %d", l, p.Constant.String(), p.Skip)
}

// Validate checks whether or not this instruction is correctly balanced.
func (p *Skip) Validate(fieldWidth uint, regs []Register) error {
	lw := regs[p.Left].Width
	//
	if p.Right != insn.UNUSED_REGISTER {
		rw := regs[p.Right].Width
		//
		if lw != rw {
			return fmt.Errorf("bit mismatch (%dbits vs %dbits)", lw, rw)
		}
	} else {
		cw := uint(p.Constant.BitLen())
		//
		if lw < cw {
			return fmt.Errorf("bit overflow (%dbits vs %dbits)", lw, cw)
		}
	}
	//
	return nil
}

/*
// Translate this instruction into low-level constraints.
func (p *Skip) Translate(st *StateTranslator) {
	st.Jump(p.Target)
}
*/
