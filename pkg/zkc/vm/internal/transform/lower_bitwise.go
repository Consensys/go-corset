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
package transform

import (
	"fmt"
	"math/big"
	"slices"

	"github.com/LFDT-Lineth/zkc/pkg/util"
	"github.com/LFDT-Lineth/zkc/pkg/zkc/vm/internal/bytecode"
	"github.com/LFDT-Lineth/zkc/pkg/zkc/vm/internal/descriptor"
	"github.com/LFDT-Lineth/zkc/pkg/zkc/vm/internal/transform/split"
	"github.com/LFDT-Lineth/zkc/pkg/zkc/vm/internal/word"
)

// LowerBitwise rewrites VM-level NOT and SHL/SHR bytecodes: NOT is inlined as
// (MASK - x), while SHL/SHR become CALLs to helper functions whose modules are
// appended to the returned program.  AND/OR/XOR are left untouched here and are
// lowered after register splitting instead (see LowerOrXorAnd).
//
// We assume this lowering happens BEFORE vectorization and register splitting.
func LowerBitwise[W word.Word[W]](program descriptor.Program[W]) descriptor.Program[W] {
	var (
		out     = slices.Clone(program.Modules())
		helpers = newShiftHelpers[W](uint(len(out)), scanShiftParams(out))
	)

	for i, mod := range out {
		if fn, ok := mod.(*descriptor.Function[W]); ok {
			out[i] = lowerBitwiseFunction(fn, func(b Bytecode[W], alloc split.Allocator[W]) []Bytecode[W] {
				return lowerBitwiseCode(b, alloc, helpers)
			})
		}
	}

	return descriptor.NewProgram(
		program.Field(),
		program.MaxStaticHeight(),
		append(out, helpers.modules()...)...)
}

// lowerBitwiseFunction rewrites each bytecode of a function via codeFn,
// threading a per-function register allocator (for any temporaries codeFn
// introduces).  The helper registry, if any, is captured by the codeFn closure
// — this driver is shared by LowerBitwise and LowerOrXorAnd, whose registries
// differ.
func lowerBitwiseFunction[W word.Word[W]](fn *descriptor.Function[W],
	codeFn func(Bytecode[W], split.Allocator[W]) []Bytecode[W],
) *descriptor.Function[W] {
	var (
		vectors = fn.Vectors()
		nvecs   = make([]BytecodeVector[W], len(vectors))
		alloc   = split.NewAllocator(fn)
	)

	for i, vec := range vectors {
		nvecs[i] = vec.Map(func(_ uint, b Bytecode[W]) []Bytecode[W] {
			return codeFn(b, alloc)
		})
	}

	return descriptor.NewFunction(fn.Name(), alloc.Registers(), fn.Kind(), fn.Effects(), nvecs)
}

func lowerBitwiseCode[W word.Word[W]](
	b Bytecode[W],
	registers split.Allocator[W],
	helpers *shiftHelpers[W],
) []Bytecode[W] {
	//
	bw, ok := b.(*bytecode.Bitwise[W])
	if !ok {
		return []Bytecode[W]{b}
	}
	//
	switch bw.Op {
	case bytecode.OP_NOT:
		return inlineBitwiseNot(bw, registers)
	case bytecode.OP_SHL, bytecode.OP_SHR:
		return lowerBitwiseShlShr(bw, registers, helpers)
	default:
		// AND/OR/XOR are lowered after register splitting; see LowerOrXorAnd.
		return []Bytecode[W]{b}
	}
}

