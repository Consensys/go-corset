package trace

import "strings"

// PadColumns pads every column in a given trace with a given amount of padding.
func PadColumns(tr Trace, padding uint) {
	modules := tr.Modules()
	for i := uint(0); i < modules.Len(); i++ {
		modules.Pad(i, padding)
	}
}

// MaxHeight determines the maximum height of any column in the trace.  This is
// useful in some scenarios for bounding the number of rows for any column.
// This is done by computing the maximum height of any module.
func MaxHeight(tr Trace) uint {
	modules := tr.Modules()
	h := uint(0)
	// Iterate over modules
	for i := uint(0); i < modules.Len(); i++ {
		h = max(h, modules.Get(i).Height())
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

		names.WriteString(trace.Columns().Get(c).Name())
	}
	// Done
	return names.String()
}
