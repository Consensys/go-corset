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
	"github.com/consensys/go-corset/pkg/schema"
	sc "github.com/consensys/go-corset/pkg/schema"
	"github.com/consensys/go-corset/pkg/schema/agnostic"
	"github.com/consensys/go-corset/pkg/trace"
	tr "github.com/consensys/go-corset/pkg/trace"
	"github.com/consensys/go-corset/pkg/util"
	"github.com/consensys/go-corset/pkg/util/collection/word"
)

// TraceValidation validates that values held in trace columns match the
// expected type.  This is really a sanity check that the trace is not
// malformed.
func TraceValidation(parallel bool, schema sc.AnySchema, tr tr.Trace) []error {
	// Validate expanded trace
	if parallel {
		// Run (parallel) trace validation
		return ParallelTraceValidation(schema, tr)
	}
	// Run (sequential) trace validation
	return SequentialTraceValidation(schema, tr)
}

// TraceExpansion expands a given trace according to a given schema. More
// specifically, that means computing the actual values for any assignments.
// This is done using a straightforward sequential algorithm.
func TraceExpansion(parallel bool, batchsize uint, schema sc.AnySchema, trace *tr.ArrayTrace) error {
	if parallel {
		// Run (parallel) trace expansion
		return ParallelTraceExpansion(batchsize, schema, trace)
	}
	//
	return SequentialTraceExpansion(schema, trace)
}

// SplitRawColumns splits the given columns according to a given register
// mapping or, otherwise, simply lowers them.
func SplitRawColumns(rawCols []trace.RawColumn[word.BigEndian], expand bool,
	mapping schema.RegisterMap) []trace.RawFrColumn {
	//
	var (
		stats = util.NewPerfStats()
		cols  []trace.RawFrColumn
	)
	// Split raw columns according to the mapping (if applicable).  Note that
	// expansion being disabled implies the trace is already split
	// appropriately.
	if mapping != nil && expand {
		cols = agnostic.SplitRawColumns(rawCols, mapping)
	} else {
		cols = agnostic.LowerRawColumns(rawCols)
	}
	//
	stats.Log("Splitting trace")
	//
	return cols
}