// lowerBitwiseShlShr rewrites a SHL/SHR into a call to the shared cascade,
// which operates at helpers.maxWidth() and returns both directions as
// (out_shl, out_shr).  A call site therefore has to adapt in two ways:
//
//   - Width: a value narrower than maxWidth is zero-extended on the way in and
//     the result truncated on the way out.  This is sound in both directions —
//     for SHR the extended operand is already zero above bit w, and for SHL the
//     bits truncation drops are exactly those that overflowed w anyway.  It
//     also subsumes out-of-range amounts: n >= w pushes every bit of a w-bit
//     value out of the low w bits, so the truncated result is zero.
//   - Direction: a partial call binds only the requested output and discards
//     the other.
func lowerBitwiseShlShr[W word.Word[W]](
	b *bytecode.Bitwise[W],
	registers split.Allocator[W],
	helpers *shiftHelpers[W],
) []Bytecode[W] {
	var (
		// NOTE: the shift amount is always a register (constant operands are
		// only supported for AND/OR/XOR).
		amount = b.Right.AsRegister()
		// NOTE: bitwidth of shift (e.g. "x << y") determined by width of first
		// argument only (i.e. "x").
		amtWidth = registers.Registers()[amount].Bitwidth().Unwrap()
		width    = uint(b.Bitwidth)
		maxWidth = helpers.maxWidth()
		id       = helpers.ensureShift(amtWidth)
		narrow   = width < maxWidth
		zero     W
		code     []Bytecode[W]
	)
	// Zero-extend the value into the cascade's width.
	value := b.Left
	if narrow {
		value = registers.Allocate("", util.Some(maxWidth))
		code = append(code, bytecode.AddConst(value, []bytecode.RegisterId{b.Left}, zero))
	}
	// Bind only the requested direction, discarding the other.
	result := b.Target
	if narrow {
		result = registers.Allocate("", util.Some(maxWidth))
	}
	//
	returns := []bytecode.RegisterId{result, bytecode.DISCARD}
	if b.Op == bytecode.OP_SHR {
		returns = []bytecode.RegisterId{bytecode.DISCARD, result}
	}
	//
	code = append(code, bytecode.CallFun[W](uint16(id),
		[]bytecode.RegisterId{value, amount}, returns))
	// Truncate back down by destructing off the high bits (little-endian, so
	// the target takes the low width bits).
	if narrow {
		high := registers.Allocate("", util.Some(maxWidth-width))
		code = append(code, bytecode.AddVec[W](
			[]bytecode.RegisterId{b.Target, high}, []bytecode.RegisterId{result}))
	}

	return code
}

// inlineBitwiseNot emits ~x as (MASK - x) directly into the caller's bytecode
// stream, where MASK = 2^width - 1.  No helper module is created.
func inlineBitwiseNot[W word.Word[W]](b *bytecode.Bitwise[W], registers split.Allocator[W]) []Bytecode[W] {
	var (
		width, _ = maxBitwidthOf(registers.Registers(), b.Uses()...)
		maskBig  = new(big.Int).Sub(new(big.Int).Lsh(big.NewInt(1), width), big.NewInt(1))
		zeroW    W
		mask     = zeroW.SetBigInt(maskBig)
		zero     W
	)

	maskReg := registers.Allocate("", util.Some(width))
	// TODO: CSUB, see: https://github.com/LFDT-Lineth/zkc/issues/2062
	return []Bytecode[W]{
		bytecode.LoadConst(maskReg, mask),
		bytecode.SubConst(b.Target, []bytecode.RegisterId{maskReg, b.Left}, zero),
	}
}

func maxBitwidthOf[W word.Word[W]](regs []descriptor.Register[W], targets ...bytecode.RegisterId) (uint, bool) {
	var w uint
	//
	for _, src := range targets {
		reg := regs[src]

		if reg.IsNative() {
			panic("unexpected native register in bitwise lowering")
		} else if reg.Bitwidth().Unwrap() == 0 {
			panic(fmt.Sprintf("zero-width register: %s", reg.Name()))
		}
		//
		w = max(w, reg.Bitwidth().Unwrap())
	}

	return w, w&(w-1) == 0
}

// bitwiseOpName is the short name used in helper module names for a bitwise
// operation.  SHL/SHR have no entry: both directions are served by a single
// merged cascade whose name is fixed (see shiftHelperName).
func bitwiseOpName(op bytecode.Operation) string {
	switch op {
	case bytecode.OP_AND:
		return "and"
	case bytecode.OP_OR:
		return "or"
	case bytecode.OP_XOR:
		return "xor"
	case bytecode.OP_NOT:
		return "not"
	default:
		panic(fmt.Sprintf("unexpected op: %v", op))
	}
}
