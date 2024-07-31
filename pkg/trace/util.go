package trace

import (
	"fmt"
	"strings"
)

// MaxHeight determines the maximum height of any column in the trace.  This is
// useful in some scenarios for bounding the number of rows for any column.
// This is done by computing the maximum height of any module.
func MaxHeight(tr Trace) uint {
	h := uint(0)
	// Iterate over modules
	for i := uint(0); i < tr.Width(); i++ {
		ctx := tr.Column(i).Context()
		h = max(h, tr.Height(ctx))
	}
	// Done
	return h
}

// QualifiedColumnNamesToCommaSeparatedString produces a suitable string for use
// in error messages from a list of one or more column identifies.
func QualifiedColumnNamesToCommaSeparatedString(columns []uint, trace Trace) string {
	var names strings.Builder

	for i, c := range columns {
		if i != 0 {
			names.WriteString(",")
		}

		names.WriteString(trace.Column(c).Name())
	}
	// Done
	return names.String()
}

// QualifiedColumnName returns the fully qualified name of a given column.
func QualifiedColumnName(module string, column string) string {
	if module == "" {
		return column
	}

	return fmt.Sprintf("%s.%s", module, column)
}
