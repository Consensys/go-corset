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
				p.Enter(uint(i), nil, nil, nil)
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
func (p *Base[W]) Enter(id uint, frame []W, args []register.Id, returns []register.Id) {
	var (
		mainFn    = p.modules[id].(*function.Boot[W])
		bootFrame = NewFrame[W](id, mainFn.Width(), uint(len(args)), returns)
	)
	// Initialise arguments
	for i, arg := range args {
		bootFrame.Store(uint(i), frame[arg.Unwrap()])
	}
	// Push frame onto call stack
	p.callstack = append(p.callstack, bootFrame)
}

// Leave implementation for the machine.Core interface.
func (p *Base[W]) Leave() bool {
	var (
		n     = len(p.callstack) - 1
		frame = p.callstack[n]
	)
	// pop call stack
	p.callstack = p.callstack[:n]
	// write returns (if applicable)
	if n >= 1 {
		//
		for i, r := range frame.returns {
			val := frame.Return(uint(i))
			p.callstack[n-1].Store(r.Unwrap(), val)
		}
	}
	//
	return n == 0
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
	var uInsn, width, regs, err = p.decode()
	//
	if err == nil {
		// Execute
		err = p.executeInstruction(uInsn, width, regs)
	}
	//
	return err
}

func (p *Base[W]) decode() (instruction.MicroInstruction[W], uint, []register.Register, error) {
	var (
		n = len(p.callstack) - 1
		// Extract executing frame
		frame = p.callstack[n]
		// Identify enclosing function
		fn = p.modules[frame.Function()].(*function.Boot[W])
		// Determine current PC position
		pc = frame.PC()
		// Lookup instruction to execute
		insn = fn.CodeAt(pc.Macro())
		//
		uInsn instruction.MicroInstruction[W]
		//
		width = uint(1)
	)
	//
	// Check for vector instruction
	if insn == nil {
		return nil, 0, nil, errors.New("invalid macro instruction")
	} else if vInsn, ok := insn.(*instruction.Vector[W]); ok {
		uInsn = vInsn.Codes[pc.Micro()]
		width = uint(len(vInsn.Codes))
	} else if pc.Micro() != 0 {
		return nil, 0, nil, errors.New("invalid micro instruction")
	} else if insn, ok := insn.(instruction.MicroInstruction[W]); ok {
		uInsn = insn
	} else {
		return nil, 0, nil, errors.New("invalid instruction")
	}
	//
	return uInsn, width, fn.Registers(), nil
}

func (p *Base[W]) executeInstruction(insn instruction.MicroInstruction[W], width uint, regs []register.Register,
) (err error) {
	//
	var (
		fp         = len(p.callstack) - 1
		stackframe = p.callstack[fp]
		frame      = stackframe.registers
		pc         = stackframe.PC()
	)
	//
	//nolint
	switch insn := insn.(type) {

	// ==============================================================
	// Control-Flow Instructions
	// ==============================================================
	case *instruction.Call:
		// Enter callee stack frame
		p.Enter(insn.Id, frame, insn.Arguments, insn.Returns)
		// Don't fall thru
		return nil
	case *instruction.Fail:
		return errors.New("machine panic")
	case *instruction.Jmp:
		// Goto target instruction in current frame
		p.callstack[fp].Goto(pc.Goto(insn.Target))
		return nil
	case *instruction.Return:
		if done := p.Leave(); done {
			return nil
		}
		// adjust frame pointer
		fp = fp - 1
		// redecode instruction to update width
		_, width, _, err = p.decode()
		// reload pc
		pc = p.callstack[fp].PC()
		// fall thru

	// ==============================================================
	// Arithmetic Instructions
	// ==============================================================
	case *instruction.Add[W]:
		err = executeAdd(*insn, frame, regs)
	case *instruction.Div[W]:
		err = executeDiv(*insn, frame, regs)
	case *instruction.Mul[W]:
		err = executeMul(*insn, frame, regs)
	case *instruction.Rem[W]:
		err = executeRem(*insn, frame, regs)
	case *instruction.Sub[W]:
		err = executeSub(*insn, frame, regs)

	// ==============================================================
	// Bitwise Instructions
	// ==============================================================
	case *instruction.And[W]:
		err = executeAnd(*insn, frame, regs)
	case *instruction.Not[W]:
		err = executeNot(*insn, frame, regs)
	case *instruction.Or[W]:
		err = executeOr(*insn, frame, regs)
	case *instruction.Xor[W]:
		err = executeXor(*insn, frame, regs)

	// ==============================================================
	// Shift Instructions
	// ==============================================================
	case *instruction.Shl[W]:
		err = executeShl(*insn, frame, regs)
	case *instruction.Shr[W]:
		err = executeShr(*insn, frame, regs)

	// ==============================================================
	// Memory Instructions
	// ==============================================================
	case *instruction.MemRead:
		var rom = p.modules[insn.Id].(memory.Memory[W])
		// Read data words from tiven address
		err = rom.FrameRead(frame, insn.Address, insn.Data)
		// Fall thru
	case *instruction.MemWrite:
		var rom = p.modules[insn.Id].(memory.Memory[W])
		// Read data words from tiven address
		err = rom.FrameWrite(frame, insn.Address, insn.Data)
		// Fall thru

	// ==============================================================
	// Misc Instructions
	// ==============================================================
	case *instruction.Cast[W]:
		err = executeCast(*insn, frame, regs)
	case *instruction.Skip:
		// Skip some micro-instructions
		pc = pc.Skip(insn.Skip)
		// Fall thru
	case *instruction.SkipIf:
		// Skip (conditionally) micro-instructions
		if executeCondition(frame, insn.Cond, insn.Left, insn.Right) {
			pc = pc.Skip(insn.Skip)
		}
		// Fall thru
	default:
		panic("unknown instruction encountered")
	}
	// Fall through to next instruction if no error.
	if err == nil {
		p.callstack[fp].Goto(pc.Next(width))
	}
	// Done
	return err
}

