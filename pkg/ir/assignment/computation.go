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
	"encoding/gob"
	"fmt"

	"github.com/consensys/go-corset/pkg/schema"
	sc "github.com/consensys/go-corset/pkg/schema"
	tr "github.com/consensys/go-corset/pkg/trace"
	"github.com/consensys/go-corset/pkg/util"
	"github.com/consensys/go-corset/pkg/util/collection/array"
	"github.com/consensys/go-corset/pkg/util/field"
	"github.com/consensys/go-corset/pkg/util/field/bls12_377"
	"github.com/consensys/go-corset/pkg/util/source/sexp"
	"github.com/consensys/go-corset/pkg/util/word"
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
func (p *Computation) Compute(trace tr.Trace[bls12_377.Element], schema sc.AnySchema) ([]tr.ArrayColumn[bls12_377.Element], error) {
	// Identify Computation
	fn := findNative[bls12_377.Element](p.Function)
	// Go!
	return computeNative(p.Sources, p.Targets, fn, trace, schema), nil
}

// Consistent performs some simple checks that the given schema is consistent.
// This provides a double check of certain key properties, such as that
// registers used for assignments are large enough, etc.
func (p *Computation) Consistent(_ sc.AnySchema) []error {
	// NOTE: this is where we could (in principle) check the type of the
	// function being defined to ensure it is, for example, typed correctly.
	return nil
}

