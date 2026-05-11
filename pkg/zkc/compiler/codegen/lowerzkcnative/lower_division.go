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

// LowerDivision replaces each INT_DIV / INT_REM instruction with an inline
// witness-and-constraint sequence.  Each replacement occupies its own vector
// slot so that Jmp targets are in slot-level units, consistent with the rest
// of the pre-vectorisation instruction stream.
//
// For a division x / y where width(x) = n_x and width(y) = n_y:
//   - quotient width:  n_q    = |n_x - n_y|
//   - remainder width: n_y    (remainder is always < y)
//   - product width:   n_prod = n_y + width(q)
//
// The expansion (8 slots per division):
//
//	[+0]  r      = INT_REM(x, y)              // witness; aborts on y=0
//	[+1]  q      = INT_DIV(x, y)              // witness; aborts on y=0
//	[+2]  y_wide = CAST(y, n_prod)
//	[+3]  q_wide = CAST(q, n_prod)
//	[+4]  r_wide = CAST(r, n_prod)
//	[+5]  prod   = INT_MUL(y_wide, q_wide)    // constraint: prod = y*q
//	[+6]  x_wide = CAST(x, n_prod)            // constraint: x_wide = zero_extend(x)
//	[+7]  x_wide = INT_ADD(r_wide, prod)      // constraint: x_wide = r + y*q
//	                                           // +6 and +7 together enforce x = r + y*q
//
// Jmp targets in non-division slots are remapped to account for the 7 extra
// slots inserted per division expansion.
//
// Division by zero is caught natively by INT_REM / INT_DIV.
func LowerDivision[W vm.Word[W]](modules []vm.Module) []vm.Module {
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
		registers = append([]register.Register{}, fn.Registers()...)
	)

	// Pre-pass: build old slot index → new slot index mapping.
	// Each non-division slot maps 1:1; each division slot expands to 11.
	oldToNew := buildSlotMapping[W](code)

	// Main pass: expand division slots and remap Jmps in all slots.
	ncode := make([]vectorInstruction, 0, int(oldToNew[len(code)]))

	for _, insn := range code {
		if !containsDivision[W](insn.Codes) {
			ncode = append(ncode, vectorInstruction{Codes: remapJmps(insn.Codes, oldToNew)})
		} else {
			ncode = append(ncode, expandDivisionSlot[W](insn.Codes, oldToNew, &registers)...)
		}
	}

	return vm.NewFunction(fn.Name(), registers, ncode)
}

// buildSlotMapping returns a mapping from old slot index to new slot index.
// oldToNew[len(code)] holds the total number of new slots.
func buildSlotMapping[W vm.Word[W]](code []vectorInstruction) []uint {
	oldToNew := make([]uint, len(code)+1)
	newIdx := uint(0)

	for i, insn := range code {
		oldToNew[i] = newIdx
		newIdx += slotExpansionSize[W](insn.Codes)
	}

	oldToNew[len(code)] = newIdx

	return oldToNew
}

// slotExpansionSize returns the number of vector slots that a slot's codes
// will expand to.  Non-division slots stay at 1.  Division slots produce:
// 11 per division instruction, 1 per non-division instruction that precedes
// the last division, and 1 for all non-division instructions that follow the
// last division (they are kept together to preserve SKIP_IF+Jmp+Jmp grouping).
func slotExpansionSize[W vm.Word[W]](codes []vm.WordInstruction) uint {
	lastDivIdx := lastDivisionIndex[W](codes)

	if lastDivIdx < 0 {
		return 1
	}

	size := uint(0)

	for i, c := range codes {
		switch c.(type) {
		case *instruction.IntDiv[W], *instruction.IntRem[W]:
			size += 8
		default:
			if i > lastDivIdx {
				// All trailing non-division codes collapse to one slot.
				return size + 1
			}

			size++ // pre-division: own slot
		}
	}

	return size
}

// lastDivisionIndex returns the index of the last INT_DIV or INT_REM in codes,
// or -1 if none is found.
func lastDivisionIndex[W vm.Word[W]](codes []vm.WordInstruction) int {
	idx := -1

	for i, c := range codes {
		switch c.(type) {
		case *instruction.IntDiv[W], *instruction.IntRem[W]:
			idx = i
		}
	}

	return idx
}

// containsDivision reports whether any code in the slice is INT_DIV or INT_REM.
func containsDivision[W vm.Word[W]](codes []vm.WordInstruction) bool {
	return lastDivisionIndex[W](codes) >= 0
}

