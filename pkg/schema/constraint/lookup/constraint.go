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
package lookup

import (
	"encoding/binary"

	"github.com/consensys/go-corset/pkg/ir"
	"github.com/consensys/go-corset/pkg/schema"
	"github.com/consensys/go-corset/pkg/schema/constraint"
	"github.com/consensys/go-corset/pkg/trace"
	"github.com/consensys/go-corset/pkg/util"
	"github.com/consensys/go-corset/pkg/util/collection/bit"
	"github.com/consensys/go-corset/pkg/util/collection/hash"
	"github.com/consensys/go-corset/pkg/util/field/bls12_377"
	"github.com/consensys/go-corset/pkg/util/source/sexp"
)

// Constraint (sometimes also called an inclusion constraint) constrains
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
type Constraint[E ir.Evaluable[bls12_377.Element]] struct {
	// Handle returns the handle for this lookup constraint which is simply an
	// identifier useful when debugging (i.e. to know which lookup failed, etc).
	Handle string
	// Targets returns the target expressions which are used to lookup into the
	// target expressions.  NOTE: the first element here is *always* the target
	// selector.
	Targets []Vector[bls12_377.Element, E]
	// Sources returns the source expressions which are used to lookup into the
	// target expressions.  NOTE: the first element here is *always* the source
	// selector.
	Sources []Vector[bls12_377.Element, E]
}

// NewConstraint creates a new lookup constraint with a given handle.
func NewConstraint[E ir.Evaluable[bls12_377.Element]](handle string, targets []Vector[bls12_377.Element, E],
	sources []Vector[bls12_377.Element, E]) Constraint[E] {
	var width uint
	// Check sources
	for i, ith := range sources {
		if i != 0 && ith.Len() != width {
			panic("inconsistent number of source lookup columns")
		}

		width = ith.Len()
	}
	// Check targets
	for _, ith := range targets {
		if ith.Len() != width {
			panic("inconsistent number of target lookup columns")
		}
	}

	return Constraint[E]{Handle: handle,
		Targets: targets,
		Sources: sources,
	}
}