// RegistersExpanded identifies registers expanded by this assignment.
func (p *Computation) RegistersExpanded() []sc.RegisterRef {
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

// Subdivide implementation for the FieldAgnostic interface.
func (p *Computation) Subdivide(mapping schema.LimbsMap) sc.Assignment {
	return p
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
		ithName := ith.QualifiedName(module)
		sources.Append(sexp.NewSymbol(ithName))
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

// NativeComputation defines the type of a native function for computing a given
// set of output columns as a function of a given set of input columns.
type NativeComputation[F field.Element[F]] func([]array.Array[F], word.Pool[uint, F]) []array.Array[F]

func computeNative[F field.Element[F]](sources []sc.RegisterRef, targets []sc.RegisterRef, fn NativeComputation[F],
	trace tr.Trace[F], schema sc.AnySchema) []tr.ArrayColumn[F] {
	// Read inputs
	inputs := ReadRegisters(trace, sources...)
	// Read inputs
	for i, ref := range sources {
		mid, rid := ref.Module(), ref.Register().Unwrap()
		inputs[i] = trace.Module(mid).Column(rid).Data()
	}
	// Apply native function
	data := fn(inputs, trace.Pool())
	// Write outputs
	return WriteRegisters(schema, targets, data)
}

// ============================================================================
// Native Function Definitions
// ============================================================================

func findNative[F field.Element[F]](name string) NativeComputation[F] {
	switch name {
	case "id":
		return idNativeFunction[F]
	// case "interleave":
	// 	return interleaveNativeFunction
	// case "filter":
	// 	return filterNativeFunction
	// case "map-if":
	// 	return mapIfNativeFunction
	// case "fwd-changes-within":
	// 	return fwdChangesWithinNativeFunction
	// case "fwd-unchanged-within":
	// 	return fwdUnchangedWithinNativeFunction
	// case "bwd-changes-within":
	// 	return bwdChangesWithinNativeFunction
	// case "fwd-fill-within":
	// 	return fwdFillWithinNativeFunction
	// case "bwd-fill-within":
	// 	return bwdFillWithinNativeFunction
	default:
		panic(fmt.Sprintf("unknown native function: %s", name))
	}
}

// id assigns the target column with the corresponding value of the source
// column
func idNativeFunction[F field.Element[F]](sources []array.Array[F], pool word.Pool[uint, F]) []array.Array[F] {
	if len(sources) != 1 {
		panic("incorrect number of arguments")
	}
	// Clone source column (that's it)
	return []array.Array[F]{sources[0]}
}

// // interleaving constructs a single interleaved column from a give set of source
// // columns.  The assumption is that the height of all columns is the same.
// func interleaveNativeFunction[F field.Element[F]](sources []field.FrArray, pool field.Pool) []field.FrArray {
// 	var (
// 		height     = sources[0].Len()
// 		bitwidth   = sources[0].BitWidth()
// 		multiplier = uint(len(sources))
// 	)
// 	// Sanity check column heights
// 	for _, src := range sources {
// 		if src.Len() != height {
// 			panic("inconsistent column height for interleaving")
// 		} else if src.BitWidth() != bitwidth {
// 			panic("inconsistent column bitwidth for interleaving")
// 		}
// 	}
// 	// Construct interleaved column
// 	target := word.NewArray(height*multiplier, bitwidth, pool)
// 	//
// 	for i := range multiplier {
// 		src := sources[i]
// 		//
// 		for j := range height {
// 			row := (j * multiplier) + i
// 			target.Set(row, src.Get(j))
// 		}
// 	}
// 	// Done
// 	return []field.FrArray{target}
// }

// // filter assigns the target column with the corresponding value of the source
// // column *when* a given selector column is non-zero.  Otherwise, the target
// // column remains zero at the given position.
// func filterNativeFunction[F field.Element[F]](sources []field.FrArray, pool field.Pool) []field.FrArray {
// 	if len(sources) != 2 {
// 		panic("incorrect number of arguments")
// 	}

// 	var (
// 		// Extract input column info
// 		srcCol = sources[0]
// 		selCol = sources[1]
// 		// Clone source column
// 		data = word.NewArray(srcCol.Len(), srcCol.BitWidth(), pool)
// 	)
// 	//
// 	for i := uint(0); i < data.Len(); i++ {
// 		selector := selCol.Get(i)
// 		// Check whether selctor non-zero
// 		if !selector.IsZero() {
// 			ithValue := srcCol.Get(i)
// 			data.Set(i, ithValue)
// 		}
// 	}
// 	// Done
// 	return []field.FrArray{data}
// }

// // apply a key-value map conditionally.
// func mapIfNativeFunction[F field.Element[F]](sources []field.FrArray, pool field.Pool) []field.FrArray {
// 	n := len(sources) - 3
// 	if n%2 != 0 {
// 		panic(fmt.Sprintf("map-if expects 3 + 2*n columns (given %d)", len(sources)))
// 	}
// 	//
// 	n = n / 2
// 	// Setup what we need
// 	sourceSelector := sources[1+n]
// 	sourceKeys := make([]array.MutArray[fr.Element], n)
// 	sourceValue := sources[2+n+n]
// 	sourceMap := hash.NewMap[hash.BytesKey, fr.Element](sourceValue.Len())
// 	targetSelector := sources[0]
// 	targetKeys := make([]array.MutArray[fr.Element], n)
// 	targetValue := word.NewArray(targetSelector.Len(), sourceValue.BitWidth(), pool)
// 	// Initialise source / target keys
// 	for i := 0; i < n; i++ {
// 		targetKeys[i] = sources[1+i]
// 		sourceKeys[i] = sources[2+n+i]
// 	}
// 	// Build source map
// 	for i := uint(0); i < sourceValue.Len(); i++ {
// 		ithSelector := sourceSelector.Get(i)
// 		if !ithSelector.IsZero() {
// 			ithValue := sourceValue.Get(i)
// 			ithKey := extractIthKey(i, sourceKeys)
// 			//
// 			if val, ok := sourceMap.Get(ithKey); ok && val.Cmp(&ithValue) != 0 {
// 				// Conflicting item already in map, so fail with useful error.
// 				ithRow := extractIthColumns(i, sourceKeys)
// 				lhs := fmt.Sprintf("%v=>%s", ithRow, ithValue.String())
// 				rhs := fmt.Sprintf("%v=>%s", ithRow, val.String())
// 				panic(fmt.Sprintf("conflicting values in source map (row %d): %s vs %s", i, lhs, rhs))
// 			} else if !ok {
// 				// Item not previously in map
// 				sourceMap.Insert(ithKey, ithValue)
// 			}
// 		}
// 	}
// 	// Construct target value column
// 	for i := uint(0); i < targetValue.Len(); i++ {
// 		ithSelector := targetSelector.Get(i)
// 		if !ithSelector.IsZero() {
// 			ithKey := extractIthKey(i, targetKeys)
// 			//nolint:revive
// 			if val, ok := sourceMap.Get(ithKey); !ok {
// 				// Couldn't find key in source map, so fail with useful error.
// 				ith_row := extractIthColumns(i, targetKeys)
// 				panic(fmt.Sprintf("target key (%v) missing from source map (row %d)", ith_row, i))
// 			} else {
// 				// Assign target value
// 				targetValue.Set(i, val)
// 			}
// 		}
// 	}
// 	// Done
// 	return []field.FrArray{targetValue}
// }

// func extractIthKey[F field.Element[F]](index uint, cols []field.FrArray, pool field.Pool) hash.BytesKey {
// 	// Each fr.Element is 4 x 64bit words.
// 	bytes := make([]byte, 32*len(cols))
// 	// Slice provides an access window for writing
// 	slice := bytes
// 	// Evaluate each expression in turn
// 	for i := 0; i < len(cols); i++ {
// 		ith := cols[i].Get(index)
// 		// Copy over each element
// 		binary.BigEndian.PutUint64(slice, ith[0])
// 		binary.BigEndian.PutUint64(slice[8:], ith[1])
// 		binary.BigEndian.PutUint64(slice[16:], ith[2])
// 		binary.BigEndian.PutUint64(slice[24:], ith[3])
// 		// Move slice over
// 		slice = slice[32:]
// 	}
// 	// Done
// 	return hash.NewBytesKey(bytes)
// }

// // determines changes of a given set of columns within a given region.
// func fwdChangesWithinNativeFunction[F field.Element[F]](sources []field.FrArray, pool field.Pool) []field.FrArray {
// 	if len(sources) < 2 {
// 		panic("incorrect number of arguments")
// 	}
// 	// Useful constant
// 	one := fr.One()
// 	// Extract input column info
// 	selectorCol := sources[0]
// 	sourceCols := make([]array.MutArray[fr.Element], len(sources)-1)
// 	//
// 	for i := 1; i < len(sources); i++ {
// 		sourceCols[i-1] = sources[i]
// 	}
// 	// Construct (binary) output column
// 	data := word.NewArray(selectorCol.Len(), 1, pool)
// 	// Set current value
// 	current := make([]fr.Element, len(sourceCols))
// 	started := false
// 	//
// 	for i := uint(0); i < selectorCol.Len(); i++ {
// 		ithSelector := selectorCol.Get(i)
// 		// Check whether within region or not.
// 		if !ithSelector.IsZero() {
// 			//
// 			row := extractIthColumns(i, sourceCols)
// 			// Trigger required?
// 			if !started || !slices.Equal(current, row) {
// 				started = true
// 				current = row
// 				//
// 				data.Set(i, one)
// 			}
// 		}
// 	}
// 	// Done
// 	return []field.FrArray{data}
// }

// func fwdUnchangedWithinNativeFunction[F field.Element[F]](sources []field.FrArray, pool field.Pool) []field.FrArray {
// 	if len(sources) < 2 {
// 		panic("incorrect number of arguments")
// 	}
// 	// Useful constant
// 	one := fr.One()
// 	zero := fr.NewElement(0)
// 	// Extract input column info
// 	selectorCol := sources[0]
// 	sourceCols := make([]array.MutArray[fr.Element], len(sources)-1)
// 	//
// 	for i := 1; i < len(sources); i++ {
// 		sourceCols[i-1] = sources[i]
// 	}
// 	// Construct (binary) output column
// 	data := word.NewArray(selectorCol.Len(), 1, pool)
// 	// Set current value
// 	current := make([]fr.Element, len(sourceCols))
// 	started := false
// 	//
// 	for i := uint(0); i < selectorCol.Len(); i++ {
// 		ithSelector := selectorCol.Get(i)
// 		// Check whether within region or not.
// 		if !ithSelector.IsZero() {
// 			//
// 			row := extractIthColumns(i, sourceCols)
// 			// Trigger required?
// 			if !started || !slices.Equal(current, row) {
// 				started = true
// 				current = row
// 				//
// 				data.Set(i, zero)
// 			} else {
// 				data.Set(i, one)
// 			}
// 		}
// 	}
// 	// Done
// 	return []field.FrArray{data}
// }

// // determines changes of a given set of columns within a given region.
// func bwdChangesWithinNativeFunction[F field.Element[F]](sources []field.FrArray, pool field.Pool) []field.FrArray {
// 	if len(sources) < 2 {
// 		panic("incorrect number of arguments")
// 	}
// 	// Useful constant
// 	one := fr.One()
// 	// Extract input column info
// 	selectorCol := sources[0]
// 	sourceCols := make([]array.MutArray[fr.Element], len(sources)-1)
// 	//
// 	for i := 1; i < len(sources); i++ {
// 		sourceCols[i-1] = sources[i]
// 	}
// 	// Construct (binary) output column
// 	data := field.NewFrArray(selectorCol.Len(), 1)
// 	// Set current value
// 	current := make([]fr.Element, len(sourceCols))
// 	started := false
// 	//
// 	for i := selectorCol.Len(); i > 0; i-- {
// 		ithSelector := selectorCol.Get(i - 1)
// 		// Check whether within region or not.
// 		if !ithSelector.IsZero() {
// 			//
// 			row := extractIthColumns(i-1, sourceCols)
// 			// Trigger required?
// 			if !started || !slices.Equal(current, row) {
// 				started = true
// 				current = row
// 				//
// 				data.Set(i-1, one)
// 			}
// 		}
// 	}
// 	// Done
// 	return []field.FrArray{data}
// }

// func fwdFillWithinNativeFunction[F field.Element[F]](sources []field.FrArray, pool field.Pool) []field.FrArray {
// 	if len(sources) != 3 {
// 		panic("incorrect number of arguments")
// 	}
// 	// Extract input column info
// 	selectorCol := sources[0]
// 	firstCol := sources[1]
// 	sourceCol := sources[2]
// 	// Construct (binary) output column
// 	data := field.NewFrArray(sourceCol.Len(), sourceCol.BitWidth())
// 	// Set current value
// 	current := fr.NewElement(0)
// 	//
// 	for i := uint(0); i < selectorCol.Len(); i++ {
// 		ithSelector := selectorCol.Get(i)
// 		// Check whether within region or not.
// 		if !ithSelector.IsZero() {
// 			ithFirst := firstCol.Get(i)
// 			//
// 			if !ithFirst.IsZero() {
// 				current = sourceCol.Get(i)
// 			}
// 			//
// 			data.Set(i, current)
// 		}
// 	}
// 	// Done
// 	return []field.FrArray{data}
// }

// func bwdFillWithinNativeFunction[F field.Element[F]](sources []field.FrArray, pool field.Pool) []field.FrArray {
// 	if len(sources) != 3 {
// 		panic("incorrect number of arguments")
// 	}
// 	// Extract input column info
// 	selectorCol := sources[0]
// 	firstCol := sources[1]
// 	sourceCol := sources[2]
// 	// Construct (binary) output column
// 	data := field.NewFrArray(sourceCol.Len(), sourceCol.BitWidth())
// 	// Set current value
// 	current := fr.NewElement(0)
// 	//
// 	for i := selectorCol.Len(); i > 0; i-- {
// 		ithSelector := selectorCol.Get(i - 1)
// 		// Check whether within region or not.
// 		if !ithSelector.IsZero() {
// 			ithFirst := firstCol.Get(i - 1)
// 			//
// 			if !ithFirst.IsZero() {
// 				current = sourceCol.Get(i - 1)
// 			}
// 			//
// 			data.Set(i-1, current)
// 		}
// 	}
// 	// Done
// 	return []field.FrArray{data}
// }

// func extractIthColumns[F field.Element[F]](index uint, cols []array.MutArray[fr.Element]) []fr.Element {
// 	row := make([]fr.Element, len(cols))
// 	//
// 	for i := range row {
// 		row[i] = cols[i].Get(index)
// 	}
// 	//
// 	return row
// }

// ============================================================================
// Encoding / Decoding
// ============================================================================

func init() {
	gob.Register(sc.Assignment(&Computation{}))
}
