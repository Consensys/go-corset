package assignment

import (
	"fmt"

	"github.com/consensys/gnark-crypto/ecc/bls12-377/fr"
	"github.com/consensys/go-corset/pkg/schema"
	sc "github.com/consensys/go-corset/pkg/schema"
	"github.com/consensys/go-corset/pkg/trace"
	"github.com/consensys/go-corset/pkg/util"
)

// ComputedColumn describes a column whose values are computed on-demand, rather
// than being stored in a data array.  Typically computed columns read values
// from other columns in a trace in order to calculate their value.  There is an
// expectation that this computation is acyclic.  Furthermore, computed columns
// give rise to "trace expansion".  That is where the initial trace provided by
// the user is expanded by determining the value of all computed columns.
type ComputedColumn[E sc.Evaluable] struct {
	target schema.Column
	// The computation which accepts a given trace and computes
	// the value of this column at a given row.
	expr E
}

// NewComputedColumn constructs a new computed column with a given name and
// determining expression.  More specifically, that expression is used to
// compute the values for this column during trace expansion.
func NewComputedColumn[E sc.Evaluable](module uint, name string, multiplier uint, expr E) *ComputedColumn[E] {
	// FIXME: Determine computed columns type?
	column := schema.NewColumn(module, name, multiplier, &schema.FieldType{})
	return &ComputedColumn[E]{column, expr}
}

// nolint:revive
func (p *ComputedColumn[E]) String() string {
	return fmt.Sprintf("(compute %s %s)", p.Name(), any(p.expr))
}

// Name returns the name of this computed column.
func (p *ComputedColumn[E]) Name() string {
	return p.target.Name()
}

// ============================================================================
// Declaration Interface
// ============================================================================

// Columns returns the columns declared by this computed column.
func (p *ComputedColumn[E]) Columns() util.Iterator[schema.Column] {
	// TODO: figure out appropriate type for computed column
	return util.NewUnitIterator[schema.Column](p.target)
}

// IsComputed Determines whether or not this declaration is computed (which it
// is).
func (p *ComputedColumn[E]) IsComputed() bool {
	return true
}

// ============================================================================
// Assignment Interface
// ============================================================================

// RequiredSpillage returns the minimum amount of spillage required to ensure
// this column can be correctly computed in the presence of arbitrary (front)
// padding.
func (p *ComputedColumn[E]) RequiredSpillage() uint {
	// NOTE: Spillage is only currently considered to be necessary at the front
	// (i.e. start) of a trace.  This is because padding is always inserted at
	// the front, never the back.  As such, it is the maximum positive shift
	// which determines how much spillage is required for this comptuation.
	return p.expr.Bounds().End
}

// ExpandTrace attempts to a new column to the trace which contains the result
// of evaluating a given expression on each row.  If the column already exists,
// then an error is flagged.
func (p *ComputedColumn[E]) ExpandTrace(tr trace.Trace) error {
	columns := tr.Columns()
	// Check whether a column already exists with the given name.
	if columns.HasColumn(p.Name()) {
		return fmt.Errorf("column already exists ({%s})", p.Name())
	}
	// Extract length multipiler
	multiplier := p.target.LengthMultiplier()
	// Determine multiplied height
	height := tr.Modules().Get(p.target.Module()).Height() * multiplier
	// Make space for computed data
	data := make([]*fr.Element, height)
	// Expand the trace
	for i := 0; i < len(data); i++ {
		val := p.expr.EvalAt(i, tr)
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
	padding := p.expr.EvalAt(-1, tr)
	// Colunm needs to be expanded.
	columns.Add(trace.NewFieldColumn(p.target.Module(), p.Name(), multiplier, data, padding))
	// Done
	return nil
}
