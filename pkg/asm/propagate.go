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
package asm

import (
	"github.com/consensys/go-corset/pkg/asm/io"
	"github.com/consensys/go-corset/pkg/trace"
	"github.com/consensys/go-corset/pkg/trace/lt"
	"github.com/consensys/go-corset/pkg/util/field"
	"github.com/consensys/go-corset/pkg/util/word"
)

// RawColumn provides a convenient alias
type RawColumn = trace.RawColumn[word.BigEndian]

// PropagateAll propagates secondary (i.e. derivative) function instances
// throughout one or more traces.  NOTES:
//
// Parallelism?
// Validation?
// Batch size?
// Recursion limit (to prevent infinite loops)
func PropagateAll[F field.Element[F], T io.Instruction[T]](p MixedProgram[F, T], ts []lt.TraceFile) []lt.TraceFile {
	var ntraces = make([]lt.TraceFile, len(ts))
	//
	for i, trace := range ts {
		ntraces[i] = Propagate(p.program, trace)
	}
	//
	return ntraces
}

// Propagate secondary (i.e. derivative) function instances throughout a trace.
// For example, suppose two functions "f(x)" and "g(y)", where f(x) calls g(y).
// Then, consider a trace which contains exactly one instance "f(a)=b". Since
// f(x) calls g(y) we may (depending on the exact implementation of f(x) and the
// parameter given) require a secondary instance, say "g(y)=c", to be added to
// make the trace complete with respect to the original instance. Trace
// propagation is about figuring out what secondary instances are required, and
// adding them to the trace.
func Propagate[T io.Instruction[T]](program io.Program[T], trace lt.TraceFile) lt.TraceFile {
	// Construct suitable executior for the given program
	var (
		executor = io.NewExecutor(program)
		// Clone heap in trace file, since will mutate this.
		heap = trace.Heap.Clone()
	)
	// Write seed instances
	writeInstances(trace.Columns, executor)
	// Read out generated instances
	columns := readInstances(&heap, executor)
	//
	return lt.NewTraceFile(trace.Header.MetaData, heap, columns)
}

// WriteInstances writes all of the instances defined in the given trace columns
// into the executor which, in turn, forces it to execute the relevant
// functions, and functions they call, etc.
func writeInstances[T io.Instruction[T]](columns []RawColumn, executor *io.Executor[T]) []RawColumn {
	panic("todo")
}

// ReadInstances simply traverses all internal states generated within the
// executor and converts them back into raw columns.
func readInstances[T io.Instruction[T]](heap *lt.WordHeap, executor *io.Executor[T]) []RawColumn {
	panic("todo")
}
