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

	"github.com/consensys/go-corset/pkg/ir"
	"github.com/consensys/go-corset/pkg/schema"
	"github.com/consensys/go-corset/pkg/trace"
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
	// Relevant context for source expressions.
	Context trace.Context
	// Source expressions which were missing
	Sources []ir.Evaluable
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
func (p *LookupFailure) RequiredCells(tr trace.Trace) *set.AnySortedSet[trace.CellRef] {
	module := tr.Module(p.Context.ModuleId)
	res := set.NewAnySortedSet[trace.CellRef]()
	//
	for _, e := range p.Sources {
		res.InsertSorted(e.RequiredCells(int(p.Row), module))
	}
	//
	return res
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
type LookupConstraint[E ir.Evaluable] struct {
	// Handle returns the handle for this lookup constraint which is simply an
	// identifier useful when debugging (i.e. to know which lookup failed, etc).
	Handle string
	// Context in which all source columns are evaluated.
	SourceContext trace.Context
	// Context in which all target columns are evaluated.
	TargetContext trace.Context
	// Sources returns the source expressions which are used to lookup into the
	// target expressions.
	Sources []E
	// Targets returns the target expressions which are used to lookup into the
	// target expressions.
	Targets []E
}

// NewLookupConstraint creates a new lookup constraint with a given handle.
func NewLookupConstraint[E ir.Evaluable](handle string, source trace.Context,
	target trace.Context, sources []E, targets []E) LookupConstraint[E] {
	if len(targets) != len(sources) {
		panic("differeng number of target / source lookup columns")
	}

	return LookupConstraint[E]{handle, source, target, sources, targets}
}

// Consistent applies a number of internal consistency checks.  Whilst not
// strictly necessary, these can highlight otherwise hidden problems as an aid
// to debugging.
func (p LookupConstraint[E]) Consistent(schema schema.AnySchema) []error {
	var (
		srcErrors = checkConsistent[E](p.SourceContext.ModuleId, schema, p.Sources...)
		dstErrors = checkConsistent[E](p.TargetContext.ModuleId, schema, p.Targets...)
		errors    = append(srcErrors, dstErrors...)
	)
	// Check consistent register widths
	if len(p.Sources) != len(p.Targets) {
		err := fmt.Errorf("inconsistent number of source / target registers (%d vs %d)", len(p.Sources), len(p.Targets))
		errors = append(errors, err)
	}
	// TODO: check lookup widths (using range analysis?)
	// Done
	return errors
}

// Name returns a unique name for a given constraint.  This is useful
// purely for identifying constraints in reports, etc.
func (p LookupConstraint[E]) Name() string {
	return p.Handle
}

// Contexts returns the evaluation contexts (i.e. enclosing module + length
// multiplier) for this constraint.  Most constraints have only a single
// evaluation context, though some (e.g. lookups) have more.  Note that all
// constraints have at least one context (which we can call the "primary"
// context).
func (p LookupConstraint[E]) Contexts() []trace.Context {
	// source context designated as primary.
	return []trace.Context{p.SourceContext, p.TargetContext}
}

// Bounds determines the well-definedness bounds for this constraint for both
// the negative (left) or positive (right) directions.  For example, consider an
// expression such as "(shift X -1)".  This is technically undefined for the
// first row of any trace and, by association, any constraint evaluating this
// expression on that first row is also undefined (and hence must pass).
//
//nolint:revive
func (p LookupConstraint[E]) Bounds(module uint) util.Bounds {
	var bound util.Bounds
	//
	if module == p.SourceContext.Module() {
		for _, e := range p.Sources {
			eth := e.Bounds()
			bound.Union(&eth)
		}
	} else if module == p.TargetContext.Module() {
		for _, e := range p.Targets {
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
func (p LookupConstraint[E]) Accepts(tr trace.Trace) (bit.Set, schema.Failure) {
	var (
		coverage  bit.Set
		srcModule = tr.Module(p.SourceContext.ModuleId)
		tgtModule = tr.Module(p.TargetContext.ModuleId)
		// Determine height of enclosing module for source columns
		srcHeight = tr.Height(p.SourceContext)
		tgtHeight = tr.Height(p.TargetContext)
		//
		rows = hash.NewSet[hash.BytesKey](tgtHeight)
	)
	// Add all target columns to the set
	for i := 0; i < int(tgtHeight); i++ {
		ith_bytes, err := evalExprsAsBytes(i, p.Targets, p.Handle, p.TargetContext, tgtModule)
		// error check
		if err != nil {
			return coverage, err
		}

		rows.Insert(hash.NewBytesKey(ith_bytes))
	}
	// Check all source columns are contained
	for i := 0; i < int(srcHeight); i++ {
		ith_bytes, err := evalExprsAsBytes(i, p.Sources, p.Handle, p.SourceContext, srcModule)
		// error check
		if err != nil {
			return coverage, err
		}
		// Check whether contained.
		if !rows.Contains(hash.NewBytesKey(ith_bytes)) {
			sources := make([]ir.Evaluable, len(p.Sources))
			for i, e := range p.Sources {
				sources[i] = e
			}
			// Construct failures
			return coverage, &LookupFailure{
				p.Handle,
				p.SourceContext,
				sources,
				uint(i),
			}
		}
	}
	//
	return coverage, nil
}

func evalExprsAsBytes[E ir.Evaluable](k int, sources []E, handle string, ctx trace.Context,
	module trace.Module) ([]byte, schema.Failure) {
	// Each fr.Element is 4 x 64bit words.
	bytes := make([]byte, 32*len(sources))
	// Slice provides an access window for writing
	slice := bytes
	// Evaluate each expression in turn
	for i := 0; i < len(sources); i++ {
		ith, err := sources[i].EvalAt(k, module)
		// error check
		if err != nil {
			return nil, &InternalFailure{
				handle, ctx, uint(i), sources[i], err.Error(),
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

// Lisp converts this schema element into a simple S-Expression, for example
// so it can be printed.
//
//nolint:revive
func (p LookupConstraint[E]) Lisp(schema schema.AnySchema) sexp.SExp {
	var (
		sourceModule = schema.Module(p.SourceContext.ModuleId)
		targetModule = schema.Module(p.TargetContext.ModuleId)
		sources      = sexp.EmptyList()
		targets      = sexp.EmptyList()
	)
	// Iterate source expressions
	for i := range p.Sources {
		sources.Append(p.Sources[i].Lisp(sourceModule))
	}
	// Iterate source expressions
	for i := range p.Targets {
		targets.Append(p.Targets[i].Lisp(targetModule))
	}
	// Done
	return sexp.NewList([]sexp.SExp{
		sexp.NewSymbol("lookup"),
		sexp.NewSymbol(p.Handle),
		targets,
		sources,
	})
}
