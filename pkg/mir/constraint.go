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

// FALSE represents a constraint which does not hold (i.e. evaluates to 0)
var FALSE Constraint

// Constraint represents a formula in conjunctive normal form.
type Constraint struct {
	// Terms here are disjuncted to formulate the final logical result.
	conjuncts []Disjunction
}

// NewConstraint constructs a new atomic constraint representing a given
// equation.
func NewConstraint(equation Equation) Constraint {
	disjunct := Disjunction{[]Equation{equation}}
	return Constraint{[]Disjunction{disjunct}}
}

// Simplify a given constraint.  Observe this can result in the given constraint
// being reduced to TRUE or FALSE.
func (e Constraint) Simplify() Constraint {
	var ne Constraint
	//
	for _, c := range e.conjuncts {
		d, tt := c.Simplify()
		//
		if !tt {
			ne.conjuncts = append(ne.conjuncts, d)
		}
	}
	//
	if ne.Is(true) {
		return TRUE
	} else if ne.Is(false) {
		return FALSE
	}
	//
	return ne
}

// Is checks whether this constraint trivially evaluates to true or false.
func (e Constraint) Is(val bool) bool {
	if len(e.conjuncts) == 0 {
		// true
		return val
	} else if len(e.conjuncts) == 1 && len(e.conjuncts[0].atoms) == 0 {
		// false
		return !val
	}
	// unknown
	return false
}

// Bounds returns max shift in either the negative (left) or positive
// direction (right).
func (e Constraint) Bounds() util.Bounds {
	// Determine min/max shift
	minShift, maxShift := shiftRangeOfDisjuncts(e.conjuncts...)
	// Convert to bounds
	start := uint(-min(0, minShift))
	end := uint(max(0, maxShift))
	// Done
	return util.NewBounds(start, end)
}

// Branches returns the number of unique evaluation paths through the given
// constraint.
func (e Constraint) Branches() uint {
	n := uint(1)

	for _, conjunct := range e.conjuncts {
		n *= uint(len(conjunct.atoms))
	}
	//
	return n
}

// Context determines the evaluation context (i.e. enclosing module) for this
func (e Constraint) Context(schema sc.Schema) trace.Context {
	return contextOfConjunction(e, schema)
}

// Lisp converts this schema element into a simple S-Termession, for example
// so it can be printed.
func (e Constraint) Lisp(schema sc.Schema) sexp.SExp {
	return lispOfConjunction(schema, e)
}

// RequiredCells returns the set of trace cells on which this term depends.
// That is, evaluating this term at the given row in the given trace will read
// these cells.
func (e Constraint) RequiredCells(row int, tr trace.Trace) *set.AnySortedSet[trace.CellRef] {
	return requiredCellsOfConjunction(e, row, tr)
}

// RequiredColumns returns the set of columns on which this term depends.
// That is, columns whose values may be accessed when evaluating this term
// on a given trace.
func (e Constraint) RequiredColumns() *set.SortedSet[uint] {
	return requiredColumnsOfConjunction(e)
}

// TestAt evaluates this constraint in a given tabular context and checks it
// against zero. Observe that if this expression is *undefined* within this
// context then it returns "nil".  An expression can be undefined for
// several reasons: firstly, if it accesses a row which does not exist (e.g.
// at index -1); secondly, if it accesses a column which does not exist.
func (e Constraint) TestAt(k int, tr trace.Trace) (bool, uint, error) {
	val, err := evalAtConstraint(e, k, tr)
	//
	if err != nil {
		return false, 0, err
	} else if val.IsZero() {
		return true, 0, nil
	}
	//
	return false, 0, nil
}

func init() {
	// True is the empty conjunct
	TRUE = Constraint{nil}
	// False is the empty disjunct
	FALSE = Constraint{[]Disjunction{Disjunction{nil}}}
}

// ============================================================================
// Constructors
// ============================================================================

// Conjunct builds the logical conjunction (i.e. and) for a given set of constraints.
func Conjunct(constraints ...Constraint) Constraint {
	var disjuncts []Disjunction
	//
	for _, c := range constraints {
		if c.Is(false) {
			return FALSE
		} else if !c.Is(true) {
			//
			disjuncts = append(disjuncts, c.conjuncts...)
		}
	}
	//
	return Constraint{disjuncts}
}

