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
	"fmt"
	"math"
	"math/big"

	"github.com/consensys/gnark-crypto/ecc/bls12-377/fr"
	"github.com/consensys/go-corset/pkg/schema"
	"github.com/consensys/go-corset/pkg/schema/agnostic"
	"github.com/consensys/go-corset/pkg/trace"
	"github.com/consensys/go-corset/pkg/util"
	"github.com/consensys/go-corset/pkg/util/collection/set"
	"github.com/consensys/go-corset/pkg/util/field"
	util_math "github.com/consensys/go-corset/pkg/util/math"
	"github.com/consensys/go-corset/pkg/util/source/sexp"
)

// Constant represents a constant value within an expression.
type Constant[F field.Element[F], T Term[T]] struct{ Value F }

// Const construct an AIR expression representing a given constant.
func Const[F field.Element[F], T Term[T]](val F) T {
	var term Term[T] = &Constant[F, T]{Value: val}
	return term.(T)
}

// Const64 construct an AIR expression representing a given constant from a
// uint64.
func Const64[F field.Element[F], T Term[T]](val uint64) T {
	var (
		element         = fr.NewElement(val)
		term    Term[T] = &Constant[F, T]{Value: element}
	)
	//
	return term.(T)
}

// IsConstant checks whether an artibrary term corresponds to a constant or not.
func IsConstant[F field.Element[F], T Term[T]](term T) *fr.Element {
	var tmp Term[T] = term
	//
	if c, ok := tmp.(*Constant[F, T]); ok {
		return &c.Value
	}
	//
	return nil
}

// Air indicates this term can be used at the AIR level.
func (p *Constant[F, T]) Air() {}

// ApplyShift implementation for Term interface.
func (p *Constant[F, T]) ApplyShift(int) T {
	var term Term[T] = p
	return term.(T)
}

// Bounds implementation for Boundable interface.
func (p *Constant[F, T]) Bounds() util.Bounds {
	return util.EMPTY_BOUND
}

// EvalAt implementation for Evaluable interface.
func (p *Constant[F, T]) EvalAt(k int, _ trace.Module[F], _ schema.Module) (fr.Element, error) {
	return p.Value, nil
}

// IsDefined implementation for Evaluable interface.
func (p *Constant[F, T]) IsDefined() bool {
	return true
}

// Lisp implementation for Lispifiable interface.
func (p *Constant[F, T]) Lisp(global bool, mapping schema.RegisterMap) sexp.SExp {
	var val big.Int
	//
	p.Value.BigInt(&val)
	// Check if power of 2
	if n, ok := agnostic.IsPowerOf2(val); ok && n > 8 {
		// Not power of 2
		return sexp.NewSymbol(fmt.Sprintf("2^%d", n))
	}
	//
	return sexp.NewSymbol(p.Value.String())
}

// RequiredRegisters implementation for Contextual interface.
func (p *Constant[F, T]) RequiredRegisters() *set.SortedSet[uint] {
	return set.NewSortedSet[uint]()
}

// RequiredCells implementation for Contextual interface
func (p *Constant[F, T]) RequiredCells(row int, mid trace.ModuleId) *set.AnySortedSet[trace.CellRef] {
	return set.NewAnySortedSet[trace.CellRef]()
}

// ShiftRange implementation for Term interface.
func (p *Constant[F, T]) ShiftRange() (int, int) {
	return math.MaxInt, math.MinInt
}

// Substitute implementation for Substitutable interface.
func (p *Constant[F, T]) Substitute(mapping map[string]fr.Element) {

}

// Simplify implementation for Term interface.
func (p *Constant[F, T]) Simplify(casts bool) T {
	var tmp Term[T] = p
	return tmp.(T)
}

// ValueRange implementation for Term interface.
func (p *Constant[F, T]) ValueRange(_ schema.RegisterMap) util_math.Interval {
	var c big.Int
	// Extract big integer from field element
	c.SetBytes(p.Value.Bytes())
	// Return as interval
	return util_math.NewInterval(c, c)
}
