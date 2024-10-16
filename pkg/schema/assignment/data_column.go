package assignment

import (
	sc "github.com/consensys/go-corset/pkg/schema"
	"github.com/consensys/go-corset/pkg/sexp"
	"github.com/consensys/go-corset/pkg/trace"
	"github.com/consensys/go-corset/pkg/util"
)

// DataColumn represents a column of user-provided values.
type DataColumn struct {
	// Context where this data column is located.
	context trace.Context
	// Name of this datacolumn
	name string
	// Expected type of values held in this column.  Observe that this should be
	// true for the input columns for any valid trace and, furthermore, every
	// computed column should have values of this type.
	datatype sc.Type
}

// NewDataColumn constructs a new data column with a given name.
func NewDataColumn(context trace.Context, name string, base sc.Type) *DataColumn {
	return &DataColumn{context, name, base}
}

// Context returns the evaluation context for this column.
func (p *DataColumn) Context() trace.Context {
	return p.context
}

// Module identifies the module which encloses this column.
func (p *DataColumn) Module() uint {
	return p.context.Module()
}

// Name provides access to information about the ith column in a schema.
func (p *DataColumn) Name() string {
	return p.name
}

// Type Returns the expected type of data in this column
func (p *DataColumn) Type() sc.Type {
	return p.datatype
}

// ============================================================================
// Declaration Interface
// ============================================================================

// Columns returns the columns declared by this computed column.
func (p *DataColumn) Columns() util.Iterator[sc.Column] {
	// Datacolumns always have a multiplier of 1.
	column := sc.NewColumn(p.context, p.name, p.datatype)
	return util.NewUnitIterator[sc.Column](column)
}

// IsComputed Determines whether or not this declaration is computed (which data
// columns never are).
func (p *DataColumn) IsComputed() bool {
	return false
}

// ============================================================================
// Lispify Interface
// ============================================================================

// Lisp converts this schema element into a simple S-Expression, for example
// so it can be printed.
func (p *DataColumn) Lisp(schema sc.Schema) sexp.SExp {
	col := sexp.NewSymbol("defcolumns")
	name := sexp.NewSymbol(p.Columns().Next().QualifiedName(schema))
	//
	if p.datatype.AsField() != nil {
		return sexp.NewList([]sexp.SExp{col, name})
	}
	//
	datatype := sexp.NewSymbol(p.datatype.String())
	def := sexp.NewList([]sexp.SExp{name, datatype})
	//
	return sexp.NewList([]sexp.SExp{col, def})
}
