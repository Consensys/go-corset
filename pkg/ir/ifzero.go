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

// IfZero returns the (optional) true branch when the condition evaluates to zero, and
// the (optional false branch otherwise.
type IfZero[T Term[T]] struct {
	// Elements contained within this list.
	Condition Term[T]
	// True branch (optional).
	TrueBranch Term[T]
	// False branch (optional).
	FalseBranch Term[T]
}

// If constructs a new conditional branch, where the true branch is taken when
// the condition evaluates to zero.
func If[T Term[T]](condition T, trueBranch T) T {
	var term Term[T] = &IfZero[T]{condition, trueBranch, nil}
	return term.(T)
}

// IfElse constructs a new conditional branch, where either the true branch or
// the false branch can (optionally) be nil (but both cannot).  Note, the true
// branch is taken when the condition evaluates to zero.
func IfElse[T Term[T]](condition T, trueBranch T, falseBranch T) T {
	var term Term[T] = &IfZero[T]{condition, trueBranch, falseBranch}
	return term.(T)
}

// ApplyShift implementation for Term interface.
func (p *IfZero[T]) ApplyShift(int) T {
	panic("todo")
}

// Bounds returns max shift in either the negative (left) or positive
// direction (right).
func (p *IfZero[T]) Bounds() util.Bounds {
	c := p.Condition.Bounds()
	// Get bounds for true branch (if applicable)
	if p.TrueBranch != nil {
		tbounds := p.TrueBranch.Bounds()
		c.Union(&tbounds)
	}
	// Get bounds for false branch (if applicable)
	if p.FalseBranch != nil {
		fbounds := p.FalseBranch.Bounds()
		c.Union(&fbounds)
	}
	// Done
	return c
}

// EvalAt implementation for Evaluable interface.
func (p *IfZero[T]) EvalAt(k int, tr trace.Module) (fr.Element, error) {
	// Evaluate condition
	cond, err := p.Condition.EvalAt(k, tr)
	//
	if err != nil {
		return cond, err
	} else if cond.IsZero() && p.TrueBranch != nil {
		return p.TrueBranch.EvalAt(k, tr)
	} else if !cond.IsZero() && p.FalseBranch != nil {
		return p.FalseBranch.EvalAt(k, tr)
	}
	//
	return frZERO, nil
}

// Lisp implementation for Lispifiable interface.
func (p *IfZero[T]) Lisp(module schema.Module) sexp.SExp {
	// Translate Condition
	condition := p.Condition.Lisp(module)
	// Dispatch on type
	if p.FalseBranch == nil {
		return sexp.NewList([]sexp.SExp{
			sexp.NewSymbol("if"),
			condition,
			p.TrueBranch.Lisp(module),
		})
	} else if p.TrueBranch == nil {
		return sexp.NewList([]sexp.SExp{
			sexp.NewSymbol("ifnot"),
			condition,
			p.FalseBranch.Lisp(module),
		})
	}

	return sexp.NewList([]sexp.SExp{
		sexp.NewSymbol("if"),
		condition,
		p.TrueBranch.Lisp(module),
		p.FalseBranch.Lisp(module),
	})
}

// RequiredRegisters implementation for Contextual interface.
func (p *IfZero[T]) RequiredRegisters() *set.SortedSet[uint] {
	set := p.Condition.RequiredRegisters()
	// Include true branch (if applicable)
	if p.TrueBranch != nil {
		set.InsertSorted(p.TrueBranch.RequiredRegisters())
	}
	// Include false branch (if applicable)
	if p.FalseBranch != nil {
		set.InsertSorted(p.FalseBranch.RequiredRegisters())
	}
	// Done
	return set
}

// RequiredCells implementation for Contextual interface
func (p *IfZero[T]) RequiredCells(row int, tr trace.Module) *set.AnySortedSet[trace.CellRef] {
	set := p.Condition.RequiredCells(row, tr)
	// Include true branch (if applicable)
	if p.TrueBranch != nil {
		set.InsertSorted(p.TrueBranch.RequiredCells(row, tr))
	}
	// Include false branch (if applicable)
	if p.FalseBranch != nil {
		set.InsertSorted(p.FalseBranch.RequiredCells(row, tr))
	}
	// Done
	return set
}

// ShiftRange implementation for Term interface.
func (p *IfZero[T]) ShiftRange() (int, int) {
	panic("todo")
}

// Simplify implementation for Term interface.
func (p *IfZero[T]) Simplify(casts bool) T {
	panic("todo")
}

// ValueRange implementation for Term interface.
func (p *IfZero[T]) ValueRange(module schema.Module) *util.Interval {
	panic("todo")
}
