package air

import (
	"github.com/consensys/gnark-crypto/ecc/bls12-377/fr"
	"github.com/consensys/go-corset/pkg/table"
	"github.com/consensys/go-corset/pkg/util"
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

// Permutation captures the notion of a simple column permutation at the AIR
// level.
type Permutation = *table.Permutation

// Schema for AIR traces which is parameterised on a notion of computation as
// permissible in computed columns.
type Schema struct {
	// The data columns of this schema.
	dataColumns []DataColumn
	// The permutation columns of this schema.
	permutations []Permutation
	// The vanishing constraints of this schema.
	vanishing []VanishingConstraint
	// The range constraints of this schema.
	ranges []*table.RangeConstraint
	// Property assertions.
	assertions []PropertyAssertion
	// The computations used to construct traces which adhere to
	// this schema.  Such computations are not expressible at the
	// prover level and, hence, can only be used to pre-process
	// traces prior to prove generation.
	computations []table.TraceComputation
}

// EmptySchema is used to construct a fresh schema onto which new columns and
// constraints will be added.
func EmptySchema[C table.Evaluable]() *Schema {
	p := new(Schema)
	p.dataColumns = make([]DataColumn, 0)
	p.permutations = make([]Permutation, 0)
	p.vanishing = make([]VanishingConstraint, 0)
	p.ranges = make([]*table.RangeConstraint, 0)
	p.assertions = make([]PropertyAssertion, 0)
	p.computations = make([]table.TraceComputation, 0)
	// Done
	return p
}

// Width returns the number of column groups in this schema.
func (p *Schema) Width() uint {
	return uint(len(p.dataColumns))
}

// ColumnGroup returns information about the ith column group in this schema.
func (p *Schema) ColumnGroup(i uint) table.ColumnGroup {
	return p.dataColumns[i]
}

// Column returns information about the ith column in this schema.
func (p *Schema) Column(i uint) table.ColumnSchema {
	return p.dataColumns[i]
}

// Size returns the number of declarations in this schema.
func (p *Schema) Size() int {
	return len(p.dataColumns) + len(p.permutations) + len(p.vanishing) +
		len(p.ranges) + len(p.assertions) + len(p.computations)
}

// GetDeclaration returns the ith declaration in this schema.
func (p *Schema) GetDeclaration(index int) table.Declaration {
	ith := util.FlatArrayIndexOf_6(index, p.dataColumns, p.permutations,
		p.vanishing, p.ranges, p.assertions, p.computations)
	return ith.(table.Declaration)
}

// Columns returns the set of data columns.
func (p *Schema) Columns() []DataColumn {
	return p.dataColumns
}

// HasColumn checks whether a given schema has a given column.
func (p *Schema) HasColumn(name string) bool {
	for _, c := range p.dataColumns {
		if c.Name() == name {
			return true
		}
	}

	return false
}

// IndexOf determines the column index for a given column in this schema, or
// returns false indicating an error.
func (p *Schema) IndexOf(name string) (uint, bool) {
	for i, c := range p.dataColumns {
		if c.Name() == name {
			return uint(i), true
		}
	}

	return 0, false
}

// RequiredSpillage returns the minimum amount of spillage required to ensure
// valid traces are accepted in the presence of arbitrary padding.  Spillage can
// only arise from computations as this is where values outside of the user's
// control are determined.
func (p *Schema) RequiredSpillage() uint {
	// Ensures always at least one row of spillage (referred to as the "initial
	// padding row")
	mx := uint(1)
	// Determine if any more spillage required
	for _, c := range p.computations {
		mx = max(mx, c.RequiredSpillage())
	}

	return mx
}

// AddColumn appends a new data column which is either synthetic or
// not.  A synthetic column is one which has been introduced by the
// process of lowering from HIR / MIR to AIR.  That is, it is not a
// column which was original specified by the user.  Columns also support a
// "padding sign", which indicates whether padding should occur at the front
// (positive sign) or the back (negative sign).
func (p *Schema) AddColumn(name string, synthetic bool) uint {
	// NOTE: the air level has no ability to enforce the type specified for a
	// given column.
	p.dataColumns = append(p.dataColumns, table.NewDataColumn(name, &table.FieldType{}, synthetic))
	// Calculate column index
	return uint(len(p.dataColumns) - 1)
}

// AddComputation appends a new computation to be used during trace
// expansion for this schema.
func (p *Schema) AddComputation(c table.TraceComputation) {
	p.computations = append(p.computations, c)
}

// AddPermutationConstraint appends a new permutation constraint which
// ensures that one column is a permutation of another.
func (p *Schema) AddPermutationConstraint(targets []uint, sources []uint) {
	p.permutations = append(p.permutations, table.NewPermutation(targets, sources))
}

// AddVanishingConstraint appends a new vanishing constraint.
func (p *Schema) AddVanishingConstraint(handle string, domain *int, expr Expr) {
	p.vanishing = append(p.vanishing, table.NewRowConstraint(handle, domain, table.ZeroTest[Expr]{Expr: expr}))
}

// AddRangeConstraint appends a new range constraint.
func (p *Schema) AddRangeConstraint(column uint, bound *fr.Element) {
	p.ranges = append(p.ranges, table.NewRangeConstraint(column, bound))
}

// Accepts determines whether this schema will accept a given trace.  That
// is, whether or not the given trace adheres to the schema.  A trace can fail
// to adhere to the schema for a variety of reasons, such as having a constraint
// which does not hold.
func (p *Schema) Accepts(trace table.Trace) error {
	// Check vanishing constraints
	err := table.ConstraintsAcceptTrace(trace, p.vanishing)
	if err != nil {
		return err
	}
	// Check permutation constraints
	err = table.ConstraintsAcceptTrace(trace, p.permutations)
	if err != nil {
		return err
	}
	// Check range constraints
	err = table.ConstraintsAcceptTrace(trace, p.ranges)
	if err != nil {
		return err
	}
	// Check computations
	err = table.ConstraintsAcceptTrace(trace, p.computations)
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
func (p *Schema) ExpandTrace(tr table.Trace) error {
	// Execute all computations
	for _, c := range p.computations {
		err := c.ExpandTrace(tr)
		if err != nil {
			return err
		}
	}
	// Done
	return nil
}
