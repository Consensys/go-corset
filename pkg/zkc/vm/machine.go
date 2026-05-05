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
package vm

import (
	"github.com/consensys/go-corset/pkg/schema/register"
	"github.com/consensys/go-corset/pkg/util/field"
	"github.com/consensys/go-corset/pkg/zkc/vm/instruction"
	"github.com/consensys/go-corset/pkg/zkc/vm/internal/function"
	"github.com/consensys/go-corset/pkg/zkc/vm/internal/machine"
)

// Machine represents the state of an executing "machine", including all
// registers, memories and functions.  A machine may be executing or terminated.
// Machines are abstracted over a given type of word W, and instruction I.  For
// example, a machine could be operating over 16bit words or 8bit words, or over
// field elements directly.  Furthermore, a machine may be operating over
// instructions compiled into bytes (for efficient execution), or instructions
// represented at a higher level (e.g. for analysis or compilation).
type Machine[W any] = machine.Core[W]

// Executor captures the notion of an instruction-specific executor.  That is,
// an executor designed for executing certain instructions over a given type of
// machine word (e.g. a Word or a field.Element, etc).  A key aspect of the
// executor is that its really only intended for straight-line instructions, and
// other control-flow instructions (e.g. skipping, calling, etc) are handled by
// the base machine (since they are common to all machines).
type Executor[W any, I any] = machine.Executor[W, I]

// StackFrame captures the state of an executing function on the call stack.
// Specifically, it contains the state of all registers at the current point of
// execution.
type StackFrame[W any] = machine.Frame[W]

// ProgramCounter abstracts the notion of a program counter in a machine.  A key
// aspect is that it two dimensional to account for so-called "vector"
// instructions: (1) it identifies the (macro) instruction being executed; (2)
// it identifies the (micro) instruction within that being executed.
type ProgramCounter = machine.ProgramCounter

// BaseWord captures the minimal set of requirements for a word used in the base
// machine.
type BaseWord[W any] = machine.BaseWord[W]

// BaseMachine provides a fundamental machine implementation.  The intention is
// that other machine variations build off this by providing executors specific
// to their instruction set.
type BaseMachine[W BaseWord[W], I Instruction, E Executor[W, I]] = machine.Base[W, I, E]

// ============================================================================
// Word Machine
// ============================================================================

// WordMachine is a machine which operates over standard machine words.
type WordMachine[W Word[W]] = machine.Word[W]

// WordFunction is a function made up of word instructions.
type WordFunction = Function[instruction.Word]

// WordInstruction is an instruction which operates over standard machine words.
type WordInstruction = instruction.Word

// ============================================================================
// Field Machine
// ============================================================================

// FieldFunction is a function made up of field instructions.
type FieldFunction = Function[instruction.Field]

// FieldMachine is a machine which operates over field elements only.
type FieldMachine[F field.Element[F]] = machine.Field[F]

// FieldInstruction is an instruction which operates over field elements only.
type FieldInstruction = instruction.Field

// ============================================================================
// Constructors
// ============================================================================

// NewFunction constructs a new function with the given components.
func NewFunction[I instruction.Instruction](name string, registers []register.Register,
	code []instruction.Vector[I]) *Function[I] {
	return function.New(name, registers, code)
}

// NewWordMachine constructs a new word machine with a given set of modules and
// a given field configuration (for native field instructions).
func NewWordMachine[W Word[W]](field field.Config, modules ...Module) *WordMachine[W] {
	return machine.NewWord[W](field, modules...)
}

// ============================================================================
// Utils
// ============================================================================

// ExecuteAll executes a given machine to completion in chunks of n steps,
// returning the number of steps executed and/or any error arising.
func ExecuteAll[W any, M Machine[W]](machine M, n uint) (uint, error) {
	var nsteps uint
	//
	for {
		// Execute upto n steps
		m, err := machine.Execute(n)
		// update the tally
		nsteps += m
		// check for termination
		if err != nil || m < n {
			return nsteps, err
		}
	}
}
