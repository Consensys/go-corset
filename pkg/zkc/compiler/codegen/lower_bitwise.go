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
package codegen

import (
	"fmt"
	"math/big"

	"github.com/consensys/go-corset/pkg/schema/register"
	"github.com/consensys/go-corset/pkg/util/field"
	"github.com/consensys/go-corset/pkg/zkc/vm/function"
	"github.com/consensys/go-corset/pkg/zkc/vm/instruction"
	"github.com/consensys/go-corset/pkg/zkc/vm/instruction/opcode"
	"github.com/consensys/go-corset/pkg/zkc/vm/machine"
	"github.com/consensys/go-corset/pkg/zkc/vm/word"
)

// LowerBitwise rewrites VM-level bitwise micro-instructions into CALLs to
// helper functions. The helper modules are appended to the returned module
// slice.
func LowerBitwise[W word.Word[W]](modules []machine.Module, cfg field.Config) []machine.Module {
	var (
		out     = append([]machine.Module{}, modules...)
		helpers = newBitwiseHelpers[W](uint(len(out)), cfg)
	)

	for i, mod := range out {
		if fn, ok := mod.(*function.Boot); ok {
			out[i] = lowerBitwiseFunction(fn, helpers)
		}
	}

	return append(out, helpers.modules()...)
}

func lowerBitwiseFunction[W word.Word[W]](
	fn *function.Boot, helpers *bitwiseHelpers[W],
) *function.Boot {
	var (
		code      = fn.Code()
		ncode     = make([]VectorInstruction, len(code))
		registers = append([]register.Register{}, fn.Registers()...)
	)

	for i, insn := range code {
		ncodes := lowerBitwiseCodes(insn.Codes, &registers, helpers)
		ncode[i] = VectorInstruction{Codes: ncodes}
	}

	return function.New(fn.Name(), registers, ncode)
}

func lowerBitwiseCodes[W word.Word[W]](
	codes []instruction.Instruction,
	registers *[]register.Register,
	helpers *bitwiseHelpers[W],
) []instruction.Instruction {
	ncodes := make([]instruction.Instruction, 0, len(codes))

	for _, code := range codes {
		ncodes = append(ncodes, lowerBitwiseCode(code, registers, helpers)...)
	}

	return ncodes
}

func lowerBitwiseCode[W word.Word[W]](
	code instruction.Instruction,
	registers *[]register.Register,
	helpers *bitwiseHelpers[W],
) []instruction.Instruction {
	if !isBitwiseOpcode(code.OpCode()) {
		return []instruction.Instruction{code}
	}

	origWidth, isPowerOfTwo := lowerableWidth(*registers, code.Definitions()[0], helpers.field.BandWidth)

	p := origWidth
	if !isPowerOfTwo {
		p = nextPowerOfTwo(origWidth)
	}

	switch t := code.(type) {
	case *instruction.BitAnd[W]:
		id := helpers.ensure(t.OpCode(), p, len(t.Sources), t.Constant)
		return bitwiseCall(id, t.Target, t.Sources, origWidth, p, registers)
	case *instruction.BitOr[W]:
		id := helpers.ensure(t.OpCode(), p, len(t.Sources), t.Constant)
		return bitwiseCall(id, t.Target, t.Sources, origWidth, p, registers)
	case *instruction.BitXor[W]:
		id := helpers.ensure(t.OpCode(), p, len(t.Sources), t.Constant)
		return bitwiseCall(id, t.Target, t.Sources, origWidth, p, registers)
	case *instruction.BitNot[W]:
		id := helpers.ensure(t.OpCode(), p, len(t.Sources), zeroWord[W]())
		if origWidth == p {
			return bitwiseCall(id, t.Target, t.Sources, origWidth, p, registers)
		}
		// NOT flips all p bits including the zero-padding beyond origWidth,
		// so mask the result back to origWidth before the narrowing cast.
		return bitwiseCallNot[W](id, t.Target, t.Sources[0], origWidth, p, registers)
	case *instruction.BitShl[W]:
		id := helpers.ensure(t.OpCode(), p, len(t.Sources), zeroWord[W]())
		return bitwiseCall(id, t.Target, t.Sources, origWidth, p, registers)
	case *instruction.BitShr[W]:
		id := helpers.ensure(t.OpCode(), p, len(t.Sources), zeroWord[W]())
		return bitwiseCall(id, t.Target, t.Sources, origWidth, p, registers)
	default:
		panic(fmt.Sprintf("unexpected non-bitwise opcode: %d", code.OpCode()))
	}
}

