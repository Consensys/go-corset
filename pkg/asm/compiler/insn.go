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
	"github.com/consensys/go-corset/pkg/asm/io/micro"
	"github.com/consensys/go-corset/pkg/schema/agnostic"
)

func (p *StateTranslator[T, E, M]) translateCode(cc uint, codes []micro.Code) E {
	switch codes[cc].(type) {
	case *micro.Assign:
		return p.translateAssign(cc, codes)
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
func (p *StateTranslator[T, E, M]) translateAssign(cc uint, codes []micro.Code) E {
	var (
		code = codes[cc].(*micro.Assign)
		// build rhs
		rhs = p.translatePolynomial(code.Source)
		// build lhs (must be after rhs)
		lhs = p.WriteAndShiftRegisters(code.Targets)
	)
	// Construct equation
	eqn := Sum(lhs).Equals(rhs)
	// Continue
	return eqn.And(p.translateCode(cc+1, codes))
}

func (p *StateTranslator[T, E, M]) translateInOut(cc uint, codes []micro.Code) E {
	var code = codes[cc].(*micro.InOut)
	// In/Out codes are really nops from the perspective of compilation.  Their
	// primary purposes is to assist trace expansion.
	//
	// NOTE: we have to pretend that we've written registers here, otherwise
	// forwarding will not be enabled.
	p.WriteRegisters(code.RegistersWritten())
	//
	return p.translateCode(cc+1, codes)
}

func (p *StateTranslator[T, E, M]) translateJmp(cc uint, codes []micro.Code) E {
	var (
		code   = codes[cc].(*micro.Jmp)
		pc_ip1 = p.WritePc()
		dst    = Number[T, E](code.Target)
	)
	// PC[i+1] = target
	eqn := pc_ip1.Equals(dst)
	//
	return p.WithLocalConstancies(eqn)
}

func (p *StateTranslator[T, E, M]) translateRet() E {
	var (
		stamp_i   = p.Stamp(false)
		stamp_ip1 = p.Stamp(true)
		one       = Number[T, E](1)
	)
	// STAMP[i]+1 == STAMP[i+1]
	eqn := one.Add(stamp_i).Equals(stamp_ip1)
	// force stamp increment
	return p.WithLocalConstancies(eqn)
}

func (p *StateTranslator[T, E, M]) translateSkip(cc uint, codes []micro.Code) E {
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

// // Consider an assignment b, X := Y - 1.  This should be translated into the
// // constraint: X + 1 == Y - 256.b (assuming b is u1, and X/Y are u8).
// func (p *StateTranslator[T, E, M]) rebalanceSub(lhs []E, rhs []E, regs []io.Register, code *micro.Sub) ([]E, []E) {
// 	//
// 	pivot := 0
// 	width := int(regs[code.Sources[0].Unwrap()].Width)
// 	//
// 	for width > 0 {
// 		reg := regs[code.Targets[pivot].Unwrap()]
// 		//
// 		pivot++
// 		width -= int(reg.Width)
// 	}
// 	// Sanity check
// 	if width < 0 {
// 		// Should be caught earlier, hence unreachable.
// 		panic("failed rebalancing subtraction")
// 	}
// 	//
// 	nlhs := slices.Clone(lhs[:pivot])
// 	nrhs := []E{rhs[0]}
// 	// rebalance
// 	nlhs = append(nlhs, rhs[1:]...)
// 	nrhs = append(nrhs, lhs[pivot:]...)
// 	// done
// 	return nlhs, nrhs
// }

// Translate polynomial (c0*x0$0*...*xn$0) + ... + (cm*x0$m*...*xn$m) where cX
// are constant coefficients.
func (p *StateTranslator[T, E, M]) translatePolynomial(poly agnostic.Polynomial) E {
	var (
		terms []E = make([]E, poly.Len())
	)
	//
	for i := range poly.Len() {
		terms[i] = p.translateMonomial(poly.Term(i))
	}
	// Optimisation
	if len(terms) == 1 {
		return terms[0]
	}
	// Normal case
	return Sum(terms)
}

// Translate a monomial of the form c*x0*...*xn where c is a constant coefficient.
func (p *StateTranslator[T, E, M]) translateMonomial(mono agnostic.Monomial) E {
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
