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

// RequiredSpillage returns the minimum amount of spillage required for a given
// module to ensure valid traces are accepted in the presence of arbitrary
// padding.  Spillage can only arise from computations as this is where values
// outside of the user's control are determined.
func RequiredSpillage(module uint, schema Schema) uint {
	// Ensures always at least one row of spillage (referred to as the "initial
	// padding row")
	mx := uint(1)
	// Determine if any more spillage required
	for i := schema.Assignments(); i.HasNext(); {
		// Get ith assignment
		ith := i.Next()
		//
		if ith.Context().Module() == module {
			// NOTE: Spillage is only currently considered to be necessary at
			// the front (i.e. start) of a trace.  This is because the prover
			// always inserts padding at the front, never the back.  As such, it
			// is the maximum positive shift which determines how much spillage
			// is required for a comptuation.
			mx = max(mx, ith.Bounds().End)
		}
	}

	return mx
}

// DefensivePadding returns the maximum amount of front padding required to
// ensure no constraint operating in the active region is clipped.  Observe that
// only front padding is considered because, for now, we assume the prover will
// only pad at the front.
func DefensivePadding(module uint, schema Schema) uint {
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

// QualifiedName returns the fully qualified name for a given (indexed) column in a given schema.
func QualifiedName(schema Schema, column uint) string {
	col := schema.Columns().Nth(column)
	return col.QualifiedName(schema)
}

// JoinContexts combines one or more evaluation contexts together.  If all
// expressions have the void context, then this is returned.  Likewise, if any
// expression has a conflicting context then this is returned.  Finally, if any
// two expressions have conflicting contexts between them, then the conflicting
// context is returned.  Otherwise, the common context to all expressions is
// returned.
func JoinContexts[E Contextual](args []E, schema Schema) tr.Context {
	ctx := tr.VoidContext[uint]()
	//
	for _, e := range args {
		ctx = ctx.Join(e.Context(schema))
	}
	// If we get here, then no conflicts were detected.
	return ctx
}

// ContextOfColumns determines the enclosing context for a given set of columns.
// If all columns have the void context, then this is returned.  Likewise,
// if any column has a conflicting context then this is returned.  Finally,
// if any two columns have conflicting contexts between them, then the
// conflicting context is returned.  Otherwise, the common context to all
// columns is returned.
func ContextOfColumns(cols []uint, schema Schema) tr.Context {
	ctx := tr.VoidContext[uint]()
	//
	for i := 0; i < len(cols); i++ {
		col := schema.Columns().Nth(cols[i])
		ctx = ctx.Join(col.Context)
	}
	// Done
	return ctx
}

// Accepts determines whether this schema will accept a given trace.  That is,
// whether or not the given trace adheres to the schema constraints.  A trace
// can fail to adhere to the schema for a variety of reasons, such as having a
// constraint which does not hold.  Observe that this does not check assertions
// within the schema hold.
//
//nolint:revive
func Accepts(parallel bool, batchsize uint, schema Schema, trace tr.Trace) (CoverageMap, []Failure) {
	return accepts(parallel, batchsize, schema.Constraints(), trace, "Constraint")
}

// Asserts determines whether or not this schema will "assert" a given trace.
// That is, whether or not the given trace adheres to the schema assertions.
func Asserts(parallel bool, batchsize uint, schema Schema, trace tr.Trace) (CoverageMap, []Failure) {
	return accepts(parallel, batchsize, schema.Assertions(), trace, "Assertion")
}

//nolint:revive
func accepts(parallel bool, batchsize uint, iter iter.Iterator[Constraint], trace tr.Trace,
	kind string) (CoverageMap, []Failure) {
	//
	if parallel {
		return parallelAccepts(batchsize, iter, trace, kind)
	}
	// sequential
	return sequentialAccepts(iter, trace)
}

func sequentialAccepts(iter iter.Iterator[Constraint], trace tr.Trace) (CoverageMap, []Failure) {
	coverage := NewBranchCoverage()
	errors := make([]Failure, 0)
	//
	for iter.HasNext() {
		ith := iter.Next()
		//
		data, err := ith.Accepts(trace)
		if err != nil {
			errors = append(errors, err)
		}
		//
		coverage.Insert(ith.Name(), data)
	}
	//
	return coverage, errors
}

func parallelAccepts(batchsize uint, iter iter.Iterator[Constraint], trace tr.Trace,
	kind string) (CoverageMap, []Failure) {
	//
	coverage := NewBranchCoverage()
	errors := make([]Failure, 0)
	// Initialise batch number (for debugging purposes)
	batch := uint(0)
	// Process constraints in batches
	for iter.HasNext() {
		errs := processConstraintBatch(kind, batch, batchsize, iter, &coverage, trace)
		errors = append(errors, errs...)
		// Increment batch number
		batch++
	}
	// Success
	return coverage, errors
}

// Process a given set of constraints in a single batch whilst recording all constraint failures.
func processConstraintBatch(logtitle string, batch uint, batchsize uint, iter iter.Iterator[Constraint],
	coverage *CoverageMap, trace tr.Trace) []Failure {
	n := uint(0)
	c := make(chan pcOutcome, 1024)
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
			c <- pcOutcome{ith.Name(), cov, err}
		}()
	}
	//
	for i := uint(0); i < n; i++ {
		p := <-c
		// Read from channel
		if p.error != nil {
			errors = append(errors, p.error)
		}
		// Update coverage
		coverage.Insert(p.handle, p.data)
	}
	// Log stats about this batch
	stats.Log(fmt.Sprintf("%s batch %d (%d items)", logtitle, batch, n))
	//
	return errors
}

type pcOutcome struct {
	handle string
	data   bit.Set
	error  Failure
}

// ColumnIndexOf returns the column index of the column with the given name, or
// returns false if no matching column exists.
func ColumnIndexOf(schema Schema, module uint, name string) (uint, bool) {
	return schema.Columns().Find(func(c Column) bool {
		return c.Context.Module() == module && c.Name == name
	})
}
