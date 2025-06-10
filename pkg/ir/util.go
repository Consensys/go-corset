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
	"math"

	"github.com/consensys/gnark-crypto/ecc/bls12-377/fr"
	"github.com/consensys/go-corset/pkg/schema"
	"github.com/consensys/go-corset/pkg/trace"
	"github.com/consensys/go-corset/pkg/util/collection/set"
	"github.com/consensys/go-corset/pkg/util/source/sexp"
)

var (
	frZERO fr.Element = fr.NewElement(0)
	frONE  fr.Element = fr.NewElement(1)
)

// Check whether a given term corresponds with the constant zero.
func isZero[T Term[T]](term T) bool {
	var t Term[T] = term
	//
	if t, ok := t.(*Constant[T]); ok {
		return t.Value.IsZero()
	}
	//
	return false
}

// Check whether a given term corresponds with the constant one.
func isOne[T Term[T]](term T) bool {
	var t Term[T] = term
	//
	if t, ok := t.(*Constant[T]); ok {
		return t.Value.IsOne()
	}
	//
	return false
}

func lispOfTerms[T Term[T]](module schema.Module, op string, exprs []T) sexp.SExp {
	arr := make([]sexp.SExp, 1+len(exprs))
	arr[0] = sexp.NewSymbol(op)
	// Translate arguments
	for i, e := range exprs {
		arr[i+1] = e.Lisp(module)
	}
	// Done
	return sexp.NewList(arr)
}

func lispOfLogicalTerms[T LogicalTerm[T]](module schema.Module, op string, exprs []T) sexp.SExp {
	arr := make([]sexp.SExp, 1+len(exprs))
	arr[0] = sexp.NewSymbol(op)
	// Translate arguments
	for i, e := range exprs {
		arr[i+1] = e.Lisp(module)
	}
	// Done
	return sexp.NewList(arr)
}

func requiredRegistersOfTerms[T Contextual](args []T) *set.SortedSet[uint] {
	return set.UnionSortedSets(args, func(term T) *set.SortedSet[uint] {
		return term.RequiredRegisters()
	})
}

func requiredCellsOfTerms[T Contextual](args []T, row int, tr trace.Module) *set.AnySortedSet[trace.CellRef] {
	return set.UnionAnySortedSets(args, func(term T) *set.AnySortedSet[trace.CellRef] {
		return term.RequiredCells(row, tr)
	})
}

func shiftRangeOfTerms[T Term[T]](terms []T) (int, int) {
	minShift := math.MaxInt
	maxShift := math.MinInt
	//
	for _, term := range terms {
		tMin, tMax := term.ShiftRange()
		minShift = min(minShift, tMin)
		maxShift = max(maxShift, tMax)
	}
	//
	return minShift, maxShift
}

func applyShiftOfTerms[T Term[T]](terms []T, shift int) []T {
	nterms := make([]T, len(terms))
	//
	for i := range terms {
		nterms[i] = terms[i].ApplyShift(shift)
	}
	//
	return nterms
}

type binop func(fr.Element, fr.Element) fr.Element

// Simplify logical terms
func simplifyLogicalTerms[T LogicalTerm[T]](terms []T, casts bool) []T {
	var nterms = make([]T, len(terms))
	//
	for i, t := range terms {
		nterms[i] = t.Simplify(casts)
	}
	//
	return nterms
}

// General purpose constant propagation mechanism.  This reduces all terms to
// constants (where possible) and combines terms according to a given
// combinator.
func simplifyTerms[T Term[T]](terms []T, fn binop, acc fr.Element, casts bool) []T {
	// Count how many terms reduced to constants.
	var (
		count  = 0
		nterms = make([]T, len(terms))
		ith    Term[T]
	)
	// Propagate through all children
	for i, e := range terms {
		nterms[i] = e.Simplify(casts)
		ith = nterms[i]
		// Check for constant
		c, ok := ith.(*Constant[T])
		// Try to continue sum
		if ok {
			// Apply combinator
			acc = fn(acc, c.Value)
			// Increase count of constants
			count++
		}
	}
	// Merge all constants
	return mergeConstants(acc, nterms)
}

// Replace all constants within a given sequence of expressions with a single
// constant (whose value has been precomputed from those constants).  The new
// value replaces the first constant in the list.
func mergeConstants[T Term[T]](constant fr.Element, terms []T) []T {
	var (
		j     = 0
		first = true
	)
	//
	for i := range terms {
		var tmp Term[T] = terms[i]
		// Check for constant
		if _, ok := tmp.(*Constant[T]); ok && first {
			tmp = &Constant[T]{constant}
			terms[j] = tmp.(T)
			first = false
			j++
		} else if !ok {
			// Retain non-constant expression
			terms[j] = terms[i]
			j++
		}
	}
	// Return slice
	return terms[0:j]
}

func addBinOp(lhs fr.Element, rhs fr.Element) fr.Element {
	return *lhs.Add(&lhs, &rhs)
}

func mulBinOp(lhs fr.Element, rhs fr.Element) fr.Element {
	return *lhs.Mul(&lhs, &rhs)
}
