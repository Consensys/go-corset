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
package ir

import (
	"fmt"
	"math"

	"github.com/consensys/go-corset/pkg/ir/builder"
	sc "github.com/consensys/go-corset/pkg/schema"
	"github.com/consensys/go-corset/pkg/trace"
	"github.com/consensys/go-corset/pkg/trace/lt"
	"github.com/consensys/go-corset/pkg/util/field"
	"github.com/consensys/go-corset/pkg/util/field/bls12_377"
	"github.com/consensys/go-corset/pkg/util/word"
)

// TraceBuilder provides a mechanical means of constructing a trace from a given
// schema and set of input columns.  The goal is to encapsulate all of the logic
// around building a trace.
type TraceBuilder struct {
	// Indicates whether or not to perform defensive padding.  This is where
	// padding rows are appended and/or prepended to ensure no constraint in the
	// active region of the trace is clipped.  Whilst not strictly necessary,
	// this can be helpful for identifying invalid constraints which are only
	// exposed with a given amount of padding.
	defensive bool
	// Indicates whether or not to perform trace expansion.  The default should
	// be to apply trace expansion.  However, for testing purposes, it can be
	// useful to provide an already expanded trace to ensure a set of
	// constraints correctly rejects it.
	expand bool
	// Indicates whether or not to validate all column types.  That is, check
	// that the values supplied for all columns (both input and computed) are
	// within their declared type.
	validate bool
	// Indicates whether or not to apply other sanity checks, such as ensuring
	// the number of lines actually added to a trace matches the expected
	// amount.
	checks bool
	// Determines the amount of padding to apply to each module in the trace.
	// At the moment, this is applied uniformly across all modules.  This is
	// somewhat cumbersome, and it would make sense to support different
	// protocols.  For example, one obvious protocol is to expand a module's
	// length upto a power-of-two.
	padding uint
	// Determines whether or not trace expansion should be performed in
	// parallel.  This should be the default, but a sequential option is
	// retained for debugging purposes.
	parallel bool
	// Specify the maximum size of any dispatched batch.
	batchSize uint
	// Mapping specifies whether or not columns in the trace need to be split to
	// match the given field configuration.
	mapping sc.LimbsMap
}

// A column key is used as a key for the column map
type columnKey struct {
	module string
	column string
}

type columnId struct {
	module uint
	column uint
}

// NewTraceBuilder constructs a default trace builder.  The idea is that this
// could then be customized as needed following the builder pattern.
func NewTraceBuilder() TraceBuilder {
	return TraceBuilder{true, true, true, true, 0, true, math.MaxUint, nil}
}

// WithDefensivePadding updates a given builder configuration to apply defensive padding
// (or not).
func (tb TraceBuilder) WithDefensivePadding(flag bool) TraceBuilder {
	ntb := tb
	ntb.defensive = flag
	//
	return ntb
}

// WithExpansionChecks enables runtime safety checks on the expanded trace.
func (tb TraceBuilder) WithExpansionChecks(flag bool) TraceBuilder {
	ntb := tb
	ntb.checks = flag
	//
	return ntb
}

// WithExpansion updates a given builder configuration to perform trace expansion (or
// not).
func (tb TraceBuilder) WithExpansion(flag bool) TraceBuilder {
	ntb := tb
	ntb.expand = flag
	//
	return ntb
}

// Expanding indicates whether or not this builder will expand the trace.
func (tb TraceBuilder) Expanding() bool {
	return tb.expand
}

// WithRegisterMapping updates a given builder configuration to split the trace
// according to a given mapping of registers.
func (tb TraceBuilder) WithRegisterMapping(mapping sc.LimbsMap) TraceBuilder {
	ntb := tb
	ntb.mapping = mapping
	//
	return ntb
}

// WithValidation updates a given builder configuration to perform trace validation (or
// not).
func (tb TraceBuilder) WithValidation(flag bool) TraceBuilder {
	ntb := tb
	ntb.validate = flag
	//
	return ntb
}

// WithPadding updates a given builder configuration to use a given amount of padding
func (tb TraceBuilder) WithPadding(padding uint) TraceBuilder {
	ntb := tb
	ntb.padding = padding
	//
	return ntb
}

// WithParallelism updates a given builder configuration to allow trace expansion to be
// performed concurrently (or not).
func (tb TraceBuilder) WithParallelism(flag bool) TraceBuilder {
	ntb := tb
	ntb.parallel = flag
	//
	return ntb
}

// Parallelism checks whether parallelism is enabled for this builder.
func (tb TraceBuilder) Parallelism() bool {
	return tb.parallel
}

// WithBatchSize sets the maximum number of batches to run in parallel during trace
// expansion.
func (tb TraceBuilder) WithBatchSize(batchSize uint) TraceBuilder {
	ntb := tb
	ntb.batchSize = batchSize
	//
	return ntb
}

