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
	"github.com/consensys/go-corset/pkg/util/collection/array"
	"github.com/consensys/go-corset/pkg/util/collection/set"
	"github.com/consensys/go-corset/pkg/util/source/sexp"
)

// Disjunct erpresents the logical OR of zero or more terms.  Observe that if
// there are no terms, then this is equivalent to logical falsehood.
type Disjunct[T LogicalTerm[T]] struct {
	Args []T
}

// False constructs a logical falsehood
func False[T LogicalTerm[T]]() T {
	return Disjunction[T]()
}

// IsFalse Check whether a given term corresponds to logical falsehood which, in
// this system, corresponds to an empty disjunct.
func IsFalse[T LogicalTerm[T]](term T) bool {
	var t LogicalTerm[T] = term
	//
	if t, ok := t.(*Disjunct[T]); ok {
		return len(t.Args) == 0
	}
	//
	return false
}

// Disjunction creates a constraint representing the disjunction of a given set of
// constraints.
func Disjunction[T LogicalTerm[T]](terms ...T) T {
	var term LogicalTerm[T] = &Disjunct[T]{terms}
	return term.(T)
}

// ApplyShift implementation for LogicalTerm interface.
func (p *Disjunct[T]) ApplyShift(shift int) T {
	return Disjunction(applyShiftOfTerms(p.Args, shift)...)
}

// ShiftRange implementation for LogicalTerm interface.
func (p *Disjunct[T]) ShiftRange() (int, int) {
	return shiftRangeOfTerms(p.Args...)
}

// Bounds implementation for Boundable interface.
func (p *Disjunct[T]) Bounds() util.Bounds {
	return util.BoundsForArray(p.Args)
}

// TestAt implementation for Testable interface.
func (p *Disjunct[T]) TestAt(k int, tr trace.Module) (bool, uint, error) {
	//
	for _, disjunct := range p.Args {
		val, _, err := disjunct.TestAt(k, tr)
		//
		if err != nil {
			return val, 0, err
		} else if val {
			// Success
			return val, 0, nil
		}
	}
	// Failure
	return false, 0, nil
}

// Lisp returns a lisp representation of this equation, which is useful for
// debugging.
func (p *Disjunct[T]) Lisp(module schema.Module) sexp.SExp {
	if len(p.Args) == 0 {
		return sexp.NewSymbol("⊥")
	}

	return lispOfLogicalTerms(module, "∨", p.Args)
}

// RequiredRegisters implementation for Contextual interface.
func (p *Disjunct[T]) RequiredRegisters() *set.SortedSet[uint] {
	return requiredRegistersOfTerms(p.Args)
}

// RequiredCells implementation for Contextual interface
func (p *Disjunct[T]) RequiredCells(row int, mid trace.ModuleId) *set.AnySortedSet[trace.CellRef] {
	return requiredCellsOfTerms(p.Args, row, mid)
}

// Simplify this term as much as reasonably possible.
func (p *Disjunct[T]) Simplify(casts bool) T {
	// Simplify terms
	terms := simplifyLogicalTerms(p.Args, casts)
	// Flatten any nested disjuncts
	terms = array.Flatten(terms, flatternDisjunct)
	// True if contains True
	if array.ContainsMatching(terms, IsTrue) {
		return True[T]()
	}
	// Remove false values
	terms = array.RemoveMatching(terms, IsFalse)
	// Final checks
	switch len(terms) {
	case 0:
		return False[T]()
	case 1:
		return terms[0]
	default:
		return Disjunction(terms...)
	}
}

// Substitute implementation for Substitutable interface.
func (p *Disjunct[T]) Substitute(mapping map[string]fr.Element) {
	substituteTerms(mapping, p.Args...)
}

func flatternDisjunct[T LogicalTerm[T]](term T) []T {
	var e LogicalTerm[T] = term
	if t, ok := e.(*Disjunct[T]); ok {
		return t.Args
	}
	//
	return nil
}
