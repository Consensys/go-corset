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
	"strings"

	"github.com/consensys/go-corset/pkg/asm/io"
	"github.com/consensys/go-corset/pkg/schema/agnostic"
	"github.com/consensys/go-corset/pkg/schema/register"
)

var biZERO *big.Int = big.NewInt(0)

// Division operator divides either a register (or constant) by another register
// (or constant) producing a quotient and a remainder.
type Division struct {
	// Target registers
	Quotient, Remainder io.RegisterId
	// Dividend and right comparisons
	Dividend, Divisor Expr
}

// Clone this micro code.
func (p *Division) Clone() Code {
	return &Division{
		Quotient:  p.Quotient,
		Remainder: p.Remainder,
		Dividend:  p.Dividend.Clone(),
		Divisor:   p.Divisor.Clone(),
	}
}

// MicroExecute implementation for Code interface.
func (p *Division) MicroExecute(state io.State) (uint, uint) {
	var (
		lhs  = p.Dividend.Eval(state)
		rhs  = p.Divisor.Eval(state)
		quot big.Int
		rem  big.Int
	)
	// Check for division by zero
	if rhs.Cmp(biZERO) == 0 {
		return 0, io.FAIL
	}
	// Compute quotient / remainder
	quot.Div(lhs, rhs)
	rem.Mod(lhs, rhs)
	// Write target registers
	state.Store(p.Quotient, quot)
	state.Store(p.Remainder, rem)
	// Continue to next instruction
	return 1, 0
}

// RegistersRead implementation for Code interface.
func (p *Division) RegistersRead() []io.RegisterId {
	var regs []io.RegisterId
	//
	if p.Dividend.HasFirst() {
		regs = append(regs, p.Dividend.First())
	}
	//
	if p.Divisor.HasFirst() {
		regs = append(regs, p.Divisor.First())
	}
	//
	return regs
}

// RegistersWritten implementation for Code interface.
func (p *Division) RegistersWritten() []io.RegisterId {
	return []io.RegisterId{p.Quotient, p.Remainder}
}

// Split implementation for Code interface.
func (p *Division) Split(mapping register.LimbsMap, _ agnostic.RegisterAllocator) []Code {
	panic("todo")
}

func (p *Division) String(fn register.Map) string {
	var builder strings.Builder
	//
	builder.WriteString(fn.Register(p.Quotient).Name)
	builder.WriteString(", ")
	builder.WriteString(fn.Register(p.Remainder).Name)
	builder.WriteString(" = ")
	builder.WriteString(p.Dividend.String(fn))
	builder.WriteString(" / ")
	builder.WriteString(p.Divisor.String(fn))
	//
	return builder.String()
}

// Validate implementation for Code interface.
func (p *Division) Validate(fieldWidth uint, fn register.Map) error {
	var (
		qBits = fn.Register(p.Quotient).Width
		rBits = fn.Register(p.Remainder).Width
		dBits = p.Dividend.Bitwidth(fn)
		vBits = p.Divisor.Bitwidth(fn)
	)
	// check
	if qBits < dBits {
		return fmt.Errorf("quotient bit overflow (u%d into u%d)", dBits, qBits)
	} else if rBits < vBits {
		return fmt.Errorf("remainder bit overflow (u%d into u%d)", vBits, rBits)
	}
	//
	return nil
}
