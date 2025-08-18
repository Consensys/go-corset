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
	"github.com/consensys/go-corset/pkg/schema"
	"github.com/consensys/go-corset/pkg/trace"
	"github.com/consensys/go-corset/pkg/util"
	"github.com/consensys/go-corset/pkg/util/collection/set"
	"github.com/consensys/go-corset/pkg/util/field"
	"github.com/consensys/go-corset/pkg/util/source/sexp"
)

// LessThan constructs an Inequality representing the X < Y.
func LessThan[F field.Element[F], S LogicalTerm[F, S], T Term[F, T]](lhs T, rhs T) S {
	var term LogicalTerm[F, S] = &Inequality[F, S, T]{
		Strict: true,
		Lhs:    lhs,
		Rhs:    rhs,
	}
	//
	return term.(S)
}

// LessThanOrEquals constructs an Inequality representing the X <= Y.
func LessThanOrEquals[F field.Element[F], S LogicalTerm[F, S], T Term[F, T]](lhs T, rhs T) S {
	var term LogicalTerm[F, S] = &Inequality[F, S, T]{
		Strict: false,
		Lhs:    lhs,
		Rhs:    rhs,
	}
	//
	return term.(S)
}

// GreaterThan constructs an Inequality representing the X > Y.
func GreaterThan[F field.Element[F], S LogicalTerm[F, S], T Term[F, T]](lhs T, rhs T) S {
	var term LogicalTerm[F, S] = &Inequality[F, S, T]{
		Strict: true,
		Lhs:    rhs,
		Rhs:    lhs,
	}
	//
	return term.(S)
}

// GreaterThanOrEquals constructs an Inequality representing the X >= Y.
func GreaterThanOrEquals[F field.Element[F], S LogicalTerm[F, S], T Term[F, T]](lhs T, rhs T) S {
	var term LogicalTerm[F, S] = &Inequality[F, S, T]{
		Strict: false,
		Lhs:    rhs,
		Rhs:    lhs,
	}
	//
	return term.(S)
}

// ============================================================================

// Inequality represents an inequality between two terms (e.g. "X<Y", or "X<=Y+1",
// etc).  Inequalitys are either Inequalityities (or negated Inequalityities) or
// inInequalityities.
type Inequality[F field.Element[F], S LogicalTerm[F, S], T Term[F, T]] struct {
	// Strict indicates whether its strictly less-than, or whether its less-than
	// or equals.
	Strict bool
	// Left hand side of the inequality
	Lhs Term[F, T]
	// Right hand side of the inequality
	Rhs Term[F, T]
}

// ApplyShift implementation for LogicalTerm interface.
func (p *Inequality[F, S, T]) ApplyShift(shift int) S {
	if p.Strict {
		return LessThan[F, S](p.Lhs.ApplyShift(shift),
			p.Rhs.ApplyShift(shift))
	}
	//
	return LessThanOrEquals[F, S](p.Lhs.ApplyShift(shift),
		p.Rhs.ApplyShift(shift))
}

// ShiftRange implementation for LogicalTerm interface.
func (p *Inequality[F, S, T]) ShiftRange() (int, int) {
	return shiftRangeOfTerms(p.Lhs.(T), p.Rhs.(T))
}

// Bounds implementation for Boundable interface.
func (p *Inequality[F, S, T]) Bounds() util.Bounds {
	l := p.Lhs.Bounds()
	r := p.Rhs.Bounds()
	//
	l.Union(&r)
	//
	return l
}

// TestAt implementation for Testable interface.
func (p *Inequality[F, S, T]) TestAt(k int, tr trace.Module[F], sc schema.Module[F]) (bool, uint, error) {
	lhs, err1 := p.Lhs.EvalAt(k, tr, sc)
	rhs, err2 := p.Rhs.EvalAt(k, tr, sc)
	// error check
	if err1 != nil {
		return false, 0, err1
	} else if err2 != nil {
		return false, 0, err2
	}
	// perform comparison
	c := lhs.Cmp(rhs)
	//
	if p.Strict {
		return c < 0, 0, nil
	}
	//
	return c <= 0, 0, nil
}

// Lisp returns a lisp representation of this Inequality, which is useful for
// debugging.
func (p *Inequality[F, S, T]) Lisp(global bool, mapping schema.RegisterMap) sexp.SExp {
	var (
		l      = p.Lhs.Lisp(global, mapping)
		r      = p.Rhs.Lisp(global, mapping)
		symbol string
	)
	//
	if p.Strict {
		symbol = "<"
	} else {
		symbol = "<="
	}
	//
	return sexp.NewList([]sexp.SExp{
		sexp.NewSymbol(symbol), l, r})
}

// RequiredRegisters implementation for Contextual interface.
func (p *Inequality[F, S, T]) RequiredRegisters() *set.SortedSet[uint] {
	set := p.Lhs.RequiredRegisters()
	set.InsertSorted(p.Rhs.RequiredRegisters())
	//
	return set
}

// RequiredCells implementation for Contextual interface
func (p *Inequality[F, S, T]) RequiredCells(row int, tr trace.ModuleId) *set.AnySortedSet[trace.CellRef] {
	set := p.Lhs.RequiredCells(row, tr)
	set.InsertSorted(p.Rhs.RequiredCells(row, tr))
	//
	return set
}

// Simplify this term as much as reasonably possible.
// nolint
func (p *Inequality[F, S, T]) Simplify(casts bool) S {
	var (
		lhs = p.Lhs.Simplify(casts)
		rhs = p.Rhs.Simplify(casts)
	)
	//
	lc, lok := IsConstant[F](lhs)
	rc, rok := IsConstant[F](rhs)
	//
	if lok && rok {
		c := lc.Cmp(rc)
		// Can simplify
		if p.Strict && c < 0 {
			return True[F, S]()
		} else if !p.Strict && c <= 0 {
			return True[F, S]()
		}
		// Fail
		return False[F, S]()
	}
	// Cannot simplify
	var tmp LogicalTerm[F, S] = &Inequality[F, S, T]{p.Strict, lhs, rhs}
	// Done
	return tmp.(S)
}

// Substitute implementation for Substitutable interface.
func (p *Inequality[F, S, T]) Substitute(mapping map[string]F) {
	p.Lhs.Substitute(mapping)
	p.Rhs.Substitute(mapping)
}
