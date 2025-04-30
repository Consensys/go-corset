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
package asm

import (
	"fmt"

	"github.com/consensys/go-corset/pkg/asm/insn"
	"github.com/consensys/go-corset/pkg/util/collection/bit"
	"github.com/consensys/go-corset/pkg/util/source"
)

// Instruction is an alias for instruction.Instruction
type Instruction = insn.Instruction

// Register describes a single register within a function.
type Register = insn.Register

// Assemble takes a given set of assembly files, and parses them into a given
// set of functions.  This includes performing various checks on the files, such
// as type checking, etc.
func Assemble(assembly ...source.File) ([]Function, source.Maps[Instruction], []source.SyntaxError) {
	var (
		fns     []Function
		errors  []source.SyntaxError
		srcmaps source.Maps[Instruction] = *source.NewSourceMaps[Instruction]()
	)
	// Parse each file in turn.
	for _, asm := range assembly {
		fn, srcmap, errs := Parse(&asm)
		if len(errs) == 0 {
			fns = append(fns, fn...)
		}
		// Join srcmap
		srcmaps.Join(srcmap)
		//
		errors = append(errors, errs...)
	}
	// Well-formedness checks
	for _, fn := range fns {
		errors = append(errors, checkWellFormed(fn, srcmaps)...)
	}
	// Done
	return fns, srcmaps, errors
}

// check that a given set of functions are well-formed.  For example, an
// assignment "x,y = z" must be balanced (i.e. number of bits on lhs must match
// number on rhs).  Likewise, registers cannot be used before they are defined,
// and all control-flow paths must reach a "ret" instruction.  Finally, we
// cannot assign to an input register under the current calling convention.
func checkWellFormed(fn Function, srcmaps source.Maps[Instruction]) []source.SyntaxError {
	balance_errs := checkInstructionsBalance(fn, srcmaps)
	flow_errs := checkInstructionsFlow(fn, srcmaps)
	//
	return append(balance_errs, flow_errs...)
}

// Check that each instruction in the function's body is correctly balanced.
// Amongst other things, this means ensuring the right number of bits are used
// on the left-hand side given the right-hand side.  For example, suppose "x :=
// y + 1" where both x and y are byte registers.  This does not balance because
// the right-hand side generates 9 bits but the left-hand side can only consume
// 8bits.
func checkInstructionsBalance(fn Function, srcmaps source.Maps[Instruction]) []source.SyntaxError {
	var errors []source.SyntaxError

	for _, insn := range fn.Code {
		err := insn.IsWellFormed(fn.Registers)
		//
		if err != nil {
			errors = append(errors, *srcmaps.SyntaxError(insn, err.Error()))
		}
	}
	//
	return errors
}

// Check for issues related to the control-flow of a function.  For example,
// where a register is not definitely assigned before being used.  Or, a
// control-flow path exists which is not terminated with ret.  This is
// implemented using a straightforward dataflow analysis.  One aspect worth
// noting is that the dataflow sets hold true for registers which are undefined,
// and false for registers which are defined.
func checkInstructionsFlow(fn Function, srcmaps source.Maps[Instruction]) []source.SyntaxError {
	var (
		n          = uint(len(fn.Code))
		errors     []source.SyntaxError
		entryState bit.Set
	)
	// Initialise entry state (since these are assigned on entry)
	for i, r := range fn.Registers {
		if r.Kind != INPUT_REGISTER {
			entryState.Insert(uint(i))
		}
	}
	// Construct the worklist which is the heart of this algorithm.
	worklist := NewWorklist(n, 0, entryState)
	// Continue until all reachable instructions visited
	for !worklist.Empty() {
		// Pop the next item from the stack
		pc, state := worklist.Pop()
		//
		nstate, errs := applyInstructionFlow(pc, state, fn, srcmaps)
		errors = append(errors, errs...)
		// Update control flow
		switch insn := fn.Code[pc].(type) {
		case *insn.Jmp:
			// Unconditional jump target
			worklist.Join(insn.Target, nstate)
		case *insn.Jznz:
			// Conditional jump target
			worklist.Join(insn.Target, nstate)
			// Unconditional jump target
			if pc+1 < n {
				worklist.Join(pc+1, nstate)
			} else {
				errors = append(errors, *srcmaps.SyntaxError(fn.Code[pc], "missing ret"))
			}
		case *insn.Ret:
			// Terminate
			errors = append(errors, checkOutputsAssigned(pc, state, fn, srcmaps)...)
		default:
			// All others fall through
			if pc+1 < n {
				worklist.Join(pc+1, nstate)
			} else {
				errors = append(errors, *srcmaps.SyntaxError(fn.Code[pc], "missing ret"))
			}
		}
	}
	// Sanity check all instructions reachable.
	for pc := 0; pc < len(fn.Code); pc++ {
		if !worklist.Visited(uint(pc)) {
			errors = append(errors, *srcmaps.SyntaxError(fn.Code[pc], "unreachable"))
		}
	}
	//
	return errors
}

