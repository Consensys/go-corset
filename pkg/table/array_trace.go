package table

import (
	"fmt"
	"strings"

	"github.com/consensys/gnark-crypto/ecc/bls12-377/fr"
)

// ArrayTrace provides an implementation of Trace which stores columns as an
// array.
type ArrayTrace struct {
	// Holds the maximum height of any column in the trace
	height uint
	// Holds the name of each column
	columns []*ArrayTraceColumn
}

// EmptyArrayTrace constructs an empty array trace into which column data can be
// added.
func EmptyArrayTrace() *ArrayTrace {
	p := new(ArrayTrace)
	// Initially empty columns
	p.columns = make([]*ArrayTraceColumn, 0)
	// Initialise height as 0
	p.height = 0
	// done
	return p
}

// Width returns the number of columns in this trace.
func (p *ArrayTrace) Width() uint {
	return uint(len(p.columns))
}

// ColumnName returns the name of the ith column in this trace.
func (p *ArrayTrace) ColumnName(index int) string {
	return p.columns[index].Name()
}

// IndexOf returns the index of the given name in this trace.
func (p *ArrayTrace) IndexOf(name string) (uint, bool) {
	for i, c := range p.columns {
		if c.name == name {
			return uint(i), true
		}
	}
	// Column does not exist
	return 0, false
}

// Clone creates an identical clone of this trace.
func (p *ArrayTrace) Clone() *ArrayTrace {
	clone := new(ArrayTrace)
	clone.columns = make([]*ArrayTraceColumn, len(p.columns))
	clone.height = p.height
	//
	for i, c := range p.columns {
		// TODO: can this be avoided?
		clone.columns[i] = c.Clone()
	}
	// done
	return clone
}

// HasColumn checks whether the trace has a given column or not.
func (p *ArrayTrace) HasColumn(name string) bool {
	for _, c := range p.columns {
		if c.name == name {
			return true
		}
	}

	return false
}

// Compatible determines whether or not this trace is "input compatible" with a
// given schema.  Specifically, whether or not this trace can be fed into the
// schema and expanded into a full trace.  For this to be true, the columns of
// the trace must exactly match the non-synthetic columns of the schema.
func (p *ArrayTrace) Compatible(schema Schema) error {
	index := 0
	// Check each column described in this schema is present in the trace.
	for i := uint(0); i < schema.Width(); i++ {
		group := schema.ColumnGroup(i)
		if !group.IsSynthetic() {
			for j := uint(0); j < group.Width(); j++ {
				// Determine column name
				schemaName := group.NameOf(j)
				// Check column exists in this trace
				if !p.HasColumn(schemaName) {
					return fmt.Errorf("trace missing input column %s", schemaName)
				}

				index++
			}
		}
	}
	// Check perfect match
	if index == len(p.columns) {
		// Success
		return nil
	}
	// Error case
	return fmt.Errorf("trace has %d unknown input column(s)", len(p.columns)-index)
}

// AlignWith attempts to align this trace with a given schema.  This means
// ensuring the order of columns in this trace matches the order in the schema.
// Thus, column indexes used by constraints in the schema can directly access in
// this trace (i.e. without name lookup).  Alignment can fail, however, if there
// is a mismatch between columns in the trace and those expected by the schema.
func (p *ArrayTrace) AlignWith(schema Schema) error {
	ncols := len(p.columns)
	index := 0
	// Check each column described in this schema is present in the trace.
	for i := uint(0); i < schema.Width(); i++ {
		group := schema.ColumnGroup(i)
		for j := uint(0); j < group.Width(); j++ {
			// Determine column name
			schemaName := group.NameOf(j)
			// Sanity check column exists
			if index >= ncols {
				return fmt.Errorf("trace missing column %s", schemaName)
			}

			traceName := p.columns[index].name
			// Check alignment
			if traceName != schemaName {
				// Not aligned --- so fix
				k, ok := p.IndexOf(schemaName)
				// check exists
				if !ok {
					return fmt.Errorf("trace missing column %s", schemaName)
				}
				// Swap columns
				tmp := p.columns[index]
				p.columns[index] = p.columns[k]
				p.columns[k] = tmp
			}
			// Continue
			index++
		}
	}
	// Check whether all columns matched
	if index == ncols {
		// Yes, alignment complete.
		return nil
	}
	// Error Case.
	unknowns := p.columns[index:]
	//
	return fmt.Errorf("trace contains unknown columns: %v", unknowns)
}

// AddColumn adds a new column of data to this trace.
func (p *ArrayTrace) AddColumn(name string, data []*fr.Element, padding *fr.Element) {
	// Sanity check the column does not already exist.
	if p.HasColumn(name) {
		panic("column already exists")
	}
	// Construct new column
	column := ArrayTraceColumn{name, data, padding}
	// Append it
	p.columns = append(p.columns, &column)
	// Update maximum height
	if uint(len(data)) > p.height {
		p.height = uint(len(data))
	}
}

// Columns returns the set of columns in this trace.
func (p *ArrayTrace) Columns() []*ArrayTraceColumn {
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
		if name == c.name {
			// Matched column
			return c
		}
	}

	return nil
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

func (p *ArrayTrace) String() string {
	// Use string builder to try and make this vaguely efficient.
	var id strings.Builder

	id.WriteString("{")

	for i := 0; i < len(p.columns); i++ {
		if i != 0 {
			id.WriteString(",")
		}

		id.WriteString(p.columns[i].name)
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

// ===================================================================
// Array Trace Column
// ===================================================================

// ArrayTraceColumn represents a column of data within an array trace.
type ArrayTraceColumn struct {
	// Holds the name of this column
	name string
	// Holds the raw data making up this column
	data []*fr.Element
	// Value to be used when padding this column
	padding *fr.Element
}

// Name returns the name of the given column.
func (p *ArrayTraceColumn) Name() string {
	return p.name
}

// Height determines the height of this column.
func (p *ArrayTraceColumn) Height() uint {
	return uint(len(p.data))
}

// Padding returns the value which will be used for padding this column.
func (p *ArrayTraceColumn) Padding() *fr.Element {
	return p.padding
}

// Data returns the data for the given column.
func (p *ArrayTraceColumn) Data() []*fr.Element {
	return p.data
}

// Get the value at a given row in this column.  If the row is
// out-of-bounds, then the column's padding value is returned instead.
// Thus, this function always succeeds.
func (p *ArrayTraceColumn) Get(row int) *fr.Element {
	if row < 0 || row >= len(p.data) {
		// out-of-bounds access
		return p.padding
	}
	// in-bounds access
	return p.data[row]
}

// Clone an ArrayTraceColumn
func (p *ArrayTraceColumn) Clone() *ArrayTraceColumn {
	clone := new(ArrayTraceColumn)
	clone.name = p.name
	clone.data = make([]*fr.Element, len(p.data))
	clone.padding = p.padding
	copy(clone.data, p.data)

	return clone
}

// Pad this column with n copies of a given value, either at the front
// (sign=true) or the back (sign=false).
func (p *ArrayTraceColumn) Pad(n uint) {
	// Allocate sufficient memory
	ndata := make([]*fr.Element, uint(len(p.data))+n)
	// Copy over the data
	copy(ndata[n:], p.data)
	// Go padding!
	for i := uint(0); i < n; i++ {
		ndata[i] = p.padding
	}
	// Copy over
	p.data = ndata
}
