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
	"github.com/consensys/go-corset/pkg/trace"
	"github.com/consensys/go-corset/pkg/util"
	"github.com/consensys/go-corset/pkg/util/collection/set"
	"github.com/consensys/go-corset/pkg/util/source/sexp"
)

// ============================================================================
// Expressions
// ============================================================================

// Expr is an expression in the High-Level Intermediate Representation (HIR).
// Expressions at this level have a many-2-one correspondance with expressions
// in the HIR level.  For example, an "if" expression at this level will be
// "compiled out" into one or more expressions at the MIR level.
type Expr struct {
	// Termession to be evaluated, etc.
	Term Term
}

// ZERO represents the constant expression equivalent to 1.
var ZERO Expr

// ONE represents the constant expression equivalent to 1.
var ONE Expr

// VOID represents the empty expression.
var VOID Expr

// NewColumnAccess constructs an HIR expression representing the value of a given
// column on the current row.
func NewColumnAccess(column uint, shift int) Expr {
	return Expr{&ColumnAccess{column, shift}}
}

// NewConst construct an HIR expression representing a given constant.
func NewConst(val fr.Element) Expr {
	return Expr{&Constant{val}}
}

// NewLabelledConst construct an HIR expression representing a constant with a given label.
func NewLabelledConst(label string, val fr.Element) Expr {
	return Expr{&LabelledConstant{label, val}}
}

// NewConst64 construct an HIR expression representing a given constant from a
// uint64.
func NewConst64(val uint64) Expr {
	element := fr.NewElement(val)
	return Expr{&Constant{element}}
}

// Context determines the evaluation context (i.e. enclosing module) for this
func (e Expr) Context(schema sc.Schema) trace.Context {
	return contextOfTerm(e.Term, schema)
}

// Branches returns the total number of logical branches this term can take
// during evaluation.
func (e Expr) Branches() uint {
	return 0
}

// Bounds returns max shift in either the negative (left) or positive
// direction (right).
func (e Expr) Bounds() util.Bounds { return e.Term.Bounds() }

// BitWidth determines bitwidth required to hold the result of evaluating this expression.
func (e Expr) BitWidth(schema sc.Schema) uint {
	switch e := e.Term.(type) {
	case *ColumnAccess:
		bitwidth := schema.Columns().Nth(e.Column).DataType.BitWidth()
		return bitwidth
	default:
		// For now, we only supports simple column accesses.
		panic("bitwidth calculation only supported for column accesses")
	}
}

// EvalAt evaluates an expression at a given row in a trace, returning the
// resulting value.
func (e Expr) EvalAt(k int, tr trace.Trace) (fr.Element, error) {
	val, err := evalAtTerm(e.Term, k, tr)
	//
	return val, err
}

// TestAt checks whether a given expression holds (i.e. evaluates to zero) on a
// given row in the trace.
func (e Expr) TestAt(k int, tr trace.Trace) (bool, uint, error) {
	val, err := evalAtTerm(e.Term, k, tr)
	//
	return val.IsZero(), 0, err
}

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

// Exponent constructs a new expression representing the given argument
// raised to a given a given power.
func Exponent(arg Expr, pow uint64) Expr {
	return Expr{&Exp{arg.Term, pow}}
}

// CastOf constructs a new expression which has been annotated by the user to be
// within a given range.
func CastOf(arg Expr, bitwidth uint) Expr {
	return Expr{&Cast{arg.Term, bitwidth}}
}

// Disjunction constructs a logical disjunction of one or more terms.
func Disjunction(exprs ...Expr) Expr {
	terms := asTerms(exprs...)
	//
	terms = util.Flatten(terms, func(t Term) []Term {
		if d, ok := t.(*Connective); ok && d.Sign {
			return d.Args
		}
		//
		return nil
	})
	//
	return Expr{&Connective{true, terms}}
}

// Conjunction constructs a logical conjunction of one or more terms.
func Conjunction(exprs ...Expr) Expr {
	terms := asTerms(exprs...)
	//
	terms = util.Flatten(terms, func(t Term) []Term {
		if d, ok := t.(*Connective); ok && !d.Sign {
			return d.Args
		}
		//
		return nil
	})
	//
	return Expr{&Connective{false, terms}}
}

