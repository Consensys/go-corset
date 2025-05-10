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
	"math/big"

	"github.com/consensys/go-corset/pkg/asm/io"
	"github.com/consensys/go-corset/pkg/asm/io/micro"
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

// Execute a ret instruction by signaling a return from the enclosing function.
func (p *Ret) Execute(pc uint, state []big.Int, regs []io.Register) uint {
	return io.RETURN
}

// Lower this instruction into a exactly one more micro instruction.
func (p *Ret) Lower(pc uint) micro.Instruction {
	// Lowering here produces an instruction containing a single microcode.
	return micro.NewInstruction(&micro.Ret{})
}

// Link any buses used within this instruction using the given bus map.
func (p *Ret) Link(buses []uint) {
	// nothing to link
}

// RegistersRead returns the set of registers read by this instruction.
func (p *Ret) RegistersRead() []uint {
	return nil
}

// RegistersWritten returns the set of registers written by this instruction.
func (p *Ret) RegistersWritten() []uint {
	return nil
}

func (p *Ret) String(env io.Environment[Instruction]) string {
	return "return"
}

// Validate checks whether or not this instruction is correctly balanced.
func (p *Ret) Validate(env io.Environment[Instruction]) error {
	return nil
}
