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

	"github.com/consensys/go-corset/pkg/asm/io"
	"github.com/consensys/go-corset/pkg/util/collection/set"
	"github.com/consensys/go-corset/pkg/util/field"
)

// IoState captures the mapping from inputs (i.e. parameters) to outputs (i.e.
// returns) for a particular instance of a given function.
type IoState struct {
	ninputs uint
	state   []big.Int
}

// Cmp comparator for the I/O registers of a particular function instance.
// Observe that, since functions are always deterministic, this only considers
// the inputs (as the outputs follow directly from this).
func (p IoState) Cmp(other IoState) int {
	for i := range p.ninputs {
		if c := p.state[i].Cmp(&other.state[i]); c != 0 {
			return c
		}
	}
	//
	return 0
}

// Outputs returns the output values for this function instance.
func (p IoState) Outputs() []big.Int {
	return p.state[p.ninputs:]
}

// Executor provides a mechanism for executing a program efficiently and
// generating a suitable top-level trace.  Executor implements the io.Map
// interface.
type Executor[F field.Element[F], T io.Instruction[T]] struct {
	program io.Program[F, T]
	states  []set.AnySortedSet[IoState]
}

// NewExecutor constructs a new executor.
func NewExecutor[F field.Element[F], T io.Instruction[T]](program io.Program[F, T]) *Executor[F, T] {
	// Construct initially empty set of states
	states := make([]set.AnySortedSet[IoState], len(program.Functions()))
	// Construct new executor
	return &Executor[F, T]{program, states}
}

// Read implementation for the io.Map interface.
func (p *Executor[F, T]) Read(bus uint, address []big.Int) []big.Int {
	var (
		iostate = IoState{uint(len(address)), address}
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
func (p *Executor[F, T]) Write(bus uint, address []big.Int, values []big.Int) {
	panic("todo")
}

func (p *Executor[F, T]) call(bus uint, inputs []big.Int) []big.Int {
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
	state := io.InitialState(states, fn.Registers(), p)
	// Keep executing until we're done.
	for pc != io.RETURN && pc != io.FAIL {
		insn := fn.CodeAt(pc)
		// execute given instruction
		pc = insn.Execute(state)
		// update state pc
		state.Goto(pc)
	}
	// Cache I/O instance
	instance := IoState{fn.NumInputs(), states[:nio]}
	p.states[bus].Insert(instance)
	// Extract outputs
	return state.Outputs()
}
