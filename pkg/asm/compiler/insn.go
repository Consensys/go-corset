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
package compiler

import (
	"fmt"
	"strings"

	"github.com/consensys/go-corset/pkg/asm/io/micro"
	"github.com/consensys/go-corset/pkg/schema"
	sc "github.com/consensys/go-corset/pkg/schema"
	"github.com/consensys/go-corset/pkg/schema/agnostic"
	"github.com/consensys/go-corset/pkg/util/poly"
)

func (p *StateTranslator[F, T, E, M]) translateCode(cc uint, codes []micro.Code) E {
	switch codes[cc].(type) {
	case *micro.Assign:
		return p.translateAssign(cc, codes)
	case *micro.Fail:
		return False[T, E]()
	case *micro.InOut:
		return p.translateInOut(cc, codes)
	case *micro.Jmp:
		return p.translateJmp(cc, codes)
	case *micro.Ret:
		return p.translateRet()
	case *micro.Skip:
		return p.translateSkip(cc, codes)
	default:
		panic("unreachable")
	}
}

// Translate this instruction into low-level constraints.
func (p *StateTranslator[F, T, E, M]) translateAssign(cc uint, codes []micro.Code) E {
	var (
		code = codes[cc].(*micro.Assign)
		// build rhs
		rhs, signed = p.translatePolynomial(code.Source)
		// build lhs (must be after rhs)
		lhs = p.WriteAndShiftRegisters(code.Targets)
		// equation
		eqn E
	)
	// Construct equation
	if signed && !hasSignBit(code.Targets, p.mapping.Registers) {
		str := assignToString(p.mapping.Registers, code.Targets, code.Source)
		panic(fmt.Sprintf("assignment missing sign bit (%s)", str))
	} else if signed {
		// Signed case, so rebalance
		lhs, rhs = p.rebalanceAssign(lhs, rhs)
	}
	//
	eqn = Sum(lhs).Equals(Sum(rhs))
	// Continue
	return eqn.And(p.translateCode(cc+1, codes))
}

func (p *StateTranslator[F, T, E, M]) translateInOut(cc uint, codes []micro.Code) E {
	var code = codes[cc].(*micro.InOut)
	// In/Out codes are really nops from the perspective of compilation.  Their
	// primary purposes is to assist trace expansion.
	//
	// NOTE: we have to pretend that we've written registers here, otherwise
	// forwarding will not be enabled.
	p.WriteRegisters(code.RegistersWritten())
	p.WriteRegister(code.Bus().EnableLine)
	// Enable line must be set high
	enabled := p.ReadRegister(code.Bus().EnableLine).Equals(Number[T, E](1))
	//
	return enabled.And(p.translateCode(cc+1, codes))
}

func (p *StateTranslator[F, T, E, M]) translateJmp(cc uint, codes []micro.Code) E {
	var code = codes[cc].(*micro.Jmp)
	//
	return p.Goto(code.Target)
}

func (p *StateTranslator[F, T, E, M]) translateRet() E {
	return p.Terminate()
}

func (p *StateTranslator[F, T, E, M]) translateSkip(cc uint, codes []micro.Code) E {
	var (
		code  = codes[cc].(*micro.Skip)
		clone = p.Clone()
		lhs   = clone.translateCode(cc+1, codes)
		rhs   = p.translateCode(cc+1+code.Skip, codes)
		left  = p.ReadRegister(code.Left)
		right E
	)
	//
	if !code.Right.IsUsed() {
		right = BigNumber[T, E](&code.Constant)
	} else {
		right = p.ReadRegister(code.Right)
	}
	//
	return IfElse(left.Equals(right), lhs, rhs)
}

// Consider an assignment b, X := Y - 1.  This should be translated into the
// constraint: X + 1 == Y + 256.b (assuming b is u1, and X/Y are u8).
func (p *StateTranslator[F, T, E, M]) rebalanceAssign(lhs []E, rhs []E) ([]E, []E) {
	var (
		n = len(lhs) - 1
		// Extract sign bit
		sign = lhs[n]
	)
	// Remove sign bit
	lhs = lhs[:n]
	// Move sign bit onto rhs
	rhs = append(rhs, sign)
	// Done
	return lhs, rhs
}

// Translate polynomial (c0*x0$0*...*xn$0) + ... + (cm*x0$m*...*xn$m) where cX
// are constant coefficients.  This generates a given translation of terms,
// along with an indication as to whether this is signed or not.
func (p *StateTranslator[F, T, E, M]) translatePolynomial(poly agnostic.Polynomial) (pos []E, signed bool) {
	var (
		terms []E
	)
	//
	for i := range poly.Len() {
		ith := poly.Term(i)
		//
		signed = signed || ith.IsNegative()
		terms = append(terms, p.translateMonomial(ith))
	}
	// Done
	return terms, signed
}

// Translate a monomial of the form c*x0*...*xn where c is a constant coefficient.
func (p *StateTranslator[F, T, E, M]) translateMonomial(mono agnostic.Monomial) E {
	var (
		n         = mono.Len()
		coeff     = mono.Coefficient()
		terms []E = make([]E, n+1)
	)
	//
	for i := range mono.Len() {
		terms[i] = p.ReadRegister(mono.Nth(i))
	}
	// Optimise for case where coeff == 1?
	terms[n] = BigNumber[T, E](&coeff)
	//
	return Product(terms)
}

func hasSignBit(targets []schema.RegisterId, regs []schema.Register) bool {
	var (
		n = len(targets) - 1
	)
	//
	if n < 0 {
		// This should be unreachable in practice.
		return false
	}
	// Look for single sign bit
	return regs[targets[n].Unwrap()].Width == 1
}

// useful for debugging
//
// nolint
func assignToString(registers []schema.Register, lhs []schema.RegisterId, rhs agnostic.Polynomial) string {
	var builder strings.Builder
	//
	for i, ith := range lhs {
		if i != 0 {
			builder.WriteString(",")
		}
		builder.WriteString(registers[ith.Unwrap()+1].Name)
	}
	//
	builder.WriteString(" := ")
	//
	builder.WriteString(poly.String(rhs, func(id sc.RegisterId) string {
		return registers[id.Unwrap()+1].Name
	}))
	//
	return builder.String()
}
