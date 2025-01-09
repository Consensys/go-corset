package schema

import (
	"fmt"
	"math"

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
	schema Schema
	// Indicates whether or not to perform trace expansion.  The default should
	// be to apply trace expansion.  However, for testing purposes, it can be
	// useful to provide an already expanded trace to ensure a set of
	// constraints correctly rejects it.
	expand bool
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
	return TraceBuilder{schema, true, 0, true, math.MaxUint}
}

// Expand updates a given builder configuration to perform trace expansion (or
// not).
func (tb TraceBuilder) Expand(flag bool) TraceBuilder {
	return TraceBuilder{tb.schema, flag, tb.padding, tb.parallel, tb.batchSize}
}

// Padding updates a given builder configuration to use a given amount of padding
func (tb TraceBuilder) Padding(padding uint) TraceBuilder {
	return TraceBuilder{tb.schema, tb.expand, padding, tb.parallel, tb.batchSize}
}

// Parallel updates a given builder configuration to allow trace expansion to be
// performed concurrently (or not).
func (tb TraceBuilder) Parallel(parallel bool) TraceBuilder {
	return TraceBuilder{tb.schema, tb.expand, tb.padding, parallel, tb.batchSize}
}

// BatchSize sets the maximum number of batches to run in parallel during trace
// expansion.
func (tb TraceBuilder) BatchSize(batchSize uint) TraceBuilder {
	return TraceBuilder{tb.schema, tb.expand, tb.padding, tb.parallel, batchSize}
}

// Build takes the given builder configuration, along with a given set of input
// columns and constructs a trace.
func (tb TraceBuilder) Build(columns []trace.RawColumn) (trace.Trace, []error) {
	tr, errs := tb.initialiseTrace(columns)

	if tr == nil {
		// Critical failure
		return nil, errs
	} else if tb.expand {
		// Apply spillage
		applySpillage(tr, tb.schema)
		// Expand trace
		if tb.parallel {
			// Run (parallel) trace expansion
			if err := parallelTraceExpansion(tb.batchSize, tb.schema, tr); err != nil {
				return nil, append(errs, err)
			}
		} else if err := sequentialTraceExpansion(tb.schema, tr); err != nil {
			// Expansion errors are fatal as well
			return nil, append(errs, err)
		}
	}
	// Padding
	if tb.padding > 0 {
		padColumns(tr, tb.padding)
	}

	return tr, errs
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
	err, warnings2 := validateTraceColumns(tb.schema, tr)
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

func validateTraceColumns(schema Schema, tr *trace.ArrayTrace) (error, []error) {
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
			tr.FillColumn(i, util.NewFrArray(0, 256), zero)
		}
	}
	// Done
	return nil, warnings
}

// applySpillage pads each module with its given level of spillage
func applySpillage(tr *trace.ArrayTrace, schema Schema) {
	n := tr.Modules().Count()
	// Iterate over modules
	for i := uint(0); i < n; i++ {
		spillage := RequiredSpillage(i, schema)
		tr.Pad(i, spillage)
	}
}

// PadColumns pads every column in a given trace with a given amount of padding.
func padColumns(tr *trace.ArrayTrace, padding uint) {
	n := tr.Modules().Count()
	// Iterate over modules
	for i := uint(0); i < n; i++ {
		tr.Pad(i, padding)
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
