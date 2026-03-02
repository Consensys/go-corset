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
	"fmt"
	"strings"

	"github.com/consensys/go-corset/pkg/asm/io/micro/dfa"
	"github.com/consensys/go-corset/pkg/schema/register"
	"github.com/consensys/go-corset/pkg/util/collection/array"
	"github.com/consensys/go-corset/pkg/util/collection/bit"
	"github.com/consensys/go-corset/pkg/util/field"
	"github.com/consensys/go-corset/pkg/zkc/vm/word"
)

// Vector instructions are instructions which are composed of two or more micro
// instructions.   The intuition here is that micro instructions can be executed
// by the underlying machine "in parallel".  This is analoguous to the concept
// of "Very-Long Instruction Words (VLIW)" but takes this to an extreme ---
// there is not bound on the width of a vector instruction.
type Vector[W word.Word[W]] struct {
	codes []MicroInstruction[W]
}

// MicroInstruction characterises the kinds of instructions which can be
// vectorized.  They key is that, whilst many instructions are also micro
// instructions, this is not always the case.  Specifically, there are
// instructions which are not valid micro-instructions and, likewise,
// micro-instructions which are not valid instructions.
type MicroInstruction[W word.Word[W]] interface {
	// Uses returns the set of variables used (i.e. read) by this instruction.
	Uses() []register.Id
	// Definitions returns the set of variables registers defined (i.e. written)
	// by this instruction.
	Definitions() []register.Id
	// Validate that this micro-instruction is well-formed.  For example, that
	// it is balanced, that there are no conflicting writes, that all
	// temporaries have been allocated, etc.
	MicroValidate(width uint, field field.Config, env register.Map) []error
	// Provide human readable form of instruction
	String(env register.Map) string
}

// NewVector constructs a new vector instruction composed of zero or more
// micro-instructions.  Observe that an empty vector instruction is a no-op.
func NewVector[W word.Word[W], I MicroInstruction[W]](insns ...I) *Vector[W] {
	// Map array of I to array of MicroInstruction
	array := array.Map(insns, func(_ uint, insn I) MicroInstruction[W] { return insn })
	//
	return &Vector[W]{array}
}

// Uses implementation for Instruction interface
func (p *Vector[W]) Uses() []register.Id {
	var (
		regs bit.Set
		read []register.Id
	)
	//
	for _, c := range p.codes {
		for _, id := range c.Uses() {
			if !regs.Contains(id.Unwrap()) {
				regs.Insert(id.Unwrap())
				read = append(read, id)
			}
		}
	}
	//
	return read
}

// Definitions implementation for Instruction interface
func (p *Vector[W]) Definitions() []register.Id {
	var (
		regs    bit.Set
		written []register.Id
	)
	//
	for _, c := range p.codes {
		for _, id := range c.Definitions() {
			if !regs.Contains(id.Unwrap()) {
				regs.Insert(id.Unwrap())
				written = append(written, id)
			}
		}
	}
	//
	return written
}

// Validate that this micro-instruction is well-formed.  For example, each
// micro-instruction contained within must be well-formed, and the overall
// requirements for a vector instruction must be met, etc.
func (p *Vector[W]) Validate(field field.Config, mapping register.Map) []error {
	// Construct write map
	var (
		errors   []error
		nCodes   = uint(len(p.codes))
		writeMap = p.Writes()
	)
	// Validate individual instructions
	for _, r := range p.codes {
		errs := r.MicroValidate(nCodes, field, mapping)
		errors = append(errors, errs...)
	}
	// Validate no Read/Write conflicts
	for i := range nCodes {
		var (
			ithState = writeMap.StateOf(i)
			ith      = p.codes[i]
		)
		// Sanity check for conflicting reads
		for _, r := range ith.Uses() {
			if ithState.MaybeAssigned(r) && !ithState.DefinitelyAssigned(r) {
				name := mapping.Register(r).Name()
				errors = append(errors,
					fmt.Errorf("conflicting read on register \"%s\" in \"%s\"", name, ith.String(mapping)))
			}
		}
		// Sanity check conflicting writes
		for _, r := range ith.Definitions() {
			if ithState.MaybeAssigned(r) {
				name := mapping.Register(r).Name()
				errors = append(errors,
					fmt.Errorf("conflicting write on register \"%s\" in \"%s\"", name, ith.String(mapping)))
			}
		}
	}
	// Done
	return errors
}

// String implementation for Instruction interface
func (p *Vector[W]) String(env register.Map) string {
	var builder strings.Builder
	//
	for i, code := range p.codes {
		if i != 0 {
			builder.WriteString(" ; ")
		}
		//
		builder.WriteString(code.String(env))
	}
	//
	return builder.String()
}

// Writes constructs the write map for this micro instruction.
func (p *Vector[W]) Writes() dfa.Result[dfa.Writes] {
	return dfa.Construct(dfa.Writes{}, p.codes, writeDfaTransfer[W])
}

// Data-flow transfer function for the writes analysis
func writeDfaTransfer[W word.Word[W]](offset uint, code MicroInstruction[W], state dfa.Writes,
) []dfa.Transfer[dfa.Writes] {
	//
	var arcs []dfa.Transfer[dfa.Writes]
	//
	switch code := code.(type) {
	case *Fail, *Return, *Jmp:
		return nil
	case *Skip:
		// join into branch target
		return append(arcs, dfa.NewTransfer(state, offset+code.Skip+1))
	case *SkipIf:
		// join into branch target
		arcs = append(arcs, dfa.NewTransfer(state, offset+code.Skip+1))
		// fall through
	}
	// Construct state after this code
	nState := state.Write(code.Uses()...)
	// Transfer to following instruction
	arcs = append(arcs, dfa.NewTransfer(nState, offset+1))
	// Done
	return arcs
}
