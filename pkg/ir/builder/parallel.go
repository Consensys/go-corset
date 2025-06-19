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

	sc "github.com/consensys/go-corset/pkg/schema"
	"github.com/consensys/go-corset/pkg/trace"
	tr "github.com/consensys/go-corset/pkg/trace"
	"github.com/consensys/go-corset/pkg/util"
)

// ParallelTraceValidation validates that values held in trace columns match the
// expected type.  This is really a sanity check that the trace is not
// malformed.
func ParallelTraceValidation(schema sc.AnySchema, tr tr.Trace) []error {
	var (
		errors []error
		// Start timer
		stats = util.NewPerfStats()
		// Construct a communication channel for errors.
		c = make(chan error, 1024)
		// Number of columns to validate
		ntodo = uint(0)
	)
	// Check each module in turn
	for mid := uint(0); mid < tr.Width(); mid++ {
		var (
			scMod = schema.Module(mid)
			trMod = tr.Module(mid)
		)
		// Check each column within each module
		for i := uint(0); i < trMod.Width(); i++ {
			rid := sc.NewRegisterId(i)
			// Check elements
			go func(reg sc.Register, data trace.Column) {
				// Send outcome back
				c <- validateColumnBitWidth(reg.Width, data, scMod)
			}(scMod.Register(rid), trMod.Column(i))
			//
			ntodo++
		}
	}
	// Collect up all the results
	for i := uint(0); i < ntodo; i++ {
		// Read from channel
		if e := <-c; e != nil {
			errors = append(errors, e)
		}
	}
	// Log stats about this batch
	stats.Log("Validating trace")
	// Done
	return errors
}

// ParallelTraceExpansion performs trace expansion using concurrently executing
// jobs.  The chosen algorithm operates in waves, rather than using an
// continuous approach.  This is for two reasons: firstly, the latter would
// require locks that would slow down evaluation performance; secondly, the vast
// majority of jobs are run in the very first wave.
func ParallelTraceExpansion(batchsize uint, schema sc.AnySchema, trace *tr.ArrayTrace) error {
	var (
		batchNum = 0
		// Construct a communication channel for errors.
		ch = make(chan columnBatch, 1024)
		//
		expander = NewExpander(schema.Width(), schema.Assignments())
	)
	// Iterate until all assignments processed.
	for !expander.Done() {
		var (
			stats = util.NewPerfStats()
			batch = expander.Next(batchsize)
		)
		// Dispatch next batch of assignments.
		dispatchReadyAssignments(batch, schema, trace, ch)
		//
		batches := make([]columnBatch, len(batch))
		// Collect all the results
		for i := 0; i < len(batch); i++ {
			batches[i] = <-ch
			// Read from channel
			if batches[i].err != nil {
				// Fail immediately
				return batches[i].err
			}
		}
		// Once we get here, all go rountines are complete and we are sequential
		// again.
		for _, r := range batches {
			fillComputedColumns(r.targets, r.columns, trace)
		}
		// Log stats about this batch
		stats.Log(fmt.Sprintf("Expansion batch %d (remaining %d)", batch, expander.Count()))
		// Increment batch
		batchNum++
	}
	// Done
	return nil
}

// Dispatch the given set of assignments with results being fed back into the
// shared channel.
func dispatchReadyAssignments(batch []sc.Assignment, schema sc.AnySchema, trace *tr.ArrayTrace, ch chan columnBatch) {
	// Dispatch each assignment in the batch
	for _, ith := range batch {
		// Dispatch!
		go func(targets []sc.RegisterRef) {
			cols, err := ith.Compute(trace, schema)
			// Send outcome back
			ch <- columnBatch{targets, cols, err}
		}(ith.RegistersWritten())
	}
}

// Result from given computation.
type columnBatch struct {
	// Target registers for this batch
	targets []sc.RegisterRef
	// The computed columns in this batch.
	columns []trace.ArrayColumn
	// An error (should one arise)
	err error
}
