package trace

import (
	"errors"
	"fmt"
	"math/big"
)

// Table represents a table of constraints.
type Table interface {
	// Check that every constraint holds for every row of this table.
	Check() error

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

	// AddConstraint adds a new constraint to this table.
	AddConstraint(constraint Constraint)

	// Height determines the height of this table, which is defined as the
	// height of the largest column.
	Height() int
}

// =============================================================================
// Lazy Table
// =============================================================================

// LazyTable is a table which lazily evaluates its computed columns when they are
// accessed.  This is less efficient, perhaps, than doing it strictly
// upfront.  But, for the purposes of testing, it is sufficient.
type LazyTable struct {
	height int
	// Column array (either data or computed).  Columns are stored
	// such that the dependencies of a column always come before
	// that column (i.e. have a lower index).  Thus, data columns
	// always precede computed columns, etc.
	columns []Column
	// Constraint slice.
	constraints []Constraint
}

// NOTE: This is used for compile time type checking if the given type satisfies the given interface.
var _ Table = (*LazyTable)(nil)

// NewEmptyLazyTable constructs a new LazyTable initialised with empty values.
func NewEmptyLazyTable() *LazyTable {
	return &LazyTable{
		columns:     make([]Column, 0),
		constraints: make([]Constraint, 0),
	}
}

// NewLazyTable constructs a new LazyTable initialised with a given set of columns
// and constraints.
func NewLazyTable(columns []Column, constraints []Constraint) *LazyTable {
	var height int
	for _, c := range columns {
		if c.Height() > height {
			height = c.Height()
		}
	}

	return &LazyTable{
		height:      height,
		columns:     columns,
		constraints: constraints,
	}
}

// Check whether all constraints on the given table evaluate to zero.
// If not, produce an error.
func (p *LazyTable) Check() error {
	for _, c := range p.constraints {
		err := c.Check(p)
		if err != nil {
			return err
		}
	}

	return nil
}

// AddConstraint adds a constraint to the table.
func (p *LazyTable) AddConstraint(constraint Constraint) {
	p.constraints = append(p.constraints, constraint)
}

// Height returns the height of the table.
func (p *LazyTable) Height() int {
	return p.height
}

// GetByName matches a column by column name.
func (p *LazyTable) GetByName(name string, row int) (*big.Int, error) {
	// NOTE: Could improve performance here if names were kept in
	// sorted order.
	for _, c := range p.columns {
		if name == c.Name() {
			// Matched column
			return c.Get(row)
		}
	}
	// Failed to find column
	msg := fmt.Sprintf("Invalid column: {%s}", name)

	return nil, errors.New(msg)
}

// GetByIndex matches a column by column index.
func (p *LazyTable) GetByIndex(col int, row int) (*big.Int, error) {
	if col < 0 || col >= len(p.columns) {
		return nil, errors.New("column access out-of-bounds")
	}

	return p.columns[col].Get(row)
}
