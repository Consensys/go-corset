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

	"github.com/consensys/go-corset/pkg/asm/io"
)

// Ret signals a return from the enclosing function.
type Ret struct {
	// dummy is included to force Ret structs to be stored in the heap.
	//nolint
	dummy uint
}

// Bind any labels contained within this instruction using the given label map.
func (p *Ret) Bind(labels []uint) {
	// no-op
}

// Clone this micro code.
func (p *Ret) Clone() Code {
	return p
}

// Execute a ret instruction by signaling a return from the enclosing function.
func (p *Ret) Execute(pc uint, state []big.Int, regs []io.Register) uint {
	return io.RETURN
}

// MicroExecute a given micro-code, using a given set of register values.  This
// may update the register values, and returns either the number of micro-codes
// to "skip over" when executing the enclosing instruction or, if skip==0, a
// destination program counter (which can signal return of enclosing function).
func (p *Ret) MicroExecute(state []big.Int, regs []io.Register) (uint, uint) {
	return 0, io.RETURN
}

// Lower this instruction into a exactly one more micro instruction.
func (p *Ret) Lower(pc uint) Instruction {
	// Lowering here produces an instruction containing a single microcode.
	return Instruction{[]Code{p}}
}

// Registers returns the set of registers read/written by this instruction.
func (p *Ret) Registers() []uint {
	return nil
}

// RegistersRead returns the set of registers read by this instruction.
func (p *Ret) RegistersRead() []uint {
	return nil
}

// RegistersWritten returns the set of registers written by this instruction.
func (p *Ret) RegistersWritten() []uint {
	return nil
}

// Split this micro code using registers of arbirary width into one or more
// micro codes using registers of a fixed maximum width.
func (p *Ret) Split(env *RegisterSplittingEnvironment) []Code {
	return []Code{p}
}

func (p *Ret) String(regs []io.Register) string {
	return "ret"
}

// Validate checks whether or not this instruction is correctly balanced.
func (p *Ret) Validate(fieldWidth uint, regs []io.Register) error {
	return nil
}
