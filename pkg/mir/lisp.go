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
package mir

import (
	"fmt"
	"reflect"

	sc "github.com/consensys/go-corset/pkg/schema"
	"github.com/consensys/go-corset/pkg/util/source/sexp"
)

func lispOfConjunction(schema sc.Schema, conjunction Constraint) sexp.SExp {
	switch len(conjunction.conjuncts) {
	case 0:
		return sexp.NewSymbol("⊤")
	case 1:
		return lispOfDisjunction(schema, conjunction.conjuncts[0])
	default:
		arr := make([]sexp.SExp, 1+len(conjunction.conjuncts))
		arr[0] = sexp.NewSymbol("∧")
		// Translate arguments
		for i, e := range conjunction.conjuncts {
			arr[i+1] = lispOfDisjunction(schema, e)
		}
		// Done
		return sexp.NewList(arr)
	}
}

func lispOfDisjunction(schema sc.Schema, disjunction Disjunction) sexp.SExp {
	switch len(disjunction.atoms) {
	case 0:
		return sexp.NewSymbol("⊥")
	case 1:
		return lispOfEquation(schema, disjunction.atoms[0])
	default:
		arr := make([]sexp.SExp, 1+len(disjunction.atoms))
		arr[0] = sexp.NewSymbol("∨")
		// Translate arguments
		for i, e := range disjunction.atoms {
			arr[i+1] = lispOfEquation(schema, e)
		}
		// Done
		return sexp.NewList(arr)
	}
}

func lispOfEquation(schema sc.Schema, e Equation) sexp.SExp {
	var symbol string

	switch e.kind {
	case EQUALS:
		symbol = "=="
	case NOT_EQUALS:
		symbol = "!="
	case LESS_THAN:
		symbol = "<"
	case LESS_THAN_EQUALS:
		symbol = "<="
	case GREATER_THAN_EQUALS:
		symbol = ">="
	case GREATER_THAN:
		symbol = ">"
	default:
		panic("unreachable")
	}
	//
	return sexp.NewList([]sexp.SExp{
		sexp.NewSymbol(symbol),
		lispOfTerm(e.lhs, schema),
		lispOfTerm(e.rhs, schema),
	})
}

func lispOfTerm(e Term, schema sc.Schema) sexp.SExp {
	switch e := e.(type) {
	case *Add:
		return lispOfTerms(schema, "+", e.Args)
	case *Cast:
		return lispOfCast(e, schema)
	case *Constant:
		return sexp.NewSymbol(e.Value.String())
	case *ColumnAccess:
		return lispOfColumnAccess(e, schema)
	case *Exp:
		return lispOfExp(e, schema)
	case *Mul:
		return lispOfTerms(schema, "*", e.Args)
	case *Norm:
		return lispOfNormalise(e, schema)
	case *Sub:
		return lispOfTerms(schema, "-", e.Args)
	default:
		name := reflect.TypeOf(e).Name()
		panic(fmt.Sprintf("unknown MIR expression \"%s\"", name))
	}
}

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

func lispOfTerms(schema sc.Schema, op string, exprs []Term) sexp.SExp {
	arr := make([]sexp.SExp, 1+len(exprs))
	arr[0] = sexp.NewSymbol(op)
	// Translate arguments
	for i, e := range exprs {
		arr[i+1] = lispOfTerm(e, schema)
	}
	// Done
	return sexp.NewList(arr)
}

func lispOfCast(e *Cast, schema sc.Schema) sexp.SExp {
	arg := lispOfTerm(e.Arg, schema)
	name := sexp.NewSymbol(fmt.Sprintf(":u%d", e.BitWidth))

	return sexp.NewList([]sexp.SExp{name, arg})
}

func lispOfNormalise(e *Norm, schema sc.Schema) sexp.SExp {
	arg := lispOfTerm(e.Arg, schema)
	return sexp.NewList([]sexp.SExp{sexp.NewSymbol("~"), arg})
}

func lispOfExp(e *Exp, schema sc.Schema) sexp.SExp {
	arg := lispOfTerm(e.Arg, schema)
	pow := sexp.NewSymbol(fmt.Sprintf("%d", e.Pow))

	return sexp.NewList([]sexp.SExp{sexp.NewSymbol("^"), arg, pow})
}
