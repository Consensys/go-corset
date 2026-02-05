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
	"github.com/consensys/go-corset/pkg/asm/io/micro/dfa"
	"github.com/consensys/go-corset/pkg/util/collection/bit"
	"github.com/consensys/go-corset/pkg/util/field"
)

// RegisterReader is a simplified view of a translator which is suitable for
// reading registers only.
type RegisterReader[T any, E Expr[T, E]] interface {
	// ReadRegister constructs a suitable accessor for referring to a given register.
	// This applies forwarding as appropriate.
	ReadRegister(reg io.RegisterId, forwarding bool) E
}

// Translator encapsulates general information related to the mapping from
// instructions down to constraints.
type Translator[F field.Element[F], T any, E Expr[T, E], M Module[F, T, E, M]] struct {
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

// Translate a micro instruction at a given Program Counter value into a given
// constraint.
func (p *Translator[F, T, E, M]) Translate(pc uint, insn micro.Instruction) E {
	var (
		nCodes              = uint(len(insn.Codes))
		constants           = p.determineConstants(insn)
		writes, branchTable = constructBranchTable[T, E](insn, p)
		constraint          = True[T, E]()
	)
	//
	for cc := uint(0); cc < nCodes; cc++ {
		var local E
		//
		switch c := insn.Codes[cc].(type) {
		case *micro.Assign:
			var str = StateTranslator[F, T, E, M]{*p, writes.StateOf(cc)}
			//
			local = str.translateAssign(c)
		case *micro.Division, *micro.InOut:
			// do nothing
			continue
		case *micro.Fail:
			local = False[T, E]()
		case *micro.Jmp:
			local = p.WithLocalConstancies(pc, p.Framing.Goto(c.Target))
		case *micro.Ret:
			local = p.WithLocalConstancies(pc, p.Framing.Return())
		case *micro.SkipIf, *micro.Skip:
			// do nothing
			continue
		default:
			panic("unreachable")
		}
		// Add control-flow requirements
		local = If(branchTable[cc], local)
		// Include local constraint
		constraint = constraint.And(local)
	}
	// Apply global constancies
	constraint = p.WithGlobalConstancies(pc, constants, constraint)
	// Add framing guards
	return If(p.Framing.Guard(pc), constraint)
}

// ReadRegister constructs a suitable accessor for referring to a given register.
// This applies forwarding as appropriate.
func (p *Translator[F, T, E, M]) ReadRegister(regId io.RegisterId, forwarding bool) E {
	var (
		reg   = p.Registers[regId.Unwrap()]
		colId = p.Columns[regId.Unwrap()]
	)
	//
	if reg.IsInput() {
		// Inputs don't need to refer back
		return Variable[T, E](colId, reg.Width(), 0)
	} else if forwarding {
		// Forwarded
		return Variable[T, E](colId, reg.Width(), 0)
	}
	// Not forwarded
	return Variable[T, E](colId, reg.Width(), -1)
}

func (p *Translator[F, T, E, M]) determineConstants(insn micro.Instruction) bit.Set {
	var constants bit.Set
	// Initially include all registers
	for i := range p.Registers {
		// I/O lines are never considered global constants.
		if !p.ioLines.Contains(uint(i)) {
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
	return constants
}

// WithGlobalConstancies adds constancy constraints for all registers not
// mutated at all by an instruction.
func (p *Translator[F, T, E, M]) WithGlobalConstancies(pc uint, constants bit.Set, condition E) E {
	if pc > 0 {
		//
		for i, reg := range p.Registers {
			rid := p.Columns[i]
			//
			if !reg.IsInput() && constants.Contains(uint(i)) {
				r_i := Variable[T, E](rid, reg.Width(), 0)
				r_im1 := Variable[T, E](rid, reg.Width(), -1)
				constancy := r_i.Equals(r_im1)
				//
				condition = condition.And(constancy)
			}
		}
	}
	//
	return condition
}

// WithLocalConstancies adds constancy constraints for all registers not
// mutated by a given branch through an instruction.
func (p *Translator[F, T, E, M]) WithLocalConstancies(pc uint, condition E) E {
	if pc > 0 {
		// for i, reg := range p.mapping.Registers {
		// 	rid := p.mapping.Columns[i]
		// 	//
		// 	if p.IsLocalConstancy(uint(i)) {
		// 		r_i := Variable[T, E](rid, reg.Width(), 0)
		// 		r_im1 := Variable[T, E](rid, reg.Width(), -1)
		// 		constancy := r_i.Equals(r_im1)
		// 		//
		// 		condition = condition.And(constancy)
		// 	}
		// }
		//panic("todo")
		fmt.Printf("TODO: implement local consistency")
	}
	//
	return condition
}

// IsLocalConstancy determines whether a given register should be given a
// constancy constraint to ensure its current value matches its previous value.
func (p *Translator[F, T, E, M]) IsLocalConstancy(id uint) bool {
	// r := p.Registers[id]
	// //
	// return !r.IsInput() && !p.constants.Contains(id) &&
	// 	!p.mutated.Contains(id) && !p.mapping.ioLines.Contains(id)
	panic("todo")
}

// StateTranslator packages up key information regarding how an individual state
// of the machine is compiled down to the lower level.
type StateTranslator[F field.Element[F], T any, E Expr[T, E], M Module[F, T, E, M]] struct {
	mapping Translator[F, T, E, M]
	// Set of registers writes on the current branch.
	writes dfa.Writes
}

// WriteRegister constructs a suitable accessors for a register written by a
// given microinstruction.  This activates forwarding for that register for all
// states after this, and returns a suitable expression for the assignment.
func (p *StateTranslator[F, T, E, M]) WriteRegister(dst io.RegisterId) E {
	var (
		ith = p.mapping.Registers[dst.Unwrap()]
		lhs = Variable[T, E](p.mapping.Columns[dst.Unwrap()], ith.Width(), 0)
	)
	//
	return lhs
}

// WriteRegisters constructs suitable accessors for the those registers written
// by a given microinstruction.  This activates forwarding for those registers
// for all states after this, and returns suitable expressions for the
// assignment.
func (p *StateTranslator[F, T, E, M]) WriteRegisters(targets []io.RegisterId) []E {
	lhs := make([]E, len(targets))
	// build up the lhs
	for i, dst := range targets {
		var ith = p.mapping.Registers[dst.Unwrap()]

		lhs[i] = Variable[T, E](p.mapping.Columns[dst.Unwrap()], ith.Width(), 0)
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
		var ith = p.mapping.Registers[dst.Unwrap()]
		//
		lhs[i] = Variable[T, E](p.mapping.Columns[dst.Unwrap()], ith.Width(), 0)
		//
		if i != 0 {
			lhs[i] = BigNumber[T, E](offset).Multiply(lhs[i])
		}
		// left shift offset by given register width.
		offset.Lsh(offset, p.mapping.Registers[dst.Unwrap()].Width())
	}
	//
	return lhs
}

// ReadRegister constructs a suitable accessor for referring to a given register.
// This applies forwarding as appropriate.
func (p *StateTranslator[F, T, E, M]) ReadRegister(regId io.RegisterId) E {
	var (
		reg   = p.mapping.Registers[regId.Unwrap()]
		colId = p.mapping.Columns[regId.Unwrap()]
	)
	//
	if reg.IsInput() {
		// Inputs don't need to refer back
		return Variable[T, E](colId, reg.Width(), 0)
	} else if p.writes.MaybeAssigned(regId) {
		// Forwarded
		return Variable[T, E](colId, reg.Width(), 0)
	}
	// Not forwarded
	return Variable[T, E](colId, reg.Width(), -1)
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
