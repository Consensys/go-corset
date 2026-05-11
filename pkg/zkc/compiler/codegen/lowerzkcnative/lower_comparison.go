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
	"github.com/consensys/go-corset/pkg/util/field"
	"github.com/consensys/go-corset/pkg/zkc/vm"
	"github.com/consensys/go-corset/pkg/zkc/vm/instruction"
	"github.com/consensys/go-corset/pkg/zkc/vm/instruction/opcode"
)

// LowerComparisons rewrites SkipIf instructions with LT/GT/LTEQ/GTEQ conditions
// into arithmetic-only sequences using biased subtraction and sign-bit extraction.
// EQ and NEQ conditions are left unchanged.
// This pass must run after LowerBitwise.
func LowerComparisons[W vm.Word[W]](modules []vm.Module, cfg field.Config) []vm.Module {
	out := append([]vm.Module{}, modules...)

	for i, mod := range out {
		if fn, ok := mod.(*vm.WordFunction); ok {
			out[i] = lowerComparisonFunction[W](fn, cfg.BandWidth)
		}
	}

	return out
}

func lowerComparisonFunction[W vm.Word[W]](fn *vm.WordFunction, bandWidth uint) *vm.WordFunction {
	var (
		code      = fn.Code()
		ncode     = make([]vectorInstruction, len(code))
		registers = append([]register.Register{}, fn.Registers()...)
	)

	for i, insn := range code {
		ncodes := lowerComparisonCodes[W](insn.Codes, &registers, bandWidth)
		ncode[i] = vectorInstruction{Codes: ncodes}
	}

	return vm.NewFunction(fn.Name(), registers, ncode)
}

func lowerComparisonCodes[W vm.Word[W]](
	codes []vm.WordInstruction,
	registers *[]register.Register,
	bandWidth uint,
) []vm.WordInstruction {
	ncodes := make([]vm.WordInstruction, 0, len(codes))

	for _, code := range codes {
		ncodes = append(ncodes, lowerComparisonCode[W](code, registers, bandWidth)...)
	}

	return ncodes
}

func lowerComparisonCode[W vm.Word[W]](
	code vm.WordInstruction,
	registers *[]register.Register,
	bandWidth uint,
) []vm.WordInstruction {
	si, ok := code.(*instruction.SkipIf)
	if !ok || !isRelationalCondition(si.Cond) {
		return []vm.WordInstruction{code}
	}

	return lowerRelationalSkipIf[W](si, registers, bandWidth)
}

func isRelationalCondition(cond opcode.Condition) bool {
	switch cond {
	case opcode.LT, opcode.GT, opcode.LTEQ, opcode.GTEQ:
		return true
	default:
		return false
	}
}

// lowerRelationalSkipIf lowers a SkipIf with a relational condition into an
// arithmetic sequence. For LT(a, b) with widths uA, uB and W = max(uA,uB)+1:
//
//	a_base = cast(a, W-1)                 // zero-extend a to W-1 bits
//	b_wide = cast(b, W)                   // zero-extend b to W bits
//	one    = 1                            // 1-bit constant
//	biased = BitConcat([a_base, one])     // 1::a_base, W bits = a_base + 2^(W-1)
//	diff   = biased - b_wide              // always in [1, 2^W-1], no underflow
//	lo, sign = Destruct(diff)             // sign=1 iff diff >= 2^(W-1) iff a >= b
//	zero   = 0                            // 1-bit constant
//	SkipIf(EQ, sign, zero, skip)          // skip iff sign==0 i.e. a < b
//
// GT and LTEQ are reduced by swapping operands; GTEQ uses NEQ instead of EQ.
func lowerRelationalSkipIf[W vm.Word[W]](
	si *instruction.SkipIf,
	registers *[]register.Register,
	bandWidth uint,
) []vm.WordInstruction {
	lhs, rhs, skipOnZero := normalizeRelational(si)

	lhsWidth := resolveRegisterWidth(*registers, lhs, bandWidth)
	rhsWidth := resolveRegisterWidth(*registers, rhs, bandWidth)

	castBandWidth := lhsWidth
	if rhsWidth > castBandWidth {
		castBandWidth = rhsWidth
	}

	castBandWidth++

	zero := vm.Uint64[W](0)
	one := vm.Uint64[W](1)

	aBase := allocTmp(registers, castBandWidth-1)
	bWide := allocTmp(registers, castBandWidth)
	oneReg := allocTmp(registers, 1)
	biased := allocTmp(registers, castBandWidth)
	diff := allocTmp(registers, castBandWidth)
	lo := allocTmp(registers, castBandWidth-1)
	sign := allocTmp(registers, 1)
	zeroReg := allocTmp(registers, 1)

	insns := []vm.WordInstruction{
		instruction.NewCast(aBase, lhs, castBandWidth-1),
		instruction.NewCast(bWide, rhs, castBandWidth),
		instruction.NewIntAdd(oneReg, nil, one),
		instruction.NewBitConcat[W](biased, []register.Id{aBase, oneReg}),
		instruction.NewIntSub(diff, []register.Id{biased, bWide}, zero),
		instruction.NewDestruct([]register.Id{lo, sign}, diff),
		instruction.NewIntAdd(zeroReg, nil, zero),
	}

	finalCond := opcode.EQ
	if !skipOnZero {
		finalCond = opcode.NEQ
	}

	return append(insns, instruction.NewSkipIf(finalCond, sign, zeroReg, si.Skip))
}

// normalizeRelational returns (lhs, rhs, skipOnZero) for a relational SkipIf,
// normalizing GT and LTEQ by swapping operands into the LT/GTEQ basis:
//
//	LT(a,b)   → lhs=a, rhs=b, skipOnZero=true  (skip if sign==0 i.e. a < b)
//	GTEQ(a,b) → lhs=a, rhs=b, skipOnZero=false (skip if sign==1 i.e. a >= b)
//	GT(a,b)   → lhs=b, rhs=a, skipOnZero=true  (= LT(b,a))
//	LTEQ(a,b) → lhs=b, rhs=a, skipOnZero=false (= GTEQ(b,a))
func normalizeRelational(si *instruction.SkipIf) (lhs, rhs register.Id, skipOnZero bool) {
	switch si.Cond {
	case opcode.LT:
		return si.Left, si.Right, true
	case opcode.GTEQ:
		return si.Left, si.Right, false
	case opcode.GT:
		return si.Right, si.Left, true
	case opcode.LTEQ:
		return si.Right, si.Left, false
	default:
		panic("normalizeRelational called with non-relational condition")
	}
}

// resolveRegisterWidth returns the effective width of a register, using
// bandWidth for native (field-sized) registers.
func resolveRegisterWidth(registers []register.Register, id register.Id, bandWidth uint) uint {
	reg := registers[id.Unwrap()]
	if reg.IsNative() {
		return bandWidth
	}

	return reg.Width()
}
