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

	"github.com/consensys/go-corset/pkg/hir"
	"github.com/consensys/go-corset/pkg/mir"
	"github.com/consensys/go-corset/pkg/schema"
	"github.com/consensys/go-corset/pkg/util/source/sexp"
)

// PrintSchemas is responsible for printing out a human-readable description of
// a given schema.
func PrintSchemas(hirSchema *hir.Schema, hir bool, mir bool, air bool,
	optConfig mir.OptimisationConfig, textwidth uint) {
	//
	if hir {
		printSchema(hirSchema, textwidth)
	}

	if mir {
		printSchema(hirSchema.LowerToMir(), textwidth)
	}

	if air {
		printSchema(hirSchema.LowerToMir().LowerToAir(optConfig), textwidth)
	}
}

// Print out all declarations included in a given
func printSchema(schema schema.Schema, width uint) {
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
