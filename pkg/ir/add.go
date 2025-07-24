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

// Add represents the addition of zero or more expressions.
type Add[T Term[T]] struct{ Args []T }

// Sum zero or more expressions together.
func Sum[T Term[T]](terms ...T) T {
	// Flatten any nested sums
	terms = array.Flatten(terms, flatternAdd)
	// Remove any zeros
	terms = array.RemoveMatching(terms, isZero)
	// Final simplifications
	switch len(terms) {
	case 0:
		return Const64[T](0)
	case 1:
		return terms[0]
	default:
		var term Term[T] = &Add[T]{terms}
		//
		return term.(T)
	}
}

// Air indicates this term can be used at the AIR level.
func (p *Add[T]) Air() {}

// ApplyShift implementation for Term interface.
func (p *Add[T]) ApplyShift(shift int) T {
	var term Term[T] = &Add[T]{applyShiftOfTerms(p.Args, shift)}
	return term.(T)
}

// Bounds implementation for Boundable interface.
func (p *Add[T]) Bounds() util.Bounds { return util.BoundsForArray(p.Args) }

// EvalAt implementation for Evaluable interface.
func (p *Add[T]) EvalAt(k int, tr trace.Module, sc schema.Module) (fr.Element, error) {
	// Evaluate first argument
	val, err := p.Args[0].EvalAt(k, tr, sc)
	// Continue evaluating the rest
	for i := 1; err == nil && i < len(p.Args); i++ {
		var ith fr.Element
		// Evaluate ith argument
		ith, err = p.Args[i].EvalAt(k, tr, sc)
		val.Add(&val, &ith)
	}
	// Done
	return val, err
}

// IsDefined implementation for Evaluable interface.
func (p *Add[T]) IsDefined() bool {
	// NOTE: this is technically safe given the limited way that IsDefined is
	// used for lookup selectors.
	return true
}

// Lisp implementation for Lispifiable interface.
func (p *Add[T]) Lisp(mapping schema.RegisterMap) sexp.SExp {
	return lispOfTerms(mapping, "+", p.Args)
}

// RequiredRegisters implementation for Contextual interface.
func (p *Add[T]) RequiredRegisters() *set.SortedSet[uint] {
	return requiredRegistersOfTerms(p.Args)
}

// RequiredCells implementation for Contextual interface
func (p *Add[T]) RequiredCells(row int, mid trace.ModuleId) *set.AnySortedSet[trace.CellRef] {
	return requiredCellsOfTerms(p.Args, row, mid)
}

// ShiftRange implementation for Term interface.
func (p *Add[T]) ShiftRange() (int, int) {
	return shiftRangeOfTerms(p.Args...)
}

// Simplify implementation for Term interface.
func (p *Add[T]) Simplify(casts bool) T {
	return simplifySum(p.Args, casts)
}

// Substitute implementation for Substitutable interface.
func (p *Add[T]) Substitute(mapping map[string]fr.Element) {
	substituteTerms(mapping, p.Args...)
}

// ValueRange implementation for Term interface.
func (p *Add[T]) ValueRange(mapping schema.RegisterMap) math.Interval {
	var res math.Interval

	for i, arg := range p.Args {
		ith := arg.ValueRange(mapping)
		if i == 0 {
			res.Set(ith)
		} else {
			res.Add(ith)
		}
	}
	//
	return res
}

func simplifySum[T Term[T]](args []T, casts bool) T {
	var (
		terms = simplifyTerms(args, addBinOp, frZERO, casts)
		tmp   Term[T]
	)
	// Flatten any nested sums
	terms = array.Flatten(terms, flatternAdd)
	// Remove any zeros
	terms = array.RemoveMatching(terms, isZero)
	// Check anything left
	switch len(terms) {
	case 0:
		tmp = &Constant[T]{frZERO}
	case 1:
		return terms[0]
	default:
		tmp = &Add[T]{terms}
	}
	// Done
	return tmp.(T)
}

func flatternAdd[T Term[T]](term T) []T {
	var e Term[T] = term
	if t, ok := e.(*Add[T]); ok {
		return t.Args
	}
	//
	return nil
}