// Disjunct creates a constraint representing the disjunction of a given set of
// constraints.
func Disjunct(constraints ...Constraint) Constraint {
	switch len(constraints) {
	case 0:
		return FALSE
	case 1:
		return constraints[0]
	default:
		lhs := constraints[0]
		// Recurse
		rhs := Disjunct(constraints[1:]...)
		// Base case
		return disjunct(lhs, rhs)
	}
}

// Negate constructs the logical negation of the given constraint.
func Negate(constraint Constraint) Constraint {
	if constraint.Is(true) {
		return FALSE
	} else if constraint.Is(false) {
		return TRUE
	}
	//
	conjuncts := make([]Constraint, len(constraint.conjuncts))
	//
	for i, disjunct := range constraint.conjuncts {
		conjuncts[i] = disjunct.Negate()
	}
	//
	return Disjunct(conjuncts...)
}

// Equals constructs an equation representing the equality of two expressions.
func Equals(lhs Expr, rhs Expr) Constraint {
	eq := Equation{EQUALS, lhs.term, rhs.term}
	dis := Disjunction{[]Equation{eq}}

	return Constraint{[]Disjunction{dis}}
}

// NotEquals constructs an equation representing the non-equality of two
// expressions.
func NotEquals(lhs Expr, rhs Expr) Constraint {
	eq := Equation{NOT_EQUALS, lhs.term, rhs.term}
	dis := Disjunction{[]Equation{eq}}

	return Constraint{[]Disjunction{dis}}
}

// LessThan constructs an equation representing the inequality of two
// expressions.
func LessThan(lhs Expr, rhs Expr) Constraint {
	eq := Equation{LESS_THAN, lhs.term, rhs.term}
	dis := Disjunction{[]Equation{eq}}

	return Constraint{[]Disjunction{dis}}
}

// LessThanOrEquals constructs an equation representing the inequality of two
// expressions.
func LessThanOrEquals(lhs Expr, rhs Expr) Constraint {
	eq := Equation{LESS_THAN_EQUALS, lhs.term, rhs.term}
	dis := Disjunction{[]Equation{eq}}

	return Constraint{[]Disjunction{dis}}
}

// GreaterThanOrEquals constructs an equation representing the inequality of two
// expressions.
func GreaterThanOrEquals(lhs Expr, rhs Expr) Constraint {
	eq := Equation{GREATER_THAN_EQUALS, lhs.term, rhs.term}
	dis := Disjunction{[]Equation{eq}}

	return Constraint{[]Disjunction{dis}}
}

// GreaterThan constructs an equation representing the inequality of two
// expressions.
func GreaterThan(lhs Expr, rhs Expr) Constraint {
	eq := Equation{GREATER_THAN, lhs.term, rhs.term}
	dis := Disjunction{[]Equation{eq}}

	return Constraint{[]Disjunction{dis}}
}

// ============================================================================
// Disjunction
// ============================================================================

// Disjunction represents a logical disjunction of equations.
type Disjunction struct {
	atoms []Equation
}

// Negate a given disjunction
func (e *Disjunction) Negate() Constraint {
	conjuncts := make([]Constraint, len(e.atoms))
	//
	for i, atom := range e.atoms {
		conjuncts[i] = NewConstraint(atom.Negate())
	}
	//
	return Conjunct(conjuncts...)
}

// Simplify a disjunction, either returning the simplified result or an
// indicator that the disjunction is trivially true.  Observe that, if the
// disjunction is trivially false, then it returns an empty disjunction.  This
// is consistent with the representation used in Constraint.
func (e *Disjunction) Simplify() (Disjunction, bool) {
	var natoms []Equation
	//
	for _, atom := range e.atoms {
		natom := atom.Simplify()
		if natom.Is(true) {
			// trivially true
			return Disjunction{}, true
		} else if !natom.Is(false) {
			natoms = append(natoms, natom)
		}
	}
	//
	return Disjunction{natoms}, false
}

// ============================================================================
// Equation
// ============================================================================

