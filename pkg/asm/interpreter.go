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
	"math"
	"math/big"
	"strings"

	"github.com/consensys/go-corset/pkg/asm/io"
)

// CheckInstance checks whether a given function instance is valid with respect
// to a given set of functions.  It returns an error if something goes wrong
// (e.g. the instance is malformed), and either true or false to indicate
// whether the trace is accepted or not.
func CheckInstance[T io.Instruction[T]](instance io.FunctionInstance, program io.Program[T]) (uint, error) {
	fn := program.Function(instance.Function)
	// Initialise a new interpreter
	interpreter := NewInterpreter(program)
	//
	init := interpreter.Bind(instance.Function, instance.Inputs)
	// Enter function
	interpreter.Enter(instance.Function, init)
	// Execute function to completion
	interpreter.Execute(math.MaxUint)
	// Extract outputs
	outputs := interpreter.Leave()
	// Check results
	for i, reg := range fn.Outputs() {
		expected, ok := instance.Outputs[reg.Name]
		actual := outputs[i]
		// Check actual output matches expected output
		if !ok {
			return math.MaxUint, fmt.Errorf("missing output (%s)", reg.Name)
		} else if expected.Cmp(&actual) != 0 {
			// failure
			return 1, fmt.Errorf("incorrect output \"%s\" (was %s, expected %s)", reg.Name, actual.String(), expected.String())
		}
	}
	//
	if len(outputs) != len(instance.Outputs) {
		msg := fmt.Errorf("incorrect number of outputs (was %d but expected %d)", len(outputs), len(instance.Outputs))
		return math.MaxUint, msg
	}
	// Success
	return 0, nil
}

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
type Interpreter[T io.Instruction[T]] struct {
	// Program being interpreted
	program io.Program[T]
	// Set of interpreter states
	states []InterpreterState
}

// NewInterpreter intialises an interpreter for executing a given instruction
// sequence.
func NewInterpreter[T io.Instruction[T]](program io.Program[T]) *Interpreter[T] {
	return &Interpreter[T]{program, nil}
}

// Bind converts a set of named inputs into the internal state as needed by the
// interpreter.
func (p *Interpreter[T]) Bind(fn uint, arguments map[string]big.Int) []big.Int {
	var (
		f     = p.program.Function(fn)
		state = make([]big.Int, len(f.Registers))
	)
	// Initialise arguments
	for i, reg := range f.Registers {
		if reg.IsInput() {
			var (
				val, ok = arguments[reg.Name]
				ith     big.Int
			)
			// Sanity check
			if !ok {
				panic(fmt.Sprintf("missing value for input register %s", reg.Name))
			}
			// Clone big int
			ith.Set(&val)
			//
			state[i] = ith
		}
	}
	//
	return state
}

// BindInner binds input arguments represented as an array into a true internal
// state.
func (p *Interpreter[T]) BindInner(fn uint, arguments []big.Int) []big.Int {
	var (
		f     = p.program.Function(fn)
		state = make([]big.Int, len(f.Registers))
		index = 0
	)
	// Initialise arguments
	for i, reg := range f.Registers {
		if reg.IsInput() {
			var (
				val = arguments[index]
				ith big.Int
			)
			// Clone big int
			ith.Set(&val)
			//
			state[i] = ith
			index = index + 1
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
func (p *Interpreter[T]) Leave() []big.Int {
	var (
		n  = len(p.states) - 1
		st = p.states[n]
		f  = p.program.Function(st.fid)
	)
	// Construct outputs
	outputs := make([]big.Int, 0)
	//
	for i, reg := range f.Registers {
		if reg.IsOutput() {
			outputs = append(outputs, st.registers[i])
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
		f    = p.program.Function(st.fid)
		step = uint(0)
	)
	// Construct local state
	state := io.State{Pc: st.pc, State: st.registers, Registers: f.Registers}
	//
	for state.Pc != io.RETURN && step < nsteps {
		insn := f.Code[state.Pc]
		state.Pc = insn.Execute(state, p)
		step++
	}
	// Write back local state
	st.pc = state.Pc
	// Done
	return step
}

// HasTerminated checks whether or not the enclosing function has terminated.
func (p *Interpreter[T]) HasTerminated() bool {
	return p.State().pc == io.RETURN
}

// Read a set of values at a given address on a bus.
func (p *Interpreter[T]) Read(bus uint, address []big.Int) []big.Int {
	// Convert address into true internal state
	state := p.BindInner(bus, address)
	// Enter function
	p.Enter(bus, state)
	// Execute function to completion
	p.Execute(math.MaxUint)
	// Extract outputs
	return p.Leave()
}

// Write a set of values to a given address on a bus.
func (p *Interpreter[T]) Write(bus uint, address []big.Int, values []big.Int) {
	panic("to do")
}

func (p *Interpreter[T]) String() string {
	var (
		builder strings.Builder
		state   = p.State()
		fn      = p.program.Function(state.fid)
	)
	//
	for i := 1; i < len(p.states); i++ {
		builder.WriteString("\t")
	}
	//
	if p.State().pc == math.MaxUint {
		builder.WriteString("------- ")
	} else {
		pc := fmt.Sprintf("(pc=%02x) ", p.State().pc)
		builder.WriteString(pc)
	}
	//
	for i := 0; i != len(fn.Registers); i++ {
		if i != 0 {
			builder.WriteString(", ")
		}
		//
		val := state.registers[i].Text(16)
		builder.WriteString(fmt.Sprintf("%s=0x%s", fn.Registers[i].Name, val))
	}
	//
	return builder.String()
}
