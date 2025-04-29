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
package asm

import (
	"math/big"

	"github.com/consensys/go-corset/pkg/asm/instruction"
)

// Instruction provides an abstract notion of a "machine instruction".
type Instruction interface {
	// Bind any labels contained within this instruction using the given label map.
	Bind(labels []uint)
	// Execute a given instruction at a given program counter position, using a
	// given set of register values.  This may update the register values, and
	// returns the next program counter position.  If the program counter is
	// math.MaxUint then a return is signaled.
	Execute(pc uint, state []big.Int, regs []instruction.Register) uint
	// Check whether or not this instruction is well-formed (e.g. correctly balanced).
	IsWellFormed(regs []instruction.Register) error
	// Registers returns the set of registers read/written by this instruction.
	Registers() []uint
	// Registers returns the set of registers read this instruction.
	RegistersRead() []uint
	// Registers returns the set of registers written by this instruction.
	RegistersWritten() []uint
}
