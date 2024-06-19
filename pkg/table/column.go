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
	// Expected type of values held in this column.  Observe that this type is
	// enforced only when checking is enabled.  Unchecked typed columns can
	// still make sense when their values are implied by some other constraint.
	Type T
	// Indicates whether or not this column was created by the compiler (i.e. is
	// synthetic), or was specified by the user (i.e. is natural).
	Synthetic bool
}

// NewDataColumn constructs a new data column with a given name.
func NewDataColumn[T Type](name string, base T, synthetic bool) *DataColumn[T] {
	return &DataColumn[T]{name, base, synthetic}
}

// Width forms part of the ColumnGroup interface, and provides access to the
// ith column in a group.  Data columns already represent a group of size 1.
func (c *DataColumn[T]) Width() uint {
	return 1
}

// NameOf forms part of the ColumnGroup interface, and provides access to the
// ith column in a group.  Since a data column represents a group of size 1,
// there is only ever one name.
func (c *DataColumn[T]) NameOf(index uint) string {
	return c.Name
}

// Accepts determines whether or not this column accepts the given trace.  For a
// data column, this means ensuring that all elements are value for the columns
// type.
//
//nolint:revive
func (c *DataColumn[T]) Accepts(tr Trace) error {
	// Only check for non-field types.  This is simply because a column with the
	// field type always accepts everything.
	if c.Type.AsField() == nil {
		// Access corresponding column in trace
		col := tr.ColumnByName(c.Name)
		// Check constraints accepted
		for i := 0; i < int(tr.Height()); i++ {
			val := col.Get(i)

			if !c.Type.Accept(val) {
				// Construct useful error message
				msg := fmt.Sprintf("column %s value out-of-bounds (row %d, %s)", c.Name, i, val)
				// Evaluation failure
				return errors.New(msg)
			}
		}
	}
	// All good
	return nil
}

//nolint:revive
func (c *DataColumn[T]) String() string {
	if c.Type.AsField() != nil {
		return fmt.Sprintf("(column %s)", c.Name)
	}

	return fmt.Sprintf("(column %s :%s)", c.Name, c.Type)
}

// ComputedColumn describes a column whose values are computed on-demand, rather
// than being stored in a data array.  Typically computed columns read values
// from other columns in a trace in order to calculate their value.  There is an
// expectation that this computation is acyclic.  Furthermore, computed columns
// give rise to "trace expansion".  That is where the initial trace provided by
// the user is expanded by determining the value of all computed columns.
type ComputedColumn[E Evaluable] struct {
	Name string
	// The computation which accepts a given trace and computes
	// the value of this column at a given row.
	Expr E
}

// NewComputedColumn constructs a new computed column with a given name and
// determining expression.  More specifically, that expression is used to
// compute the values for this column during trace expansion.
func NewComputedColumn[E Evaluable](name string, expr E) *ComputedColumn[E] {
	return &ComputedColumn[E]{
		Name: name,
		Expr: expr,
	}
}

// RequiredSpillage returns the minimum amount of spillage required to ensure
// this column can be correctly computed in the presence of arbitrary (front)
// padding.
func (c *ComputedColumn[E]) RequiredSpillage() uint {
	// NOTE: Spillage is only currently considered to be necessary at the front
	// (i.e. start) of a trace.  This is because padding is always inserted at
	// the front, never the back.  As such, it is the maximum positive shift
	// which determines how much spillage is required for this comptuation.
	return c.Expr.Bounds().End
}

// Accepts determines whether or not this column accepts the given trace.  For a
// data column, this means ensuring that all elements are value for the columns
// type.
//
//nolint:revive
func (c *ComputedColumn[E]) Accepts(tr Trace) error {
	// Check column in trace!
	if !tr.HasColumn(c.Name) {
		return fmt.Errorf("Trace missing computed column ({%s})", c.Name)
	}

	return nil
}

// ExpandTrace attempts to a new column to the trace which contains the result
// of evaluating a given expression on each row.  If the column already exists,
// then an error is flagged.
func (c *ComputedColumn[E]) ExpandTrace(tr Trace) error {
	if tr.HasColumn(c.Name) {
		msg := fmt.Sprintf("Computed column already exists ({%s})", c.Name)
		return errors.New(msg)
	}

	data := make([]*fr.Element, tr.Height())
	// Expand the trace
	for i := 0; i < len(data); i++ {
		val := c.Expr.EvalAt(i, tr)
		if val != nil {
			data[i] = val
		} else {
			zero := fr.NewElement(0)
			data[i] = &zero
		}
	}
	// Determine padding value.  A negative row index is used here to ensure
	// that all columns return their padding value which is then used to compute
	// the padding value for *this* column.
	padding := c.Expr.EvalAt(-1, tr)
	// Colunm needs to be expanded.
	tr.AddColumn(c.Name, data, padding)
	// Done
	return nil
}

// nolint:revive
func (c *ComputedColumn[E]) String() string {
	return fmt.Sprintf("(compute %s %s)", c.Name, any(c.Expr))
}

