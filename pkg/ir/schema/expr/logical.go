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
	"github.com/consensys/go-corset/pkg/ir/schema"
	"github.com/consensys/go-corset/pkg/trace"
	"github.com/consensys/go-corset/pkg/util"
	"github.com/consensys/go-corset/pkg/util/collection/set"
	"github.com/consensys/go-corset/pkg/util/source/sexp"
)

// ============================================================================

type Logical[T schema.LogicalTerm[T]] struct {
	Term T
}

// Bounds returns max shift in either the negative (left) or positive
// direction (right).
func (e Logical[T]) Bounds() util.Bounds { return e.Term.Bounds() }

// Branches returns the number of unique evaluation paths through the given
// constraint.
func (e Logical[T]) Branches() uint {
	panic("todo")
}

// Context determines the evaluation context (i.e. enclosing module) for this
func (e Logical[T]) Context(module schema.Module) trace.Context {
	return e.Term.Context(module)
}

// Lisp converts this schema element into a simple S-schema.Termession, for example
// so it can be printed.
func (e Logical[T]) Lisp(module schema.Module) sexp.SExp {
	return e.Term.Lisp(module)
}

// RequiredColumns returns the set of columns on which this schema.Term depends.
// That is, columns whose values may be accessed when evaluating this schema.Term
// on a given trace.
func (e Logical[T]) RequiredColumns() *set.SortedSet[uint] {
	return e.Term.RequiredColumns()
}

// RequiredCells returns the set of trace cells on which this schema.Term depends.
// That is, evaluating this schema.Term at the given row in the given trace will read
// these cells.
func (e Logical[T]) RequiredCells(row int, tr trace.Module) *set.AnySortedSet[trace.CellRef] {
	return e.Term.RequiredCells(row, tr)
}

func (e Logical[T]) TestAt(k int, tr trace.Module) (bool, uint, error) {
	return e.Term.TestAt(k, tr)
}
