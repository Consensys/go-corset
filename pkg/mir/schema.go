package mir

import (
	"github.com/consensys/go-corset/pkg/air"
	"github.com/consensys/go-corset/pkg/table"
	"github.com/consensys/gnark-crypto/ecc/bls12-377/fr"
)

type Schema = table.Schema[Column,Constraint]

type Column interface {
	// hir.Column is-a Column
	table.Column
	// Lower this column to an MirColumn
	LowerTo() air.Column
}

// ===================================================================
// Column
// ===================================================================

type DataColumn struct {
	name string
}

func NewDataColumn(name string) *DataColumn {
	return &DataColumn{name}
}

func (c *DataColumn) Name() string {
	return c.name
}

func (c *DataColumn) Computable() bool {
	return false
}

func (c *DataColumn) Get(row int, tr table.Trace) (*fr.Element,error) {
	return tr.GetByName(c.name,row)
}

func (c *DataColumn) Accepts(tr table.Trace) error {
	return nil
}

func (c *DataColumn) LowerTo() air.Column {
	return c
}

// ===================================================================
// Constraints
// ===================================================================


// For now, all constraints are vanishing constraints.
type Constraint = *table.VanishingConstraint[Expr]

// Lower (or refine) an MIR table into an AIR table.  That means
// lowering all the columns and constraints, whilst adding additional
// columns / constraints as necessary to preserve the original
// semantics.
func LowerToAir(mirTbl *Schema, airTbl *air.Schema) {
	for _,col := range mirTbl.Columns() {
		airTbl.AddColumn(col)
	}
	for _,c := range mirTbl.Constraints() {
		// FIXME: this is broken because its currently
		// assuming that an AirConstraint is always a
		// VanishingConstraint.  Eventually this will not be
		// true.
		air_expr := c.Expr.LowerTo(airTbl)
		airTbl.AddConstraint(&table.VanishingConstraint[air.Expr]{Handle: c.Handle,Expr: air_expr})
	}
}
