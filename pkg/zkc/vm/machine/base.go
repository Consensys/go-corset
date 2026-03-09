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
	"errors"
	"fmt"

	"github.com/consensys/go-corset/pkg/schema/register"
	"github.com/consensys/go-corset/pkg/zkc/vm/function"
	"github.com/consensys/go-corset/pkg/zkc/vm/instruction"
	"github.com/consensys/go-corset/pkg/zkc/vm/memory"
	"github.com/consensys/go-corset/pkg/zkc/vm/word"
)

// ============================================================================
// Base Machine
// ============================================================================

// Base provides a fundamental implementation of a machine.  The intention is
// that other machine variations would build off this.
type Base[W word.Word[W]] struct {
	modules   []Module[W]
	callstack []Frame[W]
}

// New constructs a new empty base machine
func New[W word.Word[W]](modules ...Module[W]) *Base[W] {
	return &Base[W]{
		modules:   modules,
		callstack: nil,
	}
}

// Boot this machine by starting the given function with the given inputs.  This
// function assumes the given inputs are correctly formed, and will: (1) ingore
// unknown inputs; (2) initialise empty memories when no input is given for
// them.  Thus, it is recommended to perform sanity checking on input prior to
// calling this function.
func (p *Base[W]) Boot(fun string, input map[string][]W) error {
	// Look for function with the machine name
	for i, m := range p.modules {
		if _, ok := m.(*function.Boot[W]); ok {
			if m.Name() == fun {
				// Initialise memory
				p.initialise(input)
				// Boot the frame
				p.Enter(uint(i))
				//
				return nil
			}
		}
	}
	// No function found
	return fmt.Errorf("missing boot function \"%s\"", fun)
}

// Execute the machine for the given number of steps, returning the actual
// number of steps executed and an error (if execution failed).
func (p *Base[W]) Execute(steps uint) (uint, error) {
	var (
		nsteps uint
		err    error
	)
	//
	for len(p.callstack) > 0 && nsteps < steps {
		if err = p.execute(); err != nil {
			return nsteps, err
		}
		//
		nsteps++
	}
	//
	return nsteps, nil
}

// Enter implementation for the machine.Core interface.
func (p *Base[W]) Enter(id uint, args ...W) {
	var (
		mainFn    = p.modules[id].(*function.Boot[W])
		bootFrame = NewFrame[W](id, mainFn.Width())
	)
	//
	p.callstack = append(p.callstack, bootFrame)
}

// Leave implementation for the machine.Core interface.
func (p *Base[W]) Leave() {
	var n = len(p.callstack)
	//
	p.callstack = p.callstack[:n-1]
}

// Module implementation for the machine.Core interface.
func (p *Base[W]) Module(id uint) Module[W] {
	return p.modules[id]
}

// Modules implementation for the machine.Core interface.
func (p *Base[W]) Modules() []Module[W] {
	return p.modules
}

// Read implementation of machine.Core interface
func (p *Base[W]) Read(id uint, address []W) (data []W) {
	var rm = p.modules[id].(memory.Memory[W])
	// Perform read
	return rm.Read(address)
}

// Write implementation of machine.Core interface
func (p *Base[W]) Write(id uint, address []W, data []W) {
	var wm = p.modules[id].(memory.Memory[W])
	// Perform write
	wm.Write(address, data)
}

// ========================================================
// Helpers
// =======================================================

func (p *Base[W]) initialise(input map[string][]W) {
	// Initialise stack input memories
	for _, m := range p.modules {
		// Check module name
		contents, ok1 := input[m.Name()]
		// Check module is a memory
		memory, ok2 := m.(memory.Memory[W])
		//
		if ok1 && ok2 {
			memory.Initialise(contents)
		}
	}
}

// Execute implementation for the Executor interface
func (p *Base[W]) execute() error {
	// Decode
	var insn, state, regs = p.decode()
	// Execute
	fallThru, err := p.executeInstruction(insn, state, regs)
	// Fall through to next instruction (if applicable)
	if fallThru && len(p.callstack) > 0 {
		var n = len(p.callstack) - 1
		p.callstack[n].FallThru()
	}
	//
	return err
}

