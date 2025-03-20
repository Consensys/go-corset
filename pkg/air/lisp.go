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
package air

import (
	"fmt"
	"reflect"

	sc "github.com/consensys/go-corset/pkg/schema"
	"github.com/consensys/go-corset/pkg/util/source/sexp"
)

func lispOfTerm(e Term, schema sc.Schema) sexp.SExp {
	switch e := e.(type) {
	case *Add:
		return nary2Lisp(schema, "+", e.Args)
	case *Constant:
		return sexp.NewSymbol(e.Value.String())
	case *ColumnAccess:
		return lispOfColumnAccess(e, schema)
	case *Mul:
		return nary2Lisp(schema, "*", e.Args)
	case *Sub:
		return nary2Lisp(schema, "-", e.Args)
	default:
		name := reflect.TypeOf(e).Name()
		panic(fmt.Sprintf("unknown AIR expression \"%s\"", name))
	}
}

// Lisp converts this schema element into a simple S-Termession, for example
// so it can be printed.
func lispOfColumnAccess(e *ColumnAccess, schema sc.Schema) sexp.SExp {
	var name string
	// Generate name, whilst allowing for schema to be nil.
	if schema != nil {
		name = schema.Columns().Nth(e.Column).QualifiedName(schema)
	} else {
		name = fmt.Sprintf("#%d", e.Column)
	}
	//
	access := sexp.NewSymbol(name)
	// Check whether shifted (or not)
	if e.Shift == 0 {
		// Not shifted
		return access
	}
	// Shifted
	shift := sexp.NewSymbol(fmt.Sprintf("%d", e.Shift))

	return sexp.NewList([]sexp.SExp{sexp.NewSymbol("shift"), access, shift})
}

func nary2Lisp(schema sc.Schema, op string, exprs []Term) sexp.SExp {
	arr := make([]sexp.SExp, 1+len(exprs))
	arr[0] = sexp.NewSymbol(op)
	// Translate arguments
	for i, e := range exprs {
		arr[i+1] = lispOfTerm(e, schema)
	}
	// Done
	return sexp.NewList(arr)
}
