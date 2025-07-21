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
	"encoding/binary"
	"fmt"

	"github.com/consensys/go-corset/pkg/schema"
	sc "github.com/consensys/go-corset/pkg/schema"
	"github.com/consensys/go-corset/pkg/trace"
	tr "github.com/consensys/go-corset/pkg/trace"
	"github.com/consensys/go-corset/pkg/util"
	"github.com/consensys/go-corset/pkg/util/collection/bit"
	"github.com/consensys/go-corset/pkg/util/collection/hash"
	"github.com/consensys/go-corset/pkg/util/collection/set"
	"github.com/consensys/go-corset/pkg/util/source/sexp"
)

// LookupFailure provides structural information about a failing lookup constraint.
type LookupFailure struct {
	// Handle of the failing constraint
	Handle string
	// Source expressions which were missing
	Sources []sc.Evaluable
	// Row on which the constraint failed
	Row uint
}

// Message provides a suitable error message
func (p *LookupFailure) Message() string {
	return fmt.Sprintf("lookup \"%s\" failed (row %d)", p.Handle, p.Row)
}

func (p *LookupFailure) String() string {
	return p.Message()
}

// RequiredCells identifies the cells required to evaluate the failing constraint at the failing row.
func (p *LookupFailure) RequiredCells(trace tr.Trace) *set.AnySortedSet[tr.CellRef] {
	res := set.NewAnySortedSet[tr.CellRef]()
	//
	for _, e := range p.Sources {
		res.InsertSorted(e.RequiredCells(int(p.Row), trace))
	}
	//
	return res
}

// LookupVector encapsulates all columns on one side of a lookup (i.e. it
// represents all source columns or all target columns).
type LookupVector[E schema.Evaluable] struct {
	// Context in which all terms are evaluated.
	TermContext trace.Context
	// Terms making up this vector.
	Terms []E
}

// NewLookupVector constructs a new column vector in a given context.
func NewLookupVector[E schema.Evaluable](context trace.Context, terms []E) LookupVector[E] {
	return LookupVector[E]{
		context,
		terms,
	}
}

// Context returns the conterxt in which all terms of this vector must be
// evaluated.
func (p *LookupVector[E]) Context() trace.Context {
	return p.TermContext
}

// Ith returns the ith term in this vector.
func (p *LookupVector[E]) Ith(index uint) E {
	return p.Terms[index]
}

// Len returns the number of items in this lookup vector.
func (p *LookupVector[E]) Len() uint {
	return uint(len(p.Terms))
}

// LookupConstraint (sometimes also called an inclusion constraint) constrains
// two sets of columns (potentially in different modules). Specifically, every
// row in the source columns must match a row in the target columns (but not
// vice-versa).  As such, the number of source columns must be the same as the
// number of target columns.  Furthermore, every source column must be in the
// same module, and likewise for target modules.  However, the source columns
// can be in a different module from the target columns.
//
// Lookup constraints are typically used to "connect" modules together.  We can
// think of them (in some ways) as being a little like function calls.  In this
// analogy, the source module is making a "function call" into the target
// module.  That is, the target module contains the set of valid input/output
// pairs (and perhaps other constraints to ensure the required relationship) and
// the source module is just checking that a given set of input/output pairs
// makes sense.
type LookupConstraint[E schema.Evaluable] struct {
	// Handle returns the handle for this lookup constraint which is simply an
	// identifier useful when debugging (i.e. to know which lookup failed, etc).
	Handle string
	// Source encapsulates the source expressions which are used to lookup into
	// the target expressions.
	Source LookupVector[E]
	// Target encapsulates the target expressions which are used to lookup into the
	// target expressions.
	Target LookupVector[E]
}

// NewLookupConstraint creates a new lookup constraint with a given handle.
func NewLookupConstraint[E schema.Evaluable](handle string, source LookupVector[E],
	target LookupVector[E]) *LookupConstraint[E] {
	//
	if target.Len() != source.Len() {
		panic("differeng number of target / source lookup columns")
	}

	return &LookupConstraint[E]{handle, source, target}
}

// Name returns a unique name for a given constraint.  This is useful
// purely for identifying constraints in reports, etc.
func (p *LookupConstraint[E]) Name() (string, uint) {
	return p.Handle, 0
}

// Contexts returns the evaluation contexts (i.e. enclosing module + length
// multiplier) for this constraint.  Most constraints have only a single
// evaluation context, though some (e.g. lookups) have more.  Note that all
// constraints have at least one context (which we can call the "primary"
// context).
func (p *LookupConstraint[E]) Contexts() []tr.Context {
	// source context designated as primary.
	return []tr.Context{p.Source.Context(), p.Target.Context()}
}

