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
	"math/big"
)

// Jznz describes a conditional branch, which is either jz ("Jump if Zero") or
// jzn ("Jump if not Zero").  As expected, jz jumps if the source register is
// zero whilst jnz jumps it if is non-zero.
type Jznz struct {
	// Sign indicates jz (true) or jnz (false)
	Sign bool
	// Source register being tested.
	Source uint
	// Target identifies target PC
	Target uint
}

// Bind any labels contained within this instruction using the given label map.
func (p *Jznz) Bind(labels []uint) {
	p.Target = labels[p.Target]
}

// Execute an unconditional branch instruction by returning the destination
// program counter.
func (p *Jznz) Execute(pc uint, regs []big.Int, widths []uint) uint {
	var (
		val     = regs[p.Source]
		is_zero = val.Cmp(&zero) == 0
	)
	//
	if p.Sign == is_zero {
		return p.Target
	}
	//
	return pc + 1
}

// Registers returns the set of registers read/written by this instruction.
func (p *Jznz) Registers() []uint {
	return []uint{p.Source}
}
