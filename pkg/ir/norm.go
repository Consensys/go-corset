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
	"math/big"

	"github.com/consensys/gnark-crypto/ecc/bls12-377/fr"
	"github.com/consensys/go-corset/pkg/schema"
	"github.com/consensys/go-corset/pkg/trace"
	"github.com/consensys/go-corset/pkg/util"
	"github.com/consensys/go-corset/pkg/util/collection/set"
	"github.com/consensys/go-corset/pkg/util/math"
	"github.com/consensys/go-corset/pkg/util/source/sexp"
)

// Norm reduces the value of an expression to either zero (if it was zero)
// or one (otherwise).
type Norm[T Term[T]] struct{ Arg T }

// Normalise normalises the result of evaluating a given expression to be
// either 0 (if its value was 0) or 1 (otherwise).
func Normalise[T Term[T]](arg T) T {
	var term Term[T] = &Norm[T]{arg}
	return term.(T)
}

// ApplyShift implementation for Term interface.
func (p *Norm[T]) ApplyShift(shift int) T {
	return Normalise(p.Arg.ApplyShift(shift))
}

// Bounds implementation for Boundable interface.
func (p *Norm[T]) Bounds() util.Bounds {
	return p.Arg.Bounds()
}

// EvalAt implementation for Evaluable interface.
func (p *Norm[T]) EvalAt(k int, tr trace.Module, sc schema.Module) (fr.Element, error) {
	// Check whether argument evaluates to zero or not.
	val, err := p.Arg.EvalAt(k, tr, sc)
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

// RequiredRegisters implementation for Contextual interface.
func (p *Norm[T]) RequiredRegisters() *set.SortedSet[uint] {
	return p.Arg.RequiredRegisters()
}

// RequiredCells implementation for Contextual interface
func (p *Norm[T]) RequiredCells(row int, mid trace.ModuleId) *set.AnySortedSet[trace.CellRef] {
	return p.Arg.RequiredCells(row, mid)
}

// ShiftRange implementation for Term interface.
func (p *Norm[T]) ShiftRange() (int, int) {
	return p.Arg.ShiftRange()
}

// Simplify implementation for Term interface.
func (p *Norm[T]) Simplify(casts bool) T {
	var (
		arg  T       = p.Arg.Simplify(casts)
		targ Term[T] = arg
	)
	//
	if c, ok := targ.(*Constant[T]); ok {
		var val fr.Element
		// Clone value
		val.Set(&c.Value)
		// Normalise (in place)
		if !val.IsZero() {
			val.SetOne()
		}
		// Done
		targ = &Constant[T]{val}
	} else {
		targ = &Norm[T]{arg}
	}
	//
	return targ.(T)
}

// Substitute implementation for Substitutable interface.
func (p *Norm[T]) Substitute(mapping map[string]fr.Element) {
	p.Arg.Substitute(mapping)
}

// ValueRange implementation for Term interface.
func (p *Norm[T]) ValueRange(module schema.Module) *math.Interval {
	return math.NewInterval(big.NewInt(0), big.NewInt(1))
}
