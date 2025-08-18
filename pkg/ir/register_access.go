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
	"strings"

	"github.com/consensys/go-corset/pkg/schema"
	"github.com/consensys/go-corset/pkg/trace"
	"github.com/consensys/go-corset/pkg/util"
	"github.com/consensys/go-corset/pkg/util/collection/set"
	"github.com/consensys/go-corset/pkg/util/field"
	util_math "github.com/consensys/go-corset/pkg/util/math"
	"github.com/consensys/go-corset/pkg/util/source/sexp"
)

// RegisterAccess represents reading the value held at a given register in the
// tabular context.  Furthermore, the current row maybe shifted up (or down) by
// a given amount. Suppose we are evaluating a constraint on row k=5 which
// contains the register accesses "STAMP(0)" and "CT(-1)".  Then, STAMP(0)
// accesses the STAMP register at row 5, whilst CT(-1) accesses the CT register at
// row 4.
type RegisterAccess[F field.Element[F], T Term[F, T]] struct {
	Register schema.RegisterId
	Shift    int
}

// NewRegisterAccess constructs an AIR expression representing the value of a
// given register on the current row.
func NewRegisterAccess[F field.Element[F], T Term[F, T]](register schema.RegisterId, shift int) T {
	var term Term[F, T] = &RegisterAccess[F, T]{Register: register, Shift: shift}
	return term.(T)
}

// RawRegisterAccess constructs an AIR expression representing the value of a given
// register on the current row.
func RawRegisterAccess[F field.Element[F], T Term[F, T]](register schema.RegisterId, shift int) *RegisterAccess[F, T] {
	return &RegisterAccess[F, T]{Register: register, Shift: shift}
}

// Air indicates this term can be used at the AIR level.
func (p *RegisterAccess[F, T]) Air() {}

// ApplyShift implementation for Term interface.
func (p *RegisterAccess[F, T]) ApplyShift(shift int) T {
	var reg Term[F, T] = &RegisterAccess[F, T]{Register: p.Register, Shift: p.Shift + shift}
	return reg.(T)
}

// Bounds implementation for Boundable interface.
func (p *RegisterAccess[F, T]) Bounds() util.Bounds {
	if p.Shift >= 0 {
		// Positive shift
		return util.NewBounds(0, uint(p.Shift))
	}
	// Negative shift
	return util.NewBounds(uint(-p.Shift), 0)
}

// EvalAt implementation for Evaluable interface.
func (p *RegisterAccess[F, T]) EvalAt(k int, module trace.Module[F], _ schema.Module[F]) (F, error) {
	return module.Column(p.Register.Unwrap()).Get(k + p.Shift), nil
}

// IsDefined implementation for Evaluable interface.
func (p *RegisterAccess[F, T]) IsDefined() bool {
	return p.Register.IsUsed()
}

// Lisp implementation for Lispifiable interface.
func (p *RegisterAccess[F, T]) Lisp(global bool, mapping schema.RegisterMap) sexp.SExp {
	var name string
	// Generate name, whilst allowing for schema to be nil.
	if mapping != nil && global {
		name = mapping.Register(p.Register).QualifiedName(mapping)
	} else if mapping != nil {
		name = mapping.Register(p.Register).Name
		// Add quotes if suitable
		if strings.Contains(name, " ") {
			name = fmt.Sprintf("\"%s\"", name)
		}
	} else {
		name = fmt.Sprintf("#%d", p.Register)
	}
	//
	access := sexp.NewSymbol(name)
	// Check whether shifted (or not)
	if p.Shift == 0 {
		// Not shifted
		return access
	}
	// Shifted
	shift := sexp.NewSymbol(fmt.Sprintf("%d", p.Shift))

	return sexp.NewList([]sexp.SExp{sexp.NewSymbol("shift"), access, shift})
}

// RequiredRegisters implementation for Contextual interface.
func (p *RegisterAccess[F, T]) RequiredRegisters() *set.SortedSet[uint] {
	r := set.NewSortedSet[uint]()
	r.Insert(p.Register.Unwrap())
	// Done
	return r
}

// RequiredCells implementation for Contextual interface
func (p *RegisterAccess[F, T]) RequiredCells(row int, mid trace.ModuleId) *set.AnySortedSet[trace.CellRef] {
	var (
		set = set.NewAnySortedSet[trace.CellRef]()
		ref = trace.NewColumnRef(mid, p.Register)
	)
	//
	set.Insert(trace.NewCellRef(ref, row+p.Shift))
	//
	return set
}

// ShiftRange implementation for Term interface.
func (p *RegisterAccess[F, T]) ShiftRange() (int, int) {
	return p.Shift, p.Shift
}

// Simplify implementation for Term interface.
func (p *RegisterAccess[F, T]) Simplify(casts bool) T {
	var tmp Term[F, T] = p
	return tmp.(T)
}

// Substitute implementation for Substitutable interface.
func (p *RegisterAccess[F, T]) Substitute(mapping map[string]F) {

}

// ValueRange implementation for Term interface.
func (p *RegisterAccess[F, T]) ValueRange(mapping schema.RegisterMap) util_math.Interval {
	var width = mapping.Register(p.Register).Width
	// NOTE: the following is necessary because MaxUint is permitted as a signal
	// that the given register has no fixed bitwidth.  Rather, it can consume
	// all possible values of the underlying field element.
	if width == math.MaxUint {
		return util_math.INFINITY
	}
	//
	return valueRangeOfBits(width)
}

func valueRangeOfBits(bitwidth uint) util_math.Interval {
	var bound = big.NewInt(2)
	//
	bound.Exp(bound, big.NewInt(int64(bitwidth)), nil)
	// Subtract 1 because interval is inclusive.
	bound.Sub(bound, &biONE)
	// Done
	return util_math.NewInterval(biZERO, *bound)
}
