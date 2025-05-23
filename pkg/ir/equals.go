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
	"github.com/consensys/go-corset/pkg/util/source/sexp"
)

// Equals constructs an Equal representing the equality of two expressions.
func Equals[S LogicalTerm[S], T Term[T]](lhs T, rhs T) S {
	var term LogicalTerm[S] = &Equal[T]{
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
type Equal[T Term[T]] struct {
	Lhs Term[T]
	Rhs Term[T]
}

// Air indicates this term can be used at the AIR level.
func (p *Equal[T]) Air() {}

// Bounds implementation for Boundable interface.
func (p *Equal[T]) Bounds() util.Bounds {
	l := p.Lhs.Bounds()
	r := p.Rhs.Bounds()
	//
	l.Union(&r)
	//
	return l
}

// TestAt implementation for Testable interface.
func (p *Equal[T]) TestAt(k int, tr trace.Module) (bool, uint, error) {
	lhs, err1 := p.Lhs.EvalAt(k, tr)
	rhs, err2 := p.Rhs.EvalAt(k, tr)
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
func (p *Equal[T]) Lisp(module schema.Module) sexp.SExp {
	var (
		l = p.Lhs.Lisp(module)
		r = p.Rhs.Lisp(module)
	)
	//
	return sexp.NewList([]sexp.SExp{
		sexp.NewSymbol("=="), l, r})
}

// RequiredRegisters implementation for Contextual interface.
func (p *Equal[T]) RequiredRegisters() *set.SortedSet[uint] {
	panic("todo")
}

// RequiredCells implementation for Contextual interface
func (p *Equal[T]) RequiredCells(row int, tr trace.Module) *set.AnySortedSet[trace.CellRef] {
	panic("todo")
}

// Simplify this Equal as much as reasonably possible.
func (p *Equal[T]) Simplify() Equal[T] {
	panic("todo")
}
