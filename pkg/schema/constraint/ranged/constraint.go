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
package ranged

import (
	"fmt"

	"github.com/consensys/go-corset/pkg/ir"
	"github.com/consensys/go-corset/pkg/schema"
	"github.com/consensys/go-corset/pkg/schema/constraint"
	"github.com/consensys/go-corset/pkg/trace"
	"github.com/consensys/go-corset/pkg/util"
	"github.com/consensys/go-corset/pkg/util/collection/bit"
	"github.com/consensys/go-corset/pkg/util/field"
	"github.com/consensys/go-corset/pkg/util/source/sexp"
)

// Constraint restricts all values for a given expression to be within a
// range [0..n) for some bound n.  Any bound is supported, and the system will
// choose the best underlying implementation as needed.
type Constraint[F field.Element[F], E ir.Evaluable[F]] struct {
	// A unique identifier for this constraint.  This is primarily useful for
	// debugging.
	Handle string
	// Evaluation Context for this constraint which must match that of the
	// constrained expression itself.
	Context schema.ModuleId
	// The expression whose values are being constrained to within the given
	// bound.
	Expr E
	// The number of bits permitted for all values matching this constraint.
	// For example, with a bitwidth of 8, the maximum permitted value is 255.
	Bitwidth uint
}

// NewConstraint constructs a new Range constraint!
func NewConstraint[F field.Element[F], E ir.Evaluable[F]](handle string, context schema.ModuleId,
	expr E, bitwidth uint) Constraint[F, E] {
	return Constraint[F, E]{handle, context, expr, bitwidth}
}

// Consistent applies a number of internal consistency checks.  Whilst not
// strictly necessary, these can highlight otherwise hidden problems as an aid
// to debugging.
func (p Constraint[F, E]) Consistent(schema schema.AnySchema[F]) []error {
	return constraint.CheckConsistent(p.Context, schema, p.Expr)
}

// Name returns a unique name for a given constraint.  This is useful
// purely for identifying constraints in reports, etc.
func (p Constraint[F, E]) Name() string {
	return p.Handle
}

// Contexts returns the evaluation contexts (i.e. enclosing module + length
// multiplier) for this constraint.  Most constraints have only a single
// evaluation context, though some (e.g. lookups) have more.  Note that all
// constraints have at least one context (which we can call the "primary"
// context).
func (p Constraint[F, E]) Contexts() []schema.ModuleId {
	return []schema.ModuleId{p.Context}
}

// Bounds determines the well-definedness bounds for this constraint for both
// the negative (left) or positive (right) directions.  For example, consider an
// expression such as "(shift X -1)".  This is technically undefined for the
// first row of any trace and, by association, any constraint evaluating this
// expression on that first row is also undefined (and hence must pass).
//
//nolint:revive
func (p Constraint[F, E]) Bounds(module uint) util.Bounds {
	if p.Context == module {
		return p.Expr.Bounds()
	}
	//
	return util.EMPTY_BOUND
}

// Accepts checks whether a range constraint holds on every row of a table. If so, return
// nil otherwise return an error.
//
//nolint:revive
func (p Constraint[F, E]) Accepts(tr trace.Trace[F], sc schema.AnySchema[F]) (bit.Set, schema.Failure) {
	var (
		coverage bit.Set
		trModule = tr.Module(p.Context)
		scModule = sc.Module(p.Context)
		handle   = constraint.DetermineHandle(p.Handle, p.Context, tr)
		bound    F
	)
	// Compute 2^n
	bound = field.TwoPowN[F](p.Bitwidth)
	// Determine height of enclosing module
	height := tr.Module(p.Context).Height()
	// Iterate every row
	for k := 0; k < int(height); k++ {
		// Get the value on the kth row
		kth, err := p.Expr.EvalAt(k, trModule, scModule)
		// Perform the range check
		if err != nil {
			return coverage, &constraint.InternalFailure[F]{
				Handle:  p.Handle,
				Context: p.Context,
				Row:     uint(k),
				Term:    p.Expr,
				Error:   err.Error(),
			}
		} else if kth.Cmp(bound) >= 0 {
			// Evaluation failure
			return coverage, &Failure[F]{handle, p.Context, p.Expr, p.Bitwidth, uint(k)}
		}
	}
	// All good
	return coverage, nil
}

// Lisp converts this schema element into a simple S-Expression, for example so
// it can be printed.
//
//nolint:revive
func (p Constraint[F, E]) Lisp(mapping schema.ModuleRegisterMap) sexp.SExp {
	module := mapping.Module(p.Context)
	//
	return sexp.NewList([]sexp.SExp{
		sexp.NewSymbol("range"),
		p.Expr.Lisp(false, module),
		sexp.NewSymbol(fmt.Sprintf("u%d", p.Bitwidth)),
	})
}

// Substitute any matchined labelled constants within this constraint
func (p Constraint[F, E]) Substitute(mapping map[string]F) {
	p.Expr.Substitute(mapping)
}
