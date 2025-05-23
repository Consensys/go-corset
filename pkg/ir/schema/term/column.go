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
	"math/big"

	"github.com/consensys/gnark-crypto/ecc/bls12-377/fr"
	"github.com/consensys/go-corset/pkg/ir/schema"
	"github.com/consensys/go-corset/pkg/trace"
	"github.com/consensys/go-corset/pkg/util"
	"github.com/consensys/go-corset/pkg/util/collection/set"
	"github.com/consensys/go-corset/pkg/util/source/sexp"
)

// ColumnAccess represents reading the value held at a given column in the
// tabular context.  Furthermore, the current row maybe shifted up (or down) by
// a given amount. Suppose we are evaluating a constraint on row k=5 which
// contains the column accesses "STAMP(0)" and "CT(-1)".  Then, STAMP(0)
// accesses the STAMP column at row 5, whilst CT(-1) accesses the CT column at
// row 4.
type ColumnAccess[T schema.Term[T]] struct {
	Column uint
	Shift  int
}

// Air indicates this term can be used at the AIR level.
func (p *ColumnAccess[T]) Air() {}

// ApplyShift implementation for Term interface.
func (p *ColumnAccess[T]) ApplyShift(shift int) T {
	//return &ColumnAccess[T]{Column: p.Column, Shift: p.Shift + shift}
	panic("got here")
}

// Bounds implementation for Boundable interface.
func (p *ColumnAccess[T]) Bounds() util.Bounds {
	if p.Shift >= 0 {
		// Positive shift
		return util.NewBounds(0, uint(p.Shift))
	}
	// Negative shift
	return util.NewBounds(uint(-p.Shift), 0)
}

// Branches implementation for Evaluable interface.
func (p *ColumnAccess[T]) Branches() uint {
	panic("todo")
}

// Context implementation for Contextual interface.
func (p *ColumnAccess[T]) Context(module schema.Module) trace.Context {
	return module.Column(p.Column).Context
}

// EvalAt implementation for Evaluable interface.
func (p *ColumnAccess[T]) EvalAt(k int, module trace.Module) (fr.Element, error) {
	return module.Column(p.Column).Get(k + p.Shift), nil
}

// Lisp implementation for Lispifiable interface.
func (p *ColumnAccess[T]) Lisp(module schema.Module) sexp.SExp {
	var name string
	// Generate name, whilst allowing for schema to be nil.
	if module != nil {
		name = module.Column(p.Column).QualifiedName(module)
	} else {
		name = fmt.Sprintf("#%d", p.Column)
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

// RequiredColumns implementation for Contextual interface.
func (p *ColumnAccess[T]) RequiredColumns() *set.SortedSet[uint] {
	r := set.NewSortedSet[uint]()
	r.Insert(p.Column)
	// Done
	return r
}

// RequiredCells implementation for Contextual interface
func (p *ColumnAccess[T]) RequiredCells(row int, tr trace.Module) *set.AnySortedSet[trace.CellRef] {
	set := set.NewAnySortedSet[trace.CellRef]()
	set.Insert(trace.NewCellRef(p.Column, row+p.Shift))
	//
	return set
}

// ShiftRange implementation for Term interface.
func (p *ColumnAccess[T]) ShiftRange() (int, int) {
	return p.Shift, p.Shift
}

// Simplify implementation for Term interface.
func (p *ColumnAccess[T]) Simplify(casts bool) T {
	panic("todo")
}

// ValueRange implementation for Term interface.
func (p *ColumnAccess[T]) ValueRange(module schema.Module) *util.Interval {
	bound := big.NewInt(2)
	width := int64(module.Column(p.Column).DataType.BitWidth())
	bound.Exp(bound, big.NewInt(width), nil)
	// Subtract 1 because interval is inclusive.
	bound.Sub(bound, big.NewInt(1))
	// Done
	return util.NewInterval(big.NewInt(0), bound)
}
