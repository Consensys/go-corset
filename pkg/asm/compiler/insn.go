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
	"math/big"
	"slices"

	"github.com/consensys/gnark-crypto/ecc/bls12-377/fr"
	"github.com/consensys/go-corset/pkg/asm/insn"
	"github.com/consensys/go-corset/pkg/asm/micro"
	"github.com/consensys/go-corset/pkg/hir"
)

var zero = *big.NewInt(0)

var one = *big.NewInt(1)

func translate(cc uint, codes []micro.Code, st StateTranslator) hir.Expr {
	switch codes[cc].(type) {
	case *micro.Add:
		return translateAdd(cc, codes, st)
	case *micro.Jmp:
		return translateJmp(cc, codes, st)
	case *micro.Mul:
		return translateMul(cc, codes, st)
	case *micro.Ret:
		return translateRet(st)
	case *micro.Skip:
		return translateSkip(cc, codes, st)
	case *micro.Sub:
		return translateSub(cc, codes, st)
	default:
		panic("unreachable")
	}
}

// Translate this instruction into low-level constraints.
func translateAdd(cc uint, codes []micro.Code, st StateTranslator) hir.Expr {
	var (
		code = codes[cc].(*micro.Add)
		// build rhs
		rhs = st.ReadRegisters(code.Sources)
		// build lhs (must be after rhs)
		lhs = st.WriteRegisters(code.Targets)
	)
	// include constant if this makes sense
	if code.Constant.Cmp(&zero) != 0 {
		var elem fr.Element
		//
		elem.SetBigInt(&code.Constant)
		rhs = append(rhs, hir.NewConst(elem))
	}
	// Construct equation
	eqn := hir.Equals(hir.Sum(lhs...), hir.Sum(rhs...))
	// Continue
	return hir.Conjunction(eqn, translate(cc+1, codes, st))
}

func translateJmp(cc uint, codes []micro.Code, st StateTranslator) hir.Expr {
	var (
		code   = codes[cc].(*micro.Jmp)
		pc_ip1 = st.Pc(true)
		dst    = hir.NewConst64(uint64(code.Target))
	)
	// PC[i+1] = target
	eqn := hir.Equals(pc_ip1, dst)
	//
	return st.WithLocalConstancies(eqn)
}

func translateMul(cc uint, codes []micro.Code, st StateTranslator) hir.Expr {
	var (
		code = codes[cc].(*micro.Mul)
		// build rhs
		rhs = st.ReadRegisters(code.Sources)
		// build lhs (must be after rhs)
		lhs = st.WriteRegisters(code.Targets)
	)
	// include constant if this makes sense
	if code.Constant.Cmp(&one) != 0 {
		var elem fr.Element
		//
		elem.SetBigInt(&code.Constant)
		rhs = append(rhs, hir.NewConst(elem))
	}
	// Construct equation
	eqn := hir.Equals(hir.Sum(lhs...), hir.Product(rhs...))
	// Continue
	return hir.Conjunction(eqn, translate(cc+1, codes, st))
}

func translateRet(st StateTranslator) hir.Expr {
	var (
		stamp_i   = st.Stamp(false)
		stamp_ip1 = st.Stamp(true)
	)
	// STAMP[i]+1 == STAMP[i+1]
	eqn := hir.Equals(hir.Sum(hir.ONE, stamp_i), stamp_ip1)
	// force stamp increment
	return st.WithLocalConstancies(eqn)
}

func translateSkip(cc uint, codes []micro.Code, st StateTranslator) hir.Expr {
	var (
		code  = codes[cc].(*micro.Skip)
		lhs   = translate(cc+1, codes, st.Clone())
		rhs   = translate(cc+1+code.Skip, codes, st)
		left  = st.ReadRegister(code.Left)
		right hir.Expr
		elem  fr.Element
	)
	//
	if code.Right == insn.UNUSED_REGISTER {
		elem.SetBigInt(&code.Constant)
		right = hir.NewConst(elem)
	} else {
		right = st.ReadRegister(code.Right)
	}
	//
	return hir.IfElse(hir.Equals(left, right), lhs, rhs)
}

func translateSub(cc uint, codes []micro.Code, st StateTranslator) hir.Expr {
	var (
		code = codes[cc].(*micro.Sub)
		// build rhs
		rhs = st.ReadRegisters(code.Sources)
		// build lhs (must be after rhs)
		lhs = st.WriteRegisters(code.Targets)
	)
	// include constant if this makes sense
	if code.Constant.Cmp(&zero) != 0 {
		var elem fr.Element
		//
		elem.SetBigInt(&code.Constant)
		rhs = append(rhs, hir.NewConst(elem))
	}
	// Rebalance the subtraction
	lhs, rhs = rebalanceSub(lhs, rhs, st.mapping.Registers, code)
	// construct (balanced) equation
	eqn := hir.Equals(hir.Sum(lhs...), hir.Sum(rhs...))
	// continue
	return hir.Conjunction(eqn, translate(cc+1, codes, st))
}

// Consider an assignment b, X := Y - 1.  This should be translated into the
// constraint: X + 1 == Y - 256.b (assuming b is u1, and X/Y are u8).
func rebalanceSub(lhs []hir.Expr, rhs []hir.Expr, regs []insn.Register, code *micro.Sub) ([]hir.Expr, []hir.Expr) {
	//
	pivot := 0
	width := int(regs[code.Sources[0]].Width)
	//
	for width > 0 {
		reg := regs[code.Targets[pivot]]
		//
		pivot++
		width -= int(reg.Width)
	}
	// Sanity check
	if width < 0 {
		// Should be caught earlier, hence unreachable.
		panic("failed rebalancing subtraction")
	}
	//
	nlhs := slices.Clone(lhs[:pivot])
	nrhs := []hir.Expr{rhs[0]}
	// rebalance
	nlhs = append(nlhs, rhs[1:]...)
	nrhs = append(nrhs, lhs[pivot:]...)
	// done
	return nlhs, nrhs
}
