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
package instruction

import (
	"fmt"

	"github.com/consensys/go-corset/pkg/schema/register"
	"github.com/consensys/go-corset/pkg/util/field"
)

// ============================================================================
// JMP
// ============================================================================

// Jmp performs an unconditional branch to the instruction identified by the
// immediate operand.  The immediate is interpreted as the target program
// counter within the enclosing function, so executing a Jmp simply transfers
// control to that PC.  Jmp is one of the three control-flow terminators (along
// with Return and Fail) recognised by the vectoriser as ending a basic block.
type Jmp[W any] struct{ OpImm[W] }

// NewJmp constructs a fresh unconditional jump instruction to the given PC
// location.
func NewJmp[W any](target uint) *Jmp[W] {
	return &Jmp[W]{OpImm[W]{JUMP, target}}
}

// ============================================================================
// FAIL
// ============================================================================

// Fail aborts execution of the machine, signalling an unrecoverable error
// (a "machine panic").  No further instructions are executed and the
// surrounding call stack is not unwound; the executor returns an error to its
// caller.  The immediate operand is unused.  Fail is one of the three
// control-flow terminators (along with Jmp and Return) recognised by the
// vectoriser as ending a basic block.
type Fail[W any] struct{ OpImm[W] }

// NewFail constructs a fresh fail instruction.
func NewFail[W any]() *Fail[W] {
	return &Fail[W]{OpImm[W]{FAIL, 0}}
}

// ============================================================================
// RETURN
// ============================================================================

// Return leaves the current stack frame, copying the function's return
// registers into the caller's frame and resuming execution at the instruction
// following the originating Call.  When executed at the outermost (boot)
// frame, Return halts the machine.  The immediate operand is unused.  Return
// is one of the three control-flow terminators (along with Jmp and Fail)
// recognised by the vectoriser as ending a basic block.
type Return[W any] struct{ OpImm[W] }

// NewReturn constructs a fresh return instruction.
func NewReturn[W any]() *Return[W] {
	return &Return[W]{OpImm[W]{RETURN, 0}}
}

// ============================================================================
// Opcode-Immediate (OpImm) instruction type
// ============================================================================

// OpImm represents an instruction parameterised solely by an opcode and a uint
// immediate value.  This is used for control-flow instructions such as JUMP
// (where the immediate identifies the branch target), as well as RETURN and
// FAIL (which ignore the immediate).
type OpImm[W any] struct {
	Op        OpCode
	Immediate uint
}

// OpCode implementation for Instruction interface
func (p *OpImm[W]) OpCode() OpCode {
	return p.Op
}

// Uses implementation for Instruction interface.
func (p *OpImm[W]) Uses() []register.Id {
	return nil
}

// Definitions implementation for Instruction interface.
func (p *OpImm[W]) Definitions() []register.Id {
	return nil
}

func (p *OpImm[W]) String(_ SystemMap[W]) string {
	switch p.Op {
	case JUMP:
		return fmt.Sprintf("jmp %d", p.Immediate)
	case RETURN:
		return "ret"
	case FAIL:
		return "fail"
	default:
		panic("unknown OpImm32 opcode")
	}
}

// MicroValidate implementation for MicroInstruction interface.
func (p *OpImm[W]) MicroValidate(_ uint, _ field.Config, _ SystemMap[W]) []error {
	return nil
}
