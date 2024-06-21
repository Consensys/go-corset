package hir

import (
	"github.com/consensys/go-corset/pkg/schema"
	"github.com/consensys/go-corset/pkg/schema/assignment"
	"github.com/consensys/go-corset/pkg/schema/constraint"
	"github.com/consensys/go-corset/pkg/trace"
	"github.com/consensys/go-corset/pkg/util"
)

// ZeroArrayTest is a wrapper which converts an array of expressions into a
// Testable constraint.  Specifically, by checking whether or not the each
// expression vanishes (i.e. evaluates to zero).
type ZeroArrayTest struct {
	Expr Expr
}

// TestAt determines whether or not every element from a given array of
// expressions evaluates to zero. Observe that any expressions which are
// undefined are assumed to hold.
func (p ZeroArrayTest) TestAt(row int, tr trace.Trace) bool {
	// Evalues expression yielding zero or more values.
	vals := p.Expr.EvalAllAt(row, tr)
	// Check each value in turn against zero.
	for _, val := range vals {
		if val != nil && !val.IsZero() {
			// This expression does not evaluat to zero, hence failure.
			return false
		}
	}
	// Success
	return true
}

func (p ZeroArrayTest) String() string {
	return p.Expr.String()
}

// Bounds determines the bounds for this zero test.
func (p ZeroArrayTest) Bounds() util.Bounds {
	return p.Expr.Bounds()
}

// DataColumn captures the essence of a data column at AIR level.
type DataColumn = *assignment.DataColumn

// VanishingConstraint captures the essence of a vanishing constraint at the HIR
// level. A vanishing constraint is a row constraint which must evaluate to
// zero.
type VanishingConstraint = *constraint.VanishingConstraint[ZeroArrayTest]

// PropertyAssertion captures the notion of an arbitrary property which should
// hold for all acceptable traces.  However, such a property is not enforced by
// the prover.
type PropertyAssertion = *schema.PropertyAssertion[ZeroArrayTest]

// Permutation captures the notion of a (sorted) permutation at the HIR level.
type Permutation = *assignment.SortedPermutation

// Schema for HIR constraints and columns.
type Schema struct {
	// The data columns of this schema.
	inputs []schema.Declaration
	// The sorted permutations of this schema.
	assignments []schema.Assignment
	// Constraints of this schema, which are either vanishing, lookup or type
	// constraints.
	constraints []schema.Constraint
	// The property assertions for this schema.
	assertions []PropertyAssertion
}

// EmptySchema is used to construct a fresh schema onto which new columns and
// constraints will be added.
func EmptySchema() *Schema {
	p := new(Schema)
	p.inputs = make([]schema.Declaration, 0)
	p.assignments = make([]schema.Assignment, 0)
	p.constraints = make([]schema.Constraint, 0)
	p.assertions = make([]PropertyAssertion, 0)
	// Done
	return p
}

// AddDataColumn appends a new data column with a given type.  Furthermore, the
// type is enforced by the system when checking is enabled.
func (p *Schema) AddDataColumn(name string, base schema.Type) {
	p.inputs = append(p.inputs, assignment.NewDataColumn(name, base))
}

// AddPermutationColumns introduces a permutation of one or more
// existing columns.  Specifically, this introduces one or more
// computed columns which represent a (sorted) permutation of the
// source columns.  Each source column is associated with a "sign"
// which indicates the direction of sorting (i.e. ascending versus
// descending).
func (p *Schema) AddPermutationColumns(targets []schema.Column, signs []bool, sources []string) {
	p.assignments = append(p.assignments, assignment.NewSortedPermutation(targets, signs, sources))
}

// AddVanishingConstraint appends a new vanishing constraint.
func (p *Schema) AddVanishingConstraint(handle string, domain *int, expr Expr) {
	p.constraints = append(p.constraints, constraint.NewVanishingConstraint(handle, domain, ZeroArrayTest{expr}))
}

// AddTypeConstraint appends a new range constraint.
func (p *Schema) AddTypeConstraint(target string, t schema.Type) {
	// Check whether is a field type, as these can actually be ignored.
	if t.AsField() == nil {
		p.constraints = append(p.constraints, constraint.NewTypeConstraint(target, t))
	}
}

// AddPropertyAssertion appends a new property assertion.
func (p *Schema) AddPropertyAssertion(handle string, property Expr) {
	p.assertions = append(p.assertions, schema.NewPropertyAssertion[ZeroArrayTest](handle, ZeroArrayTest{property}))
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
