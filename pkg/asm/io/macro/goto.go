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
	"fmt"

	"github.com/consensys/go-corset/pkg/asm/io"
	"github.com/consensys/go-corset/pkg/asm/io/micro"
)

// Goto provides an unconditional branching instruction to a given instructon.
type Goto struct {
	Target uint
}

// Bind any labels contained within this instruction using the given label map.
func (p *Goto) Bind(labels []uint) {
	p.Target = labels[p.Target]
}

// Execute this instruction with the given local and global state.  The next
// program counter position is returned, or io.RETURN if the enclosing
// function has terminated (i.e. because a return instruction was
// encountered).
func (p *Goto) Execute(state io.State) uint {
	return p.Target
}

// Lower this instruction into a exactly one more micro instruction.
func (p *Goto) Lower(pc uint) micro.Instruction {
	// Lowering here produces an instruction containing a single microcode.
	return micro.NewInstruction(&micro.Jmp{Target: p.Target})
}

// RegistersRead returns the set of registers read by this instruction.
func (p *Goto) RegistersRead() []io.RegisterId {
	return nil
}

// RegistersWritten returns the set of registers written by this instruction.
func (p *Goto) RegistersWritten() []io.RegisterId {
	return nil
}

func (p *Goto) String(fn io.Function[Instruction]) string {
	return fmt.Sprintf("goto %d", p.Target)
}

// Validate checks whether or not this instruction is correctly balanced.
func (p *Goto) Validate(fieldWidth uint, fn io.Function[Instruction]) error {
	return nil
}
