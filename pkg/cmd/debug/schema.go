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
	"math"

	"github.com/consensys/go-corset/pkg/asm"
	"github.com/consensys/go-corset/pkg/asm/io"
	cmd_util "github.com/consensys/go-corset/pkg/cmd/util"
	"github.com/consensys/go-corset/pkg/ir/mir"
	"github.com/consensys/go-corset/pkg/schema"
	"github.com/consensys/go-corset/pkg/util/source/sexp"
)

// PrintSchemas is responsible for printing out a human-readable description of
// a given schema.
func PrintSchemas(stack cmd_util.SchemaStack, textwidth uint) {
	//
	for _, schema := range stack.Schemas() {
		printSchema(schema, textwidth)
	}
}

// Print out all declarations included in a given
func printSchema(schema schema.AnySchema, width uint) {
	first := true
	// Print out each module, one by one.
	for i := schema.Modules(); i.HasNext(); {
		ith := i.Next()
		//
		if isEmptyModule(ith) {
			// Skip empty modules as they just clutter things up.  Typically,
			// for example, the root module is empty.
			continue
		} else if !first {
			fmt.Println()
		}
		//
		switch ith := ith.(type) {
		case *asm.MacroFunction:
			printAssemblyFunction(*ith)
		case *asm.MicroFunction:
			printAssemblyFunction(*ith)
		default:
			printModule(ith, schema, width)
		}
		//
		first = false
	}
}

// ==================================================================
// Legacy module
// ==================================================================

func printModule(module schema.Module, sc schema.AnySchema, width uint) {
	formatter := sexp.NewFormatter(width)
	formatter.Add(&sexp.SFormatter{Head: "if", Priority: 0})
	formatter.Add(&sexp.SFormatter{Head: "ifnot", Priority: 0})
	formatter.Add(&sexp.LFormatter{Head: "begin", Priority: 0})
	formatter.Add(&sexp.LFormatter{Head: "∧", Priority: 1})
	formatter.Add(&sexp.LFormatter{Head: "∨", Priority: 1})
	formatter.Add(&sexp.IFormatter{Head: "==", Priority: 2})
	formatter.Add(&sexp.IFormatter{Head: "!=", Priority: 2})
	formatter.Add(&sexp.IFormatter{Head: "+", Priority: 3})
	formatter.Add(&sexp.IFormatter{Head: "*", Priority: 4})

	if module.Name() == "" {
		fmt.Println("(module)")
	} else {
		fmt.Printf("(module %s)\n", module.Name())
	}
	//
	fmt.Println()
	// Print inputs / outputs
	printRegisters(module, "inputs", func(r schema.Register) bool { return r.IsInput() })
	printRegisters(module, "outputs", func(r schema.Register) bool { return r.IsOutput() })
	printRegisters(module, "computed", func(r schema.Register) bool { return r.IsComputed() })
	// Print computations
	for i := module.Assignments(); i.HasNext(); {
		ith := i.Next()
		text := formatter.Format(ith.Lisp(sc))
		fmt.Print(text)
	}
	// Print constraints
	for i := module.Constraints(); i.HasNext(); {
		ith := i.Next()
		text := formatter.Format(ith.Lisp(sc))
		//
		if requiresSpacing(ith) {
			fmt.Println()
		}
		//
		fmt.Print(text)
	}
}

func printRegisters(module schema.Module, prefix string, filter func(schema.Register) bool) {
	if countRegisters(module, filter) != 0 {
		//
		fmt.Printf("(%s\n", prefix)
		//
		for _, r := range module.Registers() {
			if filter(r) {
				if r.Width != math.MaxUint {
					fmt.Printf("   (%s u%d)\n", r.Name, r.Width)
				} else {
					fmt.Printf("   (%s 𝔽)\n", r.Name)
				}
			}
		}
		//
		fmt.Println(")")
		fmt.Println("")
	}
}

func countRegisters(module schema.Module, filter func(schema.Register) bool) uint {
	var count = uint(0)
	//
	for _, r := range module.Registers() {
		if filter(r) {
			count++
		}
	}
	//
	return count
}

func requiresSpacing(c schema.Constraint) bool {
	if c, ok := c.(mir.Constraint); ok {
		if _, ok := c.Unwrap().(mir.VanishingConstraint); ok {
			return ok
		}
	}
	//
	return false
}

func isEmptyModule(module schema.Module) bool {
	return len(module.Registers()) == 0 &&
		module.Constraints().Count() == 0 &&
		module.Assignments().Count() == 0
}

// ==================================================================
// Assembly Function
// ==================================================================

func printAssemblyFunction[T io.Instruction[T]](f io.Function[T]) {
	printAssemblySignature(f)
	printAssemblyRegisters(f)
	//
	for pc, insn := range f.Code() {
		fmt.Printf("[%d]\t%s\n", pc, insn.String(&f))
	}
	//
	fmt.Println("}")
}

func printAssemblySignature[T io.Instruction[T]](f io.Function[T]) {
	first := true
	//
	fmt.Printf("fn %s(", f.Name())
	//
	for _, r := range f.Registers() {
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
	for _, r := range f.Registers() {
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

func printAssemblyRegisters[T io.Instruction[T]](f io.Function[T]) {
	for _, r := range f.Registers() {
		if !r.IsInput() && !r.IsOutput() {
			fmt.Printf("\tvar %s u%d\n", r.Name, r.Width)
		}
	}
}