// Check that all output registers have been definitely assigned at the point of
// a return.
func checkOutputsAssigned(pc uint, state bit.Set, fn Function, srcmaps source.Maps[Instruction]) []source.SyntaxError {
	var errors []source.SyntaxError
	//
	for i, r := range fn.Registers {
		if r.Kind == OUTPUT_REGISTER && state.Contains(uint(i)) {
			msg := fmt.Sprintf("output %s possibly undefined", r.Name)
			errors = append(errors, *srcmaps.SyntaxError(fn.Code[pc], msg))
		}
	}
	//
	return errors
}

// Apply the dataflow transfer function (i.e. the effects of given instruction
// on the record of which registesr are definitely assigned).
func applyInstructionFlow(pc uint, regs bit.Set, fn Function,
	srcmaps source.Maps[Instruction]) (bit.Set, []source.SyntaxError) {
	//
	var (
		errors []source.SyntaxError
		insn   = fn.Code[pc]
	)
	// Ensure every register read has been defined.
	for _, r := range insn.RegistersRead() {
		if regs.Contains(r) {
			msg := fmt.Sprintf("register %s possibly undefined", fn.Registers[r].Name)
			errors = append(errors, *srcmaps.SyntaxError(insn, msg))
			// mark as defined to avoid follow on errors
			regs.Remove(r)
		}
	}
	// Mark all target registers as written.
	for _, r := range insn.RegistersWritten() {
		regs.Remove(r)
	}
	// Done
	return regs, errors
}

// Worklist encapsulates the notion of a worklist, along with the necessary
// dataflow sets for a dataflow analysis algorithm.
type Worklist struct {
	// Visited is used to determine which instruct
	visited bit.Set
	states  []bit.Set
	stack   []uint
}

// NewWorklist constructs a new worklist of a given capacity.
func NewWorklist(nstates uint, start uint, init bit.Set) Worklist {
	var visited bit.Set
	//
	states := make([]bit.Set, nstates)
	states[start] = init
	// mark start visited
	visited.Insert(start)
	//
	return Worklist{
		visited,
		states,
		[]uint{start},
	}
}

// Empty determines whether or not this worklist is empty.
func (p *Worklist) Empty() bool {
	return len(p.stack) == 0
}

// Visited checks whether a given program point was reached during the analysis.
func (p *Worklist) Visited(pc uint) bool {
	return p.visited.Contains(pc)
}

// Pop removes the next item from the stack, and also returns the relevant
// dataflow state.
func (p *Worklist) Pop() (uint, bit.Set) {
	n := len(p.stack) - 1
	pc := p.stack[n]
	bs := p.states[pc]
	p.stack = p.stack[:n]
	//
	return pc, bs.Clone()
}

// Join joins a given state into the state recorded for a given pc location.
func (p *Worklist) Join(pc uint, state bit.Set) {
	pcState := &p.states[pc]
	// Visit state if it hasn't been visited before, or there is an update to
	// its dataflow state.
	if pcState.Union(state) || !p.visited.Contains(pc) {
		// mark item as visited
		p.visited.Insert(pc)
		// push item on stack
		p.stack = append(p.stack, pc)
	}
}
