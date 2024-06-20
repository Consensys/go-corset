package mir

import (
	"fmt"

	"github.com/consensys/go-corset/pkg/air"
	air_gadgets "github.com/consensys/go-corset/pkg/air/gadgets"
	"github.com/consensys/go-corset/pkg/schema"
	"github.com/consensys/go-corset/pkg/trace"
	"github.com/consensys/go-corset/pkg/util"
)

// DataColumn captures the essence of a data column at the MIR level.
type DataColumn = *schema.DataColumn[schema.Type]

// VanishingConstraint captures the essence of a vanishing constraint at the MIR
// level. A vanishing constraint is a row constraint which must evaluate to
// zero.
type VanishingConstraint = *schema.RowConstraint[schema.ZeroTest[Expr]]

// PropertyAssertion captures the notion of an arbitrary property which should
// hold for all acceptable traces.  However, such a property is not enforced by
// the prover.
type PropertyAssertion = *schema.PropertyAssertion[schema.ZeroTest[Expr]]

// Permutation captures the notion of a (sorted) permutation at the MIR level.
type Permutation = *schema.SortedPermutation

// Schema for MIR traces
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

// Width returns the number of column groups in this schema.
func (p *Schema) Width() uint {
	return uint(len(p.dataColumns) + len(p.permutations))
}

// Column returns information about the ith column in this schema.
func (p *Schema) Column(i uint) schema.ColumnSchema {
	panic("todo")
}

// ColumnGroup returns information about the ith column group in this schema.
func (p *Schema) ColumnGroup(i uint) schema.ColumnGroup {
	n := uint(len(p.dataColumns))
	if i < n {
		return p.dataColumns[i]
	}

	return p.permutations[i-n]
}

// ColumnIndex determines the column index for a given column in this schema, or
// returns false indicating an error.
func (p *Schema) ColumnIndex(name string) (uint, bool) {
	index := uint(0)

	for i := uint(0); i < p.Width(); i++ {
		ith := p.ColumnGroup(i)
		for j := uint(0); j < ith.Width(); j++ {
			if ith.NameOf(j) == name {
				// hit
				return index, true
			}

			index++
		}
	}
	// miss
	return 0, false
}

// GetColumnByName gets a given data column based on its name.  If no such
// column exists, it panics.
func (p *Schema) GetColumnByName(name string) DataColumn {
	for _, c := range p.dataColumns {
		if c.Name() == name {
			return c
		}
	}

	msg := fmt.Sprintf("unknown column encountered (%s)", name)
	panic(msg)
}

// Size returns the number of declarations in this schema.
func (p *Schema) Size() int {
	return len(p.dataColumns) + len(p.permutations) + len(p.vanishing) + len(p.assertions)
}

// RequiredSpillage returns the minimum amount of spillage required to ensure
// valid traces are accepted in the presence of arbitrary padding.
func (p *Schema) RequiredSpillage() uint {
	// Ensures always at least one row of spillage (referred to as the "initial
	// padding row")
	return uint(1)
}

// GetDeclaration returns the ith declaration in this schema.
func (p *Schema) GetDeclaration(index int) schema.Declaration {
	ith := util.FlatArrayIndexOf_4(index, p.dataColumns, p.permutations, p.vanishing, p.assertions)
	return ith.(schema.Declaration)
}

// AddDataColumn appends a new data column.
func (p *Schema) AddDataColumn(name string, base schema.Type) {
	p.dataColumns = append(p.dataColumns, schema.NewDataColumn(name, base, false))
}

// AddPermutationColumns introduces a permutation of one or more
// existing columns.  Specifically, this introduces one or more
// computed columns which represent a (sorted) permutation of the
// source columns.  Each source column is associated with a "sign"
// which indicates the direction of sorting (i.e. ascending versus
// descending).
func (p *Schema) AddPermutationColumns(targets []string, signs []bool, sources []string) {
	p.permutations = append(p.permutations, schema.NewSortedPermutation(targets, signs, sources))
}

// AddVanishingConstraint appends a new vanishing constraint.
func (p *Schema) AddVanishingConstraint(handle string, domain *int, expr Expr) {
	p.vanishing = append(p.vanishing, schema.NewRowConstraint(handle, domain, schema.ZeroTest[Expr]{Expr: expr}))
}

// AddPropertyAssertion appends a new property assertion.
func (p *Schema) AddPropertyAssertion(handle string, expr Expr) {
	test := schema.ZeroTest[Expr]{Expr: expr}
	p.assertions = append(p.assertions, schema.NewPropertyAssertion(handle, test))
}

// Accepts determines whether this schema will accept a given trace.  That
// is, whether or not the given trace adheres to the schema.  A trace can fail
// to adhere to the schema for a variety of reasons, such as having a constraint
// which does not hold.
func (p *Schema) Accepts(trace trace.Trace) error {
	// Check (typed) data columns
	if err := schema.ConstraintsAcceptTrace(trace, p.dataColumns); err != nil {
		return err
	}
	// Check permutations
	if err := schema.ConstraintsAcceptTrace(trace, p.permutations); err != nil {
		return err
	}
	// Check vanishing constraints
	if err := schema.ConstraintsAcceptTrace(trace, p.vanishing); err != nil {
		return err
	}
	// Check property assertions
	if err := schema.ConstraintsAcceptTrace(trace, p.assertions); err != nil {
		return err
	}
	// Done
	return nil
}

