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

	"github.com/consensys/gnark-crypto/ecc/bls12-377/fr"
	"github.com/consensys/go-corset/pkg/hir"
	"github.com/consensys/go-corset/pkg/trace"
	"github.com/consensys/go-corset/pkg/util"
	"github.com/consensys/go-corset/pkg/util/collection/bit"
)

// StateMapping encapsulates general information related to the mapping from
// instructions down to constraints.
type StateMapping struct {
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

// StateTranslator packages up key information regarding how an individual state
// of the machine is compiled down to the lower level.
type StateTranslator struct {
	mapping StateMapping
	// Program counter for given state
	pc uint
	// Indicates can still fall through to next instruction.
	fallThru bool
	// Terminal indicates whether node is terminating, or not.
	terminal bool
	// Set of registers which can be mutated by the enclosing instruction.  This
	// is used to determine the "common constancies".
	mutated bit.Set
	// Set of registers currently being forwarded.
	forwarded bit.Set
	// Currently active condition, which can be VOID if there is not active
	// condition.
	condition hir.Expr
}

// NewStateTranslator constructs a new translated for a given state (i.e.
// program counter location) with a given mapping.
func NewStateTranslator(mapping StateMapping, pc uint) StateTranslator {
	//
	return StateTranslator{
		mapping:   mapping,
		pc:        pc,
		fallThru:  true,
		terminal:  false,
		mutated:   bit.Set{},
		forwarded: bit.Set{},
		condition: hir.VOID,
	}
}

// Pc constructs an accessor which refers to the program counter either in this
// state, or the next.
func (p *StateTranslator) Pc(next bool) hir.Expr {
	if next {
		return hir.NewColumnAccess(p.mapping.PcID, 1)
	}
	//
	return hir.NewColumnAccess(p.mapping.PcID, 0)
}

// Stamp constructs an accessor which refers to the stamp counter either in this
// state, or the next.
func (p *StateTranslator) Stamp(next bool) hir.Expr {
	if next {
		return hir.NewColumnAccess(p.mapping.StampID, 1)
	}
	//
	return hir.NewColumnAccess(p.mapping.StampID, 0)
}

// Constrain a given state in some way (for example, relating the forward state
// with the current state, etc).
func (p *StateTranslator) Constrain(name string, constraint hir.Expr) {
	var (
		pc_i       = p.Pc(false)
		pcGuard    = hir.NotEquals(pc_i, hir.NewConst64(uint64(p.pc)))
		stamp_i    = p.Stamp(false)
		stampGuard = hir.Equals(stamp_i, hir.ZERO)
	)
	//
	name = fmt.Sprintf("pc%d_%s", p.pc, name)
	// Apply necessary guards to ensure the given constraint only applies in the given state.
	if p.condition != hir.VOID {
		constraint = hir.Disjunction(stampGuard, pcGuard, p.condition, constraint)
	} else {
		constraint = hir.Disjunction(stampGuard, pcGuard, constraint)
	}
	//
	p.mapping.Schema.AddVanishingConstraint(name, p.mapping.Context, util.None[int](), constraint)
}

// WriteRegisters identifies the set of registers written by the current microinstruction
// being translated.  This activates forwarding for those registers for all
// states after this, and returns suitable expressions for the assignment.
func (p *StateTranslator) WriteRegisters(targets []uint) []hir.Expr {
	lhs := make([]hir.Expr, len(targets))
	offset := big.NewInt(1)
	// build up the lhs
	for i, dst := range targets {
		lhs[i] = hir.NewColumnAccess(p.mapping.RegIDs[dst], 1)
		//
		if i != 0 {
			var elem fr.Element
			//
			elem.SetBigInt(offset)
			lhs[i] = hir.Product(hir.NewConst(elem), lhs[i])
		}
		// left shift offset by given register width.
		offset.Lsh(offset, p.mapping.Registers[dst].Width)
		// Activate forwarding for this register
		p.forwarded.Insert(dst)
		// Mark register as having been written.
		p.mutated.Insert(dst)
	}
	//
	return lhs
}

// ReadRegister constructs a suitable accessor for referring to a given register.
// This applies forwarding as appropriate.
func (p *StateTranslator) ReadRegister(reg uint) hir.Expr {
	rid := p.mapping.RegIDs[reg]
	//
	if p.forwarded.Contains(reg) {
		// Forwarded
		return hir.NewColumnAccess(rid, 1)
	}
	// Not forwarded
	return hir.NewColumnAccess(rid, 0)
}

// ReadRegisters constructs appropriate column accesses for a given set of
// registers.  When appropriate, forwarding will be applied automatically.
func (p *StateTranslator) ReadRegisters(sources []uint) []hir.Expr {
	rhs := make([]hir.Expr, len(sources))
	// build up the lhs
	for i, src := range sources {
		rhs[i] = p.ReadRegister(src)
	}
	//
	return rhs
}

// Condition updates the condition which applies to all remaining
// microinstructions in this state.
func (p *StateTranslator) Condition(cond hir.Expr) {
	if p.condition == hir.VOID {
		p.condition = cond
	} else {
		p.condition = hir.Disjunction(p.condition, cond)
	}
}

// Jump indicates that a given instruction wants to proceed unconditionally to a
// given location.  In many cases, this will be the following program counter
// location.
func (p *StateTranslator) Jump(target uint) {
	var (
		pc_ip1 = hir.NewColumnAccess(p.mapping.PcID, 1)
		nextPc = hir.NewConst64(uint64(target))
	)
	//
	p.Constrain("jmp", hir.Equals(pc_ip1, nextPc))
	p.fallThru = false
}

// JumpEq adds logic as needed to represent a conditional branch at this point.
func (p *StateTranslator) JumpEq(target uint, lhs hir.Expr, rhs hir.Expr) {
	p.jumpIf("je", target, hir.Equals(lhs, rhs), hir.NotEquals(lhs, rhs))
}

// JumpNe adds logic as needed to represent a conditional branch at this point.
func (p *StateTranslator) JumpNe(target uint, lhs hir.Expr, rhs hir.Expr) {
	p.jumpIf("jne", target, hir.NotEquals(lhs, rhs), hir.Equals(lhs, rhs))
}

func (p *StateTranslator) jumpIf(name string, target uint, trueBranch hir.Expr, falseBranch hir.Expr) {
	var (
		pc_ip1 = hir.NewColumnAccess(p.mapping.PcID, 1)
		nextPc = hir.NewConst64(uint64(target))
	)
	//
	p.Constrain(name, hir.Disjunction(falseBranch, hir.Equals(pc_ip1, nextPc)))
	// Apply condition for remaining microinstructions.
	p.Condition(trueBranch)
	// Constancies for true branch.  NOTE: this could be made more optimal by
	// pulling out the "common constancies".
	p.ConstantExcept(p.mutated, falseBranch)
}

// Terminate indicates that the current state is a terminal node.
func (p *StateTranslator) Terminate() {
	var (
		stamp_i   = hir.NewColumnAccess(p.mapping.StampID, 0)
		stamp_ip1 = hir.NewColumnAccess(p.mapping.StampID, 1)
	)
	//
	p.fallThru = false
	p.terminal = true
	// Force a new frame.
	p.Constrain("ret", hir.Equals(stamp_ip1, hir.Sum(hir.ONE, stamp_i)))
}

// Finalise translation
func (p *StateTranslator) Finalise() {
	if p.fallThru {
		// Handle fall through
		pc_i := hir.NewColumnAccess(p.mapping.PcID, 0)
		pc_ip1 := hir.NewColumnAccess(p.mapping.PcID, 1)
		// pc = pc + 1
		p.Constrain("next", hir.Equals(pc_ip1, hir.Sum(hir.ONE, pc_i)))
		// constancies
		p.ConstantExcept(p.mutated, hir.VOID)
	} else if p.terminal && p.forwarded.Count() != 0 {
		panic("cannot forward registers in terminating instruction")
	}
}

// ConstantExcept adds constancy constraints for all registers not assigned by a given insn.
func (p *StateTranslator) ConstantExcept(mutated bit.Set, condition hir.Expr) {
	for i, r := range p.mapping.Registers {
		rid := p.mapping.RegIDs[i]
		//
		if !r.IsInput() && !mutated.Contains(uint(i)) {
			name := fmt.Sprintf("const_%s", r.Name)
			r_i := hir.NewColumnAccess(rid, 0)
			r_ip1 := hir.NewColumnAccess(rid, 1)
			//
			if condition == hir.VOID {
				p.Constrain(name, hir.Equals(r_i, r_ip1))
			} else {
				p.Constrain(name, hir.Disjunction(condition, hir.Equals(r_i, r_ip1)))
			}
		}
	}
}
