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
	"math/big"

	"github.com/consensys/go-corset/pkg/schema/register"
	"github.com/consensys/go-corset/pkg/trace"
	"github.com/consensys/go-corset/pkg/util/field"
	"github.com/consensys/go-corset/pkg/util/field/bls12_377"
	"github.com/consensys/go-corset/pkg/util/field/gf251"
	"github.com/consensys/go-corset/pkg/util/field/gf8209"
	"github.com/consensys/go-corset/pkg/util/field/koalabear"
	"github.com/consensys/go-corset/pkg/zkc/compiler/ast"
	"github.com/consensys/go-corset/pkg/zkc/compiler/ast/decl"
	"github.com/consensys/go-corset/pkg/zkc/compiler/ast/variable"
	"github.com/consensys/go-corset/pkg/zkc/vm/function"
	"github.com/consensys/go-corset/pkg/zkc/vm/machine"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var compileCmd = &cobra.Command{
	Use:   "compile [flags] file1.zkc file2.zkc ...",
	Short: "compile zkc source files into a binary package.",
	Long:  `Compile a given set of source file(s) into a single binary package.`,
	Run: func(cmd *cobra.Command, args []string) {
		runFieldAgnosticCmd(cmd, args, compileCmds)
	},
}

// Available instances
var compileCmds = []FieldAgnosticCmd{
	{field.GF_251, runCompileCmd[gf251.Element]},
	{field.GF_8209, runCompileCmd[gf8209.Element]},
	{field.KOALABEAR_16, runCompileCmd[koalabear.Element]},
	{field.BLS12_377, runCompileCmd[bls12_377.Element]},
}

func runCompileCmd[F field.Element[F]](cmd *cobra.Command, args []string) {
	// Configure log level
	if GetFlag(cmd, "verbose") {
		log.SetLevel(log.DebugLevel)
	}
	//
	ir := GetFlag(cmd, "ir")
	ast := GetFlag(cmd, "ast")
	// Compile source files, or print errors
	program := CompileSourceFiles(args)
	//
	if ast {
		writeAbstractSyntaxTree(program)
	}
	//
	if ir {
		writeIntermediateRepresentation(program.BuildMachine())
	}
}

// ============================================================================
// AST
// ============================================================================

func writeAbstractSyntaxTree(program ast.Program) {
	for i, d := range program.Components() {
		if i != 0 {
			fmt.Println()
		}
		//
		writeDeclaration(d)
	}
}

func writeDeclaration(decl ast.Declaration) {
	switch decl := decl.(type) {
	case *ast.Function:
		writeFunction(decl)
	case *ast.Memory:
		writeMemory(decl)
	default:
		panic("unknown declaration encountered")
	}
}

func writeMemory(m *ast.Memory) {
	switch m.Kind {
	case decl.PUBLIC_READ_ONLY_MEMORY:
		fmt.Printf("public input")
	case decl.PRIVATE_READ_ONLY_MEMORY:
		fmt.Printf("private input")
	case decl.PUBLIC_WRITE_ONCE_MEMORY:
		fmt.Printf("public output")
	case decl.PRIVATE_WRITE_ONCE_MEMORY:
		fmt.Printf("private output")
	case decl.PUBLIC_STATIC_MEMORY:
		fmt.Printf("public static")
	case decl.PRIVATE_STATIC_MEMORY:
		fmt.Printf("private static")
	case decl.RANDOM_ACCESS_MEMORY:
		fmt.Printf("memory")
	}
	// address lines
	fmt.Printf(" %s(", m.Name())
	writeMemoryParams(m.Address)
	fmt.Printf(") -> (")
	writeMemoryParams(m.Data)
	fmt.Printf(")")
	//
	if m.Contents != nil {
		fmt.Println(" = {")
		writeMemoryContents(m.Contents)
		fmt.Printf("}")
	}
	//
	fmt.Println()
}

func writeMemoryParams(params []variable.Descriptor) {
	for i, p := range params {
		if i > 0 {
			fmt.Printf(", ")
		}

		fmt.Printf("%s %s", p.DataType.String(), p.Name)
	}
}

