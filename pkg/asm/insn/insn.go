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
	"math/big"
	"slices"

	"github.com/consensys/gnark-crypto/ecc/bls12-377/fr"
	"github.com/consensys/go-corset/pkg/hir"
	"github.com/consensys/go-corset/pkg/trace"
	"github.com/consensys/go-corset/pkg/util"
)

// Instruction provides an abstract notion of a "machine instruction".
type Instruction interface {
	// Bind any labels contained within this instruction using the given label map.
	Bind(labels []uint)
	// Execute a given instruction at a given program counter position, using a
	// given set of register values.  This may update the register values, and
	// returns the next program counter position.  If the program counter is
	// math.MaxUint then a return is signaled.
	Execute(pc uint, state []big.Int, regs []Register) uint
	// Check whether or not this instruction is well-formed (e.g. correctly balanced).
	IsWellFormed(regs []Register) error
	// Registers returns the set of registers read/written by this instruction.
	Registers() []uint
	// Registers returns the set of registers read this instruction.
	RegistersRead() []uint
	// Registers returns the set of registers written by this instruction.
	RegistersWritten() []uint
	// Translate this instruction into constraints.
	Translate(pc uint, st StateTranslator)
}

// StateTranslator packages up key information regarding how an individual state
// of the machine is compiled down to the lower level.
type StateTranslator struct {
	// Enclosing schema to which constraints are added.
	Schema *hir.Schema
	// Column ID for STAMP HIR column.
	StampID uint
	// Column ID for PC HIR column.
	PcID uint
	// Context of enclosing HIR module.
	Context trace.Context
	// Mapping from registers to respective HIR column IDs.
	RegIDs []uint
	// Registers of the given machine
	Registers []Register
}

// Constrain a given state in some way (for example, relating the forward state
// with the current state, etc).
func (p *StateTranslator) Constrain(name string, pc uint, constraint hir.Expr) {
	var (
		pc_i       = hir.NewColumnAccess(p.PcID, 0)
		pcGuard    = hir.NotEquals(pc_i, hir.NewConst64(uint64(pc)))
		stamp_i    = hir.NewColumnAccess(p.StampID, 0)
		stampGuard = hir.Equals(stamp_i, hir.ZERO)
	)
	//
	name = fmt.Sprintf("pc%d_%s", pc, name)
	// Apply necessary guards to ensure the given constraint only applies in the given state.
	constraint = hir.Disjunction(stampGuard, pcGuard, constraint)
	//
	p.Schema.AddVanishingConstraint(name, p.Context, util.None[int](), constraint)
}

// Construct a suitable left-hand side
func (p *StateTranslator) buildAssignmentLhs(targets []uint) []hir.Expr {
	lhs := make([]hir.Expr, len(targets))
	offset := big.NewInt(1)
	// build up the lhs
	for i, dst := range targets {
		lhs[i] = hir.NewColumnAccess(p.RegIDs[dst], 1)
		//
		if i != 0 {
			var elem fr.Element
			//
			elem.SetBigInt(offset)
			lhs[i] = hir.Product(hir.NewConst(elem), lhs[i])
		}
		// left shift offset by given register width.
		offset.Lsh(offset, p.Registers[dst].Width)
	}
	//
	return lhs
}

func (p *StateTranslator) buildAssignmentRhs(sources []uint) []hir.Expr {
	rhs := make([]hir.Expr, len(sources))
	// build up the lhs
	for i, src := range sources {
		rhs[i] = hir.NewColumnAccess(p.RegIDs[src], 0)
	}
	//
	return rhs
}

// ConstantExcept adds constancy constraints for all registers not assigned by a given insn.
func (p *StateTranslator) ConstantExcept(pc uint, targets []uint) {
	//
	var (
		pc_i       = hir.NewColumnAccess(p.PcID, 0)
		pcGuard    = hir.NotEquals(pc_i, hir.NewConst64(uint64(pc)))
		stamp_i    = hir.NewColumnAccess(p.StampID, 0)
		stampGuard = hir.Equals(stamp_i, hir.ZERO)
	)
	//
	for i, r := range p.Registers {
		if !slices.Contains(targets, uint(i)) {
			r_i := hir.NewColumnAccess(p.RegIDs[i], 0)
			r_ip1 := hir.NewColumnAccess(p.RegIDs[i], 1)
			eqn := hir.Equals(r_i, r_ip1)
			name := fmt.Sprintf("pc%d_%s", pc, r.Name)
			p.Schema.AddVanishingConstraint(name, p.Context, util.None[int](), hir.Disjunction(stampGuard, pcGuard, eqn))
		}
	}
}

// pc = pc + 1
func (p *StateTranslator) pcIncrement(pc uint) {
	stamp_i := hir.NewColumnAccess(p.StampID, 0)
	pc_i := hir.NewColumnAccess(p.PcID, 0)
	pc_ip1 := hir.NewColumnAccess(p.PcID, 1)
	//
	name := fmt.Sprintf("pc%d_clk", pc)
	//
	stGuard := hir.Equals(stamp_i, hir.ZERO)
	// pc != $PC
	pcGuard := hir.NotEquals(pc_i, hir.NewConst64(uint64(pc)))
	// pc = pc + 1
	inc := hir.Equals(pc_ip1, hir.Sum(hir.ONE, pc_i))
	//
	p.Schema.AddVanishingConstraint(name, p.Context, util.None[int](), hir.Disjunction(stGuard, pcGuard, inc))
}
