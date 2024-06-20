package schema

import (
	"fmt"

	tr "github.com/consensys/go-corset/pkg/trace"
)

// AlignInputs attempts to align this trace with the input columns of a given
// schema.  This means ensuring the order of columns in this trace matches the
// order of input columns in the schema. Thus, column indexes used by
// constraints in the schema can directly access in this trace (i.e. without
// name lookup). Alignment can fail, however, if there is a mismatch between
// columns in the trace and those expected by the schema.
func AlignInputs(p tr.Trace, schema Schema) error {
	return alignWith(false, p, schema)
}

// Align attempts to align this trace with a given schema.  This means ensuring
// the order of columns in this trace matches the order in the schema. Thus,
// column indexes used by constraints in the schema can directly access in this
// trace (i.e. without name lookup).  Alignment can fail, however, if there is a
// mismatch between columns in the trace and those expected by the schema.
func Align(p tr.Trace, schema Schema) error {
	return alignWith(true, p, schema)
}

// Alignment algorithm which operates either in unexpanded or expanded mode.  In
// expanded mode, all columns must be accounted for and will be aligned.  In
// unexpanded mode, the trace is only expected to contain input (i.e.
// non-synthetic) columns.  Furthermore, in the schema these are expected to be
// allocated before synthetic columns.  As such, alignment of these input
// columns is performed.
func alignWith(expand bool, p tr.Trace, schema Schema) error {
	ncols := p.Width()
	index := uint(0)
	// Check each column described in this schema is present in the trace.
	for i := uint(0); i < schema.Width(); i++ {
		group := schema.ColumnGroup(i)
		if expand || !group.IsSynthetic() {
			for j := uint(0); j < group.Width(); j++ {
				// Determine column name
				schemaName := group.NameOf(j)
				// Sanity check column exists
				if index >= ncols {
					return fmt.Errorf("trace missing column %s", schemaName)
				}

				traceName := p.ColumnByIndex(index).Name()
				// Check alignment
				if traceName != schemaName {
					// Not aligned --- so fix
					k, ok := p.ColumnIndex(schemaName)
					// check exists
					if !ok {
						return fmt.Errorf("trace missing column %s", schemaName)
					}
					// Swap columns
					p.Swap(index, k)
				}
				// Continue
				index++
			}
		}
	}
	// Check whether all columns matched
	if index == ncols {
		// Yes, alignment complete.
		return nil
	}
	// Error Case.
	n := ncols - index
	unknowns := make([]string, n)
	// Determine names of unknown columns.
	for i := index; i < ncols; i++ {
		unknowns[i-index] = p.ColumnByIndex(i).Name()
	}
	//
	return fmt.Errorf("trace contains unknown columns: %v", unknowns)
}
