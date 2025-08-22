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
	"github.com/consensys/go-corset/pkg/util/collection/array"
	"github.com/consensys/go-corset/pkg/util/collection/set"
	"github.com/consensys/go-corset/pkg/util/field"
	"github.com/consensys/go-corset/pkg/util/math"
	"github.com/consensys/go-corset/pkg/util/source/sexp"
)

// Sub represents the subtraction over zero or more expressions.
type Sub[F field.Element[F], T Term[F, T]] struct{ Args []T }

// Subtract returns the subtraction of the subsequent expressions from the
// first.
func Subtract[F field.Element[F], T Term[F, T]](terms ...T) T {
	// Remove any zeros
	terms = array.RemoveMatchingIndexed(terms, isZeroExcept)
	// Final simplifications
	switch len(terms) {
	case 0:
		return Const64[F, T](0)
	case 1:
		return terms[0]
	default:
		var term Term[F, T] = &Sub[F, T]{terms}
		//
		return term.(T)
	}
}

// Air indicates this term can be used at the AIR level.
func (p *Sub[F, T]) Air() {}

// ApplyShift implementation for Term interface.
func (p *Sub[F, T]) ApplyShift(shift int) T {
	var term Term[F, T] = &Sub[F, T]{applyShiftOfTerms(p.Args, shift)}
	return term.(T)
}

// Bounds implementation for Boundable interface.
func (p *Sub[F, T]) Bounds() util.Bounds { return util.BoundsForArray(p.Args) }

// EvalAt implementation for Evaluable interface.
func (p *Sub[F, T]) EvalAt(k int, tr trace.Module[F], sc schema.Module[F]) (F, error) {
	// Evaluate first argument
	val, err := p.Args[0].EvalAt(k, tr, sc)
	// Continue evaluating the rest
	for i := 1; err == nil && i < len(p.Args); i++ {
		var ith F
		// Evaluate ith argument
		ith, err = p.Args[i].EvalAt(k, tr, sc)
		val = val.Sub(ith)
	}
	// Done
	return val, err
}

// IsDefined implementation for Evaluable interface.
func (p *Sub[F, T]) IsDefined() bool {
	// NOTE: this is technically safe given the limited way that IsDefined is
	// used for lookup selectors.
	return true
}

// Lisp implementation for Lispifiable interface.
func (p *Sub[F, T]) Lisp(global bool, mapping schema.RegisterMap) sexp.SExp {
	return lispOfTerms(global, mapping, "-", p.Args)
}

// RequiredRegisters implementation for Contextual interface.
func (p *Sub[F, T]) RequiredRegisters() *set.SortedSet[uint] {
	return requiredRegistersOfTerms(p.Args)
}

// RequiredCells implementation for Contextual interface
func (p *Sub[F, T]) RequiredCells(row int, mid trace.ModuleId) *set.AnySortedSet[trace.CellRef] {
	return requiredCellsOfTerms(p.Args, row, mid)
}

// ShiftRange implementation for Term interface.
func (p *Sub[F, T]) ShiftRange() (int, int) {
	return shiftRangeOfTerms(p.Args...)
}

// Substitute implementation for Substitutable interface.
func (p *Sub[F, T]) Substitute(mapping map[string]F) {
	substituteTerms(mapping, p.Args...)
}

// ValueRange implementation for Term interface.
func (p *Sub[F, T]) ValueRange(mapping schema.RegisterMap) math.Interval {
	var res math.Interval

	for i, arg := range p.Args {
		ith := arg.ValueRange(mapping)
		if i == 0 {
			res.Set(ith)
		} else {
			res.Sub(ith)
		}
	}
	//
	return res
}

// Simplify implementation for Term interface.
func (p *Sub[F, T]) Simplify(casts bool) T {
	var (
		targ  Term[F, T]
		lhs   T          = p.Args[0].Simplify(casts)
		lhs_t Term[F, T] = lhs
		// Subtraction is harder to optimise for.  What we do is view "a - b - c" as
		// "a - (b+c)", and optimise the right-hand side as though it were addition.
		rhs   T          = simplifySum(p.Args[1:], casts)
		rhs_t Term[F, T] = rhs
	)
	// Check what's left
	lc, l_const := lhs_t.(*Constant[F, T])
	rc, r_const := rhs_t.(*Constant[F, T])
	ra, r_add := rhs_t.(*Add[F, T])
	r_zero := isZero(rhs)
	//
	switch {
	case r_zero:
		// Right-hand side zero, nothing to subtract.
		return lhs
	case l_const && r_const:
		// Both sides constant, result is constant.
		c := lc.Value.Sub(rc.Value)
		//
		targ = &Constant[F, T]{c}
	case l_const && r_add:
		nterms := array.Prepend(lhs, ra.Args)
		// if rhs has constant, subtract it.
		if rc, ok := findConstant(ra.Args); ok {
			c := lc.Value.Sub(rc)
			nterms = mergeConstants(c, nterms)
		}
		//
		targ = &Sub[F, T]{nterms}
	case r_add:
		// Default case, recombine.
		targ = &Sub[F, T]{array.Prepend(lhs, ra.Args)}
	default:
		targ = &Sub[F, T]{[]T{lhs, rhs}}
	}
	//
	return targ.(T)
}

func findConstant[F field.Element[F], T Term[F, T]](terms []T) (F, bool) {
	for _, t := range terms {
		var ith Term[F, T] = t
		if c, ok := ith.(*Constant[F, T]); ok {
			return c.Value, true
		}
	}
	//
	return field.Zero[F](), false
}

func isZeroExcept[F field.Element[F], T Term[F, T]](i int, term T) bool {
	return i != 0 && isZero(term)
}
