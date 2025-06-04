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
	"math"
	"math/big"

	"github.com/consensys/gnark-crypto/ecc/bls12-377/fr"
	"github.com/consensys/go-corset/pkg/trace"
	tr "github.com/consensys/go-corset/pkg/trace"
	"github.com/consensys/go-corset/pkg/util"
)

// TraceBuilder provides a mechanical means of constructing a trace from a given
// schema and set of input columns.  The goal is to encapsulate all of the logic
// around building a trace.
type TraceBuilder struct {
	// Schema to be used when building the trace
	schema AnySchema
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
func NewTraceBuilder[C Constraint](schema Schema[C]) TraceBuilder {
	return TraceBuilder{Any(schema), true, true, true, true, 0, true, math.MaxUint}
}

// Defensive updates a given builder configuration to apply defensive padding
// (or not).
func (tb TraceBuilder) Defensive(flag bool) TraceBuilder {
	ntb := tb
	ntb.defensive = flag
	//
	return ntb
}

// ExpansionChecks enables runtime safety checks on the expanded trace.
func (tb TraceBuilder) ExpansionChecks(flag bool) TraceBuilder {
	ntb := tb
	ntb.checks = flag
	//
	return ntb
}

// Expand updates a given builder configuration to perform trace expansion (or
// not).
func (tb TraceBuilder) Expand(flag bool) TraceBuilder {
	ntb := tb
	ntb.expand = flag
	//
	return ntb
}

// Validate updates a given builder configuration to perform trace validation (or
// not).
func (tb TraceBuilder) Validate(flag bool) TraceBuilder {
	ntb := tb
	ntb.validate = flag
	//
	return ntb
}

// Padding updates a given builder configuration to use a given amount of padding
func (tb TraceBuilder) Padding(padding uint) TraceBuilder {
	ntb := tb
	ntb.padding = padding
	//
	return ntb
}

// Parallel updates a given builder configuration to allow trace expansion to be
// performed concurrently (or not).
func (tb TraceBuilder) Parallel(flag bool) TraceBuilder {
	ntb := tb
	ntb.parallel = flag
	//
	return ntb
}

// BatchSize sets the maximum number of batches to run in parallel during trace
// expansion.
func (tb TraceBuilder) BatchSize(batchSize uint) TraceBuilder {
	ntb := tb
	ntb.batchSize = batchSize
	//
	return ntb
}

// Build attempts to construct a trace for a given schema, producing errors if
// there are inconsistencies (e.g. missing columns, duplicate columns, etc).
func (tb TraceBuilder) Build(cols []trace.RawColumn) (trace.Trace, []error) {
	tr, errors := initialiseTrace(tb.schema, cols)
	//
	if len(errors) > 0 {
		// Critical failure
		return nil, errors
	} else if tb.expand {
		// Save original line counts
		moduleHeights := determineModuleHeights(tr)
		// Apply spillage
		applySpillageAndDefensivePadding(tb.defensive, tr, tb.schema)
		// Sanity checks
		if tb.checks {
			if err := checkModuleHeights(moduleHeights, tb.defensive, tr, tb.schema); err != nil {
				return nil, append(errors, err)
			}
		}
		// Expand trace
		if tb.parallel {
			// Run (parallel) trace expansion
			if err := parallelTraceExpansion(tb.batchSize, tb.schema, tr); err != nil {
				return nil, append(errors, err)
			}
		} else if err := sequentialTraceExpansion(tb.schema, tr); err != nil {
			// Expansion errors are fatal as well
			return nil, append(errors, err)
		}
		// Validate expanded trace
		if tb.validate && tb.parallel {
			// Run (parallel) trace validation
			if errs := parallelTraceValidation(tb.schema, tr); len(errs) > 0 {
				return nil, append(errors, errs...)
			}
		} else if tb.validate {
			// Run (sequential) trace validation
			if errs := sequentialTraceValidation(tb.schema, tr); len(errs) > 0 {
				return nil, append(errors, errs...)
			}
		}
	}
	// Padding
	if tb.padding > 0 {
		padColumns(tr, tb.padding)
	}
	//
	return tr, errors
}

func initialiseTrace(schema AnySchema, cols []trace.RawColumn) (*trace.ArrayTrace, []error) {
	var (
		// Initialise modules
		modmap  = initialiseModuleMap(schema)
		modules = make([]trace.ArrayModule, schema.Width())
	)
	//
	columns, errors := splitTraceColumns(schema, modmap, cols)
	//
	for i := uint(0); i != schema.Width(); i++ {
		var name = schema.Module(i).Name()
		//
		modules[i] = fillTraceModule(i, name, columns[i])
	}
	// Done
	return trace.NewArrayTrace(modules), errors
}

func initialiseModuleMap(schema AnySchema) map[string]uint {
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

func splitTraceColumns(schema AnySchema, modmap map[string]uint,
	cols []trace.RawColumn) ([][]trace.RawColumn, []error) {
	//
	var (
		// Errs contains the set of filling errors which are accumulated
		errs []error
		//
		seen map[columnKey]bool = make(map[columnKey]bool, 0)
	)
	//
	colmap, modules := initialiseColumnMap(schema)
	// Assign data from each input column given
	for _, c := range cols {
		// Lookup the module
		if _, ok := modmap[c.Module]; !ok {
			errs = append(errs, fmt.Errorf("unknown module '%s' in trace", c.Module))
		} else {
			key := columnKey{c.Module, c.Name}
			// Determine enclosiong module height
			cid, ok := colmap[key]
			// More sanity checks
			if !ok {
				errs = append(errs, fmt.Errorf("unknown column '%s' in trace", c.QualifiedName()))
			} else if _, ok := seen[key]; ok {
				errs = append(errs, fmt.Errorf("duplicate column '%s' in trace", c.QualifiedName()))
			} else {
				seen[key] = true
				modules[cid.module][cid.column] = trace.RawColumn{
					Module: c.Module,
					Name:   c.Name,
					Data:   c.Data.Clone(),
				}
			}
		}
	}
	// Sanity check everything was assigned
	for i, m := range modules {
		mod := schema.Module(uint(i))
		//
		for j, c := range m {
			reg := mod.Register(uint(j))
			//
			if reg.IsInputOutput() && c.Data == nil {
				errs = append(errs, fmt.Errorf("missing input/output column '%s' from trace", c.QualifiedName()))
			}
		}
	}
	//
	return modules, errs
}

func initialiseColumnMap(schema AnySchema) (map[columnKey]columnId, [][]trace.RawColumn) {
	var (
		colmap  = make(map[columnKey]columnId, 100)
		modules = make([][]trace.RawColumn, schema.Width())
	)
	// Initialise modules
	for i := uint(0); i != schema.Width(); i++ {
		m := schema.Module(i)
		columns := make([]trace.RawColumn, m.Width())
		//
		for j := uint(0); j != m.Width(); j++ {
			col := m.Register(j)
			key := columnKey{m.Name(), col.Name}
			id := columnId{i, j}
			//
			if _, ok := colmap[key]; ok {
				panic(fmt.Sprintf("duplicate column '%s' in schema", trace.QualifiedColumnName(m.Name(), col.Name)))
			}
			//
			colmap[key] = id
			// Add dummy column for debugging purposes
			columns[j] = trace.RawColumn{
				Module: m.Name(),
				Name:   col.Name,
				Data:   nil,
			}
		}
		// Initialise empty columns for this module.
		modules[i] = columns
	}
	// Done
	return colmap, modules
}

func fillTraceModule(mid uint, name string, rawColumns []trace.RawColumn) trace.ArrayModule {
	var (
		traceColumns = make([]trace.ArrayColumn, len(rawColumns))
		zero         = fr.NewElement(0)
	)
	//
	for i := range traceColumns {
		ith := rawColumns[i]
		ctx := trace.NewContext(mid, 1)
		//
		traceColumns[i] = trace.NewArrayColumn(ctx, ith.Name, ith.Data, zero)
	}
	//
	return trace.NewArrayModule(name, traceColumns)
}

// pad each module with its given level of spillage and (optionally) ensure a
// given level of defensive padding.
func applySpillageAndDefensivePadding(defensive bool, tr *trace.ArrayTrace, schema AnySchema) {
	n := tr.Modules().Count()
	// Iterate over modules
	for i := uint(0); i < n; i++ {
		// Compute extra padding rows required
		padding := RequiredPaddingRows(i, defensive, schema)
		// Pad extract rows with 0
		tr.Pad(i, padding, 0)
	}
}

// determineModuleHeights returns the height for each module in the trace.
func determineModuleHeights(tr *trace.ArrayTrace) []uint {
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
func checkModuleHeights(original []uint, defensive bool, tr *trace.ArrayTrace, schema AnySchema) error {
	expanded := determineModuleHeights(tr)
	//
	for mid := uint(0); mid < uint(len(expanded)); mid++ {
		spillage := RequiredPaddingRows(mid, defensive, schema)
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
func padColumns(tr *trace.ArrayTrace, padding uint) {
	n := tr.Modules().Count()
	// Iterate over modules
	for i := uint(0); i < n; i++ {
		tr.Pad(i, padding, 0)
	}
}

// ============================================================================
// Sequential Expansion / Validation
// ============================================================================

// sequentialTraceExpansion expands a given trace according to a given schema.
// More specifically, that means computing the actual values for any
// assignments.  This is done using a straightforward sequential algorithm.
func sequentialTraceExpansion(schema AnySchema, trace *trace.ArrayTrace) error {
	var err error
	// Compute each assignment in turn
	for i := schema.Assignments(); i.HasNext(); {
		var cols []tr.ArrayColumn
		// Get ith assignment
		ith := i.Next()
		// Compute ith assignment(s)
		if cols, err = ith.Compute(trace, schema); err != nil {
			return err
		}
		// Fill all computed columns
		fillComputedColumns(ith.Registers(), ith.Module(), cols, trace)
	}
	// Done
	return nil
}

// Validate that values held in trace columns match the expected type.  This is
// really a sanity check that the trace is not malformed.
func sequentialTraceValidation(schema AnySchema, tr trace.Trace) []error {
	var errors []error
	//
	for i := uint(0); i < max(schema.Width(), tr.Width()); i++ {
		// Sanity checks first
		if i >= tr.Width() {
			err := fmt.Errorf("module %s missing from trace", schema.Module(i).Name())
			errors = append(errors, err)
		} else if i >= schema.Width() {
			err := fmt.Errorf("unknown module %s in trace", tr.Module(i).Name())
			errors = append(errors, err)
		} else {
			var (
				scMod = schema.Module(i)
				trMod = tr.Module(i)
			)
			// Validate module
			errors = append(errors, sequentialModuleValidation(scMod, trMod)...)
		}
	}
	// Done
	return errors
}

func sequentialModuleValidation(scMod Module, trMod trace.Module) []error {
	var (
		errors []error
		// Extract module registers
		registers = scMod.Registers()
	)
	// Sanity check
	if scMod.Name() != trMod.Name() {
		err := fmt.Errorf("misaligned module during trace expansion (%s vs %s)", scMod.Name(), trMod.Name())
		errors = append(errors, err)
	} else {
		for i := uint(0); i < max(trMod.Width(), scMod.Width()); i++ {
			// Sanity checks first
			if i >= trMod.Width() {
				err := fmt.Errorf("register %s.%s missing from trace", trMod.Name(), registers[i].Name)
				errors = append(errors, err)
			} else if i >= scMod.Width() {
				err := fmt.Errorf("unknown register %s.%s in trace", trMod.Name(), trMod.Column(i).Name())
				errors = append(errors, err)
			} else {
				var (
					reg  Register     = scMod.Register(i)
					data trace.Column = trMod.Column(i)
				)
				// Sanity check data has expected bitwidth
				if err := validateColumnBitWidth(reg.Width, data, scMod); err != nil {
					errors = append(errors, err)
				}
			}
		}
	}
	// Done
	return errors
}

// ============================================================================
// Parallel Expansion / Validation
// ============================================================================

// Perform trace expansion using concurrently executing jobs.  The chosen
// algorithm operates in waves, rather than using an continuous approach.  This
// is for two reasons: firstly, the latter would require locks that would slow
// down evaluation performance; secondly, the vast majority of jobs are run in
// the very first wave.
func parallelTraceExpansion(batchsize uint, schema AnySchema, trace *tr.ArrayTrace) error {
	batch := 0
	// Construct a communication channel for errors.
	ch := make(chan columnBatch, 1024)
	// Determine number of input columns
	ninputs := schema.InputColumns().Count()
	// Determine number of columns to compute
	ntodo := schema.Assignments().Count()
	// Iterate until all columns completed.
	for ntodo > 0 {
		stats := util.NewPerfStats()
		// Dispatch next batch of assignments.
		n := dispatchReadyAssignments(batchsize, ninputs, schema, trace, ch)
		//
		batches := make([]columnBatch, n)
		// Collect all the results
		for i := uint(0); i < n; i++ {
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
			fillComputedColumns(r.index, r.columns, trace)
			//
			ntodo--
		}
		// Log stats about this batch
		stats.Log(fmt.Sprintf("Expansion batch %d (remaining %d)", batch, ntodo))
		// Increment batch
		batch++
	}
	// Done
	return nil
}

// Find any assignments which are ready to compute, and dispatch them with
// results being fed back into the shared channel.  This returns the number of
// jobs which have been dispatched (i.e. so the caller knows how many results to
// expect).
func dispatchReadyAssignments(batchsize uint, ninputs uint, schema AnySchema,
	trace *tr.ArrayTrace, ch chan columnBatch) uint {
	count := uint(0)
	//
	for iter, cid := schema.Assignments(), ninputs; iter.HasNext() && count < batchsize; {
		ith := iter.Next()
		// Check whether this assignment has already been computed and, if not,
		// whether or not it is ready.
		if trace.Column(cid).Data() == nil && isReady(ith, trace) {
			// Dispatch!
			go func(index uint) {
				cols, err := ith.ComputeColumns(trace)
				// Send outcome back
				ch <- columnBatch{index, cols, err}
			}(cid)
			// Increment dispatch count
			count++
		}
		// Update the column identifier
		cid += ith.Columns().Count()
	}
	// Done
	return count
}

// Check whether all dependencies for this assignment are available (that is,
// have their data already).
func isReady(assignment Assignment, trace *tr.ArrayTrace) bool {
	for _, cid := range assignment.Dependencies() {
		if trace.Column(cid).Data() == nil {
			return false
		}
	}
	// Done
	return true
}

// Result from given computation.
type columnBatch struct {
	// The column index of the first computed column in this batch.
	index uint
	// The computed columns in this batch.
	columns []trace.ArrayColumn
	// An error (should one arise)
	err error
}

// Validate that values held in trace columns match the expected type.  This is
// really a sanity check that the trace is not malformed.
func parallelTraceValidation(schema AnySchema, tr tr.Trace) []error {
	var errors []error

	schemaCols := schema.Columns()
	// Start timer
	stats := util.NewPerfStats()
	// Construct a communication channel for errors.
	c := make(chan error, 1024)
	// Check each column in turn
	for i := uint(0); i < tr.Width(); i++ {
		// Extract ith column
		col := tr.Column(i)
		// Extract schema for ith column
		scCol := schemaCols.Next()
		// Determine enclosing module
		mod := schema.Modules().Nth(scCol.Context.Module())
		// Extract type for ith column
		colType := scCol.DataType
		// Check elements
		go func() {
			// Send outcome back
			c <- validateColumnType(colType, col, mod)
		}()
	}
	// Collect up all the results
	for i := uint(0); i < tr.Width(); i++ {
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

// ============================================================================
// Helpers
// ============================================================================

// Fill a set of columns with their computed results.  The column index is that
// of the first column in the sequence, and subsequent columns are index
// consecutively.
func fillComputedColumns(cids []uint, mid uint, cols []tr.ArrayColumn, trace *tr.ArrayTrace) {
	module := trace.RawModule(mid)
	// Add all columns
	for i, col := range cols {
		dst := module.Column(cids[i])
		// Sanity checks
		if dst.Name() != col.Name() {
			mod := trace.Module(col.Context().Module()).Name()
			panic(fmt.Sprintf("misaligned computed column %s.%s during trace expansion", mod, col.Name()))
		} else if dst.Data() != nil {
			mod := trace.Module(col.Context().Module()).Name()
			panic(fmt.Sprintf("computed column %s.%s already exists in trace", mod, col.Name()))
		}
		// Looks good
		module.FillColumn(cids[i], col.Data(), col.Padding())
	}
}

// Validate that all elements of a given column fit within a given bitwidth.
func validateColumnBitWidth(bitwidth uint, col trace.Column, mod Module) error {
	var (
		biBound big.Int
	)
	// Compute 2^n
	biBound.Exp(big.NewInt(2), big.NewInt(int64(bitwidth)), nil)
	//
	for j := 0; j < int(col.Data().Len()); j++ {
		var (
			bi  big.Int
			jth = col.Get(j)
		)
		// Convert field element to bigint
		jth.BigInt(&bi)
		//
		if bi.Cmp(&biBound) >= 0 {
			qualColName := trace.QualifiedColumnName(mod.Name(), col.Name())
			return fmt.Errorf("row %d of column %s is out-of-bounds (%s)", j, qualColName, jth.String())
		}
	}
	// success
	return nil
}
