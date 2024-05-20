package mir

import (
	"fmt"

	"github.com/consensys/gnark-crypto/ecc/bls12-377/fr"
	"github.com/consensys/go-corset/pkg/air"
	"github.com/consensys/go-corset/pkg/table"
)

// DataColumn captures the essence of a data column at the MIR level.
type DataColumn = *table.DataColumn[table.Type]

// VanishingConstraint captures the essence of a vanishing constraint at the MIR
// level. A vanishing constraint is a row constraint which must evaluate to
// zero.
type VanishingConstraint = *table.RowConstraint[table.ZeroTest[Expr]]

// PropertyAssertion captures the notion of an arbitrary property which should
// hold for all acceptable traces.  However, such a property is not enforced by
// the prover.
type PropertyAssertion = *table.PropertyAssertion[Expr]

// Schema for MIR traces
type Schema struct {
	// The data columns of this schema.
	dataColumns []DataColumn
	// The sorted permutations of this schema.
	permutations []*table.SortedPermutation
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
	p.permutations = make([]*table.SortedPermutation, 0)
	p.vanishing = make([]VanishingConstraint, 0)
	p.assertions = make([]PropertyAssertion, 0)
	// Done
	return p
}

// AddDataColumn appends a new data column.
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
	p.vanishing = append(p.vanishing, table.NewRowConstraint(handle, domain, table.ZeroTest[Expr]{Expr: expr}))
}

// AddPropertyAssertion appends a new property assertion.
func (p *Schema) AddPropertyAssertion(handle string, expr Expr) {
	p.assertions = append(p.assertions, table.NewPropertyAssertion(handle, expr))
}

// Accepts determines whether this schema will accept a given trace.  That
// is, whether or not the given trace adheres to the schema.  A trace can fail
// to adhere to the schema for a variety of reasons, such as having a constraint
// which does not hold.
func (p *Schema) Accepts(trace table.Trace) error {
	// Check (typed) data columns
	err := table.ForallAcceptTrace(trace, p.dataColumns)
	if err != nil {
		return err
	}
	// Check permutations
	err = table.ForallAcceptTrace(trace, p.permutations)
	if err != nil {
		return err
	}
	// Check vanishing constraints
	err = table.ForallAcceptTrace(trace, p.vanishing)
	if err != nil {
		return err
	}
	// Check property assertions
	err = table.ForallAcceptTrace(trace, p.assertions)
	if err != nil {
		return err
	}

	return nil
}

// LowerToAir lowers (or refines) an MIR table into an AIR table.  That means
// lowering all the columns and constraints, whilst adding additional columns /
// constraints as necessary to preserve the original semantics.
func (p *Schema) LowerToAir() *air.Schema {
	airSchema := air.EmptySchema[Expr]()
	// Lower data columns
	for _, col := range p.dataColumns {
		lowerColumnToAir(col, airSchema)
	}
	// Lower permutations columns
	for _, col := range p.permutations {
		lowerPermutationToAir(col, airSchema)
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
func lowerColumnToAir(c *table.DataColumn[table.Type], schema *air.Schema) {
	// Check whether a constraint is implied by the column's type
	if t := c.Type.AsUint(); t != nil {
		// Yes, a constraint is implied.  Now, decide whether to use a range
		// constraint or just a vanishing constraint.
		if t.HasBound(2) {
			// u1 => use vanishing constraint X * (X - 1)
			addBinaryConstraint(c.Name, schema)
		} else if t.HasBound(256) {
			// u2..8 use range constraints
			schema.AddRangeConstraint(c.Name, t.Bound())
		} else {
			// u9+ use byte decompositions.
			addBitwidthConstraint(c.Name, t.BitWidth(), schema)
		}
	}
	// Finally, add an (untyped) data column representing this
	// data column.
	schema.AddColumn(c.Name, false)
}

// Add a binarity constraint for a given column in the schema which
// enforces that all values in the given column are either 0 or 1.
// For a column X, this corresponds to the vanishing constraint X *
// (X-1) == 0.
func addBinaryConstraint(col string, schema *air.Schema) {
	one := fr.NewElement(1)
	// Construct X
	X := &air.ColumnAccess{Column: col, Shift: 0}
	// Construct X-1
	X_m1 := &air.Sub{Args: []air.Expr{X, &air.Constant{Value: &one}}}
	// Construct X * (X-1)
	X_X_m1 := &air.Mul{Args: []air.Expr{X, X, X_m1}}
	// Done!
	schema.AddVanishingConstraint(col, nil, X_X_m1)
}

// Enforce a constraint for a given column in the schema to ensure all
// values in that column fit within a given number of bits.  This is
// implemented using a byte decomposition which adds n columns and a
// vanishing constraint (where n*8 >= nbits).
func addBitwidthConstraint(col string, nbits uint, schema *air.Schema) {
	if nbits%8 != 0 {
		panic("asymetric bitwidth constraints not yet supported")
	} else if nbits == 0 {
		panic("zero bitwidth constraint encountered")
	}
	// Calculate how many bytes required.
	n := nbits / 8
	es := make([]air.Expr, n)
	fr256 := fr.NewElement(256)
	coefficient := fr.NewElement(1)
	// Construct Columns
	for i := uint(0); i < n; i++ {
		// Determine name for the ith byte column
		colName := fmt.Sprintf("%s:%d", col, i)
		// Create Column + Constraint
		schema.AddColumn(colName, true)
		schema.AddRangeConstraint(colName, &fr256)
		// Create ith term (for final sum)
		var coeff fr.Element
		// Clone coefficient
		coeff.Set(&coefficient)

		Xi := &air.ColumnAccess{Column: colName, Shift: 0}
		es[i] = &air.Mul{Args: []air.Expr{Xi, &air.Constant{Value: &coeff}}}
		// Update coefficient
		coefficient.Mul(&coefficient, &fr256)
	}
	// Construct (X:0 * 1) + ... + (X:n * 2^n)
	sum := &air.Add{Args: es}
	// Construct X == (X:0 * 1) + ... + (X:n * 2^n)
	X := &air.ColumnAccess{Column: col, Shift: 0}
	eq := &air.Sub{Args: []air.Expr{X, sum}}
	schema.AddVanishingConstraint(col, nil, eq)
	// Finally, add the necessary byte decomposition computation.
	schema.AddComputation(table.NewByteDecomposition(col, nbits))
}

// Lower a permutation to the AIR level.  This has quite a few
// effects.  Firstly, permutation constraints are added for all of the
// new columns.  Secondly, sorting constraints (and their associated
// synthetic columns) must also be added.  Finally, a trace
// computation is required to ensure traces are correctly expanded to
// meet the requirements of a sorted permutation.
func lowerPermutationToAir(c *table.SortedPermutation, schema *air.Schema) {
	ncols := len(c.Targets)
	// Add individual permutation constraints
	for i := 0; i < ncols; i++ {
		schema.AddColumn(c.Targets[i], true)
		schema.AddPermutationConstraint(c.Targets[i], c.Sources[i])
	}
	// Add sorting constraints + columns
	// Add trace computation
	schema.AddComputation(c)
}

// ExpandTrace expands a given trace according to this schema.
func (p *Schema) ExpandTrace(tr table.Trace) error {
	// Expand all the permutation columns
	for _, perm := range p.permutations {
		err := perm.ExpandTrace(tr)
		if err != nil {
			return err
		}
	}

	return nil
}
