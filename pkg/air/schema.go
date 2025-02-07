package air

import (
	"fmt"

	"github.com/consensys/gnark-crypto/ecc/bls12-377/fr"
	"github.com/consensys/go-corset/pkg/schema"
	"github.com/consensys/go-corset/pkg/schema/assignment"
	"github.com/consensys/go-corset/pkg/schema/constraint"
	"github.com/consensys/go-corset/pkg/trace"
	"github.com/consensys/go-corset/pkg/util"
	"github.com/consensys/go-corset/pkg/util/collection/iter"
)

// DataColumn captures the essence of a data column at AIR level.
type DataColumn = *assignment.DataColumn

// LookupConstraint captures the essence of a lookup constraint at the AIR
// level.  At the AIR level, lookup constraints are only permitted between
// columns (i.e. not arbitrary expressions).
type LookupConstraint = *constraint.LookupConstraint[*ColumnAccess]

// VanishingConstraint captures the essence of a vanishing constraint at the AIR level.
type VanishingConstraint = *constraint.VanishingConstraint[Expr]

// RangeConstraint captures the essence of a range constraints at the AIR level.
type RangeConstraint = *constraint.RangeConstraint[*ColumnAccess]

// PermutationConstraint captures the essence of a permutation constraint at the AIR level.
// Specifically, it represents a constraint that one (or more) columns are a permutation of another.
type PermutationConstraint = *constraint.PermutationConstraint

// PropertyAssertion captures the notion of an arbitrary property which should
// hold for all acceptable traces.  However, such a property is not enforced by
// the prover.
type PropertyAssertion = *schema.PropertyAssertion[schema.Testable]

// Schema for AIR traces which is parameterised on a notion of computation as
// permissible in computed columns.
type Schema struct {
	// The modules of the schema
	modules []schema.Module
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
	// Cache list of columns declared in inputs and assignments.
	column_cache []schema.Column
}

// EmptySchema is used to construct a fresh schema onto which new columns and
// constraints will be added.
func EmptySchema[C schema.Evaluable]() *Schema {
	p := new(Schema)
	p.modules = make([]schema.Module, 0)
	p.inputs = make([]schema.Declaration, 0)
	p.assignments = make([]schema.Assignment, 0)
	p.constraints = make([]schema.Constraint, 0)
	p.assertions = make([]PropertyAssertion, 0)
	p.column_cache = make([]schema.Column, 0)
	// Done
	return p
}

// AddModule adds a new module to this schema, returning its module index.
func (p *Schema) AddModule(name string) uint {
	mid := uint(len(p.modules))
	p.modules = append(p.modules, schema.NewModule(name))

	return mid
}

// AddColumn appends a new data column whose values must be provided by the
// user.
func (p *Schema) AddColumn(context trace.Context, name string, datatype schema.Type) uint {
	if context.Module() >= uint(len(p.modules)) {
		panic(fmt.Sprintf("invalid module index (%d)", context.Module()))
	}

	col := assignment.NewDataColumn(context, name, datatype)
	// NOTE: the air level has no ability to enforce the type specified for a
	// given column.
	p.inputs = append(p.inputs, col)
	// Update column cache
	for c := col.Columns(); c.HasNext(); {
		p.column_cache = append(p.column_cache, c.Next())
	}
	// Calculate column index
	return uint(len(p.inputs) - 1)
}

// AddAssignment appends a new assignment (i.e. set of computed columns) to be
// used during trace expansion for this schema.  Computed columns are introduced
// by the process of lowering from HIR / MIR to AIR.
func (p *Schema) AddAssignment(c schema.Assignment) uint {
	index := p.Columns().Count()
	p.assignments = append(p.assignments, c)
	// Update column cache
	for c := c.Columns(); c.HasNext(); {
		p.column_cache = append(p.column_cache, c.Next())
	}

	return index
}

