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
	sc "github.com/consensys/go-corset/pkg/schema"
	"github.com/consensys/go-corset/pkg/trace"
	"github.com/consensys/go-corset/pkg/util"
	"github.com/consensys/go-corset/pkg/util/collection/set"
	"github.com/consensys/go-corset/pkg/util/source/sexp"
)

// TRUE represents a constraint which holds (i.e. evaluates to 0)
var TRUE Constraint

// Constraint represents a logical disjunction of terms which holds if at least
// one term holds.
type Constraint struct {
	// Terms here are disjuncted to formulate the final logical result.
	terms []Term
}

// AsExpr converts a constraint into an equivalent expression by taking the
// product of all disjuncted terms.
func (e Constraint) AsExpr() Expr {
	return termProduct(e.terms...)
}

// Bounds returns max shift in either the negative (left) or positive
// direction (right).
func (e Constraint) Bounds() util.Bounds {
	// Determine min/max shift
	minShift, maxShift := shiftRangeOfTerms(e.terms)
	// Convert to bounds
	start := uint(-min(0, minShift))
	end := uint(max(0, maxShift))
	// Done
	return util.NewBounds(start, end)
}

// Branches returns the number of unique evaluation paths through the given
// constraint.
func (e Constraint) Branches() uint {
	return uint(len(e.terms))
}

// Context determines the evaluation context (i.e. enclosing module) for this
func (e Constraint) Context(schema sc.Schema) trace.Context {
	return contextOfTerms(e.terms, schema)
}

// Lisp converts this schema element into a simple S-Termession, for example
// so it can be printed.
func (e Constraint) Lisp(schema sc.Schema) sexp.SExp {
	switch len(e.terms) {
	case 0:
		return sexp.NewSymbol("⊤")
	case 1:
		return lispOfTerm(e.terms[0], schema)
	default:
		return nary2Lisp(schema, "∨", e.terms)
	}
}

// RequiredCells returns the set of trace cells on which this term depends.
// That is, evaluating this term at the given row in the given trace will read
// these cells.
func (e Constraint) RequiredCells(row int, tr trace.Trace) *set.AnySortedSet[trace.CellRef] {
	return requiredCellsOfTerms(e.terms, row, tr)
}

// RequiredColumns returns the set of columns on which this term depends.
// That is, columns whose values may be accessed when evaluating this term
// on a given trace.
func (e Constraint) RequiredColumns() *set.SortedSet[uint] {
	return requiredColumnsOfTerms(e.terms)
}

// TestAt evaluates this constraint in a given tabular context and checks it
// against zero. Observe that if this expression is *undefined* within this
// context then it returns "nil".  An expression can be undefined for
// several reasons: firstly, if it accesses a row which does not exist (e.g.
// at index -1); secondly, if it accesses a column which does not exist.
func (e Constraint) TestAt(k int, tr trace.Trace) (bool, uint, error) {
	for i, t := range e.terms {
		val, err := evalAtTerm(t, k, tr)
		//
		if err != nil {
			return false, uint(i), err
		} else if val.IsZero() {
			return true, uint(i), nil
		}
	}
	//
	return false, uint(len(e.terms)), nil
}

func init() {
	TRUE = Constraint{nil}
}

// ============================================================================
// Constructors
// ============================================================================

// Disjunct creates a constraint representing the disjunction of a given set of
// constraints.
func Disjunct(constraints ...Constraint) Constraint {
	var nterms []Term
	//
	for _, c := range constraints {
		nterms = append(nterms, c.terms...)
	}
	// TODO: opportunity here for simplification?
	//
	return Constraint{nterms}
}
