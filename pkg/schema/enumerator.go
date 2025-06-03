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
	"github.com/consensys/gnark-crypto/ecc/bls12-377/fr"
	tr "github.com/consensys/go-corset/pkg/trace"
	"github.com/consensys/go-corset/pkg/util/collection/enum"
	"github.com/consensys/go-corset/pkg/util/field"
	log "github.com/sirupsen/logrus"
)

// ============================================================================
// TraceEnumerator
// ============================================================================

// TraceEnumerator is an adaptor which surrounds an enumerator and, essentially,
// converts flat sequences of elements into traces.
type TraceEnumerator struct {
	// Schema for which traces are being generated
	schema Schema
	// Number of lines
	lines uint
	// Enumerate sequences of elements
	enumerator enum.Enumerator[[]fr.Element]
}

// NewTraceEnumerator constructs an enumerator for all traces matching the
// given column specifications using elements sourced from the given pool.
func NewTraceEnumerator(lines uint, schema Schema, pool []fr.Element) enum.Enumerator[tr.Trace] {
	ncells := schema.InputColumns().Count() * lines
	// Construct the enumerator
	enumerator := enum.Power(ncells, pool)
	// Done
	return &TraceEnumerator{schema, lines, enumerator}
}

// Nth returns the nth item in this iterator.  This will mutate the iterator.
func (p *TraceEnumerator) Nth(n uint) tr.Trace {
	return p.buildTrace(p.enumerator.Nth(n))
}

// Count returns the number of items left in this enumeration.
//
//nolint:revive
func (p *TraceEnumerator) Count() uint {
	return p.enumerator.Count()
}

// Next returns the next trace in the enumeration
func (p *TraceEnumerator) Next() tr.Trace {
	return p.buildTrace(p.enumerator.Next())
}

func (p *TraceEnumerator) buildTrace(elems []fr.Element) tr.Trace {
	ncols := p.schema.InputColumns().Count()
	cols := make([]tr.RawColumn, ncols)
	//
	i, j := 0, 0
	// Construct each column from the sequence
	for iter := p.schema.InputColumns(); iter.HasNext(); {
		col := iter.Next()
		data := field.NewFrArray(p.lines, 256)
		// Slice nrows values from elems
		for k := uint(0); k < p.lines; k++ {
			data.Set(k, elems[j])
			// Consume element from generated sequence
			j++
		}
		// Construct raw column
		modName := p.schema.Modules().Nth(col.Context.Module()).Name
		cols[i] = tr.RawColumn{Module: modName, Name: col.Name, Data: data}
		i++
	}
	// Finally, build the trace.
	builder := NewTraceBuilder(p.schema).Expand(true).Parallel(false).Padding(0)
	// Build the trace
	trace, errs := builder.Build(cols)
	// Handle errors
	if errs != nil {
		// Should be unreachable, since control the trace!
		for _, err := range errs {
			log.Error(err)
		}
		// Fail
		panic("invalid trace constructed")
	}
	// Done
	return trace
}

// HasNext checks whether the enumeration has more elements (or not).
func (p *TraceEnumerator) HasNext() bool {
	return p.enumerator.HasNext()
}
