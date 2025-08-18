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
package assembler

import (
	"fmt"

	"github.com/consensys/go-corset/pkg/asm/io"
	"github.com/consensys/go-corset/pkg/asm/io/macro"
	"github.com/consensys/go-corset/pkg/asm/io/micro"
	"github.com/consensys/go-corset/pkg/util/collection/bit"
	"github.com/consensys/go-corset/pkg/util/field"
	"github.com/consensys/go-corset/pkg/util/field/bls12_377"
	"github.com/consensys/go-corset/pkg/util/source"
)

// MicroProgram is a program using micro instructions.
type MicroProgram = io.Program[bls12_377.Element, micro.Instruction]

// MacroProgram is a program using macro instructions.
type MacroProgram = io.Program[bls12_377.Element, macro.Instruction]

// Validate checks that a given set of functions are well-formed.  For
// example, an assignment "x,y = z" must be balanced (i.e. number of bits on lhs
// must match number on rhs).  Likewise, registers cannot be used before they
// are defined, and all control-flow paths must reach a "ret" instruction.
// Finally, we cannot assign to an input register under the current calling
// convention.
func Validate(fieldWidth uint, program MacroProgram, srcmaps source.Maps[any]) []source.SyntaxError {
	var errors []source.SyntaxError
	//
	for _, fn := range program.Functions() {
		errors = append(errors, validateInstructions(fieldWidth, *fn, srcmaps)...)
		errors = append(errors, validateControlFlow(*fn, srcmaps)...)
	}
	//
	return errors
}

// ValidateMicro a micro program.  This is more challenging as we have no
// available source mapping information.  Instead, we just panic upon
// encountering an error.
func ValidateMicro(fieldWidth uint, program MicroProgram) {
	var srcmap source.Maps[any]
	//
	for _, fn := range program.Functions() {
		// TODO: support control-flow checks as well.
		validateInstructions(fieldWidth, *fn, srcmap)
	}
}

// Check that each instruction in the function's body is correctly balanced.
// Amongst other things, this means ensuring the right number of bits are used
// on the left-hand side given the right-hand side.  For example, suppose "x :=
// y + 1" where both x and y are byte registers.  This does not balance because
// the right-hand side generates 9 bits but the left-hand side can only consume
// 8bits.
func validateInstructions[F field.Element[F], T io.Instruction[T]](fieldWidth uint, fn io.Function[F, T],
	srcmaps source.Maps[any]) []source.SyntaxError {
	//
	var errors []source.SyntaxError

	for _, insn := range fn.Code() {
		err := insn.Validate(fieldWidth, &fn)
		//
		if err != nil {
			if !srcmaps.Has(insn) {
				panic(err)
			}
			//
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
func validateControlFlow(fn MacroFunction, srcmaps source.Maps[any]) []source.SyntaxError {
	var (
		n          = uint(len(fn.Code()))
		errors     []source.SyntaxError
		entryState bit.Set
	)
	// Initialise entry state (since these are assigned on entry)
	for i, r := range fn.Registers() {
		if !r.IsInput() {
			entryState.Insert(uint(i))
		}
	}
	// Construct the worklist which is the heart of this algorithm.
	worklist := NewWorklist(n, 0, entryState)
	// Continue until all reachable instructions visited
	for !worklist.Empty() {
		// Abstract execute instruction
		errs := applyInstructionSemantics(&worklist, fn, srcmaps)
		// Collect any errors
		errors = append(errors, errs...)
	}
	// Sanity check all instructions reachable.
	for pc, insn := range fn.Code() {
		if !worklist.Visited(uint(pc)) {
			errors = append(errors, *srcmaps.SyntaxError(insn, "unreachable"))
		}
	}
	//
	return errors
}

// Abstractly execute a given vector instruction with respect to a given state
// at the beginning of the instruction.
func applyInstructionSemantics(worklist *Worklist, fn MacroFunction,
	srcmaps source.Maps[any]) []source.SyntaxError {
	//
	var errors []source.SyntaxError
	// Pop the next item from the stack
	pc, state := worklist.Pop()
	insn := fn.CodeAt(pc)
	// Apply effect of instruction on state
	state, errors = applyInstructionFlow(insn, state, fn, srcmaps)
	// Propagate state along branches
	switch insn := insn.(type) {
	case *macro.Goto:
		// Unconditional jump target
		worklist.Join(insn.Target, state)
	case *macro.IfGoto:
		// Conditional jump target
		worklist.Join(insn.Target, state)
		// Fall thru
		worklist.Join(pc+1, state)
	case *macro.Return:
		// Check all outputs are assigned
		errs := checkOutputsAssigned(insn, state, fn, srcmaps)
		errors = append(errors, errs...)
	default:
		// Check not falling off the end
		if pc+1 == uint(len(fn.Code())) {
			errors = append(errors, *srcmaps.SyntaxError(insn, "missing ret"))
		} else {
			// fall through cases
			worklist.Join(pc+1, state)
		}
	}
	//
	return errors
}

// Apply the dataflow transfer function (i.e. the effects of given instruction
// on the record of which registesr are definitely assigned).
func applyInstructionFlow(microinsn macro.Instruction, state bit.Set, fn MacroFunction,
	srcmaps source.Maps[any]) (bit.Set, []source.SyntaxError) {
	//
	var errors []source.SyntaxError
	// Ensure every register read has been defined.
	for _, r := range microinsn.RegistersRead() {
		if state.Contains(r.Unwrap()) {
			msg := fmt.Sprintf("register %s possibly undefined", fn.Register(r).Name)
			errors = append(errors, *srcmaps.SyntaxError(microinsn, msg))
			// mark as defined to avoid follow on errors
			state.Remove(r.Unwrap())
		}
	}
	// Mark all target registers as written.
	for _, r := range microinsn.RegistersWritten() {
		state.Remove(r.Unwrap())
	}
	// Done
	return state, errors
}

// Check that all output registers have been definitely assigned at the point of
// a return.
func checkOutputsAssigned(insn macro.Instruction, state bit.Set, fn MacroFunction,
	srcmaps source.Maps[any]) []source.SyntaxError {
	//
	var errors []source.SyntaxError
	//
	for i, r := range fn.Registers() {
		if r.IsOutput() && state.Contains(uint(i)) {
			msg := fmt.Sprintf("output %s possibly undefined", r.Name)
			errors = append(errors, *srcmaps.SyntaxError(insn, msg))
		}
	}
	//
	return errors
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