// bitwiseCall emits a call to a bitwise helper module.  When origWidth equals p
// (already a power of two) the call is direct with no temporaries.  Otherwise
// each source is zero-extended from origWidth to p bits via Cast before the
// call, and the p-wide result is truncated back to origWidth bits afterwards.
func bitwiseCall(
	id uint,
	target register.Id,
	sources []register.Id,
	origWidth, p uint,
	registers *[]register.Register,
) []instruction.Instruction {
	if origWidth == p {
		return []instruction.Instruction{
			instruction.NewCall(id, append([]register.Id{}, sources...), []register.Id{target}),
		}
	}

	insns := make([]instruction.Instruction, 0, 2+len(sources))

	pSources := make([]register.Id, len(sources))
	for i, src := range sources {
		pTmp := allocTmp(registers, p)
		insns = append(insns, instruction.NewCast(pTmp, src, p))
		pSources[i] = pTmp
	}

	pResult := allocTmp(registers, p)
	insns = append(insns, instruction.NewCall(id, pSources, []register.Id{pResult}))
	// TODO @Dave: is a cast safe here for truncation?
	insns = append(insns, instruction.NewCast(target, pResult, origWidth))

	return insns
}

// bitwiseCallNot emits the instruction sequence for a NOT with non-power-of-2
// width.  Unlike AND/OR/XOR, NOT flips the zero-padding bits when called on a
// padded-up value, so the result must be masked back to origWidth before the
// narrowing cast.
func bitwiseCallNot[W word.Word[W]](
	id uint,
	target, source register.Id,
	origWidth, p uint,
	registers *[]register.Register,
) []instruction.Instruction {
	var zero W

	maskBig := new(big.Int).Sub(new(big.Int).Lsh(big.NewInt(1), origWidth), big.NewInt(1))
	mask := zero.SetBigInt(maskBig)

	pTmp := allocTmp(registers, p)
	pResult := allocTmp(registers, p)
	pMasked := allocTmp(registers, p)

	return []instruction.Instruction{
		instruction.NewCast(pTmp, source, p),
		instruction.NewCall(id, []register.Id{pTmp}, []register.Id{pResult}),
		instruction.NewBitAnd[W](pMasked, []register.Id{pResult}, mask),
		instruction.NewCast(target, pMasked, origWidth),
	}
}

func allocTmp(registers *[]register.Register, width uint) register.Id {
	var padding big.Int

	id := register.NewId(uint(len(*registers)))
	name := fmt.Sprintf("$cast%d", len(*registers))
	*registers = append(*registers, register.NewComputed(name, width, padding))

	return id
}

func zeroWord[W word.Word[W]]() W {
	var z W
	return z
}

func nextPowerOfTwo(w uint) uint {
	p := uint(1)
	for p < w {
		p <<= 1
	}

	return p
}

func lowerableWidth(registers []register.Register, target register.Id, bandWidth uint) (uint, bool) {
	reg := registers[target.Unwrap()]

	var w uint
	if reg.IsNative() {
		w = bandWidth
	} else {
		w = reg.Width()
	}

	if w == 0 {
		panic(fmt.Sprintf("zero-width register: %s", reg.Name()))
	}

	return w, w&(w-1) == 0
}

type bitwiseHelperKey struct {
	opcode   instruction.OpCode
	width    uint
	arity    int
	constant string
}

type bitwiseHelpers[W word.Word[W]] struct {
	baseID uint
	field  field.Config
	ids    map[bitwiseHelperKey]uint
	items  []machine.Module
}

func newBitwiseHelpers[W word.Word[W]](baseID uint, cfg field.Config) *bitwiseHelpers[W] {
	return &bitwiseHelpers[W]{
		baseID: baseID,
		field:  cfg,
		ids:    make(map[bitwiseHelperKey]uint),
	}
}

func (p *bitwiseHelpers[W]) modules() []machine.Module {
	return p.items
}

