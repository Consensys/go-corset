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
	"math/big"
	"slices"

	"github.com/consensys/go-corset/pkg/asm/io"
)

// Call represents a function call providing one or more arguments and accepting
// zero or more values in return.  A function call requires a "bus" to read and
// write its arguments / returns.  A bus is a set of dedicated registers
// providing an I/O communication channel to a given peripheral (in this case,
// another function).
type Call struct {
	// Bus identifies the relevant bus for this instruction.
	Bus uint
	// Target registers for addition
	Targets []uint
	// Source registers (i.e. arguments) for call
	Sources []uint
}

// Bind any labels contained within this instruction using the given label map.
func (p *Call) Bind(labels []uint) {
	// no-op
}

// Clone this micro code.
func (p *Call) Clone() Code {
	//
	return &Call{
		p.Bus,
		slices.Clone(p.Targets),
		slices.Clone(p.Sources),
	}
}

// Execute a given instruction at a given program counter position, using a
// given set of register values.  This may update the register values, and
// returns the next program counter position.  If the program counter is
// math.MaxUint then a return is signaled.
func (p *Call) Execute(pc uint, state []big.Int, regs []io.Register) uint {
	p.MicroExecute(state, regs)
	return pc + 1
}

// MicroExecute a given micro-code, using a given set of register values.  This
// may update the register values, and returns either the number of micro-codes
// to "skip over" when executing the enclosing instruction or, if skip==0, a
// destination program counter (which can signal return of enclosing function).
func (p *Call) MicroExecute(state []big.Int, regs []io.Register) (uint, uint) {
	panic("todo")
}

// Lower this instruction into a exactly one more micro instruction.
func (p *Call) Lower(pc uint) Instruction {
	panic("todo")
}

// Registers returns the set of registers read/written by this instruction.
func (p *Call) Registers() []uint {
	return append(p.Targets, p.Sources...)
}

// RegistersRead returns the set of registers read by this instruction.
func (p *Call) RegistersRead() []uint {
	return p.Sources
}

// RegistersWritten returns the set of registers written by this instruction.
func (p *Call) RegistersWritten() []uint {
	return p.Targets
}

func (p *Call) String(regs []io.Register) string {
	panic("todo")
}

// Split this micro code using registers of arbirary width into one or more
// micro codes using registers of a fixed maximum width.  Here, regsBefore
// represents the registers are they are for this code, whilst regsAfter
// represent those for the resulting split codes.  The regMap provides a
// mapping from registers in regsBefore to those in regsAfter.
func (p *Call) Split(env *RegisterSplittingEnvironment) []Code {
	panic("todo")
}

// Validate checks whether or not this instruction well-formed.
func (p *Call) Validate(fieldWidth uint, regs []io.Register) error {
	panic("todo")
}