// remapJmps returns a copy of codes with each Jmp target remapped through
// oldToNew.  SKIP / SKIP_IF use relative offsets so they are left unchanged.
// Returns codes unchanged if there are no Jmps (common case).
func remapJmps(codes []vm.WordInstruction, oldToNew []uint) []vm.WordInstruction {
	hasJmp := false

	for _, c := range codes {
		if c.OpCode() == opcode.JUMP {
			hasJmp = true

			break
		}
	}

	if !hasJmp {
		return codes
	}

	result := make([]vm.WordInstruction, len(codes))
	copy(result, codes)

	for i, c := range result {
		if c.OpCode() == opcode.JUMP {
			jmp := c.(*instruction.Jmp)
			result[i] = instruction.NewJmp(oldToNew[jmp.Immediate])
		}
	}

	return result
}

// expandDivisionSlot processes the micro-instructions of one vector slot.
// If the slot contains no INT_DIV or INT_REM it is returned unchanged.
// A slot containing INT_DIV / INT_REM is replaced by multiple slots:
//   - Each non-division code before the last division gets its own slot.
//   - Each INT_DIV / INT_REM produces 11 slots via inlineDivRem.
//   - All non-division codes after the last division are kept in ONE slot so
//     that SKIP_IF+Jmp+Jmp condition sequences remain intact for vectorisation.
//
// baseSlot is the absolute index in the function's code where this slot begins.
// oldToNew is used to remap Jmp targets in non-division codes within the slot.
func expandDivisionSlot[W vm.Word[W]](
	codes []vm.WordInstruction,
	oldToNew []uint,
	registers *[]register.Register,
) []vectorInstruction {
	lastDiv := lastDivisionIndex[W](codes)

	var result []vectorInstruction

	for i, code := range codes {
		switch t := code.(type) {
		case *instruction.IntDiv[W]:
			result = append(result, inlineDivRem[W](t.Target, t.Sources[0], t.Sources[1], true, registers)...)
		case *instruction.IntRem[W]:
			result = append(result, inlineDivRem[W](t.Target, t.Sources[0], t.Sources[1], false, registers)...)
		default:
			if i > lastDiv {
				// Batch all remaining trailing codes into one slot so that
				// SKIP_IF+Jmp+Jmp sequences from compileCondition stay together.
				result = append(result, vectorInstruction{Codes: remapJmps(codes[i:], oldToNew)})

				return result
			}
			// Pre-division: each code gets its own slot.
			result = append(result, vectorInstruction{Codes: remapJmps([]vm.WordInstruction{code}, oldToNew)})
		}
	}

	return result
}

// absDiff returns the absolute difference of two unsigned integers.
func absDiff(a, b uint) uint {
	if a >= b {
		return a - b
	}

	return b - a
}

// inlineDivRem returns the 8-slot sequence that replaces one binary INT_DIV
// or INT_REM.  wantQuotient=true means target holds q; false means target holds
// r.  Correctness is enforced by two polynomial constraints on x_wide:
// CAST(x_wide, x, n_prod) and INT_ADD(x_wide, [r_wide, prod], 0) together
// enforce x = r + y*q in the ZK proof system.
func inlineDivRem[W vm.Word[W]](
	target, x, y register.Id,
	wantQuotient bool,
	registers *[]register.Register,
) []vectorInstruction {
	nX := (*registers)[x.Unwrap()].Width()
	nY := (*registers)[y.Unwrap()].Width()
	nQ := absDiff(nX, nY) // |n_x - n_y|

	var q, r register.Id

	if wantQuotient {
		q = target // target already carries the type-system quotient width
		r = allocTmp(registers, nY)
	} else {
		q = allocTmp(registers, nQ)
		r = target // target already carries the type-system remainder width
	}

	nProd := nY + (*registers)[q.Unwrap()].Width()

	yWide := allocTmp(registers, nProd)
	qWide := allocTmp(registers, nProd)
	rWide := allocTmp(registers, nProd)
	xWide := allocTmp(registers, nProd)
	prod := allocTmp(registers, nProd)

	var zero W

	return []vectorInstruction{
		// +0, +1: non-deterministic witnesses; abort if y == 0.
		{Codes: []vm.WordInstruction{instruction.NewIntRem[W](r, x, y)}},
		{Codes: []vm.WordInstruction{instruction.NewIntDiv[W](q, x, y)}},
		// +2 .. +4: widen operands to n_prod bits.
		{Codes: []vm.WordInstruction{instruction.NewCast(yWide, y, nProd)}},
		{Codes: []vm.WordInstruction{instruction.NewCast(qWide, q, nProd)}},
		{Codes: []vm.WordInstruction{instruction.NewCast(rWide, r, nProd)}},
		// +5: prod = y * q
		{Codes: []vm.WordInstruction{instruction.NewIntMul(prod, []register.Id{yWide, qWide}, zero)}},
		// +6, +7: two definitions of x_wide enforce x = r + y*q in ZK.
		{Codes: []vm.WordInstruction{instruction.NewCast(xWide, x, nProd)}},
		{Codes: []vm.WordInstruction{instruction.NewIntAdd(xWide, []register.Id{rWide, prod}, zero)}},
	}
}
