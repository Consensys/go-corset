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

	"github.com/consensys/go-corset/pkg/asm"
	"github.com/consensys/go-corset/pkg/asm/insn"
)

// PrintAssemblyProgram is responsible for printing out a given assembly program
// in a human-readable format.  This is either done at the macro- or
// micro-assembly level, as indicated by the given flag.  When micro-assembly is
// requested, the lowering configuration is used to lower the macro program into
// a micro program.
func PrintAssemblyProgram(micro bool, cfg asm.LoweringConfig, program asm.MacroProgram) {
	//
	if micro {
		// Lower the program.
		uprogram := program.Lower(cfg)
		//
		printAssemblyFunctions(uprogram.Functions())
	} else {
		printAssemblyFunctions(program.Functions())
	}
}

func printAssemblyFunctions[T insn.Instruction](fns []asm.Function[T]) {
	for _, f := range fns {
		printAssemblyFunction(f)
	}
}

func printAssemblyFunction[T insn.Instruction](f asm.Function[T]) {
	printAssemblySignature(f)
	printAssemblyRegisters(f)
	//
	for pc, insn := range f.Code {
		fmt.Printf("[%d]\t%s\n", pc, insn.String(f.Registers))
	}
	//
	fmt.Println("}")
}

func printAssemblySignature[T any](f asm.Function[T]) {
	first := true
	//
	fmt.Printf("fn %s(", f.Name)
	//
	for _, r := range f.Registers {
		if r.IsInput() {
			if !first {
				fmt.Printf(", ")
			} else {
				first = false
			}
			//
			fmt.Printf("%s u%d", r.Name, r.Width)
		}
	}
	//
	fmt.Printf(") -> (")
	// reset
	first = true
	//
	for _, r := range f.Registers {
		if r.IsOutput() {
			if !first {
				fmt.Printf(", ")
			} else {
				first = false
			}
			//
			fmt.Printf("%s u%d", r.Name, r.Width)
		}
	}
	//
	fmt.Println(") {")
}

func printAssemblyRegisters[T any](f asm.Function[T]) {
	for _, r := range f.Registers {
		if !r.IsInput() && !r.IsOutput() {
			fmt.Printf("\tvar %s u%d\n", r.Name, r.Width)
		}
	}
}
