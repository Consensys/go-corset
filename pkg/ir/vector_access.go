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
	"github.com/consensys/go-corset/pkg/util/field"
	"github.com/consensys/go-corset/pkg/util/math"
	"github.com/consensys/go-corset/pkg/util/source/sexp"
)

// VectorAccess represents the bitwise concatenation of one or more registers.
// Registers are organised in little endian form.  That is, the least
// significant register comes first (i.e. has index 0 in the array).
type VectorAccess[F field.Element[F], T Term[F, T]] struct{ Vars []*RegisterAccess[F, T] }

// NewVectorAccess constructs a new vector access for a given set of registers.
func NewVectorAccess[F field.Element[F], T Term[F, T]](vars []*RegisterAccess[F, T]) T {
	var term Term[F, T] = &VectorAccess[F, T]{vars}
	//
	return term.(T)
}

// ApplyShift implementation for Term interface.
func (p *VectorAccess[F, T]) ApplyShift(shift int) T {
	nterms := make([]*RegisterAccess[F, T], len(p.Vars))
	//
	for i := range p.Vars {
		var ith Term[F, T] = p.Vars[i].ApplyShift(shift)
		nterms[i] = ith.(*RegisterAccess[F, T])
	}
	//
	var term Term[F, T] = &VectorAccess[F, T]{nterms}
	//
	return term.(T)
}

// Bounds implementation for Boundable interface.
func (p *VectorAccess[F, T]) Bounds() util.Bounds { return util.BoundsForArray(p.Vars) }

// EvalAt implementation for Evaluable interface.
func (p *VectorAccess[F, T]) EvalAt(k int, tr trace.Module[F], sc schema.Module) (F, error) {
	var shift = sc.Register(p.Vars[0].Register).Width
	// Evaluate first argument
	val, err := p.Vars[0].EvalAt(k, tr, sc)
	// Continue evaluating the rest
	for i := 1; err == nil && i < len(p.Vars); i++ {
		var (
			ith       F
			ith_width = sc.Register(p.Vars[i].Register).Width
		)
		// Evaluate ith argument
		ith, err = p.Vars[i].EvalAt(k, tr, sc)
		//
		val = val.Add(shiftValue(ith, shift))
		//
		shift = shift + ith_width
	}
	// Done
	return val, err
}

// IsDefined implementation for Evaluable interface.
func (p *VectorAccess[F, T]) IsDefined() bool {
	// NOTE: this is technically safe given the limited way that IsDefined is
	// used for lookup selectors.
	return true
}

// Lisp implementation for Lispifiable interface.
func (p *VectorAccess[F, T]) Lisp(global bool, mapping schema.RegisterMap) sexp.SExp {
	return lispOfTerms(global, mapping, "::", p.Vars)
}

// RequiredRegisters implementation for Contextual interface.
func (p *VectorAccess[F, T]) RequiredRegisters() *set.SortedSet[uint] {
	return requiredRegistersOfTerms(p.Vars)
}

// RequiredCells implementation for Contextual interface
func (p *VectorAccess[F, T]) RequiredCells(row int, mid trace.ModuleId) *set.AnySortedSet[trace.CellRef] {
	return requiredCellsOfTerms(p.Vars, row, mid)
}

// ShiftRange implementation for Term interface.
func (p *VectorAccess[F, T]) ShiftRange() (int, int) {
	return shiftRangeOfTerms(p.Vars...)
}

// Simplify implementation for Term interface.
func (p *VectorAccess[F, T]) Simplify(casts bool) T {
	var term Term[F, T] = p
	return term.(T)
}

// Substitute implementation for Substitutable interface.
func (p *VectorAccess[F, T]) Substitute(mapping map[string]F) {
	substituteTerms(mapping, p.Vars...)
}

// ValueRange implementation for Term interface.
func (p *VectorAccess[F, T]) ValueRange(mapping schema.RegisterMap) math.Interval {
	var width = uint(0)
	// Determine total bitwidth of the vector
	for _, arg := range p.Vars {
		ith_width := mapping.Register(arg.Register).Width
		width += ith_width
	}
	//
	return valueRangeOfBits(width)
}

func shiftValue[F field.Element[F]](val F, width uint) F {
	// Determine 2^width
	coeff := field.TwoPowN[F](width)
	// Determine val * 2^width
	return val.Mul(coeff)
}
