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

	"github.com/consensys/gnark-crypto/ecc/bls12-377/fr"
	"github.com/consensys/go-corset/pkg/schema"
	"github.com/consensys/go-corset/pkg/schema/agnostic"
	"github.com/consensys/go-corset/pkg/trace"
	"github.com/consensys/go-corset/pkg/util"
	"github.com/consensys/go-corset/pkg/util/collection/array"
	"github.com/consensys/go-corset/pkg/util/collection/bit"
	"github.com/consensys/go-corset/pkg/util/field"
	"github.com/consensys/go-corset/pkg/util/word"
)

// TraceSplitting splits a given set of raw columns according to a given
// register mapping or, otherwise, simply lowers them.
func TraceSplitting(parallel bool, rawCols []trace.BigEndianColumn,
	mapping schema.RegisterMap) []trace.RawFrColumn {
	var (
		stats = util.NewPerfStats()
		cols  []trace.RawFrColumn
	)
	//
	if parallel {
		cols = parallelTraceSplitting(rawCols, mapping)
	} else {
		cols = sequentialTraceSplitting(rawCols, mapping)
	}
	//
	stats.Log("Trace splitting")
	//
	return cols
}

func sequentialTraceSplitting(columns []trace.BigEndianColumn, mapping schema.RegisterMap) []trace.RawFrColumn {
	var splitColumns []trace.RawFrColumn
	//
	for _, ith := range columns {
		split := splitRawColumn(ith, mapping)
		splitColumns = append(splitColumns, split...)
	}
	//
	return splitColumns
}

func parallelTraceSplitting(columns []trace.BigEndianColumn, mapping schema.RegisterMap) []trace.RawFrColumn {
	var (
		splits [][]trace.RawFrColumn = make([][]trace.RawFrColumn, len(columns))
		// Construct a communication channel split columns.
		c = make(chan util.Pair[int, []trace.RawFrColumn], len(columns))
		//
		total int
	)
	// Split column concurrently
	for i, ith := range columns {
		go func(index int, column trace.BigEndianColumn, mapping schema.RegisterMap) {
			// Send outcome back
			c <- util.NewPair(index, splitRawColumn(column, mapping))
		}(i, ith, mapping)
	}
	// Collect results
	for range len(splits) {
		// Read from channel
		res := <-c
		// Assign split
		splits[res.Left] = res.Right
		//
		total += len(res.Right)
	}
	// Flatten split
	return flatten(total, splits)
}

func flatten(total int, splits [][]trace.RawFrColumn) []trace.RawFrColumn {
	var (
		columns = make([]trace.RawFrColumn, total)
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
func splitRawColumn(column trace.RawColumn[word.BigEndian], mapping schema.RegisterMap) []trace.RawFrColumn {
	var (
		height = column.Data.Len()
		// Access mapping for enclosing module
		modmap = mapping.ModuleOf(column.Module)
		// Determine register id for this column
		reg = modmap.RegisterOf(column.Name)
		// Determine limbs of this register
		limbIds = modmap.LimbIds(reg)
	)
	// Check whether any work actually required
	if len(limbIds) == 1 {
		// No, this register was not split into any limbs.  Therefore, no need
		// to split the column into any limbs.
		return []trace.RawColumn[fr.Element]{lowerRawColumn(column)}
	}
	// Yes, must split into two or more limbs of given widths.
	limbWidths := agnostic.WidthsOfLimbs(modmap, modmap.LimbIds(reg))
	// Determine limbs of this register
	limbs := agnostic.LimbsOf(modmap, limbIds)
	// Construct temporary place holder for new array data.
	arrays := make([]field.FrArray, len(limbIds))
	//
	for i, limb := range limbs {
		arrays[i] = field.NewFrArray(height, limb.Width)
	}
	// Deconstruct all data
	for i := range height {
		// Extract ith data
		ith := column.Data.Get(i)
		// Assign split components
		for j, v := range splitFieldElement(ith, limbWidths) {
			arrays[j].Set(i, v)
		}
	}
	// Construct final columns
	columns := make([]trace.RawFrColumn, len(limbIds))
	// Construct final columns
	for i, limb := range limbs {
		columns[i] = trace.RawFrColumn{
			Module: column.Module,
			Name:   limb.Name,
			Data:   arrays[i]}
	}
	// Done
	return columns
}

// split a given field element into a given set of limbs, where the least
// significant comes first.  NOTE: this is really a temporary function which
// should be eliminated when RawColumn is moved away from fr.Element.
func splitFieldElement(val word.BigEndian, widths []uint) []fr.Element {
	var (
		n = len(widths)
		//
		bitwidth = sum(widths...)
		// Determine bytewidth
		bytewidth = word.ByteWidth(bitwidth)
		// Extract bytes whilst ensuring they are in little endian form, and
		// that they match the expected bitwidth.
		bytes = reverseAndPad(val.Bytes(), bytewidth)
		//
		bits     = bit.NewReader(bytes[:])
		buf      [32]byte
		elements = make([]fr.Element, n)
	)
	// read actual bits
	for i, w := range widths {
		var ith fr.Element
		// Read bits
		m := bits.ReadInto(w, buf[:])
		// Convert back to big endian
		array.ReverseInPlace(buf[:m])
		// Done
		ith.SetBytes(buf[:m])
		elements[i] = ith
	}
	//
	return elements
}

func reverseAndPad(bytes []byte, n uint) []byte {
	// Make sure bytes is both padded and cloned.
	switch {
	case n > uint(len(bytes)):
		bytes = array.FrontPad(bytes, n, 0)
	case n == uint(len(bytes)):
		bytes = slices.Clone(bytes)
	case n < uint(len(bytes)):
		panic(fmt.Sprintf("have %d bytes, expected at most %d", len(bytes), n))
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
