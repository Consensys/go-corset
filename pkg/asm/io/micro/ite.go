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
	"github.com/consensys/go-corset/pkg/schema"
	"github.com/consensys/go-corset/pkg/schema/agnostic"
	"github.com/consensys/go-corset/pkg/util/math"
)

const (
	// EQ indicates an equality condition
	EQ uint8 = 0
	// NEQ indicates a non-equality condition
	NEQ uint8 = 1
)

// Ite represents a ternary operation.
type Ite struct {
	Targets []io.RegisterId
	// Cond indicates the condition
	Cond uint8
	// Left-hand side
	Left io.RegisterId
	// Left-hand size
	Right big.Int
	// Then/Else branches
	Then, Else big.Int
}

// Clone this micro code.
func (p *Ite) Clone() Code {
	return p
}

// MicroExecute implementation for the micro.Code interface.
func (p *Ite) MicroExecute(state io.State) (uint, uint) {
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
	return 1, 0
}

// RegistersRead returns the set of registers read by this instruction.
func (p *Ite) RegistersRead() []io.RegisterId {
	return []io.RegisterId{p.Left}
}

// RegistersWritten returns the set of registers written by this instruction.
func (p *Ite) RegistersWritten() []io.RegisterId {
	return p.Targets
}

// Split this micro code using registers of arbirary width into one or more
// micro codes using registers of a fixed maximum width.
func (p *Ite) Split(env schema.RegisterAllocator) []Code {
	// Split targets
	targets := agnostic.ApplyMapping(env, p.Targets...)
	// Split left-hand register
	left := agnostic.ApplyMapping(env, p.Left)
	// Sanity check for nwo
	if len(left) != 1 {
		panic(fmt.Sprintf("if-then-else cannot split register \"%s\"", env.Register(p.Left).Name))
	}
	// Construct split instruction
	code := &Ite{targets, p.Cond, left[0], p.Right, p.Then, p.Else}
	// FIXME: sanity check arguments
	return []Code{code}
}

func (p *Ite) String(fn schema.RegisterMap) string {
	var (
		regs    = fn.Registers()
		targets = io.RegistersToString(p.Targets, regs)
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
func (p *Ite) Validate(fieldWidth uint, fn schema.RegisterMap) error {
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
