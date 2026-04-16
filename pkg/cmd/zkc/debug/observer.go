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
package debug

import (
	"fmt"
	"strings"

	"github.com/consensys/go-corset/pkg/schema/register"
	"github.com/consensys/go-corset/pkg/trace"
	"github.com/consensys/go-corset/pkg/zkc/vm/function"
	"github.com/consensys/go-corset/pkg/zkc/vm/instruction"
	"github.com/consensys/go-corset/pkg/zkc/vm/machine"
	"github.com/consensys/go-corset/pkg/zkc/vm/word"
)

// TraceObserver prints a trace
type TraceObserver[W word.Word[W]] struct {
	depth uint
	fun   *function.Function[instruction.Instruction[W]]
	insn  instruction.MicroInstruction[W]
	pc    machine.ProgramCounter
}

// PreExecution implementation for Observer interface
func (p *TraceObserver[W]) PreExecution(machine *machine.Base[W]) {
	var (
		n = machine.Depth()
	)
	//
	if n > 0 {
		if n != p.depth {
			fmt.Println()
			p.enterFunction(machine)
			fmt.Print(p.callStack(machine))
			fmt.Println()
		}
		//
		p.writeInstruction(machine)
	}
}

// PostExecution implementation for Observer interface
func (p *TraceObserver[W]) PostExecution(machine *machine.Base[W]) {
	var (
		n = machine.Depth()
	)
	//
	if n > 0 {
		switch {
		case n == p.depth:
			// Normal instruction: depth unchanged, use current top frame.
			p.writeStateFromFrame(machine, machine.StackFrame(n-1), false)
		case n == p.depth+1:
			// Call instruction: depth increased. Use caller's frame and skip
			// return register annotations (returns are not yet written).
			p.writeStateFromFrame(machine, machine.StackFrame(n-2), true)
		case n+1 == p.depth:
			// Ret instruction: depth decreased. Callee frame is gone; use
			// current caller frame. ret has no Uses/Definitions so no loads occur.
			p.writeStateFromFrame(machine, machine.StackFrame(n-1), false)
		}

		fmt.Println()
	}
}

func (p *TraceObserver[W]) enterFunction(machine *machine.Base[W]) {
	var (
		n     = machine.Depth()
		frame = machine.StackFrame(n - 1)
	)
	//
	p.depth = n
	p.fun = machine.Module(frame.Function()).(*function.Function[instruction.Instruction[W]])
	p.insn = nil
}

func (p *TraceObserver[W]) writeInstruction(machine *machine.Base[W]) {
	var (
		frame = machine.StackFrame(p.depth - 1)
	)
	//
	p.insn = decode(frame, p.fun)
	p.pc = frame.PC()
}

// writeStateFromFrame prints the current instruction with register values annotated
// inline. frame is the stack frame to read register values from. If skipDefs is
// true, Definitions are excluded from annotation (used for call instructions where
// return registers are not yet written).
func (p *TraceObserver[W]) writeStateFromFrame(machine *machine.Base[W], frame machine.Frame[W], skipDefs bool) {
	var (
		name   = trace.ModuleName{Name: p.fun.Name(), Multiplier: 1}
		base   = instruction.NewSystemMap(register.ArrayMap(name, p.fun.Registers()...), machine.Modules())
		values = make(map[uint]string)
	)
	// Collect register values. In PostExecution, sources still hold their pre-execution values
	// (unmodified by the instruction), while definitions hold their post-execution values.
	// Definitions are added last so that when a register appears on both sides, the
	// post-execution value is shown.
	for _, r := range p.insn.Uses() {
		values[r.Unwrap()] = frame.Load(r.Unwrap()).Text(16)
	}

	if !skipDefs {
		for _, r := range p.insn.Definitions() {
			values[r.Unwrap()] = frame.Load(r.Unwrap()).Text(16)
		}
	}
	//
	annotated := &annotatedMap[W]{base: base, values: values}
	insnStr := fmt.Sprintf("[%02x.%02x] %s", p.pc.Macro(), p.pc.Micro(), p.insn.String(annotated))
	fmt.Print(insnStr)
}

func (p *TraceObserver[W]) callStack(machine *machine.Base[W]) string {
	var builder strings.Builder
	//
	for i := uint(0); i < p.depth; i++ {
		var (
			ith = machine.StackFrame(i)
			fun = machine.Module(ith.Function()).(*function.Function[instruction.Instruction[W]])
		)
		//
		if i+1 == p.depth {
			inputs := functionInputs(ith, fun)
			fmt.Fprintf(&builder, "> %s(%s) ", fun.Name(), inputs)
		} else {
			fmt.Fprintf(&builder, "> %s ", fun.Name())
		}
	}
	//
	return builder.String()
}

func functionInputs[W word.Word[W]](frame machine.Frame[W], fun *function.Function[instruction.Instruction[W]]) string {
	var builder strings.Builder

	for i, r := range fun.Registers() {
		var ith = frame.Load(uint(i))
		//
		if !r.IsInput() {
			break
		} else if i != 0 {
			builder.WriteString(", ")
		}
		//
		fmt.Fprintf(&builder, "%s=0x%s", r.Name(), ith.Text(16))
	}

	return builder.String()
}

func decode[W word.Word[W]](frame machine.Frame[W],
	fn *function.Function[instruction.Instruction[W]]) instruction.MicroInstruction[W] {
	//
	var (
		pc   = frame.PC()
		insn = fn.CodeAt(pc.Macro())
	)
	// nolint
	if uInsn, ok := insn.(*instruction.Vector[W]); ok {
		return uInsn.Codes[pc.Micro()]
	} else if uInsn, ok := insn.(instruction.MicroInstruction[W]); ok {
		return uInsn
	}
	//
	panic("invalid micro instruction")
}

// annotatedMap wraps a SystemMap and annotates each register name with its
// current value as "[0xVAL]", producing inline value display in instruction strings.
type annotatedMap[W word.Word[W]] struct {
	base   instruction.SystemMap[W]
	values map[uint]string // register index → hex value string (no "0x" prefix)
}

func (a *annotatedMap[W]) Register(id register.Id) register.Register {
	reg := a.base.Register(id)
	if val, ok := a.values[id.Unwrap()]; ok {
		return register.New(reg.Kind(), reg.Name()+" [0x"+val+"]", reg.Width(), *reg.Padding())
	}

	return reg
}

func (a *annotatedMap[W]) Module(id uint) instruction.Module[W] { return a.base.Module(id) }

func (a *annotatedMap[W]) Name() trace.ModuleName { return a.base.Name() }

func (a *annotatedMap[W]) HasRegister(name string) (register.Id, bool) {
	return a.base.HasRegister(name)
}

func (a *annotatedMap[W]) Registers() []register.Register { return a.base.Registers() }

func (a *annotatedMap[W]) String() string { return a.base.String() }
