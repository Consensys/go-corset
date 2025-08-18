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

	"github.com/consensys/go-corset/pkg/schema"
	"github.com/consensys/go-corset/pkg/trace"
	"github.com/consensys/go-corset/pkg/util"
	"github.com/consensys/go-corset/pkg/util/collection/set"
	"github.com/consensys/go-corset/pkg/util/field"
	"github.com/consensys/go-corset/pkg/util/math"
	"github.com/consensys/go-corset/pkg/util/source/sexp"
)

// Exp represents the a given value taken to a power.
type Exp[F field.Element[F], T Term[F, T]] struct {
	Arg T
	Pow uint64
}

// Exponent constructs a new expression representing the given argument
// raised to a given a given power.
func Exponent[F field.Element[F], T Term[F, T]](arg T, pow uint64) T {
	var term Term[F, T] = &Exp[F, T]{arg, pow}
	return term.(T)
}

// ApplyShift implementation for Term interface.
func (p *Exp[F, T]) ApplyShift(shift int) T {
	return Exponent[F, T](p.Arg.ApplyShift(shift), p.Pow)
}

// Bounds implementation for Boundable interface.
func (p *Exp[F, T]) Bounds() util.Bounds {
	return p.Arg.Bounds()
}

// EvalAt implementation for Evaluable interface.
func (p *Exp[F, T]) EvalAt(k int, tr trace.Module[F], sc schema.Module[F]) (F, error) {
	// Check whether argument evaluates to zero or not.
	val, err := p.Arg.EvalAt(k, tr, sc)
	// Compute exponent
	val = field.Pow(val, p.Pow)
	// Done
	return val, err
}

// Lisp implementation for Lispifiable interface.
func (p *Exp[F, T]) Lisp(global bool, mapping schema.RegisterMap) sexp.SExp {
	arg := p.Arg.Lisp(global, mapping)
	pow := sexp.NewSymbol(fmt.Sprintf("%d", p.Pow))

	return sexp.NewList([]sexp.SExp{sexp.NewSymbol("^"), arg, pow})
}

// RequiredRegisters implementation for Contextual interface.
func (p *Exp[F, T]) RequiredRegisters() *set.SortedSet[uint] {
	return p.Arg.RequiredRegisters()
}

// RequiredCells implementation for Contextual interface
func (p *Exp[F, T]) RequiredCells(row int, mid trace.ModuleId) *set.AnySortedSet[trace.CellRef] {
	return p.Arg.RequiredCells(row, mid)
}

// ShiftRange implementation for Term interface.
func (p *Exp[F, T]) ShiftRange() (int, int) {
	return p.Arg.ShiftRange()
}

// Substitute implementation for Substitutable interface.
func (p *Exp[F, T]) Substitute(mapping map[string]F) {
	p.Arg.Substitute(mapping)
}

// Simplify implementation for Term interface.
func (p *Exp[F, T]) Simplify(casts bool) T {
	var (
		arg  T          = p.Arg.Simplify(casts)
		targ Term[F, T] = arg
	)
	//
	if c, ok := targ.(*Constant[F, T]); ok {
		// Compute exponent (in place)
		val := field.Pow(c.Value, p.Pow)
		// Done
		targ = &Constant[F, T]{val}
	} else {
		targ = &Exp[F, T]{arg, p.Pow}
	}
	//
	return targ.(T)
}

// ValueRange implementation for Term interface.
func (p *Exp[F, T]) ValueRange(mapping schema.RegisterMap) math.Interval {
	bounds := p.Arg.ValueRange(mapping)
	bounds.Exp(uint(p.Pow))
	//
	return bounds
}
