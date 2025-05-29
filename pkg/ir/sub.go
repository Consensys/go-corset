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
	"github.com/consensys/gnark-crypto/ecc/bls12-377/fr"
	"github.com/consensys/go-corset/pkg/schema"
	"github.com/consensys/go-corset/pkg/trace"
	"github.com/consensys/go-corset/pkg/util"
	"github.com/consensys/go-corset/pkg/util/collection/set"
	"github.com/consensys/go-corset/pkg/util/source/sexp"
)

// Sub represents the subtraction over zero or more expressions.
type Sub[T Term[T]] struct{ Args []T }

// Subtract returns the subtraction of the subsequent expressions from the
// first.
func Subtract[T Term[T]](exprs ...T) T {
	var term Term[T] = &Sub[T]{exprs}
	return term.(T)
}

// Air indicates this term can be used at the AIR level.
func (p *Sub[T]) Air() {}

// ApplyShift implementation for Term interface.
func (p *Sub[T]) ApplyShift(shift int) T {
	var term Term[T] = &Sub[T]{applyShiftOfTerms(p.Args, shift)}
	return term.(T)
}

// Bounds implementation for Boundable interface.
func (p *Sub[T]) Bounds() util.Bounds { return util.BoundsForArray(p.Args) }

// EvalAt implementation for Evaluable interface.
func (p *Sub[T]) EvalAt(k int, tr trace.Module) (fr.Element, error) {
	// Evaluate first argument
	val, err := p.Args[0].EvalAt(k, tr)
	// Continue evaluating the rest
	for i := 1; err == nil && i < len(p.Args); i++ {
		var ith fr.Element
		// Evaluate ith argument
		ith, err = p.Args[i].EvalAt(k, tr)
		val.Sub(&val, &ith)
	}
	// Done
	return val, err
}

// Lisp implementation for Lispifiable interface.
func (p *Sub[T]) Lisp(module schema.Module) sexp.SExp {
	return lispOfTerms(module, "-", p.Args)
}

// RequiredRegisters implementation for Contextual interface.
func (p *Sub[T]) RequiredRegisters() *set.SortedSet[uint] {
	return requiredRegistersOfTerms(p.Args)
}

// RequiredCells implementation for Contextual interface
func (p *Sub[T]) RequiredCells(row int, tr trace.Module) *set.AnySortedSet[trace.CellRef] {
	return requiredCellsOfTerms(p.Args, row, tr)
}

// ShiftRange implementation for Term interface.
func (p *Sub[T]) ShiftRange() (int, int) {
	return shiftRangeOfTerms(p.Args)
}

// Simplify implementation for Term interface.
func (p *Sub[T]) Simplify(casts bool) T {
	panic("todo")
}

// ValueRange implementation for Term interface.
func (p *Sub[T]) ValueRange(module schema.Module) *util.Interval {
	var res util.Interval

	for i, arg := range p.Args {
		ith := arg.ValueRange(module)
		if i == 0 {
			res.Set(ith)
		} else {
			res.Sub(ith)
		}
	}
	//
	return &res
}
