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

	"github.com/consensys/go-corset/pkg/asm/io"
	"github.com/consensys/go-corset/pkg/asm/io/micro"
	"github.com/consensys/go-corset/pkg/util"
	"github.com/consensys/go-corset/pkg/util/collection/bit"
)

// Translator encapsulates general information related to the mapping from
// instructions down to constraints.
type Translator[T any, E Expr[T, E], M Module[T, E, M]] struct {
	// Enclosing module to which constraints are added.
	Module M
	// Column ID for STAMP HIR column.
	Stamp T
	// Column ID for PC HIR column.
	ProgramCounter T
	// Registers of the given machine
	Registers []io.Register
	// Mapping from registers to column IDs in the underlying constraint system.
	Columns []T
}

// Translate a micro-instruction at a given program counter position.
func (p *Translator[T, E, M]) Translate(pc uint, insn micro.Instruction) {
	var (
		tr         = NewStateTranslator(*p, pc, insn)
		constraint = tr.translateCode(0, insn.Codes)
		pcGuard    = tr.ReadPc().Equals(Number[T, E](pc))
		stampGuard = tr.Stamp(false).NotEquals(Number[T, E](0))
		name       = fmt.Sprintf("pc%d", pc)
	)
	// Apply global constancies
	constraint = tr.WithGlobalConstancies(constraint)
	// Apply state guards
	constraint = If(stampGuard.And(pcGuard), constraint)
	//
	p.Module.NewConstraint(name, util.None[int](), constraint)
}

// StateTranslator packages up key information regarding how an individual state
// of the machine is compiled down to the lower level.
type StateTranslator[T any, E Expr[T, E], M Module[T, E, M]] struct {
	mapping Translator[T, E, M]
	// Program counter
	pc uint
	// Set of registers not mutated by the enclosing instruction.
	constants bit.Set
	// Set of registers mutated on the current branch.
	mutated bit.Set
	// Set of registers currently being forwarded.
	forwarded bit.Set
}

// NewStateTranslator constructs a new translated for a given state (i.e.
// program counter location) with a given mapping.
func NewStateTranslator[T any, E Expr[T, E], M Module[T, E, M]](mapping Translator[T, E, M],
	pc uint, insn micro.Instruction) StateTranslator[T, E, M] {
	//
	var constants bit.Set
	// Initially include all registers
	for i := range mapping.Registers {
		constants.Insert(uint(i))
	}
	// Remove those which are actually modified
	for _, code := range insn.Codes {
		for _, reg := range code.RegistersWritten() {
			constants.Remove(reg.Unwrap())
		}
	}
	//
	return StateTranslator[T, E, M]{
		mapping:   mapping,
		pc:        pc,
		constants: constants,
		mutated:   bit.Set{},
		forwarded: bit.Set{},
	}
}

// Stamp returns a column access for either the stamp on this row, or the stamp
// on the next row.
func (p *StateTranslator[T, E, M]) Stamp(next bool) E {
	if next {
		return Variable[T, E](p.mapping.Stamp, 1)
	}
	//
	return Variable[T, E](p.mapping.Stamp, 0)
}

// ReadPc returns a column access for current the pc value.
func (p *StateTranslator[T, E, M]) ReadPc() E {
	return Variable[T, E](p.mapping.ProgramCounter, 0)
}

// WritePc returns a column access for suitable for setting the next PC value.
// This also marks the PC as mutated, meaning it will not be included in any
// constancy calculations.
func (p *StateTranslator[T, E, M]) WritePc() E {
	// Mark register as having been written.
	p.mutated.Insert(io.PC_INDEX)
	return Variable[T, E](p.mapping.ProgramCounter, 1)
}

// Clone creates a fresh copy of this translator.
func (p *StateTranslator[T, E, M]) Clone() StateTranslator[T, E, M] {
	return StateTranslator[T, E, M]{
		mapping:   p.mapping,
		constants: p.constants.Clone(),
		mutated:   p.mutated.Clone(),
		forwarded: p.forwarded.Clone(),
	}
}