// Consistent applies a number of internal consistency checks.  Whilst not
// strictly necessary, these can highlight otherwise hidden problems as an aid
// to debugging.
func (p Constraint[E]) Consistent(_ schema.AnySchema) []error {
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
func (p Constraint[E]) Bounds(module uint) util.Bounds {
	var bound util.Bounds
	// sources
	for _, ith := range p.Sources {
		eth := ith.Bounds(module)
		bound.Union(&eth)
	}
	// targets
	for _, ith := range p.Targets {
		eth := ith.Bounds(module)
		bound.Union(&eth)
	}
	//
	return bound
}

// Accepts checks whether a lookup constraint into the target columns holds for
// all rows of the source columns.
//
//nolint:revive
func (p Constraint[E]) Accepts(tr trace.Trace[bls12_377.Element], sc schema.AnySchema) (bit.Set, schema.Failure) {
	var (
		coverage bit.Set
		// Determine width (in columns) of this lookup
		width uint = p.Sources[0].Len()
		//
		rows *hash.Set[hash.BytesKey]
		// Construct reusable buffer
		bytes = make([]byte, 32*width)
		err   schema.Failure
	)
	// Insert all active target vectors
	if rows, err = p.insertTargetVectors(tr, sc, bytes); err != nil {
		return coverage, err
	}
	// Check against all active source vectors
	if err = p.checkSourceVectors(rows, tr, sc, bytes); err != nil {
		return coverage, err
	}
	//
	return coverage, nil
}

func (p *Constraint[E]) insertTargetVectors(tr trace.Trace[bls12_377.Element], sc schema.AnySchema,
	bytes []byte) (*hash.Set[hash.BytesKey], schema.Failure) {
	//
	var (
		rows = hash.NewSet[hash.BytesKey](tr.Module(p.Targets[0].Module).Height())
	)
	// Choose optimised loop
	for _, target := range p.Targets {
		var (
			trModule = tr.Module(target.Module)
			scModule = sc.Module(target.Module)
			height   = trModule.Height()
		)
		//
		if target.HasSelector() {
			// unfiltered
			for i := range int(height) {
				if err := insertFilteredTargetVector(i, target, p.Handle, rows, trModule, scModule, bytes); err != nil {
					return nil, err
				}
			}
		} else {
			// unfiltered
			for i := range int(height) {
				if err := insertTargetVector(i, target, p.Handle, rows, trModule, scModule, bytes); err != nil {
					return nil, err
				}
			}
		}
	}
	//
	return rows, nil
}

func (p *Constraint[E]) checkSourceVectors(
	rows *hash.Set[hash.BytesKey],
	tr trace.Trace[bls12_377.Element],
	sc schema.AnySchema,
	bytes []byte,
) schema.Failure {
	// Choose optimised loop
	for _, source := range p.Sources {
		var (
			trModule = tr.Module(source.Module)
			scModule = sc.Module(source.Module)
			height   = trModule.Height()
		)
		//
		if source.HasSelector() {
			// filtered
			for i := range int(height) {
				if err := checkFilteredSourceVector(i, source, p.Handle, rows, trModule, scModule, bytes); err != nil {
					return err
				}
			}
		} else {
			// unfiltered
			for i := range int(height) {
				if err := checkSourceVector(i, source, p.Handle, rows, trModule, scModule, bytes); err != nil {
					return err
				}
			}
		}
	}
	// success
	return nil
}

func insertFilteredTargetVector[E ir.Evaluable[bls12_377.Element]](k int, vec Vector[bls12_377.Element, E], handle string, rows *hash.Set[hash.BytesKey],
	trModule trace.Module[bls12_377.Element], scModule schema.Module, bytes []byte) schema.Failure {
	// If no selector, then always selected
	var selected bool = !vec.HasSelector()
	//
	if vec.HasSelector() {
		// Otherwise, check whether selector enabled (or not).
		var (
			selector = vec.Selector.Unwrap()
			ith, err = selector.EvalAt(k, trModule, scModule)
		)
		//
		if err != nil {
			return &constraint.InternalFailure{
				Handle:  handle,
				Context: vec.Module,
				Row:     uint(k),
				Term:    vec.Selector.Unwrap(),
				Error:   err.Error(),
			}
		}
		// Selected when non-zero
		selected = !ith.IsZero()
	}
	// If row selected, then insert contents!
	if selected {
		return insertTargetVector(k, vec, handle, rows, trModule, scModule, bytes)
	}
	//
	return nil
}

func insertTargetVector[E ir.Evaluable[bls12_377.Element]](k int, vec Vector[bls12_377.Element, E], handle string,
	rows *hash.Set[hash.BytesKey], trModule trace.Module[bls12_377.Element], scModule schema.Module, bytes []byte) schema.Failure {
	//
	// Check each source is included
	if err := evalExprsAsBytes(k, vec, handle, trModule, scModule, bytes); err != nil {
		return err
	}
	//
	rows.Insert(hash.NewBytesKey(bytes))
	//
	return nil
}

func checkFilteredSourceVector[E ir.Evaluable[bls12_377.Element]](k int, vec Vector[bls12_377.Element, E], handle string, rows *hash.Set[hash.BytesKey],
	trModule trace.Module[bls12_377.Element], scModule schema.Module, bytes []byte) schema.Failure {
	// If no selector, then always selected
	var selected bool = !vec.HasSelector()
	//
	if vec.HasSelector() {
		// Otherwise, check whether selector enabled (or not).
		var (
			selector = vec.Selector.Unwrap()
			ith, err = selector.EvalAt(k, trModule, scModule)
		)
		//
		if err != nil {
			return &constraint.InternalFailure{
				Handle:  handle,
				Context: vec.Module,
				Row:     uint(k),
				Term:    vec.Selector.Unwrap(),
				Error:   err.Error(),
			}
		}
		// Selected when non-zero
		selected = !ith.IsZero()
	}
	// If row selected, then check contents!
	if selected {
		return checkSourceVector(k, vec, handle, rows, trModule, scModule, bytes)
	}
	//
	return nil
}

func checkSourceVector[E ir.Evaluable[bls12_377.Element]](k int, vec Vector[bls12_377.Element, E], handle string, rows *hash.Set[hash.BytesKey],
	trModule trace.Module[bls12_377.Element], scModule schema.Module, bytes []byte) schema.Failure {
	// Check each source is included
	if err := evalExprsAsBytes(k, vec, handle, trModule, scModule, bytes); err != nil {
		return err
	}
	// Check whether contained.
	if !rows.Contains(hash.NewBytesKey(bytes)) {
		sources := make([]ir.Evaluable[bls12_377.Element], vec.Len())
		for i, e := range vec.Terms {
			sources[i] = e
		}
		// Construct failures
		return &Failure{handle, vec.Module, sources, uint(k)}
	}
	// success
	return nil
}

func evalExprsAsBytes[E ir.Evaluable[bls12_377.Element]](k int, vec Vector[bls12_377.Element, E], handle string, trModule trace.Module[bls12_377.Element],
	scModule schema.Module, bytes []byte) schema.Failure {
	var (
		// Slice provides an access window for writing
		slice = bytes
	)
	// Evaluate each expression in turn (remembering that the first element is
	// the selector)
	for i := uint(0); i < vec.Len(); i++ {
		ith, err := vec.Ith(i).EvalAt(k, trModule, scModule)
		// error check
		if err != nil {
			return &constraint.InternalFailure{
				Handle:  handle,
				Context: vec.Module,
				Row:     uint(k),
				Term:    vec.Ith(i),
				Error:   err.Error(),
			}
		} else {
			// Copy over each element
			binary.BigEndian.PutUint64(slice, ith.Element[0])
			binary.BigEndian.PutUint64(slice[8:], ith.Element[1])
			binary.BigEndian.PutUint64(slice[16:], ith.Element[2])
			binary.BigEndian.PutUint64(slice[24:], ith.Element[3])
			// Move slice over
			slice = slice[32:]
		}
	}
	// Done
	return nil
}

// Lisp converts this schema element into a simple S-Expression, for example
// so it can be printed.
//
//nolint:revive
func (p Constraint[E]) Lisp(schema schema.AnySchema) sexp.SExp {
	var (
		sources = sexp.EmptyList()
		targets = sexp.EmptyList()
	)
	// Iterate source expressions
	for _, ith := range p.Sources {
		sources.Append(ith.Lisp(schema))
	}
	// Iterate target expressions
	for _, ith := range p.Targets {
		targets.Append(ith.Lisp(schema))
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
func (p Constraint[E]) Substitute(mapping map[string]bls12_377.Element) {
	// Sources
	for _, ith := range p.Sources {
		ith.Substitute(mapping)
	}
	// Targets
	for _, ith := range p.Targets {
		ith.Substitute(mapping)
	}
}
