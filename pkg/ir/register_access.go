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

// RegisterAccess represents reading the value held at a given register in the
// tabular context.  Furthermore, the current row maybe shifted up (or down) by
// a given amount. Suppose we are evaluating a constraint on row k=5 which
// contains the register accesses "STAMP(0)" and "CT(-1)".  Then, STAMP(0)
// accesses the STAMP register at row 5, whilst CT(-1) accesses the CT register at
// row 4.
type RegisterAccess[T Term[T]] struct {
	Register schema.RegisterId
	Shift    int
}

// NewRegisterAccess constructs an AIR expression representing the value of a
// given register on the current row.
func NewRegisterAccess[T Term[T]](register schema.RegisterId, shift int) T {
	var term Term[T] = &RegisterAccess[T]{Register: register, Shift: shift}
	return term.(T)
}

// RawRegisterAccess constructs an AIR expression representing the value of a given
// register on the current row.
func RawRegisterAccess[T Term[T]](register schema.RegisterId, shift int) *RegisterAccess[T] {
	return &RegisterAccess[T]{Register: register, Shift: shift}
}

// Air indicates this term can be used at the AIR level.
func (p *RegisterAccess[T]) Air() {}

// ApplyShift implementation for Term interface.
func (p *RegisterAccess[T]) ApplyShift(shift int) T {
	var reg Term[T] = &RegisterAccess[T]{Register: p.Register, Shift: p.Shift + shift}
	return reg.(T)
}

// Bounds implementation for Boundable interface.
func (p *RegisterAccess[T]) Bounds() util.Bounds {
	if p.Shift >= 0 {
		// Positive shift
		return util.NewBounds(0, uint(p.Shift))
	}
	// Negative shift
	return util.NewBounds(uint(-p.Shift), 0)
}

// EvalAt implementation for Evaluable interface.
func (p *RegisterAccess[T]) EvalAt(k int, module trace.Module) (fr.Element, error) {
	return module.Column(p.Register.Unwrap()).Get(k + p.Shift), nil
}

// Lisp implementation for Lispifiable interface.
func (p *RegisterAccess[T]) Lisp(module schema.Module) sexp.SExp {
	var name string
	// Generate name, whilst allowing for schema to be nil.
	if module != nil {
		name = module.Register(p.Register).QualifiedName(module)
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
func (p *RegisterAccess[T]) RequiredRegisters() *set.SortedSet[uint] {
	r := set.NewSortedSet[uint]()
	r.Insert(p.Register.Unwrap())
	// Done
	return r
}

// RequiredCells implementation for Contextual interface
func (p *RegisterAccess[T]) RequiredCells(row int, tr trace.Module) *set.AnySortedSet[trace.CellRef] {
	set := set.NewAnySortedSet[trace.CellRef]()
	set.Insert(trace.NewCellRef(p.Register.Unwrap(), row+p.Shift))
	//
	return set
}

// ShiftRange implementation for Term interface.
func (p *RegisterAccess[T]) ShiftRange() (int, int) {
	return p.Shift, p.Shift
}

// Simplify implementation for Term interface.
func (p *RegisterAccess[T]) Simplify(casts bool) T {
	var tmp Term[T] = p
	return tmp.(T)
}

// ValueRange implementation for Term interface.
func (p *RegisterAccess[T]) ValueRange(module schema.Module) *util.Interval {
	bound := big.NewInt(2)
	width := int64(module.Register(p.Register).Width)
	bound.Exp(bound, big.NewInt(width), nil)
	// Subtract 1 because interval is inclusive.
	bound.Sub(bound, big.NewInt(1))
	// Done
	return util.NewInterval(big.NewInt(0), bound)
}
