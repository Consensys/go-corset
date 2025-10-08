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
	"github.com/consensys/go-corset/pkg/util/field"
)

// Translator encapsulates general information related to the mapping from
// instructions down to constraints.
type Translator[F field.Element[F], T any, E Expr[T, E], M Module[F, T, E, M]] struct {
	// Enclosing module to which constraints are added.
	Module M
	// Framining identifies any required control lines.
	Framing Framing[T, E]
	// Registers of the given machine
	Registers []io.Register
	// ioLines identifies which registers are used purely for buses.  This is
	// because these registers can be treated in a more relaxed fashion than
	// other registers.
	ioLines bit.Set
	// Mapping from registers to column IDs in the underlying constraint system.
	Columns []T
}

// Translate a micro-instruction at a given program counter position.
func (p *Translator[F, T, E, M]) Translate(pc uint, insn micro.Instruction) {
	var (
		tr         = NewStateTranslator(*p, pc, insn)
		constraint = tr.translateCode(0, insn.Codes)
		name       = fmt.Sprintf("pc%d", pc)
	)
	// Apply global constancies
	constraint = tr.WithGlobalConstancies(constraint)
	// Apply framing guards (if applicable)
	constraint = If(p.Framing.Guard(pc), constraint)
	//
	p.Module.NewConstraint(name, util.None[int](), constraint)
}

// StateReader is a simplified view of a state translator which is suitable for
// reading registers only.
type StateReader[T any, E Expr[T, E]] interface {
	// ReadRegister constructs a suitable accessor for referring to a given register.
	// This applies forwarding as appropriate.
	ReadRegister(reg io.RegisterId) E
}

// StateTranslator packages up key information regarding how an individual state
// of the machine is compiled down to the lower level.
type StateTranslator[F field.Element[F], T any, E Expr[T, E], M Module[F, T, E, M]] struct {
	mapping Translator[F, T, E, M]
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
func NewStateTranslator[F field.Element[F], T any, E Expr[T, E], M Module[F, T, E, M]](mapping Translator[F, T, E, M],
	pc uint, insn micro.Instruction) StateTranslator[F, T, E, M] {
	//
	var constants bit.Set
	// Initially include all registers
	for i := range mapping.Registers {
		// I/O lines are never considered global constants.
		if !mapping.ioLines.Contains(uint(i)) {
			constants.Insert(uint(i))
		}
	}
	// Remove those which are actually modified
	for _, code := range insn.Codes {
		for _, reg := range code.RegistersWritten() {
			constants.Remove(reg.Unwrap())
		}
	}
	//
	return StateTranslator[F, T, E, M]{
		mapping:   mapping,
		pc:        pc,
		constants: constants,
		mutated:   bit.Set{},
		forwarded: bit.Set{},
	}
}

// Terminate current frame, and setup for next frame.
func (p *StateTranslator[F, T, E, M]) Terminate() E {
	return p.WithLocalConstancies(p.mapping.Framing.Return())
}

// Goto returns an expression suitable for ensuring that the given instruction
// transitions to the state representing the given PC value.
func (p *StateTranslator[F, T, E, M]) Goto(pc uint) E {
	// Apply framing
	return p.WithLocalConstancies(p.mapping.Framing.Goto(pc))
}

// Clone creates a fresh copy of this translator.
func (p *StateTranslator[F, T, E, M]) Clone() StateTranslator[F, T, E, M] {
	return StateTranslator[F, T, E, M]{
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
func (p *StateTranslator[F, T, E, M]) WriteRegisters(targets []io.RegisterId) []E {
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
func (p *StateTranslator[F, T, E, M]) WriteAndShiftRegisters(targets []io.RegisterId) []E {
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
func (p *StateTranslator[F, T, E, M]) ReadRegister(reg io.RegisterId) E {
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
func (p *StateTranslator[F, T, E, M]) ReadRegisters(sources []io.RegisterId) []E {
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
func (p *StateTranslator[F, T, E, M]) WithLocalConstancies(condition E) E {
	if p.pc > 0 {
		for i := range p.mapping.Registers {
			rid := p.mapping.Columns[i]
			//
			if p.IsLocalConstancy(uint(i)) {
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

// IsLocalConstancy determines whether a given register should be given a
// constancy constraint to ensure its current value matches its previous value.
func (p *StateTranslator[F, T, E, M]) IsLocalConstancy(id uint) bool {
	r := p.mapping.Registers[id]
	//
	return !r.IsInput() && !p.constants.Contains(id) &&
		!p.mutated.Contains(id) && !p.mapping.ioLines.Contains(id)
}

// WithGlobalConstancies adds constancy constraints for all registers not
// mutated at all by an instruction.
func (p *StateTranslator[F, T, E, M]) WithGlobalConstancies(condition E) E {
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

func (p *StateTranslator[F, T, E, M]) translateCode(cc uint, codes []micro.Code) E {
	switch codes[cc].(type) {
	case *micro.Assign:
		return p.translateAssign(cc, codes)
	case *micro.Fail:
		return False[T, E]()
	case *micro.InOut:
		return p.translateInOut(cc, codes)
	case *micro.Ite:
		return p.translateIte(cc, codes)
	case *micro.Jmp:
		return p.translateJmp(cc, codes)
	case *micro.Ret:
		return p.translateRet()
	case *micro.Skip:
		return p.translateSkip(cc, codes)
	default:
		panic("unreachable")
	}
}
