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
package micro

import (
	"fmt"
	"math/big"
)

// Jmp provides an unconditional branching instruction to a given instructon.
type Jmp struct {
	Target uint
}

// Bind any labels contained within this instruction using the given label map.
func (p *Jmp) Bind(labels []uint) {
	p.Target = labels[p.Target]
}

// Clone this micro code.
func (p *Jmp) Clone() Code {
	return &Jmp{p.Target}
}

// Sequential indicates whether or not this microinstruction can execute
// sequentially onto the next.
func (p *Jmp) Sequential() bool {
	return false
}

// Terminal indicates whether or not this microinstruction terminates the
// enclosing function.
func (p *Jmp) Terminal() bool {
	return false
}

// Execute an unconditional branch instruction by returning the destination
// program counter.
func (p *Jmp) Execute(state []big.Int, regs []Register) uint {
	return p.Target
}

// Lower this instruction into a exactly one more micro instruction.
func (p *Jmp) Lower(pc uint) Instruction {
	// Lowering here produces an instruction containing a single microcode.
	return Instruction{[]Code{p}}
}

// Registers returns the set of registers read/written by this instruction.
func (p *Jmp) Registers() []uint {
	return nil
}

// RegistersRead returns the set of registers read by this instruction.
func (p *Jmp) RegistersRead() []uint {
	return nil
}

// RegistersWritten returns the set of registers written by this instruction.
func (p *Jmp) RegistersWritten() []uint {
	return nil
}

// Split this micro code using registers of arbirary width into one or more
// micro codes using registers of a fixed maximum width.
func (p *Jmp) Split(env *RegisterSplittingEnvironment) []Code {
	return []Code{p}
}

func (p *Jmp) String(regs []Register) string {
	return fmt.Sprintf("jmp %d", p.Target)
}

// Validate checks whether or not this instruction is correctly balanced.
func (p *Jmp) Validate(fieldWidth uint, regs []Register) error {
	return nil
}

/*
// Translate this instruction into low-level constraints.
func (p *Jmp) Translate(st *StateTranslator) {
	st.Jump(p.Target)
}
*/
