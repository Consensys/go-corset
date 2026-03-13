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
package zkc

import (
	"fmt"
	"strings"

	"github.com/consensys/go-corset/pkg/schema/register"
	"github.com/consensys/go-corset/pkg/trace"
	"github.com/consensys/go-corset/pkg/util/field"
	"github.com/consensys/go-corset/pkg/util/field/bls12_377"
	"github.com/consensys/go-corset/pkg/util/field/gf251"
	"github.com/consensys/go-corset/pkg/util/field/gf8209"
	"github.com/consensys/go-corset/pkg/util/field/koalabear"
	"github.com/consensys/go-corset/pkg/zkc/vm/function"
	"github.com/consensys/go-corset/pkg/zkc/vm/instruction"
	"github.com/consensys/go-corset/pkg/zkc/vm/machine"
	"github.com/consensys/go-corset/pkg/zkc/vm/word"
	"github.com/spf13/cobra"
)

var debugCmd = &cobra.Command{
	Use:     "debug [flags] input.json file1.zkc file2.zkc ...",
	Short:   "Debug a zkc program.",
	Long:    `Debug a zkc program to produce a set of outputs a from given a set of inputs.`,
	Aliases: []string{"exec"},
	Run: func(cmd *cobra.Command, args []string) {
		runFieldAgnosticCmd(cmd, args, debugCmds)
	},
}

// Available instances
var debugCmds = []FieldAgnosticCmd{
	{field.GF_251, runDebugCmd[gf251.Element]},
	{field.GF_8209, runDebugCmd[gf8209.Element]},
	{field.KOALABEAR_16, runDebugCmd[koalabear.Element]},
	{field.BLS12_377, runDebugCmd[bls12_377.Element]},
}

func runDebugCmd[F field.Element[F]](cmd *cobra.Command, args []string) {
	leftWidth := GetUint(cmd, "left-width")
	midWidth := GetUint(cmd, "mid-width")
	//
	input := ParseInputFile(args[0])
	// Compile source files, or print errors
	program := CompileSourceFiles(args[1:]...)
	//
	observer := TraceObserver[word.Uint]{
		leftPane: leftWidth,
		midPane:  midWidth,
	}
	//
	executeIrProgram[*TraceObserver[word.Uint]]("main", program, input, &observer)
	//
	fmt.Println()
}

// TraceObserver prints a trace
type TraceObserver[W word.Word[W]] struct {
	depth       uint
	fun         *function.Function[instruction.Instruction[W]]
	uses        string
	definitions []register.Id
	regWidth    uint
	leftPane    uint
	midPane     uint
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
		if n == p.depth {
			p.writeState(machine)
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
	p.uses = ""
	p.definitions = nil
	p.regWidth = 0
	//
	for _, r := range p.fun.Registers() {
		p.regWidth = max(p.regWidth, uint(len(r.Name())))
	}
}

func (p *TraceObserver[W]) writeInstruction(machine *machine.Base[W]) {
	var (
		frame   = machine.StackFrame(p.depth - 1)
		insn    = decode(frame, p.fun)
		name    = trace.ParseModuleName("")
		pc      = frame.PC()
		insnStr = insn.String(register.ArrayMap(name, p.fun.Registers()...))
		builder strings.Builder
	)
	//
	insnStr = fmt.Sprintf("[%02x.%02x] %s", pc.Macro(), pc.Micro(), insnStr)
	//
	fmt.Print(leftAligned(insnStr, p.leftPane))
	// write uses
	for i, r := range insn.Uses() {
		var (
			ith  = frame.Load(uint(i))
			name = p.fun.Register(r).Name()
		)
		//
		if i != 0 {
			builder.WriteString("; ")
		}
		//
		builder.WriteString(fmt.Sprintf("%s==0x%s", name, ith.Text(16)))
	}
	//
	p.uses = builder.String()
	p.definitions = insn.Definitions()
}

func (p *TraceObserver[W]) writeState(machine *machine.Base[W]) {
	fmt.Print(rightAligned(p.defs(machine), p.midPane))
	//
	if len(p.definitions) > 0 || p.uses != "" {
		fmt.Printf(" ; %s", p.uses)
	}
}

func (p *TraceObserver[W]) defs(machine *machine.Base[W]) string {
	var (
		n       = machine.Depth()
		frame   = machine.StackFrame(n - 1)
		builder strings.Builder
	)
	//
	for i, r := range p.definitions {
		var (
			ith = frame.Load(r.Unwrap())
			reg = p.fun.Register(r)
		)
		//
		if i != 0 {
			builder.WriteString("; ")
		}
		//
		builder.WriteString(rightAligned(reg.Name(), p.regWidth))
		builder.WriteString(":=0x")
		builder.WriteString(ith.Text(16))
	}
	//
	return builder.String()
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
			builder.WriteString(fmt.Sprintf("> %s(%s) ", fun.Name(), inputs))
		} else {
			builder.WriteString(fmt.Sprintf("> %s ", fun.Name()))
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
		builder.WriteString(fmt.Sprintf("%s=0x%s", r.Name(), ith.Text(16)))
	}

	return builder.String()
}

func leftAligned(str string, width uint) string {
	var (
		n       = min(uint(len(str)), width)
		builder strings.Builder
	)
	//
	if uint(len(str)) > width {
		str = str[:width-2]
		builder.WriteString(fmt.Sprintf("%s..", str))
	} else {
		builder.WriteString(str)
	}
	//
	for i := n; i < width; i++ {
		builder.WriteString(" ")
	}
	//
	return builder.String()
}

func rightAligned(str string, width uint) string {
	var (
		n       = min(uint(len(str)), width)
		builder strings.Builder
	)
	//
	for i := n; i < width; i++ {
		builder.WriteString(" ")
	}
	//
	builder.WriteString(str)
	//
	return builder.String()
}

func decode[W word.Word[W]](frame machine.Frame[W],
	fn *function.Function[instruction.Instruction[W]]) instruction.MicroInstruction[W] {
	//
	var (
		pc   = frame.PC()
		insn = fn.CodeAt(pc.Macro())
	)
	//
	if uInsn, ok := insn.(*instruction.Vector[W]); ok {
		return uInsn.Codes[pc.Micro()]
	} else if uInsn, ok := insn.(instruction.MicroInstruction[W]); ok {
		return uInsn
	}
	//
	panic("invalid micro instruction")
}

// ============================================================================
// Misc
// ============================================================================

//nolint:errcheck
func init() {
	rootCmd.AddCommand(debugCmd)
	debugCmd.Flags().Uint("left-width", 40, "width of instruction panel")
	debugCmd.Flags().Uint("mid-width", 40, "width of assignment panel")
}
