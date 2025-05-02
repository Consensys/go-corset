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
	"strings"

	"github.com/consensys/go-corset/pkg/asm/insn"
	"github.com/consensys/go-corset/pkg/util/collection/bit"
)

// Alias for big integer representation of 0.
var zero big.Int = *big.NewInt(0)

// Alias for big integer representation of 1.
var one big.Int = *big.NewInt(1)

// Register is an alias for insn.Register
type Register = insn.Register

// Code provides an abstract notion of an atomic "machine operation", where a
// single instruction is comprised of multiple such microcodes.  To ensure
// efficiency, we want to pack as many microcodes into each instruction as we
// can.  However, there are restrictions here meaning we cannot pack arbitrarily
// many microcodes into a single instruction.  For example, we cannot pack two
// microcodes together which have conflicting writes (i.e. both write to the
// same register).
type Code interface {
	insn.Instruction
	// Sequential indicates whether or not this microinstruction can execute
	// sequentially onto the next.
	Sequential() bool
	// Terminal indicates whether or not this microinstruction terminates the
	// enclosing function.
	Terminal() bool
}

// Instruction represents the composition of one or more micro instructions
// which are to be executed "in parallel".  This roughly following the ideas of
// vector machines and vectorisation.  In order to ensure parallel execution is
// safe, there are restrictions on how microcodes can be combined.  For example,
// two microcodes writing to the same register are said to be "conflicting" and,
// hence, this is not permitted.  Likewise, it is not possible to branch into
// the middle of a microinstruction.
type Instruction struct {
	Codes []Code
}

// Sequential indicates whether or not this microinstruction can execute
// sequentially onto the next.
func (p *Instruction) Sequential() bool {
	n := len(p.Codes) - 1
	// Only need to check last instruction to determine this.
	return p.Codes[n].Sequential()
}

// Terminal indicates whether or not this instruction is a terminating
// instruction (or not).  That is, whether or not its possible for control-flow
// to "fall through" to the next instruction.
func (p Instruction) Terminal() bool {
	n := len(p.Codes) - 1
	// Only need to check last instruction to determine this.
	return p.Codes[n].Terminal()
}

// Execute a given instruction at a given program counter position, using a
// given set of register values.  This may update the register values, and
// returns the next program counter position.  If the program counter is
// math.MaxUint then a return is signaled.
func (p Instruction) Execute(state []big.Int, regs []Register) uint {
	//
	for _, r := range p.Codes {
		npc := r.Execute(state, regs)
		// Sanity check
		if npc != insn.FALL_THRU {
			return npc
		}
	}
	// Fall through
	return insn.FALL_THRU
}

// Registers returns the set of registers read/written by this instruction.
func (p Instruction) Registers() []uint {
	return append(p.RegistersRead(), p.RegistersWritten()...)
}

// RegistersRead returns the set of registers read by this instruction.
func (p Instruction) RegistersRead() []uint {
	var regs bit.Set
	//
	for _, c := range p.Codes {
		regs.InsertAll(c.RegistersRead()...)
	}
	//
	return regs.Iter().Collect()
}

// RegistersWritten returns the set of registers written by this instruction.
func (p Instruction) RegistersWritten() []uint {
	var regs bit.Set
	//
	for _, c := range p.Codes {
		regs.InsertAll(c.RegistersWritten()...)
	}
	//
	return regs.Iter().Collect()
}

func (p Instruction) String(regs []Register) string {
	var builder strings.Builder
	//
	for i, code := range p.Codes {
		if i != 0 {
			builder.WriteString(" ; ")
		}
		//
		builder.WriteString(code.String(regs))
	}
	//
	return builder.String()
}

// Validate that this micro-instruction is well-formed.  For example, each
// micro-instruction contained within must be well-formed, and the overall
// requirements for a vector instruction must be met, etc.
func (p Instruction) Validate(regs []Register) error {
	var (
		written bit.Set
		n       = len(p.Codes) - 1
	)
	//
	for _, r := range p.Codes {
		if err := r.Validate(regs); err != nil {
			return err
		}
	}
	// Check read-after-write conflicts
	for _, r := range p.Codes {
		for _, dst := range r.RegistersWritten() {
			if written.Contains(dst) {
				// Forwarding required for this
				return fmt.Errorf("conflicting write")
			}
			//
			written.Insert(dst)
		}
	}
	// Check for unreachable instructions
	for i, r := range p.Codes {
		if i != n && !r.Sequential() {
			return fmt.Errorf("unreachable")
		}
	}
	//
	return nil
}
