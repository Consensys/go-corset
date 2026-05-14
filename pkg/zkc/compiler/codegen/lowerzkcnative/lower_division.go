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
package lowerzkcnative

import (
	"github.com/consensys/go-corset/pkg/schema/register"
	"github.com/consensys/go-corset/pkg/zkc/vm"
	"github.com/consensys/go-corset/pkg/zkc/vm/instruction"
	"github.com/consensys/go-corset/pkg/zkc/vm/instruction/opcode"
)

// LowerDivisions rewrites INT_DIV and INT_REM instructions into a
// non-deterministic hint followed by arithmetic validation:
//
//	FieldHint{targets:[q, r], sources:[x, y]}  // prover fills q=x/y, r=x%y
//	sum = q * y
//	sum = sum + r
//	SkipIf(EQ, sum, x, 1)
//	Fail
//	SkipIf(LT, r, y, 1)                        // expanded later by LowerComparisons
//	Fail
//
// This pass must run before LowerComparisons.
func LowerDivisions[W vm.Word[W]](modules []vm.Module) []vm.Module {
	out := append([]vm.Module{}, modules...)

	for i, mod := range out {
		if fn, ok := mod.(*vm.WordFunction); ok {
			out[i] = lowerDivisionFunction[W](fn)
		}
	}

	return out
}

func lowerDivisionFunction[W vm.Word[W]](fn *vm.WordFunction) *vm.WordFunction {
	var (
		code      = fn.Code()
		ncode     = make([]vectorInstruction, len(code))
		registers = append([]register.Register{}, fn.Registers()...)
	)

	for i, insn := range code {
		ncodes := lowerDivisionCodes[W](insn.Codes, &registers)
		ncode[i] = vectorInstruction{Codes: ncodes}
	}

	return vm.NewFunction(fn.Name(), fn.IsNative(), registers, ncode)
}

func lowerDivisionCodes[W vm.Word[W]](
	codes []vm.WordInstruction,
	registers *[]register.Register,
) []vm.WordInstruction {
	ncodes := make([]vm.WordInstruction, 0, len(codes))

	for _, code := range codes {
		ncodes = append(ncodes, lowerDivisionCode[W](code, registers)...)
	}

	return ncodes
}

func lowerDivisionCode[W vm.Word[W]](
	code vm.WordInstruction,
	registers *[]register.Register,
) []vm.WordInstruction {
	switch code.OpCode() {
	case opcode.INT_DIV:
		insn := code.(*instruction.IntDiv[W])
		return expandDivision[W](insn.Target, insn.Sources[0], insn.Sources[1], registers)
	case opcode.INT_REM:
		insn := code.(*instruction.IntRem[W])
		return expandRemainder[W](insn.Target, insn.Sources[0], insn.Sources[1], registers)
	default:
		return []vm.WordInstruction{code}
	}
}

// expandDivision replaces INT_DIV(q, x, y) with the hint+validation sequence.
// Since q*y + rTmp = x, sum fits in width(x) bits; rTmp < y fits in width(y) bits.
// Constant divisors may have a smaller bitwidth than the dividend (e.g. "2" typed as u2),
// so we derive sum width from x, not y.
func expandDivision[W vm.Word[W]](q, x, y register.Id, registers *[]register.Register) []vm.WordInstruction {
	var (
		nX   = resolveRegisterWidth(*registers, x, 0)
		nY   = resolveRegisterWidth(*registers, y, 0)
		zero = vm.Uint64[W](0)
		one  = vm.Uint64[W](1)
		rTmp = allocTmp(registers, nY)
		sum  = allocTmp(registers, nX)
	)

	return []vm.WordInstruction{
		instruction.NewFieldHint([]register.Id{q, rTmp}, []register.Id{x, y}),
		instruction.NewIntMul(sum, []register.Id{q, y}, one),
		instruction.NewIntAdd(sum, []register.Id{sum, rTmp}, zero),
		instruction.NewSkipIf(opcode.EQ, sum, x, 1),
		instruction.NewFail(),
		instruction.NewSkipIf(opcode.LT, rTmp, y, 1),
		instruction.NewFail(),
	}
}

// expandRemainder replaces INT_REM(r, x, y) with the hint+validation sequence.
// Since qTmp*y + r = x, sum fits in width(x) bits; r < y fits in width(y) bits.
// Constant divisors may have a smaller bitwidth than the dividend (e.g. "2" typed as u2),
// so we derive sum and qTmp widths from x, not y.
func expandRemainder[W vm.Word[W]](r, x, y register.Id, registers *[]register.Register) []vm.WordInstruction {
	var (
		nX   = resolveRegisterWidth(*registers, x, 0)
		zero = vm.Uint64[W](0)
		one  = vm.Uint64[W](1)
		qTmp = allocTmp(registers, nX)
		sum  = allocTmp(registers, nX)
	)

	return []vm.WordInstruction{
		instruction.NewFieldHint([]register.Id{qTmp, r}, []register.Id{x, y}),
		instruction.NewIntMul(sum, []register.Id{qTmp, y}, one),
		instruction.NewIntAdd(sum, []register.Id{sum, r}, zero),
		instruction.NewSkipIf(opcode.EQ, sum, x, 1),
		instruction.NewFail(),
		instruction.NewSkipIf(opcode.LT, r, y, 1),
		instruction.NewFail(),
	}
}
