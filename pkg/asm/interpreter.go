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
	"math"
	"math/big"

	"github.com/consensys/go-corset/pkg/asm/insn"
)

// InterpreterState represents the state of a function being executed by the interpreter.
type InterpreterState struct {
	// Function index determines which function this state corresponds with.
	fid uint
	// Program counter
	pc uint
	// Current register values
	registers []big.Int
}

// Registers returns the current state of all registers.
func (p *InterpreterState) Registers() []big.Int {
	return p.registers
}

// PC returns the current Program Counter position.
func (p *InterpreterState) PC() uint {
	return p.pc
}

// Interpreter encapsulates all state needed for executing a given instruction
// sequence.
type Interpreter[T insn.Instruction] struct {
	// Set of functions being interpreted
	functions []Function[T]
	// Set of interpreter states
	states []InterpreterState
}

// NewInterpreter intialises an interpreter for executing a given instruction
// sequence.
func NewInterpreter[T insn.Instruction](fns ...Function[T]) *Interpreter[T] {
	return &Interpreter[T]{fns, nil}
}

// Bind converts a set of name inputs into the internal state as needed by the
// interpreter.
func (p *Interpreter[T]) Bind(fn uint, arguments map[string]big.Int) []big.Int {
	var (
		f     = p.functions[fn]
		state = make([]big.Int, len(f.Registers))
	)
	// Initialise arguments
	for i, reg := range f.Registers {
		if reg.IsInput() {
			var (
				val = arguments[reg.Name]
				ith big.Int
			)
			// Clone big int
			ith.Set(&val)
			//
			state[i] = ith
		}
	}
	//
	return state
}

// State returns the interpreter's (raw) register state for the currently
// executing function.  This state is raw, hence changes to this can impact the
// interpreter's subsequent execution.
func (p *Interpreter[T]) State() InterpreterState {
	var n = len(p.states) - 1
	return p.states[n]
}

// Enter beings execution of a given function, using a given initial state.  The
// currently executing function (if any) is paused.
func (p *Interpreter[T]) Enter(fn uint, state []big.Int) {
	//
	p.states = append(p.states, InterpreterState{
		fn, uint(0), state,
	})
}

// Leave exits the currently executing function, extracting its output values.
func (p *Interpreter[T]) Leave() map[string]big.Int {
	var (
		n  = len(p.states) - 1
		st = p.states[n]
		f  = p.functions[st.fid]
	)
	// Construct outputs
	outputs := make(map[string]big.Int, 0)
	//
	for i, reg := range f.Registers {
		if reg.IsOutput() {
			outputs[reg.Name] = st.registers[i]
		}
	}
	// Remove last state
	p.states = p.states[:n]
	//
	return outputs
}

// Execute n steps of the given program, returning the number of steps actually
// executed.  The number of steps can differ from that requested if: the
// enclosing function has already terminated; the enclosing function terminates
// before executing all steps.
func (p *Interpreter[T]) Execute(nsteps uint) uint {
	var (
		n    = len(p.states) - 1
		st   = &p.states[n]
		f    = p.functions[st.fid]
		step = uint(0)
	)
	//
	for st.pc != math.MaxUint && step < nsteps {
		st.pc = execute(st.pc, st.registers, f)
		step++
	}
	//
	return step
}

// HasTerminated checks whether or not the enclosing function has terminated.
func (p *Interpreter[T]) HasTerminated() bool {
	var (
		n  = len(p.states) - 1
		st = p.states[n]
	)
	//
	return st.pc == math.MaxUint
}

func execute[T insn.Instruction](pc uint, state []big.Int, f Function[T]) uint {
	npc := f.Code[pc].Execute(state, f.Registers)
	// Handle return values
	if npc == insn.FALL_THRU {
		return pc + 1
	}
	//
	return npc
}