func writeMemoryContents(values []big.Int) {
	var N = 20
	//
	for i := 0; i < len(values); i += N {
		var left = len(values) - i
		//
		for j := range min(N, left) {
			fmt.Printf("0x%s", values[i+j].Text(16))
			//
			if i+j+1 != len(values) {
				fmt.Printf(", ")
			}
		}
		//
		fmt.Println()
	}
}

func writeFunction(f *ast.Function) {
	fmt.Printf("fn %s(", f.Name())
	// parameters
	writeFunctionArgs(variable.PARAMETER, f.Variables)
	//
	fmt.Printf(") -> (")
	// returns
	writeFunctionArgs(variable.RETURN, f.Variables)
	//
	fmt.Println(") {")
	//
	writeFunctionVariables(f)
	//
	for pc, insn := range f.Code {
		fmt.Printf("[%d]\t%s\n", pc, insn.String(f))
	}
	// Done
	fmt.Println("}")
}

func writeFunctionArgs(kind variable.Kind, variables []variable.Descriptor) {
	var first = true
	//
	for _, r := range variables {
		if r.Kind == kind {
			if !first {
				fmt.Printf(", ")
			} else {
				first = false
			}
			//
			fmt.Printf("%s %s", r.DataType.String(), r.Name)
		}
	}
}

func writeFunctionVariables(f *ast.Function) {
	for _, r := range f.Variables {
		if r.IsLocal() {
			fmt.Printf("\t%s %s\n", r.DataType.String(), r.Name)
		}
	}
}

// ============================================================================
// Intermediate Representation (IR)
// ============================================================================

func writeIntermediateRepresentation(machine machine.Boot) {
	var (
		state = machine.State()
	)
	// Write statics
	for i := range state.NumStatics() {
		writeIrMemory("static", state.Static(i).Name())
	}
	// Write inputs
	for i := range state.NumInputs() {
		writeIrMemory("input", state.Input(i).Name())
	}
	// Write outputs
	for i := range state.NumOutputs() {
		writeIrMemory("output", state.Output(i).Name())
	}
	// Write memories
	for i := range state.NumMemories() {
		writeIrMemory("memory", state.Memory(i).Name())
	}
	// Write functions
	for i := range state.NumFunctions() {
		writeIrFunction(state.Function(i))
	}
}

func writeIrMemory(kind string, name string) {
	fmt.Printf("%s %s(?) -> (?)\n", kind, name)
}

func writeIrFunction(f function.Boot) {
	var (
		name   = trace.ModuleName{Name: f.Name(), Multiplier: 1}
		regMap = register.ArrayMap(name, f.Registers()...)
	)
	fmt.Printf("fn %s(", f.Name())
	// parameters
	writeIrFunctionArgs(register.INPUT_REGISTER, f.Registers())
	//
	fmt.Printf(")")
	//
	if f.NumOutputs() != 0 {
		//
		fmt.Printf(" -> (")
		// returns
		writeIrFunctionArgs(register.OUTPUT_REGISTER, f.Registers())
		//
		fmt.Printf(")")
	}
	//
	fmt.Println(" {")
	//
	writeIrFunctionVariables(f)
	//
	for pc, insn := range f.Code() {
		fmt.Printf("[%d]\t%s\n", pc, insn.String(regMap))
	}
	// Done
	fmt.Println("}")
}

func writeIrFunctionArgs(kind register.Type, regs []register.Register) {
	var first = true
	//
	for _, r := range regs {
		if r.Kind() == kind {
			if !first {
				fmt.Printf(", ")
			} else {
				first = false
			}
			//
			fmt.Printf("u%d %s", r.Width(), r.Name())
		}
	}
}

func writeIrFunctionVariables(f function.Boot) {
	for _, r := range f.Registers() {
		if !r.IsInputOutput() {
			fmt.Printf("\tu%d %s\n", r.Width(), r.Name())
		}
	}
}

// ============================================================================
// Misc
// ============================================================================

//nolint:errcheck
func init() {
	rootCmd.AddCommand(compileCmd)
	compileCmd.Flags().Bool("ast", false, "Output abstract syntax tree (AST)")
	compileCmd.Flags().Bool("ir", false, "Output intermediate representation (IR)")
}
