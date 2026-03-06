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

// ExecuteAll executes a given machine to completion in chunks of n steps,
// returning the number of steps executed and/or any error arising.
func ExecuteAll[W any, M Core[W]](machine M, n uint) (uint, error) {
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

// Core represents the state of an executing machine, including the state of
// all registers, memories and functions.  A machine may be executing or
// terminated.  Machines are abstracted over a given type of word W, and
// instruction I.  For example, a machine could be operating over 16bit words or
// 8bit words, etc (i.e. as determined by the underlying field).  Furthermore, a
// machine may be operating over instructions compiled into bytes (for efficient
// execution), or instructions represented at a higher level (e.g. for analysis
// or compilation).
type Core[W any] interface {
	// Execute the machine for the given number of steps, returning the actual
	// number of steps executed and an error (if execution failed).
	Execute(steps uint) (uint, error)
	// Return ith module in this machine (either a function or some form of memory).
	Module(id uint) Module[W]
	// Return set of modules in this machine.
	Modules() []Module[W]
	// Enter a new function on the call-stack
	Enter(id uint, args ...W)
	// Leave
	Leave()
	// Read location from ith module.  This must be a readable memory, otherwise
	// this will panic.
	Read(id uint, address []W) (data []W)
	// Write location in ith module.  This must be a writeable memory, otherwise
	// this will panic.
	Write(id uint, address []W, data []W)
}

// Module represents an either a function or memory within the machine.
type Module[W any] interface {
	// Name of this module
	Name() string
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

// FallThru to the next instruction in the frame.
func (p *Frame[W]) FallThru() {
	p.pc++
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
