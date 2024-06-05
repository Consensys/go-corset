package air

import (
	"errors"
	"fmt"

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
		if c.Name == name {
			return true
		}
	}

	return false
}

// IsInputTrace determines whether a given input trace is a suitable
// input (i.e. non-expanded) trace for this schema.  Specifically, the
// input trace must contain a matching column for each non-synthetic
// column in this trace.
func (p *Schema) IsInputTrace(tr table.Trace) error {
	count := 0

	for _, c := range p.dataColumns {
		if !c.Synthetic && !tr.HasColumn(c.Name) {
			msg := fmt.Sprintf("Trace missing input column ({%s})", c.Name)
			return errors.New(msg)
		} else if c.Synthetic && tr.HasColumn(c.Name) {
			msg := fmt.Sprintf("Trace has synthetic column ({%s})", c.Name)
			return errors.New(msg)
		} else if !c.Synthetic {
			count = count + 1
		}
	}
	// Check geometry
	if tr.Width() != count {
		// Determine the unknown columns for error reporting.
		unknown := make([]string, 0)

		for i := 0; i < tr.Width(); i++ {
			n := tr.ColumnName(i)
			if !p.HasColumn(n) {
				unknown = append(unknown, n)
			}
		}

		msg := fmt.Sprintf("Trace has unknown columns {%s}", unknown)

		return errors.New(msg)
	}
	// Done
	return nil
}

// IsOutputTrace determines whether a given input trace is a suitable
// output (i.e. expanded) trace for this schema.  Specifically, the
// output trace must contain a matching column for each column in this
// trace (synthetic or otherwise).
func (p *Schema) IsOutputTrace(tr table.Trace) error {
	count := 0

	for _, c := range p.dataColumns {
		if !tr.HasColumn(c.Name) {
			msg := fmt.Sprintf("Trace missing input column ({%s})", c.Name)
			return errors.New(msg)
		}

		count++
	}
	// Check geometry
	if tr.Width() != count {
		return errors.New("Trace has unknown columns")
	}
	// Done
	return nil
}

// AddColumn appends a new data column which is either synthetic or
// not.  A synthetic column is one which has been introduced by the
// process of lowering from HIR / MIR to AIR.  That is, it is not a
// column which was original specified by the user.
func (p *Schema) AddColumn(name string, synthetic bool) {
	// NOTE: the air level has no ability to enforce the type specified for a
	// given column.
	p.dataColumns = append(p.dataColumns, table.NewDataColumn(name, &table.FieldType{}, synthetic))
}

// AddComputation appends a new computation to be used during trace
// expansion for this schema.
func (p *Schema) AddComputation(c table.TraceComputation) {
	p.computations = append(p.computations, c)
}

// AddPermutationConstraint appends a new permutation constraint which
// ensures that one column is a permutation of another.
func (p *Schema) AddPermutationConstraint(target string, source string) {
	p.permutations = append(p.permutations, table.NewPermutation(target, source))
}

// AddVanishingConstraint appends a new vanishing constraint.
func (p *Schema) AddVanishingConstraint(handle string, domain *int, expr Expr) {
	p.vanishing = append(p.vanishing, table.NewRowConstraint(handle, domain, table.ZeroTest[Expr]{Expr: expr}))
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
	for _, c := range p.computations {
		err := c.ExpandTrace(tr)
		if err != nil {
			return err
		}
	}
	// Done
	return nil
}
