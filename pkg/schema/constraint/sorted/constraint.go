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
package sorted

import (
	"fmt"
	"math/big"

	"github.com/consensys/gnark-crypto/ecc/bls12-377/fr"
	"github.com/consensys/go-corset/pkg/ir"
	"github.com/consensys/go-corset/pkg/schema"
	"github.com/consensys/go-corset/pkg/schema/constraint"
	"github.com/consensys/go-corset/pkg/trace"
	"github.com/consensys/go-corset/pkg/util"
	"github.com/consensys/go-corset/pkg/util/collection/bit"
	bls12_377 "github.com/consensys/go-corset/pkg/util/field/bls12-377"
	"github.com/consensys/go-corset/pkg/util/source/sexp"
)

// Constraint declares a constraint that one (or more) columns are
// lexicographically sorted.
type Constraint[E ir.Evaluable] struct {
	Handle string
	// Evaluation Context for this constraint which must match that of the
	// source expressions.
	Context schema.ModuleId
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
func NewSortedConstraint[E ir.Evaluable](handle string, context schema.ModuleId, bitwidth uint, selector util.Option[E],
	sources []E, signs []bool, strict bool) Constraint[E] {
	//
	return Constraint[E]{handle, context, bitwidth, selector, sources, signs, strict}
}

// Consistent applies a number of internal consistency checks.  Whilst not
// strictly necessary, these can highlight otherwise hidden problems as an aid
// to debugging.
func (p Constraint[E]) Consistent(schema schema.AnySchema) []error {
	// TODO: add more useful checks
	return nil
}

// Name returns a unique name for a given constraint.  This is useful
// purely for identifying constraints in reports, etc.
func (p Constraint[E]) Name() string {
	return p.Handle
}

// Contexts returns the evaluation contexts (i.e. enclosing module + length
// multiplier) for this constraint.  Most constraints have only a single
// evaluation context, though some (e.g. lookups) have more.  Note that all
// constraints have at least one context (which we can call the "primary"
// context).
func (p Constraint[E]) Contexts() []schema.ModuleId {
	return []schema.ModuleId{p.Context}
}

// Bounds determines the well-definedness bounds for this constraint for both
// the negative (left) or positive (right) directions.  For example, consider an
// expression such as "(shift X -1)".  This is technically undefined for the
// first row of any trace and, by association, any constraint evaluating this
// expression on that first row is also undefined (and hence must pass).
func (p Constraint[E]) Bounds(module uint) util.Bounds {
	var bound util.Bounds
	//
	if module == p.Context {
		for _, e := range p.Sources {
			eth := e.Bounds()
			bound.Union(&eth)
		}
	}

	return bound
}

// Accepts checks whether a Sorted holds between the source and
// target columns.
func (p Constraint[E]) Accepts(tr trace.Trace[bls12_377.Element], sc schema.AnySchema) (bit.Set, schema.Failure) {
	var (
		coverage bit.Set
		// Determine enclosing module
		trModule = tr.Module(p.Context)
		scModule = sc.Module(p.Context)
		//
		height = trModule.Height()
		// Determine well-definedness bounds for this constraint
		bounds = p.Bounds(p.Context)
	)
	// Sanity check enough rows
	if bounds.End < height {
		// Determine permitted range on delta value
		deltaBound := p.deltaBound()
		// Construct temporary buffers which are reused between evaluations to
		// reduce memory pressure.
		lhs := make([]fr.Element, len(p.Sources))
		rhs := make([]fr.Element, len(p.Sources))
		// Check all in-bounds values
		for k := bounds.Start + 1; k < (height - bounds.End); k++ {
			// Check selector
			if p.Selector.HasValue() {
				selector := p.Selector.Unwrap()
				// Evaluate selector expression
				val, err := selector.EvalAt(int(k), trModule, scModule)
				// Check whether active (or not)
				if err != nil {
					return coverage, &constraint.InternalFailure{
						Handle:  p.Handle,
						Context: p.Context,
						Row:     k,
						Term:    selector,
						Error:   err.Error()}
				} else if val.IsZero() {
					continue
				}
			}
			// Check sorting between rows k-1 and k
			if ok, err := sorted(k-1, k, deltaBound, p.Sources, p.Signs, p.Strict, trModule, scModule, lhs, rhs); err != nil {
				return coverage, &constraint.InternalFailure{Handle: p.Handle, Context: p.Context, Row: k, Error: err.Error()}
			} else if !ok {
				return coverage, &Failure{fmt.Sprintf("sorted constraint \"%s\" failed (rows %d ~ %d)", p.Handle, k-1, k)}
			}
		}
	}
	//
	return coverage, nil
}

// Lisp converts this schema element into a simple S-Expression, for example
// so it can be printed.
func (p Constraint[E]) Lisp(schema schema.AnySchema) sexp.SExp {
	var (
		module  = schema.Module(p.Context)
		kind    = "sorted"
		sources = sexp.EmptyList()
	)
	//
	if p.Strict {
		kind = "strictsorted"
	}
	// Iterate source expressions
	for i := 0; i < len(p.Sources); i++ {
		ith := p.Sources[i].Lisp(module)
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
			p.Selector.Unwrap().Lisp(module),
			sources,
		})
	}
}

// Substitute any matchined labelled constants within this constraint
func (p Constraint[E]) Substitute(mapping map[string]fr.Element) {
	for _, s := range p.Sources {
		s.Substitute(mapping)
	}
	//
	if p.Selector.HasValue() {
		p.Selector.Unwrap().Substitute(mapping)
	}
}

func (p Constraint[E]) deltaBound() fr.Element {
	var (
		two   fr.Element = fr.NewElement(2)
		bound fr.Element
	)
	//
	bound.Exp(two, big.NewInt(int64(p.BitWidth)))
	//
	return bound
}

func sorted[E ir.Evaluable](first, second uint, bound fr.Element, sources []E, signs []bool, strict bool,
	trMod trace.Module, scMod schema.Module, lhs []fr.Element, rhs []fr.Element) (bool, error) {
	//
	var (
		delta fr.Element
		err   error
	)
	// Evaluate lhs
	if err = evalExprsAt(first, sources, trMod, scMod, lhs); err != nil {
		return false, err
	}
	// Evaluate rhs
	if err = evalExprsAt(second, sources, trMod, scMod, rhs); err != nil {
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

func evalExprsAt[E ir.Evaluable](k uint, sources []E, trMod trace.Module, scMod schema.Module,
	buffer []fr.Element) error {
	//
	var err error
	// Evaluate each expression in turn
	for i := 0; err == nil && i < len(sources); i++ {
		buffer[i], err = sources[i].EvalAt(int(k), trMod, scMod)
	}
	//
	return err
}
