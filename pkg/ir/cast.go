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
	"math/big"

	"github.com/consensys/gnark-crypto/ecc/bls12-377/fr"
	"github.com/consensys/go-corset/pkg/schema"
	"github.com/consensys/go-corset/pkg/trace"
	"github.com/consensys/go-corset/pkg/util"
	"github.com/consensys/go-corset/pkg/util/collection/set"
	"github.com/consensys/go-corset/pkg/util/source/sexp"
)

// Cast attempts to narrow the width a given expression.
type Cast[T Term[T]] struct {
	Arg      T
	BitWidth uint
	Bound    fr.Element
}

// CastOf constructs a new expression which has been annotated by the user to be
// within a given range.
func CastOf[T Term[T]](arg T, bitwidth uint) T {
	bound := fr.NewElement(2)
	util.Pow(&bound, uint64(bitwidth))
	// Construct term
	var term Term[T] = &Cast[T]{Arg: arg, BitWidth: bitwidth, Bound: bound}
	// Done
	return term.(T)
}

// ApplyShift implementation for Term interface.
func (p *Cast[T]) ApplyShift(shift int) T {
	return CastOf(p.Arg.ApplyShift(shift), p.BitWidth)
}

// Bounds implementation for Boundable interface.
func (p *Cast[T]) Bounds() util.Bounds {
	return p.Arg.Bounds()
}

// EvalAt implementation for Evaluable interface.
func (p *Cast[T]) EvalAt(k int, tr trace.Module) (fr.Element, error) {
	// Check whether argument evaluates to zero or not.
	val, err := p.Arg.EvalAt(k, tr)
	// Dynamic cast check
	if err == nil && val.Cmp(&p.Bound) >= 0 {
		// Construct error
		err = fmt.Errorf("cast failure (value %s not a u%d)", val.String(), p.BitWidth)
	}
	// All good
	return val, err
}

// Lisp implementation for Lispifiable interface.
func (p *Cast[T]) Lisp(module schema.Module) sexp.SExp {
	arg := p.Arg.Lisp(module)
	name := sexp.NewSymbol(fmt.Sprintf(":u%d", p.BitWidth))

	return sexp.NewList([]sexp.SExp{name, arg})
}

// RequiredRegisters implementation for Contextual interface.
func (p *Cast[T]) RequiredRegisters() *set.SortedSet[uint] {
	return p.Arg.RequiredRegisters()
}

// RequiredCells implementation for Contextual interface
func (p *Cast[T]) RequiredCells(row int, mid trace.ModuleId) *set.AnySortedSet[trace.CellRef] {
	return p.Arg.RequiredCells(row, mid)
}

// Range returns the range of values which this cast represents.
func (p *Cast[T]) Range() *util.Interval {
	var bound = big.NewInt(2)
	// Determine bound for static type check
	bound.Exp(bound, big.NewInt(int64(p.BitWidth)), nil)
	// Subtract 1 because interval is inclusive.
	bound.Sub(bound, &biONE)
	// Determine casted interval
	return util.NewInterval(&biZERO, bound)
}

// ShiftRange implementation for Term interface.
func (p *Cast[T]) ShiftRange() (int, int) {
	return p.Arg.ShiftRange()
}

// Simplify implementation for Term interface.
func (p *Cast[T]) Simplify(casts bool) T {
	var bound fr.Element = fr.NewElement(2)
	// Determine bound for static type check
	util.Pow(&bound, uint64(p.BitWidth))
	// Propagate constants in the argument

	var (
		arg  T       = p.Arg.Simplify(casts)
		targ Term[T] = arg
	)
	//
	if c, ok := targ.(*Constant[T]); ok && c.Value.Cmp(&bound) < 0 {
		// Done
		return arg
	} else if ok {
		// Type failure
		panic(fmt.Sprintf("type cast failure (have %s with expected bitwidth %d)", c.Value.String(), p.BitWidth))
	} else if casts {
		targ = CastOf(arg, p.BitWidth)
		arg = targ.(T)
	}
	// elide cast
	return arg
}

// Substitute implementation for Substitutable interface.
func (p *Cast[T]) Substitute(mapping map[string]fr.Element) {
	p.Arg.Substitute(mapping)
}

// ValueRange implementation for Term interface.
func (p *Cast[T]) ValueRange(module schema.Module) *util.Interval {
	cast := p.Range()
	// Compute actual interval
	res := p.Arg.ValueRange(module)
	// Check whether is within (or not)
	if res.Within(cast) {
		return res
	}
	//
	return cast
}
