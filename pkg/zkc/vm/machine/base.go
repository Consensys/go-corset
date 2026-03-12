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
		// Check module is a memory
		mem, ok := m.(memory.Memory[W])
		if !ok {
			continue
		}
		// Initialise with provided contents, or reset to empty if not supplied.
		mem.Initialise(input[m.Name()])
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
	//nolint
	switch insn := insn.(type) {

	// ==============================================================
	// Control-Flow Instructions
	// ==============================================================

	case *instruction.Fail:
		return false, errors.New("machine panic")

	case *instruction.Jmp:
		n := len(p.callstack) - 1
		// Goto target instruction in current frame
		p.callstack[n].Goto(insn.Target)
		// Don't fall through to next instruction
		return false, nil

	case *instruction.Return:
		p.Leave()
		// Fall through to next instruction in the callee frame.
		return true, nil

	// ==============================================================
	// Arithmetic Instructions
	// ==============================================================

	case *instruction.Add[W]:
		return executeAdd(*insn, frame, regs)
	case *instruction.Mul[W]:
		return executeMul(*insn, frame, regs)
	case *instruction.Sub[W]:
		return executeSub(*insn, frame, regs)

	// ==============================================================
	// Bitwise Instructions
	// ==============================================================

	case *instruction.And[W]:
		return executeAnd(*insn, frame, regs)
	case *instruction.Cast[W]:
		return executeCast(*insn, frame, regs)
	case *instruction.Not[W]:
		return executeNot(*insn, frame, regs)
	case *instruction.Or[W]:
		return executeOr(*insn, frame, regs)
	case *instruction.Xor[W]:
		return executeXor(*insn, frame, regs)

	// ==============================================================
	// Shift Instructions
	// ==============================================================
	case *instruction.Shl[W]:
		return executeShl(*insn, frame, regs)
	case *instruction.Shr[W]:
		return executeShr(*insn, frame, regs)

	// ==============================================================
	// Memory Instructions
	// ==============================================================

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
	case *instruction.MemWrite:
		var (
			address = make([]W, len(insn.Targets))
			data    = make([]W, len(insn.Sources))
		)
		// Read address lines
		for i, arg := range insn.Targets {
			address[i] = frame[arg.Unwrap()]
		}
		// Read data lines
		for i, arg := range insn.Sources {
			data[i] = frame[arg.Unwrap()]
		}
		// Write data words at given address
		p.Write(insn.Id, address, data)
		//
		return true, nil

	// ==============================================================
	// Misc
	// ==============================================================

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

// ==============================================================
// Arithmetic Instructions
// ==============================================================

func executeAdd[W word.Word[W]](insn instruction.Add[W], frame []W, regs []register.Register) (bool, error) {
	var (
		val      W = insn.Constant
		bitwidth   = regs[insn.Target.Unwrap()].Width()
		overflow bool
	)
	//
	for _, arg := range insn.Sources {
		val, overflow = val.Add(bitwidth, frame[arg.Unwrap()])
		//
		if overflow {
			return false, errors.New("arithmetic overflow")
		}
	}
	//
	frame[insn.Target.Unwrap()] = val
	//
	return true, nil
}

func executeMul[W word.Word[W]](insn instruction.Mul[W], frame []W, regs []register.Register) (bool, error) {
	var (
		val      W = insn.Constant
		bitwidth   = regs[insn.Target.Unwrap()].Width()
		overflow bool
	)
	//
	for _, arg := range insn.Sources {
		val, overflow = val.Mul(bitwidth, frame[arg.Unwrap()])
		//
		if overflow {
			return false, errors.New("arithmetic overflow")
		}
	}
	//
	frame[insn.Target.Unwrap()] = val
	//
	return true, nil
}

func executeSub[W word.Word[W]](insn instruction.Sub[W], frame []W, regs []register.Register) (bool, error) {
	var (
		val       W
		bitwidth  = regs[insn.Target.Unwrap()].Width()
		underflow bool
	)
	//
	for i, arg := range insn.Sources {
		if i == 0 {
			val = frame[arg.Unwrap()]
		} else {
			if val, underflow = val.Sub(bitwidth, frame[arg.Unwrap()]); underflow {
				return false, errors.New("arithmetic underflow")
			}
		}
	}
	// Subtract constant
	if val, underflow = val.Sub(bitwidth, insn.Constant); underflow {
		return false, errors.New("arithmetic underflow")
	}
	//
	frame[insn.Target.Unwrap()] = val
	//
	return true, nil
}

// ==============================================================
// Bitwise Instructions
// ==============================================================

func executeAnd[W word.Word[W]](insn instruction.And[W], frame []W, regs []register.Register) (bool, error) {
	var (
		val      W = insn.Constant
		bitwidth   = regs[insn.Target.Unwrap()].Width()
	)
	//
	for _, arg := range insn.Sources {
		val = val.And(bitwidth, frame[arg.Unwrap()])
	}
	//
	frame[insn.Target.Unwrap()] = val
	//
	return true, nil
}
func executeOr[W word.Word[W]](insn instruction.Or[W], frame []W, regs []register.Register) (bool, error) {
	var (
		val      W = insn.Constant
		bitwidth   = regs[insn.Target.Unwrap()].Width()
	)
	//
	for _, arg := range insn.Sources {
		val = val.Or(bitwidth, frame[arg.Unwrap()])
	}
	//
	frame[insn.Target.Unwrap()] = val
	//
	return true, nil
}

func executeXor[W word.Word[W]](insn instruction.Xor[W], frame []W, regs []register.Register) (bool, error) {
	var (
		val      W = insn.Constant
		bitwidth   = regs[insn.Target.Unwrap()].Width()
	)
	//
	for _, arg := range insn.Sources {
		val = val.Xor(bitwidth, frame[arg.Unwrap()])
	}
	//
	frame[insn.Target.Unwrap()] = val
	//
	return true, nil
}

func executeCast[W word.Word[W]](insn instruction.Cast[W], frame []W, _ []register.Register) (bool, error) {
	src := frame[insn.Source.Unwrap()]
	sliced := src.Slice(insn.Width)
	// Panic if the source value doesn't fit within the target bit width.
	if src.Cmp(sliced) != 0 {
		return false, errors.New("cast overflow")
	}
	//
	frame[insn.Target.Unwrap()] = sliced
	//
	return true, nil
}

func executeNot[W word.Word[W]](insn instruction.Not[W], frame []W, regs []register.Register) (bool, error) {
	var (
		bitwidth = regs[insn.Target.Unwrap()].Width()
		arg      = frame[insn.Source.Unwrap()]
	)
	//
	frame[insn.Target.Unwrap()] = arg.Not(bitwidth)
	//
	return true, nil
}

// ==============================================================
// Shift Instructions
// ==============================================================

func executeShl[W word.Word[W]](insn instruction.Shl[W], frame []W, regs []register.Register) (bool, error) {
	var (
		bitwidth = regs[insn.Target.Unwrap()].Width()
		lhs      = frame[insn.Value.Unwrap()]
		rhs      = frame[insn.Amount.Unwrap()]
	)
	//
	frame[insn.Target.Unwrap()] = lhs.Shl(bitwidth, rhs)
	//
	return true, nil
}

func executeShr[W word.Word[W]](insn instruction.Shr[W], frame []W, regs []register.Register) (bool, error) {
	var (
		bitwidth = regs[insn.Target.Unwrap()].Width()
		lhs      = frame[insn.Value.Unwrap()]
		rhs      = frame[insn.Amount.Unwrap()]
	)
	//
	frame[insn.Target.Unwrap()] = lhs.Shr(bitwidth, rhs)
	//
	return true, nil
}

// ==============================================================
// Conditions
// ==============================================================
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
