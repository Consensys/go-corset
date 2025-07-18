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

	"github.com/consensys/gnark-crypto/ecc/bls12-377/fr"
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
	Context schema.ModuleId
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
func (p *LookupFailure) RequiredCells(_ trace.Trace) *set.AnySortedSet[trace.CellRef] {
	res := set.NewAnySortedSet[trace.CellRef]()
	//
	for i, e := range p.Sources {
		if i != 0 || e.IsDefined() {
			res.InsertSorted(e.RequiredCells(int(p.Row), p.Context))
		}
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
	// Targets returns the target expressions which are used to lookup into the
	// target expressions.  NOTE: the first element here is *always* the target
	// selector.
	Targets []ir.Enclosed[[]E]
	// Sources returns the source expressions which are used to lookup into the
	// target expressions.  NOTE: the first element here is *always* the source
	// selector.
	Sources []ir.Enclosed[[]E]
}

// NewLookupConstraint creates a new lookup constraint with a given handle.
func NewLookupConstraint[E ir.Evaluable](handle string, targets []ir.Enclosed[[]E],
	sources []ir.Enclosed[[]E]) LookupConstraint[E] {
	var width int
	// Check sources
	for i, ith := range sources {
		if i != 0 && len(ith.Item) != width {
			panic("inconsistent number of source lookup columns")
		}

		width = len(ith.Item)
	}
	// Check targets
	for _, ith := range targets {
		if len(ith.Item) != width {
			panic("inconsistent number of target lookup columns")
		}
	}

	return LookupConstraint[E]{Handle: handle,
		Targets: targets,
		Sources: sources,
	}
}

// Consistent applies a number of internal consistency checks.  Whilst not
// strictly necessary, these can highlight otherwise hidden problems as an aid
// to debugging.
func (p LookupConstraint[E]) Consistent(_ schema.AnySchema) []error {
	return nil
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
func (p LookupConstraint[E]) Contexts() []schema.ModuleId {
	var contexts []schema.ModuleId
	// source contexts
	for _, source := range p.Sources {
		contexts = append(contexts, source.Module)
	}
	// target contexts
	for _, target := range p.Targets {
		contexts = append(contexts, target.Module)
	}
	//
	return contexts
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
	// sources
	for _, ith := range p.Sources {
		if module == ith.Module {
			for i, e := range ith.Item {
				if i != 0 || e.IsDefined() {
					eth := e.Bounds()
					bound.Union(&eth)
				}
			}
		}
	}
	// targets
	for _, ith := range p.Targets {
		if module == ith.Module {
			for i, e := range ith.Item {
				if i != 0 || e.IsDefined() {
					eth := e.Bounds()
					bound.Union(&eth)
				}
			}
		}
	}
	//
	return bound
}

// Accepts checks whether a lookup constraint into the target columns holds for
// all rows of the source columns.
//
//nolint:revive
func (p LookupConstraint[E]) Accepts(tr trace.Trace, sc schema.AnySchema) (bit.Set, schema.Failure) {
	var (
		coverage bit.Set
		// Determine width (in columns) of this lookup
		width int = len(p.Sources[0].Item) - 1
		//
		rows = hash.NewSet[hash.BytesKey](128)
		// Construct reusable buffer
		buffer = make([]byte, 32*width)
	)
	// Add all target columns to the set
	for _, ith := range p.Targets {
		var (
			tgtTrMod = tr.Module(ith.Module)
			tgtScMod = sc.Module(ith.Module)
			selector = ith.Item[0].IsDefined()
		)
		// Add each row in the target module
		for i := range tgtTrMod.Height() {
			ith_bytes, err := evalExprsAsBytes(int(i), selector, ith, p.Handle, tgtTrMod, tgtScMod, buffer[:])
			// error check
			if err != nil {
				return coverage, err
			} else if ith_bytes != nil {
				// Insert item, whilst checking whether the buffer was consumed or not.
				if !rows.Insert(hash.NewBytesKey(ith_bytes)) {
					// Yes, buffer consumed.  Therefore, construct a fresh buffer.
					buffer = make([]byte, 32*width)
				}
			}
		}
	}
	// Check all source columns are contained
	for _, ith := range p.Sources {
		var (
			srcTrMod = tr.Module(ith.Module)
			srcScMod = sc.Module(ith.Module)
			selector = ith.Item[0].IsDefined()
		)
		//
		for i := range srcTrMod.Height() {
			ith_bytes, err := evalExprsAsBytes(int(i), selector, ith, p.Handle, srcTrMod, srcScMod, buffer[:])
			// error check
			if err != nil {
				return coverage, err
			}
			// Check whether contained.
			if ith_bytes != nil && !rows.Contains(hash.NewBytesKey(ith_bytes)) {
				sources := make([]ir.Evaluable, width)
				for i, e := range ith.Item[1:] {
					sources[i] = e
				}
				// Construct failures
				return coverage, &LookupFailure{
					p.Handle, ith.Module, sources, i,
				}
			}
		}
	}
	//
	return coverage, nil
}

func evalExprsAsBytes[E ir.Evaluable](k int, selector bool, terms ir.Enclosed[[]E], handle string,
	trModule trace.Module, scModule schema.Module, bytes []byte) ([]byte, schema.Failure) {
	var (
		// Slice provides an access window for writing
		slice = bytes
		//
		sources = terms.Item
		i       = 0
	)
	// Check whether selector is defined.  If no selector exists, then it will
	// not be defined.
	if !selector {
		// If no selector, then skip that column.
		i = 1
	}
	// Evaluate each expression in turn (remembering that the first element is
	// the selector)
	for ; i < len(sources); i++ {
		ith, err := sources[i].EvalAt(k, trModule, scModule)
		// error check
		if err != nil {
			return nil, &InternalFailure{
				handle, terms.Module, uint(i), sources[i], err.Error(),
			}
		} else if i == 0 {
			// Selector determines whether or not this row is enabled.  If the
			// selector is 0 then this row is not enabled.
			if ith.Cmp(&frZero) == 0 {
				// Row is not enabled to ignore
				return nil, nil
			}
		} else {
			// Copy over each element
			binary.BigEndian.PutUint64(slice, ith[0])
			binary.BigEndian.PutUint64(slice[8:], ith[1])
			binary.BigEndian.PutUint64(slice[16:], ith[2])
			binary.BigEndian.PutUint64(slice[24:], ith[3])
			// Move slice over
			slice = slice[32:]
		}
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
		sources = sexp.EmptyList()
		targets = sexp.EmptyList()
	)
	// Iterate source expressions
	for _, ith := range p.Sources {
		var (
			source = sexp.EmptyList()
			module = schema.Module(ith.Module)
		)
		//
		for i, item := range ith.Item {
			if i == 0 && !item.IsDefined() {
				source.Append(sexp.NewSymbol("_"))
			} else {
				source.Append(item.Lisp(module))
			}
		}
		//
		sources.Append(source)
	}
	// Iterate target expressions
	for _, ith := range p.Targets {
		var (
			target = sexp.EmptyList()
			module = schema.Module(ith.Module)
		)
		//
		for i, item := range ith.Item {
			if i == 0 && !item.IsDefined() {
				target.Append(sexp.NewSymbol("_"))
			} else {
				target.Append(item.Lisp(module))
			}
		}
		//
		targets.Append(target)
	}
	// Done
	return sexp.NewList([]sexp.SExp{
		sexp.NewSymbol("lookup"),
		sexp.NewSymbol(p.Handle),
		targets,
		sources,
	})
}

// Substitute any matchined labelled constants within this constraint
func (p LookupConstraint[E]) Substitute(mapping map[string]fr.Element) {
	// Sources
	for _, ith := range p.Sources {
		for _, s := range ith.Item {
			s.Substitute(mapping)
		}
	}
	// Targets
	for _, ith := range p.Targets {
		for _, s := range ith.Item {
			s.Substitute(mapping)
		}
	}
}
