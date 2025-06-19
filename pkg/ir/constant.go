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

// Constant represents a constant value within an expression.
type Constant[T Term[T]] struct{ Value fr.Element }

// Const construct an AIR expression representing a given constant.
func Const[T Term[T]](val fr.Element) T {
	var term Term[T] = &Constant[T]{Value: val}
	return term.(T)
}

// Const64 construct an AIR expression representing a given constant from a
// uint64.
func Const64[T Term[T]](val uint64) T {
	var (
		element         = fr.NewElement(val)
		term    Term[T] = &Constant[T]{Value: element}
	)
	//
	return term.(T)
}

// IsConstant checks whether an artibrary term corresponds to a constant or not.
func IsConstant[T Term[T]](term T) *fr.Element {
	var tmp Term[T] = term
	//
	if c, ok := tmp.(*Constant[T]); ok {
		return &c.Value
	}
	//
	return nil
}

// Air indicates this term can be used at the AIR level.
func (p *Constant[T]) Air() {}

// ApplyShift implementation for Term interface.
func (p *Constant[T]) ApplyShift(int) T {
	var term Term[T] = p
	return term.(T)
}

// Bounds implementation for Boundable interface.
func (p *Constant[T]) Bounds() util.Bounds {
	return util.EMPTY_BOUND
}

// EvalAt implementation for Evaluable interface.
func (p *Constant[T]) EvalAt(k int, module trace.Module) (fr.Element, error) {
	return p.Value, nil
}

// Lisp implementation for Lispifiable interface.
func (p *Constant[T]) Lisp(module schema.Module) sexp.SExp {
	return sexp.NewSymbol(p.Value.String())
}

// RequiredRegisters implementation for Contextual interface.
func (p *Constant[T]) RequiredRegisters() *set.SortedSet[uint] {
	return set.NewSortedSet[uint]()
}

// RequiredCells implementation for Contextual interface
func (p *Constant[T]) RequiredCells(row int, mid trace.ModuleId) *set.AnySortedSet[trace.CellRef] {
	return set.NewAnySortedSet[trace.CellRef]()
}

// ShiftRange implementation for Term interface.
func (p *Constant[T]) ShiftRange() (int, int) {
	return math.MaxInt, math.MinInt
}

// Substitute implementation for Substitutable interface.
func (p *Constant[T]) Substitute(mapping map[string]fr.Element) {

}

// Simplify implementation for Term interface.
func (p *Constant[T]) Simplify(casts bool) T {
	var tmp Term[T] = p
	return tmp.(T)
}

// ValueRange implementation for Term interface.
func (p *Constant[T]) ValueRange(module schema.Module) *util_math.Interval {
	var c big.Int
	// Extract big integer from field element
	p.Value.BigInt(&c)
	// Return as interval
	return util_math.NewInterval(&c, &c)
}
