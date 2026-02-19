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
package vm

import (
	"github.com/consensys/go-corset/pkg/asmv2/vm/fun"
	"github.com/consensys/go-corset/pkg/asmv2/vm/ram"
	"github.com/consensys/go-corset/pkg/asmv2/vm/rom"
	"github.com/consensys/go-corset/pkg/asmv2/vm/wom"
)

// Machine represents the state of an executing machine, including the state of
// all registers, and memories.  A machine may be executing or terminated.
type Machine[W any] interface {
	// Current call stack of the machine.  This consists of zero or more stack
	// frames, where that with highest index is currently executing.  If the
	// call stack is empty, then the machine has terminated.
	CallStack() []StackFrame[W]
	// Return the ith function in this machine in order, for example, to access
	// its compiled bytecode.
	Function(id uint) fun.Function[W]
	// Return the ith Read-Only Memory (ROM) in this machine.  ROMs are used as
	// inputs in one of two ways: firstly, as inputs to a given execution;
	// secondly, as static inputs to all executions.  The latter, for example,
	// would correspond with a static reference table declared in the source
	// program.  Eitherway, the contents of a ROM is defined prior to execution.
	Rom(id uint) rom.ReadOnlyMemory[W]
	// Return the ith Random-Access Memory in this machine.  Such memories are
	// the workhorse of execution, providing unbounded storage.  Initially, all
	// RAMs are empty and maybe assigned values during machine execution.  The
	// size of a RAM expands dynamically as it is written, with all locations
	// initially holding zero.
	Ram(id uint) ram.RandomAccessMemory[W]
	// Return the ith Write-Once Memory (WOM) in this machine.  ROMs are used
	// for writing the outputs of the machine and can be thought of (roughly
	// speaking) as output streams.  All WOMs are empty at the start of
	// execution, and may be written values as the program executes.
	Wom(id uint) wom.WriteOnceMemory[W]
}

// StackFrame represents an executing function on the call stack.  Specifically,
// it contains the state of all registers at the current point of execution for
// that function.
type StackFrame[W any] struct {
	// Function identifier
	functionId uint
	// Program Counter
	pc uint
	// Registers
	registers []W
}

// Function identifies the function to which this stack frame corresponds.
func (p *StackFrame[W]) Function() uint {
	return p.functionId
}

// PC returns the current Program Counter position.
func (p *StackFrame[W]) PC() uint {
	return p.pc
}

// Load the value of the ith register from this stack frame.
func (p *StackFrame[W]) Load(reg uint) W {
	return p.registers[reg]
}

// Store a given value into the ith register of this stack frame, overwriting
// its previous contents.
func (p *StackFrame[W]) Store(reg uint, value W) {
	p.registers[reg] = value
}
