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
package assignment

import (
	"encoding/binary"
	"encoding/gob"
	"fmt"
	"slices"

	"github.com/consensys/gnark-crypto/ecc/bls12-377/fr"
	sc "github.com/consensys/go-corset/pkg/schema"
	tr "github.com/consensys/go-corset/pkg/trace"
	"github.com/consensys/go-corset/pkg/util"
	"github.com/consensys/go-corset/pkg/util/collection/hash"
	"github.com/consensys/go-corset/pkg/util/field"
	"github.com/consensys/go-corset/pkg/util/source/sexp"
)

// Computation currently describes a native computation which accepts a set of
// input columns, and assigns a set of output columns.
type Computation struct {
	// Name of the function being invoked.
	Function string
	// Target columns declared by this sorted permutation (in the order
	// of declaration).
	Targets []sc.RegisterRef
	// Source columns which define the new (sorted) columns.
	Sources []sc.RegisterRef
}

// NewComputation defines a set of target columns which are assigned from a
// given set of source columns using a function to multiplex input to output.
func NewComputation(fn string, targets []sc.RegisterRef, sources []sc.RegisterRef) *Computation {
	//
	return &Computation{fn, targets, sources}
}

// ============================================================================
// Assignment Interface
// ============================================================================

// Bounds determines the well-definedness bounds for this assignment for both
// the negative (left) or positive (right) directions.  For example, consider an
// expression such as "(shift X -1)".  This is technically undefined for the
// first row of any trace and, by association, any constraint evaluating this
// expression on that first row is also undefined (and hence must pass).
func (p *Computation) Bounds(_ sc.ModuleId) util.Bounds {
	return util.EMPTY_BOUND
}

// Compute computes the values of columns defined by this assignment. This
// requires copying the data in the source columns, and sorting that data
// according to the permutation criteria.
func (p *Computation) Compute(trace tr.Trace, schema sc.AnySchema) ([]tr.ArrayColumn, error) {
	var (
		fn func([]field.FrArray) []field.FrArray
		ok bool
	)
	// Sanity check
	if fn, ok = NATIVES[p.Function]; !ok {
		panic(fmt.Sprintf("unknown native function: %s", p.Function))
	}
	// Go!
	return computeNative(p.Sources, p.Targets, fn, trace, schema), nil
}

// Consistent performs some simple checks that the given schema is consistent.
// This provides a double check of certain key properties, such as that
// registers used for assignments are large enough, etc.
func (p *Computation) Consistent(schema sc.AnySchema) []error {
	// NOTE: this is where we could (in principle) check the type of the
	// function being defined to ensure it is, for example, typed correctly.
	return nil
}

// RegistersRead returns the set of columns that this assignment depends upon.
// That can include both input columns, as well as other computed columns.
func (p *Computation) RegistersRead() []sc.RegisterRef {
	return p.Sources
}

// RegistersWritten identifies registers assigned by this assignment.
func (p *Computation) RegistersWritten() []sc.RegisterRef {
	return p.Targets
}

// ============================================================================
// Lispify Interface
// ============================================================================

// Lisp converts this schema element into a simple S-Expression, for example
// so it can be printed.
func (p *Computation) Lisp(schema sc.AnySchema) sexp.SExp {
	var (
		targets = sexp.EmptyList()
		sources = sexp.EmptyList()
	)

	for _, ref := range p.Targets {
		module := schema.Module(ref.Module())
		ith := module.Register(ref.Register())
		name := sexp.NewSymbol(ith.QualifiedName(module))
		datatype := sexp.NewSymbol(fmt.Sprintf("u%d", ith.Width))
		def := sexp.NewList([]sexp.SExp{name, datatype})
		targets.Append(def)
	}

	for _, ref := range p.Sources {
		module := schema.Module(ref.Module())
		ith := module.Register(ref.Register())
		ith_name := ith.QualifiedName(module)
		sources.Append(sexp.NewSymbol(ith_name))
	}

	return sexp.NewList([]sexp.SExp{
		sexp.NewSymbol("compute"),
		targets,
		sexp.NewSymbol(p.Function),
		sources,
	})
}

