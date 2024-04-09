package trace

import (
	"errors"
	"fmt"
	"math/big"
)

type Constraint interface {
	// Get the handle for this constraint (i.e. its name).
	GetHandle() string
}

type Table interface {
	// GetByName gets the value of a given column by its name.  If the column
	// does not exist or if the index is out-of-bounds then an
	// error is returned.
	//
	// NOTE: this operation is expected to be slower than
	// GetByIndex as, depending on the underlying data format,
	// this may first resolve the name into a physical column
	// index.
	GetByName(name string, row int) (*big.Int, error)

	// GetByIndex gets the value of a given column by its index. If the column
	// does not exist or if the index is out-of-bounds then an
	// error is returned.
	GetByIndex(col int, row int) (*big.Int, error)

	// Height gets the number of rows in this table
	Height() int
}

// =============================================================================
// Lazy Table
// =============================================================================

// A table which lazily evaluates its computed columns when they are
// accessed.  This is less efficient, perhaps, than doing it strictly
// upfront.  But, for the purposes of testing, it is sufficient.
type LazyTable struct {
	// Index of columns
	columns []string
	// A mapping from column names to data arrays.
	rows [][]*big.Int
}

func EmptyLazyTable() *LazyTable {
	p := new(LazyTable)
	// Initially columns empty
	p.columns = make([]string, 0)
	// Initially columns empty
	p.rows = make([][]*big.Int, 0)
	//
	return p
}

// NewLazyTable constructs a new LazyTable initialised with a given schema and
// corresponding data.  Observe that this operation can fail if the
// schema and data are mal-formed.  For example, if data for one or
// more columns is missing; likewise, if there is data for non-existant
// columns; finally, if some of the columns have differing height.
func NewLazyTable(columns []string, data ...[]*big.Int) (*LazyTable, error) {
	if len(columns) != len(data) {
		return nil, errors.New("Column data does not match schema")
	} else if len(columns) > 0 {
		// Sanity check data columns all have same height.
		n := len(data[0])
		for _, d := range data {
			if len(d) != n {
				return nil, errors.New("Column data has differing heights")
			}
		}
	}
	// At this point, we are happy.
	p := new(LazyTable)
	p.columns = columns
	p.rows = data
	// Done
	return p, nil
}

// AddColumn adds a given column to a lazy table.
func (p *LazyTable) AddColumn(name string, data []*big.Int) {
	p.columns = append(p.columns, name)
	p.rows = append(p.rows, data)
}

func (p *LazyTable) Height() int {
	if len(p.rows) == 0 {
		return 0
	} else {
		return len(p.rows[0])
	}
}

func (p *LazyTable) GetByName(name string, row int) (*big.Int, error) {
	// NOTE: Could improve performance here if names were kept in
	// sorted order.
	for i, n := range p.columns {
		if n == name {
			// Matched column
			return p.GetByIndex(i, row)
		}
	}
	// Failed to find column
	msg := fmt.Sprintf("Invalid column: {%s}", name)

	return nil, errors.New(msg)
}

func (p *LazyTable) GetByIndex(col int, row int) (*big.Int, error) {
	if row < 0 || row >= p.Height() {
		return nil, errors.New("Column access out-of-bounds")
	}

	return p.rows[col][row], nil
}
