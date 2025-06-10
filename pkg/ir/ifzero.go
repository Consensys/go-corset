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

// IfZero returns the true branch when the condition evaluates to zero, and the
// false branch otherwise.
type IfZero[S LogicalTerm[S], T Term[T]] struct {
	// Elements contained within this list.
	Condition S
	// True branch
	TrueBranch T
	// False branch
	FalseBranch T
}

// IfElse constructs a new conditional with true and false branches.  Note, the
// true branch is taken when the condition evaluates to zero.
func IfElse[S LogicalTerm[S], T Term[T]](condition S, trueBranch T, falseBranch T) T {
	var term Term[T] = &IfZero[S, T]{condition, trueBranch, falseBranch}
	return term.(T)
}

// ApplyShift implementation for Term interface.
func (p *IfZero[S, T]) ApplyShift(int) T {
	// Need to add ApplyShift to LogicalTerm
	panic("todo")
}

// Bounds returns max shift in either the negative (left) or positive
// direction (right).
func (p *IfZero[S, T]) Bounds() util.Bounds {
	c := p.Condition.Bounds()
	// Get bounds for true branch
	tbounds := p.TrueBranch.Bounds()
	c.Union(&tbounds)
	// Get bounds for false branch
	fbounds := p.FalseBranch.Bounds()
	c.Union(&fbounds)
	// Done
	return c
}

// EvalAt implementation for Evaluable interface.
func (p *IfZero[S, T]) EvalAt(k int, tr trace.Module) (fr.Element, error) {
	// Evaluate condition
	cond, _, err := p.Condition.TestAt(k, tr)
	//
	if err != nil {
		return fr.Element{}, err
	} else if cond {
		return p.TrueBranch.EvalAt(k, tr)
	}
	//
	return p.FalseBranch.EvalAt(k, tr)
}

// Lisp implementation for Lispifiable interface.
func (p *IfZero[S, T]) Lisp(module schema.Module) sexp.SExp {
	// Translate Condition
	condition := p.Condition.Lisp(module)
	// Dispatch on type
	return sexp.NewList([]sexp.SExp{
		sexp.NewSymbol("if"),
		condition,
		p.TrueBranch.Lisp(module),
		p.FalseBranch.Lisp(module),
	})
}

// RequiredRegisters implementation for Contextual interface.
func (p *IfZero[S, T]) RequiredRegisters() *set.SortedSet[uint] {
	set := p.Condition.RequiredRegisters()
	// Include true branch
	set.InsertSorted(p.TrueBranch.RequiredRegisters())
	// Include false branch
	set.InsertSorted(p.FalseBranch.RequiredRegisters())
	// Done
	return set
}

// RequiredCells implementation for Contextual interface
func (p *IfZero[S, T]) RequiredCells(row int, tr trace.Module) *set.AnySortedSet[trace.CellRef] {
	set := p.Condition.RequiredCells(row, tr)
	// Include true branch
	set.InsertSorted(p.TrueBranch.RequiredCells(row, tr))
	// Include false branch
	set.InsertSorted(p.FalseBranch.RequiredCells(row, tr))
	// Done
	return set
}

// ShiftRange implementation for Term interface.
func (p *IfZero[S, T]) ShiftRange() (int, int) {
	panic("todo")
}

// ValueRange implementation for Term interface.
func (p *IfZero[S, T]) ValueRange(module schema.Module) *util.Interval {
	panic("todo")
}

// Simplify implementation for Term interface.
//
// nolint
func (p *IfZero[S, T]) Simplify(casts bool) T {
	var (
		cond        = p.Condition.Simplify(casts)
		trueBranch  = p.TrueBranch.Simplify(casts)
		falseBranch = p.FalseBranch.Simplify(casts)
	)
	// Handle reductive cases
	if IsTrue(cond) {
		return trueBranch
	} else if IsFalse(cond) {
		return falseBranch
	}
	// Done
	var term Term[T] = &IfZero[S, T]{cond, trueBranch, falseBranch}
	//
	return term.(T)
}
