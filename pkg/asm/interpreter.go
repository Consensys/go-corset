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

	"github.com/consensys/go-corset/pkg/asm/io"
)

// CheckInstance checks whether a given function instance is valid with respect
// to a given set of functions.  It returns an error if something goes wrong
// (e.g. the instance is malformed), and either true or false to indicate
// whether the trace is accepted or not.
func CheckInstance[T io.Instruction[T]](instance io.FunctionInstance, program io.Program[T]) (uint, error) {
	fn := program.Function(instance.Function)
	//
	exec := &SystematicExecutor[T]{program}
	//
	arguments := extractInstanceArguments(instance, fn)
	// Run function using given executor
	outputs := exec.Read(instance.Function, arguments)
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

// Execute a given function with the given arguments in a given I/O environment
// to produce a given set of output values.
func Execute[T io.Instruction[T]](arguments []big.Int, fn io.Function[T], iomap io.Map) []big.Int {
	var code = fn.Code()
	// Construct local state
	state := io.InitialState(arguments, fn, iomap)
	// Keep executing until we're done.
	for state.Pc != io.RETURN {
		insn := code[state.Pc]
		state.Pc = insn.Execute(state)
	}
	// Done
	return state.Outputs()
}

// Trace a given function with the given arguments in a given I/O environment to
// produce a given set of output values, along with the complete set of internal
// traces.
func Trace[T io.Instruction[T]](arguments []big.Int, fn io.Function[T], iomap io.Map) ([]big.Int, []io.State) {
	var (
		code   = fn.Code()
		states []io.State
		// Construct local state
		state = io.InitialState(arguments, fn, iomap)
	)
	// Keep executing until we're done.
	for state.Pc != io.RETURN {
		insn := code[state.Pc]
		// execute given instruction
		pc := insn.Execute(state)
		// record internal state
		states = append(states, state.Clone())
		// advance to next instruction
		state.Pc = pc
	}
	// Done
	return state.Outputs(), states
}

// ============================================================================
// Base Executor
// ============================================================================

// SystematicExecutor executes functions exactly in the order they arise.  For
// example, if a given function is called twice with the same set of arguments,
// then it is executed twice accordingly.
type SystematicExecutor[T io.Instruction[T]] struct {
	program io.Program[T]
}

// Read a set of values at a given address on a bus.
func (p *SystematicExecutor[T]) Read(bus uint, address []big.Int) []big.Int {
	fn := p.program.Function(bus)
	//
	return Execute(address, fn, p)
}

// Write a set of values to a given address on a bus.
func (p *SystematicExecutor[T]) Write(bus uint, address []big.Int, values []big.Int) {
	// Placeholder until mutable bus components are supported.
	panic("cannot write to read-only bus")
}

// ============================================================================
// Trace Executor
// ============================================================================

// TracingExecutor is an executor which additionally traces (i.e. records) the
// internal state of all functions executed.  Thus, it can be used to fill (i.e.
// expand) the trace for a function.
type TracingExecutor[T io.Instruction[T]] struct {
	program io.Program[T]
	//
	traces [][]io.State
}

// NewTracingExecutor constrcts a new tracing executor for a given program.
func NewTracingExecutor[T io.Instruction[T]](program io.Program[T]) *TracingExecutor[T] {
	traces := make([][]io.State, len(program.Functions()))
	return &TracingExecutor[T]{program, traces}
}

// Read a set of values at a given address on a bus.
func (p *TracingExecutor[T]) Read(bus uint, address []big.Int) []big.Int {
	fn := p.program.Function(bus)
	// Trace the given function
	outputs, states := Trace(address, fn, p)
	// Append all generated internal states
	p.traces[bus] = append(p.traces[bus], states...)
	// Done
	return outputs
}

// Write a set of values to a given address on a bus.
func (p *TracingExecutor[T]) Write(bus uint, address []big.Int, values []big.Int) {
	// Placeholder until mutable bus components are supported.
	panic("cannot write to read-only bus")
}

// Traces returns all recorded states for a given bus.
func (p *TracingExecutor[T]) Traces(bus uint) []io.State {
	return p.traces[bus]
}

// ============================================================================
// Helpers
// ============================================================================

func extractInstanceArguments[T io.Instruction[T]](instance io.FunctionInstance, fn io.Function[T]) []big.Int {
	var state = make([]big.Int, len(fn.Registers()))
	// Initialise arguments
	for i, reg := range fn.Registers() {
		if reg.IsInput() {
			var (
				val, ok = instance.Inputs[reg.Name]
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
