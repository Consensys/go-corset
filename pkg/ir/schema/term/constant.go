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
	"math"

	"github.com/consensys/gnark-crypto/ecc/bls12-377/fr"
	"github.com/consensys/go-corset/pkg/ir/schema"
	"github.com/consensys/go-corset/pkg/trace"
	"github.com/consensys/go-corset/pkg/util"
	"github.com/consensys/go-corset/pkg/util/collection/set"
	"github.com/consensys/go-corset/pkg/util/source/sexp"
)

// Constant represents a constant value within an expression.
type Constant[T schema.Term[T]] struct{ Value fr.Element }

// Air indicates this term can be used at the AIR level.
func (p *Constant[T]) Air() {}

// ApplyShift implementation for Term interface.
func (p *Constant[T]) ApplyShift(int) T {
	panic("todo")
}

// Bounds implementation for Boundable interface.
func (p *Constant[T]) Bounds() util.Bounds {
	return util.EMPTY_BOUND
}

// Branches implementation for Evaluable interface.
func (p *Constant[T]) Branches() uint {
	panic("todo")
}

// Context implementation for Contextual interface.
func (p *Constant[T]) Context(module schema.Module) trace.Context {
	return trace.VoidContext[uint]()
}

// EvalAt implementation for Evaluable interface.
func (p *Constant[T]) EvalAt(k int, module trace.Module) (fr.Element, error) {
	return p.Value, nil
}

// Lisp implementation for Lispifiable interface.
func (p *Constant[T]) Lisp(module schema.Module) sexp.SExp {
	return sexp.NewSymbol(p.Value.String())
}

// RequiredColumns implementation for Contextual interface.
func (p *Constant[T]) RequiredColumns() *set.SortedSet[uint] {
	return set.NewSortedSet[uint]()
}

// RequiredCells implementation for Contextual interface
func (p *Constant[T]) RequiredCells(row int, tr trace.Module) *set.AnySortedSet[trace.CellRef] {
	return set.NewAnySortedSet[trace.CellRef]()
}

// ShiftRange implementation for Term interface.
func (p *Constant[T]) ShiftRange() (int, int) {
	return math.MaxInt, math.MinInt
}

// Simplify implementation for Term interface.
func (p *Constant[T]) Simplify(casts bool) T {
	panic("todo")
}

// ValueRange implementation for Term interface.
func (p *Constant[T]) ValueRange(module schema.Module) *util.Interval {
	panic("todo")
}
