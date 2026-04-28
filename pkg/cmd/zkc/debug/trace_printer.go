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
	"bufio"
	"fmt"
	"io"

	"github.com/consensys/go-corset/pkg/schema/register"
	"github.com/consensys/go-corset/pkg/trace"
	"github.com/consensys/go-corset/pkg/util/termio"
	"github.com/consensys/go-corset/pkg/zkc/vm/function"
	"github.com/consensys/go-corset/pkg/zkc/vm/instruction"
	"github.com/consensys/go-corset/pkg/zkc/vm/machine"
	"github.com/consensys/go-corset/pkg/zkc/vm/word"
)

type TracePrinter[W word.Word[W]] struct {
	out bufio.Writer
	// Underlying machine
	vm machine.Base[W]
	// Formatting to use for the Program Counter
	pcFormat termio.AnsiEscape
	// Formatting to use for the body of the instruction
	insnFormat termio.AnsiEscape
	// Formatting to use for the debugging values
	valueFormat termio.AnsiEscape
}

func NewTracePrinter[W word.Word[W]](out io.Writer, vm *machine.Base[W]) TracePrinter[W] {
	return TracePrinter[W]{
		out:         *bufio.NewWriter(out),
		vm:          *vm,
		pcFormat:    termio.NewAnsiEscape().FgColour(termio.TERM_YELLOW),
		insnFormat:  termio.NewAnsiEscape().FgColour(termio.TERM_WHITE),
		valueFormat: termio.NewAnsiEscape().Fg256Colour(250),
	}
}

// PrintAll prints one (or more) execution steps.
func (p *TracePrinter[W]) PrintAll(steps []ExecutionStep[W]) error {
	//
	for _, step := range steps {
		p.print(step)
		// Write new line
		if _, err := p.out.WriteString("\n"); err != nil {
			return err
		}
	}
	//
	return p.out.Flush()
}

func (p *TracePrinter[W]) print(step ExecutionStep[W]) {
	var (
		fun = p.vm.Module(step.Fun).(*function.Boot[W])
	)
	//
	p.printPc(step.Pc)
	p.printInstruction(step.Pc, *fun)
}

func (p *TracePrinter[W]) printPc(pc machine.ProgramCounter) {
	// Construct string representation of PC
	pcStr := fmt.Sprintf("[%02x.%02x] ", pc.Macro(), pc.Micro())
	// Add formatting
	ansi := termio.NewFormattedText(pcStr, p.pcFormat)
	// Write out
	p.out.WriteString(string(ansi.Bytes()))
}

func (p *TracePrinter[W]) printInstruction(pc machine.ProgramCounter, fun function.Boot[W]) {
	var (
		insn = fun.CodeAt(pc.Macro())
		m, n = pc.Micro(), uint(1)
	)
	// Handle vector instructions
	if vec, ok := insn.(*instruction.Vector[W]); ok {
		insn = vec.Codes[pc.Micro()]
		n = uint(len(vec.Codes))
	}
	//
	p.printMicroInstruction(m, n, insn.(instruction.MicroInstruction[W]), fun)
}

func (p *TracePrinter[W]) printMicroInstruction(m, n uint, insn instruction.MicroInstruction[W], fun function.Boot[W]) {
	var (
		name = trace.ModuleName{Name: fun.Name(), Multiplier: 1}
		base = instruction.NewSystemMap(register.ArrayMap(name, fun.Registers()...), p.vm.Modules())
	)
	//
	if n == 1 {
		p.out.WriteString(insn.String(base))
	} else if m == 0 {
		p.out.WriteString(insn.String(base))
		p.out.WriteString(" ; ... ")
	} else if m+1 == n {
		p.out.WriteString("... ; ")
		p.out.WriteString(insn.String(base))
	} else {
		p.out.WriteString("... ; ")
		p.out.WriteString(insn.String(base))
		p.out.WriteString(" ; ... ")
	}
}