func (p *bitwiseHelpers[W]) ensure(op instruction.OpCode, width uint, arity int, constant W) uint {
	key := bitwiseHelperKey{
		opcode:   op,
		width:    width,
		arity:    arity,
		constant: helperConstant(op, constant),
	}

	if id, ok := p.ids[key]; ok {
		return id
	}

	// Sub-helpers (e.g. recursive NOT halves) are appended to p.items inside
	// the factory, so id must be derived after the factory returns.
	mod := newBitwiseHelperModule(p, key, constant)

	id := p.baseID + uint(len(p.items))
	p.items = append(p.items, mod)
	p.ids[key] = id

	return id
}

func helperConstant[W word.Word[W]](op instruction.OpCode, constant W) string {
	switch op {
	case opcode.BIT_AND, opcode.BIT_OR, opcode.BIT_XOR:
		return constant.BigInt().Text(16)
	default:
		return ""
	}
}

func newBitwiseHelperModule[W word.Word[W]](
	helpers *bitwiseHelpers[W],
	key bitwiseHelperKey,
	constant W,
) machine.Module {
	if key.opcode == opcode.BIT_AND || key.opcode == opcode.BIT_OR ||
		key.opcode == opcode.BIT_XOR {
		// Recursive: sub-helpers are appended inside; id is recomputed there.
		return newDecomposedNaryHelper(helpers, key, constant)
	}

	if key.opcode == opcode.BIT_NOT {
		// Recursive: sub-helpers are appended inside; id is recomputed there.
		return newDecomposedNotHelper[W](helpers, key)
	}

	var (
		padding big.Int
		regs    = make([]register.Register, 0, key.arity+1)
		sources = make([]register.Id, key.arity)
	)

	for i := 0; i < key.arity; i++ {
		regs = append(regs, register.NewInput(fmt.Sprintf("arg%d", i+1), key.width, padding))
		sources[i] = register.NewId(uint(i))
	}

	target := register.NewId(uint(key.arity))
	regs = append(regs, register.NewOutput("out", key.width, padding))

	var op instruction.Instruction

	switch key.opcode {
	case opcode.BIT_SHL:
		op = instruction.NewBitShl[W](target, sources[0], sources[1])
	case opcode.BIT_SHR:
		op = instruction.NewBitShr[W](target, sources[0], sources[1])
	default:
		panic(fmt.Sprintf("unsupported bitwise helper opcode: %d", key.opcode))
	}

	code := []instruction.Instruction{op, instruction.NewReturn()}
	name := helperName(key)

	return function.New(name, regs, []VectorInstruction{VectorInstruction{Codes: code}})
}

// newDecomposedNaryHelper builds a helper module for bitwise AND/OR/XOR using
// recursive halving.  Each module body is O(arity) instructions: it splits
// every source and the constant into two half-wide pieces, calls the
// half-wide sub-helpers for each piece, and recombines.  Sub-helpers are
// shared across call sites via the helpers cache, so the total number of
// unique modules is O(log(width)) when the constant is uniform across halves
// (e.g. all-zeros or all-ones masks).
func newDecomposedNaryHelper[W word.Word[W]](
	helpers *bitwiseHelpers[W],
	key bitwiseHelperKey,
	constant W,
) machine.Module {
	b := newHelperBuilder[W](key.width, key.arity)

	out := b.output
	zero := word.Uint64[W](0)

	// TODO: we will want to stop before width == 1 to reduce the number of tiny modules.
	if key.width == 1 {
		// Base case: single-bit operation.  Seed agg with the constant bit then
		// fold each source in using the appropriate pairwise identity.
		one := word.Uint64[W](1)
		agg := b.newComputed("agg")

		if constant.BigInt().Bit(0) == 0 {
			b.emit(instruction.NewIntAdd(agg, nil, zero))
		} else {
			b.emit(instruction.NewIntAdd(agg, nil, one))
		}

		for _, inp := range b.inputs {
			agg = b.combineBit(key.opcode, agg, inp)
		}

		b.emit(instruction.NewIntAdd(out, []register.Id{agg}, zero))
	} else {
		// Recursive case.
		half := key.width / 2

		// Split the constant at generation time.
		constBig := constant.BigInt()
		splitBig := new(big.Int).Lsh(big.NewInt(1), half)
		constLow := constant.SetBigInt(new(big.Int).Mod(constBig, splitBig))
		constHigh := constant.SetBigInt(new(big.Int).Rsh(constBig, half))

		// Ensure sub-helpers for each constant half (may be the same module
		// when constLow == constHigh, e.g. all-zeros or all-ones masks).
		subIDlow := helpers.ensure(key.opcode, half, key.arity, constLow)
		subIDhigh := helpers.ensure(key.opcode, half, key.arity, constHigh)

		lowSrcs := make([]register.Id, key.arity)
		highSrcs := make([]register.Id, key.arity)

		for i, arg := range b.inputs {
			lo := b.newComputedNamed(fmt.Sprintf("low%d", i+1), half)
			hi := b.newComputedNamed(fmt.Sprintf("high%d", i+1), half)
			b.emit(instruction.NewDestruct([]register.Id{lo, hi}, arg))
			lowSrcs[i] = lo
			highSrcs[i] = hi
		}

		resLow := b.newComputedNamed("rlow", half)
		resHigh := b.newComputedNamed("rhigh", half)

		b.emit(instruction.NewCall(subIDlow, lowSrcs, []register.Id{resLow}))
		b.emit(instruction.NewCall(subIDhigh, highSrcs, []register.Id{resHigh}))

		b.emit(instruction.NewBitConcat[W](out, []register.Id{resLow, resHigh}))
	}

	b.emit(instruction.NewReturn())

	// Sub-helpers (if any) have already been appended; our slot is next.
	name := helperName(key)

	return function.New(name, b.regs(), []VectorInstruction{{Codes: b.code}})
}

