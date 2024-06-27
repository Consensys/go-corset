package trace

// PadColumns pads every column in a given trace with a given amount of padding.
func PadColumns(tr Trace, padding uint) {
	cols := tr.Columns()
	n := cols.Len()

	for i := uint(0); i < n; i++ {
		cols.Get(i).Pad(padding)
	}
}
