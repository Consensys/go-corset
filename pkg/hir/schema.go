package hir

import (
	"github.com/consensys/go-corset/pkg/mir"
	"github.com/consensys/go-corset/pkg/table"
)

// DataColumn captures the essence of a data column at the HIR level.
type DataColumn = *table.DataColumn[table.Type]

// VanishingConstraint captures the essence of a vanishing constraint at the HIR
// level.
type VanishingConstraint = *table.VanishingConstraint[Expr]

// Schema for HIR constraints and columns.
type Schema struct {
	// The data columns of this schema.
	dataColumns []DataColumn
	// The vanishing constraints of this schema.
	vanishing []VanishingConstraint
}

// EmptySchema is used to construct a fresh schema onto which new columns and
// constraints will be added.
func EmptySchema() *Schema {
	p := new(Schema)
	p.dataColumns = make([]DataColumn, 0)
	p.vanishing = make([]VanishingConstraint, 0)
	// Done
	return p
}

// Columns returns the set of (data) columns declared within this schema.
func (p *Schema) Columns() []DataColumn {
	return p.dataColumns
}

// Constraints returns the set of (vanishing) constraints declared within this schema.
func (p *Schema) Constraints() []VanishingConstraint {
	return p.vanishing
}

// AddDataColumn appends a new data column.
func (p *Schema) AddDataColumn(name string, base table.Type) {
	p.dataColumns = append(p.dataColumns, table.NewDataColumn(name, base))
}

// AddVanishingConstraint appends a new vanishing constraint.
func (p *Schema) AddVanishingConstraint(handle string, domain *int, expr Expr) {
	p.vanishing = append(p.vanishing, table.NewVanishingConstraint(handle, domain, expr))
}

// Accepts determines whether this schema will accept a given trace.  That
// is, whether or not the given trace adheres to the schema.  A trace can fail
// to adhere to the schema for a variety of reasons, such as having a constraint
// which does not hold.
func (p *Schema) Accepts(trace table.Trace) (bool, error) {
	// Check (typed) data columns
	warning, err := table.ForallAcceptTrace(trace, p.dataColumns)
	if err != nil {
		return warning, err
	}
	// Check range constraints
	warning, err = table.ForallAcceptTrace(trace, p.vanishing)
	if err != nil {
		return warning, err
	}

	return false, nil
}

// LowerToMir lowers (or refines) an HIR table into an MIR table.  That means
// lowering all the columns and constraints, whilst adding additional columns /
// constraints as necessary to preserve the original semantics.
func (p *Schema) LowerToMir() *mir.Schema {
	mirSchema := mir.EmptySchema()
	// First, lower columns
	for _, col := range p.dataColumns {
		mirSchema.AddDataColumn(col.Name, col.Type)
	}
	// Second, lower constraints
	for _, c := range p.vanishing {
		mir_exprs := c.Expr.LowerTo()
		// Add individual constraints arising
		for _, mir_expr := range mir_exprs {
			mirSchema.AddVanishingConstraint(c.Handle, c.Domain, mir_expr)
		}
	}
	//
	return mirSchema
}
