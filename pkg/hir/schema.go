package hir

import (
	"github.com/consensys/go-corset/pkg/mir"
	"github.com/consensys/go-corset/pkg/table"
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
func (p ZeroArrayTest) TestAt(row int, tr table.Trace) bool {
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

// DataColumn captures the essence of a data column at AIR level.
type DataColumn = *table.DataColumn[table.Type]

// VanishingConstraint captures the essence of a vanishing constraint at the HIR
// level. A vanishing constraint is a row constraint which must evaluate to
// zero.
type VanishingConstraint = *table.RowConstraint[ZeroArrayTest]

// PropertyAssertion captures the notion of an arbitrary property which should
// hold for all acceptable traces.  However, such a property is not enforced by
// the prover.
type PropertyAssertion = mir.PropertyAssertion

// Permutation captures the notion of a (sorted) permutation at the HIR level.
type Permutation = *table.SortedPermutation

// Schema for HIR constraints and columns.
type Schema struct {
	// The data columns of this schema.
	dataColumns []DataColumn
	// The sorted permutations of this schema.
	permutations []Permutation
	// The vanishing constraints of this schema.
	vanishing []VanishingConstraint
	// The property assertions for this schema.
	assertions []PropertyAssertion
}

// EmptySchema is used to construct a fresh schema onto which new columns and
// constraints will be added.
func EmptySchema() *Schema {
	p := new(Schema)
	p.dataColumns = make([]DataColumn, 0)
	p.permutations = make([]Permutation, 0)
	p.vanishing = make([]VanishingConstraint, 0)
	p.assertions = make([]PropertyAssertion, 0)
	// Done
	return p
}

// HasColumn checks whether a given schema has a given column.
func (p *Schema) HasColumn(name string) bool {
	for _, c := range p.dataColumns {
		if (*c).Name == name {
			return true
		}
	}

	return false
}

// Columns returns the set of (data) columns declared within this schema.
func (p *Schema) Columns() []*table.DataColumn[table.Type] {
	return p.dataColumns
}

// Constraints returns the set of (vanishing) constraints declared within this schema.
func (p *Schema) Constraints() []VanishingConstraint {
	return p.vanishing
}

// Size returns the number of declarations in this schema.
func (p *Schema) Size() int {
	return len(p.dataColumns) + len(p.permutations) + len(p.vanishing) + len(p.assertions)
}

// GetDeclaration returns the ith declaration in this schema.
func (p *Schema) GetDeclaration(index int) table.Declaration {
	ith := util.FlatArrayIndexOf_4(index, p.dataColumns, p.permutations, p.vanishing, p.assertions)
	return ith.(table.Declaration)
}

// AddDataColumn appends a new data column with a given type.  Furthermore, the
// type is enforced by the system when checking is enabled.
func (p *Schema) AddDataColumn(name string, base table.Type) {
	p.dataColumns = append(p.dataColumns, table.NewDataColumn(name, base, false))
}

// AddPermutationColumns introduces a permutation of one or more
// existing columns.  Specifically, this introduces one or more
// computed columns which represent a (sorted) permutation of the
// source columns.  Each source column is associated with a "sign"
// which indicates the direction of sorting (i.e. ascending versus
// descending).
func (p *Schema) AddPermutationColumns(targets []string, signs []bool, sources []string) {
	p.permutations = append(p.permutations, table.NewSortedPermutation(targets, signs, sources))
}

// AddVanishingConstraint appends a new vanishing constraint.
func (p *Schema) AddVanishingConstraint(handle string, domain *int, expr Expr) {
	p.vanishing = append(p.vanishing, table.NewRowConstraint(handle, domain, ZeroArrayTest{expr}))
}

// AddPropertyAssertion appends a new property assertion.
func (p *Schema) AddPropertyAssertion(handle string, expr mir.Expr) {
	p.assertions = append(p.assertions, table.NewPropertyAssertion[mir.Expr](handle, expr))
}

// Accepts determines whether this schema will accept a given trace.  That
// is, whether or not the given trace adheres to the schema.  A trace can fail
// to adhere to the schema for a variety of reasons, such as having a constraint
// which does not hold.
func (p *Schema) Accepts(trace table.Trace) error {
	// Check (typed) data columns
	if err := table.ConstraintsAcceptTrace(trace, p.dataColumns); err != nil {
		return err
	}
	// Check permutations
	if err := table.ConstraintsAcceptTrace(trace, p.permutations); err != nil {
		return err
	}
	// Check vanishing constraints
	if err := table.ConstraintsAcceptTrace(trace, p.vanishing); err != nil {
		return err
	}
	// Check properties
	if err := table.ConstraintsAcceptTrace(trace, p.assertions); err != nil {
		return err
	}
	// Done
	return nil
}

// ExpandTrace expands a given trace according to this schema.
func (p *Schema) ExpandTrace(tr table.Trace) error {
	// Insert initial padding row
	table.PadTrace(1, tr)
	// Expand all the permutation columns
	for _, perm := range p.permutations {
		err := perm.ExpandTrace(tr)
		if err != nil {
			return err
		}
	}

	return nil
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
	// Second, lower permutations
	for _, col := range p.permutations {
		mirSchema.AddPermutationColumns(col.Targets, col.Signs, col.Sources)
	}
	// Third, lower constraints
	for _, c := range p.vanishing {
		mir_exprs := c.Constraint.Expr.LowerTo()
		// Add individual constraints arising
		for _, mir_expr := range mir_exprs {
			mirSchema.AddVanishingConstraint(c.Handle, c.Domain, mir_expr)
		}
	}
	// Fourth, copy property assertions.  Observe, these do not require lowering
	// because they are already MIR-level expressions.
	for _, c := range p.assertions {
		mirSchema.AddPropertyAssertion(c.Handle, c.Expr)
	}
	//
	return mirSchema
}
