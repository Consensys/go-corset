package trace

// PadColumns pads every column in a given trace with a given amount of padding.
func PadColumns(tr Trace, padding uint) {
	modules := tr.Modules()
	for i := uint(0); i < modules.Len(); i++ {
		modules.Pad(i, padding)
	}
	// columns := tr.Columns()
	// for i := uint(0); i < columns.Len(); i++ {
	// 	columns.Get(i).Pad(padding)
	// }
}

// MaxHeight determines the maximum height of any column in the trace.  This is
// useful in some scenarios for bounding the number of rows for any column.
// This is done by computing the maximum height of any module.
func MaxHeight(tr Trace) uint {
	modules := tr.Modules()
	h := uint(0)
	for i := uint(0); i < modules.Len(); i++ {
		h = max(h, modules.Get(i).Height())
	}
	// Done
	return h
}
