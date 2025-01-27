package assignment

import (
	sc "github.com/consensys/go-corset/pkg/schema"
	"github.com/consensys/go-corset/pkg/trace"
	"github.com/consensys/go-corset/pkg/util"
	"github.com/consensys/go-corset/pkg/util/field"
	"github.com/consensys/go-corset/pkg/util/sexp"
)

// ComputedColumn describes a column whose values are computed on-demand, rather
// than being stored in a data array.  Typically computed columns read values
// from other columns in a trace in order to calculate their value.  There is an
// expectation that this computation is acyclic.  Furthermore, computed columns
// give rise to "trace expansion".  That is where the initial trace provided by
// the user is expanded by determining the value of all computed columns.
type ComputedColumn[E sc.Evaluable] struct {
	target sc.Column
	// The computation which accepts a given trace and computes
	// the value of this column at a given row.
	expr E
}

// NewComputedColumn constructs a new computed column with a given name and
// determining expression.  More specifically, that expression is used to
// compute the values for this column during trace expansion.
func NewComputedColumn[E sc.Evaluable](context trace.Context, name string, expr E) *ComputedColumn[E] {
	column := sc.NewColumn(context, name, &sc.FieldType{})
	// FIXME: Determine computed columns type?
	return &ComputedColumn[E]{column, expr}
}

// Name returns the name of this computed column.
func (p *ComputedColumn[E]) Name() string {
	return p.target.Name
}

// ============================================================================
// Declaration Interface
// ============================================================================

// Context returns the evaluation context for this computed column.
func (p *ComputedColumn[E]) Context() trace.Context {
	return p.target.Context
}

// Columns returns the columns declared by this computed column.
func (p *ComputedColumn[E]) Columns() util.Iterator[sc.Column] {
	// TODO: figure out appropriate type for computed column
	return util.NewUnitIterator[sc.Column](p.target)
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

// ComputeColumns computes the values of columns defined by this assignment.
// Specifically, this creates a new column which contains the result of
// evaluating a given expression on each row.
func (p *ComputedColumn[E]) ComputeColumns(tr trace.Trace) ([]trace.ArrayColumn, error) {
	// Determine multiplied height
	height := tr.Height(p.target.Context)
	// Make space for computed data
	data := field.NewFrArray(height, 256)
	// Expand the trace
	for i := uint(0); i < data.Len(); i++ {
		val := p.expr.EvalAt(int(i), tr)
		data.Set(i, val)
	}
	// Determine padding value.  A negative row index is used here to ensure
	// that all columns return their padding value which is then used to compute
	// the padding value for *this* column.
	padding := p.expr.EvalAt(-1, tr)
	// Construct column
	col := trace.NewArrayColumn(p.target.Context, p.Name(), data, padding)
	// Done
	return []trace.ArrayColumn{col}, nil
}

// Dependencies returns the set of columns that this assignment depends upon.
// That can include both input columns, as well as other computed columns.
func (p *ComputedColumn[E]) Dependencies() []uint {
	return *p.expr.RequiredColumns()
}

// ============================================================================
// Lispify Interface
// ============================================================================

// Lisp converts this schema element into a simple S-Expression, for example
// so it can be printed.
func (p *ComputedColumn[E]) Lisp(schema sc.Schema) sexp.SExp {
	col := sexp.NewSymbol("computed")
	name := sexp.NewSymbol(p.Columns().Next().QualifiedName(schema))
	expr := p.expr.Lisp(schema)

	return sexp.NewList([]sexp.SExp{col, name, expr})
}
