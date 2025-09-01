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
package io

import (
	"math"
	"math/big"

	"github.com/consensys/go-corset/pkg/util/collection/set"
)

// FunctionInstance captures the mapping from inputs (i.e. parameters) to outputs (i.e.
// returns) for a particular instance of a given function.
type FunctionInstance struct {
	ninputs uint
	state   []big.Int
}

// Cmp comparator for the I/O registers of a particular function instance.
// Observe that, since functions are always deterministic, this only considers
// the inputs (as the outputs follow directly from this).
func (p FunctionInstance) Cmp(other FunctionInstance) int {
	for i := range p.ninputs {
		if c := p.state[i].Cmp(&other.state[i]); c != 0 {
			return c
		}
	}
	//
	return 0
}

// Outputs returns the output values for this function instance.
func (p FunctionInstance) Outputs() []big.Int {
	return p.state[p.ninputs:]
}

// Executor provides a mechanism for executing a program efficiently and
// generating a suitable top-level trace.  Executor implements the io.Map
// interface.
type Executor[T Instruction[T]] struct {
	program Program[T]
	states  []set.AnySortedSet[FunctionInstance]
}

// NewExecutor constructs a new executor.
func NewExecutor[T Instruction[T]](program Program[T]) *Executor[T] {
	// Construct initially empty set of states
	states := make([]set.AnySortedSet[FunctionInstance], len(program.Functions()))
	// Construct new executor
	return &Executor[T]{program, states}
}

// Read implementation for the io.Map interface.
func (p *Executor[T]) Read(bus uint, address []big.Int) []big.Int {
	var (
		iostate = FunctionInstance{uint(len(address)), address}
		states  = p.states[bus]
	)
	// Check whether this instance has already been computed.
	if index := states.Find(iostate); index != math.MaxUint {
		// Yes, therefore return precomputed outputs
		return states[index].Outputs()
	}
	// Execute function to determine new outputs.
	return p.call(bus, address)
}

// Write implementation for the io.Map interface.
func (p *Executor[T]) Write(bus uint, address []big.Int, values []big.Int) {
	panic("todo")
}

func (p *Executor[T]) call(bus uint, inputs []big.Int) []big.Int {
	var (
		fn = p.program.Function(bus)
		// Determine how many I/O registers
		nio = fn.NumInputs() + fn.NumOutputs()
		//
		pc = uint(0)
		//
		states = make([]big.Int, len(fn.Registers()))
	)
	// Initialise input arguments
	copy(states, inputs)
	// Construct initial state
	state := InitialState(states, fn.Registers(), p)
	// Keep executing until we're done.
	for pc != RETURN && pc != FAIL {
		insn := fn.CodeAt(pc)
		// execute given instruction
		pc = insn.Execute(state)
		// update state pc
		state.Goto(pc)
	}
	// Cache I/O instance
	instance := FunctionInstance{fn.NumInputs(), states[:nio]}
	p.states[bus].Insert(instance)
	// Extract outputs
	return state.Outputs()
}
