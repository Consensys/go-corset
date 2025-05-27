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
	"github.com/consensys/go-corset/pkg/ir"
	"github.com/consensys/go-corset/pkg/schema"
	"github.com/consensys/go-corset/pkg/trace"
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
	Expr ir.Evaluable
	// Range restriction
	Bitwidth uint
	// Row on which the constraint failed
	Row uint
}

// Message provides a suitable error message
func (p *RangeFailure) Message() string {
	// Construct useful error message
	return fmt.Sprintf("range \"%s\" is u%d does not hold (row %d)", p.Handle, p.Bitwidth, p.Row)
}

func (p *RangeFailure) String() string {
	return p.Message()
}

// RequiredCells identifies the cells required to evaluate the failing constraint at the failing row.
func (p *RangeFailure) RequiredCells(module trace.Module) *set.AnySortedSet[trace.CellRef] {
	return p.Expr.RequiredCells(int(p.Row), module)
}

// RangeConstraint restricts all values for a given expression to be within a
// range [0..n) for some bound n.  Any bound is supported, and the system will
// choose the best underlying implementation as needed.
type RangeConstraint[E ir.Evaluable] struct {
	// A unique identifier for this constraint.  This is primarily useful for
	// debugging.
	Handle string
	// Evaluation Context for this constraint which must match that of the
	// constrained expression itself.
	Context trace.Context
	// The expression whose values are being constrained to within the given
	// bound.
	Expr E
	// The number of bits permitted for all values matching this constraint.
	// For example, with a bitwidth of 8, the maximum permitted value is 255.
	Bitwidth uint
}

// NewRangeConstraint constructs a new Range constraint!
func NewRangeConstraint[E ir.Evaluable](handle string, context trace.Context,
	expr E, bitwidth uint) RangeConstraint[E] {
	return RangeConstraint[E]{handle, context, expr, bitwidth}
}

// Name returns a unique name for a given constraint.  This is useful
// purely for identifying constraints in reports, etc.
func (p RangeConstraint[E]) Name() string {
	return p.Handle
}

// Contexts returns the evaluation contexts (i.e. enclosing module + length
// multiplier) for this constraint.  Most constraints have only a single
// evaluation context, though some (e.g. lookups) have more.  Note that all
// constraints have at least one context (which we can call the "primary"
// context).
func (p RangeConstraint[E]) Contexts() []trace.Context {
	return []trace.Context{p.Context}
}

// Bounds determines the well-definedness bounds for this constraint for both
// the negative (left) or positive (right) directions.  For example, consider an
// expression such as "(shift X -1)".  This is technically undefined for the
// first row of any trace and, by association, any constraint evaluating this
// expression on that first row is also undefined (and hence must pass).
//
//nolint:revive
func (p RangeConstraint[E]) Bounds(module uint) util.Bounds {
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
func (p RangeConstraint[E]) Accepts(tr trace.Trace) (bit.Set, schema.Failure) {
	var (
		coverage bit.Set
		module   = tr.Module(p.Context.ModuleId)
		handle   = determineHandle(p.Handle, p.Context, tr)
		bound    = big.NewInt(2)
		frBound  fr.Element
	)
	// Compute 2^n
	bound.Exp(bound, big.NewInt(int64(p.Bitwidth)), nil)
	// Construct bound
	frBound.SetBigInt(bound)
	// Determine height of enclosing module
	height := tr.Height(p.Context)
	// Iterate every row
	for k := 0; k < int(height); k++ {
		// Get the value on the kth row
		kth, err := p.Expr.EvalAt(k, module)
		// Perform the range check
		if err != nil {
			return coverage, &schema.InternalFailure{
				Handle: p.Handle,
				Row:    uint(k),
				Term:   p.Expr,
				Error:  err.Error(),
			}
		} else if kth.Cmp(&frBound) >= 0 {
			// Evaluation failure
			return coverage, &RangeFailure{handle, p.Expr, p.Bitwidth, uint(k)}
		}
	}
	// All good
	return coverage, nil
}

// Lisp converts this schema element into a simple S-Expression, for example so
// it can be printed.
//
//nolint:revive
func (p RangeConstraint[E]) Lisp(schema schema.AnySchema) sexp.SExp {
	module := schema.Module(p.Context.ModuleId)
	//
	return sexp.NewList([]sexp.SExp{
		sexp.NewSymbol("range"),
		p.Expr.Lisp(module),
		sexp.NewSymbol(fmt.Sprintf("u%d", p.Bitwidth)),
	})
}
