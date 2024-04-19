package trace

import (
	"fmt"
	"errors"
	"math/big"
)

type Trace interface {
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
	// Determine the height of this table, which is defined as the
	// height of the largest column.
	Height() int
}

// =============================================================================
// Table
// =============================================================================

type Table[C any] interface {
	// Check whether all constraints for a given trace evaluate to zero.
	// If not, produce an error.
	Check() error
	// Access Columns
	Columns() []Column
	// Access Constraints
	Constraints() []C
	// Add a new column to this table.
	AddColumn(column Column)
	// Add a new constraint to this table.
	AddConstraint(constraint C)
}

// =============================================================================
// Lazy Table
// =============================================================================

// A table which lazily evaluates its computed columns when they are
// accessed.  This is less efficient, perhaps, than doing it strictly
// upfront.  But, for the purposes of testing, it is sufficient.
type LazyTable[C Constraint] struct {
	height int
	// Column array (either data or computed).  Columns are stored
	// such that the dependencies of a column always come before
	// that column (i.e. have a lower index).  Thus, data columns
	// always precede computed columns, etc.
	columns []Column
	// Constaint array.
	constraints []C
}

func EmptyLazyTable[C Constraint]() *LazyTable[C] {
	p := new(LazyTable[C])
	// Initially empty columns
	p.columns = make([]Column,0)
	// Initially empty constraints
	p.constraints = make([]C,0)
	// Initialise height as 0
	return p
}

// Construct a new LazyTable initialised with a given set of columns
// and constraints.
func NewLazyTable[C Constraint](columns []Column, constraints []C) *LazyTable[C] {
	p := new(LazyTable[C])
	p.columns = columns
	p.constraints = constraints
	// initialise height
	for _,c := range columns {
		if c.MinHeight() > p.height { p.height = c.MinHeight() }
	}
	//
	return p
}

// Check whether all constraints on the given table evaluate to zero.
// If not, produce an error.
func (p *LazyTable[C]) Check() error {
	for _,c := range p.constraints {
		err := c.Check(p)
		if err != nil { return err }
	}
	return nil
}

func (p *LazyTable[C]) Columns() []Column {
	return p.columns
}

func (p *LazyTable[C]) Constraints() []C {
	return p.constraints
}

func (p *LazyTable[C]) AddConstraint(constraint C) {
	p.constraints = append(p.constraints,constraint)
}

func (p *LazyTable[C]) AddColumn(column Column) {
	p.columns = append(p.columns,column)
	// Update maximum height
	if column.MinHeight() > p.height {
		p.height = column.MinHeight()
	}
}

func (p *LazyTable[C]) Height() int {
	return p.height
}

func (p *LazyTable[C]) GetByName(name string, row int) (*big.Int,error) {
	// NOTE: Could improve performance here if names were kept in
	// sorted order.
	for _,c := range p.columns {
		if name == c.Name() {
			// Matched column
			return c.Get(row,p)
		}
	}
	// Failed to find column
	msg := fmt.Sprintf("Invalid column: {%s}",name)
	return nil,errors.New(msg)
}

func (p *LazyTable[C]) GetByIndex(col int, row int) (*big.Int,error) {
	if col < 0 || col >= len(p.columns) {
		return nil,errors.New("Column access out-of-bounds")
	} else {
		return p.columns[col].Get(row,p)
	}
}
