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
package constraint

import (
	"fmt"

	"github.com/consensys/gnark-crypto/ecc/bls12-377/fr"
	"github.com/consensys/go-corset/pkg/schema"
	sc "github.com/consensys/go-corset/pkg/schema"
	"github.com/consensys/go-corset/pkg/trace"
	tr "github.com/consensys/go-corset/pkg/trace"
	"github.com/consensys/go-corset/pkg/util"
	"github.com/consensys/go-corset/pkg/util/collection/bit"
	"github.com/consensys/go-corset/pkg/util/sexp"
)

// SortedFailure provides structural information about a failing Sorted constraint.
type SortedFailure struct {
	Msg string
}

// Message provides a suitable error message
func (p *SortedFailure) Message() string {
	return p.Msg
}

func (p *SortedFailure) String() string {
	return p.Msg
}

// SortedConstraint declares a constraint that one (or more) columns are
// lexicographically sorted.
type SortedConstraint[E schema.Evaluable] struct {
	Handle string
	// Evaluation Context for this constraint which must match that of the
	// source expressions.
	Context tr.Context
	// Sources returns the indices of the columns composing the "right" table of the
	// Sorted.
	Sources []E
	// Signs returns sorting direction of all columns.
	Signs []bool
}

// NewSortedConstraint creates a new Sorted
func NewSortedConstraint[E schema.Evaluable](handle string, context tr.Context, sources []E,
	signs []bool) *SortedConstraint[E] {
	//
	return &SortedConstraint[E]{handle, context, sources, signs}
}

// Name returns a unique name for a given constraint.  This is useful
// purely for identifying constraints in reports, etc.
func (p *SortedConstraint[E]) Name() (string, uint) {
	return p.Handle, 0
}

// Contexts returns the evaluation contexts (i.e. enclosing module + length
// multiplier) for this constraint.  Most constraints have only a single
// evaluation context, though some (e.g. lookups) have more.  Note that all
// constraints have at least one context (which we can call the "primary"
// context).
func (p *SortedConstraint[E]) Contexts() []tr.Context {
	return []tr.Context{p.Context}
}

// Branches returns the total number of logical branches this constraint can
// take during evaluation.
func (p *SortedConstraint[E]) Branches() uint {
	sum := uint(1)
	// Include source branches
	for _, e := range p.Sources {
		sum += e.Branches()
	}
	// Done
	return sum
}

// Bounds determines the well-definedness bounds for this constraint for both
// the negative (left) or positive (right) directions.  For example, consider an
// expression such as "(shift X -1)".  This is technically undefined for the
// first row of any trace and, by association, any constraint evaluating this
// expression on that first row is also undefined (and hence must pass).
func (p *SortedConstraint[E]) Bounds(module uint) util.Bounds {
	var bound util.Bounds
	//
	if module == p.Context.Module() {
		for _, e := range p.Sources {
			eth := e.Bounds()
			bound.Union(&eth)
		}
	}

	return bound
}

// Accepts checks whether a Sorted holds between the source and
// target columns.
func (p *SortedConstraint[E]) Accepts(trace tr.Trace) (bit.Set, sc.Failure) {
	var coverage bit.Set
	//
	height := trace.Height(p.Context)
	// Determine well-definedness bounds for this constraint
	bounds := p.Bounds(p.Context.Module())
	// Sanity check enough rows
	if bounds.End < height {
		// Check all in-bounds values
		for k := bounds.Start + 1; k < (height - bounds.End); k++ {
			// Check sorting between rows k-1 and k
			if !sorted(k-1, k, p.Sources, p.Signs, trace) {
				return coverage, &SortedFailure{fmt.Sprintf("sorted constraint \"%s\" failed (rows %d ~ %d)", p.Handle, k-1, k)}
			}
		}
	}
	//
	return coverage, nil
}

// Lisp converts this schema element into a simple S-Expression, for example
// so it can be printed.
func (p *SortedConstraint[E]) Lisp(schema sc.Schema) sexp.SExp {
	sources := sexp.EmptyList()
	// Iterate source expressions
	for i := 0; i < len(p.Sources); i++ {
		ith := p.Sources[i].Lisp(schema)
		//
		if i >= len(p.Signs) {
			//
		} else if p.Signs[i] {
			ith = sexp.NewList([]sexp.SExp{sexp.NewSymbol("+"), ith})
		} else {
			ith = sexp.NewList([]sexp.SExp{sexp.NewSymbol("-"), ith})
		}
		//
		sources.Append(ith)
	}

	return sexp.NewList([]sexp.SExp{
		sexp.NewSymbol("sorted"),
		sexp.NewSymbol(p.Handle),
		sources,
	})
}

func sorted[E schema.Evaluable](first, second uint, sources []E, signs []bool, trace tr.Trace) bool {
	lhs := evalExprsAt2(first, sources, trace)
	rhs := evalExprsAt2(second, sources, trace)
	//
	for i := range signs {
		// Compare value
		c := lhs[i].Cmp(&rhs[i])
		// Check sorting criteria
		if c > 0 {
			return !signs[i]
		} else if c < 0 {
			return signs[i]
		}
	}
	//
	return true
}

func evalExprsAt2[E schema.Evaluable](k uint, sources []E, tr trace.Trace) []fr.Element {
	values := make([]fr.Element, len(sources))
	// Evaluate each expression in turn
	for i := 0; i < len(sources); i++ {
		values[i] = sources[i].EvalAt(int(k), tr)
	}
	//
	return values
}
