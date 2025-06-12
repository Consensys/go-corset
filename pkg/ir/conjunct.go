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

// Conjunct erpresents the logical AND of zero or more terms.  Observe that if
// there are no terms, then this is equivalent to logical truth.
type Conjunct[T LogicalTerm[T]] struct {
	// Terms here are disjuncted to formulate the final logical result.
	Args []T
}

// True constructs a logical truth
func True[T LogicalTerm[T]]() T {
	return Conjunction[T]()
}

// IsTrue checks whether a given term corresponds to logical truth which, in
// this system, corresponds to an empty conjunct.
func IsTrue[T LogicalTerm[T]](term T) bool {
	var t LogicalTerm[T] = term
	//
	if t, ok := t.(*Conjunct[T]); ok {
		return len(t.Args) == 0
	}
	//
	return false
}

// Conjunction builds the logical conjunction (i.e. and) for a given set of constraints.
func Conjunction[T LogicalTerm[T]](terms ...T) T {
	var term LogicalTerm[T] = &Conjunct[T]{terms}
	return term.(T)
}

// ApplyShift implementation for LogicalTerm interface.
func (p *Conjunct[T]) ApplyShift(shift int) T {
	return Conjunction(applyShiftOfTerms(p.Args, shift)...)
}

// Bounds implementation for Boundable interface.
func (p *Conjunct[T]) Bounds() util.Bounds {
	return util.BoundsForArray(p.Args)
}

// ShiftRange implementation for LogicalTerm interface.
func (p *Conjunct[T]) ShiftRange() (int, int) {
	return shiftRangeOfTerms(p.Args...)
}

// TestAt implementation for Testable interface.
func (p *Conjunct[T]) TestAt(k int, tr trace.Module) (bool, uint, error) {
	//
	for _, disjunct := range p.Args {
		val, _, err := disjunct.TestAt(k, tr)
		//
		if err != nil {
			return val, 0, err
		} else if !val {
			// Failure
			return val, 0, nil
		}
	}
	// Success
	return true, 0, nil
}

// Lisp returns a lisp representation of this equation, which is useful for
// debugging.
func (p *Conjunct[T]) Lisp(module schema.Module) sexp.SExp {
	if len(p.Args) == 0 {
		return sexp.NewSymbol("⊤")
	}

	return lispOfLogicalTerms(module, "∧", p.Args)
}

// RequiredRegisters implementation for Contextual interface.
func (p *Conjunct[T]) RequiredRegisters() *set.SortedSet[uint] {
	return requiredRegistersOfTerms(p.Args)
}

// RequiredCells implementation for Contextual interface
func (p *Conjunct[T]) RequiredCells(row int, tr trace.Module) *set.AnySortedSet[trace.CellRef] {
	return requiredCellsOfTerms(p.Args, row, tr)
}

// Simplify this term as much as reasonably possible.
func (p *Conjunct[T]) Simplify(casts bool) T {
	// Simplify terms
	terms := simplifyLogicalTerms(p.Args, casts)
	// Flatten any nested conjuncts
	terms = util.Flatten(terms, flatternConjunct)
	// False if contains false
	if util.ContainsMatching(terms, IsFalse) {
		return False[T]()
	}
	// Remove true values
	terms = util.RemoveMatching(terms, IsTrue)
	// Final checks
	switch len(terms) {
	case 0:
		return True[T]()
	case 1:
		return terms[0]
	default:
		return Conjunction(terms...)
	}
}

func flatternConjunct[T LogicalTerm[T]](term T) []T {
	var e LogicalTerm[T] = term
	if t, ok := e.(*Conjunct[T]); ok {
		return t.Args
	}
	//
	return nil
}
