package air

import (
	"github.com/consensys/gnark-crypto/ecc/bls12-377/fr"
	"github.com/consensys/go-corset/pkg/table"
)

// DataColumn captures the essence of a data column at AIR level.
type DataColumn = *table.DataColumn[*table.FieldType]

// VanishingConstraint captures the essence of a vanishing constraint at the HIR
// level.  A vanishing constraint is a row constraint which must evaluate to
// zero.
type VanishingConstraint = *table.RowConstraint[table.ZeroTest[Expr]]

// PropertyAssertion captures the notion of an arbitrary property which should
// hold for all acceptable traces.  However, such a property is not enforced by
// the prover.
type PropertyAssertion = *table.PropertyAssertion[table.Evaluable]

// Schema for AIR traces which is parameterised on a notion of computation as
// permissible in computed columns.
type Schema struct {
	// The data columns of this schema.
	dataColumns []DataColumn
	// The computed columns of this schema.
	computedColumns []*table.ComputedColumn
	// The vanishing constraints of this schema.
	vanishing []VanishingConstraint
	// The range constraints of this schema.
	ranges []*table.RangeConstraint
	// Property assertions.
	assertions []PropertyAssertion
}

// EmptySchema is used to construct a fresh schema onto which new columns and
// constraints will be added.
func EmptySchema[C table.Evaluable]() *Schema {
	p := new(Schema)
	p.dataColumns = make([]DataColumn, 0)
	p.computedColumns = make([]*table.ComputedColumn, 0)
	p.vanishing = make([]VanishingConstraint, 0)
	p.ranges = make([]*table.RangeConstraint, 0)
	p.assertions = make([]PropertyAssertion, 0)
	// Done
	return p
}

// HasColumn checks whether a given schema has a given column.
func (p *Schema) HasColumn(name string) bool {
	for _, c := range p.dataColumns {
		if c.Name == name {
			return true
		}
	}

	for _, c := range p.computedColumns {
		if c.Name == name {
			return true
		}
	}

	return false
}

// AddDataColumn appends a new data column.
func (p *Schema) AddDataColumn(name string) {
	p.dataColumns = append(p.dataColumns, table.NewDataColumn(name, &table.FieldType{}))
}

// AddComputedColumn appends a new computed column.
func (p *Schema) AddComputedColumn(name string, expr table.Evaluable) {
	p.computedColumns = append(p.computedColumns, table.NewComputedColumn(name, expr))
}

// AddVanishingConstraint appends a new vanishing constraint.
func (p *Schema) AddVanishingConstraint(handle string, domain *int, expr Expr) {
	p.vanishing = append(p.vanishing, table.NewRowConstraint(handle, domain, expr))
}

// AddRangeConstraint appends a new range constraint.
func (p *Schema) AddRangeConstraint(column string, bound *fr.Element) {
	p.ranges = append(p.ranges, table.NewRangeConstraint(column, bound))
}

// Accepts determines whether this schema will accept a given trace.  That
// is, whether or not the given trace adheres to the schema.  A trace can fail
// to adhere to the schema for a variety of reasons, such as having a constraint
// which does not hold.
func (p *Schema) Accepts(trace table.Trace) error {
	// Check vanishing constraints
	err := table.ForallAcceptTrace(trace, p.vanishing)
	if err != nil {
		return err
	}
	// Check range constraints
	err = table.ForallAcceptTrace(trace, p.ranges)
	if err != nil {
		return err
	}
	// TODO: handle assertions.  These cannot be checked in the same way as for
	// other constraints at the AIR level because the prover does not support
	// them.

	return nil
}

// ExpandTrace expands a given trace according to this schema.  More
// specifically, that means computing the actual values for any computed
// columns. Observe that computed columns have to be computed in the correct
// order.
func (p *Schema) ExpandTrace(tr table.Trace) {
	for _, c := range p.computedColumns {
		if !tr.HasColumn(c.Name) {
			data := make([]*fr.Element, tr.Height())
			// Expand the trace
			for i := 0; i < len(data); i++ {
				var err error
				// NOTE: at the moment Get cannot return an error anyway
				data[i], err = c.Get(i, tr)
				// FIXME: we need proper error handling
				if err != nil {
					panic(err)
				}
			}
			// Colunm needs to be expanded.
			tr.AddColumn(c.Name, data)
		}
	}
}
