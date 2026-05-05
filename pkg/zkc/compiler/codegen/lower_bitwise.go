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
		ncodes := lowerBitwiseCodes(insn.Codes, registers, helpers)
		ncode[i] = VectorInstruction{Codes: ncodes}
	}

	return function.New(fn.Name(), registers, ncode)
}

func lowerBitwiseCodes[W word.Word[W]](
	codes []instruction.Instruction,
	registers []register.Register,
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
	registers []register.Register,
	helpers *bitwiseHelpers[W],
) []instruction.Instruction {
	if !isBitwiseOpcode(code.OpCode()) {
		return []instruction.Instruction{code}
	}

	width, powerOfTwo := lowerableWidth(registers, code.Definitions()[0], helpers.field.BandWidth)

	if !powerOfTwo {
		width = nextPowerOfTwo(width)
	}
	switch t := code.(type) {
	case *instruction.BitAnd[W]:
		id := helpers.ensure(t.OpCode(), width, len(t.Sources), t.Constant)

		return []instruction.Instruction{
			instruction.NewCall(id, append([]register.Id{}, t.Sources...), []register.Id{t.Target}),
		}
	case *instruction.BitOr[W]:
		id := helpers.ensure(t.OpCode(), width, len(t.Sources), t.Constant)

		return []instruction.Instruction{
			instruction.NewCall(id, append([]register.Id{}, t.Sources...), []register.Id{t.Target}),
		}
	case *instruction.BitXor[W]:
		id := helpers.ensure(t.OpCode(), width, len(t.Sources), t.Constant)

		return []instruction.Instruction{
			instruction.NewCall(id, append([]register.Id{}, t.Sources...), []register.Id{t.Target}),
		}
	case *instruction.BitNot[W]:
		id := helpers.ensure(t.OpCode(), width, len(t.Sources), zeroWord[W]())

		return []instruction.Instruction{
			instruction.NewCall(id, append([]register.Id{}, t.Sources...), []register.Id{t.Target}),
		}
	case *instruction.BitShl[W]:
		id := helpers.ensure(t.OpCode(), width, len(t.Sources), zeroWord[W]())

		return []instruction.Instruction{
			instruction.NewCall(id, append([]register.Id{}, t.Sources...), []register.Id{t.Target}),
		}
	case *instruction.BitShr[W]:
		id := helpers.ensure(t.OpCode(), width, len(t.Sources), zeroWord[W]())

		return []instruction.Instruction{
			instruction.NewCall(id, append([]register.Id{}, t.Sources...), []register.Id{t.Target}),
		}
	default:
		panic(fmt.Sprintf("unexpected non-bitwise opcode: %d", code.OpCode()))
	}
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

	id := p.baseID + uint(len(p.items))
	mod := newBitwiseHelperModule(id, key, constant)

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
	id uint,
	key bitwiseHelperKey,
	constant W,
) machine.Module {
	if key.opcode == opcode.BIT_AND || key.opcode == opcode.BIT_OR ||
		key.opcode == opcode.BIT_XOR {
		return newDecomposedNaryHelper(id, key, constant)
	}

	if key.opcode == opcode.BIT_NOT {
		return newDecomposedNotHelper[W](id, key)
	}

	var (
		padding big.Int
		regs    = make([]register.Register, 0, key.arity+1)
		sources = make([]register.Id, key.arity)
	)

	for i := 0; i < key.arity; i++ {
		regs = append(regs, register.NewInput(fmt.Sprintf("arg%d", i), key.width, padding))
		sources[i] = register.NewId(uint(i))
	}

	target := register.NewId(uint(key.arity))
	regs = append(regs, register.NewOutput("out", key.width, padding))

	var op instruction.Instruction

	switch key.opcode {
	case opcode.BIT_AND:
		op = instruction.NewBitAnd(target, sources, constant)
	case opcode.BIT_OR:
		op = instruction.NewBitOr(target, sources, constant)
	case opcode.BIT_XOR:
		op = instruction.NewBitXor(target, sources, constant)
	case opcode.BIT_NOT:
		op = instruction.NewBitNot[W](target, sources[0])
	case opcode.BIT_SHL:
		op = instruction.NewBitShl[W](target, sources[0], sources[1])
	case opcode.BIT_SHR:
		op = instruction.NewBitShr[W](target, sources[0], sources[1])
	default:
		panic(fmt.Sprintf("unsupported bitwise helper opcode: %d", key.opcode))
	}

	code := []instruction.Instruction{op, instruction.NewReturn()}
	name := helperName(id, key)

	return function.New(name, regs, []VectorInstruction{VectorInstruction{Codes: code}})
}