// WriteRegisters constructs suitable accessors for the those registers written
// by a given microinstruction.  This activates forwarding for those registers
// for all states after this, and returns suitable expressions for the
// assignment.
func (p *StateTranslator[T, E, M]) WriteRegisters(targets []io.RegisterId) []E {
	lhs := make([]E, len(targets))
	// build up the lhs
	for i, dst := range targets {
		lhs[i] = Variable[T, E](p.mapping.Columns[dst.Unwrap()], 0)
		// Activate forwarding for this register
		p.forwarded.Insert(dst.Unwrap())
		// Mark register as having been written.
		p.mutated.Insert(dst.Unwrap())
	}
	//
	return lhs
}

// WriteAndShiftRegisters constructs suitable accessors for the those registers
// written by a given microinstruction, and also shifts them (i.e. so they can
// be combined in a sum).  This activates forwarding for those registers for all
// states after this, and returns suitable expressions for the assignment.
func (p *StateTranslator[T, E, M]) WriteAndShiftRegisters(targets []io.RegisterId) []E {
	lhs := make([]E, len(targets))
	offset := big.NewInt(1)
	// build up the lhs
	for i, dst := range targets {
		lhs[i] = Variable[T, E](p.mapping.Columns[dst.Unwrap()], 0)
		//
		if i != 0 {
			lhs[i] = BigNumber[T, E](offset).Multiply(lhs[i])
		}
		// left shift offset by given register width.
		offset.Lsh(offset, p.mapping.Registers[dst.Unwrap()].Width)
		// Activate forwarding for this register
		p.forwarded.Insert(dst.Unwrap())
		// Mark register as having been written.
		p.mutated.Insert(dst.Unwrap())
	}
	//
	return lhs
}

// ReadRegister constructs a suitable accessor for referring to a given register.
// This applies forwarding as appropriate.
func (p *StateTranslator[T, E, M]) ReadRegister(reg io.RegisterId) E {
	rid := p.mapping.Columns[reg.Unwrap()]
	//
	if p.mapping.Registers[reg.Unwrap()].IsInput() {
		// Inputs don't need to refer back
		return Variable[T, E](rid, 0)
	} else if p.forwarded.Contains(reg.Unwrap()) {
		// Forwarded
		return Variable[T, E](rid, 0)
	}
	// Not forwarded
	return Variable[T, E](rid, -1)
}

// ReadRegisters constructs appropriate column accesses for a given set of
// registers.  When appropriate, forwarding will be applied automatically.
func (p *StateTranslator[T, E, M]) ReadRegisters(sources []io.RegisterId) []E {
	rhs := make([]E, len(sources))
	// build up the lhs
	for i, src := range sources {
		rhs[i] = p.ReadRegister(src)
	}
	//
	return rhs
}

// WithLocalConstancies adds constancy constraints for all registers not
// mutated by a given branch through an instruction.
func (p *StateTranslator[T, E, M]) WithLocalConstancies(condition E) E {
	// FIXME: following check is temporary hack
	if p.pc > 0 {
		//
		for i, r := range p.mapping.Registers {
			rid := p.mapping.Columns[i]
			//
			if !r.IsInput() && !p.constants.Contains(uint(i)) && !p.mutated.Contains(uint(i)) {
				r_i := Variable[T, E](rid, 0)
				r_im1 := Variable[T, E](rid, -1)
				constancy := r_i.Equals(r_im1)
				//
				condition = condition.And(constancy)
			}
		}
	}
	//
	return condition
}

// WithGlobalConstancies adds constancy constraints for all registers not
// mutated at all by an instruction.
func (p *StateTranslator[T, E, M]) WithGlobalConstancies(condition E) E {
	// FIXME: following check is temporary hack
	if p.pc > 0 {
		for i, r := range p.mapping.Registers {
			rid := p.mapping.Columns[i]
			//
			if !r.IsInput() && p.constants.Contains(uint(i)) {
				r_i := Variable[T, E](rid, 0)
				r_im1 := Variable[T, E](rid, -1)
				constancy := r_i.Equals(r_im1)
				//
				condition = condition.And(constancy)
			}
		}
	}
	//
	return condition
}
