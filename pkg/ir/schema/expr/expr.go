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
package expr

import (
	"github.com/consensys/gnark-crypto/ecc/bls12-377/fr"
	"github.com/consensys/go-corset/pkg/ir/schema"
	"github.com/consensys/go-corset/pkg/trace"
	"github.com/consensys/go-corset/pkg/util"
	"github.com/consensys/go-corset/pkg/util/collection/set"
	"github.com/consensys/go-corset/pkg/util/source/sexp"
)

// ============================================================================

type Expr[T schema.Term[T]] struct {
	// Term to be evaluated, etc.
	Term T
}

// AsConstant deschema.Termines whether or not this is a constant expression.  If
// so, the constant is returned; otherwise, nil is returned.  NOTE: this
// does not perform any form of simplification to deschema.Termine this.
func (e Expr[T]) AsConstant() *fr.Element {
	panic("todo")
}

// Context deschema.Termines the evaluation context (i.e. enclosing module) for this
func (e Expr[T]) Context(module schema.Module) trace.Context {
	return e.Term.Context(module)
}

// Bounds returns max shift in either the negative (left) or positive
// direction (right).
func (e Expr[T]) Bounds() util.Bounds { return e.Term.Bounds() }

// Lisp converts this schema element into a simple S-schema.Termession, for example
// so it can be printed.
func (e Expr[T]) Lisp(module schema.Module) sexp.SExp {
	return e.Term.Lisp(module)
}

// RequiredColumns returns the set of columns on which this schema.Term depends.
// That is, columns whose values may be accessed when evaluating this schema.Term
// on a given trace.
func (e Expr[T]) RequiredColumns() *set.SortedSet[uint] {
	return e.Term.RequiredColumns()
}

// RequiredCells returns the set of trace cells on which this schema.Term depends.
// That is, evaluating this schema.Term at the given row in the given trace will read
// these cells.
func (e Expr[T]) RequiredCells(row int, tr trace.Module) *set.AnySortedSet[trace.CellRef] {
	return e.Term.RequiredCells(row, tr)
}

// EvalAt evaluates a column access at a given row in a trace, which returns the
// value at that row of the column in question or nil is that row is
// out-of-bounds.
func (e Expr[T]) EvalAt(k int, tr trace.Module) (fr.Element, error) {
	return e.Term.EvalAt(k, tr)
}

// Shift all column accesses within the expression by a given amount.
func (e Expr[T]) Shift(shift int) Expr[T] {
	return Expr[T]{e.Term.ApplyShift(shift)}
}

// TestAt evaluates this expression in a given tabular context and checks it
// against zero. Observe that if this expression is *undefined* within this
// context then it returns "nil".  An expression can be undefined for
// several reasons: firstly, if it accesses a row which does not exist (e.g.
// at index -1); secondly, if it accesses a column which does not exist.
func (e Expr[T]) TestAt(k int, tr trace.Module) (bool, uint, error) {
	val, err := e.Term.EvalAt(k, tr)
	//
	return val.IsZero(), 0, err
}

// Branches returns the number of unique evaluation paths through the given
// constraint.
func (e Expr[T]) Branches() uint {
	// NOTE: currently branch coverage is not supported at the AIR level.
	return 1
}
