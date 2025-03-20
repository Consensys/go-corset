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
package hir

import (
	"github.com/consensys/gnark-crypto/ecc/bls12-377/fr"
	sc "github.com/consensys/go-corset/pkg/schema"
	tr "github.com/consensys/go-corset/pkg/trace"
	"github.com/consensys/go-corset/pkg/util"
	"github.com/consensys/go-corset/pkg/util/collection/set"
	"github.com/consensys/go-corset/pkg/util/source/sexp"
)

// ============================================================================
// ZeroArrayTest
// ============================================================================

// ZeroArrayTest is a wrapper which converts an array of expressions into a
// Testable constraint.  Specifically, by checking whether or not the each
// expression vanishes (i.e. evaluates to zero).
type ZeroArrayTest struct {
	Expr Expr
}

// TestAt determines whether or not every element from a given array of
// expressions evaluates to zero. Observe that any expressions which are
// undefined are assumed to hold.
func (p ZeroArrayTest) TestAt(row int, trace tr.Trace) (bool, uint, error) {
	// Evalues expression yielding zero or more values.
	vals, err := evalAtTerm(p.Expr.Term, row, trace)
	//
	if err != nil {
		return false, 0, err
	}
	// Check each value in turn against zero.
	for _, val := range vals {
		if !val.IsZero() {
			// This expression does not evaluate to zero, hence failure.
			return false, 0, nil
		}
	}
	// Success
	return true, 0, nil
}

// Branches returns the number of unique evaluation paths through the given
// constraint.
func (p ZeroArrayTest) Branches() uint {
	// FIXME
	return 1
}

// Bounds determines the bounds for this zero test.
func (p ZeroArrayTest) Bounds() util.Bounds {
	return p.Expr.Bounds()
}

// Context determines the evaluation context (i.e. enclosing module) for this
// expression.
func (p ZeroArrayTest) Context(schema sc.Schema) tr.Context {
	return p.Expr.Context(schema)
}

// RequiredColumns returns the set of columns on which this term depends.
// That is, columns whose values may be accessed when evaluating this term
// on a given trace.
func (p ZeroArrayTest) RequiredColumns() *set.SortedSet[uint] {
	return p.Expr.RequiredColumns()
}

// RequiredCells returns the set of trace cells on which evaluation of this
// constraint element depends.
func (p ZeroArrayTest) RequiredCells(row int, trace tr.Trace) *set.AnySortedSet[tr.CellRef] {
	return p.Expr.RequiredCells(row, trace)
}

// Lisp converts this schema element into a simple S-Expression, for example
// so it can be printed.
func (p ZeroArrayTest) Lisp(schema sc.Schema) sexp.SExp {
	return p.Expr.Lisp(schema.(*Schema))
}

// ============================================================================
// UnitExpr
// ============================================================================

// UnitExpr is an adaptor for a general expression which can be used in
// situations where an Evaluable expression is required.  This performs a
// similar function to the ZeroArrayTest, but actually produces a value.  A
// strict requirement is placed that the given expression always returns (via
// EvalAll) exactly one result.  This means the presence of certain constructs,
// such as lists and if conditions can result in Eval causing a panic.
type UnitExpr struct {
	//
	Expr Expr
}

// NewUnitExpr constructs a unit wrapper around an HIR expression.  In essence,
// this introduces a runtime check that the given expression only every reduces
// to a single value.  Evaluation of this expression will panic if that
// condition does not hold.  The intention is that this error is checked for
// upstream (e.g. as part of the compiler front end).
func NewUnitExpr(expr Expr) UnitExpr {
	return UnitExpr{expr}
}

// EvalAt evaluates a column access at a given row in a trace, which returns the
// value at that row of the column in question or nil is that row is
// out-of-bounds.
func (e UnitExpr) EvalAt(k int, trace tr.Trace) (fr.Element, error) {
	vals, err := evalAtTerm(e.Expr.Term, k, trace)
	// Check we got exactly one thing
	if len(vals) == 1 {
		return vals[0], err
	}
	// Fail
	panic("invalid unitary expression")
}

