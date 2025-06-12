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
package schema

import (
	"fmt"

	tr "github.com/consensys/go-corset/pkg/trace"
	"github.com/consensys/go-corset/pkg/util"
	"github.com/consensys/go-corset/pkg/util/collection/bit"
	"github.com/consensys/go-corset/pkg/util/collection/iter"
)

// RequiredPaddingRows determines the number of additional (spillage / padding)
// rows that will be added during trace expansion.  The exact value depends on
// whether defensive padding is enabled or not.
func RequiredPaddingRows(module uint, defensive bool, schema AnySchema) uint {
	var (
		multiplier = schema.Module(module).LengthMultiplier()
		padding    = requiredSpillage(module, schema)
	)
	//
	if defensive {
		// determine minimum levels of defensive padding required.
		padding = max(padding, defensivePadding(module, schema))
	}
	// Technically, we could avoid multiplying by the multiplier here, but in
	// practice it shouldn't matter.  That's because of the very limited ways in
	// which interleaved columns are used in practice.
	return padding * multiplier
}

// RequiredSpillage returns the minimum amount of spillage required for a given
// module to ensure valid traces are accepted in the presence of arbitrary
// padding.  Spillage can only arise from computations as this is where values
// outside of the user's control are determined.
func requiredSpillage(module uint, schema AnySchema) uint {
	// Ensures always at least one row of spillage (referred to as the "initial
	// padding row")
	mx := uint(1)
	// Determine if any more spillage required
	for i := schema.Assignments(); i.HasNext(); {
		// Get ith assignment
		ith := i.Next()
		// NOTE: Spillage is only currently considered to be necessary at
		// the front (i.e. start) of a trace.  This is because the prover
		// always inserts padding at the front, never the back.  As such, it
		// is the maximum positive shift which determines how much spillage
		// is required for a comptuation.
		mx = max(mx, ith.Bounds(module).End)
	}
	//
	return mx
}

// DefensivePadding returns the maximum amount of front padding required to
// ensure no constraint operating in the active region is clipped.  Observe that
// only front padding is considered because, for now, we assume the prover will
// only pad at the front.
func defensivePadding(module uint, schema AnySchema) uint {
	front := uint(0)
	// Determine maximum amounts of defensive padding required for constraints.
	for i := schema.Constraints(); i.HasNext(); {
		bounds := i.Next().Bounds(module)
		//
		front = max(front, bounds.Start)
	}
	//
	return front
}

// Accepts determines whether this schema will accept a given trace.  That is,
// whether or not the given trace adheres to the schema constraints.  A trace
// can fail to adhere to the schema for a variety of reasons, such as having a
// constraint which does not hold.
//
//nolint:revive
func Accepts[C Constraint](parallel bool, batchsize uint, schema Schema[C], trace tr.Trace) []Failure {
	return accepts(parallel, batchsize, schema.Constraints(), trace, "Constraint")
}

//nolint:revive
func accepts[C Constraint](parallel bool, batchsize uint, iter iter.Iterator[C], trace tr.Trace,
	kind string) []Failure {
	//
	if parallel {
		return parallelAccepts(batchsize, iter, trace, kind)
	}
	// sequential
	return sequentialAccepts(iter, trace)
}

func sequentialAccepts[C Constraint](iter iter.Iterator[C], trace tr.Trace) []Failure {
	errors := make([]Failure, 0)
	//
	for iter.HasNext() {
		ith := iter.Next()
		//
		_, err := ith.Accepts(trace)
		if err != nil {
			errors = append(errors, err)
		}
	}
	//
	return errors
}

func parallelAccepts[C Constraint](batchsize uint, iter iter.Iterator[C], trace tr.Trace,
	kind string) []Failure {
	//
	errors := make([]Failure, 0)
	// Initialise batch number (for debugging purposes)
	batch := uint(0)
	// Process constraints in batches
	for iter.HasNext() {
		errs := processConstraintBatch(kind, batch, batchsize, iter, trace)
		errors = append(errors, errs...)
		// Increment batch number
		batch++
	}
	// Success
	return errors
}

// Process a given set of constraints in a single batch whilst recording all constraint failures.
func processConstraintBatch[C Constraint](logtitle string, batch uint, batchsize uint, iter iter.Iterator[C],
	trace tr.Trace) []Failure {
	n := uint(0)
	c := make(chan batchOutcome, 1024)
	errors := make([]Failure, 0)
	stats := util.NewPerfStats()
	// Launch at most 100 go-routines.
	for ; n < batchsize && iter.HasNext(); n++ {
		// Get ith constraint
		ith := iter.Next()
		// Launch checker for constraint
		go func() {
			// Send outcome back
			cov, err := ith.Accepts(trace)
			context := ith.Contexts()[0]
			name := ith.Name()
			c <- batchOutcome{context, name, cov, err}
		}()
	}
	//
	for i := uint(0); i < n; i++ {
		p := <-c
		// Read from channel
		if p.error != nil {
			errors = append(errors, p.error)
		}
	}
	// Log stats about this batch
	stats.Log(fmt.Sprintf("%s batch %d (%d items)", logtitle, batch, n))
	//
	return errors
}

type batchOutcome struct {
	module uint
	handle string
	data   bit.Set
	error  Failure
}
