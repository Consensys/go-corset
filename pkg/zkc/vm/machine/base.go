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
package machine

import (
	"github.com/consensys/go-corset/pkg/util/collection/stack"
	"github.com/consensys/go-corset/pkg/zkc/vm/fun"
	"github.com/consensys/go-corset/pkg/zkc/vm/memory"
)

// ============================================================================
// Base Machine
// ============================================================================

// Base provides a fundamental implementation of a machine.  The intention is
// that other machine variations would build off this.
type Base[W any, N any, M memory.Memory[W], E Executor[W, N, BaseState[W, N, M]]] struct {
	state    BaseState[W, N, M]
	executor E
}

// New constructs a new empty base machine
func New[W, N any, M memory.Memory[W], E Executor[W, N, BaseState[W, N, M]]]() Base[W, N, M, E] {
	var executor E
	//
	return Base[W, N, M, E]{NewBaseState[W, N, M](), executor}
}

// Boot this machine by starting the given function.
func (p Base[W, N, M, E]) Boot(main uint) *Base[W, N, M, E] {
	var (
		base      = p
		mainFn    = p.state.functions[main]
		bootFrame = NewFrame[W](main, mainFn.Width())
	)
	// Boot the frame
	base.state.callstack.Push(bootFrame)
	// Done
	return &base
}

// WithFunctions returns a base machine updated with the given set of functions,
// but which is otherwise identical to before.
func (p Base[W, N, M, E]) WithFunctions(fns ...fun.Function[N]) Base[W, N, M, E] {
	var base = p
	//
	base.state.functions = fns
	//
	return base
}

// WithInputs returns a base machine updated with the given set of inputs but
// which is otherwise identical to before.
func (p Base[W, N, M, E]) WithInputs(inputs ...M) Base[W, N, M, E] {
	var base = p
	//
	base.state.inputs = inputs
	//
	return base
}

// WithOutputs returns a base machine updated with the given set of outputs but
// which is otherwise identical to before.
func (p Base[W, N, M, E]) WithOutputs(outputs ...M) Base[W, N, M, E] {
	var base = p
	//
	base.state.outputs = outputs
	//
	return base
}

// WithMemories returns a base machine updated with the given set of random
// access memories but which is otherwise identical to before.
func (p Base[W, N, M, E]) WithMemories(rams ...M) Base[W, N, M, E] {
	var base = p
	//
	base.state.rams = rams
	//
	return base
}

// WithStatics returns a base machine updated with the given set of inputs but
// which is otherwise identical to before.
func (p Base[W, N, M, E]) WithStatics(statics ...M) Base[W, N, M, E] {
	var base = p
	//
	base.state.statics = statics
	//
	return base
}

// Execute the machine for the given number of steps, returning the actual
// number of steps executed and an error (if execution failed).
func (p *Base[W, N, M, E]) Execute(steps uint) (uint, error) {
	var (
		nsteps uint
		err    error
	)
	//
	for !p.state.callstack.IsEmpty() {
		if p.state, err = p.executor.Execute(p.state); err != nil {
			return nsteps, err
		}
		//
		nsteps++
	}
	//
	return nsteps, nil
}

// State implementation for the Machine interface.
func (p *Base[W, N, M, E]) State() State[W, N] {
	return p.state
}

// ============================================================================
// Base State
// ============================================================================

// BaseState provides the base implementation of the StaticState
// interface.
type BaseState[W any, N any, M memory.Memory[W]] struct {
	functions []fun.Function[N]
	statics   []M
	inputs    []M
	outputs   []M
	rams      []M
	callstack *stack.Stack[Frame[W]]
}

// NewBaseState creates an empty base state.
func NewBaseState[W any, N any, M memory.Memory[W]]() BaseState[W, N, M] {
	return BaseState[W, N, M]{
		functions: nil,
		statics:   nil,
		inputs:    nil,
		outputs:   nil,
		rams:      nil,
		callstack: stack.NewStack[Frame[W]](),
	}
}

// ========================================================
// Static State
// ========================================================

// Function implementation of StaticState interface
func (p BaseState[W, N, M]) Function(id uint) fun.Function[N] {
	return p.functions[id]
}

// NumFunctions implementation of StaticState interface
func (p BaseState[W, N, M]) NumFunctions() uint {
	return uint(len(p.functions))
}

// NumStatics implementation of StaticState interface
func (p BaseState[W, N, M]) NumStatics() uint {
	return uint(len(p.statics))
}

// Static implementation of StaticState interface
func (p BaseState[W, N, M]) Static(id uint) memory.ReadOnlyMemory[W] {
	return p.statics[id]
}

// ========================================================
// Dynamic State
// ========================================================

// CallStack implementation of DynamicState interface
func (p BaseState[W, N, M]) CallStack() *stack.Stack[Frame[W]] {
	return p.callstack
}

// Input implementation of DynamicState interface
func (p BaseState[W, N, M]) Input(id uint) memory.ReadOnlyMemory[W] {
	return p.inputs[id]
}

// Output implementation of DynamicState interface
func (p BaseState[W, N, M]) Output(id uint) memory.WriteOnceMemory[W] {
	return p.outputs[id]
}

// Memory implementation of DynamicState interface
func (p BaseState[W, N, M]) Memory(id uint) memory.Memory[W] {
	return p.rams[id]
}

// NumInputs implementation of DynamicState interface
func (p BaseState[W, N, M]) NumInputs() uint {
	return uint(len(p.inputs))
}

// NumOutputs implementation of DynamicState interface
func (p BaseState[W, N, M]) NumOutputs() uint {
	return uint(len(p.outputs))
}

// NumMemories implementation of DynamicState interface
func (p BaseState[W, N, M]) NumMemories() uint {
	return uint(len(p.rams))
}
