package hir

import (
	"errors"
	"fmt"
	"github.com/consensys/go-corset/pkg/mir"
	"github.com/consensys/go-corset/pkg/table"
	"github.com/consensys/gnark-crypto/ecc/bls12-377/fr"
)

type Schema = table.Schema[Column,Constraint]

type Column interface {
	// hir.Column is-a Column
	table.Column
	// Get the type associated with this column
	Type() mir.Type
	// Lower this column to an MirColumn
	LowerTo() mir.Column
}

// ===================================================================
// Data Column
// ===================================================================

type DataColumn struct {
	name string
	base mir.Type
}

func NewDataColumn(name string, base mir.Type) *DataColumn {
       return &DataColumn{name,base}
}

func (c *DataColumn) Name() string {
       return c.name
}

func (c *DataColumn) Type() mir.Type {
       return c.base
}

func (c *DataColumn) Computable() bool {
       return false
}

func (c *DataColumn) Get(row int, tr table.Trace) (*fr.Element,error) {
       return tr.GetByName(c.name,row)
}

func (c *DataColumn) Accepts(tr table.Trace) error {
	for i := 0; i < tr.Height(); i++ {
		val,err := tr.GetByName(c.name,i)
		if err != nil { return err }
		if !c.base.Accepts(val) {
			// Construct useful error message
			msg := fmt.Sprintf("column %s value out-of-bounds (row %d, %s)",c.name,i,val)
			// Evaluation failure
			return errors.New(msg)
		}
	}
	// All good
	return nil
}

func (c *DataColumn) LowerTo() mir.Column {
	return mir.NewDataColumn(c.name,c.base)
}

// ===================================================================
// Lowering
// ===================================================================

// Lower (or refine) an HIR table into an MIR table.  That means
// lowering all the columns and constraints, whilst adding additional
// columns / constraints as necessary to preserve the original
// semantics.
func LowerToMir(hirSchema *Schema, mirSchema *mir.Schema) {
	// First, lower columns
	for _,col := range hirSchema.Columns() {
		mirSchema.AddColumn(col.LowerTo())
	}
	// Second, lower constraints
	for _,c := range hirSchema.Constraints() {
		mir_exprs := c.Expr.LowerTo()
		// Add individual constraints arising
		for _,mir_expr := range mir_exprs {
			mirSchema.AddConstraint(&mir.VanishingConstraint{Handle: c.Handle,Expr: mir_expr})
		}
	}
}
