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
)

// PrintSchemas is responsible for printing out a human-readable description of
// a given schema.
func PrintSchemas(hirSchema *hir.Schema, hir bool, mir bool, air bool, optConfig mir.OptimisationConfig) {
	if hir {
		printSchema(hirSchema)
	}

	if mir {
		printSchema(hirSchema.LowerToMir())
	}

	if air {
		printSchema(hirSchema.LowerToMir().LowerToAir(optConfig))
	}
}

// Print out all declarations included in a given
func printSchema(schema schema.Schema) {
	for i := schema.Declarations(); i.HasNext(); {
		ith := i.Next()
		fmt.Println(ith.Lisp(schema).String(true))
	}

	for i := schema.Constraints(); i.HasNext(); {
		ith := i.Next()
		fmt.Println(ith.Lisp(schema).String(true))
	}

	for i := schema.Assertions(); i.HasNext(); {
		ith := i.Next()
		fmt.Println(ith.Lisp(schema).String(true))
	}
}
