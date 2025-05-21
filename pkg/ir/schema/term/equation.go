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
package term

import (
	"github.com/consensys/go-corset/pkg/ir/schema"
	"github.com/consensys/go-corset/pkg/trace"
	"github.com/consensys/go-corset/pkg/util"
	"github.com/consensys/go-corset/pkg/util/collection/set"
	"github.com/consensys/go-corset/pkg/util/source/sexp"
)

const (
	// EQUALS indicates an equals (==) relationship
	EQUALS uint8 = 0
	// NOT_EQUALS indicates a not-equals (!=) relationship
	NOT_EQUALS uint8 = 1
	// LESS_THAN indicates a less-than (<) relationship
	LESS_THAN uint8 = 2
	// LESS_THAN_EQUALS indicates a less-than-or-equals (<=) relationship
	LESS_THAN_EQUALS uint8 = 3
	// GREATER_THAN indicates a greater-than (>) relationship
	GREATER_THAN uint8 = 4
	// GREATER_THAN_EQUALS indicates a greater-than-or-equals (>=) relationship
	GREATER_THAN_EQUALS uint8 = 5
)

// Equation represents an equation between two terms (e.g. "X==Y", or "X!=Y+1",
// etc).  Equations are either equalities (or negated equalities) or
// inequalities.
type Equation[T schema.Term[T]] struct {
	Kind uint8
	Lhs  schema.Term[T]
	Rhs  schema.Term[T]
}

// Bounds implementation for Boundable interface.
func (p *Equation[T]) Bounds() util.Bounds {
	panic("todo")
}

// Branches implementation for Evaluable interface.
func (p *Equation[T]) Branches() uint {
	panic("todo")
}

// Context implementation for Contextual interface.
func (p *Equation[T]) Context(module schema.Module) trace.Context {
	panic("todo")
}

func (p *Equation[T]) TestAt(k int, tr trace.Module) (bool, uint, error) {
	lhs, err1 := p.Lhs.EvalAt(k, tr)
	rhs, err2 := p.Rhs.EvalAt(k, tr)
	// error check
	if err1 != nil {
		return false, 0, err1
	} else if err2 != nil {
		return false, 0, err2
	}
	// perform comparison
	c := lhs.Cmp(&rhs)
	//
	switch p.Kind {
	case EQUALS:
		return c == 0, 0, nil
	case NOT_EQUALS:
		return c != 0, 0, nil
	case LESS_THAN:
		return c < 0, 0, nil
	case LESS_THAN_EQUALS:
		return c <= 0, 0, nil
	case GREATER_THAN_EQUALS:
		return c >= 0, 0, nil
	case GREATER_THAN:
		return c > 0, 0, nil
	}
	// failure
	panic("unreachable")
}

// Lisp returns a lisp representation of this equation, which is useful for
// debugging.
func (e Equation[T]) Lisp(module schema.Module) sexp.SExp {
	var (
		symbol string
		l      = e.Lhs.Lisp(module)
		r      = e.Rhs.Lisp(module)
	)
	//
	switch e.Kind {
	case EQUALS:
		symbol = "=="
	case NOT_EQUALS:
		symbol = "!="
	case LESS_THAN:
		symbol = "<"
	case LESS_THAN_EQUALS:
		symbol = "<="
	case GREATER_THAN:
		symbol = ">"
	case GREATER_THAN_EQUALS:
		symbol = ">="
	default:
		panic("unreachable")
	}
	//
	return sexp.NewList([]sexp.SExp{
		sexp.NewSymbol(symbol), l, r})
}

// RequiredColumns implementation for Contextual interface.
func (p *Equation[T]) RequiredColumns() *set.SortedSet[uint] {
	panic("todo")
}

// RequiredCells implementation for Contextual interface
func (p *Equation[T]) RequiredCells(row int, tr trace.Module) *set.AnySortedSet[trace.CellRef] {
	panic("todo")
}

// Simplify this equation as much as reasonably possible.
func (e Equation[T]) Simplify() Equation[T] {
	panic("todo")
}

// Negate a given equation
func (e Equation[T]) Negate() Equation[T] {
	var kind uint8
	//
	switch e.Kind {
	case EQUALS:
		kind = NOT_EQUALS
	case NOT_EQUALS:
		kind = EQUALS
	case LESS_THAN:
		kind = GREATER_THAN_EQUALS
	case LESS_THAN_EQUALS:
		kind = GREATER_THAN
	case GREATER_THAN_EQUALS:
		kind = LESS_THAN
	case GREATER_THAN:
		kind = LESS_THAN_EQUALS
	}
	//
	return Equation[T]{kind, e.Lhs, e.Rhs}
}

// Is determines whether or not this equation is known to evaluate to true or
// false.  For example, "0 == 0" evaluates to true, whilst "0 != 0" evaluates to
// false.
func (e Equation[T]) Is(val bool) bool {
	// Attempt to disprove non-equality
	lc, l_ok := e.Lhs.(*Constant[T])
	rc, r_ok := e.Rhs.(*Constant[T])
	//
	if l_ok && r_ok {
		var (
			cmp  = lc.Value.Cmp(&rc.Value)
			sign bool
		)
		//
		switch e.Kind {
		case EQUALS:
			sign = (cmp == 0)
		case NOT_EQUALS:
			sign = (cmp != 0)
		case LESS_THAN:
			sign = (cmp < 0)
		case LESS_THAN_EQUALS:
			sign = (cmp <= 0)
		case GREATER_THAN:
			sign = (cmp > 0)
		case GREATER_THAN_EQUALS:
			sign = (cmp >= 0)
		default:
			panic("unreachable")
		}
		//
		return val == sign
	}
	// Give up
	return false
}
