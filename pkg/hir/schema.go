package hir

import (
	"errors"
	"fmt"

	"github.com/consensys/gnark-crypto/ecc/bls12-377/fr"
	"github.com/consensys/go-corset/pkg/mir"
	"github.com/consensys/go-corset/pkg/table"
)

// Schema for HIR constraints and columns.
type Schema = table.Schema[Column, Constraint]

// LowerToMir lowers (or refines) an HIR table into an MIR table.  That means
// lowering all the columns and constraints, whilst adding additional columns /
// constraints as necessary to preserve the original semantics.
func LowerToMir(hirSchema *Schema, mirSchema *mir.Schema) {
	// First, lower columns
	for _, col := range hirSchema.Columns() {
		mirSchema.AddColumn(col.LowerTo())
	}
	// Second, lower constraints
	for _, c := range hirSchema.Constraints() {
		mir_exprs := c.Expr.LowerTo()
		// Add individual constraints arising
		for _, mir_expr := range mir_exprs {
			mirSchema.AddConstraint(&mir.VanishingConstraint{Handle: c.Handle, Expr: mir_expr})
		}
	}
}

// Column captures the essence of a column at the HIR level.  Specifically, an
// MIR column can be lowered to an MIR column.
type Column interface {
	// hir.Column is-a Column
	table.Column
	// Get the type associated with this column
	Type() mir.Type
	// Lower this column to an MirColumn
	LowerTo() mir.Column
}

// DataColumn represents a column of user-provided values.
type DataColumn struct {
	name string
	base mir.Type
}

// NewDataColumn constructs a new data column with a given name and base type.
func NewDataColumn(name string, base mir.Type) *DataColumn {
	return &DataColumn{name, base}
}

// Name returns the name of this column.
func (c *DataColumn) Name() string {
	return c.name
}

// Type returns the type of this column.
func (c *DataColumn) Type() mir.Type {
	return c.base
}

// Computable determines whether or not this column can be computed from the
// existing columns of a trace.  That is, whether or not there is a known
// expression which determines the values for this column based on others in the
// trace.  Data columns are not computable.
func (c *DataColumn) Computable() bool {
	return false
}

// Get the value of this column at a given row in a given trace.
func (c *DataColumn) Get(row int, tr table.Trace) (*fr.Element, error) {
	return tr.GetByName(c.name, row)
}

// Accepts determines whether or not this column accepts the given trace.  For a
// data column, this means ensuring that all elements are value for the columns
// type.
func (c *DataColumn) Accepts(tr table.Trace) error {
	for i := 0; i < tr.Height(); i++ {
		val, err := tr.GetByName(c.name, i)
		if err != nil {
			return err
		}

		if !c.base.Accepts(val) {
			// Construct useful error message
			msg := fmt.Sprintf("column %s value out-of-bounds (row %d, %s)", c.name, i, val)
			// Evaluation failure
			return errors.New(msg)
		}
	}
	// All good
	return nil
}

// LowerTo lowers this datacolumn to the MIR level.  This has relatively little
// effect at this time.
func (c *DataColumn) LowerTo() mir.Column {
	return mir.NewDataColumn(c.name, c.base)
}
