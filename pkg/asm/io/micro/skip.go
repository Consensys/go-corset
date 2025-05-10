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

	"github.com/consensys/go-corset/pkg/asm/io"
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

// MicroExecute a given micro-code, using a given local state.  This may update
// the register values, and returns either the number of micro-codes to "skip
// over" when executing the enclosing instruction or, if skip==0, a destination
// program counter (which can signal return of enclosing function).
func (p *Skip) MicroExecute(state io.State, iomap io.Map) (uint, uint) {
	var (
		lhs = state.Read(p.Left)
		rhs *big.Int
	)
	//
	if p.Right != io.UNUSED_REGISTER {
		rhs = state.Read(p.Right)
	} else {
		rhs = &p.Constant
	}
	//
	if lhs.Cmp(rhs) != 0 {
		return 1 + p.Skip, 0
	} else {
		return 1, 0
	}
}

// RegistersRead returns the set of registers read by this instruction.
func (p *Skip) RegistersRead() []uint {
	if p.Right != io.UNUSED_REGISTER {
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
	if p.Right != io.UNUSED_REGISTER {
		rhsLimbs := env.SplitTargetRegisters(p.Right)
		for i := uint(0); i < n; i++ {
			ncode := &Skip{lhsLimbs[i], rhsLimbs[i], p.Constant, skip - i}
			ncodes = append(ncodes, ncode)
		}
	} else {
		constantLimbs := env.SplitConstant(p.Constant, n)
		for i := uint(0); i < n; i++ {
			ncode := &Skip{lhsLimbs[i], io.UNUSED_REGISTER, constantLimbs[i], skip - i}
			ncodes = append(ncodes, ncode)
		}
	}
	//
	return ncodes
}

func (p *Skip) String(env io.Environment[Instruction]) string {
	var (
		regs = env.Enclosing().Registers
		l    = regs[p.Left].Name
	)
	//
	if p.Right != io.UNUSED_REGISTER {
		return fmt.Sprintf("skip %s!=%s %d", l, regs[p.Right].Name, p.Skip)
	}
	//
	return fmt.Sprintf("skip %s!=%s %d", l, p.Constant.String(), p.Skip)
}

// Validate checks whether or not this instruction is correctly balanced.
func (p *Skip) Validate(env io.Environment[Instruction]) error {
	var (
		regs = env.Enclosing().Registers
		lw   = regs[p.Left].Width
	)
	//
	if p.Right != io.UNUSED_REGISTER {
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