// ============================================================================
// Native Generic Computation
// ============================================================================

type NativeComputation func([]field.FrArray) []field.FrArray

func computeNative(sources []sc.RegisterRef, targets []sc.RegisterRef, fn NativeComputation, trace tr.Trace, schema sc.AnySchema) []tr.ArrayColumn {
	// Read inputs
	inputs := readRegisters(trace, sources...)
	// Read inputs
	for i, ref := range sources {
		mid, rid := ref.Module(), ref.Register().Unwrap()
		inputs[i] = trace.Module(mid).Column(rid).Data()
	}
	// Apply native function
	data := fn(inputs)
	// Write outputs
	return writeRegisters(schema, targets, data)
}

// ============================================================================
// Native Function Definitions
// ============================================================================

// NATIVES map holds the supported set of native computations.
var NATIVES = map[string]func([]field.FrArray) []field.FrArray{
	"id":                   idNativeFunction,
	"interleave":           interleaveNativeFunction,
	"filter":               filterNativeFunction,
	"map-if":               mapIfNativeFunction,
	"fwd-changes-within":   fwdChangesWithinNativeFunction,
	"fwd-unchanged-within": fwdUnchangedWithinNativeFunction,
	"bwd-changes-within":   bwdChangesWithinNativeFunction,
	"fwd-fill-within":      fwdFillWithinNativeFunction,
	"bwd-fill-within":      bwdFillWithinNativeFunction,
}

// id assigns the target column with the corresponding value of the source
// column
func idNativeFunction(sources []field.FrArray) []field.FrArray {
	if len(sources) != 1 {
		panic("incorrect number of arguments")
	}
	// Clone source column (that's it)
	return []field.FrArray{sources[0].Clone()}
}

// interleaving constructs a single interleaved column from a give set of source
// columns.  The assumption is that the height of all columns is the same.
func interleaveNativeFunction(sources []field.FrArray) []field.FrArray {
	var (
		height     = sources[0].Len()
		bitwidth   = sources[0].BitWidth()
		multiplier = uint(len(sources))
	)
	// Sanity check column heights
	for _, src := range sources {
		if src.Len() != height {
			panic("inconsistent column height for interleaving")
		} else if src.BitWidth() != bitwidth {
			panic("inconsistent column bitwidth for interleaving")
		}
	}
	// Construct interleaved column
	target := field.NewFrArray(height*multiplier, bitwidth)
	//
	for i := range multiplier {
		src := sources[i]
		//
		for j := range height {
			row := (j * multiplier) + i
			target.Set(row, src.Get(j))
		}
	}
	// Done
	return []field.FrArray{target}
}

// filter assigns the target column with the corresponding value of the source
// column *when* a given selector column is non-zero.  Otherwise, the target
// column remains zero at the given position.
func filterNativeFunction(sources []field.FrArray) []field.FrArray {
	if len(sources) != 2 {
		panic("incorrect number of arguments")
	}

	var (
		// Extract input column info
		src_col = sources[0]
		sel_col = sources[1]
		// Clone source column
		data = field.NewFrArray(src_col.Len(), src_col.BitWidth())
	)
	//
	for i := uint(0); i < data.Len(); i++ {
		selector := sel_col.Get(i)
		// Check whether selctor non-zero
		if !selector.IsZero() {
			ith_value := src_col.Get(i)
			data.Set(i, ith_value)
		}
	}
	// Done
	return []field.FrArray{data}
}

