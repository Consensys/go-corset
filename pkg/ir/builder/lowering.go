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
	"github.com/consensys/go-corset/pkg/util/word"
)

// TraceLowering simply converts columns from their current big endian word
// representation into the appropriate field representation without performing
// any splitting.  This is only required for traces which are "pre-expanded".
// Such traces typically arise in testing, etc.
func TraceLowering[T word.Word[T]](parallel bool, tf lt.TraceFile) (word.Pool[uint, T], []trace.RawColumn[T]) {
	var (
		stats = util.NewPerfStats()
		pool  word.Pool[uint, T]
		cols  []trace.RawColumn[T]
	)
	//
	if parallel {
		pool, cols = parallelTraceLowering[T](tf)
	} else {
		pool, cols = sequentialTraceLowering[T](tf)
	}
	//
	stats.Log("Trace lowering")
	//
	return pool, cols
}

func sequentialTraceLowering[T word.Word[T]](tf lt.TraceFile) (word.Pool[uint, T], []trace.RawColumn[T]) {
	var (
		pool           = word.NewHeapPool[T]()
		loweredColumns []trace.RawColumn[T]
	)
	//
	for _, ith := range tf.Columns {
		lowered := lowerRawColumn(pool, ith)
		loweredColumns = append(loweredColumns, lowered)
	}
	//
	return pool, loweredColumns
}

func parallelTraceLowering[T word.Word[T]](tf lt.TraceFile) (word.Pool[uint, T], []trace.RawColumn[T]) {
	var (
		pool           = word.NewHeapPool[T]()
		columns        = tf.Columns
		loweredColumns = make([]trace.RawColumn[T], len(columns))
		// Construct a communication channel split columns.
		c = make(chan util.Pair[int, trace.RawColumn[T]], len(columns))
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
func lowerRawColumn[F word.Word[F], T word.Word[T]](pool word.Pool[uint, T], column trace.RawColumn[F]) trace.RawColumn[T] {
	var (
		data = column.Data
		arr  = word.NewArray(data.Len(), data.BitWidth(), pool)
		buf  []byte
	)
	//
	for i := range data.Len() {
		var val T
		// Write data into byte array
		buf = data.Get(i).Put(buf)
		// Initialise target from source bytes
		val.Set(buf)
		// Write into ith row of array being constructed.
		arr.Set(i, val)
	}
	//
	return trace.RawColumn[T]{
		Module: column.Module,
		Name:   column.Name,
		Data:   arr,
	}
}
