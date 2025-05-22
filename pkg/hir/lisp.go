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

	sc "github.com/consensys/go-corset/pkg/schema"
	"github.com/consensys/go-corset/pkg/util/source/sexp"
)

func lispOfTerm(e Term, module sc.Module) sexp.SExp {
	switch e := e.(type) {
	case *Add:
		return nary2Lisp(module, "+", e.Args...)
	case *Cast:
		return lispOfCast(e, module)
	case *Connective:
		if e.Sign {
			return nary2Lisp(module, "∨", e.Args...)
		}
		//
		return nary2Lisp(module, "∧", e.Args...)
	case *Constant:
		return sexp.NewSymbol(e.Value.String())
	case *ColumnAccess:
		return lispOfColumnAccess(e, module)
	case *Equation:
		return lispOfEquation(e, module)
	case *Exp:
		return lispOfExp(e, module)
	case *IfZero:
		return lispOfIfZero(e, module)
	case *LabelledConstant:
		lab := fmt.Sprintf("%s:%s", e.Label, e.Value.String())
		return sexp.NewSymbol(lab)
	case *List:
		return nary2Lisp(module, "begin", e.Args...)
	case *Mul:
		return nary2Lisp(module, "*", e.Args...)
	case *Norm:
		return lispOfNormalise(e, module)
	case *Not:
		return nary2Lisp(module, "¬", e.Arg)
	case *Sub:
		return nary2Lisp(module, "-", e.Args...)
	default:
		name := reflect.TypeOf(e).Name()
		panic(fmt.Sprintf("unknown HIR expression \"%s\"", name))
	}
}

func lispOfColumnAccess(e *ColumnAccess, module sc.Module) sexp.SExp {
	var name string
	// Generate name, whilst allowing for schema to be nil.
	if module != nil {
		name = module.Columns().Nth(e.Column).QualifiedName(module)
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

func lispOfEquation(e *Equation, module sc.Module) sexp.SExp {
	var symbol string

	switch e.Kind {
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
		lispOfTerm(e.Lhs, module),
		lispOfTerm(e.Rhs, module)})
}

// Lisp converts this schema element into a simple S-Expression, for example
// so it can be printed.
func lispOfIfZero(e *IfZero, module sc.Module) sexp.SExp {
	// Translate Condition
	condition := lispOfTerm(e.Condition, module)
	// Dispatch on type
	if e.FalseBranch == nil {
		return sexp.NewList([]sexp.SExp{
			sexp.NewSymbol("if"),
			condition,
			lispOfTerm(e.TrueBranch, module),
		})
	} else if e.TrueBranch == nil {
		return sexp.NewList([]sexp.SExp{
			sexp.NewSymbol("ifnot"),
			condition,
			lispOfTerm(e.FalseBranch, module),
		})
	}

	return sexp.NewList([]sexp.SExp{
		sexp.NewSymbol("if"),
		condition,
		lispOfTerm(e.TrueBranch, module),
		lispOfTerm(e.FalseBranch, module),
	})
}

func lispOfNormalise(e *Norm, module sc.Module) sexp.SExp {
	arg := lispOfTerm(e.Arg, module)
	return sexp.NewList([]sexp.SExp{sexp.NewSymbol("~"), arg})
}

func lispOfCast(e *Cast, module sc.Module) sexp.SExp {
	arg := lispOfTerm(e.Arg, module)
	name := sexp.NewSymbol(fmt.Sprintf(":u%d", e.BitWidth))

	return sexp.NewList([]sexp.SExp{name, arg})
}

func lispOfExp(e *Exp, module sc.Module) sexp.SExp {
	arg := lispOfTerm(e.Arg, module)
	pow := sexp.NewSymbol(fmt.Sprintf("%d", e.Pow))

	return sexp.NewList([]sexp.SExp{sexp.NewSymbol("^"), arg, pow})
}

func nary2Lisp(module sc.Module, op string, exprs ...Term) sexp.SExp {
	arr := make([]sexp.SExp, 1+len(exprs))
	arr[0] = sexp.NewSymbol(op)
	// Translate arguments
	for i, e := range exprs {
		arr[i+1] = lispOfTerm(e, module)
	}
	// Done
	return sexp.NewList(arr)
}
