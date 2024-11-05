package mir

import (
	"fmt"

	"github.com/consensys/gnark-crypto/ecc/bls12-377/fr"
	"github.com/consensys/go-corset/pkg/schema"
	"github.com/consensys/go-corset/pkg/schema/assignment"
	"github.com/consensys/go-corset/pkg/schema/constraint"
	"github.com/consensys/go-corset/pkg/trace"
	"github.com/consensys/go-corset/pkg/util"
)

// DataColumn captures the essence of a data column at the MIR level.
type DataColumn = *assignment.DataColumn

// LookupConstraint captures the essence of a lookup constraint at the HIR
// level.
type LookupConstraint = *constraint.LookupConstraint[Expr]

// VanishingConstraint captures the essence of a vanishing constraint at the MIR
// level. A vanishing constraint is a row constraint which must evaluate to
// zero.
type VanishingConstraint = *constraint.VanishingConstraint[constraint.ZeroTest[Expr]]

// RangeConstraint captures the essence of a range constraints at the MIR level.
type RangeConstraint = *constraint.RangeConstraint[Expr]

// PropertyAssertion captures the notion of an arbitrary property which should
// hold for all acceptable traces.  However, such a property is not enforced by
// the prover.
type PropertyAssertion = *schema.PropertyAssertion[constraint.ZeroTest[Expr]]

// Permutation captures the notion of a (sorted) permutation at the MIR level.
type Permutation = *assignment.SortedPermutation

// Interleaving captures the notion of an interleaving at the MIR level.
type Interleaving = *assignment.Interleaving

// Schema for MIR traces
type Schema struct {
	// The modules of the schema
	modules []schema.Module
	// The data columns of this schema.
	inputs []schema.Declaration
	// The sorted permutations of this schema.
	assignments []schema.Assignment
	// The constraints of this schema, which are either vanishing constraints,
	// type constraints or lookup constraints.
	constraints []schema.Constraint
	// The property assertions for this schema.
	assertions []PropertyAssertion
	// Cache list of columns declared in inputs and assignments.
	column_cache []schema.Column
}

// EmptySchema is used to construct a fresh schema onto which new columns and
// constraints will be added.
func EmptySchema() *Schema {
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

// AddDataColumn appends a new data column.
func (p *Schema) AddDataColumn(context trace.Context, name string, base schema.Type) {
	if context.Module() >= uint(len(p.modules)) {
		panic(fmt.Sprintf("invalid module index (%d)", context.Module()))
	}
	// Create column
	col := assignment.NewDataColumn(context, name, base)
	p.inputs = append(p.inputs, col)
	// Update column cache
	for c := col.Columns(); c.HasNext(); {
		p.column_cache = append(p.column_cache, c.Next())
	}
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
func (p *Schema) AddLookupConstraint(handle string, source trace.Context, target trace.Context,
	sources []Expr, targets []Expr) {
	if len(targets) != len(sources) {
		panic("differeng number of target / source lookup columns")
	}
	// TODO: sanity source columns are in the same module, and likewise target
	// columns (though they don't have to be in the same column together).
	p.constraints = append(p.constraints,
		constraint.NewLookupConstraint(handle, source, target, sources, targets))
}

// AddVanishingConstraint appends a new vanishing constraint.
func (p *Schema) AddVanishingConstraint(handle string, context trace.Context, domain *int, expr Expr) {
	if context.Module() >= uint(len(p.modules)) {
		panic(fmt.Sprintf("invalid module index (%d)", context.Module()))
	}

	p.constraints = append(p.constraints,
		constraint.NewVanishingConstraint(handle, context, domain, constraint.ZeroTest[Expr]{Expr: expr}))
}

// AddRangeConstraint appends a new range constraint.
func (p *Schema) AddRangeConstraint(handle string, context trace.Context, expr Expr, bound fr.Element) {
	p.constraints = append(p.constraints, constraint.NewRangeConstraint(handle, context, expr, bound))
}

// AddPropertyAssertion appends a new property assertion.
func (p *Schema) AddPropertyAssertion(handle string, context trace.Context, expr Expr) {
	test := constraint.ZeroTest[Expr]{Expr: expr}
	p.assertions = append(p.assertions, schema.NewPropertyAssertion(handle, context, test))
}

// ============================================================================
// Schema Interface
// ============================================================================

// InputColumns returns an array over the input columns of this schema.  That
// is, the subset of columns whose trace values must be provided by the
// user.
func (p *Schema) InputColumns() util.Iterator[schema.Column] {
	inputs := util.NewArrayIterator(p.inputs)
	return util.NewFlattenIterator[schema.Declaration, schema.Column](inputs,
		func(d schema.Declaration) util.Iterator[schema.Column] { return d.Columns() })
}

// Assertions returns an iterator over the property assertions of this
// schema.  These are properties which should hold true for any valid trace
// (though, of course, may not hold true for an invalid trace).
func (p *Schema) Assertions() util.Iterator[schema.Constraint] {
	properties := util.NewArrayIterator(p.assertions)
	return util.NewCastIterator[PropertyAssertion, schema.Constraint](properties)
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
	return util.NewArrayIterator(p.column_cache)
}

// Constraints returns an array over the underlying constraints of this
// schema.
func (p *Schema) Constraints() util.Iterator[schema.Constraint] {
	return util.NewArrayIterator(p.constraints)
}

// Declarations returns an array over the column declarations of this
// schema.
func (p *Schema) Declarations() util.Iterator[schema.Declaration] {
	inputs := util.NewArrayIterator(p.inputs)
	ps := util.NewCastIterator[schema.Assignment, schema.Declaration](p.Assignments())

	return inputs.Append(ps)
}

// Modules returns an iterator over the declared set of modules within this
// schema.
func (p *Schema) Modules() util.Iterator[schema.Module] {
	return util.NewArrayIterator(p.modules)
}
