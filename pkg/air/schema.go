package air

import (
	"github.com/consensys/gnark-crypto/ecc/bls12-377/fr"
	"github.com/consensys/go-corset/pkg/schema"
	"github.com/consensys/go-corset/pkg/schema/assignment"
	"github.com/consensys/go-corset/pkg/schema/constraint"
	"github.com/consensys/go-corset/pkg/util"
)

// DataColumn captures the essence of a data column at AIR level.
type DataColumn = *assignment.DataColumn

// PropertyAssertion captures the notion of an arbitrary property which should
// hold for all acceptable traces.  However, such a property is not enforced by
// the prover.
type PropertyAssertion = *schema.PropertyAssertion[constraint.ZeroTest[schema.Evaluable]]

// Schema for AIR traces which is parameterised on a notion of computation as
// permissible in computed columns.
type Schema struct {
	// The set of data columns corresponding to the inputs of this schema.
	inputs []schema.Declaration
	// Assignments defines the set of column declarations whose trace values are
	// computed from the inputs.
	assignments []schema.Assignment
	// The constraints of this schema.  A constraint is either a vanishing
	// constraint, a permutation constraint, a lookup constraint or a range
	// constraint.
	constraints []schema.Constraint
	// Property assertions.
	assertions []PropertyAssertion
}

// EmptySchema is used to construct a fresh schema onto which new columns and
// constraints will be added.
func EmptySchema[C schema.Evaluable]() *Schema {
	p := new(Schema)
	p.inputs = make([]schema.Declaration, 0)
	p.assignments = make([]schema.Assignment, 0)
	p.constraints = make([]schema.Constraint, 0)
	p.assertions = make([]PropertyAssertion, 0)
	// Done
	return p
}

// AddColumn appends a new data column whose values must be provided by the
// user.
func (p *Schema) AddColumn(name string, datatype schema.Type) uint {
	// NOTE: the air level has no ability to enforce the type specified for a
	// given column.
	p.inputs = append(p.inputs, assignment.NewDataColumn(name, datatype))
	// Calculate column index
	return uint(len(p.inputs) - 1)
}

// AddAssignment appends a new assignment (i.e. set of computed columns) to be
// used during trace expansion for this schema.  Computed columns are introduced
// by the process of lowering from HIR / MIR to AIR.
func (p *Schema) AddAssignment(c schema.Assignment) uint {
	index := p.Columns().Count()
	p.assignments = append(p.assignments, c)

	return index
}

// AddPermutationConstraint appends a new permutation constraint which
// ensures that one column is a permutation of another.
func (p *Schema) AddPermutationConstraint(targets []uint, sources []uint) {
	p.constraints = append(p.constraints, constraint.NewPermutationConstraint(targets, sources))
}

// AddVanishingConstraint appends a new vanishing constraint.
func (p *Schema) AddVanishingConstraint(handle string, domain *int, expr Expr) {
	p.constraints = append(p.constraints,
		constraint.NewVanishingConstraint(handle, domain, constraint.ZeroTest[Expr]{Expr: expr}))
}

// AddRangeConstraint appends a new range constraint.
func (p *Schema) AddRangeConstraint(column uint, bound *fr.Element) {
	p.constraints = append(p.constraints, constraint.NewRangeConstraint(column, bound))
}

// ============================================================================
// Schema Interface
// ============================================================================

// Inputs returns an array over the input declarations of this schema.  That is,
// the subset of declarations whose trace values must be provided by the user.
func (p *Schema) Inputs() util.Iterator[schema.Declaration] {
	return util.NewArrayIterator(p.inputs)
}

// Assignments returns an array over the assignments of this schema.  That
// is, the subset of declarations whose trace values can be computed from
// the inputs.
func (p *Schema) Assignments() util.Iterator[schema.Assignment] {
	return util.NewArrayIterator(p.assignments)
}

// Columns returns an array over the underlying columns of this schema.
// Specifically, the index of a column in this array is its column index.
func (p *Schema) Columns() util.Iterator[schema.Column] {
	is := util.NewFlattenIterator[schema.Declaration, schema.Column](p.Inputs(),
		func(d schema.Declaration) util.Iterator[schema.Column] { return d.Columns() })
	ps := util.NewFlattenIterator[schema.Assignment, schema.Column](p.Assignments(),
		func(d schema.Assignment) util.Iterator[schema.Column] { return d.Columns() })
	//
	return is.Append(ps)
}

// Constraints returns an array over the underlying constraints of this
// schema.
func (p *Schema) Constraints() util.Iterator[schema.Constraint] {
	return util.NewArrayIterator(p.constraints)
}

// Declarations returns an array over the column declarations of this
// schema.
func (p *Schema) Declarations() util.Iterator[schema.Declaration] {
	ps := util.NewCastIterator[schema.Assignment, schema.Declaration](p.Assignments())
	return p.Inputs().Append(ps)
}
