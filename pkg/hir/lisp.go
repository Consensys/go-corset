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
package hir

import (
	"fmt"
	"reflect"

	"github.com/consensys/go-corset/pkg/util/source/sexp"
)

func lispOfTerm(e Term, schema *Schema) sexp.SExp {
	switch e := e.(type) {
	case *Add:
		return nary2Lisp(schema, "+", e.Args...)
	case *Cast:
		return lispOfCast(e, schema)
	case *Constant:
		return sexp.NewSymbol(e.Value.String())
	case *ColumnAccess:
		return lispOfColumnAccess(e, schema)
	case *Equation:
		if e.Sign {
			return nary2Lisp(schema, "==", e.Lhs, e.Rhs)
		}
		//
		return nary2Lisp(schema, "!=", e.Lhs, e.Rhs)
	case *Exp:
		return lispOfExp(e, schema)
	case *IfZero:
		return lispOfIfZero(e, schema)
	case *LabelledConstant:
		lab := fmt.Sprintf("%s:%s", e.Label, e.Value.String())
		return sexp.NewSymbol(lab)
	case *List:
		return nary2Lisp(schema, "begin", e.Args...)
	case *Mul:
		return nary2Lisp(schema, "*", e.Args...)
	case *Norm:
		return lispOfNormalise(e, schema)
	case *Sub:
		return nary2Lisp(schema, "-", e.Args...)
	default:
		name := reflect.TypeOf(e).Name()
		panic(fmt.Sprintf("unknown HIR expression \"%s\"", name))
	}
}

func lispOfColumnAccess(e *ColumnAccess, schema *Schema) sexp.SExp {
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

// Lisp converts this schema element into a simple S-Expression, for example
// so it can be printed.
func lispOfIfZero(e *IfZero, schema *Schema) sexp.SExp {
	// Translate Condition
	condition := lispOfTerm(e.Condition, schema)
	// Dispatch on type
	if e.FalseBranch == nil {
		return sexp.NewList([]sexp.SExp{
			sexp.NewSymbol("if"),
			condition,
			lispOfTerm(e.TrueBranch, schema),
		})
	} else if e.TrueBranch == nil {
		return sexp.NewList([]sexp.SExp{
			sexp.NewSymbol("ifnot"),
			condition,
			lispOfTerm(e.FalseBranch, schema),
		})
	}

	return sexp.NewList([]sexp.SExp{
		sexp.NewSymbol("if"),
		condition,
		lispOfTerm(e.TrueBranch, schema),
		lispOfTerm(e.FalseBranch, schema),
	})
}

func lispOfNormalise(e *Norm, schema *Schema) sexp.SExp {
	arg := lispOfTerm(e.Arg, schema)
	return sexp.NewList([]sexp.SExp{sexp.NewSymbol("~"), arg})
}

func lispOfCast(e *Cast, schema *Schema) sexp.SExp {
	arg := lispOfTerm(e.Arg, schema)
	name := sexp.NewSymbol(fmt.Sprintf(":u%d", e.BitWidth))

	return sexp.NewList([]sexp.SExp{name, arg})
}

func lispOfExp(e *Exp, schema *Schema) sexp.SExp {
	arg := lispOfTerm(e.Arg, schema)
	pow := sexp.NewSymbol(fmt.Sprintf("%d", e.Pow))

	return sexp.NewList([]sexp.SExp{sexp.NewSymbol("^"), arg, pow})
}

func nary2Lisp(schema *Schema, op string, exprs ...Term) sexp.SExp {
	arr := make([]sexp.SExp, 1+len(exprs))
	arr[0] = sexp.NewSymbol(op)
	// Translate arguments
	for i, e := range exprs {
		arr[i+1] = lispOfTerm(e, schema)
	}
	// Done
	return sexp.NewList(arr)
}