// apply a key-value map conditionally.
func mapIfNativeFunction(sources []field.FrArray) []field.FrArray {
	n := len(sources) - 3
	if n%2 != 0 {
		panic(fmt.Sprintf("map-if expects 3 + 2*n columns (given %d)", len(sources)))
	}
	//
	n = n / 2
	// Setup what we need
	source_selector := sources[1+n]
	source_keys := make([]util.Array[fr.Element], n)
	source_value := sources[2+n+n]
	source_map := hash.NewMap[hash.BytesKey, fr.Element](source_value.Len())
	target_selector := sources[0]
	target_keys := make([]util.Array[fr.Element], n)
	target_value := field.NewFrArray(target_selector.Len(), source_value.BitWidth())
	// Initialise source / target keys
	for i := 0; i < n; i++ {
		target_keys[i] = sources[1+i]
		source_keys[i] = sources[2+n+i]
	}
	// Build source map
	for i := uint(0); i < source_value.Len(); i++ {
		ith_selector := source_selector.Get(i)
		if !ith_selector.IsZero() {
			ith_value := source_value.Get(i)
			ith_key := extractIthKey(i, source_keys)
			//
			if val, ok := source_map.Get(ith_key); ok && val.Cmp(&ith_value) != 0 {
				// Conflicting item already in map, so fail with useful error.
				ith_row := extractIthColumns(i, source_keys)
				lhs := fmt.Sprintf("%v=>%s", ith_row, ith_value.String())
				rhs := fmt.Sprintf("%v=>%s", ith_row, val.String())
				panic(fmt.Sprintf("conflicting values in source map (row %d): %s vs %s", i, lhs, rhs))
			} else if !ok {
				// Item not previously in map
				source_map.Insert(ith_key, ith_value)
			}
		}
	}
	// Construct target value column
	for i := uint(0); i < target_value.Len(); i++ {
		ith_selector := target_selector.Get(i)
		if !ith_selector.IsZero() {
			ith_key := extractIthKey(i, target_keys)
			//nolint:revive
			if val, ok := source_map.Get(ith_key); !ok {
				// Couldn't find key in source map, so fail with useful error.
				ith_row := extractIthColumns(i, target_keys)
				panic(fmt.Sprintf("target key (%v) missing from source map (row %d)", ith_row, i))
			} else {
				// Assign target value
				target_value.Set(i, val)
			}
		}
	}
	// Done
	return []field.FrArray{target_value}
}

func extractIthKey(index uint, cols []field.FrArray) hash.BytesKey {
	// Each fr.Element is 4 x 64bit words.
	bytes := make([]byte, 32*len(cols))
	// Slice provides an access window for writing
	slice := bytes
	// Evaluate each expression in turn
	for i := 0; i < len(cols); i++ {
		ith := cols[i].Get(index)
		// Copy over each element
		binary.BigEndian.PutUint64(slice, ith[0])
		binary.BigEndian.PutUint64(slice[8:], ith[1])
		binary.BigEndian.PutUint64(slice[16:], ith[2])
		binary.BigEndian.PutUint64(slice[24:], ith[3])
		// Move slice over
		slice = slice[32:]
	}
	// Done
	return hash.NewBytesKey(bytes)
}

// determines changes of a given set of columns within a given region.
func fwdChangesWithinNativeFunction(sources []field.FrArray) []field.FrArray {
	if len(sources) < 2 {
		panic("incorrect number of arguments")
	}
	// Useful constant
	one := fr.One()
	// Extract input column info
	selector_col := sources[0]
	source_cols := make([]util.Array[fr.Element], len(sources)-1)
	//
	for i := 1; i < len(sources); i++ {
		source_cols[i-1] = sources[i]
	}
	// Construct (binary) output column
	data := field.NewFrArray(selector_col.Len(), 1)
	// Set current value
	current := make([]fr.Element, len(source_cols))
	started := false
	//
	for i := uint(0); i < selector_col.Len(); i++ {
		ith_selector := selector_col.Get(i)
		// Check whether within region or not.
		if !ith_selector.IsZero() {
			//
			row := extractIthColumns(i, source_cols)
			// Trigger required?
			if !started || !slices.Equal(current, row) {
				started = true
				current = row
				//
				data.Set(i, one)
			}
		}
	}
	// Done
	return []field.FrArray{data}
}

