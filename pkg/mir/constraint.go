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
package mir

import (
	"reflect"

	"github.com/consensys/gnark-crypto/ecc/bls12-377/fr"
	sc "github.com/consensys/go-corset/pkg/schema"
	"github.com/consensys/go-corset/pkg/trace"
	"github.com/consensys/go-corset/pkg/util"
	"github.com/consensys/go-corset/pkg/util/collection/set"
	"github.com/consensys/go-corset/pkg/util/source/sexp"
)

// TRUE represents a constraint which holds (i.e. evaluates to 0)
var TRUE Constraint

// FALSE represents a constraint which does not hold (i.e. evaluates to 0)
var FALSE Constraint

// Constraint represents a logical disjunction of terms which holds if at least
// one term holds.
type Constraint struct {
	// Terms here are disjuncted to formulate the final logical result.
	disjuncts []Equation
}

// AsExpr converts a constraint into an equivalent expression by taking the
// product of all disjuncted terms.
func (e Constraint) AsExpr() Expr {
	terms := make([]Term, len(e.disjuncts))
	//
	for i, t := range e.disjuncts {
		terms[i] = t.AsTerm()
	}
	//
	return termProduct(terms...)
}

// Is checks whether this constraint trivially evaluates to true or false.
func (e Constraint) Is(val bool) bool {
	for _, d := range e.disjuncts {
		if d.Is(true) {
			return true
		} else if !d.Is(false) {
			return false
		}
	}
	// If we get here, either there are no disjuncts or every disjunct was
	// provably false.
	return !val
}

// Bounds returns max shift in either the negative (left) or positive
// direction (right).
func (e Constraint) Bounds() util.Bounds {
	// Determine min/max shift
	minShift, maxShift := shiftRangeOfEquations(e.disjuncts)
	// Convert to bounds
	start := uint(-min(0, minShift))
	end := uint(max(0, maxShift))
	// Done
	return util.NewBounds(start, end)
}

// Branches returns the number of unique evaluation paths through the given
// constraint.
func (e Constraint) Branches() uint {
	return uint(len(e.disjuncts))
}

// Context determines the evaluation context (i.e. enclosing module) for this
func (e Constraint) Context(schema sc.Schema) trace.Context {
	return contextOfEquations(e.disjuncts, schema)
}

// Lisp converts this schema element into a simple S-Termession, for example
// so it can be printed.
func (e Constraint) Lisp(schema sc.Schema) sexp.SExp {
	switch len(e.disjuncts) {
	case 0:
		return sexp.NewSymbol("⊥")
	case 1:
		return lispOfEquation(e.disjuncts[0], schema)
	default:
		return lispOfEquations(schema, "∨", e.disjuncts)
	}
}

// RequiredCells returns the set of trace cells on which this term depends.
// That is, evaluating this term at the given row in the given trace will read
// these cells.
func (e Constraint) RequiredCells(row int, tr trace.Trace) *set.AnySortedSet[trace.CellRef] {
	return requiredCellsOfEquations(e.disjuncts, row, tr)
}

// RequiredColumns returns the set of columns on which this term depends.
// That is, columns whose values may be accessed when evaluating this term
// on a given trace.
func (e Constraint) RequiredColumns() *set.SortedSet[uint] {
	return requiredColumnsOfEquations(e.disjuncts)
}

// TestAt evaluates this constraint in a given tabular context and checks it
// against zero. Observe that if this expression is *undefined* within this
// context then it returns "nil".  An expression can be undefined for
// several reasons: firstly, if it accesses a row which does not exist (e.g.
// at index -1); secondly, if it accesses a column which does not exist.
func (e Constraint) TestAt(k int, tr trace.Trace) (bool, uint, error) {
	for i, t := range e.disjuncts {
		val, err := evalAtEquation(t, k, tr)
		//
		if err != nil {
			return false, uint(i), err
		} else if val.IsZero() {
			return true, uint(i), nil
		}
	}
	//
	return false, uint(len(e.disjuncts)), nil
}

func init() {
	zero := Constant{frZERO}
	eq := Equation{kind: EQUALS, lhs: &zero, rhs: &zero}
	TRUE = Constraint{[]Equation{eq}}
	FALSE = Constraint{nil}
}

// ============================================================================
// Constructors
// ============================================================================

// Disjunct creates a constraint representing the disjunction of a given set of
// constraints.
func Disjunct(constraints ...Constraint) Constraint {
	var equations []Equation
	//
	for _, c := range constraints {
		for _, d := range c.disjuncts {
			d = d.Simplify()
			//
			if d.Is(true) {
				return TRUE
			} else if !d.Is(false) {
				equations = append(equations, d)
			}
		}
	}
	//
	if len(equations) > 0 {
		return Constraint{equations}
	}
	//
	return FALSE
}

// ============================================================================
// Equation
// ============================================================================

const (
	// EQUALS indicates an equals relationship
	EQUALS uint8 = 0
	// NOT_EQUALS indicates a not-equals relationship
	NOT_EQUALS uint8 = 1
)

// Equation represents an equation between two terms (e.g. "X==Y", or "X!=Y+1",
// etc).  Equations are either equalities (or negated equalities) or
// inequalities.
type Equation struct {
	kind uint8
	lhs  Term
	rhs  Term
}

// Simplify this equation as much as reasonably possible.
func (e Equation) Simplify() Equation {
	// Apply constant propagation (whilst retaining casts)
	lhs := constantPropagationForTerm(e.lhs, true, nil)
	rhs := constantPropagationForTerm(e.rhs, true, nil)
	//
	return Equation{e.kind, lhs, rhs}
}

// Is determines whether or not this equation is known to evaluate to true or
// false.  For example, "0 == 0" evaluates to true, whilst "0 != 0" evaluates to
// false.
func (e Equation) Is(val bool) bool {
	return reflect.DeepEqual(e.lhs, e.rhs)
}

// AsTerm translates this equation into a raw term.
func (e Equation) AsTerm() Term {
	t := &Sub{[]Term{e.lhs, e.rhs}}
	//
	switch e.kind {
	case EQUALS:
		// don't do anything
	case NOT_EQUALS:
		// (1 - NORM(cb))
		normBody := &Norm{t}
		one := &Constant{fr.NewElement(1)}
		t = &Sub{[]Term{one, normBody}}
	default:
		panic("unknown equation")
	}
	//
	return t
}
