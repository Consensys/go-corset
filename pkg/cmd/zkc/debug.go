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
	"github.com/consensys/go-corset/pkg/util/collection/bit"
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
	width := GetUint(cmd, "width")
	//
	input := ParseInputFile(args[0])
	// Compile source files, or print errors
	program := CompileSourceFiles(args[1:]...)
	//
	executeIrProgram[*TraceObserver[word.Uint]]("main", program, input, &TraceObserver[word.Uint]{width: width})
}

// TraceObserver prints a trace
type TraceObserver[W word.Word[W]] struct {
	depth  uint
	fun    *function.Function[instruction.Instruction[W]]
	last   []string
	width  uint
	widths []uint
}

// PreExecution implementation for Observer interface
func (p *TraceObserver[W]) PreExecution(machine *machine.Base[W]) {

}

// PostExecution implementation for Observer interface
func (p *TraceObserver[W]) PostExecution(machine *machine.Base[W]) {
	var (
		n = machine.Depth()
	)
	//
	if n > 0 {
		if n != p.depth {
			fmt.Println()
			p.enterFunction(machine)
			p.writeFunctionTitle(machine)
			fmt.Println()
		}
		//
		p.writeInstruction(machine)
		p.writeState(machine, false)
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
	p.last = nil
	// construct widths
	p.widths = make([]uint, p.fun.Width())
	//
	for i, r := range p.fun.Registers() {
		var (
			nameLen = uint(len(r.Name()))
			valLen  = 2 + (2 * bit.BytesRequiredFor(r.Width()))
		)
		//
		if r.Width() == 0 {
			valLen = 3
		}
		//
		p.widths[i] = 1 + max(nameLen, valLen)
	}
}

func (p *TraceObserver[W]) writeFunctionTitle(machine *machine.Base[W]) {
	var (
		frame = machine.StackFrame(p.depth - 1)
		pc    = frame.PC()
	)
	//
	printLeftAligned(p.callStack(machine), p.width)
	//
	for i, r := range p.fun.Registers() {
		printRightAligned(r.Name(), p.widths[i])
	}
	//
	fmt.Println()
	// Print initial state
	printLeftAligned("", p.width)
	p.writeState(machine, pc.First())
}

func (p *TraceObserver[W]) writeInstruction(machine *machine.Base[W]) {
	var (
		frame   = machine.StackFrame(p.depth - 1)
		insn    = decode(frame, p.fun)
		name    = trace.ParseModuleName("")
		pc      = frame.PC()
		insnStr = insn.String(register.ArrayMap(name, p.fun.Registers()...))
	)
	//
	insnStr = fmt.Sprintf("[%02x.%02x] %s", pc.Macro(), pc.Micro(), insnStr)
	//
	printLeftAligned(insnStr, p.width)
}

func (p *TraceObserver[W]) writeState(machine *machine.Base[W], inputs bool) {
	var (
		n     = machine.Depth()
		frame = machine.StackFrame(n - 1)
		last  = make([]string, p.fun.Width())
	)
	//
	for i, r := range p.fun.Registers() {
		//
		val := fmt.Sprintf("0x%s", frame.Load(uint(i)).Text(16))
		// Record value for next instruction
		last[i] = val
		//
		if p.last != nil && p.last[i] == val {
			val = "."
		}
		//
		if !inputs || r.IsInput() {
			printRightAligned(val, p.widths[i])
		}
	}
	//
	p.last = last
}

func (p *TraceObserver[W]) callStack(machine *machine.Base[W]) string {
	var builder strings.Builder
	//
	for i := uint(0); i < p.depth; i++ {
		ith := machine.StackFrame(i)
		fun := machine.Module(ith.Function())
		builder.WriteString(fmt.Sprintf("> %s ", fun.Name()))
	}
	//
	return builder.String()
}

func printLeftAligned(str string, width uint) {
	var (
		n = min(uint(len(str)), width)
	)
	//
	if uint(len(str)) > width {
		str = str[:width-2]
		fmt.Printf("%s..", str)
	} else {
		fmt.Print(str)
	}
	//
	for i := n; i < width; i++ {
		fmt.Printf(" ")
	}
}

func printRightAligned(str string, width uint) {
	var (
		n = min(uint(len(str)), width)
	)
	//
	for i := n; i < width; i++ {
		fmt.Printf(" ")
	}
	//
	if uint(len(str)) > width {
		str = str[:width-2]
		fmt.Printf("%s..", str)
	} else {
		fmt.Print(str)
	}
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
	debugCmd.Flags().Uint("width", 50, "width of instruction display panel")
}
