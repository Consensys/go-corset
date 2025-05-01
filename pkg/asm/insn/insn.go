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
	"fmt"
	"math"
	"math/big"

	"github.com/consensys/go-corset/pkg/util/collection/bit"
)

// MicroInstruction provides an abstract notion of an atomic "machine
// instruction".  In practice, we want to pack as many microinstructions
// together as we can into a single vector instruction.
type MicroInstruction interface {
	// Bind any labels contained within this instruction using the given label map.
	Bind(labels []uint)
	// Execute a given instruction at a given program counter position, using a
	// given set of register values.  This may update the register values, and
	// returns the next program counter position.  If the program counter is
	// math.MaxUint then a return is signaled.
	Execute(pc uint, state []big.Int, regs []Register) uint
	// Sequential indicates whether or not this microinstruction can execute
	// sequentially onto the next.
	Sequential() bool
	// Terminal indicates whether or not this microinstruction terminates the
	// enclosing function.
	Terminal() bool
	// Validate that this micro-instruction is well-formed.  For example, that
	// it is balanced.
	IsWellFormed(regs []Register) error
	// Registers returns the set of registers read/written by this micro instruction.
	Registers() []uint
	// Registers returns the set of registers read this micro instruction.
	RegistersRead() []uint
	// Registers returns the set of registers written by this micro instruction.
	RegistersWritten() []uint
	// Translate this micro instruction into constraints using the given state
	// translator.
	Translate(st *StateTranslator)
}

// Instruction represents the composition of one or more micro instructions
// which are to be executed "in parallel".  This roughly following the ideas of
// vector machines and vectorisation.  In order to ensure parallel execution is
// safe, there are restrictions on how microinstructions can be combined.  For
// example, two microinstructions writing to the same register are said to be
// "conflicting" and, hence, this is not permitted.  Likewise, it is not
// possible to branch into the middle of a microinstruction.
type Instruction struct {
	Instructions []MicroInstruction
}

// Sequential indicates whether or not this microinstruction can execute
// sequentially onto the next.
func (p *Instruction) Sequential() bool {
	n := len(p.Instructions) - 1
	// Only need to check last instruction to determine this.
	return p.Instructions[n].Sequential()
}

// Terminal indicates whether or not this instruction is a terminating
// instruction (or not).  That is, whether or not its possible for control-flow
// to "fall through" to the next instruction.
func (p *Instruction) Terminal() bool {
	n := len(p.Instructions) - 1
	// Only need to check last instruction to determine this.
	return p.Instructions[n].Terminal()
}

// Bind any labels contained within this instruction using the given label map.
func (p *Instruction) Bind(labels []uint) {
	for _, insn := range p.Instructions {
		insn.Bind(labels)
	}
}

// Execute a given instruction at a given program counter position, using a
// given set of register values.  This may update the register values, and
// returns the next program counter position.  If the program counter is
// math.MaxUint then a return is signaled.
func (p *Instruction) Execute(pc uint, state []big.Int, regs []Register) uint {
	var npc uint = pc + 1

	for _, r := range p.Instructions {
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

// Validate that this micro-instruction is well-formed.  For example, each
// micro-instruction contained within must be well-formed, and the overall
// requirements for a vector instruction must be met, etc.
func (p *Instruction) Validate(regs []Register) (uint, error) {
	var (
		written bit.Set
		n       = len(p.Instructions) - 1
	)
	//
	for i, r := range p.Instructions {
		if err := r.IsWellFormed(regs); err != nil {
			return uint(i), err
		}
	}
	// Check read-after-write conflicts
	for i, r := range p.Instructions {
		for _, src := range r.RegistersRead() {
			if written.Contains(src) {
				// Forwarding required for this
				return uint(i), fmt.Errorf("conflicting reading (requires forwarding)")
			}
		}
		//
		for _, dst := range r.RegistersWritten() {
			written.Insert(dst)
		}
	}
	// Check for unreachable instructions
	for i, r := range p.Instructions {
		if i != n && !r.Sequential() {
			return uint(i), fmt.Errorf("unreachable")
		}
	}
	//
	return math.MaxUint, nil
}
