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

// Conjunction builds the logical conjunction (i.e. and) for a given set of constraints.
func Conjunction[T LogicalTerm[T]](terms ...T) T {
	panic("todo")
}

// Bounds implementation for Boundable interface.
func (p *Conjunct[T]) Bounds() util.Bounds {
	return util.BoundsForArray(p.Args)
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
	panic("todo")
}

// RequiredRegisters implementation for Contextual interface.
func (p *Conjunct[T]) RequiredRegisters() *set.SortedSet[uint] {
	return requiredRegistersOfTerms(p.Args)
}

// RequiredCells implementation for Contextual interface
func (p *Conjunct[T]) RequiredCells(row int, tr trace.Module) *set.AnySortedSet[trace.CellRef] {
	return requiredCellsOfTerms(p.Args, row, tr)
}
