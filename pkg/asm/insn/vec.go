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
package insn

import (
	"math/big"

	"github.com/consensys/go-corset/pkg/util/collection/bit"
)

// Vec is an instruction which contains one or more instructions
// that are to be executed "in parallel", roughly following the ideas of vector
// machines and vectorisation.
type Vec struct {
	Code []Instruction
}

// Bind any labels contained within this instruction using the given label map.
func (p *Vec) Bind(labels []uint) {
	for _, insn := range p.Code {
		insn.Bind(labels)
	}
}

// Execute a given instruction at a given program counter position, using a
// given set of register values.  This may update the register values, and
// returns the next program counter position.  If the program counter is
// math.MaxUint then a return is signaled.
func (p *Vec) Execute(pc uint, state []big.Int, regs []Register) uint {
	var npc uint = pc + 1

	for _, r := range p.Code {
		p := r.Execute(pc, state, regs)
		// Sanity check
		if p != pc+1 && npc != pc+1 {
			panic("conflicting jump targets")
		}
		//
		npc = p
	}
	//
	return npc
}

// IsWellFormed checks whether or not this instruction is correctly balanced.
func (p *Vec) IsWellFormed(regs []Register) error {
	for _, r := range p.Code {
		if err := r.IsWellFormed(regs); err != nil {
			return err
		}
	}
	// TODO: sanity check expected requirements for vector instruction.
	//
	return nil
}

// Registers returns the set of registers read/written by this instruction.
func (p *Vec) Registers() []uint {
	var set bit.Set
	//
	for _, insn := range p.Code {
		set.InsertAll(insn.Registers()...)
	}
	//
	return set.Iter().Collect()
}

// RegistersRead returns the set of registers read by this instruction.
func (p *Vec) RegistersRead() []uint {
	var set bit.Set
	//
	for _, insn := range p.Code {
		set.InsertAll(insn.RegistersRead()...)
	}
	//
	return set.Iter().Collect()
}

// RegistersWritten returns the set of registers written by this instruction.
func (p *Vec) RegistersWritten() []uint {
	var set bit.Set
	//
	for _, insn := range p.Code {
		set.InsertAll(insn.RegistersWritten()...)
	}
	//
	return set.Iter().Collect()
}

// Translate this instruction into low-level constraints.
func (p *Vec) Translate(pc uint, st StateTranslator) {
	for _, insn := range p.Code {
		insn.Translate(pc, st)
	}
}
