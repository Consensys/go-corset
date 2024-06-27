package trace

import (
	"strings"

	"github.com/consensys/go-corset/pkg/util"
)

// ArrayTrace provides an implementation of Trace which stores columns as an
// array.
type ArrayTrace struct {
	// Holds the name of each column
	columns util.Array_1[Column]
}

// EmptyArrayTrace constructs an empty array trace into which column data can be
// added.
func EmptyArrayTrace() *ArrayTrace {
	p := new(ArrayTrace)
	// Initially empty columns
	p.columns = util.NewArray_1(make([]Column, 0))
	// done
	return p
}

// NewArrayTrace constructs a new trace from a given array of columns.
func NewArrayTrace(columns []Column) (*ArrayTrace, error) {
	return &ArrayTrace{util.NewArray_1(columns)}, nil
}

// Columns returns the set of columns in this trace.  Observe that mutating
// the returned array will mutate the trace.
func (p *ArrayTrace) Columns() util.Array[Column] {
	return &p.columns
}

// ColumnIndex returns the column index of the column with the given name in
// this trace, or returns false if no such column exists.
func (p *ArrayTrace) ColumnIndex(name string) (uint, bool) {
	for i := uint(0); i < p.columns.Len(); i++ {
		c := p.columns.Get(i)
		if c.Name() == name {
			return i, true
		}
	}
	// Column does not exist
	return 0, false
}

// HasColumn checks whether the trace has a given named column (or not).
func (p *ArrayTrace) HasColumn(name string) bool {
	_, ok := p.ColumnIndex(name)
	return ok
}

// Clone creates an identical clone of this trace.
func (p *ArrayTrace) Clone() Trace {
	clone := new(ArrayTrace)
	clone.columns = p.columns.Copy()
	// done
	return clone
}

// Height calculates the maximum height of any column within this trace.
func (p *ArrayTrace) Height() uint {
	h := uint(0)
	for i := uint(0); i < p.columns.Len(); i++ {
		h = max(h, p.columns.Get(i).Height())
	}
	// Done
	return h
}

// Pad each column in this trace with n items at the front.  An iterator over
// the padding values to use for each column must be given.
func (p *ArrayTrace) Pad(n uint) {
	for i := uint(0); i < p.columns.Len(); i++ {
		p.columns.Get(i).Pad(n)
	}
}

func (p *ArrayTrace) String() string {
	// Use string builder to try and make this vaguely efficient.
	var id strings.Builder

	id.WriteString("{")

	for i := uint(0); i < p.columns.Len(); i++ {
		ith := p.columns.Get(i)

		if i != 0 {
			id.WriteString(",")
		}

		id.WriteString(ith.Name())
		id.WriteString("={")

		for j := 0; j < int(ith.Height()); j++ {
			jth := ith.Get(j)

			if j != 0 {
				id.WriteString(",")
			}

			if jth == nil {
				id.WriteString("_")
			} else {
				id.WriteString(jth.String())
			}
		}
		id.WriteString("}")
	}
	id.WriteString("}")
	//
	return id.String()
}
