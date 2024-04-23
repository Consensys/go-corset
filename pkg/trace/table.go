package trace

import (
	"fmt"
	"errors"
	"github.com/consensys/gnark-crypto/ecc/bls12-377/fr"
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
	GetByName(name string, row int) (*fr.Element,error)
	// Get the value of a given column by its index. If the column
	// does not exist or if the index is out-of-bounds then an
	// error is returned.
	GetByIndex(col int, row int)  (*fr.Element,error)
	// Determine the height of this table, which is defined as the
	// height of the largest column.
	Height() int
}

// =============================================================================
// Table
// =============================================================================

type Table[C any, R any] interface {
	// Check whether all constraints for a given trace evaluate to zero.
	// If not, produce an error.
	Check() error
	// Check whether a given column already exists
	HasColumn(string) bool
	// Access Columns
	Columns() []C
	// Access Constraints
	Constraints() []R
	// Add a new column to this table.
	AddColumn(column C)
	// Add a new constraint to this table.
	AddConstraint(constraint R)
}

// =============================================================================
// Lazy Table
// =============================================================================

// A table which lazily evaluates its computed columns when they are
// accessed.  This is less efficient, perhaps, than doing it strictly
// upfront.  But, for the purposes of testing, it is sufficient.
type LazyTable[C Column, R Constraint] struct {
	height int
	// Column array (either data or computed).  Columns are stored
	// such that the dependencies of a column always come before
	// that column (i.e. have a lower index).  Thus, data columns
	// always precede computed columns, etc.
	columns []C
	// Constaint array.
	constraints []R
}

func EmptyLazyTable[C Column, R Constraint]() *LazyTable[C,R] {
	p := new(LazyTable[C,R])
	// Initially empty columns
	p.columns = make([]C,0)
	// Initially empty constraints
	p.constraints = make([]R,0)
	// Initialise height as 0
	return p
}

// Construct a new LazyTable initialised with a given set of columns
// and constraints.
func NewLazyTable[C Column, R Constraint](columns []C, constraints []R) *LazyTable[C,R] {
	p := new(LazyTable[C,R])
	p.columns = columns
	p.constraints = constraints
	// initialise height
	for _,c := range columns {
		if c.MinHeight() > p.height { p.height = c.MinHeight() }
	}
	//
	return p
}

func (p *LazyTable[C, R]) HasColumn(name string) bool {
	for _,c := range p.columns {
		if c.Name() == name {
			return true
		}
	}
	return false
}

// Check whether all constraints on the given table evaluate to zero.
// If not, produce an error.
func (p *LazyTable[C,R]) Check() error {
	for _,c := range p.constraints {
		err := c.Check(p)
		if err != nil { return err }
	}
	return nil
}

func (p *LazyTable[C, R]) Columns() []C {
	return p.columns
}

func (p *LazyTable[C, R]) Constraints() []R {
	return p.constraints
}

func (p *LazyTable[C, R]) AddConstraint(constraint R) {
	p.constraints = append(p.constraints,constraint)
}

func (p *LazyTable[C, R]) AddColumn(column C) {
	p.columns = append(p.columns,column)
	// Update maximum height
	if column.MinHeight() > p.height {
		p.height = column.MinHeight()
	}
}

func (p *LazyTable[C,R]) Height() int {
	return p.height
}

func (p *LazyTable[C,R]) GetByName(name string, row int) (*fr.Element,error) {
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

func (p *LazyTable[C,R]) GetByIndex(col int, row int) (*fr.Element,error) {
	if col < 0 || col >= len(p.columns) {
		return nil,errors.New("Column access out-of-bounds")
	} else {
		return p.columns[col].Get(row,p)
	}
}