// BatchSize returns the configure batch size for this builder.
func (tb TraceBuilder) BatchSize() uint {
	return tb.batchSize
}

// Build attempts to construct a trace for a given schema, producing errors if
// there are inconsistencies (e.g. missing columns, duplicate columns, etc).
func (tb TraceBuilder) Build(schema sc.AnySchema, tf lt.TraceFile) (trace.Trace[bls12_377.Element], []error) {
	var (
		pool   word.Pool[uint, bls12_377.Element]
		cols   []trace.RawColumn[bls12_377.Element]
		errors []error
	)
	// If expansion is enabled, then we must split the trace according to the
	// given mapping; otherwise, we simply lower the trace as is.
	if tb.mapping != nil && tb.expand {
		// Split raw columns, and handle any errors arising.
		if pool, cols, errors = builder.TraceSplitting[bls12_377.Element](tb.parallel, tf, tb.mapping); len(errors) > 0 {
			return nil, errors
		}
	} else {
		// Lower raw columns
		pool, cols = builder.TraceLowering[bls12_377.Element](tb.parallel, tf)
	}
	// Initialise the actual trace object
	tr, errors := initialiseTrace(!tb.expand, schema, pool, cols)
	//
	if len(errors) > 0 {
		// Critical failure
		return nil, errors
	} else if tb.expand {
		// Save original line counts
		moduleHeights := determineModuleHeights(tr)
		// Apply spillage
		applySpillageAndDefensivePadding(tb.defensive, tr, schema)
		// Sanity checks
		if tb.checks {
			if err := checkModuleHeights(moduleHeights, tb.defensive, tr, schema); err != nil {
				return nil, append(errors, err)
			}
		}
		// Expand trace
		if err := builder.TraceExpansion(tb.parallel, tb.batchSize, schema, tr); err != nil {
			return nil, append(errors, err)
		}
		// Validate expanded trace
		if tb.validate {
			// Run (parallel) trace validation
			if errs := builder.TraceValidation(tb.parallel, schema, tr); len(errs) > 0 {
				return nil, append(errors, errs...)
			}
		}
	}
	// Padding
	if tb.padding > 0 {
		padColumns(tr, schema, tb.padding)
	}
	//
	return tr, errors
}

func initialiseTrace[F field.Element[F]](expanded bool, schema sc.AnySchema, pool word.Pool[uint, F],
	cols []trace.RawColumn[F]) (*trace.ArrayTrace[F], []error) {
	//
	var (
		// Initialise modules
		modmap  = initialiseModuleMap(schema)
		modules = make([]trace.ArrayModule[F], schema.Width())
	)
	//
	columns, errors := splitTraceColumns(expanded, schema, modmap, cols)
	//
	for i := uint(0); i != schema.Width(); i++ {
		var mod = schema.Module(i)
		//
		modules[i] = fillTraceModule(mod.Name(), mod.LengthMultiplier(), columns[i])
	}
	// Done
	return trace.NewArrayTrace(pool, modules), errors
}

func initialiseModuleMap(schema sc.AnySchema) map[string]uint {
	modmap := make(map[string]uint, 100)
	// Initialise modules
	for i := uint(0); i != schema.Width(); i++ {
		m := schema.Module(i)
		// Sanity check module
		if _, ok := modmap[m.Name()]; ok {
			panic(fmt.Sprintf("duplicate module '%s' in schema", m.Name()))
		}

		modmap[m.Name()] = i
	}
	// Done
	return modmap
}

func splitTraceColumns[T word.Word[T]](expanded bool, schema sc.AnySchema, modmap map[string]uint,
	cols []trace.RawColumn[T]) ([][]trace.RawColumn[T], []error) {
	//
	var (
		// Errs contains the set of filling errors which are accumulated
		errs []error
		//
		seen map[columnKey]bool = make(map[columnKey]bool, 0)
	)
	//
	colmap, modules := initialiseColumnMap[T](expanded, schema)
	// Assign data from each input column given
	for _, col := range cols {
		// Lookup the module
		if _, ok := modmap[col.Module]; !ok {
			errs = append(errs, fmt.Errorf("unknown module '%s' in trace", col.Module))
		} else {
			key := columnKey{col.Module, col.Name}
			// Determine enclosiong module height
			cid, ok := colmap[key]
			// More sanity checks
			if !ok {
				errs = append(errs, fmt.Errorf("unknown column '%s' in trace", col.QualifiedName()))
			} else if _, ok := seen[key]; ok {
				errs = append(errs, fmt.Errorf("duplicate column '%s' in trace", col.QualifiedName()))
			} else {
				seen[key] = true
				modules[cid.module][cid.column] = col
			}
		}
	}
	// Sanity check everything was assigned
	for i, m := range modules {
		mod := schema.Module(uint(i))
		//
		for j, c := range m {
			rid := sc.NewRegisterId(uint(j))
			reg := mod.Register(rid)
			//
			if reg.IsInputOutput() && c.Data == nil {
				errs = append(errs, fmt.Errorf("missing input/output column '%s' from trace", c.QualifiedName()))
			} else if expanded && c.Data == nil {
				errs = append(errs, fmt.Errorf("missing computed column '%s' from expanded trace", c.QualifiedName()))
			}
		}
	}
	//
	return modules, errs
}

