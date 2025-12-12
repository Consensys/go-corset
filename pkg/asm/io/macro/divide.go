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
	"strings"

	"github.com/consensys/go-corset/pkg/asm/io"
	"github.com/consensys/go-corset/pkg/asm/io/macro/expr"
	"github.com/consensys/go-corset/pkg/asm/io/micro"
	"github.com/consensys/go-corset/pkg/schema"
)

var biZERO *big.Int = big.NewInt(0)
var biONE *big.Int = big.NewInt(1)

// Division represents an Divisionment of the form:
//
// q,r,w = e1 / e1
//
// where e1/e2 is either a variable or constant, and the witness w is a proof
// that r < e1.
type Division struct {
	// Target registers
	Quotient  expr.RegAccess
	Remainder expr.RegAccess
	Witness   expr.RegAccess
	// Dividend expression
	Dividend expr.AtomicExpr
	// Divisor expression
	Divisor expr.AtomicExpr
}

// Execute implementation for Instruction interface.
func (p *Division) Execute(state io.State) uint {
	var (
		lhs  = p.Dividend.Eval(state.Internal())
		rhs  = p.Divisor.Eval(state.Internal())
		quot big.Int
		rem  big.Int
		wit  big.Int
	)
	// Check for division by zero
	if rhs.Cmp(biZERO) == 0 {
		return io.FAIL
	}
	// Compute quotient / remainder / witntess
	quot.Div(&lhs, &rhs)
	rem.Mod(&lhs, &rhs)
	wit.Sub(&rhs, &rem)
	wit.Sub(&wit, biONE)
	// Write target registers
	state.Store(p.Quotient.Register, quot)
	state.Store(p.Remainder.Register, rem)
	state.Store(p.Witness.Register, wit)
	// Continue to next instruction
	return state.Pc() + 1
}

// Lower implementation for Instruction interface.
func (p *Division) Lower(pc uint) micro.Instruction {
	codes := []micro.Code{
		&micro.Division{
			Quotient:  p.Quotient.Register,
			Remainder: p.Remainder.Register,
			Witness:   p.Witness.Register,
			Dividend:  p.Dividend.ToMicroExpr(),
			Divisor:   p.Divisor.ToMicroExpr(),
		},
		&micro.Jmp{Target: pc + 1},
	}
	//
	return micro.Instruction{Codes: codes}
}

// RegistersRead implementation for Instruction interface.
func (p *Division) RegistersRead() []io.RegisterId {
	return expr.RegistersRead(p.Dividend, p.Divisor)
}

// RegistersWritten implementation for Instruction interface.
func (p *Division) RegistersWritten() []io.RegisterId {
	return []io.RegisterId{p.Quotient.Register, p.Remainder.Register, p.Witness.Register}
}

// String implementation for Instruction interface.
func (p *Division) String(fn schema.RegisterMap) string {
	var builder strings.Builder
	//
	builder.WriteString(p.Quotient.String(fn))
	builder.WriteString(", ")
	builder.WriteString(p.Remainder.String(fn))
	builder.WriteString(", ")
	builder.WriteString(p.Witness.String(fn))
	builder.WriteString(" = ")
	builder.WriteString(p.Dividend.String(fn))
	builder.WriteString(" / ")
	builder.WriteString(p.Divisor.String(fn))
	//
	return builder.String()
}

// Validate implementation for Instruction interface.
func (p *Division) Validate(fieldWidth uint, fn schema.RegisterMap) error {
	var (
		qBits, _ = expr.BitWidth(&p.Quotient, fn)
		rBits, _ = expr.BitWidth(&p.Remainder, fn)
		wBits, _ = expr.BitWidth(&p.Witness, fn)
		dBits, _ = expr.BitWidth(p.Dividend, fn)
		vBits, _ = expr.BitWidth(p.Divisor, fn)
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
