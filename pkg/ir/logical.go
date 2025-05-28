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
package ir

import (
	"github.com/consensys/go-corset/pkg/schema"
	"github.com/consensys/go-corset/pkg/trace"
	"github.com/consensys/go-corset/pkg/util"
	"github.com/consensys/go-corset/pkg/util/collection/set"
	"github.com/consensys/go-corset/pkg/util/source/sexp"
)

// Conjunction builds the logical conjunction (i.e. and) for a given set of constraints.
func Conjunction[T LogicalTerm[T]](terms ...Logical[T]) Logical[T] {
	panic("todo")
}

// Disjunction creates a constraint representing the disjunction of a given set of
// constraints.
func Disjunction[T LogicalTerm[T]](terms ...Logical[T]) Logical[T] {
	panic("todo")
}

// Equals constructs an equation representing the equality of two expressions.
func Equals[S LogicalTerm[S], T Term[T]](lhs Expr[T], rhs Expr[T]) Logical[S] {
	var term LogicalTerm[S] = &Equation[T]{
		Kind: EQUALS,
		Lhs:  lhs.Term,
		Rhs:  rhs.Term,
	}
	//
	return Logical[S]{Term: term.(S)}
}

// GreaterThan constructs an equation representing the inequality of two
// expressions.
func GreaterThan[S LogicalTerm[S], T Term[T]](lhs Expr[T], rhs Expr[T]) Logical[S] {
	var term LogicalTerm[S] = &Equation[T]{
		Kind: GREATER_THAN,
		Lhs:  lhs.Term,
		Rhs:  rhs.Term,
	}
	//
	return Logical[S]{Term: term.(S)}
}

// GreaterThanOrEquals constructs an equation representing the inequality of two
// expressions.
func GreaterThanOrEquals[S LogicalTerm[S], T Term[T]](lhs Expr[T], rhs Expr[T]) Logical[S] {
	var term LogicalTerm[S] = &Equation[T]{
		Kind: GREATER_THAN_EQUALS,
		Lhs:  lhs.Term,
		Rhs:  rhs.Term,
	}
	//
	return Logical[S]{Term: term.(S)}
}

// LessThan constructs an equation representing the inequality of two
// expressions.
func LessThan[S LogicalTerm[S], T Term[T]](lhs Expr[T], rhs Expr[T]) Logical[S] {
	var term LogicalTerm[S] = &Equation[T]{
		Kind: LESS_THAN,
		Lhs:  lhs.Term,
		Rhs:  rhs.Term,
	}
	//
	return Logical[S]{Term: term.(S)}
}

// LessThanOrEquals constructs an equation representing the inequality of two
// expressions.
func LessThanOrEquals[S LogicalTerm[S], T Term[T]](lhs Expr[T], rhs Expr[T]) Logical[S] {
	var term LogicalTerm[S] = &Equation[T]{
		Kind: LESS_THAN_EQUALS,
		Lhs:  lhs.Term,
		Rhs:  rhs.Term,
	}
	//
	return Logical[S]{Term: term.(S)}
}

// Negate constructs the logical negation of the given Logical[T].
func Negate[T LogicalTerm[T]](expr Logical[T]) Logical[T] {
	panic("todo")
}

// NotEquals constructs an equation representing the non-equality of two
// expressions.
func NotEquals[S LogicalTerm[S], T Term[T]](lhs Expr[T], rhs Expr[T]) Logical[S] {
	var term LogicalTerm[S] = &Equation[T]{
		Kind: NOT_EQUALS,
		Lhs:  lhs.Term,
		Rhs:  rhs.Term,
	}
	//
	return Logical[S]{Term: term.(S)}
}

// ============================================================================

// Logical encapsulates the notion of a "logical expression".  That is something
// which can be tested for truthhood or falsehood.  Logical expressions may use
// arithmetic terms internally (e.g. an equation may compare two arithmetic
// terms), but it is expected to only produce "true" or "false".
type Logical[T LogicalTerm[T]] struct {
	Term T
}

// Bounds returns max shift in either the negative (left) or positive
// direction (right).
func (e Logical[T]) Bounds() util.Bounds { return e.Term.Bounds() }

// Lisp converts this schema element into a simple S-schema.Termession, for example
// so it can be printed.
func (e Logical[T]) Lisp(module schema.Module) sexp.SExp {
	return e.Term.Lisp(module)
}

// RequiredRegisters returns the set of registers on which this schema.Term depends.
// That is, registers whose values may be accessed when evaluating this schema.Term
// on a given trace.
func (e Logical[T]) RequiredRegisters() *set.SortedSet[uint] {
	return e.Term.RequiredRegisters()
}

// RequiredCells returns the set of trace cells on which this schema.Term depends.
// That is, evaluating this schema.Term at the given row in the given trace will read
// these cells.
func (e Logical[T]) RequiredCells(row int, tr trace.Module) *set.AnySortedSet[trace.CellRef] {
	return e.Term.RequiredCells(row, tr)
}

// TestAt implementation for the Testable interface.
func (e Logical[T]) TestAt(k int, tr trace.Module) (bool, uint, error) {
	return e.Term.TestAt(k, tr)
}
