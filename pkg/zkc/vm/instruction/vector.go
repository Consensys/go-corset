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
	"github.com/consensys/go-corset/pkg/util/field"
	"github.com/consensys/go-corset/pkg/zkc/vm/instruction/opcode"
)

// Vector instructions are instructions composed of some number of micro
// instructions which, with restrictions, can be executed by the underlying
// machine "in parallel".  The approach is analoguous to the concept of
// "Very-Long Instruction Words (VLIW)" but taken to more of an extreme ---
// there is no limit on the number of micro-instructions.
//
// To better understand vector instructions, consider two instructions executed
// in sequence (the at pc location 0, the second at pc location 1):
//
// (pc=0) x = y + 1 (pc=1) z = 0
//
// When executing these instructions, there is an intermediate state after the
// first instruction is executed but before the second has been where x has been
// written but z has not.  Alternatively, the two instructions can be composed
// together to form a vector instruction, written like so:
//
// (pc=0) x = y + 1 ; z = 0
//
// In this case, both instructions are executed together and there is no
// intermediate state where x is written but z is not.
//
// To ensure easy translation into polynomial constraints, there are
// restrictions on how vector instructions can be composed.  In particular, no
// variable can be assigned twice on the same execution path.  Thus, for
// example, this is an invalid vector instruction:
//
// (pc=0) x = 0 ; x = 1
//
// These writes are said to be _conflicting_.  In contrast, the following is a
// valid vector instruction:
//
// (pc=0) skip_if x != y 2 ; r = 0 ; ret ; r = 1 ; ret
//
// In this case, whilst there are two assignments to register r, neither are on
// the same path.  These writes are said to be _non-conflicting_.  Finally, we
// should note that register forwarding is applied within vector instructions.
// Thus, for example, the following is allowed:
//
// (pc=0) x = 0; y = x + 1; ret
//
// Here, the value of x written in the instruction is "forwarded" to the
// assignment for y.  This process is, roughly speaking, analoguous to register
// forwarding as found in CPU architectures.
type Vector[I Instruction] struct {
	Codes []I
}

// NewVector constructs a new vector instruction composed of zero or more
// micro-instructions.  Observe that an empty vector instruction is a no-op.
func NewVector[I Instruction](insns ...I) Vector[I] {
	return Vector[I]{insns}
}

// IsEmpty simply identifies whether this instruction is a no-op (or not).
func (p *Vector[W]) IsEmpty() bool {
	return len(p.Codes) == 0
}

// Validate that this micro-instruction is well-formed.  For example, each
// micro-instruction contained within must be well-formed, and the overall
// requirements for a vector instruction must be met, etc.
func (p *Vector[W]) Validate(field field.Config, mapping SystemMap) []error {
	// Construct write map
	var (
		errors   []error
		nCodes   = uint(len(p.Codes))
		writeMap = p.WriteMap()
	)
	// Validate individual instructions
	for _, r := range p.Codes {
		errs := r.MicroValidate(nCodes, field, mapping)
		errors = append(errors, errs...)
	}
	// Validate no Read/Write conflicts
	for i := range nCodes {
		var (
			ithState = writeMap.StateOf(i)
			ith      = p.Codes[i]
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
func (p *Vector[W]) String(mapping SystemMap) string {
	var builder strings.Builder
	//
	for i, code := range p.Codes {
		if i != 0 {
			builder.WriteString(" ; ")
		}
		//
		builder.WriteString(code.String(mapping))
	}
	//
	return builder.String()
}

// WriteMap constructs the write map for this vector instruction.
//
// For each instruction, the write map records — on entry to that instruction —
// which registers have been written by preceding instructions (on any path to
// this point). This identifies: (1) whether a register _may_ have been written
// on some path; (2) or, whether it was _definitely_ written along all paths.
// For example, consider the following sequence:
//
// x = 0; skip_if ... 1; y = 0; ret
//
// When execution reaches the return instruction, we know that x was definitely
// written but only that y may have been written (i.e. depending on which path
// was taken).
//
// The write map serves two purposes:  firstly, it allows conflict detection;
// secondly, it identifies where register forwarding should be used.  A write
// conflict arises when a register is written which _may_ have already been
// written; likewise a read conflict arises when a register is read that _may_
// (but not _definitely_) have been written.  Finally, register forwarding
// arises when a register has _definitely_ been written by an earlier
// instruction in the vector and, hence, subsequent reads use the new value
// (rather than the previous value).
func (p *Vector[W]) WriteMap() dfa.Result[dfa.Writes] {
	return dfa.Construct(dfa.Writes{}, p.Codes, writeDfaTransfer)
}

// BranchTable returns a _branch condition_ for each instruction in the vector
// which identifies the conditions under which the given instruction will
// execute.  For example, consider the following sequence:
//
// skip_if x!=0 1; y=0; skip_if x!=1 2; y=1; ret; y = 2; ret
// --------------+----+---------------+----+----+------+----
// 0             | 1  | 2             | 3  | 4  | 5    | 6
//
// This sequence gives rise to the following branch table:
//
// --+-------------+-----------------------
// 0 | skip_if ... | TRUE
// 1 | y=0         | x==0
// 2 | skip_if ... | x!=0
// 3 | y=1         | x!=0 && x==1 ==> x==1
// 4 | ret         | x!=0 && x==1 ==> x==1
// 5 | y=2         | x!=0 && x!=1
// 6 | ret         | x!=0 && x!=1
// --+-------------+-----------------------
//
// Observe that the optimiser automatically reduces "x!=0 && x==1" to just x==1
// (this is why it is sometimes called _branch table optimisation_).
func (p *Vector[W]) BranchTable() dfa.Result[dfa.Branch] {
	// Construct suitable branch table for this instruction vector.
	var btf = branchTableTransfer[W](p.WriteMap())
	//
	return dfa.Construct(dfa.Branch{Condition: dfa.TRUE}, p.Codes, btf)
}

func branchTableTransfer[I Instruction](writeMap dfa.Result[dfa.Writes]) dfa.BranchTransferFunction[I] {
	panic("todo")
}

// Data-flow transfer function for the writes analysis
func writeDfaTransfer[I Instruction](offset uint, code I, state dfa.Writes) []dfa.Transfer[dfa.Writes] {
	//
	var (
		arcs []dfa.Transfer[dfa.Writes]
		insn Instruction = code
	)
	//
	switch code.OpCode() {
	case opcode.FAIL, opcode.RETURN, opcode.JUMP:
		return nil
	case opcode.SKIP:
		code := insn.(*Skip)
		// join into branch target
		return append(arcs, dfa.NewTransfer(state, offset+code.Skip+1))
	case opcode.SKIP_IF:
		code := insn.(*SkipIf)
		// join into branch target
		arcs = append(arcs, dfa.NewTransfer(state, offset+code.Skip+1))
		// fall through
	}
	// Construct state after this code
	nState := state.Write(code.Definitions()...)
	// Transfer to following instruction
	arcs = append(arcs, dfa.NewTransfer(nState, offset+1))
	// Done
	return arcs
}