// LowerToAir lowers (or refines) an MIR table into an AIR schema.  That means
// lowering all the columns and constraints, whilst adding additional columns /
// constraints as necessary to preserve the original semantics.
func (p *Schema) LowerToAir() *air.Schema {
	airSchema := air.EmptySchema[Expr]()
	// Allocate data and permutation columns.  This must be done first to ensure
	// alignment is preserved across lowering.
	index := uint(0)

	for i := uint(0); i < p.Width(); i++ {
		ith := p.ColumnGroup(i)
		for j := uint(0); j < ith.Width(); j++ {
			col := ith.NameOf(j)
			airSchema.AddColumn(col, ith.IsSynthetic())

			index++
		}
	}
	// Add computations. Again this has to be done first for things to work.
	// Essentially to reflect the fact that these columns have been added above
	// before others.  Realistically, the overall design of this process is a
	// bit broken right now.
	for _, perm := range p.permutations {
		airSchema.AddComputation(perm)
	}
	// Lower checked data columns
	for i, col := range p.dataColumns {
		lowerColumnToAir(uint(i), col, airSchema)
	}
	// Lower permutations columns
	for _, perm := range p.permutations {
		lowerPermutationToAir(perm, p, airSchema)
	}
	// Lower vanishing constraints
	for _, c := range p.vanishing {
		// FIXME: this is broken because its currently
		// assuming that an AirConstraint is always a
		// VanishingConstraint.  Eventually this will not be
		// true.
		air_expr := c.Constraint.Expr.LowerTo(airSchema)
		airSchema.AddVanishingConstraint(c.Handle, c.Domain, air_expr)
	}
	// Done
	return airSchema
}

// Lower a datacolumn to the AIR level.  The main effect of this is that, for
// columns with non-trivial types, we must add appropriate range constraints to
// the enclosing schema.
func lowerColumnToAir(index uint, c *schema.DataColumn[schema.Type], schema *air.Schema) {
	// Check whether a constraint is implied by the column's type
	if t := c.Type.AsUint(); t != nil && t.Checked() {
		// Yes, a constraint is implied.  Now, decide whether to use a range
		// constraint or just a vanishing constraint.
		if t.HasBound(2) {
			// u1 => use vanishing constraint X * (X - 1)
			air_gadgets.ApplyBinaryGadget(index, schema)
		} else if t.HasBound(256) {
			// u2..8 use range constraints
			schema.AddRangeConstraint(index, t.Bound())
		} else {
			// u9+ use byte decompositions.
			air_gadgets.ApplyBitwidthGadget(index, t.BitWidth(), schema)
		}
	}
}

// Lower a permutation to the AIR level.  This has quite a few
// effects.  Firstly, permutation constraints are added for all of the
// new columns.  Secondly, sorting constraints (and their associated
// synthetic columns) must also be added.  Finally, a trace
// computation is required to ensure traces are correctly expanded to
// meet the requirements of a sorted permutation.
func lowerPermutationToAir(c Permutation, mirSchema *Schema, airSchema *air.Schema) {
	ncols := len(c.Targets)
	//
	targets := make([]uint, ncols)
	sources := make([]uint, ncols)
	// Add individual permutation constraints
	for i := 0; i < ncols; i++ {
		var ok1, ok2 bool
		// TODO: REPLACE
		sources[i], ok1 = airSchema.ColumnIndex(c.Sources[i])
		targets[i], ok2 = airSchema.ColumnIndex(c.Targets[i])

		if !ok1 || !ok2 {
			panic("missing column")
		}
	}
	//
	airSchema.AddPermutationConstraint(targets, sources)
	// Add sorting constraints + synthetic columns as necessary.
	if ncols == 1 {
		// For a single column sort, its actually a bit easier because we don't
		// need to implement a multiplexor (i.e. to determine which column is
		// differs, etc).  Instead, we just need a delta column which ensures
		// there is a non-negative difference between consecutive rows.  This
		// also requires bitwidth constraints.
		bitwidth := mirSchema.GetColumnByName(c.Sources[0]).Type.AsUint().BitWidth()
		// Add column sorting constraints
		air_gadgets.ApplyColumnSortGadget(targets[0], c.Signs[0], bitwidth, airSchema)
	} else {
		// For a multi column sort, its a bit harder as we need additional
		// logicl to ensure the target columns are lexicographally sorted.
		bitwidth := uint(0)

		for i := 0; i < ncols; i++ {
			// Extract bitwidth of ith column
			ith := mirSchema.GetColumnByName(c.Sources[i]).Type.AsUint().BitWidth()
			if ith > bitwidth {
				bitwidth = ith
			}
		}
		// Add lexicographically sorted constraints
		air_gadgets.ApplyLexicographicSortingGadget(targets, c.Signs, bitwidth, airSchema)
	}
}

// ExpandTrace expands a given trace according to this schema.
func (p *Schema) ExpandTrace(tr trace.Trace) error {
	// Expand all the permutation columns
	for _, perm := range p.permutations {
		err := perm.ExpandTrace(tr)
		if err != nil {
			return err
		}
	}

	return nil
}
