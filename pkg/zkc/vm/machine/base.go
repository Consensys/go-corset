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
	"github.com/consensys/go-corset/pkg/util/field"
	"github.com/consensys/go-corset/pkg/zkc/util"
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
	field     field.Config
	modules   []Module[W]
	callstack []Frame[W]
}

// New constructs a new empty base machine
func New[W word.Word[W]](field field.Config, modules ...Module[W]) *Base[W] {
	return &Base[W]{
		field:     field,
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

// Depth returns the depth of the call stack.
func (p *Base[W]) Depth() uint {
	return uint(len(p.callstack))
}

// StackFrame returns the nth stack frame, where n==0 returns the root frame.
func (p *Base[W]) StackFrame(n uint) Frame[W] {
	return p.callstack[n]
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

// nolint
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
	switch insn.OpCode() {
	// ==============================================================
	// Control-Flow Instructions
	// ==============================================================
	case instruction.CALL:
		insn := insn.(*instruction.Call[W])
		// Enter callee stack frame
		p.Enter(insn.Id, frame, insn.Arguments, insn.Returns)
		// Don't fall thru
		return nil
	case instruction.FAIL:
		return errors.New("machine panic")
	case instruction.JUMP:
		insn := insn.(*instruction.Jmp[W])
		// Goto target instruction in current frame
		p.callstack[fp].Goto(pc.Goto(uint(insn.Immediate)))
		return nil
	case instruction.RETURN:
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
	case instruction.INT_ADD:
		insn := insn.(*instruction.IntAdd[W])
		err = executeAdd(insn.Target, insn.Sources, insn.Constant, frame, regs)
	case instruction.INT_DIV:
		insn := insn.(*instruction.IntDiv[W])
		err = executeDiv(insn.Target, insn.Sources, frame, regs)
	case instruction.INT_MUL:
		insn := insn.(*instruction.IntMul[W])
		err = executeMul(insn.Target, insn.Sources, insn.Constant, frame, regs)
	case instruction.INT_REM:
		insn := insn.(*instruction.IntRem[W])
		err = executeRem(insn.Target, insn.Sources, frame, regs)
	case instruction.INT_SUB:
		insn := insn.(*instruction.IntSub[W])
		err = executeSub(insn.Target, insn.Sources, insn.Constant, frame, regs)

	// ==============================================================
	// Bitwise Instructions
	// ==============================================================
	case instruction.BIT_AND:
		insn := insn.(*instruction.BitAnd[W])
		err = executeAnd(insn.Target, insn.Sources, insn.Constant, frame, regs)
	case instruction.BIT_NOT:
		insn := insn.(*instruction.BitNot[W])
		err = executeNot(insn.Target, insn.Sources, frame, regs)
	case instruction.BIT_OR:
		insn := insn.(*instruction.BitOr[W])
		err = executeOr(insn.Target, insn.Sources, insn.Constant, frame, regs)
	case instruction.BIT_XOR:
		insn := insn.(*instruction.BitXor[W])
		err = executeXor(insn.Target, insn.Sources, insn.Constant, frame, regs)
	case instruction.BIT_SHL:
		insn := insn.(*instruction.BitShl[W])
		err = executeShl(insn.Target, insn.Sources, frame, regs)
	case instruction.BIT_SHR:
		insn := insn.(*instruction.BitShr[W])
		err = executeShr(insn.Target, insn.Sources, frame, regs)

	// ==============================================================
	// Memory Instructions
	// ==============================================================
	case instruction.MEMORY_READ:
		var (
			insn = insn.(*instruction.MemRead[W])
			rom  = p.modules[insn.Id].(memory.Memory[W])
		)
		// Read data words from tiven address
		err = rom.Read(frame, insn.Arguments, insn.Returns)
		// Fall thru
	case instruction.MEMORY_WRITE:
		var (
			insn = insn.(*instruction.MemWrite[W])
			rom  = p.modules[insn.Id].(memory.Memory[W])
		)
		// Read data words from tiven address
		err = rom.Write(frame, insn.Arguments, insn.Returns)
		// Fall thru

	// ==============================================================
	// Misc Instructions
	// ==============================================================
	case instruction.CAST:
		insn := insn.(*instruction.Cast[W])
		err = executeCast(*insn, frame, regs)
		// Fall thru
	case instruction.BIT_CONCAT:
		insn := insn.(*instruction.BitConcat[W])
		err = executeConcat(insn.Target, insn.Sources, frame, regs)
		// Fall thru
	case instruction.BIT_DESTRUCT:
		insn := insn.(*instruction.Destruct[W])
		err = executeDestruct(*insn, frame, regs)
		// Fall thru
	case instruction.SKIP:
		insn := insn.(*instruction.Skip[W])
		// Skip some micro-instructions
		pc = pc.Skip(insn.Skip)
		// Fall thru
	case instruction.SKIP_IF:
		insn := insn.(*instruction.SkipIf[W])
		// Skip (conditionally) micro-instructions
		if executeCondition(frame, insn.Cond, insn.Left, insn.Right) {
			pc = pc.Skip(insn.Skip)
		}
		// Fall thru
	case instruction.DEBUG:
		insn := insn.(*instruction.Debug[W])
		err = executeDebug(*insn, frame, regs)
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

func executeAdd[W word.Word[W]](target register.Id, sources []register.Id, constant W, frame []W,
	regs []register.Register) error {
	//
	var (
		bitwidth = regs[target.Unwrap()].Width()
		overflow bool
	)
	//
	for _, arg := range sources {
		constant, overflow = constant.Add(bitwidth, frame[arg.Unwrap()])
		//
		if overflow {
			return errors.New("executeAdd arithmetic overflow")
		}
	}
	//
	frame[target.Unwrap()] = constant
	//
	return nil
}

func executeMul[W word.Word[W]](target register.Id, sources []register.Id, constant W, frame []W,
	regs []register.Register) error {
	//
	var (
		val      W = constant
		bitwidth   = regs[target.Unwrap()].Width()
		overflow bool
	)
	//
	for _, arg := range sources {
		val, overflow = val.Mul(bitwidth, frame[arg.Unwrap()])
		//
		if overflow {
			return errors.New("executeMul arithmetic overflow")
		}
	}
	//
	frame[target.Unwrap()] = val
	//
	return nil
}

func executeSub[W word.Word[W]](target register.Id, sources []register.Id, constant W, frame []W,
	regs []register.Register) error {
	//
	var (
		val       W
		bitwidth  = regs[target.Unwrap()].Width()
		underflow bool
	)
	//
	for i, arg := range sources {
		if i == 0 {
			val = frame[arg.Unwrap()]
		} else {
			if val, underflow = val.Sub(bitwidth, frame[arg.Unwrap()]); underflow {
				return errors.New("arithmetic underflow")
			}
		}
	}
	// Subtract constant
	if val, underflow = val.Sub(bitwidth, constant); underflow {
		return errors.New("arithmetic underflow")
	}
	//
	frame[target.Unwrap()] = val
	//
	return nil
}

func executeDiv[W word.Word[W]](target register.Id, sources []register.Id, frame []W,
	regs []register.Register) error {
	//
	var (
		bitwidth = regs[target.Unwrap()].Width()
		dividend = frame[sources[0].Unwrap()]
		divisor  = frame[sources[1].Unwrap()]
	)
	//
	if divisor.BigInt().Sign() == 0 {
		return errors.New("division by zero")
	}
	//
	frame[target.Unwrap()] = dividend.Div(bitwidth, divisor)
	//
	return nil
}

func executeRem[W word.Word[W]](target register.Id, sources []register.Id, frame []W,
	regs []register.Register) error {
	//
	var (
		bitwidth = regs[target.Unwrap()].Width()
		dividend = frame[sources[0].Unwrap()]
		divisor  = frame[sources[1].Unwrap()]
	)
	//
	if divisor.BigInt().Sign() == 0 {
		return errors.New("division by zero")
	}
	//
	frame[target.Unwrap()] = dividend.Rem(bitwidth, divisor)
	//
	return nil
}

// ==============================================================
// Bitwise Instructions
// ==============================================================

func executeAnd[W word.Word[W]](target register.Id, sources []register.Id, constant W, frame []W,
	regs []register.Register) error {
	//
	var (
		val      W = constant
		bitwidth   = regs[target.Unwrap()].Width()
	)
	//
	for _, arg := range sources {
		val = val.And(bitwidth, frame[arg.Unwrap()])
	}
	//
	frame[target.Unwrap()] = val
	//
	return nil
}
func executeOr[W word.Word[W]](target register.Id, sources []register.Id, constant W, frame []W,
	regs []register.Register) error {
	//
	var (
		val      W = constant
		bitwidth   = regs[target.Unwrap()].Width()
	)
	//
	for _, arg := range sources {
		val = val.Or(bitwidth, frame[arg.Unwrap()])
	}
	//
	frame[target.Unwrap()] = val
	//
	return nil
}

func executeXor[W word.Word[W]](target register.Id, sources []register.Id, constant W, frame []W,
	regs []register.Register) error {
	//
	var (
		val      W = constant
		bitwidth   = regs[target.Unwrap()].Width()
	)
	//
	for _, arg := range sources {
		val = val.Xor(bitwidth, frame[arg.Unwrap()])
	}
	//
	frame[target.Unwrap()] = val
	//
	return nil
}

func executeNot[W word.Word[W]](target register.Id, sources []register.Id, frame []W,
	regs []register.Register) error {
	//
	var (
		bitwidth = regs[target.Unwrap()].Width()
		arg      = frame[sources[0].Unwrap()]
	)
	//
	frame[target.Unwrap()] = arg.Not(bitwidth)
	//
	return nil
}

// ==============================================================
// Shift Instructions
// ==============================================================

func executeShl[W word.Word[W]](target register.Id, sources []register.Id, frame []W,
	regs []register.Register) error {
	//
	var (
		bitwidth = regs[target.Unwrap()].Width()
		lhs      = frame[sources[0].Unwrap()]
		rhs      = frame[sources[1].Unwrap()]
	)
	//
	frame[target.Unwrap()] = lhs.Shl(bitwidth, rhs)
	//
	return nil
}

func executeShr[W word.Word[W]](target register.Id, sources []register.Id, frame []W,
	regs []register.Register) error {
	//
	var (
		bitwidth = regs[target.Unwrap()].Width()
		lhs      = frame[sources[0].Unwrap()]
		rhs      = frame[sources[1].Unwrap()]
	)
	//
	frame[target.Unwrap()] = lhs.Shr(bitwidth, rhs)
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

func executeConcat[W word.Word[W]](target register.Id, sources []register.Id, frame []W,
	regs []register.Register) error {
	//
	var (
		val    W
		offset uint64
		width  = regs[target.Unwrap()].Width()
	)
	//
	for _, reg := range sources {
		// determine register width
		var (
			reg_width = regs[reg.Unwrap()].Width()
			reg_val   = frame[reg.Unwrap()]
		)
		// Merge bits from value at the correct position
		val = val.Or(width, reg_val.Shl64(width, offset))
		// Update width accumulate
		offset += uint64(reg_width)
	}
	//
	frame[target.Unwrap()] = val
	//
	return nil
}

func executeDestruct[W word.Word[W]](insn instruction.Destruct[W], frame []W, regs []register.Register) error {
	var val = frame[insn.Source.Unwrap()]
	//
	for _, reg := range insn.Targets {
		// determine register width
		var reg_width = regs[reg.Unwrap()].Width()
		//
		frame[reg.Unwrap()] = val.Slice(reg_width)
		// Shift val
		val = val.Shr64(uint64(reg_width))
	}
	//
	return nil
}

func executeDebug[W word.Word[W]](insn instruction.Debug[W], frame []W, _ []register.Register) error {
	for _, chunk := range insn.Chunks {
		fmt.Printf("%s", chunk.Text)
		//
		if chunk.Format.HasFormat() {
			fmt.Printf("%s", util.FormatWord(chunk.Format, frame[chunk.Argument.Unwrap()]))
		}
	}
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