func initialiseColumnMap[T word.Word[T]](expanded bool, schema sc.AnySchema) (map[columnKey]columnId, [][]trace.RawColumn[T]) {
	var (
		colmap  = make(map[columnKey]columnId, 100)
		modules = make([][]trace.RawColumn[T], schema.Width())
	)
	// Initialise modules
	for i := uint(0); i != schema.Width(); i++ {
		m := schema.Module(i)
		columns := make([]trace.RawColumn[T], m.Width())
		//
		for j := uint(0); j != m.Width(); j++ {
			var (
				rid = sc.NewRegisterId(j)
				col = m.Register(rid)
				key = columnKey{m.Name(), col.Name}
				id  = columnId{i, j}
			)
			//
			if _, ok := colmap[key]; ok {
				panic(fmt.Sprintf("duplicate column '%s' in schema", trace.QualifiedColumnName(m.Name(), col.Name)))
			}
			// Add initially empty column
			columns[j] = trace.RawColumn[T]{
				Module: m.Name(),
				Name:   col.Name,
				Data:   nil,
			}
			// Set column as expected if appropriate.
			if expanded || col.IsInputOutput() {
				colmap[key] = id
			}
		}
		// Initialise empty columns for this module.
		modules[i] = columns
	}
	// Done
	return colmap, modules
}

func fillTraceModule[F field.Element[F]](name string, multiplier uint, rawColumns []trace.RawColumn[F]) trace.ArrayModule[F] {
	var (
		traceColumns = make([]trace.ArrayColumn[F], len(rawColumns))
		zero         F
	)
	//
	for i := range traceColumns {
		var ith = rawColumns[i]
		//
		traceColumns[i] = trace.NewArrayColumn(ith.Name, ith.Data, zero)
	}
	//
	return trace.NewArrayModule(name, multiplier, traceColumns)
}

// pad each module with its given level of spillage and (optionally) ensure a
// given level of defensive padding.
func applySpillageAndDefensivePadding[F field.Element[F]](defensive bool, tr *trace.ArrayTrace[F], schema sc.AnySchema) {
	n := tr.Modules().Count()
	// Iterate over modules
	for i := uint(0); i < n; i++ {
		// Compute extra padding rows required
		padding := sc.RequiredPaddingRows(i, defensive, schema)
		// Don't pad unless we have to
		if padding > 0 {
			// Pad extract rows with 0
			tr.Pad(i, padding, 0)
		}
	}
}

// determineModuleHeights returns the height for each module in the trace.
func determineModuleHeights[F field.Element[F]](tr *trace.ArrayTrace[F]) []uint {
	n := tr.Modules().Count()
	mid := 0
	heights := make([]uint, n)
	// Iterate over modules
	for iter := tr.Modules(); iter.HasNext(); {
		ith := iter.Next()
		heights[mid] = ith.Height()
		mid++
	}
	//
	return heights
}

// checkModuleHeights checks the expanded heights match exactly what was
// expected.
func checkModuleHeights[F field.Element[F]](original []uint, defensive bool, tr *trace.ArrayTrace[F], schema sc.AnySchema) error {
	expanded := determineModuleHeights(tr)
	//
	for mid := uint(0); mid < uint(len(expanded)); mid++ {
		spillage := sc.RequiredPaddingRows(mid, defensive, schema)
		expected := original[mid] + spillage
		// Perform the check
		if expected != expanded[mid] {
			name := schema.Module(mid).Name()
			return fmt.Errorf(
				"inconsistent expanded trace height for %s (was %d but expected %d)", name, expanded[mid], expected)
		}
	}
	//
	return nil
}

// PadColumns pads every column in a given trace with a given amount of (front)
// padding. Observe that this applies on top of any spillage and/or defensive
// padding already applied.
func padColumns[F field.Element[F]](tr *trace.ArrayTrace[F], schema sc.AnySchema, padding uint) {
	n := tr.Modules().Count()
	// Iterate over modules
	for i := uint(0); i < n; i++ {
		multiplier := schema.Module(i).LengthMultiplier()
		tr.Pad(i, padding*multiplier, 0)
	}
}
