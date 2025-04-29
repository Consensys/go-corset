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

import "math/big"

// Jmp provides an unconditional branching instruction to a given instructon.
type Jmp struct {
	Target uint
}

// Bind any labels contained within this instruction using the given label map.
func (p *Jmp) Bind(labels []uint) {
	p.Target = labels[p.Target]
}

// Execute an unconditional branch instruction by returning the destination
// program counter.
func (p *Jmp) Execute(pc uint, state []big.Int, regs []Register) uint {
	return p.Target
}

// IsBalanced checks whether or not this instruction is correctly balanced.
func (p *Jmp) IsBalanced(regs []Register) error {
	return nil
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