// Bounds returns max shift in either the negative (left) or positive
// direction (right).
func (e UnitExpr) Bounds() util.Bounds {
	return e.Expr.Bounds()
}

// BitWidth determines bitwidth required to hold the result of evaluating this expression.
func (e UnitExpr) BitWidth(schema sc.Schema) uint {
	return e.Expr.BitWidth(schema)[0]
}

// Context determines the evaluation context (i.e. enclosing module) for this
// expression.
func (e UnitExpr) Context(schema sc.Schema) tr.Context {
	return e.Expr.Context(schema)
}

// Branches returns the total number of logical branches this term can take
// during evaluation.
func (e UnitExpr) Branches() uint {
	return e.Expr.Branches()
}

// RequiredColumns returns the set of columns on which this term depends.
// That is, columns whose values may be accessed when evaluating this term
// on a given trace.
func (e UnitExpr) RequiredColumns() *set.SortedSet[uint] {
	return e.Expr.RequiredColumns()
}

// RequiredCells returns the set of trace cells on which this term depends.
// In this case, that is the empty set.
func (e UnitExpr) RequiredCells(row int, trace tr.Trace) *set.AnySortedSet[tr.CellRef] {
	return e.Expr.RequiredCells(row, trace)
}

// Lisp converts this schema element into a simple S-Expression, for example
// so it can be printed.
func (e UnitExpr) Lisp(schema sc.Schema) sexp.SExp {
	return e.Expr.Lisp(schema.(*Schema))
}

// ============================================================================
// MaxExpr
// ============================================================================

// MaxExpr is an adaptor for a general expression which can be used in
// situations where an Evaluable expression is required.  This performs a
// similar function to the ZeroArrayTest, but actually produces a value.
// Specifically, the value produced is always the maximum of all values
// produced.  This is only useful in specific situations (e.g. checking range
// constraints).
type MaxExpr struct {
	//
	Expr Expr
}

// NewMaxExpr constructs a unit wrapper around an HIR expression.  In essence,
// this introduces a runtime check that the given expression only every reduces
// to a single value.  Evaluation of this expression will panic if that
// condition does not hold.  The intention is that this error is checked for
// upstream (e.g. as part of the compiler front end).
func NewMaxExpr(expr Expr) MaxExpr {
	return MaxExpr{expr}
}

// EvalAt evaluates a column access at a given row in a trace, which returns the
// value at that row of the column in question or nil is that row is
// out-of-bounds.
func (e MaxExpr) EvalAt(k int, trace tr.Trace) (fr.Element, error) {
	vals, err := evalAtTerm(e.Expr.Term, k, trace)
	//
	max := fr.NewElement(0)
	//
	for _, v := range vals {
		if max.Cmp(&v) < 0 {
			max = v
		}
	}
	//
	return max, err
}

// Bounds returns max shift in either the negative (left) or positive
// direction (right).
func (e MaxExpr) Bounds() util.Bounds {
	return e.Expr.Bounds()
}

// Branches returns the total number of logical branches this term can take
// during evaluation.
func (e MaxExpr) Branches() uint {
	return e.Expr.Branches()
}

// Context determines the evaluation context (i.e. enclosing module) for this
// expression.
func (e MaxExpr) Context(schema sc.Schema) tr.Context {
	return e.Expr.Context(schema)
}

// RequiredColumns returns the set of columns on which this term depends.
// That is, columns whose values may be accessed when evaluating this term
// on a given trace.
func (e MaxExpr) RequiredColumns() *set.SortedSet[uint] {
	return e.Expr.RequiredColumns()
}

// RequiredCells returns the set of trace cells on which this term depends.
// In this case, that is the empty set.
func (e MaxExpr) RequiredCells(row int, trace tr.Trace) *set.AnySortedSet[tr.CellRef] {
	return e.Expr.RequiredCells(row, trace)
}

// Lisp converts this schema element into a simple S-Expression, for example
// so it can be printed.
func (e MaxExpr) Lisp(schema sc.Schema) sexp.SExp {
	return e.Expr.Lisp(schema.(*Schema))
}