// newDecomposedNotHelper builds a helper module that computes bitwise NOT using
// recursive halving rather than bit-by-bit iteration.  For a width-2^n input
// the module body is O(1) instructions: it splits into two half-wide halves,
// calls the width-2^(n-1) NOT helper for each, and recombines.  This keeps
// every individual module body small while the shared sub-helpers are reused
// across all call sites.
func newDecomposedNotHelper[W word.Word[W]](helpers *bitwiseHelpers[W], key bitwiseHelperKey) machine.Module {
	b := newHelperBuilder[W](key.width, key.arity)

	out := b.output
	zero := word.Uint64[W](0)

	// TODO: we will want to stop before width == 1 to reduce the number of tiny modules.
	if key.width == 1 {
		// Base case: NOT of a single bit = 1 - bit.
		one := word.Uint64[W](1)
		oneReg := b.newComputed("one")
		b.emit(instruction.NewIntAdd(oneReg, nil, one))
		b.emit(instruction.NewIntSub(out, []register.Id{oneReg, b.inputs[0]}, zero))
	} else {
		// Recursive case: split into two half-wide halves, NOT each, recombine.
		half := key.width / 2

		var zeroW W

		subID := helpers.ensure(opcode.BIT_NOT, half, 1, zeroW)

		low := b.newComputedNamed("low", half)
		high := b.newComputedNamed("high", half)
		b.emit(instruction.NewDestruct([]register.Id{low, high}, b.inputs[0]))

		nlow := b.newComputedNamed("nlow", half)
		nhigh := b.newComputedNamed("nhigh", half)

		b.emit(instruction.NewCall(subID, []register.Id{low}, []register.Id{nlow}))
		b.emit(instruction.NewCall(subID, []register.Id{high}, []register.Id{nhigh}))

		b.emit(instruction.NewBitConcat[W](out, []register.Id{nlow, nhigh}))
	}

	b.emit(instruction.NewReturn())

	// Sub-helpers (if any) have already been appended; our slot is next.
	name := helperName(key)

	return function.New(name, b.regs(), []VectorInstruction{VectorInstruction{Codes: b.code}})
}

type helperBuilder[W word.Word[W]] struct {
	width   uint
	inputs  []register.Id
	output  register.Id
	base    []register.Register
	code    []instruction.Instruction
	nextTmp uint
}

func newHelperBuilder[W word.Word[W]](width uint, arity int) *helperBuilder[W] {
	var (
		padding big.Int
		base    = make([]register.Register, 0, arity+1)
		inputs  = make([]register.Id, arity)
	)

	for i := 0; i < arity; i++ {
		inputs[i] = register.NewId(uint(i))
		base = append(base, register.NewInput(fmt.Sprintf("arg%d", i+1), width, padding))
	}

	output := register.NewId(uint(arity))

	base = append(base, register.NewOutput("out", width, padding))

	return &helperBuilder[W]{
		width:  width,
		inputs: inputs,
		output: output,
		base:   base,
	}
}

