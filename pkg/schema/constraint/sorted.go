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
	"math/big"

	"github.com/consensys/gnark-crypto/ecc/bls12-377/fr"
	"github.com/consensys/go-corset/pkg/schema"
	sc "github.com/consensys/go-corset/pkg/schema"
	"github.com/consensys/go-corset/pkg/trace"
	tr "github.com/consensys/go-corset/pkg/trace"
	"github.com/consensys/go-corset/pkg/util"
	"github.com/consensys/go-corset/pkg/util/collection/bit"
	"github.com/consensys/go-corset/pkg/util/source/sexp"
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
	// BitWidth of delta (i.e. maximum difference between columns)
	BitWidth uint
	// Optional selector expression which determines on which rows this
	// constraint is active.
	Selector util.Option[E]
	// Sources returns the indices of the columns composing the "right" table of the
	// Sorted.
	Sources []E
	// Signs returns sorting direction of all columns.
	Signs []bool
	// Strict determines whether or not this constraint is strict (i.e. doesn't
	// permit equal values).
	Strict bool
}

// NewSortedConstraint creates a new Sorted
func NewSortedConstraint[E schema.Evaluable](handle string, context tr.Context, bitwidth uint, selector util.Option[E],
	sources []E, signs []bool, strict bool) *SortedConstraint[E] {
	//
	return &SortedConstraint[E]{handle, context, bitwidth, selector, sources, signs, strict}
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
	// NOTE: at the moment, we don't consider branches through sorted
	// constraints.
	return 1
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
		// Determine permitted range on delta value
		deltaBound := p.deltaBound()
		// Check all in-bounds values
		for k := bounds.Start + 1; k < (height - bounds.End); k++ {
			// Check selector
			if p.Selector.HasValue() {
				// Evaluate selector expression
				val, err := p.Selector.Unwrap().EvalAt(int(k), trace)
				// Check whether active (or not)
				if err != nil {
					return coverage, &sc.InternalFailure{Handle: p.Handle, Row: k, Error: err.Error()}
				} else if val.IsZero() {
					continue
				}
			}
			// Check sorting between rows k-1 and k
			if ok, err := sorted(k-1, k, deltaBound, p.Sources, p.Signs, p.Strict, trace); err != nil {
				return coverage, &sc.InternalFailure{Handle: p.Handle, Row: k, Error: err.Error()}
			} else if !ok {
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
	var (
		kind    = "sorted"
		sources = sexp.EmptyList()
	)
	//
	if p.Strict {
		kind = "strictsorted"
	}
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
	// Handle optional selector
	if p.Selector.IsEmpty() {
		return sexp.NewList([]sexp.SExp{
			sexp.NewSymbol(kind),
			sexp.NewSymbol(p.Handle),
			sources,
		})
	} else {
		return sexp.NewList([]sexp.SExp{
			sexp.NewSymbol(kind),
			sexp.NewSymbol(p.Handle),
			p.Selector.Unwrap().Lisp(schema),
			sources,
		})
	}
}

func (p *SortedConstraint[E]) deltaBound() fr.Element {
	var (
		two   fr.Element = fr.NewElement(2)
		bound fr.Element
	)
	//
	bound.Exp(two, big.NewInt(int64(p.BitWidth)))
	//
	return bound
}

func sorted[E schema.Evaluable](first, second uint, bound fr.Element, sources []E, signs []bool, strict bool,
	trace tr.Trace) (bool, error) {
	//
	var (
		delta    fr.Element
		lhs, rhs []fr.Element
		err      error
	)
	// Evaluate lhs
	if lhs, err = evalExprsAt(first, sources, trace); err != nil {
		return false, err
	}
	// Evaluate rhs
	if rhs, err = evalExprsAt(second, sources, trace); err != nil {
		return false, err
	}
	//
	for i := range signs {
		// Compare value
		c := lhs[i].Cmp(&rhs[i])
		// Check sorting criteria
		if c > 0 {
			// Compute delta
			delta.Sub(&lhs[i], &rhs[i])
			//
			return delta.Cmp(&bound) < 0 && !signs[i], nil
		} else if c < 0 {
			// Compute delta
			delta.Sub(&rhs[i], &lhs[i])
			//
			return delta.Cmp(&bound) < 0 && signs[i], nil
		}
	}
	// If we get here, then the elements are considered equal.  Thus, this is
	// only permitted if this is a non-strict ordering.
	return !strict, nil
}

func evalExprsAt[E schema.Evaluable](k uint, sources []E, tr trace.Trace) ([]fr.Element, error) {
	var err error
	//
	values := make([]fr.Element, len(sources))
	// Evaluate each expression in turn
	for i := 0; err == nil && i < len(sources); i++ {
		values[i], err = sources[i].EvalAt(int(k), tr)
	}
	//
	return values, err
}