func newDecomposedNaryHelper[W word.Word[W]](
	id uint,
	key bitwiseHelperKey,
	constant W,
) machine.Module {
	b := newHelperBuilder[W](key.width, key.arity)

	out := b.output
	zero := word.Uint64[W](0)
	one := word.Uint64[W](1)
	two := word.Uint64[W](2)

	divisor := b.newComputed("d")
	b.emit(instruction.NewIntAdd(divisor, nil, two))
	b.emit(instruction.NewIntAdd(out, nil, zero))

	pow := b.newComputed("pow")
	b.emit(instruction.NewIntAdd(pow, nil, one))

	work := make([]register.Id, len(b.inputs))
	for i, arg := range b.inputs {
		w := b.newComputed("w")
		b.emit(instruction.NewIntAdd(w, []register.Id{arg}, zero))
		work[i] = w
	}

	for i := uint(0); i < key.width; i++ {
		agg := b.newComputed("agg")
		if constant.BigInt().Bit(int(i)) == 0 {
			b.emit(instruction.NewIntAdd(agg, nil, zero))
		} else {
			b.emit(instruction.NewIntAdd(agg, nil, one))
		}

		for j := range work {
			bit := b.newComputed("bit")
			b.emit(instruction.NewIntRem[W](bit, work[j], divisor))

			next := b.newComputed("next")
			b.emit(instruction.NewIntDiv[W](next, work[j], divisor))
			work[j] = next

			agg = b.combineBit(key.opcode, agg, bit)
		}

		scaled := b.newComputed("scaled")
		b.emit(instruction.NewIntMul(scaled, []register.Id{agg, pow}, one))
		b.emit(instruction.NewIntAdd(out, []register.Id{out, scaled}, zero))

		if i+1 < key.width {
			npow := b.newComputed("pow")
			b.emit(instruction.NewIntMul(npow, []register.Id{pow}, two))
			pow = npow
		}
	}

	b.emit(instruction.NewReturn())

	name := helperName(id, key)

	return function.New(name, b.regs(), []VectorInstruction{VectorInstruction{Codes: b.code}})
}

func newDecomposedNotHelper[W word.Word[W]](id uint, key bitwiseHelperKey) machine.Module {
	b := newHelperBuilder[W](key.width, key.arity)

	out := b.output
	zero := word.Uint64[W](0)
	one := word.Uint64[W](1)
	two := word.Uint64[W](2)

	divisor := b.newComputed("d")
	b.emit(instruction.NewIntAdd(divisor, nil, two))

	oneReg := b.newComputed("one")
	b.emit(instruction.NewIntAdd(oneReg, nil, one))

	b.emit(instruction.NewIntAdd(out, nil, zero))

	pow := b.newComputed("pow")
	b.emit(instruction.NewIntAdd(pow, nil, one))

	work := b.newComputed("w")
	b.emit(instruction.NewIntAdd(work, []register.Id{b.inputs[0]}, zero))

	for i := uint(0); i < key.width; i++ {
		bit := b.newComputed("bit")
		b.emit(instruction.NewIntRem[W](bit, work, divisor))

		next := b.newComputed("next")
		b.emit(instruction.NewIntDiv[W](next, work, divisor))
		work = next

		nbit := b.newComputed("nbit")
		b.emit(instruction.NewIntSub(nbit, []register.Id{oneReg, bit}, zero))

		scaled := b.newComputed("scaled")
		b.emit(instruction.NewIntMul(scaled, []register.Id{nbit, pow}, one))
		b.emit(instruction.NewIntAdd(out, []register.Id{out, scaled}, zero))

		if i+1 < key.width {
			npow := b.newComputed("pow")
			b.emit(instruction.NewIntMul(npow, []register.Id{pow}, two))
			pow = npow
		}
	}

	b.emit(instruction.NewReturn())

	name := helperName(id, key)

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
		base = append(base, register.NewInput(fmt.Sprintf("arg%d", i), width, padding))
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
	var padding big.Int

	id := register.NewId(uint(len(p.base)))
	name := fmt.Sprintf("%s%d", prefix, p.nextTmp)
	p.base = append(p.base, register.NewComputed(name, p.width, padding))
	p.nextTmp++

	return id
}

func (p *helperBuilder[W]) combineBit(op instruction.OpCode, lhs, rhs register.Id) register.Id {
	zero := word.Uint64[W](0)
	one := word.Uint64[W](1)
	two := word.Uint64[W](2)

	switch op {
	case opcode.BIT_AND:
		res := p.newComputed("and")
		p.emit(instruction.NewIntMul(res, []register.Id{lhs, rhs}, one))

		return res
	case opcode.BIT_OR:
		sum := p.newComputed("or_sum")
		p.emit(instruction.NewIntAdd(sum, []register.Id{lhs, rhs}, zero))

		prod := p.newComputed("or_prod")
		p.emit(instruction.NewIntMul(prod, []register.Id{lhs, rhs}, one))

		res := p.newComputed("or")
		p.emit(instruction.NewIntSub(res, []register.Id{sum, prod}, zero))

		return res
	case opcode.BIT_XOR:
		sum := p.newComputed("xor_sum")
		p.emit(instruction.NewIntAdd(sum, []register.Id{lhs, rhs}, zero))

		prod := p.newComputed("xor_prod")
		p.emit(instruction.NewIntMul(prod, []register.Id{lhs, rhs}, one))

		dbl := p.newComputed("xor_dbl")
		p.emit(instruction.NewIntMul(dbl, []register.Id{prod}, two))

		res := p.newComputed("xor")
		p.emit(instruction.NewIntSub(res, []register.Id{sum, dbl}, zero))

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

func helperName(id uint, key bitwiseHelperKey) string {
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

	if key.constant != "" {
		return fmt.Sprintf("$bit_%s_u%d_n%d_c%s_h%d", op, key.width, key.arity, key.constant, id)
	}

	return fmt.Sprintf("$bit_%s_u%d_n%d_h%d", op, key.width, key.arity, id)
}
