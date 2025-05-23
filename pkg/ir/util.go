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

var frZERO fr.Element = fr.NewElement(0)

func applyShiftTerms[T Term[T]](terms []T, shift int) []T {
	nterms := make([]T, len(terms))
	//
	for i := range terms {
		nterms[i] = terms[i].ApplyShift(shift)
	}
	//
	return nterms
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