// Branches returns the total number of logical branches this constraint can
// take during evaluation.
func (p *LookupConstraint[E]) Branches() uint {
	// NOTE: at the moment, we don't consider branches through lookups.  This is
	// perhaps a degree of imprecision as some lookups have selectors.
	return 1
}

// Bounds determines the well-definedness bounds for this constraint for both
// the negative (left) or positive (right) directions.  For example, consider an
// expression such as "(shift X -1)".  This is technically undefined for the
// first row of any trace and, by association, any constraint evaluating this
// expression on that first row is also undefined (and hence must pass).
//
//nolint:revive
func (p *LookupConstraint[E]) Bounds(module uint) util.Bounds {
	var bound util.Bounds
	//
	if module == p.Source.Context().Module() {
		for _, e := range p.Source.Terms {
			eth := e.Bounds()
			bound.Union(&eth)
		}
	} else if module == p.Target.Context().Module() {
		for _, e := range p.Target.Terms {
			eth := e.Bounds()
			bound.Union(&eth)
		}
	}
	//
	return bound
}

// Accepts checks whether a lookup constraint into the target columns holds for
// all rows of the source columns.
//
//nolint:revive
func (p *LookupConstraint[E]) Accepts(tr trace.Trace) (bit.Set, schema.Failure) {
	var coverage bit.Set
	// Determine height of enclosing module for source columns
	src_height := tr.Height(p.Source.Context())
	tgt_height := tr.Height(p.Target.Context())
	//
	rows := hash.NewSet[hash.BytesKey](tgt_height)
	// Add all target columns to the set
	for i := 0; i < int(tgt_height); i++ {
		ith_bytes, err := evalExprsAsBytes(i, p.Target.Terms, p.Handle, tr)
		// error check
		if err != nil {
			return coverage, err
		}

		rows.Insert(hash.NewBytesKey(ith_bytes))
	}
	// Check all source columns are contained
	for i := 0; i < int(src_height); i++ {
		ith_bytes, err := evalExprsAsBytes(i, p.Source.Terms, p.Handle, tr)
		// error check
		if err != nil {
			return coverage, err
		}
		// Check whether contained.
		if !rows.Contains(hash.NewBytesKey(ith_bytes)) {
			sources := make([]sc.Evaluable, len(p.Source.Terms))
			for i, e := range p.Source.Terms {
				sources[i] = e
			}
			// Construct failures
			return coverage, &LookupFailure{
				p.Handle,
				sources,
				uint(i),
			}
		}
	}
	//
	return coverage, nil
}

// Lisp converts this schema element into a simple S-Expression, for example
// so it can be printed.
//
//nolint:revive
func (p *LookupConstraint[E]) Lisp(schema sc.Schema) sexp.SExp {
	sources := sexp.EmptyList()
	targets := sexp.EmptyList()
	// Iterate source expressions
	for i := range p.Source.Len() {
		sources.Append(p.Source.Ith(i).Lisp(schema))
	}
	// Iterate source expressions
	for i := range p.Target.Len() {
		targets.Append(p.Target.Ith(i).Lisp(schema))
	}
	// Done
	return sexp.NewList([]sexp.SExp{
		sexp.NewSymbol("lookup"),
		sexp.NewSymbol(p.Handle),
		targets,
		sources,
	})
}

func evalExprsAsBytes[E schema.Evaluable](k int, sources []E, handle string, tr trace.Trace) ([]byte, schema.Failure) {
	// Each fr.Element is 4 x 64bit words.
	bytes := make([]byte, 32*len(sources))
	// Slice provides an access window for writing
	slice := bytes
	// Evaluate each expression in turn
	for i := 0; i < len(sources); i++ {
		ith, err := sources[i].EvalAt(k, tr)
		// error check
		if err != nil {
			return nil, &sc.InternalFailure{
				Handle: handle, Row: uint(i), Term: sources[i], Error: err.Error(),
			}
		}
		// Copy over each element
		binary.BigEndian.PutUint64(slice, ith[0])
		binary.BigEndian.PutUint64(slice[8:], ith[1])
		binary.BigEndian.PutUint64(slice[16:], ith[2])
		binary.BigEndian.PutUint64(slice[24:], ith[3])
		// Move slice over
		slice = slice[32:]
	}
	// Done
	return bytes, nil
}
