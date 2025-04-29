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

import (
	"math"
	"math/big"
)

// Ret signals a return from the enclosing function.
type Ret struct{}

// Bind any labels contained within this instruction using the given label map.
func (p *Ret) Bind(labels []uint) {
	// no-op
}

// Execute a ret instruction by signaling a return from the enclosing function.
func (p *Ret) Execute(pc uint, state []big.Int, regs []Register) uint {
	return math.MaxUint
}

// IsBalanced checks whether or not this instruction is correctly balanced.
func (p *Ret) IsBalanced(regs []Register) error {
	return nil
}

// Registers returns the set of registers read/written by this instruction.
func (p *Ret) Registers() []uint {
	return nil
}
