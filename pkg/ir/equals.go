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
	"github.com/consensys/gnark-crypto/ecc/bls12-377/fr"
	"github.com/consensys/go-corset/pkg/schema"
	"github.com/consensys/go-corset/pkg/trace"
	"github.com/consensys/go-corset/pkg/util"
	"github.com/consensys/go-corset/pkg/util/collection/set"
	"github.com/consensys/go-corset/pkg/util/source/sexp"
)

// Equals constructs an Equal representing the equality of two expressions.
func Equals[S LogicalTerm[S], T Term[T]](lhs T, rhs T) S {
	var term LogicalTerm[S] = &Equal[S, T]{
		Lhs: lhs,
		Rhs: rhs,
	}
	//
	return term.(S)
}

// ============================================================================

// Equal represents an Equal between two terms (e.g. "X==Y", or "X!=Y+1",
// etc).  Equals are either equalities (or negated equalities) or
// inequalities.
type Equal[S LogicalTerm[S], T Term[T]] struct {
	Lhs Term[T]
	Rhs Term[T]
}

// ApplyShift implementation for LogicalTerm interface.
func (p *Equal[S, T]) ApplyShift(shift int) S {
	return Equals[S](p.Lhs.ApplyShift(shift), p.Rhs.ApplyShift(shift))
}

// ShiftRange implementation for LogicalTerm interface.
func (p *Equal[S, T]) ShiftRange() (int, int) {
	return shiftRangeOfTerms[T](p.Lhs.(T), p.Rhs.(T))
}

// Bounds implementation for Boundable interface.
func (p *Equal[S, T]) Bounds() util.Bounds {
	l := p.Lhs.Bounds()
	r := p.Rhs.Bounds()
	//
	l.Union(&r)
	//
	return l
}

// TestAt implementation for Testable interface.
func (p *Equal[S, T]) TestAt(k int, tr trace.Module, sc schema.Module) (bool, uint, error) {
	lhs, err1 := p.Lhs.EvalAt(k, tr, sc)
	rhs, err2 := p.Rhs.EvalAt(k, tr, sc)
	// error check
	if err1 != nil {
		return false, 0, err1
	} else if err2 != nil {
		return false, 0, err2
	}
	// perform comparison
	c := lhs.Cmp(&rhs)
	//
	return c == 0, 0, nil
}

// Lisp returns a lisp representation of this Equal, which is useful for
// debugging.
func (p *Equal[S, T]) Lisp(mapping schema.RegisterMap) sexp.SExp {
	var (
		l = p.Lhs.Lisp(mapping)
		r = p.Rhs.Lisp(mapping)
	)
	//
	return sexp.NewList([]sexp.SExp{
		sexp.NewSymbol("=="), l, r})
}

// RequiredRegisters implementation for Contextual interface.
func (p *Equal[S, T]) RequiredRegisters() *set.SortedSet[uint] {
	set := p.Lhs.RequiredRegisters()
	set.InsertSorted(p.Rhs.RequiredRegisters())
	//
	return set
}

// RequiredCells implementation for Contextual interface
func (p *Equal[S, T]) RequiredCells(row int, mid trace.ModuleId) *set.AnySortedSet[trace.CellRef] {
	set := p.Lhs.RequiredCells(row, mid)
	set.InsertSorted(p.Rhs.RequiredCells(row, mid))
	//
	return set
}

// Simplify this term as much as reasonably possible.
// nolint
func (p *Equal[S, T]) Simplify(casts bool) S {
	var (
		lhs = p.Lhs.Simplify(casts)
		rhs = p.Rhs.Simplify(casts)
	)
	//
	lc := IsConstant(lhs)
	rc := IsConstant(rhs)
	//
	if lc != nil && rc != nil {
		// Can simplify
		if lc.Cmp(rc) == 0 {
			return True[S]()
		}
		//
		return False[S]()
	}
	// Cannot simplify
	var tmp LogicalTerm[S] = &Equal[S, T]{lhs, rhs}
	// Done
	return tmp.(S)
}

// Substitute implementation for Substitutable interface.
func (p *Equal[S, T]) Substitute(mapping map[string]fr.Element) {
	p.Lhs.Substitute(mapping)
	p.Rhs.Substitute(mapping)
}
