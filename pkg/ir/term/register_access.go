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
package term

import (
	"fmt"
	"math"
	"math/big"
	"strings"

	"github.com/consensys/go-corset/pkg/schema/register"
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
type RegisterAccess[F field.Element[F], T Expr[F, T]] struct {
	// Id for register being accessed
	register register.Id
	// Bitwidth of access.  This can be math.MaxUint to signal the entire
	// register is being read; otherwise, it can be a width below that of the
	// given register to signal a cast.
	bitwidth uint
	// Relative shift of access.  This indicates on which row (relative to the
	// current row) the given register is being read.
	shift int
	// Bound is precomputed when bitwidth != MaxUint for efficiency.
	bound F
}

// NewRegisterAccess constructs an AIR expression representing the value of a
// given register on the current row.
func NewRegisterAccess[F field.Element[F], T Expr[F, T]](register register.Id, shift int) T {
	var term Expr[F, T] = NarrowRegisterAccess[F, T](register, math.MaxUint, shift)
	return term.(T)
}

// RawRegisterAccess constructs an AIR expression representing the value of a given
// register on the current row.
func RawRegisterAccess[F field.Element[F], T Expr[F, T]](register register.Id, shift int) *RegisterAccess[F, T] {
	return NarrowRegisterAccess[F, T](register, math.MaxUint, shift)
}

// NarrowRegisterAccess constructs an AIR expression representing the value of a
// given register on the current row.  Additionally, the bitwidth can be
// specified so as to narrow the width of the register being read (i.e. for
// casting).
func NarrowRegisterAccess[F field.Element[F], T Expr[F, T]](register register.Id, bitwidth uint, shift int,
) *RegisterAccess[F, T] {
	var bound F
	// Precompute 2^bitwidth (if applicable)
	if bitwidth != math.MaxUint {
		bound = field.TwoPowN[F](bitwidth)
	}
	//
	return &RegisterAccess[F, T]{register, bitwidth, shift, bound}
}

// Air indicates this term can be used at the AIR level.
func (p *RegisterAccess[F, T]) Air() {}

// Register returns the id of the register being accessed.
func (p *RegisterAccess[F, T]) Register() register.Id {
	return p.register
}

// Bitwidth returns the width of this access.  This can be math.MaxUint to
// signal an "unbounded" access; otherwise, it can be the actual register's
// width or below.  If below, then this signals a cast.
func (p *RegisterAccess[F, T]) Bitwidth() uint {
	return p.bitwidth
}

// Shift returns the relative shift of this access.
func (p *RegisterAccess[F, T]) Shift() int {
	return p.shift
}

// ApplyShift implementation for Term interface.
func (p *RegisterAccess[F, T]) ApplyShift(shift int) T {
	var reg Expr[F, T] = &RegisterAccess[F, T]{
		p.register, p.bitwidth, p.shift + shift, p.bound,
	}
	//
	return reg.(T)
}

// Bounds implementation for Boundable interface.
func (p *RegisterAccess[F, T]) Bounds() util.Bounds {
	if p.shift >= 0 {
		// Positive shift
		return util.NewBounds(0, uint(p.shift))
	}
	// Negative shift
	return util.NewBounds(uint(-p.shift), 0)
}

// EvalAt implementation for Evaluable interface.
func (p *RegisterAccess[F, T]) EvalAt(k int, module trace.Module[F], _ register.Map) (F, error) {
	var (
		val = module.Column(p.register.Unwrap()).Get(k + p.shift)
		err error
	)
	// Dynamic cast
	if p.bitwidth != math.MaxUint && val.Cmp(p.bound) >= 0 {
		// Construct error
		err = fmt.Errorf("read failure (value %s not u%d)", val.String(), p.bitwidth)
	}
	//
	return val, err
}

// IsDefined implementation for Evaluable interface.
func (p *RegisterAccess[F, T]) IsDefined() bool {
	return p.register.IsUsed()
}

// Lisp implementation for Lispifiable interface.
func (p *RegisterAccess[F, T]) Lisp(global bool, mapping register.Map) sexp.SExp {
	var name string
	// Generate name, whilst allowing for schema to be nil.
	if mapping != nil && global {
		name = mapping.Register(p.register).QualifiedName(mapping)
	} else if mapping != nil {
		name = mapping.Register(p.register).Name
		// Add quotes if suitable
		if strings.Contains(name, " ") {
			name = fmt.Sprintf("\"%s\"", name)
		}
	} else {
		name = fmt.Sprintf("#%d", p.register)
	}
	//
	var access sexp.SExp = sexp.NewSymbol(name)
	// Check whether shifted (or not)
	if p.shift != 0 {
		// Shifted
		shift := sexp.NewSymbol(fmt.Sprintf("%d", p.shift))
		//
		access = sexp.NewList([]sexp.SExp{sexp.NewSymbol("shift"), access, shift})
	}
	//
	if p.bitwidth != math.MaxUint {
		tw := fmt.Sprintf("u%d", p.bitwidth)
		access = sexp.NewList([]sexp.SExp{sexp.NewSymbol(tw), access})
	}
	//
	return access
}

// RequiredRegisters implementation for Contextual interface.
func (p *RegisterAccess[F, T]) RequiredRegisters() *set.SortedSet[uint] {
	r := set.NewSortedSet[uint]()
	r.Insert(p.register.Unwrap())
	// Done
	return r
}

// RequiredCells implementation for Contextual interface
func (p *RegisterAccess[F, T]) RequiredCells(row int, mid trace.ModuleId) *set.AnySortedSet[trace.CellRef] {
	var (
		set = set.NewAnySortedSet[trace.CellRef]()
		ref = trace.NewColumnRef(mid, p.register)
	)
	//
	set.Insert(trace.NewCellRef(ref, row+p.shift))
	//
	return set
}

// ShiftRange implementation for Term interface.
func (p *RegisterAccess[F, T]) ShiftRange() (int, int) {
	return p.shift, p.shift
}

// Simplify implementation for Term interface.
func (p *RegisterAccess[F, T]) Simplify(casts bool) T {
	var tmp Expr[F, T] = p
	return tmp.(T)
}

// Substitute implementation for Substitutable interface.
func (p *RegisterAccess[F, T]) Substitute(mapping map[string]F) {

}

// ValueRange implementation for Term interface.
func (p *RegisterAccess[F, T]) ValueRange(mapping register.Map) util_math.Interval {
	var (
		width = mapping.Register(p.register).Width
	)
	// NOTE: the following is necessary because MaxUint is permitted as a signal
	// that the given register has no fixed bitwidth.  Rather, it can consume
	// all possible values of the underlying field element.
	if width == math.MaxUint {
		return util_math.INFINITY
	}
	//
	return valueRangeOfBits(min(width, p.bitwidth))
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