func fwdUnchangedWithinNativeFunction(sources []field.FrArray) []field.FrArray {
	if len(sources) < 2 {
		panic("incorrect number of arguments")
	}
	// Useful constant
	one := fr.One()
	zero := fr.NewElement(0)
	// Extract input column info
	selector_col := sources[0]
	source_cols := make([]util.Array[fr.Element], len(sources)-1)
	//
	for i := 1; i < len(sources); i++ {
		source_cols[i-1] = sources[i]
	}
	// Construct (binary) output column
	data := field.NewFrArray(selector_col.Len(), 1)
	// Set current value
	current := make([]fr.Element, len(source_cols))
	started := false
	//
	for i := uint(0); i < selector_col.Len(); i++ {
		ith_selector := selector_col.Get(i)
		// Check whether within region or not.
		if !ith_selector.IsZero() {
			//
			row := extractIthColumns(i, source_cols)
			// Trigger required?
			if !started || !slices.Equal(current, row) {
				started = true
				current = row
				//
				data.Set(i, zero)
			} else {
				data.Set(i, one)
			}
		}
	}
	// Done
	return []field.FrArray{data}
}

// determines changes of a given set of columns within a given region.
func bwdChangesWithinNativeFunction(sources []field.FrArray) []field.FrArray {
	if len(sources) < 2 {
		panic("incorrect number of arguments")
	}
	// Useful constant
	one := fr.One()
	// Extract input column info
	selector_col := sources[0]
	source_cols := make([]util.Array[fr.Element], len(sources)-1)
	//
	for i := 1; i < len(sources); i++ {
		source_cols[i-1] = sources[i]
	}
	// Construct (binary) output column
	data := field.NewFrArray(selector_col.Len(), 1)
	// Set current value
	current := make([]fr.Element, len(source_cols))
	started := false
	//
	for i := selector_col.Len(); i > 0; i-- {
		ith_selector := selector_col.Get(i - 1)
		// Check whether within region or not.
		if !ith_selector.IsZero() {
			//
			row := extractIthColumns(i-1, source_cols)
			// Trigger required?
			if !started || !slices.Equal(current, row) {
				started = true
				current = row
				//
				data.Set(i-1, one)
			}
		}
	}
	// Done
	return []field.FrArray{data}
}

func fwdFillWithinNativeFunction(sources []field.FrArray) []field.FrArray {
	if len(sources) != 3 {
		panic("incorrect number of arguments")
	}
	// Extract input column info
	selector_col := sources[0]
	first_col := sources[1]
	source_col := sources[2]
	// Construct (binary) output column
	data := field.NewFrArray(source_col.Len(), source_col.BitWidth())
	// Set current value
	current := fr.NewElement(0)
	//
	for i := uint(0); i < selector_col.Len(); i++ {
		ith_selector := selector_col.Get(i)
		// Check whether within region or not.
		if !ith_selector.IsZero() {
			ith_first := first_col.Get(i)
			//
			if !ith_first.IsZero() {
				current = source_col.Get(i)
			}
			//
			data.Set(i, current)
		}
	}
	// Done
	return []field.FrArray{data}
}

func bwdFillWithinNativeFunction(sources []field.FrArray) []field.FrArray {
	if len(sources) != 3 {
		panic("incorrect number of arguments")
	}
	// Extract input column info
	selector_col := sources[0]
	first_col := sources[1]
	source_col := sources[2]
	// Construct (binary) output column
	data := field.NewFrArray(source_col.Len(), source_col.BitWidth())
	// Set current value
	current := fr.NewElement(0)
	//
	for i := selector_col.Len(); i > 0; i-- {
		ith_selector := selector_col.Get(i - 1)
		// Check whether within region or not.
		if !ith_selector.IsZero() {
			ith_first := first_col.Get(i - 1)
			//
			if !ith_first.IsZero() {
				current = source_col.Get(i - 1)
			}
			//
			data.Set(i-1, current)
		}
	}
	// Done
	return []field.FrArray{data}
}

func extractIthColumns(index uint, cols []util.Array[fr.Element]) []fr.Element {
	row := make([]fr.Element, len(cols))
	//
	for i := range row {
		row[i] = cols[i].Get(index)
	}
	//
	return row
}

// ============================================================================
// Encoding / Decoding
// ============================================================================

func init() {
	gob.Register(sc.Assignment(&Computation{}))
}
