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
	"github.com/consensys/go-corset/pkg/schema"
	"github.com/consensys/go-corset/pkg/schema/agnostic"
)

var biZERO *big.Int = big.NewInt(0)
var biONE *big.Int = big.NewInt(1)

// Division operator divides either a register (or constant) by another register
// (or constant) producing a quotient and a remainder.
type Division struct {
	// Target registers
	Quotient, Remainder, Witness io.RegisterId
	// Dividend and right comparisons
	Dividend, Divisor Expr
}

// Clone this micro code.
func (p *Division) Clone() Code {
	return &Division{
		Quotient:  p.Quotient,
		Remainder: p.Remainder,
		Witness:   p.Witness,
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
		wit  big.Int
	)
	// Check for division by zero
	if rhs.Cmp(biZERO) == 0 {
		return 0, io.FAIL
	}
	// Compute quotient / remainder / witness
	quot.Div(lhs, rhs)
	rem.Mod(lhs, rhs)
	wit.Sub(rhs, &rem)
	wit.Sub(&wit, biONE)
	// Write target registers
	state.Store(p.Quotient, quot)
	state.Store(p.Remainder, rem)
	state.Store(p.Witness, wit)
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
	return []io.RegisterId{p.Quotient, p.Remainder, p.Witness}
}

// Split implementation for Code interface.
func (p *Division) Split(mapping schema.RegisterAllocator) []Code {
	var (
		qLimbs      = mapping.LimbIds(p.Quotient)
		rLimbs      = mapping.LimbIds(p.Remainder)
		wLimbs      = mapping.LimbIds(p.Witness)
		qLimbWidths = agnostic.WidthsOfLimbs(mapping, qLimbs)
		rLimbWidths = agnostic.WidthsOfLimbs(mapping, rLimbs)
	)
	// FIXME: implement this (somehow)
	if len(qLimbs) != 1 || len(rLimbs) != 1 || len(wLimbs) != 1 {
		panic("splitting for division unsupported")
	}
	//
	return []Code{&Division{
		Quotient:  qLimbs[0],
		Remainder: rLimbs[0],
		Witness:   wLimbs[0],
		Dividend:  splitExpression(qLimbWidths, p.Dividend, mapping),
		Divisor:   splitExpression(rLimbWidths, p.Divisor, mapping),
	}}
}

func splitExpression(widths []uint, expr Expr, mapping schema.RegisterLimbsMap) Expr {
	//
	if expr.HasSecond() {
		return NewConstant(expr.Second())
	}
	//
	limbs := mapping.LimbIds(expr.First())
	//
	if len(widths) != 1 {
		panic("splitting for division unsupported")
	}

	return NewRegister(limbs[0])
}

func (p *Division) String(fn schema.RegisterMap) string {
	var builder strings.Builder
	//
	builder.WriteString(fn.Register(p.Quotient).Name)
	builder.WriteString(", ")
	builder.WriteString(fn.Register(p.Remainder).Name)
	builder.WriteString(", ")
	builder.WriteString(fn.Register(p.Witness).Name)
	builder.WriteString(" = ")
	builder.WriteString(p.Dividend.String(fn))
	builder.WriteString(" / ")
	builder.WriteString(p.Divisor.String(fn))
	//
	return builder.String()
}

// Validate implementation for Code interface.
func (p *Division) Validate(fieldWidth uint, fn schema.RegisterMap) error {
	var (
		qBits = fn.Register(p.Quotient).Width
		rBits = fn.Register(p.Remainder).Width
		wBits = fn.Register(p.Witness).Width
		dBits = p.Dividend.Bitwidth(fn)
		vBits = p.Divisor.Bitwidth(fn)
	)
	// check
	if qBits < dBits {
		return fmt.Errorf("quotient bit overflow (u%d into u%d)", dBits, qBits)
	} else if rBits < vBits {
		return fmt.Errorf("remainder bit overflow (u%d into u%d)", vBits, rBits)
	} else if wBits < vBits {
		return fmt.Errorf("witness bit overflow (u%d into u%d)", vBits, wBits)
	}
	//
	return nil
}
