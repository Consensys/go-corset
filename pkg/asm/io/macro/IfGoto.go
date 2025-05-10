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
)

const (
	// EQ indicates an equality condition
	EQ uint8 = 0
	// NEQ indicates a non-equality condition
	NEQ uint8 = 1
	// LT indicates a less-than condition
	LT uint8 = 2
	// GT indicates a greater-than condition
	GT uint8 = 3
	// LTEQ indicates a less-than-or-equals condition
	LTEQ uint8 = 4
	// GTEQ indicates a greater-than-or-equals condition
	GTEQ uint8 = 5
)

// IfGoto describes a conditional branch, which is either jeq ("Jump if equal") or
// jne ("Jump if not equal").  This has two variants: register-register; and,
// register-constant.  The latter is indiciated when the right register is
// marked as UNUSED.
type IfGoto struct {
	// Cond indicates the condition
	Cond uint8
	// Left and right comparisons
	Left, Right uint
	//
	Constant big.Int
	// Target identifies target PC
	Target uint
}

// Bind any labels contained within this instruction using the given label map.
func (p *IfGoto) Bind(labels []uint) {
	p.Target = labels[p.Target]
}

// Link any buses used within this instruction using the given bus map.
func (p *IfGoto) Link(buses []uint) {
	// nothing to link
}

// Execute this instruction with the given local and global state.  The next
// program counter position is returned, or io.RETURN if the enclosing
// function has terminated (i.e. because a return instruction was
// encountered).
func (p *IfGoto) Execute(state io.State, iomap io.Map) uint {
	var (
		lhs   *big.Int = state.Read(p.Left)
		rhs   *big.Int
		taken bool
	)
	//
	if p.Right != io.UNUSED_REGISTER {
		rhs = state.Read(p.Right)
	} else {
		rhs = &p.Constant
	}
	//
	switch p.Cond {
	case EQ:
		taken = lhs.Cmp(rhs) == 0
	case NEQ:
		taken = lhs.Cmp(rhs) != 0
	case LT:
		taken = lhs.Cmp(rhs) < 0
	case LTEQ:
		taken = lhs.Cmp(rhs) <= 0
	case GT:
		taken = lhs.Cmp(rhs) > 0
	case GTEQ:
		taken = lhs.Cmp(rhs) >= 0
	default:
		panic("unreachable")
	}
	// Check if taken or not taken
	if taken {
		return p.Target
	}
	//
	return state.Next()
}

// Lower this (macro) instruction into a sequence of one or more micro
// instructions.
func (p *IfGoto) Lower(pc uint) micro.Instruction {
	var codes []micro.Code
	//
	switch p.Cond {
	case EQ:
		codes = []micro.Code{
			&micro.Skip{Left: p.Left, Right: p.Right, Constant: p.Constant, Skip: 1},
			&micro.Jmp{Target: p.Target},
			&micro.Jmp{Target: pc + 1},
		}
	case NEQ:
		codes = []micro.Code{
			&micro.Skip{Left: p.Left, Right: p.Right, Constant: p.Constant, Skip: 1},
			&micro.Jmp{Target: pc + 1},
			&micro.Jmp{Target: p.Target},
		}
	default:
		panic("unreachable")
	}
	//
	return micro.Instruction{Codes: codes}
}

// RegistersRead returns the set of registers read by this instruction.
func (p *IfGoto) RegistersRead() []uint {
	if p.Right != io.UNUSED_REGISTER {
		return []uint{p.Left}
	}
	//
	return []uint{p.Left, p.Right}
}

// RegistersWritten returns the set of registers written by this instruction.
func (p *IfGoto) RegistersWritten() []uint {
	return nil
}

func (p *IfGoto) String(env io.Environment[Instruction]) string {
	var (
		regs = env.Enclosing().Registers
		l    = regs[p.Left].Name
		r    string
		op   string
	)
	//
	switch p.Cond {
	case EQ:
		op = "=="
	case NEQ:
		op = "!="
	case LT:
		op = "<"
	case LTEQ:
		op = "<="
	case GT:
		op = ">"
	case GTEQ:
		op = ">="
	default:
		panic("unreachable")
	}
	//
	if p.Right != io.UNUSED_REGISTER {
		r = regs[p.Right].Name
	} else {
		r = p.Constant.String()
	}
	//
	return fmt.Sprintf("if %s%s%s goto %d", l, op, r, p.Target)
}

// Validate checks whether or not this instruction is correctly balanced.
func (p *IfGoto) Validate(env io.Environment[Instruction]) error {
	if p.Left == p.Right {
		switch p.Cond {
		case EQ, LTEQ, GTEQ:
			return fmt.Errorf("always taken")
		default:
			return fmt.Errorf("never taken")
		}
	}
	//
	return nil
}
