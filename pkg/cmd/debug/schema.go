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

	"github.com/consensys/go-corset/pkg/ir/mir"
	"github.com/consensys/go-corset/pkg/ir/schema"
	"github.com/consensys/go-corset/pkg/util/source/sexp"
)

// PrintSchemas is responsible for printing out a human-readable description of
// a given schema.
func PrintSchemas(mirSchema mir.Schema, mirEnable bool, airEnable bool,
	optConfig mir.OptimisationConfig, textwidth uint) {
	//
	if mirEnable {
		printSchema(mirSchema, textwidth)
	}

	if airEnable {
		printSchema(mir.LowerToAir(&mirSchema, optConfig), textwidth)
	}
}

// Print out all declarations included in a given
func printSchema[M schema.Module, C schema.Constraint](schema schema.Schema[M, C], width uint) {
	formatter := sexp.NewFormatter(width)
	formatter.Add(&sexp.SFormatter{Head: "if", Priority: 0})
	formatter.Add(&sexp.SFormatter{Head: "ifnot", Priority: 0})
	formatter.Add(&sexp.LFormatter{Head: "begin", Priority: 1})
	formatter.Add(&sexp.LFormatter{Head: "∧", Priority: 1})
	formatter.Add(&sexp.LFormatter{Head: "∨", Priority: 1})
	formatter.Add(&sexp.LFormatter{Head: "+", Priority: 2})
	formatter.Add(&sexp.LFormatter{Head: "*", Priority: 3})
	//
	for i := schema.Declarations(); i.HasNext(); {
		ith := i.Next()
		text := formatter.Format(ith.Lisp(schema))
		fmt.Println(text)
	}

	for i := schema.Constraints(); i.HasNext(); {
		ith := i.Next()
		text := formatter.Format(ith.Lisp(schema))
		fmt.Println(text)
	}

	for i := schema.Assertions(); i.HasNext(); {
		ith := i.Next()
		text := formatter.Format(ith.Lisp(schema))
		fmt.Println(text)
	}
}
