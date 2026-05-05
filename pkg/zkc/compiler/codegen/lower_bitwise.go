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
	"github.com/consensys/go-corset/pkg/zkc/vm/machine"
	"github.com/consensys/go-corset/pkg/zkc/vm/word"
)

// LowerBitwise rewrites VM-level bitwise micro-instructions into CALLs to
// helper functions. The helper modules are appended to the returned module
// slice.
func LowerBitwise[W word.Word[W]](modules []machine.Module[W], cfg field.Config) []machine.Module[W] {
	var (
		out     = append([]machine.Module[W]{}, modules...)
		helpers = newBitwiseHelpers[W](uint(len(out)), cfg)
	)

	for i, mod := range out {
		if fn, ok := mod.(*function.Boot[W]); ok {
			out[i] = lowerBitwiseFunction(fn, helpers)
		}
	}

	return append(out, helpers.modules()...)
}

func lowerBitwiseFunction[W word.Word[W]](
	fn *function.Boot[W], helpers *bitwiseHelpers[W],
) *function.Boot[W] {
	var (
		code      = fn.Code()
		ncode     = make([]instruction.Instruction[W], len(code))
		registers = append([]register.Register{}, fn.Registers()...)
	)

	for i, insn := range code {
		switch t := insn.(type) {
		case instruction.MicroInstruction[W]:
			ncodes := lowerBitwiseCodes([]instruction.MicroInstruction[W]{t}, registers, helpers)
			if len(ncodes) == 1 {
				ncode[i] = ncodes[0]
			} else {
				ncode[i] = &instruction.Vector[W]{Codes: ncodes}
			}
			// Note: shouldn't happen since the vectorization happens later
		case *instruction.Vector[W]:
			ncodes := lowerBitwiseCodes(t.Codes, registers, helpers)
			ncode[i] = &instruction.Vector[W]{Codes: ncodes}
		default:
			ncode[i] = insn
		}
	}

	return function.New(fn.Name(), registers, ncode)
}

func lowerBitwiseCodes[W word.Word[W]](
	codes []instruction.MicroInstruction[W],
	registers []register.Register,
	helpers *bitwiseHelpers[W],
) []instruction.MicroInstruction[W] {
	ncodes := make([]instruction.MicroInstruction[W], 0, len(codes))

	for _, code := range codes {
		ncodes = append(ncodes, lowerBitwiseCode(code, registers, helpers)...)
	}

	return ncodes
}

func lowerBitwiseCode[W word.Word[W]](
	code instruction.MicroInstruction[W],
	registers []register.Register,
	helpers *bitwiseHelpers[W],
) []instruction.MicroInstruction[W] {
	switch t := code.(type) {
	case *instruction.BitAnd[W]:
		width, ok := lowerableWidth(registers, t.Target, helpers.field.BandWidth)
		if !ok {
			return []instruction.MicroInstruction[W]{code}
		}

		id := helpers.ensure(t.OpCode(), width, len(t.Sources), t.Constant)

		return []instruction.MicroInstruction[W]{
			instruction.NewCall[W](id, append([]register.Id{}, t.Sources...), []register.Id{t.Target}),
		}
	case *instruction.BitOr[W]:
		width, ok := lowerableWidth(registers, t.Target, helpers.field.BandWidth)
		if !ok {
			return []instruction.MicroInstruction[W]{code}
		}

		id := helpers.ensure(t.OpCode(), width, len(t.Sources), t.Constant)

		return []instruction.MicroInstruction[W]{
			instruction.NewCall[W](id, append([]register.Id{}, t.Sources...), []register.Id{t.Target}),
		}
	case *instruction.BitXor[W]:
		width, ok := lowerableWidth(registers, t.Target, helpers.field.BandWidth)
		if !ok {
			return []instruction.MicroInstruction[W]{code}
		}

		id := helpers.ensure(t.OpCode(), width, len(t.Sources), t.Constant)

		return []instruction.MicroInstruction[W]{
			instruction.NewCall[W](id, append([]register.Id{}, t.Sources...), []register.Id{t.Target}),
		}
	case *instruction.BitNot[W]:
		width, ok := lowerableWidth(registers, t.Target, helpers.field.BandWidth)
		if !ok {
			return []instruction.MicroInstruction[W]{code}
		}

		id := helpers.ensure(t.OpCode(), width, len(t.Sources), zeroWord[W]())

		return []instruction.MicroInstruction[W]{
			instruction.NewCall[W](id, append([]register.Id{}, t.Sources...), []register.Id{t.Target}),
		}
	case *instruction.BitShl[W]:
		width, ok := lowerableWidth(registers, t.Target, helpers.field.BandWidth)
		if !ok {
			return []instruction.MicroInstruction[W]{code}
		}

		id := helpers.ensure(t.OpCode(), width, len(t.Sources), zeroWord[W]())

		return []instruction.MicroInstruction[W]{
			instruction.NewCall[W](id, append([]register.Id{}, t.Sources...), []register.Id{t.Target}),
		}
	case *instruction.BitShr[W]:
		width, ok := lowerableWidth(registers, t.Target, helpers.field.BandWidth)
		if !ok {
			return []instruction.MicroInstruction[W]{code}
		}

		id := helpers.ensure(t.OpCode(), width, len(t.Sources), zeroWord[W]())

		return []instruction.MicroInstruction[W]{
			instruction.NewCall[W](id, append([]register.Id{}, t.Sources...), []register.Id{t.Target}),
		}
	default:
		return []instruction.MicroInstruction[W]{code}
	}
}

