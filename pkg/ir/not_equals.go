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

// NotEqual represents an NotEqual between two terms (e.g. "X==Y", or "X!=Y+1",
// etc).  NotEquals are either NotEqualities (or negated NotEqualities) or
// inNotEqualities.
type NotEqual[S LogicalTerm[S], T Term[T]] struct {
	Lhs Term[T]
	Rhs Term[T]
}

// NotEquals constructs an NotEqual representing the NotEquality of two expressions.
func NotEquals[S LogicalTerm[S], T Term[T]](lhs T, rhs T) S {
	var (
		term LogicalTerm[S] = &NotEqual[S, T]{
			Lhs: lhs,
			Rhs: rhs,
		}
		res, ok = term.(S)
	)
	//
	if ok {
		return res
	}
	// Sanity check
	panic("invalid logical AIR term")
}

// ApplyShift implementation for LogicalTerm interface.
func (p *NotEqual[S, T]) ApplyShift(shift int) S {
	return NotEquals[S](p.Lhs.ApplyShift(shift), p.Rhs.ApplyShift(shift))
}

// ShiftRange implementation for LogicalTerm interface.
func (p *NotEqual[S, T]) ShiftRange() (int, int) {
	return shiftRangeOfTerms[T](p.Lhs.(T), p.Rhs.(T))
}

// Bounds implementation for Boundable interface.
func (p *NotEqual[S, T]) Bounds() util.Bounds {
	l := p.Lhs.Bounds()
	r := p.Rhs.Bounds()
	//
	l.Union(&r)
	//
	return l
}

// TestAt implementation for Testable interface.
func (p *NotEqual[S, T]) TestAt(k int, tr trace.Module, sc schema.Module) (bool, uint, error) {
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
	return c != 0, 0, nil
}

// Lisp returns a lisp representation of this NotEqual, which is useful for
// debugging.
func (p *NotEqual[S, T]) Lisp(module schema.Module) sexp.SExp {
	var (
		l = p.Lhs.Lisp(module)
		r = p.Rhs.Lisp(module)
	)
	//
	return sexp.NewList([]sexp.SExp{
		sexp.NewSymbol("!="), l, r})
}

// RequiredRegisters implementation for Contextual interface.
func (p *NotEqual[S, T]) RequiredRegisters() *set.SortedSet[uint] {
	set := p.Lhs.RequiredRegisters()
	set.InsertSorted(p.Rhs.RequiredRegisters())
	//
	return set
}

// RequiredCells implementation for Contextual interface
func (p *NotEqual[S, T]) RequiredCells(row int, mid trace.ModuleId) *set.AnySortedSet[trace.CellRef] {
	set := p.Lhs.RequiredCells(row, mid)
	set.InsertSorted(p.Rhs.RequiredCells(row, mid))
	//
	return set
}

// Simplify this term as much as reasonably possible.
//
// nolint
func (p *NotEqual[S, T]) Simplify(casts bool) S {
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
			return False[S]()
		}
		//
		return True[S]()
	}
	// Cannot simplify
	var tmp LogicalTerm[S] = &NotEqual[S, T]{lhs, rhs}
	// Done
	return tmp.(S)
}

// Substitute implementation for Substitutable interface.
func (p *NotEqual[S, T]) Substitute(mapping map[string]fr.Element) {
	p.Lhs.Substitute(mapping)
	p.Rhs.Substitute(mapping)
}
