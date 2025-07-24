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
	"github.com/consensys/go-corset/pkg/util/collection/array"
	"github.com/consensys/go-corset/pkg/util/collection/set"
	"github.com/consensys/go-corset/pkg/util/math"
	"github.com/consensys/go-corset/pkg/util/source/sexp"
)

// Mul represents the product over zero or more expressions.
type Mul[T Term[T]] struct{ Args []T }

// Product returns the product of zero or more multiplications.
func Product[T Term[T]](terms ...T) T {
	// flatten any nested products
	terms = array.Flatten(terms, flatternMul)
	// Remove all multiplications by one
	terms = array.RemoveMatching(terms, isOne)
	// Check for zero
	if array.ContainsMatching(terms, isZero) {
		return Const64[T](0)
	}
	// Final optimisation
	switch len(terms) {
	case 0:
		return Const64[T](1)
	case 1:
		return terms[0]
	default:
		var term Term[T] = &Mul[T]{terms}
		//
		return term.(T)
	}
}

// Air indicates this term can be used at the AIR level.
func (p *Mul[T]) Air() {}

// ApplyShift implementation for Term interface.
func (p *Mul[T]) ApplyShift(shift int) T {
	var term Term[T] = &Mul[T]{applyShiftOfTerms(p.Args, shift)}
	return term.(T)
}

// Bounds implementation for Boundable interface.
func (p *Mul[T]) Bounds() util.Bounds { return util.BoundsForArray(p.Args) }

// EvalAt implementation for Evaluable interface.
func (p *Mul[T]) EvalAt(k int, tr trace.Module, sc schema.Module) (fr.Element, error) {
	// Evaluate first argument
	val, err := p.Args[0].EvalAt(k, tr, sc)
	// Continue evaluating the rest
	for i := 1; err == nil && i < len(p.Args); i++ {
		var ith fr.Element
		// Can short-circuit evaluation?
		if val.IsZero() {
			return val, nil
		}
		// No
		ith, err = p.Args[i].EvalAt(k, tr, sc)
		val.Mul(&val, &ith)
	}
	// Done
	return val, err
}

// IsDefined implementation for Evaluable interface.
func (p *Mul[T]) IsDefined() bool {
	// NOTE: this is technically safe given the limited way that IsDefined is
	// used for lookup selectors.
	return true
}

// Lisp implementation for Lispifiable interface.
func (p *Mul[T]) Lisp(mapping schema.RegisterMap) sexp.SExp {
	return lispOfTerms(mapping, "*", p.Args)
}

// RequiredRegisters implementation for Contextual interface.
func (p *Mul[T]) RequiredRegisters() *set.SortedSet[uint] {
	return requiredRegistersOfTerms(p.Args)
}

// RequiredCells implementation for Contextual interface
func (p *Mul[T]) RequiredCells(row int, mid trace.ModuleId) *set.AnySortedSet[trace.CellRef] {
	return requiredCellsOfTerms(p.Args, row, mid)
}

// ShiftRange implementation for Term interface.
func (p *Mul[T]) ShiftRange() (int, int) {
	return shiftRangeOfTerms(p.Args...)
}

// Substitute implementation for Substitutable interface.
func (p *Mul[T]) Substitute(mapping map[string]fr.Element) {
	substituteTerms(mapping, p.Args...)
}

// Simplify implementation for Term interface.
func (p *Mul[T]) Simplify(casts bool) T {
	var (
		terms = simplifyTerms(p.Args, mulBinOp, frONE, casts)
		targ  Term[T]
	)
	// Flatten any nested products
	terms = array.Flatten(terms, flatternMul)
	// Check for zero
	if array.ContainsMatching(terms, isZero) {
		// Yes, is zero
		targ = &Constant[T]{fr.NewElement(0)}
	} else {
		// Remove any ones
		terms = array.RemoveMatching(terms, isOne)
		// Check whats left
		switch len(terms) {
		case 0:
			targ = &Constant[T]{frONE}
		case 1:
			return terms[0]
		default:
			// Done
			targ = &Mul[T]{terms}
		}
	}
	//
	return targ.(T)
}

// ValueRange implementation for Term interface.
func (p *Mul[T]) ValueRange(mapping schema.RegisterMap) math.Interval {
	var res math.Interval

	for i, arg := range p.Args {
		ith := arg.ValueRange(mapping)
		if i == 0 {
			res.Set(ith)
		} else {
			res.Mul(ith)
		}
	}
	//
	return res
}

func flatternMul[T Term[T]](term T) []T {
	var e Term[T] = term
	if t, ok := e.(*Mul[T]); ok {
		return t.Args
	}
	//
	return nil
}
