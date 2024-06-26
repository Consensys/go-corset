package lt

// TraceFile represents a single LT trace file, and is made up from zero or more
// columns.
type TraceFile struct {
	columns []*Column
}

// Width returns the number of columns in this trace file.
func (p *TraceFile) Width() uint {
	return uint(len(p.columns))
}

// Column returns the ith column in this trace file.
func (p *TraceFile) Column(i uint) *Column {
	return p.columns[i]
}
