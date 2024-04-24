package table

import (
	"errors"
	"fmt"
	"github.com/consensys/gnark-crypto/ecc/bls12-377/fr"
)

// Describes a set of named data columns.  Columns are not
// required to have the same height.
type Trace interface {
	// Get the value of a given column by its name.  If the column
	// does not exist or if the index is out-of-bounds then an
	// error is returned.
	//
	// NOTE: this operation is expected to be slower than
	// GetByindex as, depending on the underlying data format,
	// this may first resolve the name into a physical column
	// index.
	GetByName(name string, row int) (*fr.Element,error)
	// Get the value of a given column by its index. If the column
	// does not exist or if the index is out-of-bounds then an
	// error is returned.
	GetByIndex(col int, row int)  (*fr.Element,error)
	// Determine the height of this table, which is defined as the
	// height of the largest column.
	Height() int
}

// ===================================================================
// Array Trace
// ===================================================================

// An implementation of Trace which stores columns as an array.
type ArrayTrace struct {
	// Holds the maximum height of any column in the trace
	height int
	// Holds the name of each column
	columns []ArrayTraceColumn
}

func EmptyArrayTrace() *ArrayTrace {
	p := new(ArrayTrace)
	// Initially empty columns
	p.columns = make([]ArrayTraceColumn,0)
	// Initialise height as 0
	p.height = 0
	// done
	return p
}

func (p *ArrayTrace) HasColumn(name string) bool {
	for _,c := range p.columns {
		if c.name == name {
			return true
		}
	}
	return false
}

// Add a new column of data to this trace.
func (p *ArrayTrace) AddColumn(name string, data []*fr.Element) {
	// Construct new column
	column := ArrayTraceColumn{name,data}
	// Append it
	p.columns = append(p.columns,column)
	// Update maximum height
	if len(data) > p.height {
		p.height = len(data)
	}
}

func (p *ArrayTrace) Height() int {
	return p.height
}

func (p *ArrayTrace) GetByName(name string, row int) (*fr.Element,error) {
	// NOTE: Could improve performance here if names were kept in
	// sorted order.
	for _,c := range p.columns {
		if name == c.name {
			// Matched column
			return c.Get(row)
		}
	}
	// Failed to find column
	msg := fmt.Sprintf("Invalid column: {%s}",name)
	return nil,errors.New(msg)
}

func (p *ArrayTrace) GetByIndex(col int, row int) (*fr.Element,error) {
	if col < 0 || col >= len(p.columns) {
		return nil,errors.New("Column access out-of-bounds")
	} else {
		return p.columns[col].Get(row)
	}
}

// ===================================================================
// Array Trace Column
// ===================================================================

type ArrayTraceColumn struct {
	// Holds the name of this column
	name string
	// Holds the raw data making up this column
	data []*fr.Element
}

func (p *ArrayTraceColumn) Get(row int) (*fr.Element,error) {
	if row < 0 || row >= len(p.data) {
		return nil,errors.New("Column access out-of-bounds")
	} else {
		return p.data[row],nil
	}

}
