package schema

import (
	"fmt"

	"github.com/consensys/gnark-crypto/ecc/bls12-377/fr"
	"github.com/consensys/go-corset/pkg/trace"
	tr "github.com/consensys/go-corset/pkg/trace"
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
}

// NewTraceBuilder constructs a default trace builder.  The idea is that this
// could then be customized as needed following the builder pattern.
func NewTraceBuilder(schema Schema) TraceBuilder {
	return TraceBuilder{schema, true, 0, false}
}

// Expand updates a given builder configuration to perform trace expansion (or
// not).
func (tb TraceBuilder) Expand(flag bool) TraceBuilder {
	return TraceBuilder{tb.schema, flag, tb.padding, tb.parallel}
}

// Padding updates a given builder configuration to use a given amount of padding
func (tb TraceBuilder) Padding(padding uint) TraceBuilder {
	return TraceBuilder{tb.schema, tb.expand, padding, tb.parallel}
}

// Build takes the given builder configuration, along with a given set of input
// columns and constructs a trace.
func (tb TraceBuilder) Build(columns []trace.RawColumn) (trace.Trace, []error) {
	tr, errs := tb.initialiseTrace(columns)

	if tr == nil {
		// Critical failure
		return nil, errs
	} else if tb.expand {
		// TODO: this is not done properly.
		padColumns(tr, requiredSpillage(tb.schema))
		// Expand trace
		if tb.parallel {
			panic("parallel trace expansion not supported")
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
	context trace.Context
	column  string
}

func (tb TraceBuilder) initialiseTrace(cols []trace.RawColumn) (*trace.ArrayTrace, []error) {
	// Initialise modules
	modules, modmap := tb.initialiseTraceModules()
	// Initialise columns
	columns, colmap := tb.initialiseTraceColumns()
	// Construct (empty) trace
	tr := trace.NewArrayTrace(modules, columns)
	// Fill trace.  Note that all filling errors are non-critical.
	errs := fillTraceColumns(modmap, colmap, cols, tr)
	// Finally, sanity check all input columns provided
	ninputs := tb.schema.InputColumns().Count()
	//
	for i := uint(0); i < ninputs; i++ {
		ith := columns[i]
		if ith.Data() == nil {
			mod := tb.schema.Modules().Nth(ith.Context().Module()).name
			// Missing an input column is a critical unrecoverable failure
			err := fmt.Errorf("missing input column '%s.%s' in trace", mod, ith.Name())

			return nil, append(errs, err)
		}
	}
	// Done
	return tr, errs
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
		modules[i] = trace.EmptyArrayModule(m.name)
		// Sanity check module
		if _, ok := modmap[m.name]; ok {
			panic(fmt.Sprintf("duplicate module '%s' in schema", m.name))
		}

		modmap[m.name] = i
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
		colkey := columnKey{c.Context(), c.Name()}
		// Initially column data and padding are nil.  In some cases, we will
		// populate this information from the cols array.  However, in other
		// cases, it will need to be populated during trace expansion.
		columns[i] = trace.EmptyArrayColumn(c.Context(), c.Name())
		// Sanity check column
		if _, ok := colmap[colkey]; ok {
			mod := tb.schema.Modules().Nth(c.Context().Module())
			panic(fmt.Sprintf("duplicate column '%s' in schema", trace.QualifiedColumnName(mod.name, c.Name())))
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
	var zero fr.Element = fr.NewElement((0))
	// Errs contains the set of filling errors which are accumulated
	var errs []error
	// Assign data from each input column given
	for _, c := range cols {
		// Lookup the module
		mid, ok := modmap[c.Module]
		if !ok {
			errs = append(errs, fmt.Errorf("unknown module '%s' in trace", c.Module))
		} else {
			// We assume (for now) that user-provided columns always have a length
			// multiplier of 1.  In general, this will be true.  However, in situations
			// where we are importing expanded traces, then this might not be true.
			context := trace.NewContext(mid, 1)
			// Determine enclosiong module height
			cid, ok := colmap[columnKey{context, c.Name}]
			// More sanity checks
			if !ok {
				errs = append(errs, fmt.Errorf("unknown column '%s' in trace", c.QualifiedName()))
			} else if tr.Column(cid).Data() != nil {
				errs = append(errs, fmt.Errorf("duplicate column '%s' in trace", c.QualifiedName()))
			} else {
				// Assign data
				tr.FillColumn(cid, c.Data, &zero)
			}
		}
	}
	//
	return errs
}

// RequiredSpillage returns the minimum amount of spillage required to ensure
// valid traces are accepted in the presence of arbitrary padding.  Spillage can
// only arise from computations as this is where values outside of the user's
// control are determined.
func requiredSpillage(schema Schema) uint {
	// Ensures always at least one row of spillage (referred to as the "initial
	// padding row")
	mx := uint(1)
	// Determine if any more spillage required
	for i := schema.Assignments(); i.HasNext(); {
		// Get ith assignment
		ith := i.Next()
		// Incorporate its spillage requirements
		mx = max(mx, ith.RequiredSpillage())
	}

	return mx
}

// sequentialTraceExpansion expands a given trace according to a given schema.
// More specifically, that means computing the actual values for any
// assignments.  This is done using a straightforward sequential algorithm.
func sequentialTraceExpansion(schema Schema, trace *tr.ArrayTrace) error {
	// Column identifiers for computed columns start immediately following the
	// designated input columns.
	cid := schema.InputColumns().Count()
	// Compute each assignment in turn
	for i, j := schema.Assignments(), uint(0); i.HasNext(); j++ {
		// Get ith assignment
		ith := i.Next()
		// Compute ith assignment(s)
		cols, err := ith.ComputeColumns(trace)
		// Check error
		if err != nil {
			return err
		}
		// Add all columns
		for _, col := range cols {
			dst := trace.Column(cid)
			// Sanity checks
			if dst.Context() != col.Context() || dst.Name() != col.Name() {
				mod := schema.Modules().Nth(col.Context().Module()).name
				return fmt.Errorf("misaligned computed column %s.%s during trace expansion", mod, col.Name())
			} else if dst.Data() != nil {
				mod := schema.Modules().Nth(col.Context().Module()).name
				return fmt.Errorf("computed column %s.%s already exists in trace", mod, col.Name())
			}
			// Looks good
			trace.FillColumn(cid, col.Data(), col.Padding())
			//
			cid++
		}
	}
	//
	return nil
}

// PadColumns pads every column in a given trace with a given amount of padding.
func padColumns(tr *trace.ArrayTrace, padding uint) {
	n := tr.Modules().Count()
	// Iterate over modules
	for i := uint(0); i < n; i++ {
		tr.Pad(i, padding)
	}
}
