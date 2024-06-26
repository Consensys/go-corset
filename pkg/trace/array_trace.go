package trace

import (
	"strings"

	"github.com/consensys/gnark-crypto/ecc/bls12-377/fr"
)

// ArrayTrace provides an implementation of Trace which stores columns as an
// array.
type ArrayTrace struct {
	// Holds the maximum height of any column in the trace
	height uint
	// Holds the name of each column
	columns []Column
}

// EmptyArrayTrace constructs an empty array trace into which column data can be
// added.
func EmptyArrayTrace() *ArrayTrace {
	p := new(ArrayTrace)
	// Initially empty columns
	p.columns = make([]Column, 0)
	// Initialise height as 0
	p.height = 0
	// done
	return p
}

// NewArrayTrace constructs a new trace from a given array of columns.
func NewArrayTrace(columns []Column) (*ArrayTrace, error) {
	height := columns[0].Height()
	// for _, c := range columns {
	// 	if c.Height() != height {
	// 		return nil, errors.New("trace columns have different heights")
	// 	}
	// }
	//
	return &ArrayTrace{height, columns}, nil
}

// Width returns the number of columns in this trace.
func (p *ArrayTrace) Width() uint {
	return uint(len(p.columns))
}

// ColumnName returns the name of the ith column in this trace.
func (p *ArrayTrace) ColumnName(index int) string {
	return p.columns[index].Name()
}

// ColumnIndex returns the column index of the column with the given name in
// this trace, or returns false if no such column exists.
func (p *ArrayTrace) ColumnIndex(name string) (uint, bool) {
	for i, c := range p.columns {
		if c.Name() == name {
			return uint(i), true
		}
	}
	// Column does not exist
	return 0, false
}

// Columns returns the set of columns in this trace.  Observe that mutating the
// returned array will mutate the trace.
func (p *ArrayTrace) Columns() []Column {
	return p.columns
}

// ColumnByIndex looks up a column based on its index.
func (p *ArrayTrace) ColumnByIndex(index uint) Column {
	return p.columns[index]
}

// ColumnByName looks up a column based on its name.  If the column doesn't
// exist, then nil is returned.
func (p *ArrayTrace) ColumnByName(name string) Column {
	for _, c := range p.columns {
		if name == c.Name() {
			// Matched column
			return c
		}
	}

	return nil
}

// HasColumn checks whether the trace has a given named column (or not).
func (p *ArrayTrace) HasColumn(name string) bool {
	_, ok := p.ColumnIndex(name)
	return ok
}

// Clone creates an identical clone of this trace.
func (p *ArrayTrace) Clone() Trace {
	clone := new(ArrayTrace)
	clone.columns = make([]Column, len(p.columns))
	clone.height = p.height
	//
	for i, c := range p.columns {
		clone.columns[i] = c.Clone()
	}
	// done
	return clone
}

// Add adds a new column of data to this trace.
func (p *ArrayTrace) Add(column Column) {
	// Sanity check the column does not already exist.
	if p.HasColumn(column.Name()) {
		panic("column already exists")
	}
	// Append it
	p.columns = append(p.columns, column)
	// Update maximum height
	if column.Height() > p.height {
		p.height = column.Height()
	}
}

// AddColumn adds a new column of data to this trace.
func (p *ArrayTrace) AddColumn(name string, data []*fr.Element, padding *fr.Element) {
	// Sanity check the column does not already exist.
	if p.HasColumn(name) {
		panic("column already exists")
	}
	// Construct new column
	column := FieldColumn{name, data, padding}
	// Append it
	p.columns = append(p.columns, &column)
	// Update maximum height
	if uint(len(data)) > p.height {
		p.height = uint(len(data))
	}
}

// Height determines the maximum height of any column within this trace.
func (p *ArrayTrace) Height() uint {
	return p.height
}

// Pad each column in this trace with n items at the front.  An iterator over
// the padding values to use for each column must be given.
func (p *ArrayTrace) Pad(n uint) {
	for _, c := range p.columns {
		c.Pad(n)
	}
	// Increment height
	p.height += n
}

// Swap the order of two columns in this trace.  This is needed, in
// particular, for alignment.
func (p *ArrayTrace) Swap(l uint, r uint) {
	tmp := p.columns[l]
	p.columns[l] = p.columns[r]
	p.columns[r] = tmp
}

func (p *ArrayTrace) String() string {
	// Use string builder to try and make this vaguely efficient.
	var id strings.Builder

	id.WriteString("{")

	for i := 0; i < len(p.columns); i++ {
		if i != 0 {
			id.WriteString(",")
		}

		id.WriteString(p.columns[i].Name())
		id.WriteString("={")

		for j := 0; j < int(p.height); j++ {
			jth := p.columns[i].Get(j)

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
