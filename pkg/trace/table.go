package trace

import (
	"fmt"
	"errors"
	"math/big"
)

type Table interface {
	// Check that every constraint holds for every row of this table.
	Check() error

	// Get the value of a given column by its name.  If the column
	// does not exist or if the index is out-of-bounds then an
	// error is returned.
	//
	// NOTE: this operation is expected to be slower than
	// GetByindex as, depending on the underlying data format,
	// this may first resolve the name into a physical column
	// index.
	GetByName(name string, row int) (*big.Int,error)

	// Get the value of a given column by its index. If the column
	// does not exist or if the index is out-of-bounds then an
	// error is returned.
	GetByIndex(col int, row int)  (*big.Int,error)

	// Add a new constraint to this table.
	AddConstraint(constraint Constraint)

	// Determine the height of this table, which is defined as the
	// height of the largest column.
	Height() int
}

// =============================================================================
// Lazy Table
// =============================================================================

// A table which lazily evaluates its computed columns when they are
// accessed.  This is less efficient, perhaps, than doing it strictly
// upfront.  But, for the purposes of testing, it is sufficient.
type LazyTable struct {
	height int
	// Column array (either data or computed).  Columns are stored
	// such that the dependencies of a column always come before
	// that column (i.e. have a lower index).  Thus, data columns
	// always precede computed columns, etc.
	columns []Column
	// Constaint array.
	constraints []Constraint
}

func EmptyLazyTable() *LazyTable {
	p := new(LazyTable)
	// Initially empty columns
	p.columns = make([]Column,0)
	// Initially empty constraints
	p.constraints = make([]Constraint,0)
	// Initialise height as 0
	return p
}

// Construct a new LazyTable initialised with a given set of columns
// and constraints.
func NewLazyTable(columns []Column, constraints []Constraint) *LazyTable {
	p := new(LazyTable)
	p.columns = columns
	p.constraints = constraints
	// initialise height
	for _,c := range columns {
		if c.Height() > p.height { p.height = c.Height() }
	}
	//
	return p
}

// Check whether all constraints on the given table evaluate to zero.
// If not, produce an error.
func (p *LazyTable) Check() error {
	for _,c := range p.constraints {
		err := c.Check(p)
		if err != nil {
			return err
		}
	}
	return nil
}

func (p *LazyTable) AddConstraint(constraint Constraint) {
	p.constraints = append(p.constraints,constraint)
}

func (p *LazyTable) Height() int {
	return p.height
}

func (p *LazyTable) GetByName(name string, row int) (*big.Int,error) {
	// NOTE: Could improve performance here if names were kept in
	// sorted order.
	for _,c := range p.columns {
		if name == c.Name() {
			// Matched column
			return c.Get(row)
		}
	}
	// Failed to find column
	msg := fmt.Sprintf("Invalid column: {%s}",name)
	return nil,errors.New(msg)
}

func (p *LazyTable) GetByIndex(col int, row int) (*big.Int,error) {
	if col < 0 || col >= len(p.columns) {
		return nil,errors.New("Column access out-of-bounds")
	} else {
		return p.columns[col].Get(row)
	}
}