// Negation construct the logical negation of a given term
func Negation(expr Expr) Expr {
	return Expr{&Not{expr.Term}}
}

// Equals constructs an equality between two expressions.
func Equals(lhs Expr, rhs Expr) Expr {
	return Expr{&Equation{EQUALS, lhs.Term, rhs.Term}}
}

// NotEquals constructs a non-equality between two expressions.
func NotEquals(lhs Expr, rhs Expr) Expr {
	return Expr{&Equation{NOT_EQUALS, lhs.Term, rhs.Term}}
}

// LessThan constructs an inequality between two expressions.
func LessThan(lhs Expr, rhs Expr) Expr {
	return Expr{&Equation{LESS_THAN, lhs.Term, rhs.Term}}
}

// LessThanOrEquals constructs an inequality between two expressions.
func LessThanOrEquals(lhs Expr, rhs Expr) Expr {
	return Expr{&Equation{LESS_THAN_EQUALS, lhs.Term, rhs.Term}}
}

// GreaterThan constructs an inequality between two expressions.
func GreaterThan(lhs Expr, rhs Expr) Expr {
	return Expr{&Equation{GREATER_THAN, lhs.Term, rhs.Term}}
}

// GreaterThanOrEquals constructs an inequality between two expressions.
func GreaterThanOrEquals(lhs Expr, rhs Expr) Expr {
	return Expr{&Equation{GREATER_THAN_EQUALS, lhs.Term, rhs.Term}}
}

// If constructs a new conditional branch, where the true branch is taken when
// the condition evaluates to zero.
func If(condition Expr, trueBranch Expr) Expr {
	return Expr{&IfZero{condition.Term, trueBranch.Term, nil}}
}

// IfElse constructs a new conditional branch, where either the true branch or
// the false branch can (optionally) be VOID (but both cannot).  Note, the true
// branch is taken when the condition evaluates to zero.
func IfElse(condition Expr, trueBranch Expr, falseBranch Expr) Expr {
	return Expr{&IfZero{condition.Term, trueBranch.Term, falseBranch.Term}}
}

// Normalise normalises the result of evaluating a given expression to be
// either 0 (if its value was 0) or 1 (otherwise).
func Normalise(arg Expr) Expr {
	return Expr{&Norm{arg.Term}}
}

// ListOf constructs a list of 0 or more values
func ListOf(exprs ...Expr) Expr {
	terms := asTerms(exprs...)
	return Expr{&List{terms}}
}

// Sum zero or more expressions together.
func Sum(exprs ...Expr) Expr {
	terms := asTerms(exprs...)
	// flatten any nested sums
	terms = util.Flatten(terms, func(t Term) []Term {
		if t, ok := t.(*Add); ok {
			return t.Args
		}
		//
		return nil
	})
	// Remove any zeros
	terms = util.RemoveMatching(terms, isZero)
	// Final optimisation
	switch len(terms) {
	case 0:
		return NewConst64(0)
	case 1:
		return Expr{terms[0]}
	default:
		return Expr{&Add{terms}}
	}
}

// Product returns the product of zero or more multiplications.
func Product(exprs ...Expr) Expr {
	terms := asTerms(exprs...)
	// flatten any nested products
	terms = util.Flatten(terms, func(t Term) []Term {
		if t, ok := t.(*Mul); ok {
			return t.Args
		}
		//
		return nil
	})
	// Remove all multiplications by one
	terms = util.RemoveMatching(terms, isOne)
	// Check for zero
	if util.ContainsMatching(terms, isZero) {
		return ZERO
	}
	// Final optimisation
	switch len(terms) {
	case 0:
		return NewConst64(1)
	case 1:
		return Expr{terms[0]}
	default:
		return Expr{&Mul{terms}}
	}
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

func isOne(term Term) bool {
	if t, ok := term.(*Constant); ok {
		return t.Value.IsOne()
	}
	//
	return false
}

func isZero(term Term) bool {
	if t, ok := term.(*Constant); ok {
		return t.Value.IsZero()
	}
	//
	return false
}

func init() {
	ONE = NewConst64(1)
	ZERO = NewConst64(0)
}