func (p *helperBuilder[W]) regs() []register.Register {
	return p.base
}

func (p *helperBuilder[W]) emit(insn instruction.Instruction) {
	p.code = append(p.code, insn)
}

func (p *helperBuilder[W]) newComputed(prefix string) register.Id {
	return p.newComputedWidth(prefix, p.width)
}

func (p *helperBuilder[W]) newComputedWidth(prefix string, width uint) register.Id {
	var padding big.Int

	id := register.NewId(uint(len(p.base)))
	name := fmt.Sprintf("%s%d", prefix, p.nextTmp)
	p.base = append(p.base, register.NewComputed(name, width, padding))
	p.nextTmp++

	return id
}

func (p *helperBuilder[W]) newComputedNamed(name string, width uint) register.Id {
	var padding big.Int

	id := register.NewId(uint(len(p.base)))
	p.base = append(p.base, register.NewComputed(name, width, padding))

	return id
}

func (p *helperBuilder[W]) combineBit(op instruction.OpCode, lhs, rhs register.Id) register.Id {
	zero := word.Uint64[W](0)
	one := word.Uint64[W](1)

	switch op {
	case opcode.BIT_AND:
		res := p.newComputed("and")
		p.emit(instruction.NewIntMul(res, []register.Id{lhs, rhs}, one))

		return res
	case opcode.BIT_OR:
		// a + (1-a)*b avoids the intermediate overflow of (a+b) when a=b=1
		oneReg := p.newComputed("or_one")
		p.emit(instruction.NewIntAdd(oneReg, nil, one))

		na := p.newComputed("or_na")
		p.emit(instruction.NewIntSub(na, []register.Id{oneReg, lhs}, zero))

		prod := p.newComputed("or_prod")
		p.emit(instruction.NewIntMul(prod, []register.Id{na, rhs}, one))

		res := p.newComputed("or")
		p.emit(instruction.NewIntAdd(res, []register.Id{lhs, prod}, zero))

		return res
	case opcode.BIT_XOR:
		// a*(1-b) + (1-a)*b avoids intermediate overflow when a=b=1
		oneReg := p.newComputed("xor_one")
		p.emit(instruction.NewIntAdd(oneReg, nil, one))

		nb := p.newComputed("xor_nb")
		p.emit(instruction.NewIntSub(nb, []register.Id{oneReg, rhs}, zero))

		na := p.newComputed("xor_na")
		p.emit(instruction.NewIntSub(na, []register.Id{oneReg, lhs}, zero))

		l := p.newComputed("xor_l")
		p.emit(instruction.NewIntMul(l, []register.Id{lhs, nb}, one))

		r := p.newComputed("xor_r")
		p.emit(instruction.NewIntMul(r, []register.Id{na, rhs}, one))

		res := p.newComputed("xor")
		p.emit(instruction.NewIntAdd(res, []register.Id{l, r}, zero))

		return res
	default:
		panic(fmt.Sprintf("unsupported bit combine opcode: %d", op))
	}
}

// HasBitwiseOps checks whether any module contains VM bitwise instructions.
func HasBitwiseOps(modules []machine.Module) bool {
	for _, mod := range modules {
		fn, ok := mod.(*function.Boot)
		if !ok {
			continue
		}

		for _, insn := range fn.Code() {
			for _, code := range insn.Codes {
				if isBitwiseOpcode(code.OpCode()) {
					return true
				}
			}
		}
	}

	return false
}

func isBitwiseOpcode(op instruction.OpCode) bool {
	switch op {
	case opcode.BIT_AND, opcode.BIT_OR, opcode.BIT_XOR,
		opcode.BIT_NOT, opcode.BIT_SHL, opcode.BIT_SHR:
		return true
	default:
		return false
	}
}

func helperName(key bitwiseHelperKey) string {
	var op string

	switch key.opcode {
	case opcode.BIT_AND:
		op = "and"
	case opcode.BIT_OR:
		op = "or"
	case opcode.BIT_XOR:
		op = "xor"
	case opcode.BIT_NOT:
		op = "not"
	case opcode.BIT_SHL:
		op = "shl"
	case opcode.BIT_SHR:
		op = "shr"
	default:
		op = "unknown"
	}

	return fmt.Sprintf("$bit_%s_u%d", op, key.width)
}
