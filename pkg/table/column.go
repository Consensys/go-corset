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

// Get the value of this column at a given row in a given trace.
func (c *DataColumn[T]) Get(row int, tr Trace) *fr.Element {
	return tr.GetByName(c.Name, row)
}

// Accepts determines whether or not this column accepts the given trace.  For a
// data column, this means ensuring that all elements are value for the columns
// type.
//
//nolint:revive
func (c *DataColumn[T]) Accepts(tr Trace) error {
	// Check column in trace!
	if !tr.HasColumn(c.Name) {
		return fmt.Errorf("Trace missing data column ({%s})", c.Name)
	}
	// Check constraints accepted
	for i := 0; i < tr.Height(); i++ {
		val := tr.GetByName(c.Name, i)

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

// Accepts determines whether or not this column accepts the given trace.  For a
// data column, this means ensuring that all elements are value for the columns
// type.
//
//nolint:revive
func (c *ComputedColumn) Accepts(tr Trace) error {
	// Check column in trace!
	if !tr.HasColumn(c.Name) {
		return fmt.Errorf("Trace missing computed column ({%s})", c.Name)
	}

	return nil
}

// ExpandTrace attempts to a new column to the trace which contains the result
// of evaluating a given expression on each row.  If the column already exists,
// then an error is flagged.
func (c *ComputedColumn) ExpandTrace(tr Trace) error {
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
	// Colunm needs to be expanded.
	tr.AddColumn(c.Name, data)
	// Done
	return nil
}

func (c *ComputedColumn) String() string {
	return fmt.Sprintf("(compute %s %s)", c.Name, c.Expr)
}

// ===================================================================
// Sorted Permutations
// ===================================================================

// Permutation declares a constraint that one column is a permutation
// of another.
type Permutation struct {
	// The target column
	Target string
	// The so columns
	Source string
}

// NewPermutation creates a new permutation
func NewPermutation(target string, source string) *Permutation {
	return &Permutation{target, source}
}

// Accepts checks whether a permutation holds between the source and
// target columns.
func (p *Permutation) Accepts(tr Trace) error {
	// Check column in trace!
	if !tr.HasColumn(p.Target) {
		return fmt.Errorf("Trace missing permutation target column ({%s})", p.Target)
	} else if !tr.HasColumn(p.Source) {
		return fmt.Errorf("Trace missing permutation source column ({%s})", p.Source)
	}

	return IsPermutationOf(p.Target, p.Source, tr)
}

func (p *Permutation) String() string {
	return fmt.Sprintf("(permutation %s %s)", p.Target, p.Source)
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

// Accepts checks whether a sorted permutation holds between the
// source and target columns.
func (p *SortedPermutation) Accepts(tr Trace) error {
	ncols := len(p.Sources)
	cols := make([][]*fr.Element, ncols)
	// Check required columns in trace
	for _, n := range p.Targets {
		if !tr.HasColumn(n) {
			return fmt.Errorf("Trace missing permutation target column ({%s})", n)
		}
	}

	for _, n := range p.Sources {
		if !tr.HasColumn(n) {
			return fmt.Errorf("Trace missing permutation source ({%s})", n)
		}
	}
	// Check that target and source columns exist and are permutations of source
	// columns.
	for i := 0; i < ncols; i++ {
		dstName := p.Targets[i]
		srcName := p.Sources[i]
		// Access column data based on column name.
		err := IsPermutationOf(dstName, srcName, tr)
		if err != nil {
			return err
		}

		cols[i] = tr.ColumnByName(dstName)
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

// IsPermutationOf checks whether (or not) one column is a permutation
// of another in given trace.  The order in which columns are given is
// not important.
func IsPermutationOf(target string, source string, tr Trace) error {
	dst := tr.ColumnByName(target)
	src := tr.ColumnByName(source)
	// Sanity check whether column exists
	if dst == nil {
		msg := fmt.Sprintf("Invalid target column for permutation ({%s})", target)
		return errors.New(msg)
	} else if src == nil {
		msg := fmt.Sprintf("Invalid source column for permutation ({%s})", source)
		return errors.New(msg)
	} else if !util.IsPermutationOf(dst, src) {
		msg := fmt.Sprintf("Target column (%s) not permutation of source ({%s})", target, source)
		return errors.New(msg)
	}

	return nil
}
