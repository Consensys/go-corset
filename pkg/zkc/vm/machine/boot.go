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
package machine

import (
	"errors"

	"github.com/consensys/go-corset/pkg/util/collection/stack"
	"github.com/consensys/go-corset/pkg/zkc/vm/instruction"
	"github.com/consensys/go-corset/pkg/zkc/vm/memory"
	"github.com/consensys/go-corset/pkg/zkc/vm/word"
)

// BootState is the concrete runtime state of a booted   It bundles
// the function table together with all memory banks (statics, inputs, outputs,
// RAMs) and the call stack, and is passed by value into each BootExecutor step.
type BootState[W word.Word[W]] = BaseState[W, instruction.Instruction[W], memory.Boot[W]]

// Boot is a fully assembled machine operating over arbitrary machine words.
type Boot[W word.Word[W]] = Base[W, instruction.Instruction[W], memory.Boot[W], BootExecutor[W, BootState[W]]]

// NewBoot constructs an empty boot machine.
func NewBoot[W word.Word[W]]() Boot[W] {
	var callstack stack.Stack[Frame[W]]
	//
	return Base[W, instruction.Instruction[W], memory.Boot[W], BootExecutor[W, BootState[W]]]{
		state: BootState[W]{callstack: &callstack},
	}
}

// BootExecutor for boot machine(s).
type BootExecutor[W word.Word[W], S State[W, instruction.Instruction[W]]] struct{}

// Execute implementation for the Executor interface
func (p BootExecutor[W, S]) Execute(state S) error {
	var (
		err       error
		callstack = state.CallStack()
		// Extract executing frame
		frame = callstack.Pop()
		// Identify enclosing function
		fn = state.Function(frame.Function())
		// Determine current PC position
		pc = frame.PC()
		// Lookup instruction to execute
		insn = fn.CodeAt(pc)
	)
	//
	switch insn := insn.(type) {
	case *instruction.Add[word.Uint]:
		panic("todo add")
	case *instruction.Jmp:
		frame.Goto(insn.Target)
	case *instruction.Fail:
		err = errors.New("machine panic")
	case *instruction.Return:
		return nil
	default:
		panic("unknown instruction encountered")
	}
	//
	callstack.Push(frame)
	//
	return err
}
