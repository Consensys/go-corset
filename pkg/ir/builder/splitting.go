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
package builder

import (
	"fmt"
	"slices"

	"github.com/consensys/go-corset/pkg/schema"
	"github.com/consensys/go-corset/pkg/schema/agnostic"
	"github.com/consensys/go-corset/pkg/trace"
	"github.com/consensys/go-corset/pkg/trace/lt"
	"github.com/consensys/go-corset/pkg/util"
	"github.com/consensys/go-corset/pkg/util/collection/array"
	"github.com/consensys/go-corset/pkg/util/collection/bit"
	"github.com/consensys/go-corset/pkg/util/field"
	"github.com/consensys/go-corset/pkg/util/word"
)

// TraceSplitting splits a given set of raw columns according to a given
// register mapping or, otherwise, simply lowers them.
func TraceSplitting[F field.Element[F]](parallel bool, tf lt.TraceFile, mapping schema.LimbsMap) (word.ArrayBuilder[F],
	[]trace.RawColumn[F], []error) {
	//
	var (
		stats   = util.NewPerfStats()
		builder word.ArrayBuilder[F]
		cols    []trace.RawColumn[F]
		err     []error
	)
	//
	if parallel {
		builder, cols, err = parallelTraceSplitting[F](tf, mapping)
	} else {
		builder, cols, err = sequentialTraceSplitting[F](tf, mapping)
	}
	//
	stats.Log("Trace splitting")
	//
	return builder, cols, err
}

func sequentialTraceSplitting[F field.Element[F]](tf lt.TraceFile, gmap schema.LimbsMap) (word.ArrayBuilder[F],
	[]trace.RawColumn[F], []error) {
	//
	var (
		splitColumns []trace.RawColumn[F]
		// Allocate fresh array builder
		builder = word.NewStaticArrayBuilder[F]()
		errors  []error
	)
	//
	for _, ith := range tf.Columns {
		split, errs := splitRawColumn(ith, builder, gmap)
		splitColumns = append(splitColumns, split...)
		errors = append(errors, errs...)
	}
	//
	return builder, splitColumns, errors
}

func parallelTraceSplitting[F field.Element[F]](tf lt.TraceFile, mapping schema.LimbsMap) (word.ArrayBuilder[F],
	[]trace.RawColumn[F], []error) {
	//
	var (
		splits = make([][]trace.RawColumn[F], len(tf.Columns))
		// Allocate fresh array builder
		builder = word.NewStaticArrayBuilder[F]()
		// Construct a communication channel split columns.
		c = make(chan splitResult[F], len(tf.Columns))
		//
		errors []error
		//
		total int
	)
	// Split column concurrently
	for i, ith := range tf.Columns {
		go func(index int, column trace.RawColumn[word.BigEndian], mapping schema.LimbsMap) {
			// Send outcome back
			data, errors := splitRawColumn(column, builder, mapping)
			c <- splitResult[F]{index, data, errors}
		}(i, ith, mapping)
	}
	// Collect results
	for range len(splits) {
		// Read from channel
		res := <-c
		// Assign split
		splits[res.id] = res.data
		//
		total += len(res.data)
		errors = append(errors, res.errors...)
	}
	// Flatten split
	return builder, flatten(total, splits), errors
}

func flatten[W any](total int, splits [][]trace.RawColumn[W]) []trace.RawColumn[W] {
	var (
		columns = make([]trace.RawColumn[W], total)
		index   = 0
	)
	// Flattern all columns
	for _, ith := range splits {
		for _, jth := range ith {
			columns[index] = jth
			index++
		}
	}
	//
	return columns
}

// SplitRawColumn splits a given raw column using the given register mapping.
func splitRawColumn[F field.Element[F]](col trace.RawColumn[word.BigEndian], builder word.ArrayBuilder[F],
	mapping schema.LimbsMap) ([]trace.RawColumn[F], []error) {
	//
	var (
		height = col.Data.Len()
		// Access mapping for enclosing module
		modmap = mapping.ModuleOf(col.Module)
		//
		reg, regExists = modmap.HasRegister(col.Name)
	)
	// Check whether register is known
	if !regExists {
		// Unknown register --- this is an error
		return nil, []error{fmt.Errorf("unknown register \"%s\"", col.Name)}
	}
	// Determine register id for this column (we can assume it exists)
	limbIds := modmap.LimbIds(reg)
	//
	if len(limbIds) == 1 {
		// No, this register was not split into any limbs.  Therefore, no need
		// to split the column into any limbs.
		return []trace.RawColumn[F]{lowerRawColumn(col, builder)}, nil
	}
	// Yes, must split into two or more limbs of given widths.
	limbWidths := agnostic.WidthsOfLimbs(modmap, modmap.LimbIds(reg))
	// Determine limbs of this register
	limbs := agnostic.LimbsOf(modmap, limbIds)
	// Construct temporary place holder for new array data.
	arrays := make([]array.MutArray[F], len(limbIds))
	//
	for i, limb := range limbs {
		arrays[i] = builder.NewArray(height, limb.Width)
	}
	// Deconstruct all data
	for i := range height {
		// Extract ith data
		ith := col.Data.Get(i)
		// Assign split components
		setSplitWord(ith, i, arrays, limbWidths)
	}
	// Construct final columns
	columns := make([]trace.RawColumn[F], len(limbIds))
	// Construct final columns
	for i, limb := range limbs {
		columns[i] = trace.RawColumn[F]{
			Module: col.Module,
			Name:   limb.Name,
			Data:   arrays[i]}
	}
	// Done
	return columns, nil
}

// split a given field element into a given set of limbs, where the least
// significant comes first.  NOTE: this is really a temporary function which
// should be eliminated when RawColumn is moved away from fr.Element.
func setSplitWord[F field.Element[F]](val word.BigEndian, row uint, arrays []array.MutArray[F], widths []uint) {
	var (
		bitwidth = sum(widths...)
		// Determine bytewidth
		bytewidth = word.ByteWidth(bitwidth)
		// Extract bytes whilst ensuring they are in little endian form, and
		// that they match the expected bitwidth.
		bytes = padAndReverse(val.Bytes(), bytewidth)
		//
		bits = bit.NewReader(bytes[:])
		buf  [32]byte
	)
	// read actual bits
	for i, w := range widths {
		// Read bits
		m := bits.ReadInto(w, buf[:])
		// Convert back to big endian
		array.ReverseInPlace(buf[:m])
		// Done
		arrays[i].Set(row, field.FromBigEndianBytes[F](buf[:m]))
	}
}

func padAndReverse(bytes []byte, n uint) []byte {
	// Make sure bytes is both padded and cloned.
	switch {
	case n > uint(len(bytes)):
		bytes = array.FrontPad(bytes, n, 0)
	default:
		bytes = slices.Clone(bytes)
	}
	// In place reversal
	array.ReverseInPlace(bytes)
	//
	return bytes
}

func sum(vals ...uint) uint {
	val := uint(0)
	//
	for _, v := range vals {
		val += v
	}
	//
	return val
}

// SplitResult is returned by worker threads during parallel trace splitting.
type splitResult[W any] struct {
	id     int
	data   []trace.RawColumn[W]
	errors []error
}