func lowerableWidth(registers []register.Register, target register.Id, bandWidth uint) (uint, bool) {
	reg := registers[target.Unwrap()]
	if reg.IsNative() {
		return bandWidth, true
	}

	return reg.Width(), true
}

func zeroWord[W word.Word[W]]() W {
	var z W
	return z
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
	items  []machine.Module[W]
}

func newBitwiseHelpers[W word.Word[W]](baseID uint, cfg field.Config) *bitwiseHelpers[W] {
	return &bitwiseHelpers[W]{
		baseID: baseID,
		field:  cfg,
		ids:    make(map[bitwiseHelperKey]uint),
	}
}

func (p *bitwiseHelpers[W]) modules() []machine.Module[W] {
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
	mod := newBitwiseHelperModule[W](id, key, constant)

	p.items = append(p.items, mod)
	p.ids[key] = id

	return id
}

func helperConstant[W word.Word[W]](op instruction.OpCode, constant W) string {
	switch op {
	case instruction.BIT_AND, instruction.BIT_OR, instruction.BIT_XOR:
		return constant.BigInt().Text(16)
	default:
		return ""
	}
}

func newBitwiseHelperModule[W word.Word[W]](
	id uint,
	key bitwiseHelperKey,
	constant W,
) machine.Module[W] {
	if key.opcode == instruction.BIT_AND || key.opcode == instruction.BIT_OR ||
		key.opcode == instruction.BIT_XOR {
		return newDecomposedNaryHelper[W](id, key, constant)
	}

	if key.opcode == instruction.BIT_NOT {
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

	var op instruction.MicroInstruction[W]

	switch key.opcode {
	case instruction.BIT_AND:
		op = instruction.NewBitAnd(target, sources, constant)
	case instruction.BIT_OR:
		op = instruction.NewBitOr(target, sources, constant)
	case instruction.BIT_XOR:
		op = instruction.NewBitXor(target, sources, constant)
	case instruction.BIT_NOT:
		op = instruction.NewBitNot[W](target, sources[0])
	case instruction.BIT_SHL:
		op = instruction.NewBitShl[W](target, sources[0], sources[1])
	case instruction.BIT_SHR:
		op = instruction.NewBitShr[W](target, sources[0], sources[1])
	default:
		panic(fmt.Sprintf("unsupported bitwise helper opcode: %d", key.opcode))
	}

	code := []instruction.Instruction[W]{op, instruction.NewReturn[W]()}
	name := helperName(id, key)

	return function.New(name, regs, code)
}

func newDecomposedNaryHelper[W word.Word[W]](
	id uint,
	key bitwiseHelperKey,
	constant W,
) machine.Module[W] {
	b := newHelperBuilder[W](key.width, key.arity)

	out := b.output
	zero := word.Uint64[W](0)
	one := word.Uint64[W](1)
	two := word.Uint64[W](2)

	divisor := b.newComputed("d")
	b.emit(instruction.NewIntAdd[W](divisor, nil, two))
	b.emit(instruction.NewIntAdd[W](out, nil, zero))

	pow := b.newComputed("pow")
	b.emit(instruction.NewIntAdd[W](pow, nil, one))

	work := make([]register.Id, len(b.inputs))
	for i, arg := range b.inputs {
		w := b.newComputed("w")
		b.emit(instruction.NewIntAdd[W](w, []register.Id{arg}, zero))
		work[i] = w
	}

	for i := uint(0); i < key.width; i++ {
		agg := b.newComputed("agg")
		if constant.BigInt().Bit(int(i)) == 0 {
			b.emit(instruction.NewIntAdd[W](agg, nil, zero))
		} else {
			b.emit(instruction.NewIntAdd[W](agg, nil, one))
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
		b.emit(instruction.NewIntMul[W](scaled, []register.Id{agg, pow}, one))
		b.emit(instruction.NewIntAdd[W](out, []register.Id{out, scaled}, zero))

		if i+1 < key.width {
			npow := b.newComputed("pow")
			b.emit(instruction.NewIntMul[W](npow, []register.Id{pow}, two))
			pow = npow
		}
	}

	b.emit(instruction.NewReturn[W]())

	name := helperName(id, key)

	return function.New(name, b.regs(), b.code)
}

func newDecomposedNotHelper[W word.Word[W]](id uint, key bitwiseHelperKey) machine.Module[W] {
	b := newHelperBuilder[W](key.width, key.arity)

	out := b.output
	zero := word.Uint64[W](0)
	one := word.Uint64[W](1)
	two := word.Uint64[W](2)

	divisor := b.newComputed("d")
	b.emit(instruction.NewIntAdd[W](divisor, nil, two))

	oneReg := b.newComputed("one")
	b.emit(instruction.NewIntAdd[W](oneReg, nil, one))

	b.emit(instruction.NewIntAdd[W](out, nil, zero))

	pow := b.newComputed("pow")
	b.emit(instruction.NewIntAdd[W](pow, nil, one))

	work := b.newComputed("w")
	b.emit(instruction.NewIntAdd[W](work, []register.Id{b.inputs[0]}, zero))

	for i := uint(0); i < key.width; i++ {
		bit := b.newComputed("bit")
		b.emit(instruction.NewIntRem[W](bit, work, divisor))

		next := b.newComputed("next")
		b.emit(instruction.NewIntDiv[W](next, work, divisor))
		work = next

		nbit := b.newComputed("nbit")
		b.emit(instruction.NewIntSub[W](nbit, []register.Id{oneReg, bit}, zero))

		scaled := b.newComputed("scaled")
		b.emit(instruction.NewIntMul[W](scaled, []register.Id{nbit, pow}, one))
		b.emit(instruction.NewIntAdd[W](out, []register.Id{out, scaled}, zero))

		if i+1 < key.width {
			npow := b.newComputed("pow")
			b.emit(instruction.NewIntMul[W](npow, []register.Id{pow}, two))
			pow = npow
		}
	}

	b.emit(instruction.NewReturn[W]())

	name := helperName(id, key)

	return function.New(name, b.regs(), b.code)
}

type helperBuilder[W word.Word[W]] struct {
	width   uint
	inputs  []register.Id
	output  register.Id
	base    []register.Register
	code    []instruction.Instruction[W]
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

func (p *helperBuilder[W]) emit(insn instruction.Instruction[W]) {
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
	case instruction.BIT_AND:
		res := p.newComputed("and")
		p.emit(instruction.NewIntMul[W](res, []register.Id{lhs, rhs}, one))

		return res
	case instruction.BIT_OR:
		sum := p.newComputed("or_sum")
		p.emit(instruction.NewIntAdd[W](sum, []register.Id{lhs, rhs}, zero))

		prod := p.newComputed("or_prod")
		p.emit(instruction.NewIntMul[W](prod, []register.Id{lhs, rhs}, one))

		res := p.newComputed("or")
		p.emit(instruction.NewIntSub[W](res, []register.Id{sum, prod}, zero))

		return res
	case instruction.BIT_XOR:
		sum := p.newComputed("xor_sum")
		p.emit(instruction.NewIntAdd[W](sum, []register.Id{lhs, rhs}, zero))

		prod := p.newComputed("xor_prod")
		p.emit(instruction.NewIntMul[W](prod, []register.Id{lhs, rhs}, one))

		dbl := p.newComputed("xor_dbl")
		p.emit(instruction.NewIntMul[W](dbl, []register.Id{prod}, two))

		res := p.newComputed("xor")
		p.emit(instruction.NewIntSub[W](res, []register.Id{sum, dbl}, zero))

		return res
	default:
		panic(fmt.Sprintf("unsupported bit combine opcode: %d", op))
	}
}

// HasBitwiseOps checks whether any module contains VM bitwise instructions.
func HasBitwiseOps[W word.Word[W]](modules []machine.Module[W]) bool {
	for _, mod := range modules {
		fn, ok := mod.(*function.Boot[W])
		if !ok {
			continue
		}

		for _, insn := range fn.Code() {
			switch t := insn.(type) {
			case *instruction.Vector[W]:
				for _, code := range t.Codes {
					if isBitwiseOpcode(code.OpCode()) {
						return true
					}
				}
			case instruction.MicroInstruction[W]:
				if isBitwiseOpcode(t.OpCode()) {
					return true
				}
			}
		}
	}

	return false
}

func isBitwiseOpcode(op instruction.OpCode) bool {
	switch op {
	case instruction.BIT_AND, instruction.BIT_OR, instruction.BIT_XOR,
		instruction.BIT_NOT, instruction.BIT_SHL, instruction.BIT_SHR:
		return true
	default:
		return false
	}
}

func helperName(id uint, key bitwiseHelperKey) string {
	var op string

	switch key.opcode {
	case instruction.BIT_AND:
		op = "and"
	case instruction.BIT_OR:
		op = "or"
	case instruction.BIT_XOR:
		op = "xor"
	case instruction.BIT_NOT:
		op = "not"
	case instruction.BIT_SHL:
		op = "shl"
	case instruction.BIT_SHR:
		op = "shr"
	default:
		op = "unknown"
	}

	if key.constant != "" {
		return fmt.Sprintf("$bit_%s_u%d_n%d_c%s_h%d", op, key.width, key.arity, key.constant, id)
	}

	return fmt.Sprintf("$bit_%s_u%d_n%d_h%d", op, key.width, key.arity, id)
}
