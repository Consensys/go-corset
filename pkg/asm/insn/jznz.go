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

	"github.com/consensys/go-corset/pkg/hir"
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
func (p *Jznz) Execute(pc uint, state []big.Int, regs []Register) uint {
	var (
		val     = state[p.Source]
		is_zero = val.Cmp(&zero) == 0
	)
	//
	if p.Sign == is_zero {
		return p.Target
	}
	//
	return pc + 1
}

// IsWellFormed checks whether or not this instruction is correctly balanced.
func (p *Jznz) IsWellFormed(regs []Register) error {
	return nil
}

// Registers returns the set of registers read/written by this instruction.
func (p *Jznz) Registers() []uint {
	return []uint{p.Source}
}

// RegistersRead returns the set of registers read by this instruction.
func (p *Jznz) RegistersRead() []uint {
	return []uint{p.Source}
}

// RegistersWritten returns the set of registers written by this instruction.
func (p *Jznz) RegistersWritten() []uint {
	return nil
}

// Translate this instruction into low-level constraints.
func (p *Jznz) Translate(pc uint, state StateTranslator) {
	if p.Sign {
		p.translateJz(pc, state)
	} else {
		p.translateJnz(pc, state)
	}
}

func (p *Jznz) translateJz(pc uint, st StateTranslator) {
	//
	var (
		pc_i   = hir.NewColumnAccess(st.PcID, 0)
		pc_ip1 = hir.NewColumnAccess(st.PcID, 1)
		reg_i  = hir.NewColumnAccess(st.RegIDs[p.Source], 0)
		target = hir.NewConst64(uint64(p.Target))
	)
	// taken
	st.Constrain("jz", pc, hir.Disjunction(hir.NotEquals(reg_i, hir.ZERO), hir.Equals(pc_ip1, target)))
	// not taken
	st.Constrain("jnz", pc,
		hir.Disjunction(hir.Equals(reg_i, hir.ZERO), hir.Equals(pc_ip1, hir.Sum(pc_i, hir.ONE))))
}

func (p *Jznz) translateJnz(pc uint, st StateTranslator) {
	//
	var (
		pc_i   = hir.NewColumnAccess(st.PcID, 0)
		pc_ip1 = hir.NewColumnAccess(st.PcID, 1)
		reg_i  = hir.NewColumnAccess(st.RegIDs[p.Source], 0)
		target = hir.NewConst64(uint64(p.Target))
	)
	// taken
	st.Constrain("jnz", pc,
		hir.Disjunction(hir.Equals(reg_i, hir.ZERO), hir.Equals(pc_ip1, target)))
	// not taken
	st.Constrain("jz", pc,
		hir.Disjunction(hir.NotEquals(reg_i, hir.ZERO), hir.Equals(pc_ip1, hir.Sum(pc_i, hir.ONE))))
	// register constancies
	st.ConstantExcept(pc, nil)
}
