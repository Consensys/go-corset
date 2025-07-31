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
	"github.com/consensys/go-corset/pkg/util/field"
	"github.com/consensys/go-corset/pkg/util/math"
	"github.com/consensys/go-corset/pkg/util/source/sexp"
)

// VectorAccess represents the bitwise concatenation of one or more registers.
// Registers are organised in little endian form.  That is, the least
// significant register comes first (i.e. has index 0 in the array).
type VectorAccess[T Term[T]] struct{ Vars []*RegisterAccess[T] }

// NewVectorAccess constructs a new vector access for a given set of registers.
func NewVectorAccess[T Term[T]](vars []*RegisterAccess[T]) T {
	var term Term[T] = &VectorAccess[T]{vars}
	//
	return term.(T)
}

// ApplyShift implementation for Term interface.
func (p *VectorAccess[T]) ApplyShift(shift int) T {
	nterms := make([]*RegisterAccess[T], len(p.Vars))
	//
	for i := range p.Vars {
		var ith Term[T] = p.Vars[i].ApplyShift(shift)
		nterms[i] = ith.(*RegisterAccess[T])
	}
	//
	var term Term[T] = &VectorAccess[T]{nterms}
	//
	return term.(T)
}

// Bounds implementation for Boundable interface.
func (p *VectorAccess[T]) Bounds() util.Bounds { return util.BoundsForArray(p.Vars) }

// EvalAt implementation for Evaluable interface.
func (p *VectorAccess[T]) EvalAt(k int, tr trace.Module, sc schema.Module) (fr.Element, error) {
	var shift = sc.Register(p.Vars[0].Register).Width
	// Evaluate first argument
	val, err := p.Vars[0].EvalAt(k, tr, sc)
	// Continue evaluating the rest
	for i := 1; err == nil && i < len(p.Vars); i++ {
		var (
			ith       fr.Element
			ith_width = sc.Register(p.Vars[i].Register).Width
		)
		// Evaluate ith argument
		ith, err = p.Vars[i].EvalAt(k, tr, sc)
		//
		val.Add(&val, shiftValue(ith, shift))
		//
		shift = shift + ith_width
	}
	// Done
	return val, err
}

// IsDefined implementation for Evaluable interface.
func (p *VectorAccess[T]) IsDefined() bool {
	// NOTE: this is technically safe given the limited way that IsDefined is
	// used for lookup selectors.
	return true
}

// Lisp implementation for Lispifiable interface.
func (p *VectorAccess[T]) Lisp(global bool, mapping schema.RegisterMap) sexp.SExp {
	return lispOfTerms(global, mapping, "::", p.Vars)
}

// RequiredRegisters implementation for Contextual interface.
func (p *VectorAccess[T]) RequiredRegisters() *set.SortedSet[uint] {
	return requiredRegistersOfTerms(p.Vars)
}

// RequiredCells implementation for Contextual interface
func (p *VectorAccess[T]) RequiredCells(row int, mid trace.ModuleId) *set.AnySortedSet[trace.CellRef] {
	return requiredCellsOfTerms(p.Vars, row, mid)
}

// ShiftRange implementation for Term interface.
func (p *VectorAccess[T]) ShiftRange() (int, int) {
	return shiftRangeOfTerms(p.Vars...)
}

// Simplify implementation for Term interface.
func (p *VectorAccess[T]) Simplify(casts bool) T {
	var term Term[T] = p
	return term.(T)
}

// Substitute implementation for Substitutable interface.
func (p *VectorAccess[T]) Substitute(mapping map[string]fr.Element) {
	substituteTerms(mapping, p.Vars...)
}

// ValueRange implementation for Term interface.
func (p *VectorAccess[T]) ValueRange(mapping schema.RegisterMap) math.Interval {
	var width = uint(0)
	// Determine total bitwidth of the vector
	for _, arg := range p.Vars {
		ith_width := mapping.Register(arg.Register).Width
		width += ith_width
	}
	//
	return valueRangeOfBits(width)
}

func shiftValue(val fr.Element, width uint) *fr.Element {
	var coeff = fr.NewElement(2)
	// Determine 2^width
	field.Pow(&coeff, uint64(width))
	// Determine val & 2^width
	return val.Mul(&val, &coeff)
}
