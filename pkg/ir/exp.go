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

	"github.com/consensys/gnark-crypto/ecc/bls12-377/fr"
	"github.com/consensys/go-corset/pkg/schema"
	"github.com/consensys/go-corset/pkg/trace"
	"github.com/consensys/go-corset/pkg/util"
	"github.com/consensys/go-corset/pkg/util/collection/set"
	"github.com/consensys/go-corset/pkg/util/source/sexp"
)

// Exp represents the a given value taken to a power.
type Exp[T Term[T]] struct {
	Arg T
	Pow uint64
}

// Exponent constructs a new expression representing the given argument
// raised to a given a given power.
func Exponent[T Term[T]](arg T, pow uint64) T {
	var term Term[T] = &Exp[T]{arg, pow}
	return term.(T)
}

// ApplyShift implementation for Term interface.
func (p *Exp[T]) ApplyShift(shift int) T {
	return Exponent(p.Arg.ApplyShift(shift), p.Pow)
}

// Bounds implementation for Boundable interface.
func (p *Exp[T]) Bounds() util.Bounds {
	return p.Arg.Bounds()
}

// EvalAt implementation for Evaluable interface.
func (p *Exp[T]) EvalAt(k int, tr trace.Module) (fr.Element, error) {
	// Check whether argument evaluates to zero or not.
	val, err := p.Arg.EvalAt(k, tr)
	// Compute exponent
	util.Pow(&val, p.Pow)
	// Done
	return val, err
}

// Lisp implementation for Lispifiable interface.
func (p *Exp[T]) Lisp(module schema.Module) sexp.SExp {
	arg := p.Arg.Lisp(module)
	pow := sexp.NewSymbol(fmt.Sprintf("%d", p.Pow))

	return sexp.NewList([]sexp.SExp{sexp.NewSymbol("^"), arg, pow})
}

// RequiredRegisters implementation for Contextual interface.
func (p *Exp[T]) RequiredRegisters() *set.SortedSet[uint] {
	return p.Arg.RequiredRegisters()
}

// RequiredCells implementation for Contextual interface
func (p *Exp[T]) RequiredCells(row int, mid trace.ModuleId) *set.AnySortedSet[trace.CellRef] {
	return p.Arg.RequiredCells(row, mid)
}

// ShiftRange implementation for Term interface.
func (p *Exp[T]) ShiftRange() (int, int) {
	return p.Arg.ShiftRange()
}

// Substitute implementation for Substitutable interface.
func (p *Exp[T]) Substitute(mapping map[string]fr.Element) {
	p.Arg.Substitute(mapping)
}

// Simplify implementation for Term interface.
func (p *Exp[T]) Simplify(casts bool) T {
	var (
		arg  T       = p.Arg.Simplify(casts)
		targ Term[T] = arg
	)
	//
	if c, ok := targ.(*Constant[T]); ok {
		var val fr.Element
		// Clone value
		val.Set(&c.Value)
		// Compute exponent (in place)
		util.Pow(&val, p.Pow)
		// Done
		targ = &Constant[T]{val}
	} else {
		targ = &Exp[T]{arg, p.Pow}
	}
	//
	return targ.(T)
}

// ValueRange implementation for Term interface.
func (p *Exp[T]) ValueRange(module schema.Module) *util.Interval {
	bounds := p.Arg.ValueRange(module)
	bounds.Exp(uint(p.Pow))
	//
	return bounds
}
