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
	"errors"
	"fmt"
	"math/big"

	"github.com/consensys/go-corset/pkg/asm/io"
	"github.com/consensys/go-corset/pkg/asm/io/macro/expr"
	"github.com/consensys/go-corset/pkg/asm/io/micro"
	"github.com/consensys/go-corset/pkg/schema/register"
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
	// Constant label
	Label string
	// Then/Else branches
	Then, Else Expr
}

// Execute implementation for Instruction interface.
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
		value = p.Then.Eval(state.Internal())
	} else {
		value = p.Else.Eval(state.Internal())
	}
	// Write value
	state.StoreAcross(value, p.Targets...)
	//
	return state.Pc() + 1
}

// Lower implementation for Instruction interface.
func (p *IfThenElse) Lower(pc uint) micro.Instruction {
	var (
		codes      []micro.Code
		rhs        = register.UnusedId()
		thenBranch = p.Then.Polynomial()
		elseBranch = p.Else.Polynomial()
	)
	//
	switch p.Cond {
	case EQ:
		codes = []micro.Code{
			&micro.Skip{Left: p.Left, Right: rhs, Constant: p.Right, Skip: 2},
			// Then branch
			&micro.Assign{Targets: p.Targets, Source: thenBranch},
			&micro.Jmp{Target: pc + 1},
			// Else branch
			&micro.Assign{Targets: p.Targets, Source: elseBranch},
			&micro.Jmp{Target: pc + 1},
		}
	case NEQ:
		codes = []micro.Code{
			&micro.Skip{Left: p.Left, Right: rhs, Constant: p.Right, Skip: 2},
			// Then branch
			&micro.Assign{Targets: p.Targets, Source: elseBranch},
			&micro.Jmp{Target: pc + 1},
			// Else branch
			&micro.Assign{Targets: p.Targets, Source: thenBranch},
			&micro.Jmp{Target: pc + 1},
		}
	default:
		panic("unreachable")
	}
	//
	return micro.Instruction{Codes: codes}
}

// RegistersRead implementation for Instruction interface.
func (p *IfThenElse) RegistersRead() []io.RegisterId {
	var regs = []io.RegisterId{p.Left}
	//
	regs = append(regs, expr.RegistersRead(p.Then)...)
	regs = append(regs, expr.RegistersRead(p.Else)...)
	//
	return regs
}

// RegistersWritten implementation for Instruction interface.
func (p *IfThenElse) RegistersWritten() []io.RegisterId {
	return p.Targets
}

func (p *IfThenElse) String(fn register.Map) string {
	var (
		regs    = fn.Registers()
		targets = io.RegistersReversedToString(p.Targets, regs)
		left    = regs[p.Left.Unwrap()].Name
		right   = p.Right.String()
		tb      = p.Then.String(fn)
		fb      = p.Else.String(fn)
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

// Validate implementation for Instruction interface.
func (p *IfThenElse) Validate(fieldWidth uint, fn register.Map) error {
	var (
		regs                  = fn.Registers()
		lhs_bits              = sumTargetBits(p.Targets, regs)
		then_bits, thenSigned = expr.BitWidth(p.Then, fn)
		else_bits, elseSigned = expr.BitWidth(p.Else, fn)
		rhs_bits              = max(then_bits, else_bits)
	)
	// check
	if lhs_bits < rhs_bits {
		return fmt.Errorf("bit overflow (u%d into u%d)", rhs_bits, lhs_bits)
	} else if rhs_bits > fieldWidth {
		return fmt.Errorf("field overflow (u%d into u%d field)", rhs_bits, fieldWidth)
	} else if thenSigned || elseSigned {
		return errors.New("signed exprtession not supported here")
	}
	//
	return io.CheckTargetRegisters(p.Targets, regs)
}
