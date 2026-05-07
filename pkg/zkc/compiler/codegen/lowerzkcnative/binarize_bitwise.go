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
	"fmt"
	"math/big"

	"github.com/consensys/go-corset/pkg/schema/register"
	"github.com/consensys/go-corset/pkg/zkc/vm/function"
	"github.com/consensys/go-corset/pkg/zkc/vm/instruction"
	"github.com/consensys/go-corset/pkg/zkc/vm/instruction/opcode"
	"github.com/consensys/go-corset/pkg/zkc/vm/machine"
	"github.com/consensys/go-corset/pkg/zkc/vm/word"
)

// BinarizeBitwise splits any AND/OR/XOR instruction with more than two source
// registers into a left-fold chain of binary instructions.  This must run
// before LowerBitwise so that the helper modules it generates never need more
// than two inputs.
func BinarizeBitwise[W word.Word[W]](modules []machine.Module) []machine.Module {
	out := append([]machine.Module{}, modules...)

	for i, mod := range out {
		if fn, ok := mod.(*function.Boot); ok {
			out[i] = binarizeBitwiseFunction[W](fn)
		}
	}

	return out
}

func binarizeBitwiseFunction[W word.Word[W]](fn *function.Boot) *function.Boot {
	var (
		code      = fn.Code()
		ncode     = make([]vectorInstruction, len(code))
		registers = append([]register.Register{}, fn.Registers()...)
	)

	for i, insn := range code {
		ncodes := binarizeBitwiseCodes[W](insn.Codes, &registers)
		ncode[i] = vectorInstruction{Codes: ncodes}
	}

	return function.New(fn.Name(), registers, ncode)
}

func binarizeBitwiseCodes[W word.Word[W]](
	codes []instruction.Instruction,
	registers *[]register.Register,
) []instruction.Instruction {
	ncodes := make([]instruction.Instruction, 0, len(codes))

	for _, code := range codes {
		ncodes = append(ncodes, binarizeBitwiseCode[W](code, registers)...)
	}

	return ncodes
}

func binarizeBitwiseCode[W word.Word[W]](
	code instruction.Instruction,
	registers *[]register.Register,
) []instruction.Instruction {
	var (
		op       instruction.OpCode
		target   register.Id
		sources  []register.Id
		constant W
	)

	switch t := code.(type) {
	case *instruction.BitAnd[W]:
		op, target, sources, constant = t.OpCode(), t.Target, t.Sources, t.Constant
	case *instruction.BitOr[W]:
		op, target, sources, constant = t.OpCode(), t.Target, t.Sources, t.Constant
	case *instruction.BitXor[W]:
		op, target, sources, constant = t.OpCode(), t.Target, t.Sources, t.Constant
	default:
		return []instruction.Instruction{code}
	}

	if len(sources) <= 2 {
		return []instruction.Instruction{code}
	}

	width := (*registers)[target.Unwrap()].Width()
	identity := bitwiseIdentity[W](op, width)

	insns := make([]instruction.Instruction, 0, len(sources)-1)
	acc := sources[0]

	for _, src := range sources[1 : len(sources)-1] {
		tmp := allocTmp(registers, width)
		insns = append(insns, newBinaryBitOp[W](op, tmp, acc, src, identity))
		acc = tmp
	}

	insns = append(insns, newBinaryBitOp[W](op, target, acc, sources[len(sources)-1], constant))

	return insns
}

// bitwiseIdentity returns the identity element for the given bitwise operation.
func bitwiseIdentity[W word.Word[W]](op instruction.OpCode, width uint) W {
	var z W

	if op == opcode.BIT_AND {
		maskBig := new(big.Int).Sub(new(big.Int).Lsh(big.NewInt(1), width), big.NewInt(1))
		return z.SetBigInt(maskBig)
	}

	return z
}

func newBinaryBitOp[W word.Word[W]](
	op instruction.OpCode, target, lhs, rhs register.Id, constant W,
) instruction.Instruction {
	sources := []register.Id{lhs, rhs}

	switch op {
	case opcode.BIT_AND:
		return instruction.NewBitAnd[W](target, sources, constant)
	case opcode.BIT_OR:
		return instruction.NewBitOr[W](target, sources, constant)
	case opcode.BIT_XOR:
		return instruction.NewBitXor[W](target, sources, constant)
	default:
		panic(fmt.Sprintf("unexpected bitwise opcode: %d", op))
	}
}
