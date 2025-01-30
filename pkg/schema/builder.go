package schema

import (
	"fmt"
	"math"

	"github.com/consensys/gnark-crypto/ecc/bls12-377/fr"
	"github.com/consensys/go-corset/pkg/trace"
	tr "github.com/consensys/go-corset/pkg/trace"
	"github.com/consensys/go-corset/pkg/util"
	"github.com/consensys/go-corset/pkg/util/field"
)

// TraceBuilder provides a mechanical means of constructing a trace from a given
// schema and set of input columns.  The goal is to encapsulate all of the logic
// around building a trace.
type TraceBuilder struct {
	// Schema to be used when building the trace
	schema Schema
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

// NewTraceBuilder constructs a default trace builder.  The idea is that this
// could then be customized as needed following the builder pattern.
func NewTraceBuilder(schema Schema) TraceBuilder {
	return TraceBuilder{schema, true, true, true, 0, true, math.MaxUint}
}

// Defensive updates a given builder configuration to apply defensive padding
// (or not).
func (tb TraceBuilder) Defensive(flag bool) TraceBuilder {
	ntb := tb
	ntb.defensive = flag
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

// Build takes the given builder configuration, along with a given set of input
// columns and constructs a trace.
func (tb TraceBuilder) Build(columns []trace.RawColumn) (trace.Trace, []error) {
	tr, errors := tb.initialiseTrace(columns)

	if tr == nil {
		// Critical failure
		return nil, errors
	} else if tb.expand {
		// Apply spillage
		applySpillageAndDefensivePadding(tb.defensive, tr, tb.schema)
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
			if errs := sequentialTraceValidation(tb.schema, tr); errs != nil {
				return nil, append(errors, errs...)
			}
		}
	}
	// Padding
	if tb.padding > 0 {
		padColumns(tr, tb.padding)
	}

	return tr, errors
}

// A column key is used as a key for the column map
type columnKey struct {
	module uint
	column string
}

func (tb TraceBuilder) initialiseTrace(cols []trace.RawColumn) (*trace.ArrayTrace, []error) {
	// Initialise modules
	modules, modmap := tb.initialiseTraceModules()
	// Initialise columns
	columns, colmap := tb.initialiseTraceColumns()
	// Construct (empty) trace
	tr := trace.NewArrayTrace(modules, columns)
	// Fill trace.
	warnings1 := fillTraceColumns(modmap, colmap, cols, tr)
	// Validation
	err, warnings2 := checkForMissingInputColumns(tb.schema, tr)
	// Combine warnings together
	warnings := append(warnings1, warnings2...)
	//
	if err != nil {
		// Unrecoverable error
		return nil, append(warnings, err)
	}
	// Done
	return tr, warnings
}

func (tb TraceBuilder) initialiseTraceModules() ([]trace.ArrayModule, map[string]uint) {
	modmap := make(map[string]uint, 100)
	modules := make([]trace.ArrayModule, tb.schema.Modules().Count())
	// Initialise modules
	for i, iter := uint(0), tb.schema.Modules(); iter.HasNext(); i++ {
		m := iter.Next()
		// Initialise an empty module.  Such modules have an (as yet)
		// unspecified height.  For such a module to be usable, it needs at
		// least one (or more) filled columns.
		modules[i] = trace.EmptyArrayModule(m.Name)
		// Sanity check module
		if _, ok := modmap[m.Name]; ok {
			panic(fmt.Sprintf("duplicate module '%s' in schema", m.Name))
		}

		modmap[m.Name] = i
	}
	// Done
	return modules, modmap
}

func (tb TraceBuilder) initialiseTraceColumns() ([]trace.ArrayColumn, map[columnKey]uint) {
	colmap := make(map[columnKey]uint, 100)
	columns := make([]trace.ArrayColumn, tb.schema.Columns().Count())
	// Initialise columns and map
	for i, iter := uint(0), tb.schema.Columns(); iter.HasNext(); i++ {
		c := iter.Next()
		// Construct an appropriate key for this column
		colkey := columnKey{c.Context.Module(), c.Name}
		// Initially column data and padding are nil.  In some cases, we will
		// populate this information from the cols array.  However, in other
		// cases, it will need to be populated during trace expansion.
		columns[i] = trace.EmptyArrayColumn(c.Context, c.Name)
		// Sanity check column
		if _, ok := colmap[colkey]; ok {
			mod := tb.schema.Modules().Nth(c.Context.Module())
			panic(fmt.Sprintf("duplicate column '%s' in schema", trace.QualifiedColumnName(mod.Name, c.Name)))
		}
		// All clear
		colmap[colkey] = i
	}
	// Done
	return columns, colmap
}

// Fill columns in the corresponding trace from the given input columns
func fillTraceColumns(modmap map[string]uint, colmap map[columnKey]uint,
	cols []trace.RawColumn, tr *trace.ArrayTrace) []error {
	var zero fr.Element = fr.NewElement(0)
	// Errs contains the set of filling errors which are accumulated
	var errs []error
	// Assign data from each input column given
	for _, c := range cols {
		// Lookup the module
		mid, ok := modmap[c.Module]
		if !ok {
			errs = append(errs, fmt.Errorf("unknown module '%s' in trace", c.Module))
		} else {
			// Determine enclosiong module height
			cid, ok := colmap[columnKey{mid, c.Name}]
			// More sanity checks
			if !ok {
				errs = append(errs, fmt.Errorf("unknown column '%s' in trace", c.QualifiedName()))
			} else if tr.Column(cid).Data() != nil {
				errs = append(errs, fmt.Errorf("duplicate column '%s' in trace", c.QualifiedName()))
			} else {
				// Assign data
				tr.FillColumn(cid, c.Data, zero)
			}
		}
	}
	//
	return errs
}

func checkForMissingInputColumns(schema Schema, tr *trace.ArrayTrace) (error, []error) {
	var zero fr.Element = fr.NewElement(0)
	// Determine how many input columns to expect
	ninputs := schema.InputColumns().Count()
	warnings := []error{}
	// Finally, sanity check all input columns provided
	for i := uint(0); i < ninputs; i++ {
		ith := tr.Column(i)
		if ith.Data() == nil {
			// This looks suspect
			mid := ith.Context().Module()
			mod := schema.Modules().Nth(mid).Name
			err := fmt.Errorf("missing input column '%s.%s' in trace", mod, ith.Name())
			mod_height := tr.Height(ith.Context())
			// Check whether we have other columns for this module
			if mod_height != math.MaxUint && mod_height != 0 {
				// Yes, this is not recoverable.
				return err, nil
			}
			// Ok, treat as warning
			warnings = append(warnings, err)
			// Fill with a column of height zero.
			tr.FillColumn(i, field.NewFrArray(0, 256), zero)
		}
	}
	// Done
	return nil, warnings
}

// pad each module with its given level of spillage and (optionally) ensure a
// given level of defensive padding.
func applySpillageAndDefensivePadding(defensive bool, tr *trace.ArrayTrace, schema Schema) {
	n := tr.Modules().Count()
	// Iterate over modules
	for i := uint(0); i < n; i++ {
		padding := RequiredSpillage(i, schema)
		//
		if defensive {
			// determine minimum levels of defensive padding required.
			padding = max(padding, DefensivePadding(i, schema))
		}
		//
		tr.Pad(i, padding, 0)
	}
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

// sequentialTraceExpansion expands a given trace according to a given schema.
// More specifically, that means computing the actual values for any
// assignments.  This is done using a straightforward sequential algorithm.
func sequentialTraceExpansion(schema Schema, trace *tr.ArrayTrace) error {
	var err error
	// Column identifiers for computed columns start immediately following the
	// designated input columns.
	cid := schema.InputColumns().Count()
	// Compute each assignment in turn
	for i, j := schema.Assignments(), uint(0); i.HasNext(); j++ {
		var cols []tr.ArrayColumn
		// Get ith assignment
		ith := i.Next()
		// Compute ith assignment(s)
		if cols, err = ith.ComputeColumns(trace); err != nil {
			return err
		}
		// Fill all computed columns
		fillComputedColumns(cid, cols, trace)
		// Advance column id past this assignment
		cid += ith.Columns().Count()
	}
	// Done
	return nil
}

// Perform trace expansion using concurrently executing jobs.  The chosen
// algorithm operates in waves, rather than using an continuous approach.  This
// is for two reasons: firstly, the latter would require locks that would slow
// down evaluation performance; secondly, the vast majority of jobs are run in
// the very first wave.
func parallelTraceExpansion(batchsize uint, schema Schema, trace *tr.ArrayTrace) error {
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
func dispatchReadyAssignments(batchsize uint, ninputs uint, schema Schema,
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

// Fill a set of columns with their computed results.  The column index is that
// of the first column in the sequence, and subsequent columns are index
// consecutively.
func fillComputedColumns(cid uint, cols []tr.ArrayColumn, trace *tr.ArrayTrace) {
	// Add all columns
	for _, col := range cols {
		dst := trace.Column(cid)
		// Sanity checks
		if dst.Context() != col.Context() || dst.Name() != col.Name() {
			mod := trace.Modules().Nth(col.Context().Module()).Name()
			panic(fmt.Sprintf("misaligned computed column %s.%s during trace expansion", mod, col.Name()))
		} else if dst.Data() != nil {
			mod := trace.Modules().Nth(col.Context().Module()).Name()
			panic(fmt.Sprintf("computed column %s.%s already exists in trace", mod, col.Name()))
		}
		// Looks good
		trace.FillColumn(cid, col.Data(), col.Padding())
		//
		cid++
	}
}

// Validate that values held in trace columns match the expected type.  This is
// really a sanity check that the trace is not malformed.
func parallelTraceValidation(schema Schema, tr tr.Trace) []error {
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

// Validate that values held in trace columns match the expected type.  This is
// really a sanity check that the trace is not malformed.
func sequentialTraceValidation(schema Schema, tr tr.Trace) []error {
	var errors []error

	schemaCols := schema.Columns()
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
		errors = append(errors, validateColumnType(colType, col, mod))
	}
	// Done
	return errors
}

// Validate that all elements of a given column are within the given type.
func validateColumnType(colType Type, col tr.Column, mod Module) error {
	for j := 0; j < int(col.Data().Len()); j++ {
		jth := col.Get(j)
		if !colType.Accept(jth) {
			qualColName := tr.QualifiedColumnName(mod.Name, col.Name())
			return fmt.Errorf("row %d of column %s is out-of-bounds (%s)", j, qualColName, jth.String())
		}
	}
	// success
	return nil
}
