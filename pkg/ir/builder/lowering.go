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
	"github.com/consensys/gnark-crypto/ecc/bls12-377/fr"
	"github.com/consensys/go-corset/pkg/trace"
	"github.com/consensys/go-corset/pkg/util"
	"github.com/consensys/go-corset/pkg/util/field"
	"github.com/consensys/go-corset/pkg/util/word"
)

// TraceLowering simply converts columns from their current big endian word
// representation into the appropriate field representation without performing
// any splitting.  This is only required for traces which are "pre-expanded".
// Such traces typically arise in testing, etc.
func TraceLowering(parallel bool, rawCols []trace.RawColumn[word.BigEndian]) []trace.RawFrColumn {
	var (
		stats = util.NewPerfStats()
		cols  []trace.RawFrColumn
	)
	//
	if parallel {
		cols = parallelTraceLowering(rawCols)
	} else {
		cols = sequentialTraceLowering(rawCols)
	}
	//
	stats.Log("Trace lowering")
	//
	return cols
}

func sequentialTraceLowering(columns []trace.RawColumn[word.BigEndian]) []trace.RawFrColumn {
	var loweredColumns []trace.RawFrColumn
	//
	for _, ith := range columns {
		lowered := lowerRawColumn(ith)
		loweredColumns = append(loweredColumns, lowered)
	}
	//
	return loweredColumns
}

func parallelTraceLowering(columns []trace.RawColumn[word.BigEndian]) []trace.RawFrColumn {
	var (
		loweredColumns []trace.RawFrColumn = make([]trace.RawFrColumn, len(columns))
		// Construct a communication channel split columns.
		c = make(chan util.Pair[int, trace.RawFrColumn], len(columns))
	)
	// Split column concurrently
	for i, ith := range columns {
		go func(index int, column trace.BigEndianColumn) {
			// Send outcome back
			c <- util.NewPair(index, lowerRawColumn(column))
		}(i, ith)
	}
	// Collect results
	for range len(columns) {
		// Read from channel
		res := <-c
		// Assign split
		loweredColumns[res.Left] = res.Right
	}
	// Done
	return loweredColumns
}

// lowerRawColumn lowers a given raw column into a given field implementation.
func lowerRawColumn(column trace.RawColumn[word.BigEndian]) trace.RawFrColumn {
	var (
		data  = column.Data
		ndata = field.NewFrArray(data.Len(), data.BitWidth())
	)
	//
	for i := range data.Len() {
		var val fr.Element
		// Initial field element from big endian bytes.
		val.SetBytes(data.Get(i).Bytes())
		//
		ndata.Set(i, val)
	}
	//
	return trace.RawColumn[fr.Element]{
		Module: column.Module,
		Name:   column.Name,
		Data:   ndata,
	}
}