const (
	// EQUALS indicates an equals (==) relationship
	EQUALS uint8 = 0
	// NOT_EQUALS indicates a not-equals (!=) relationship
	NOT_EQUALS uint8 = 1
	// LESS_THAN indicates a less-than (<) relationship
	LESS_THAN uint8 = 2
	// LESS_THAN_EQUALS indicates a less-than-or-equals (<=) relationship
	LESS_THAN_EQUALS uint8 = 3
	// GREATER_THAN indicates a greater-than (>) relationship
	GREATER_THAN uint8 = 4
	// GREATER_THAN_EQUALS indicates a greater-than-or-equals (>=) relationship
	GREATER_THAN_EQUALS uint8 = 5
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

// Negate a given equation
func (e Equation) Negate() Equation {
	var kind uint8
	//
	switch e.kind {
	case EQUALS:
		kind = NOT_EQUALS
	case NOT_EQUALS:
		kind = EQUALS
	case LESS_THAN:
		kind = GREATER_THAN_EQUALS
	case LESS_THAN_EQUALS:
		kind = GREATER_THAN
	case GREATER_THAN_EQUALS:
		kind = LESS_THAN
	case GREATER_THAN:
		kind = LESS_THAN_EQUALS
	}
	//
	return Equation{kind, e.lhs, e.rhs}
}

// Is determines whether or not this equation is known to evaluate to true or
// false.  For example, "0 == 0" evaluates to true, whilst "0 != 0" evaluates to
// false.
func (e Equation) Is(val bool) bool {
	// Attempt to disprove non-equality
	lc, l_ok := e.lhs.(*Constant)
	rc, r_ok := e.rhs.(*Constant)
	//
	if l_ok && r_ok {
		var (
			cmp  = lc.Value.Cmp(&rc.Value)
			sign bool
		)
		//
		switch e.kind {
		case EQUALS:
			sign = (cmp == 0)
		case NOT_EQUALS:
			sign = (cmp != 0)
		case LESS_THAN:
			sign = (cmp < 0)
		case LESS_THAN_EQUALS:
			sign = (cmp <= 0)
		case GREATER_THAN:
			sign = (cmp > 0)
		case GREATER_THAN_EQUALS:
			sign = (cmp >= 0)
		default:
			panic("unreachable")
		}
		//
		return val == sign
	}
	// Give up
	return false
}

// Lisp returns a lisp representation of this equation, which is useful for
// debugging.
func (e Equation) Lisp() sexp.SExp {
	var (
		symbol string
		l      = lispOfTerm(e.lhs, nil)
		r      = lispOfTerm(e.rhs, nil)
	)
	//
	switch e.kind {
	case EQUALS:
		symbol = "=="
	case NOT_EQUALS:
		symbol = "!="
	case LESS_THAN:
		symbol = "<"
	case LESS_THAN_EQUALS:
		symbol = "<="
	case GREATER_THAN:
		symbol = ">"
	case GREATER_THAN_EQUALS:
		symbol = ">="
	default:
		panic("unreachable")
	}
	//
	return sexp.NewList([]sexp.SExp{
		sexp.NewSymbol(symbol), l, r})
}

// ============================================================================
// Helpers
// ============================================================================

// Construct the disjunction of two constraints.
//
// nolint
func disjunct(lhs Constraint, rhs Constraint) Constraint {
	if len(lhs.conjuncts) == 0 {
		return rhs
	} else if len(rhs.conjuncts) == 0 {
		return lhs
	}
	//
	var disjuncts []Disjunction
	//
	for _, l_d := range lhs.conjuncts {
		var l_atoms []Equation
		// left atoms
		for _, l_atom := range l_d.atoms {
			if l_atom.Is(true) {
				return TRUE
			} else if !l_atom.Is(false) {
				l_atoms = append(l_atoms, l_atom)
			}
		}
		//
		for _, r_d := range rhs.conjuncts {
			var atoms []Equation = make([]Equation, len(l_atoms))
			//
			copy(atoms, l_atoms)
			// Right atoms
			for _, r_atom := range r_d.atoms {
				if r_atom.Is(true) {
					return TRUE
				} else if !r_atom.Is(false) {
					atoms = append(atoms, r_atom)
				}
			}
			// Combine them all
			disjuncts = append(disjuncts, Disjunction{atoms})
		}
	}
	//
	return Constraint{disjuncts}
}
