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
	"github.com/consensys/go-corset/pkg/trace"
	"github.com/consensys/go-corset/pkg/trace/lt"
	"github.com/consensys/go-corset/pkg/util"
	"github.com/consensys/go-corset/pkg/util/field"
	"github.com/consensys/go-corset/pkg/util/word"
)

// TraceLowering simply converts columns from their current big endian word
// representation into the appropriate field representation without performing
// any splitting.  This is only required for traces which are "pre-expanded".
// Such traces typically arise in testing, etc.
func TraceLowering[F field.Element[F]](parallel bool, tf lt.TraceFile) (word.ArrayBuilder[F], []trace.RawColumn[F]) {
	var (
		stats   = util.NewPerfStats()
		builder word.ArrayBuilder[F]
		cols    []trace.RawColumn[F]
	)
	//
	if parallel {
		builder, cols = parallelTraceLowering[F](tf.Columns)
	} else {
		builder, cols = sequentialTraceLowering[F](tf.Columns)
	}
	//
	stats.Log("Trace lowering")
	//
	return builder, cols
}

func sequentialTraceLowering[F field.Element[F]](columns []trace.RawColumn[word.BigEndian]) (word.ArrayBuilder[F],
	[]trace.RawColumn[F]) {
	//
	var (
		loweredColumns []trace.RawColumn[F]
		builder        = word.NewStaticArrayBuilder[F]()
	)
	//
	for _, ith := range columns {
		lowered := lowerRawColumn(ith, builder)
		loweredColumns = append(loweredColumns, lowered)
	}
	//
	return builder, loweredColumns
}

func parallelTraceLowering[F field.Element[F]](columns []trace.RawColumn[word.BigEndian]) (word.ArrayBuilder[F],
	[]trace.RawColumn[F]) {
	//
	var (
		loweredColumns []trace.RawColumn[F] = make([]trace.RawColumn[F], len(columns))
		// Construct new pool
		builder = word.NewStaticArrayBuilder[F]()
		// Construct a communication channel split columns.
		c = make(chan util.Pair[int, trace.RawColumn[F]], len(columns))
	)
	// Split column concurrently
	for i, ith := range columns {
		go func(index int, column trace.RawColumn[word.BigEndian]) {
			// Send outcome back
			c <- util.NewPair(index, lowerRawColumn(column, builder))
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
	return builder, loweredColumns
}

// lowerRawColumn lowers a given raw column into a given field implementation.
func lowerRawColumn[F field.Element[F]](column trace.RawColumn[word.BigEndian], builder word.ArrayBuilder[F],
) trace.RawColumn[F] {
	var (
		data  = column.Data
		ndata = builder.NewArray(data.Len(), data.BitWidth())
	)
	//
	for i := range data.Len() {
		var val F
		// Initial word from big endian bytes.
		val = val.SetBytes(data.Get(i).Bytes())
		//
		ndata.Set(i, val)
	}
	//
	return trace.RawColumn[F]{
		Module: column.Module,
		Name:   column.Name,
		Data:   ndata,
	}
}
