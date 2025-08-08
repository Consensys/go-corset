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
func TraceLowering[F field.Element[F]](parallel bool, tf lt.TraceFile) (word.Pool[uint, F], []trace.RawColumn[F]) {
	var (
		stats = util.NewPerfStats()
		pool  word.Pool[uint, F]
		cols  []trace.RawColumn[F]
	)
	//
	if parallel {
		pool, cols = parallelTraceLowering[F](tf)
	} else {
		pool, cols = sequentialTraceLowering[F](tf)
	}
	//
	stats.Log("Trace lowering")
	//
	return pool, cols
}

func sequentialTraceLowering[F field.Element[F]](tf lt.TraceFile) (word.Pool[uint, F], []trace.RawColumn[F]) {
	var (
		pool           = word.NewHeapPool[F]()
		loweredColumns []trace.RawColumn[F]
	)
	//
	for _, ith := range tf.Columns {
		lowered := lowerRawColumn(pool, ith)
		loweredColumns = append(loweredColumns, lowered)
	}
	//
	return pool, loweredColumns
}

func parallelTraceLowering[F field.Element[F]](tf lt.TraceFile) (word.Pool[uint, F], []trace.RawColumn[F]) {
	var (
		pool           = word.NewHeapPool[F]()
		columns        = tf.Columns
		loweredColumns = make([]trace.RawColumn[F], len(columns))
		// Construct a communication channel split columns.
		c = make(chan util.Pair[int, trace.RawColumn[F]], len(columns))
	)
	// Split column concurrently
	for i, ith := range columns {
		go func(index int, column trace.BigEndianColumn) {
			// Send outcome back
			c <- util.NewPair(index, lowerRawColumn(pool, column))
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
	return pool, loweredColumns
}

// lowerRawColumn lowers a given raw column into a given field implementation.
func lowerRawColumn[W word.Word[W], F field.Element[F]](pool word.Pool[uint, F], column trace.RawColumn[W],
) trace.RawColumn[F] {
	//
	var (
		data = column.Data
		arr  = word.NewArray(data.Len(), data.BitWidth(), pool)
		buf  []byte
	)
	//
	for i := range data.Len() {
		// Write raw data into byte array.  This is safe because we know its
		// coming from an unencoded source.
		buf = data.Get(i).PutRawBytes(buf)
		// Initialise target from source bytes
		val := field.FromBigEndianBytes[F](buf)
		// Write into ith row of array being constructed.
		arr.Set(i, val)
	}
	//
	return trace.RawColumn[F]{
		Module: column.Module,
		Name:   column.Name,
		Data:   arr,
	}
}
