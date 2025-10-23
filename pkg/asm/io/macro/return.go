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
package macro

import (
	"github.com/consensys/go-corset/pkg/asm/io"
	"github.com/consensys/go-corset/pkg/asm/io/micro"
	"github.com/consensys/go-corset/pkg/schema/register"
)

// Return signals a return from the enclosing function.
type Return struct {
	// dummy is included to force Return structs to be stored in the heap.
	//nolint
	Dummy uint
}

// Execute this instruction with the given local and global state.  The next
// program counter position is returned, or io.RETURN if the enclosing
// function has terminated (i.e. because a return instruction was
// encountered).
func (p *Return) Execute(state io.State) uint {
	return io.RETURN
}

// Lower this instruction into a exactly one more micro instruction.
func (p *Return) Lower(pc uint) micro.Instruction {
	// Lowering here produces an instruction containing a single microcode.
	return micro.NewInstruction(&micro.Ret{})
}

// RegistersRead returns the set of registers read by this instruction.
func (p *Return) RegistersRead() []io.RegisterId {
	return nil
}

// RegistersWritten returns the set of registers written by this instruction.
func (p *Return) RegistersWritten() []io.RegisterId {
	return nil
}

func (p *Return) String(fn register.Map) string {
	return "return"
}

// Validate checks whether or not this instruction is correctly balanced.
func (p *Return) Validate(fieldWidth uint, fn register.Map) error {
	return nil
}
