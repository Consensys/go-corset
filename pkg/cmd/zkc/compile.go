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
	"github.com/consensys/go-corset/pkg/zkc/compiler/ast/data"
	"github.com/consensys/go-corset/pkg/zkc/compiler/ast/decl"
	"github.com/consensys/go-corset/pkg/zkc/compiler/ast/symbol"
	"github.com/consensys/go-corset/pkg/zkc/compiler/ast/variable"
	"github.com/consensys/go-corset/pkg/zkc/vm/function"
	"github.com/consensys/go-corset/pkg/zkc/vm/machine"
	"github.com/consensys/go-corset/pkg/zkc/vm/memory"
	"github.com/consensys/go-corset/pkg/zkc/vm/word"
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
	as := GetFlag(cmd, "ast")
	// Compile source files, or print errors
	program := CompileSourceFiles(args...)
	//
	if as {
		writeAbstractSyntaxTree(program)
	}
	//
	if ir {
		vm := program.Compile()
		writeIntermediateRepresentation[word.Uint](vm)
	}
}

// ============================================================================
// AST
// ============================================================================

func writeAbstractSyntaxTree(program ast.Program) {
	var (
		env = data.NewEnvironment(func(id symbol.Resolved) data.ResolvedType {
			return nil
		})
	)
	//
	for i, d := range program.Components() {
		if i != 0 {
			fmt.Println()
		}
		//
		writeDeclaration(d, env)
	}
}

func writeDeclaration(d decl.Resolved, env data.ResolvedEnvironment) {
	switch d := d.(type) {
	case *decl.ResolvedConstant:
		writeConstant(d, env)
	case *decl.ResolvedFunction:
		writeFunction(d, env)
	case *decl.ResolvedMemory:
		writeMemory(d, env)
	default:
		panic("unknown declaration encountered")
	}
}

func writeConstant(m *decl.ResolvedConstant, env data.ResolvedEnvironment) {
	var mapping = variable.ArrayMap[symbol.Resolved]()
	//
	fmt.Print("const ")
	// type
	fmt.Printf("%s ", m.DataType.String(env))
	// name
	fmt.Printf("%s = ", m.Name())
	// contents
	fmt.Println(m.ConstExpr.String(mapping))
}

func writeMemory(m *decl.ResolvedMemory, env data.ResolvedEnvironment) {
	switch m.Kind {
	case decl.PUBLIC_READ_ONLY_MEMORY:
		fmt.Printf("public input")
	case decl.PRIVATE_READ_ONLY_MEMORY:
		fmt.Printf("input")
	case decl.PUBLIC_WRITE_ONCE_MEMORY:
		fmt.Printf("public output")
	case decl.PRIVATE_WRITE_ONCE_MEMORY:
		fmt.Printf("output")
	case decl.PUBLIC_STATIC_MEMORY:
		fmt.Printf("public static")
	case decl.PRIVATE_STATIC_MEMORY:
		fmt.Printf("static")
	case decl.RANDOM_ACCESS_MEMORY:
		fmt.Printf("memory")
	}
	// address lines
	fmt.Printf(" %s(", m.Name())
	writeMemoryParams(m.Address, env)
	fmt.Printf(") -> (")
	writeMemoryParams(m.Data, env)
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

func writeMemoryParams(params []variable.ResolvedDescriptor, env data.ResolvedEnvironment) {
	for i, p := range params {
		if i > 0 {
			fmt.Printf(", ")
		}

		fmt.Printf("%s %s", p.DataType.String(env), p.Name)
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

func writeFunction(f *decl.ResolvedFunction, env data.ResolvedEnvironment) {
	fmt.Printf("fn %s(", f.Name())
	// parameters
	writeFunctionArgs(variable.PARAMETER, f.Variables, env)
	//
	fmt.Printf(") -> (")
	// returns
	writeFunctionArgs(variable.RETURN, f.Variables, env)
	//
	fmt.Println(") {")
	//
	writeFunctionVariables(f, env)
	//
	for pc, insn := range f.Code {
		fmt.Printf("[%d]\t%s\n", pc, insn.String(f))
	}
	// Done
	fmt.Println("}")
}

func writeFunctionArgs(kind variable.Kind, variables []variable.ResolvedDescriptor, env data.ResolvedEnvironment) {
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
			fmt.Printf("%s %s", r.DataType.String(env), r.Name)
		}
	}
}

func writeFunctionVariables(f *decl.ResolvedFunction, env data.ResolvedEnvironment) {
	for _, r := range f.Variables {
		if r.IsLocal() {
			fmt.Printf("\t%s %s\n", r.DataType.String(env), r.Name)
		}
	}
}

// ============================================================================
// Intermediate Representation (IR)
// ============================================================================

func writeIntermediateRepresentation[W word.Word[W]](machine *machine.Base[W]) {
	// Write memories
	for _, m := range machine.Modules() {
		switch m := m.(type) {
		case memory.Memory[W]:
			writeIrMemory(m)
		case *function.Boot[W]:
			writeIrFunction[W](m)
		}
	}
}

func writeIrMemory[W word.Word[W]](m memory.Memory[W]) {
	fmt.Printf("memory? %s(?) -> (?)\n", m.Name())
}

func writeIrFunction[W word.Word[W]](f *function.Boot[W]) {
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
	writeIrFunctionVariables[W](f)
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

func writeIrFunctionVariables[W word.Word[W]](f *function.Boot[W]) {
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
