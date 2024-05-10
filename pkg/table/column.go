package table

import (
	"errors"
	"fmt"

	"github.com/consensys/gnark-crypto/ecc/bls12-377/fr"
	"github.com/consensys/go-corset/pkg/util"
)

// DataColumn represents a column of user-provided values.
type DataColumn[T Type] struct {
	Name string
	Type T
}

// NewDataColumn constructs a new data column with a given name.
func NewDataColumn[T Type](name string, base T) *DataColumn[T] {
	return &DataColumn[T]{name, base}
}

// Get the value of this column at a given row in a given trace.
func (c *DataColumn[T]) Get(row int, tr Trace) (*fr.Element, error) {
	return tr.GetByName(c.Name, row)
}

// Accepts determines whether or not this column accepts the given trace.  For a
// data column, this means ensuring that all elements are value for the columns
// type.
//
//nolint:revive
func (c *DataColumn[T]) Accepts(tr Trace) error {
	for i := 0; i < tr.Height(); i++ {
		val, err := tr.GetByName(c.Name, i)
		if err != nil {
			return err
		}

		if !c.Type.Accept(val) {
			// Construct useful error message
			msg := fmt.Sprintf("column %s value out-of-bounds (row %d, %s)", c.Name, i, val)
			// Evaluation failure
			return errors.New(msg)
		}
	}
	// All good
	return nil
}

// ComputedColumn describes a column whose values are computed on-demand, rather
// than being stored in a data array.  Typically computed columns read values
// from other columns in a trace in order to calculate their value.  There is an
// expectation that this computation is acyclic.  Furthermore, computed columns
// give rise to "trace expansion".  That is where the initial trace provided by
// the user is expanded by determining the value of all computed columns.
type ComputedColumn struct {
	Name string
	// The computation which accepts a given trace and computes
	// the value of this column at a given row.
	Expr Evaluable
}

// NewComputedColumn constructs a new computed column with a given name and
// determining expression.  More specifically, that expression is used to
// compute the values for this column during trace expansion.
func NewComputedColumn(name string, expr Evaluable) *ComputedColumn {
	return &ComputedColumn{
		Name: name,
		Expr: expr,
	}
}

// Get reads the value at a given row in a data column. This amounts to
// looking up that value in the array of values which backs it.
func (c *ComputedColumn) Get(row int, tr Trace) (*fr.Element, error) {
	// Compute value at given row
	return c.Expr.EvalAt(row, tr), nil
}

// ===================================================================
// Sored Permutations
// ===================================================================

// SortedPermutation declares one or more columns as sorted permutations of
// existing columns.
type SortedPermutation struct {
	// The new (sorted) columns
	Targets []string
	// The sorting criteria
	Signs []bool
	// The existing columns
	Sources []string
}

// NewSortedPermutation creates a new sorted permutation
func NewSortedPermutation(targets []string, signs []bool, sources []string) *SortedPermutation {
	if len(targets) != len(signs) || len(signs) != len(sources) {
		panic("target and source column widths must match")
	}

	return &SortedPermutation{targets, signs, sources}
}

// Accepts checks whether a sorted permutation holds between the
// source and target columns.
func (p *SortedPermutation) Accepts(tr Trace) error {
	ncols := len(p.Sources)
	cols := make([][]*fr.Element, ncols)
	// Check that source columns have the same height?

	// Check that target and source columns exist and are permutations of source
	// columns.
	for i := 0; i < ncols; i++ {
		dstName := p.Targets[i]
		srcName := p.Sources[i]
		// Access column data based on column name.
		dst := tr.ColumnByName(dstName)
		src := tr.ColumnByName(srcName)
		// Sanity check whether column exists
		if dst == nil {
			msg := fmt.Sprintf("Invalid target column for permutation ({%s})", dstName)
			return errors.New(msg)
		} else if src == nil {
			msg := fmt.Sprintf("Invalid source column for permutation ({%s})", srcName)
			return errors.New(msg)
		} else if !util.IsPermutationOf(dst, src) {
			msg := fmt.Sprintf("Target column (%s) not permutation of source ({%s})", dstName, srcName)
			return errors.New(msg)
		}

		cols[i] = dst
	}

	// Check that target columns are sorted lexicographically.
	if util.AreLexicographicallySorted(cols, p.Signs) {
		return nil
	}

	msg := fmt.Sprintf("Permutation columns not lexicographically sorted ({%s})", p.Targets)

	return errors.New(msg)
}

// ExpandTrace expands a given trace to include the columns specified by a given
// SortedPermutation.  This requires copying the data in the source columns, and
// sorting that data according to the permutation criteria.
func (p *SortedPermutation) ExpandTrace(tr Trace) error {
	// Ensure target columns don't exist
	for _, col := range p.Targets {
		if tr.HasColumn(col) {
			panic("target column already exists")
		}
	}

	cols := make([][]*fr.Element, len(p.Sources))
	// Construct target columns
	for i := 0; i < len(p.Targets); i++ {
		src := p.Sources[i]
		// Read column data to initialise permutation.
		data := tr.ColumnByName(src)
		// Copy column data to initialise permutation.
		cols[i] = make([]*fr.Element, len(data))
		copy(cols[i], data)
	}
	// Sort target columns
	util.PermutationSort(cols, p.Signs)
	// Physically add the columns
	for i := 0; i < len(p.Targets); i++ {
		col := p.Targets[i]
		tr.AddColumn(col, cols[i])
	}
	//
	return nil
}