// ===================================================================
// Sorted Permutations
// ===================================================================

// Permutation declares a constraint that one column is a permutation
// of another.
type Permutation struct {
	// The target columns
	Targets []string
	// The so columns
	Sources []string
}

// NewPermutation creates a new permutation
func NewPermutation(targets []string, sources []string) *Permutation {
	if len(targets) != len(sources) {
		panic("differeng number of target / source permutation columns")
	}

	return &Permutation{targets, sources}
}

// RequiredSpillage returns the minimum amount of spillage required to ensure
// valid traces are accepted in the presence of arbitrary padding.
func (p *Permutation) RequiredSpillage() uint {
	return uint(0)
}

// Accepts checks whether a permutation holds between the source and
// target columns.
func (p *Permutation) Accepts(tr Trace) error {
	// Sanity check columns well formed.
	if err := validPermutationColumns(p.Targets, p.Sources, tr); err != nil {
		return err
	}
	// Slice out data
	src := sliceMatchingColumns(p.Sources, tr)
	dst := sliceMatchingColumns(p.Targets, tr)
	// Sanity check whether column exists
	if !util.ArePermutationOf(dst, src) {
		msg := fmt.Sprintf("Target columns (%v) not permutation of source columns ({%v})",
			p.Targets, p.Sources)
		return errors.New(msg)
	}
	// Success
	return nil
}

func (p *Permutation) String() string {
	targets := ""
	sources := ""

	for i, s := range p.Targets {
		if i != 0 {
			targets += " "
		}

		targets += s
	}

	for i, s := range p.Sources {
		if i != 0 {
			sources += " "
		}

		sources += s
	}

	return fmt.Sprintf("(permutation (%s) (%s))", targets, sources)
}

// ===================================================================
// Sorted Permutations
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

// RequiredSpillage returns the minimum amount of spillage required to ensure
// valid traces are accepted in the presence of arbitrary padding.
func (p *SortedPermutation) RequiredSpillage() uint {
	return uint(0)
}

// Accepts checks whether a sorted permutation holds between the
// source and target columns.
func (p *SortedPermutation) Accepts(tr Trace) error {
	// Sanity check columns well formed.
	if err := validPermutationColumns(p.Targets, p.Sources, tr); err != nil {
		return err
	}
	// Slice out data
	src := sliceMatchingColumns(p.Sources, tr)
	dst := sliceMatchingColumns(p.Targets, tr)
	// Sanity check whether column exists
	if !util.ArePermutationOf(dst, src) {
		msg := fmt.Sprintf("Target columns (%v) not permutation of source columns ({%v})",
			p.Targets, p.Sources)
		return errors.New(msg)
	}
	// Check that target columns are sorted lexicographically.
	if util.AreLexicographicallySorted(dst, p.Signs) {
		return nil
	}
	// Error case
	msg := fmt.Sprintf("Permutation columns not lexicographically sorted ({%s})", p.Targets)
	// Done
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
		data := tr.ColumnByName(src).Data()
		// Copy column data to initialise permutation.
		cols[i] = make([]*fr.Element, len(data))
		copy(cols[i], data)
	}
	// Sort target columns
	util.PermutationSort(cols, p.Signs)
	// Physically add the columns
	for i := 0; i < len(p.Targets); i++ {
		dstColName := p.Targets[i]
		srcCol := tr.ColumnByName(p.Sources[i])
		tr.AddColumn(dstColName, cols[i], srcCol.Padding())
	}
	//
	return nil
}

// String returns a string representation of this constraint.  This is primarily
// used for debugging.
func (p *SortedPermutation) String() string {
	targets := ""
	sources := ""

	for i, s := range p.Targets {
		if i != 0 {
			targets += " "
		}

		targets += s
	}

	for i, s := range p.Sources {
		if i != 0 {
			sources += " "
		}

		if p.Signs[i] {
			sources += fmt.Sprintf("+%s", s)
		} else {
			sources += fmt.Sprintf("-%s", s)
		}
	}

	return fmt.Sprintf("(permute (%s) (%s))", targets, sources)
}

func validPermutationColumns(targets []string, sources []string, tr Trace) error {
	ncols := len(targets)
	// Sanity check matching length
	if len(sources) != ncols {
		return fmt.Errorf("Number of source and target columns differs")
	}
	// Check required columns in trace
	for i := 0; i < ncols; i++ {
		if !tr.HasColumn(targets[i]) {
			return fmt.Errorf("Trace missing permutation target column ({%s})", targets[i])
		} else if !tr.HasColumn(sources[i]) {
			return fmt.Errorf("Trace missing permutation source ({%s})", sources[i])
		}
	}
	//
	return nil
}

func sliceMatchingColumns(names []string, tr Trace) [][]*fr.Element {
	// Allocate return array
	cols := make([][]*fr.Element, len(names))
	// Slice out the data
	for i, n := range names {
		nth := tr.ColumnByName(n)
		cols[i] = nth.Data()
	}
	// Done
	return cols
}
