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
	"github.com/consensys/gnark-crypto/ecc/bls12-377/fr"
	"github.com/consensys/go-corset/pkg/ir/schema"
	"github.com/consensys/go-corset/pkg/trace"
	"github.com/consensys/go-corset/pkg/util"
	"github.com/consensys/go-corset/pkg/util/collection/set"
	"github.com/consensys/go-corset/pkg/util/source/sexp"
)

// Norm reduces the value of an expression to either zero (if it was zero)
// or one (otherwise).
type Norm[T schema.Term[T]] struct{ Arg T }

// ApplyShift implementation for Term interface.
func (p *Norm[T]) ApplyShift(int) T {
	panic("todo")
}

// Bounds implementation for Boundable interface.
func (p *Norm[T]) Bounds() util.Bounds {
	panic("todo")
}

// Branches implementation for Evaluable interface.
func (p *Norm[T]) Branches() uint {
	panic("todo")
}

// Context implementation for Contextual interface.
func (p *Norm[T]) Context(module schema.Module) trace.Context {
	return p.Arg.Context(module)
}

// EvalAt implementation for Evaluable interface.
func (p *Norm[T]) EvalAt(k int, tr trace.Module) (fr.Element, error) {
	// Check whether argument evaluates to zero or not.
	val, err := p.Arg.EvalAt(k, tr)
	// Normalise value (if necessary)
	if !val.IsZero() {
		val.SetOne()
	}
	// Done
	return val, err
}

// Lisp implementation for Lispifiable interface.
func (p *Norm[T]) Lisp(module schema.Module) sexp.SExp {
	arg := p.Arg.Lisp(module)
	return sexp.NewList([]sexp.SExp{sexp.NewSymbol("~"), arg})
}

// RequiredColumns implementation for Contextual interface.
func (p *Norm[T]) RequiredColumns() *set.SortedSet[uint] {
	return p.Arg.RequiredColumns()
}

// RequiredCells implementation for Contextual interface
func (p *Norm[T]) RequiredCells(row int, tr trace.Module) *set.AnySortedSet[trace.CellRef] {
	return p.Arg.RequiredCells(row, tr)
}

// ShiftRange implementation for Term interface.
func (p *Norm[T]) ShiftRange() (int, int) {
	return p.Arg.ShiftRange()
}

// Simplify implementation for Term interface.
func (p *Norm[T]) Simplify(casts bool) T {
	panic("todo")
}

// ValueRange implementation for Term interface.
func (p *Norm[T]) ValueRange(module schema.Module) *util.Interval {
	panic("todo")
}
