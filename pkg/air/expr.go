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
package air

import (
	"github.com/consensys/gnark-crypto/ecc/bls12-377/fr"
	sc "github.com/consensys/go-corset/pkg/schema"
	"github.com/consensys/go-corset/pkg/trace"
	"github.com/consensys/go-corset/pkg/util"
	"github.com/consensys/go-corset/pkg/util/collection/set"
	"github.com/consensys/go-corset/pkg/util/sexp"
)

// Expr represents an expression in the Arithmetic Intermediate Representation
// (AIR). Any expression in this form can be lowered into a polynomial.
// Expressions at this level are split into those which can be arithmetised and
// those which cannot.  The latter represent expressions which cannot be
// expressed within a polynomial but can be computed externally (e.g. during
// trace expansion).
type Expr struct {
	// Termession to be evaluated, etc.
	Term Term
}

var _ sc.Evaluable = Expr{}

// NewColumnAccess constructs an AIR expression representing the value of a given
// column on the current row.
func NewColumnAccess(column uint, shift int) Expr {
	return Expr{&ColumnAccess{column, shift}}
}

// NewConst construct an AIR expression representing a given constant.
func NewConst(val fr.Element) Expr {
	return Expr{&Constant{val}}
}

// NewConst64 construct an AIR expression representing a given constant from a
// uint64.
func NewConst64(val uint64) Expr {
	element := fr.NewElement(val)
	return Expr{&Constant{element}}
}

// AsConstant determines whether or not this is a constant expression.  If
// so, the constant is returned; otherwise, nil is returned.  NOTE: this
// does not perform any form of simplification to determine this.
func (e Expr) AsConstant() *fr.Element {
	return constantOfTerm(e.Term)
}

// Context determines the evaluation context (i.e. enclosing module) for this
func (e Expr) Context(schema sc.Schema) trace.Context {
	return contextOfTerm(e.Term, schema)
}

// Bounds returns max shift in either the negative (left) or positive
// direction (right).
func (e Expr) Bounds() util.Bounds { return e.Term.Bounds() }

// Lisp converts this schema element into a simple S-Termession, for example
// so it can be printed.
func (e Expr) Lisp(schema sc.Schema) sexp.SExp {
	return lispOfTerm(e.Term, schema)
}

// RequiredColumns returns the set of columns on which this term depends.
// That is, columns whose values may be accessed when evaluating this term
// on a given trace.
func (e Expr) RequiredColumns() *set.SortedSet[uint] {
	return requiredColumnsOfTerm(e.Term)
}

// RequiredCells returns the set of trace cells on which this term depends.
// That is, evaluating this term at the given row in the given trace will read
// these cells.
func (e Expr) RequiredCells(row int, tr trace.Trace) *set.AnySortedSet[trace.CellRef] {
	return requiredCellsOfTerm(e.Term, row, tr)
}

// EvalAt evaluates a column access at a given row in a trace, which returns the
// value at that row of the column in question or nil is that row is
// out-of-bounds.
func (e Expr) EvalAt(k int, tr trace.Trace) (fr.Element, error) {
	val, _ := evalAtTerm[sc.NoMetric](e.Term, k, tr)
	//
	return val, nil
}

// TestAt evaluates this expression in a given tabular context and checks it
// against zero. Observe that if this expression is *undefined* within this
// context then it returns "nil".  An expression can be undefined for
// several reasons: firstly, if it accesses a row which does not exist (e.g.
// at index -1); secondly, if it accesses a column which does not exist.
func (e Expr) TestAt(k int, tr trace.Trace) (bool, sc.BranchMetric, error) {
	val, path := evalAtTerm[sc.BranchMetric](e.Term, k, tr)
	//
	return val.IsZero(), path, nil
}

// Branches returns the number of unique evaluation paths through the given
// constraint.
func (e Expr) Branches() uint {
	return pathsOfTerm(e.Term)
}

// Add two expressions together.
func (e Expr) Add(arg Expr) Expr {
	return Expr{&Add{Args: []Term{e.Term, arg.Term}}}
}

// Sub subtracts the argument from this expression.
func (e Expr) Sub(arg Expr) Expr {
	return Expr{&Sub{Args: []Term{e.Term, arg.Term}}}
}

// Mul multiplies this expression with the argument
func (e Expr) Mul(arg Expr) Expr {
	return Expr{&Mul{Args: []Term{e.Term, arg.Term}}}
}

// Equate equates this expression with the argument.
func (e Expr) Equate(arg Expr) Expr {
	return Expr{&Sub{Args: []Term{e.Term, arg.Term}}}
}

// Sum zero or more expressions together.
func Sum(exprs ...Expr) Expr {
	if len(exprs) == 0 {
		return NewConst64(0)
	}
	//
	return Expr{&Add{asTerms(exprs...)}}
}

// Product returns the product of zero or more multiplications.
func Product(exprs ...Expr) Expr {
	if len(exprs) == 0 {
		return NewConst64(1)
	}
	//
	return Expr{&Mul{asTerms(exprs...)}}
}

// Subtract returns the subtraction of the subsequent expressions from the
// first.
func Subtract(exprs ...Expr) Expr {
	if len(exprs) == 0 {
		return NewConst64(0)
	}
	//
	return Expr{&Sub{asTerms(exprs...)}}
}

func asTerms(exprs ...Expr) []Term {
	terms := make([]Term, len(exprs))
	//
	for i, e := range exprs {
		terms[i] = e.Term
	}
	//
	return terms
}
