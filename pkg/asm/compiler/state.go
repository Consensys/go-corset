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
package compiler

import (
	"fmt"
	"math/big"

	"github.com/consensys/gnark-crypto/ecc/bls12-377/fr"
	"github.com/consensys/go-corset/pkg/asm/io"
	"github.com/consensys/go-corset/pkg/asm/io/micro"
	"github.com/consensys/go-corset/pkg/hir"
	"github.com/consensys/go-corset/pkg/trace"
	"github.com/consensys/go-corset/pkg/util"
	"github.com/consensys/go-corset/pkg/util/collection/bit"
)

// Translator encapsulates general information related to the mapping from
// instructions down to constraints.
type Translator struct {
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
	Registers []io.Register
}

// Translate a micro-instruction at a given program counter position.
func (p *Translator) Translate(pc uint, insn micro.Instruction) {
	var (
		tr         = NewStateTranslator(*p, insn)
		constraint = translate(0, insn.Codes, tr)
		pcGuard    = hir.Equals(tr.Pc(false), hir.NewConst64(uint64(pc)))
		stampGuard = hir.NotEquals(tr.Stamp(false), hir.ZERO)
		name       = fmt.Sprintf("pc%d", pc)
	)
	// Apply global constancies
	constraint = tr.WithGlobalConstancies(constraint)
	// Apply state guards
	constraint = hir.If(hir.Conjunction(stampGuard, pcGuard), constraint)
	//
	p.Schema.AddVanishingConstraint(name, p.Context, util.None[int](), constraint)
}

// StateTranslator packages up key information regarding how an individual state
// of the machine is compiled down to the lower level.
type StateTranslator struct {
	mapping Translator
	// Set of registers not mutated by the enclosing instruction.
	constants bit.Set
	// Set of registers mutated on the current branch.
	mutated bit.Set
	// Set of registers currently being forwarded.
	forwarded bit.Set
}

// NewStateTranslator constructs a new translated for a given state (i.e.
// program counter location) with a given mapping.
func NewStateTranslator(mapping Translator, insn micro.Instruction) StateTranslator {
	var constants bit.Set
	// Initially include all registers
	for i := range mapping.Registers {
		constants.Insert(uint(i))
	}
	// Remove those which are actually modified
	for _, code := range insn.Codes {
		for _, reg := range code.RegistersWritten() {
			constants.Remove(reg)
		}
	}
	//
	return StateTranslator{
		mapping:   mapping,
		constants: constants,
		mutated:   bit.Set{},
		forwarded: bit.Set{},
	}
}

// Stamp returns a column access for either the stamp on this row, or the stamp
// on the next row.
func (p *StateTranslator) Stamp(next bool) hir.Expr {
	if next {
		return hir.NewColumnAccess(p.mapping.StampID, 1)
	}
	//
	return hir.NewColumnAccess(p.mapping.StampID, 0)
}

// Pc returns a column access for either the pc on this row, or the pc on the
// next row.
func (p *StateTranslator) Pc(next bool) hir.Expr {
	if next {
		return hir.NewColumnAccess(p.mapping.PcID, 1)
	}
	//
	return hir.NewColumnAccess(p.mapping.PcID, 0)
}

// Clone creates a fresh copy of this translator.
func (p *StateTranslator) Clone() StateTranslator {
	return StateTranslator{
		mapping:   p.mapping,
		constants: p.constants.Clone(),
		mutated:   p.mutated.Clone(),
		forwarded: p.forwarded.Clone(),
	}
}

// WriteRegisters identifies the set of registers written by the current microinstruction
// being translated.  This activates forwarding for those registers for all
// states after this, and returns suitable expressions for the assignment.
func (p *StateTranslator) WriteRegisters(targets []uint) []hir.Expr {
	lhs := make([]hir.Expr, len(targets))
	offset := big.NewInt(1)
	// build up the lhs
	for i, dst := range targets {
		lhs[i] = hir.NewColumnAccess(p.mapping.RegIDs[dst], 0)
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
	if p.mapping.Registers[reg].IsInput() {
		// Inputs don't need to refer back
		return hir.NewColumnAccess(rid, 0)
	} else if p.forwarded.Contains(reg) {
		// Forwarded
		return hir.NewColumnAccess(rid, 0)
	}
	// Not forwarded
	return hir.NewColumnAccess(rid, -1)
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

// WithLocalConstancies adds constancy constraints for all registers not
// mutated by a given branch through an instruction.
func (p *StateTranslator) WithLocalConstancies(condition hir.Expr) hir.Expr {
	//
	for i, r := range p.mapping.Registers {
		rid := p.mapping.RegIDs[i]
		//
		if !r.IsInput() && !p.constants.Contains(uint(i)) && !p.mutated.Contains(uint(i)) {
			r_i := hir.NewColumnAccess(rid, 0)
			r_im1 := hir.NewColumnAccess(rid, -1)
			constancy := hir.Equals(r_i, r_im1)
			//
			condition = hir.Conjunction(condition, constancy)
		}
	}
	//
	return condition
}

// WithGlobalConstancies adds constancy constraints for all registers not
// mutated at all by an instruction.
func (p *StateTranslator) WithGlobalConstancies(condition hir.Expr) hir.Expr {
	//
	for i, r := range p.mapping.Registers {
		rid := p.mapping.RegIDs[i]
		//
		if !r.IsInput() && p.constants.Contains(uint(i)) {
			r_i := hir.NewColumnAccess(rid, 0)
			r_im1 := hir.NewColumnAccess(rid, -1)
			constancy := hir.Equals(r_i, r_im1)
			//
			condition = hir.Conjunction(condition, constancy)
		}
	}
	//
	return condition
}
