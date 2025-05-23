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

	cmd_util "github.com/consensys/go-corset/pkg/cmd/util"
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
	// Print out each module, one by one.
	for i := schema.Modules(); i.HasNext(); {
		printModule(i.Next(), schema, width)
	}
}

func printModule(module schema.Module, schema schema.AnySchema, width uint) {
	formatter := sexp.NewFormatter(width)
	formatter.Add(&sexp.SFormatter{Head: "if", Priority: 0})
	formatter.Add(&sexp.SFormatter{Head: "ifnot", Priority: 0})
	formatter.Add(&sexp.LFormatter{Head: "begin", Priority: 1})
	formatter.Add(&sexp.LFormatter{Head: "∧", Priority: 1})
	formatter.Add(&sexp.LFormatter{Head: "∨", Priority: 1})
	formatter.Add(&sexp.LFormatter{Head: "+", Priority: 2})
	formatter.Add(&sexp.LFormatter{Head: "*", Priority: 3})

	if module.Name() == "" {
		fmt.Printf("(module)\n")
	} else {
		fmt.Printf("(module %s)\n", module.Name())
	}

	for _, r := range module.Registers() {
		if r.IsInput() {
			fmt.Printf("(input %s u%d)\n", r.Name, r.Width)
		} else if r.IsOutput() {
			fmt.Printf("(output %s u%d)\n", r.Name, r.Width)
		} else if r.IsComputed() {
			fmt.Printf("(computed %s u%d)\n", r.Name, r.Width)
		} else {
			// Fallback --- unsure what kind this is.
			fmt.Printf("(column %s u%d)\n", r.Name, r.Width)
		}
	}

	for i := module.Constraints(); i.HasNext(); {
		ith := i.Next()
		text := formatter.Format(ith.Lisp(schema))
		fmt.Println(text)
	}
}