// AddLookupConstraint appends a new lookup constraint.
func (p *Schema) AddLookupConstraint(handle string, source trace.Context,
	target trace.Context, sources []uint, targets []uint) {
	if len(targets) != len(sources) {
		panic("differeng number of target / source lookup columns")
	}
	// TODO: sanity source columns are in the same module, and likewise target
	// columns (though they don't have to be in the same column together).
	from := make([]*ColumnAccess, len(sources))
	into := make([]*ColumnAccess, len(targets))
	// Construct column accesses from column indices.
	for i := 0; i < len(from); i++ {
		from[i] = &ColumnAccess{Column: sources[i], Shift: 0}
		into[i] = &ColumnAccess{Column: targets[i], Shift: 0}
	}
	// Construct lookup constraint
	var lookup LookupConstraint = constraint.NewLookupConstraint(handle, source, target, from, into)
	// Add
	p.constraints = append(p.constraints, lookup)
}

// AddPermutationConstraint appends a new permutation constraint which
// ensures that one column is a permutation of another.
func (p *Schema) AddPermutationConstraint(targets []uint, sources []uint) {
	// TODO: sanity target and source columns are in the same module.
	p.constraints = append(p.constraints, constraint.NewPermutationConstraint(targets, sources))
}

// AddPropertyAssertion appends a new property assertion.
func (p *Schema) AddPropertyAssertion(handle string, context trace.Context, assertion schema.Testable) {
	p.assertions = append(p.assertions, schema.NewPropertyAssertion(handle, context, assertion))
}

// AddVanishingConstraint appends a new vanishing constraint.
func (p *Schema) AddVanishingConstraint(handle string, context trace.Context, domain util.Option[int], expr Expr) {
	if context.Module() >= uint(len(p.modules)) {
		panic(fmt.Sprintf("invalid module index (%d)", context.Module()))
	}
	// TODO: sanity check expression enclosed by module
	p.constraints = append(p.constraints,
		constraint.NewVanishingConstraint(handle, context, domain, expr))
}

// AddRangeConstraint appends a new range constraint.
func (p *Schema) AddRangeConstraint(column uint, bound fr.Element) {
	col := p.Columns().Nth(column)
	handle := col.QualifiedName(p)
	tc := constraint.NewRangeConstraint(handle, col.Context, &ColumnAccess{Column: column, Shift: 0}, bound)
	p.constraints = append(p.constraints, tc)
}

// ============================================================================
// Schema Interface
// ============================================================================

// InputColumns returns an array over the input columns of this schema.  That
// is, the subset of columns whose trace values must be provided by the
// user.
func (p *Schema) InputColumns() iter.Iterator[schema.Column] {
	inputs := iter.NewArrayIterator(p.inputs)
	return iter.NewFlattenIterator[schema.Declaration, schema.Column](inputs,
		func(d schema.Declaration) iter.Iterator[schema.Column] { return d.Columns() })
}

// Assertions returns an iterator over the property assertions of this
// schema.  These are properties which should hold true for any valid trace
// (though, of course, may not hold true for an invalid trace).
func (p *Schema) Assertions() iter.Iterator[schema.Constraint] {
	properties := iter.NewArrayIterator(p.assertions)
	return iter.NewCastIterator[PropertyAssertion, schema.Constraint](properties)
}

// Assignments returns an array over the assignments of this schema.  That
// is, the subset of declarations whose trace values can be computed from
// the inputs.
func (p *Schema) Assignments() iter.Iterator[schema.Assignment] {
	return iter.NewArrayIterator(p.assignments)
}

// Columns returns an array over the underlying columns of this schema.
// Specifically, the index of a column in this array is its column index.
func (p *Schema) Columns() iter.Iterator[schema.Column] {
	return iter.NewArrayIterator(p.column_cache)
}

// Constraints returns an array over the underlying constraints of this
// schema.
func (p *Schema) Constraints() iter.Iterator[schema.Constraint] {
	return iter.NewArrayIterator(p.constraints)
}

// Declarations returns an array over the column declarations of this
// schema.
func (p *Schema) Declarations() iter.Iterator[schema.Declaration] {
	inputs := iter.NewArrayIterator(p.inputs)
	ps := iter.NewCastIterator[schema.Assignment, schema.Declaration](p.Assignments())

	return inputs.Append(ps)
}

// Modules returns an iterator over the declared set of modules within this
// schema.
func (p *Schema) Modules() iter.Iterator[schema.Module] {
	return iter.NewArrayIterator(p.modules)
}
