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
	"math/big"
	"testing"

	"github.com/consensys/go-corset/pkg/schema/register"
	"github.com/consensys/go-corset/pkg/zkc/vm/function"
	"github.com/consensys/go-corset/pkg/zkc/vm/instruction"
	"github.com/consensys/go-corset/pkg/zkc/vm/machine"
	"github.com/consensys/go-corset/pkg/zkc/vm/word"
)

func Test_LowerBitwise_RewritesBitAndToCall(t *testing.T) {
	var (
		padding big.Int
		regs    = []register.Register{
			register.NewInput("x", 8, padding),
			register.NewOutput("y", 8, padding),
		}
		code = []instruction.Instruction[word.Uint]{
			instruction.NewVector[word.Uint](
				instruction.NewBitAnd[word.Uint](
					register.NewId(1),
					[]register.Id{register.NewId(0)},
					word.Uint64[word.Uint](0xff),
				),
			),
			instruction.NewReturn[word.Uint](),
		}
		mainFn = function.New("main", regs, code)
	)

	lowered := LowerBitwise[word.Uint]([]machine.Module[word.Uint]{mainFn})
	if len(lowered) != 2 {
		t.Fatalf("expected 2 modules after lowering, got %d", len(lowered))
	}

	fn := lowered[0].(*function.Boot[word.Uint])

	vec := fn.CodeAt(0).(*instruction.Vector[word.Uint])

	if len(vec.Codes) != 1 {
		t.Fatalf("expected single micro instruction in vector, got %d", len(vec.Codes))
	}

	call, ok := vec.Codes[0].(*instruction.Call[word.Uint])
	if !ok {
		t.Fatalf("expected lowered instruction to be CALL, got %T", vec.Codes[0])
	}

	if call.Id != 1 {
		t.Fatalf("expected helper id 1, got %d", call.Id)
	}

	if len(call.Arguments) != 1 || call.Arguments[0] != register.NewId(0) {
		t.Fatalf("unexpected call arguments: %+v", call.Arguments)
	}

	if len(call.Returns) != 1 || call.Returns[0] != register.NewId(1) {
		t.Fatalf("unexpected call returns: %+v", call.Returns)
	}

	helper := lowered[1].(*function.Boot[word.Uint])
	if helperHasBitwise(helper) {
		t.Fatalf("expected decomposed helper to avoid bitwise opcodes")
	}
}

func Test_LowerBitwise_DeduplicatesHelpers(t *testing.T) {
	var (
		padding big.Int
		regs    = []register.Register{
			register.NewOutput("y", 8, padding),
		}
		code = []instruction.Instruction[word.Uint]{
			instruction.NewVector[word.Uint](
				instruction.NewBitOr[word.Uint](
					register.NewId(0),
					nil,
					word.Uint64[word.Uint](7),
				),
			),
			instruction.NewReturn[word.Uint](),
		}
	)

	fn1 := function.New("main", regs, code)
	fn2 := function.New("other", regs, code)

	lowered := LowerBitwise[word.Uint]([]machine.Module[word.Uint]{fn1, fn2})
	if len(lowered) != 3 {
		t.Fatalf("expected 3 modules after lowering, got %d", len(lowered))
	}

	c1 := firstCall(lowered[0].(*function.Boot[word.Uint]))
	c2 := firstCall(lowered[1].(*function.Boot[word.Uint]))

	if c1.Id != c2.Id {
		t.Fatalf("expected calls to share helper id, got %d and %d", c1.Id, c2.Id)
	}

	if c1.Id != 2 {
		t.Fatalf("expected helper id 2, got %d", c1.Id)
	}
}

func Test_LowerBitwise_LeavesNonBitwiseUnchanged(t *testing.T) {
	var (
		padding big.Int
		regs    = []register.Register{
			register.NewOutput("y", 8, padding),
		}
		code = []instruction.Instruction[word.Uint]{
			instruction.NewVector[word.Uint](
				instruction.NewIntAdd[word.Uint](
					register.NewId(0),
					nil,
					word.Uint64[word.Uint](5),
				),
			),
			instruction.NewReturn[word.Uint](),
		}
		mainFn = function.New("main", regs, code)
	)

	lowered := LowerBitwise[word.Uint]([]machine.Module[word.Uint]{mainFn})

	if len(lowered) != 1 {
		t.Fatalf("expected no helper modules for non-bitwise function, got %d modules", len(lowered))
	}

	fn := lowered[0].(*function.Boot[word.Uint])

	vec := fn.CodeAt(0).(*instruction.Vector[word.Uint])

	if _, ok := vec.Codes[0].(*instruction.IntAdd[word.Uint]); !ok {
		t.Fatalf("expected IntAdd to remain unchanged, got %T", vec.Codes[0])
	}
}

func Test_HasBitwiseOps(t *testing.T) {
	var (
		padding big.Int
		regs    = []register.Register{
			register.NewInput("x", 8, padding),
			register.NewOutput("y", 8, padding),
		}
		code = []instruction.Instruction[word.Uint]{
			instruction.NewVector[word.Uint](
				instruction.NewBitNot[word.Uint](register.NewId(1), register.NewId(0)),
			),
			instruction.NewReturn[word.Uint](),
		}
	)

	fn := function.New("main", regs, code)

	mods := []machine.Module[word.Uint]{fn}
	if !HasBitwiseOps(mods) {
		t.Fatalf("expected HasBitwiseOps to detect bitwise opcode")
	}

	lowered := LowerBitwise[word.Uint](mods)
	if HasBitwiseOps(lowered) {
		t.Fatalf("expected lowered machine to be bitwise-opcode free for this program")
	}
}

func firstCall(fn *function.Boot[word.Uint]) *instruction.Call[word.Uint] {
	vec := fn.CodeAt(0).(*instruction.Vector[word.Uint])

	return vec.Codes[0].(*instruction.Call[word.Uint])
}

func helperHasBitwise(fn *function.Boot[word.Uint]) bool {
	for _, insn := range fn.Code() {
		switch t := insn.(type) {
		case *instruction.Vector[word.Uint]:
			for _, code := range t.Codes {
				if isBitwiseOpcode(code.OpCode()) {
					return true
				}
			}
		case instruction.MicroInstruction[word.Uint]:
			if isBitwiseOpcode(t.OpCode()) {
				return true
			}
		}
	}

	return false
}
