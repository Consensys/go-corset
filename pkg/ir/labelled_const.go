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
	"math/big"

	"github.com/consensys/gnark-crypto/ecc/bls12-377/fr"
	"github.com/consensys/go-corset/pkg/schema"
	"github.com/consensys/go-corset/pkg/trace"
	"github.com/consensys/go-corset/pkg/util"
	"github.com/consensys/go-corset/pkg/util/collection/set"
	util_math "github.com/consensys/go-corset/pkg/util/math"
	"github.com/consensys/go-corset/pkg/util/source/sexp"
)

// LabelledConst represents a constant value which is labelled with a given
// name.  The purpose of this is to allow labelled constants to be substituted
// for different values when desired.
type LabelledConst[T Term[T]] struct {
	Label string
	Value fr.Element
}

// LabelledConstant construct an expression representing a constant with a given
// label.
func LabelledConstant[T Term[T]](label string, value fr.Element) T {
	var term Term[T] = &LabelledConst[T]{Label: label, Value: value}
	return term.(T)
}

// ApplyShift implementation for Term interface.
func (p *LabelledConst[T]) ApplyShift(int) T {
	var term Term[T] = p
	return term.(T)
}

// Bounds implementation for Boundable interface.
func (p *LabelledConst[T]) Bounds() util.Bounds {
	return util.EMPTY_BOUND
}

// EvalAt implementation for Evaluable interface.
func (p *LabelledConst[T]) EvalAt(k int, _ trace.Module, _ schema.Module) (fr.Element, error) {
	return p.Value, nil
}

// IsDefined implementation for Evaluable interface.
func (p *LabelledConst[T]) IsDefined() bool {
	// NOTE: this is technically safe given the limited way that IsDefined is
	// used for lookup selectors.
	return true
}

// Lisp implementation for Lispifiable interface.
func (p *LabelledConst[T]) Lisp(_ schema.RegisterMap) sexp.SExp {
	return sexp.NewSymbol(p.Value.String())
}

// RequiredRegisters implementation for Contextual interface.
func (p *LabelledConst[T]) RequiredRegisters() *set.SortedSet[uint] {
	return set.NewSortedSet[uint]()
}

// RequiredCells implementation for Contextual interface
func (p *LabelledConst[T]) RequiredCells(row int, mid trace.ModuleId) *set.AnySortedSet[trace.CellRef] {
	return set.NewAnySortedSet[trace.CellRef]()
}

// ShiftRange implementation for Term interface.
func (p *LabelledConst[T]) ShiftRange() (int, int) {
	return math.MaxInt, math.MinInt
}

// Simplify implementation for Term interface.
func (p *LabelledConst[T]) Simplify(casts bool) T {
	var tmp Term[T] = p
	return tmp.(T)
}

// Substitute implementation for Substitutable interface.
func (p *LabelledConst[T]) Substitute(mapping map[string]fr.Element) {
	// Attempt to apply substitution
	if nval, ok := mapping[p.Label]; ok {
		p.Value = nval
	}
}

// ValueRange implementation for Term interface.
func (p *LabelledConst[T]) ValueRange(_ schema.RegisterMap) util_math.Interval {
	var c big.Int
	// Extract big integer from field element
	p.Value.BigInt(&c)
	// Return as interval
	return util_math.NewInterval(c, c)
}
