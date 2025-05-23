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
	"fmt"

	"github.com/consensys/gnark-crypto/ecc/bls12-377/fr"
	"github.com/consensys/go-corset/pkg/ir/schema"
	"github.com/consensys/go-corset/pkg/trace"
	"github.com/consensys/go-corset/pkg/util"
	"github.com/consensys/go-corset/pkg/util/collection/set"
	"github.com/consensys/go-corset/pkg/util/source/sexp"
)

// Exp represents the a given value taken to a power.
type Exp[T schema.Term[T]] struct {
	Arg T
	Pow uint64
}

// ApplyShift implementation for Term interface.
func (p *Exp[T]) ApplyShift(int) T {
	panic("todo")
}

// Bounds implementation for Boundable interface.
func (p *Exp[T]) Bounds() util.Bounds {
	panic("todo")
}

// Branches implementation for Evaluable interface.
func (p *Exp[T]) Branches() uint {
	panic("todo")
}

// Context implementation for Contextual interface.
func (p *Exp[T]) Context(module schema.Module) trace.Context {
	return p.Arg.Context(module)
}

// EvalAt implementation for Evaluable interface.
func (p *Exp[T]) EvalAt(k int, tr trace.Module) (fr.Element, error) {
	// Check whether argument evaluates to zero or not.
	val, err := p.Arg.EvalAt(k, tr)
	// Compute exponent
	util.Pow(&val, p.Pow)
	// Done
	return val, err
}

// Lisp implementation for Lispifiable interface.
func (p *Exp[T]) Lisp(module schema.Module) sexp.SExp {
	arg := p.Arg.Lisp(module)
	pow := sexp.NewSymbol(fmt.Sprintf("%d", p.Pow))

	return sexp.NewList([]sexp.SExp{sexp.NewSymbol("^"), arg, pow})
}

// RequiredColumns implementation for Contextual interface.
func (p *Exp[T]) RequiredColumns() *set.SortedSet[uint] {
	return p.Arg.RequiredColumns()
}

// RequiredCells implementation for Contextual interface
func (p *Exp[T]) RequiredCells(row int, tr trace.Module) *set.AnySortedSet[trace.CellRef] {
	return p.Arg.RequiredCells(row, tr)
}

// ShiftRange implementation for Term interface.
func (p *Exp[T]) ShiftRange() (int, int) {
	return p.Arg.ShiftRange()
}

// Simplify implementation for Term interface.
func (p *Exp[T]) Simplify(casts bool) T {
	panic("todo")
}

// ValueRange implementation for Term interface.
func (p *Exp[T]) ValueRange(module schema.Module) *util.Interval {
	panic("todo")
}
