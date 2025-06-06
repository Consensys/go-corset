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
	"github.com/consensys/go-corset/pkg/util/collection/set"
	"github.com/consensys/go-corset/pkg/util/source/sexp"
)

// RangeFailure provides structural information about a failing type constraint.
type RangeFailure struct {
	// Handle of the failing constraint
	Handle string
	// Constraint expression
	Expr sc.Evaluable
	// Range restriction
	Bound fr.Element
	// Row on which the constraint failed
	Row uint
}

// Message provides a suitable error message
func (p *RangeFailure) Message() string {
	// Construct useful error message
	return fmt.Sprintf("range \"%s\" < %s does not hold (row %d)", p.Handle, p.Bound.String(), p.Row)
}

func (p *RangeFailure) String() string {
	return p.Message()
}

// RequiredCells identifies the cells required to evaluate the failing constraint at the failing row.
func (p *RangeFailure) RequiredCells(trace tr.Trace) *set.AnySortedSet[tr.CellRef] {
	return p.Expr.RequiredCells(int(p.Row), trace)
}

// RangeConstraint restricts all values for a given expression to be within a
// range [0..n) for some bound n.  Any bound is supported, and the system will
// choose the best underlying implementation as needed.
type RangeConstraint[E sc.Evaluable] struct {
	// A unique identifier for this constraint.  This is primarily useful for
	// debugging.
	Handle string
	// A further differentiator to manage distinct low-level constraints arising
	// from high-level constraints.
	Case uint
	// Evaluation Context for this constraint which must match that of the
	// constrained expression itself.
	Context trace.Context
	// The expression whose values are being constrained to within the given
	// bound.
	Expr E
	// The upper Bound for this constraint.  Specifically, every evaluation of
	// the expression should produce a value strictly below this Bound.  NOTE:
	// an fr.Element is used here to store the Bound simply to make the
	// necessary comparison against table data more direct.
	Bound fr.Element
}

// NewRangeConstraint constructs a new Range constraint!
func NewRangeConstraint[E sc.Evaluable](handle string, casenum uint, context trace.Context,
	expr E, bound fr.Element) *RangeConstraint[E] {
	return &RangeConstraint[E]{handle, casenum, context, expr, bound}
}

// Name returns a unique name for a given constraint.  This is useful
// purely for identifying constraints in reports, etc.
func (p *RangeConstraint[E]) Name() (string, uint) {
	return p.Handle, p.Case
}

// Contexts returns the evaluation contexts (i.e. enclosing module + length
// multiplier) for this constraint.  Most constraints have only a single
// evaluation context, though some (e.g. lookups) have more.  Note that all
// constraints have at least one context (which we can call the "primary"
// context).
func (p *RangeConstraint[E]) Contexts() []tr.Context {
	return []tr.Context{p.Context}
}

// Branches returns the total number of logical branches this constraint can
// take during evaluation.
func (p *RangeConstraint[E]) Branches() uint {
	return p.Expr.Branches()
}

// BoundedAtMost determines whether the bound for this constraint is at most a given bound.
func (p *RangeConstraint[E]) BoundedAtMost(bound uint) bool {
	var n fr.Element = fr.NewElement(uint64(bound))
	return p.Bound.Cmp(&n) <= 0
}

// Bounds determines the well-definedness bounds for this constraint for both
// the negative (left) or positive (right) directions.  For example, consider an
// expression such as "(shift X -1)".  This is technically undefined for the
// first row of any trace and, by association, any constraint evaluating this
// expression on that first row is also undefined (and hence must pass).
//
//nolint:revive
func (p *RangeConstraint[E]) Bounds(module uint) util.Bounds {
	if p.Context.Module() == module {
		return p.Expr.Bounds()
	}
	//
	return util.EMPTY_BOUND
}

// Accepts checks whether a range constraint holds on every row of a table. If so, return
// nil otherwise return an error.
//
//nolint:revive
func (p *RangeConstraint[E]) Accepts(tr trace.Trace) (bit.Set, schema.Failure) {
	var (
		coverage bit.Set
		handle   = determineHandle(p.Handle, p.Context, tr)
	)
	// Determine height of enclosing module
	height := tr.Height(p.Context)
	// Iterate every row
	for k := 0; k < int(height); k++ {
		// Get the value on the kth row
		kth, err := p.Expr.EvalAt(k, tr)
		// Perform the range check
		if err != nil {
			return coverage, &sc.InternalFailure{
				Handle: p.Handle,
				Row:    uint(k),
				Term:   p.Expr,
				Error:  err.Error(),
			}
		} else if kth.Cmp(&p.Bound) >= 0 {
			// Evaluation failure
			return coverage, &RangeFailure{handle, p.Expr, p.Bound, uint(k)}
		}
	}
	// All good
	return coverage, nil
}

// Lisp converts this schema element into a simple S-Expression, for example so
// it can be printed.
//
//nolint:revive
func (p *RangeConstraint[E]) Lisp(schema sc.Schema) sexp.SExp {
	//
	return sexp.NewList([]sexp.SExp{
		sexp.NewSymbol("range"),
		p.Expr.Lisp(schema),
		sexp.NewSymbol(p.Bound.String()),
	})
}
