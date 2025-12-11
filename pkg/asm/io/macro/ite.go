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
package macro

import (
	"fmt"
	"math/big"

	"github.com/consensys/go-corset/pkg/asm/io"
	"github.com/consensys/go-corset/pkg/asm/io/micro"
	"github.com/consensys/go-corset/pkg/schema"
	"github.com/consensys/go-corset/pkg/util/math"
)

// IfThenElse represents a ternary operation.
type IfThenElse struct {
	Targets []io.RegisterId
	// Cond indicates the condition
	Cond uint8
	// Left-hand side
	Left io.RegisterId
	// Right-hand side
	Right big.Int
	// Then/Else branches
	Then, Else big.Int
}

// Execute this instruction with the given local and global state.  The next
// program counter position is returned, or io.RETURN if the enclosing
// function has terminated (i.e. because a return instruction was
// encountered).
func (p *IfThenElse) Execute(state io.State) uint {
	var (
		lhs   *big.Int = state.Load(p.Left)
		rhs   *big.Int = &p.Right
		value big.Int
		taken bool
	)
	// Check whether taken or not.
	switch p.Cond {
	case EQ:
		taken = lhs.Cmp(rhs) == 0
	case NEQ:
		taken = lhs.Cmp(rhs) != 0
	default:
		panic("unreachable")
	}
	//
	if taken {
		value = p.Then
	} else {
		value = p.Else
	}
	// Write value
	state.StoreAcross(value, p.Targets...)
	//
	return state.Pc() + 1
}

// Lower this instruction into a exactly one more micro instruction.
func (p *IfThenElse) Lower(pc uint) micro.Instruction {
	code := &micro.Ite{
		Targets: p.Targets,
		Cond:    p.Cond,
		Left:    p.Left,
		Right:   p.Right,
		Then:    p.Then,
		Else:    p.Else,
	}
	//
	return micro.NewInstruction(code, &micro.Jmp{Target: pc + 1})
}

// RegistersRead returns the set of registers read by this instruction.
func (p *IfThenElse) RegistersRead() []io.RegisterId {
	return []io.RegisterId{p.Left}
}

// RegistersWritten returns the set of registers written by this instruction.
func (p *IfThenElse) RegistersWritten() []io.RegisterId {
	return p.Targets
}

func (p *IfThenElse) String(fn schema.RegisterMap) string {
	var (
		regs    = fn.Registers()
		targets = io.RegistersReversedToString(p.Targets, regs)
		left    = regs[p.Left.Unwrap()].Name
		right   = p.Right.String()
		tb      = p.Then.String()
		fb      = p.Else.String()
		op      string
	)
	//
	switch p.Cond {
	case EQ:
		op = "=="
	case NEQ:
		op = "!="
	default:
		panic("unreachable")
	}
	//
	return fmt.Sprintf("%s = %s%s%s ? %s : %s", targets, left, op, right, tb, fb)
}

// Validate checks whether or not this instruction is correctly balanced.
func (p *IfThenElse) Validate(fieldWidth uint, fn schema.RegisterMap) error {
	var (
		regs     = fn.Registers()
		lhs_bits = sumTargetBits(p.Targets, regs)
		rhs_bits = sumThenElseBits(p.Then, p.Else)
	)
	// check
	if lhs_bits < rhs_bits {
		return fmt.Errorf("bit overflow (u%d into u%d)", rhs_bits, lhs_bits)
	} else if rhs_bits > fieldWidth {
		return fmt.Errorf("field overflow (u%d into u%d field)", rhs_bits, fieldWidth)
	}
	//
	return io.CheckTargetRegisters(p.Targets, regs)
}

func sumThenElseBits(tb, fb big.Int) uint {
	var (
		tRange = math.NewInterval(tb, tb)
		fRange = math.NewInterval(fb, fb)
		values = tRange.Union(fRange)
	)
	// Observe, values cannot be negative by construction.
	bitwidth, _ := values.BitWidth()
	//
	return bitwidth
}
