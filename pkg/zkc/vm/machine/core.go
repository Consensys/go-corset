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
	"github.com/consensys/go-corset/pkg/zkc/vm/function"
	"github.com/consensys/go-corset/pkg/zkc/vm/memory"
)

// ExecuteAll executes a given machine to completion in chunks of n steps,
// returning the number of steps executed and/or any error arising.
func ExecuteAll[W any, I any, M Core[W, I]](machine M, n uint) (uint, error) {
	var nsteps uint
	//
	for {
		// Execute upto n steps
		m, err := machine.Execute(n)
		// update the tally
		nsteps += m
		// check for termination
		if err != nil || m < n {
			return nsteps, err
		}
	}
}

// Executor captures a function which can execute a single instruction within
// the context of a given machine's state.  This may produce an error, such as
// when a fail instruction is encountered.
type Executor[W, N any, S State[W, N]] interface {
	Execute(state S) error
}

// Core represents the state of an executing machine, including the state of
// all registers, memories and functions.  A machine may be executing or
// terminated.  Machines are abstracted over a given type of word W, and
// instruction I.  For example, a machine could be operating over 16bit words or
// 8bit words, etc (i.e. as determined by the underlying field).  Furthermore, a
// machine may be operating over instructions compiled into bytes (for efficient
// execution), or instructions represented at a higher level (e.g. for analysis
// or compilation).
type Core[W any, N any] interface {
	// Execute the machine for the given number of steps, returning the actual
	// number of steps executed and an error (if execution failed).
	Execute(steps uint) (uint, error)
	// Return the dynamic state of this machine.  That is, the state which can
	// differ between executions of the same machine (e.g. ROM) and/or within a
	// single execution of the machine (e.g. RAM).
	State() State[W, N]
}

// StaticState captures the static state of an executing machine, such as the
// functions and any static ROMs (e.g. for static reference tables which do not
// change between different executions of a given machine).
type StaticState[W any, N any] interface {
	// Return the ith function in this machine in order, for example, to access
	// its compiled bytecode.
	Function(id uint) function.Function[N]
	// Return the number of functions in this machine.
	NumFunctions() uint
	// Return the number of (static) Read-Only Memory's of this machine.
	NumStatics() uint
	// Return the ith (static) Read-Only Memory' (ROM) of this machine.  Static
	// ROMs are inputs which are fixed across all executions of the given
	// machine.  For example, they might correspond with a static reference
	// table declared in the source program.
	Static(id uint) memory.ReadOnlyMemory[W]
}

// DynamicState captures the non-static state of an executing machine, including
// the call stack, all RAMs, WOMs and (non-static) ROMs.  It does not, however,
// include the functions (which are static by definition) and any static ROMs
// (e.g. for static reference tables), as these do not change between different
// executions of a given machine.
type DynamicState[W any] interface {
	// Current call stack of the machine.  This consists of zero or more stack
	// frames, where that with highest index is currently executing.  If the
	// call stack is empty, then the machine has terminated.
	CallStack() *stack.Stack[Frame[W]]
	// Return the ith Read-Only Memory (ROMS) in this machine. Non-Static ROMs are
	// used as inputs to a given execution of the machine (i.e. they can change
	// between different executions of the same machine).
	Input(id uint) memory.ReadOnlyMemory[W]
	// Return the ith Write-Once Memory (WOM) in this machine which are used for
	// writing the outputs of the machine, Roughly speaking, they can be thought
	// of as output streams.  All WOMs are empty at the start of execution, and
	// may be written values as the program executes.
	Output(id uint) memory.WriteOnceMemory[W]
	// Return the ith Random-Access Memory in this machine.  Such memories are
	// the workhorse of execution, providing unbounded storage.  Initially, all
	// RAMs are empty and maybe assigned values during machine execution.  The
	// size of a RAM expands dynamically as it is written, with all locations
	// initially holding zero.
	Memory(id uint) memory.Memory[W]
	// Return the number of input memories in this machine.
	NumInputs() uint
	// Return the number of output memories in this machine.
	NumOutputs() uint
	// Return the number of random-access memories in this machine.
	NumMemories() uint
}

// State combines the static and dynamic state of a machine into a single
// abstraction.
type State[W any, N any] interface {
	StaticState[W, N]
	DynamicState[W]
}

// Frame represents an executing function on the call stack.  Specifically,
// it contains the state of all registers at the current point of execution for
// that function.
type Frame[W any] struct {
	// Function identifier
	functionId uint
	// Program Counter
	pc uint
	// Registers
	registers []W
}

// NewFrame constructs an initially empty frame for a function with a given
// number of registers.
func NewFrame[W any](fid uint, width uint) Frame[W] {
	return Frame[W]{
		functionId: fid,
		pc:         0,
		registers:  make([]W, width),
	}
}

// Function identifies the function to which this stack frame corresponds.
func (p *Frame[W]) Function() uint {
	return p.functionId
}

// PC returns the current Program Counter position.
func (p *Frame[W]) PC() uint {
	return p.pc
}

// Goto sets the Program Counter to a given position.
func (p *Frame[W]) Goto(pc uint) {
	p.pc = pc
}

// Load the value of the ith register from this stack frame.
func (p *Frame[W]) Load(reg uint) W {
	return p.registers[reg]
}

// Store a given value into the ith register of this stack frame, overwriting
// its previous contents.
func (p *Frame[W]) Store(reg uint, value W) {
	p.registers[reg] = value
}