// ==============================================================
// Arithmetic Instructions
// ==============================================================

func executeAdd[W word.Word[W]](insn instruction.Add[W], frame []W, regs []register.Register) error {
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
			return errors.New("arithmetic overflow")
		}
	}
	//
	frame[insn.Target.Unwrap()] = val
	//
	return nil
}

func executeMul[W word.Word[W]](insn instruction.Mul[W], frame []W, regs []register.Register) error {
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
			return errors.New("arithmetic overflow")
		}
	}
	//
	frame[insn.Target.Unwrap()] = val
	//
	return nil
}

func executeSub[W word.Word[W]](insn instruction.Sub[W], frame []W, regs []register.Register) error {
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
				return errors.New("arithmetic underflow")
			}
		}
	}
	// Subtract constant
	if val, underflow = val.Sub(bitwidth, insn.Constant); underflow {
		return errors.New("arithmetic underflow")
	}
	//
	frame[insn.Target.Unwrap()] = val
	//
	return nil
}

func executeDiv[W word.Word[W]](insn instruction.Div[W], frame []W, regs []register.Register) error {
	var (
		bitwidth = regs[insn.Target.Unwrap()].Width()
		dividend = frame[insn.Dividend.Unwrap()]
		divisor  = frame[insn.Divisor.Unwrap()]
	)
	//
	if divisor.BigInt().Sign() == 0 {
		return errors.New("division by zero")
	}
	//
	frame[insn.Target.Unwrap()] = dividend.Div(bitwidth, divisor)
	//
	return nil
}

func executeRem[W word.Word[W]](insn instruction.Rem[W], frame []W, regs []register.Register) error {
	var (
		bitwidth = regs[insn.Target.Unwrap()].Width()
		dividend = frame[insn.Dividend.Unwrap()]
		divisor  = frame[insn.Divisor.Unwrap()]
	)
	//
	if divisor.BigInt().Sign() == 0 {
		return errors.New("division by zero")
	}
	//
	frame[insn.Target.Unwrap()] = dividend.Rem(bitwidth, divisor)
	//
	return nil
}

// ==============================================================
// Bitwise Instructions
// ==============================================================

func executeAnd[W word.Word[W]](insn instruction.And[W], frame []W, regs []register.Register) error {
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
	return nil
}
func executeOr[W word.Word[W]](insn instruction.Or[W], frame []W, regs []register.Register) error {
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
	return nil
}

func executeXor[W word.Word[W]](insn instruction.Xor[W], frame []W, regs []register.Register) error {
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
	return nil
}

func executeNot[W word.Word[W]](insn instruction.Not[W], frame []W, regs []register.Register) error {
	var (
		bitwidth = regs[insn.Target.Unwrap()].Width()
		arg      = frame[insn.Source.Unwrap()]
	)
	//
	frame[insn.Target.Unwrap()] = arg.Not(bitwidth)
	//
	return nil
}

// ==============================================================
// Shift Instructions
// ==============================================================

func executeShl[W word.Word[W]](insn instruction.Shl[W], frame []W, regs []register.Register) error {
	var (
		bitwidth = regs[insn.Target.Unwrap()].Width()
		lhs      = frame[insn.Value.Unwrap()]
		rhs      = frame[insn.Amount.Unwrap()]
	)
	//
	frame[insn.Target.Unwrap()] = lhs.Shl(bitwidth, rhs)
	//
	return nil
}

func executeShr[W word.Word[W]](insn instruction.Shr[W], frame []W, regs []register.Register) error {
	var (
		bitwidth = regs[insn.Target.Unwrap()].Width()
		lhs      = frame[insn.Value.Unwrap()]
		rhs      = frame[insn.Amount.Unwrap()]
	)
	//
	frame[insn.Target.Unwrap()] = lhs.Shr(bitwidth, rhs)
	//
	return nil
}

// ==============================================================
// Misc Instructions
// ==============================================================

func executeCast[W word.Word[W]](insn instruction.Cast[W], frame []W, _ []register.Register) error {
	src := frame[insn.Source.Unwrap()]
	sliced := src.Slice(insn.Width)
	// Panic if the source value doesn't fit within the target bit width.
	if src.Cmp(sliced) != 0 {
		return errors.New("cast overflow")
	}
	//
	frame[insn.Target.Unwrap()] = sliced
	//
	return nil
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