func (p *Base[W]) decode() (instruction.Instruction[W], []W, []register.Register) {
	var (
		n = len(p.callstack) - 1
		// Extract executing frame
		frame = p.callstack[n]
		// Identify enclosing function
		fn = p.modules[frame.Function()].(*function.Boot[W])
		// Determine current PC position
		pc = frame.PC()
		// Lookup instruction to execute
		insn = fn.CodeAt(pc)
	)
	//
	return insn, frame.registers, fn.Registers()
}

func (p *Base[W]) executeInstruction(insn instruction.Instruction[W], frame []W, regs []register.Register,
) (bool, error) {
	//
	switch insn := insn.(type) {
	case *instruction.Add[W]:
		var val W = insn.Constant
		//
		for _, arg := range insn.Sources {
			val = val.Add(frame[arg.Unwrap()])
		}
		//
		frame[insn.Target.Unwrap()] = val
		//
		return true, nil

	case *instruction.Fail:
		return false, errors.New("machine panic")

	case *instruction.Jmp:
		n := len(p.callstack) - 1
		// Goto target instruction in current frame
		p.callstack[n].Goto(insn.Target)
		// Don't fall through to next instruction
		return false, nil

	case *instruction.Mul[W]:
		var val W = insn.Constant
		//
		for _, arg := range insn.Sources {
			val = val.Mul(frame[arg.Unwrap()])
		}
		//
		frame[insn.Target.Unwrap()] = val
		//
		return true, nil

	case *instruction.MemRead:
		var address = make([]W, len(insn.Sources))
		// Read source registers
		for i, arg := range insn.Sources {
			address[i] = frame[arg.Unwrap()]
		}
		// Read data words from tiven address
		data := p.Read(insn.Id, address)
		// Write into target registers
		for i := range data {
			target := insn.Targets[i].Unwrap()
			frame[target] = data[i]
		}
		//
		return true, nil

	case *instruction.Return:
		p.Leave()
		// Fall through to next instruction in the callee frame.
		return true, nil

	case *instruction.Sub[W]:
		var (
			val      W
			bitwidth = regs[insn.Target.Unwrap()].Width()
		)
		//
		for i, arg := range insn.Sources {
			if i == 0 {
				val = frame[arg.Unwrap()]
			} else {
				val = val.Sub(bitwidth, frame[arg.Unwrap()])
			}
		}
		// Subtract constant
		val = val.Sub(bitwidth, insn.Constant)
		//
		frame[insn.Target.Unwrap()] = val
		//
		return true, nil
	case *instruction.Vector[W]:
		var (
			err      error
			fallThru = true
			codes    = insn.Codes
			nCodes   = uint(len(codes))
		)
		// Execute vector instructions in turn, whilst applying skips.
		for cc := uint(0); cc < nCodes && fallThru; cc++ {
			switch c := codes[cc].(type) {
			case *instruction.Skip:
				cc += c.Skip
			case *instruction.SkipIf:
				//
				if executeCondition(frame, c.Cond, c.Left, c.Right) {
					cc += c.Skip
				}
			case instruction.Instruction[W]:
				fallThru, err = p.executeInstruction(c, frame, regs)
			default:
				panic("unreachable")
			}
		}
		//
		return fallThru, err

	default:
		panic("unknown instruction encountered")
	}
}

func executeCondition[W word.Word[W]](frame []W, cond instruction.Condition, left, right register.Id) bool {
	var (
		lhs = frame[left.Unwrap()]
		rhs = frame[right.Unwrap()]
	)
	//
	switch cond {
	case instruction.EQ:
		return lhs.Cmp(rhs) == 0
	case instruction.NEQ:
		return lhs.Cmp(rhs) != 0
	case instruction.LT:
		return lhs.Cmp(rhs) < 0
	case instruction.LTEQ:
		return lhs.Cmp(rhs) <= 0
	case instruction.GT:
		return lhs.Cmp(rhs) > 0
	case instruction.GTEQ:
		return lhs.Cmp(rhs) >= 0
	default:
		panic("unreachable")
	}
}
