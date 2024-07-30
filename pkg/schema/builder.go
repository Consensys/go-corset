package schema

import (
	"fmt"

	"github.com/consensys/gnark-crypto/ecc/bls12-377/fr"
	"github.com/consensys/go-corset/pkg/trace"
	tr "github.com/consensys/go-corset/pkg/trace"
)

// BuildTrace converts a set of raw columns into a fully expanded trace.  This
// is a potentially expensive computation if the columns are large.
func BuildTrace(schema Schema, columns []trace.RawColumn, expand bool, padding uint) (trace.Trace, error) {
	//fmt.Printf("BUILDING TRACE: %s\n", columns)
	tr, err := internalBuildTrace(schema, columns)

	if err != nil {
		return nil, err
	} else if expand {
		// TODO: this is not done properly.
		padColumns(tr, requiredSpillage(schema))
		// Expand trace
		if err := doTraceExpansion(schema, tr); err != nil {
			return tr, err
		}
	}
	// Padding
	if padding > 0 {
		padColumns(tr, padding)
	}
	//fmt.Printf("BUILT TRACE: %s\n", tr)
	return tr, nil
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

// doTraceExpansion expands a given trace according to a given schema.  More
// specifically, that means computing the actual values for any assignments.
func doTraceExpansion(schema Schema, trace *tr.ArrayTrace) error {
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
	//	return util.ParExec[expandTraceJob](batchjobs)
	return nil
}

// ColumnIndexOf returns the column index of the column with the given name, or
// returns false if no matching column exists.
func ColumnIndexOf(schema Schema, module uint, name string) (uint, bool) {
	return schema.Columns().Find(func(c Column) bool {
		return c.Context().Module() == module && c.Name() == name
	})
}

// A column key is used as a key for the column map
type columnKey struct {
	context trace.Context
	column  string
}

func internalBuildTrace(schema Schema, cols []trace.RawColumn) (*trace.ArrayTrace, error) {
	// Initialise modules
	modules, modmap := initialiseTraceModules(schema)
	// Initialise columns
	columns, colmap := initialiseTraceColumns(schema)
	// Construct (empty) trace
	tr := trace.NewArrayTrace(modules, columns)
	// Fill trace
	if err := fillTraceColumns(modmap, colmap, cols, tr); err != nil {
		return nil, err
	}
	// Finally, sanity check all input columns provided
	ninputs := schema.InputColumns().Count()
	for i := uint(0); i < ninputs; i++ {
		ith := columns[i]
		if ith.Data() == nil {
			mod := schema.Modules().Nth(ith.Context().Module()).name
			return nil, fmt.Errorf("missing input column '%s.%s' in trace", mod, ith.Name())
		}
	}
	// Done
	return trace.NewArrayTrace(modules, columns), nil
}

func initialiseTraceModules(schema Schema) ([]trace.ArrayModule, map[string]uint) {
	modmap := make(map[string]uint, 100)
	modules := make([]trace.ArrayModule, schema.Modules().Count())
	// Initialise modules
	for i, iter := uint(0), schema.Modules(); iter.HasNext(); i++ {
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

func initialiseTraceColumns(schema Schema) ([]trace.ArrayColumn, map[columnKey]uint) {
	colmap := make(map[columnKey]uint, 100)
	columns := make([]trace.ArrayColumn, schema.Columns().Count())
	// Initialise columns and map
	for i, iter := uint(0), schema.Columns(); iter.HasNext(); i++ {
		c := iter.Next()
		// Construct an appropriate key for this column
		colkey := columnKey{c.Context(), c.Name()}
		// Initially column data and padding are nil.  In some cases, we will
		// populate this information from the cols array.  However, in other
		// cases, it will need to be populated during trace expansion.
		columns[i] = trace.EmptyArrayColumn(c.Context(), c.Name())
		// Sanity check column
		if _, ok := colmap[colkey]; ok {
			mod := schema.Modules().Nth(c.Context().Module())
			panic(fmt.Sprintf("duplicate column '%s.%s' in schema", mod, c.Name()))
		}
		// All clear
		colmap[colkey] = i
	}
	// Done
	return columns, colmap
}

// Fill columns in the corresponding trace from the given input columns
func fillTraceColumns(modmap map[string]uint, colmap map[columnKey]uint,
	cols []trace.RawColumn, tr *trace.ArrayTrace) error {
	var zero fr.Element = fr.NewElement((0))
	// Assign data from each input column given
	for _, c := range cols {
		// Lookup the module
		mid, ok := modmap[c.Module]
		if !ok {
			return fmt.Errorf("unknown module '%s' in trace", c.Module)
		}
		// We assume (for now) that user-provided columns always have a length
		// multiplier of 1.  In general, this will be true.  However, in situations
		// where we are importing expanded traces, then this might not be true.
		context := trace.NewContext(mid, 1)
		// Determine enclosiong module height
		cid, ok := colmap[columnKey{context, c.Name}]
		// More sanity checks
		if !ok {
			return fmt.Errorf("unknown column '%s.%s' in trace", c.Module, c.Name)
		} else if tr.Column(cid).Data() != nil {
			return fmt.Errorf("duplicate column '%s.%s' in trace", c.Module, c.Name)
		}
		// Assign data
		tr.FillColumn(cid, c.Data, &zero)
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
