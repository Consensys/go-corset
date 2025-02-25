// Copyright Consensys Software Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License"); you may not use this file except in compliance with
// the License. You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
// an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
// specific language governing permissions and limitations under the License.
//
// SPDX-License-Identifier: Apache-2.0
package assignment

import (
	"fmt"

	sc "github.com/consensys/go-corset/pkg/schema"
	"github.com/consensys/go-corset/pkg/trace"
	"github.com/consensys/go-corset/pkg/util"
	"github.com/consensys/go-corset/pkg/util/collection/iter"
	"github.com/consensys/go-corset/pkg/util/field"
	"github.com/consensys/go-corset/pkg/util/sexp"
)

// ComputedColumn describes a column whose values are computed on-demand, rather
// than being stored in a data array.  Typically computed columns read values
// from other columns in a trace in order to calculate their value.  There is an
// expectation that this computation is acyclic.  Furthermore, computed columns
// give rise to "trace expansion".  That is where the initial trace provided by
// the user is expanded by determining the value of all computed columns.
type ComputedColumn struct {
	target sc.Column
	// The computation which accepts a given trace and computes
	// the value of this column at a given row.
	expr sc.Evaluable
}

// NewComputedColumn constructs a new computed column with a given name and
// determining expression.  More specifically, that expression is used to
// compute the values for this column during trace expansion.
func NewComputedColumn(context trace.Context, name string, datatype sc.Type,
	expr sc.Evaluable) *ComputedColumn {
	column := sc.NewColumn(context, name, datatype)
	// FIXME: Determine computed columns type?
	return &ComputedColumn{column, expr}
}

// Name returns the name of this computed column.
func (p *ComputedColumn) Name() string {
	return p.target.Name
}

// ============================================================================
// Declaration Interface
// ============================================================================

// Context returns the evaluation context for this computed column.
func (p *ComputedColumn) Context() trace.Context {
	return p.target.Context
}

// Columns returns the columns declared by this computed column.
func (p *ComputedColumn) Columns() iter.Iterator[sc.Column] {
	// TODO: figure out appropriate type for computed column
	return iter.NewUnitIterator[sc.Column](p.target)
}

// IsComputed Determines whether or not this declaration is computed (which it
// is).
func (p *ComputedColumn) IsComputed() bool {
	return true
}

// ============================================================================
// Assignment Interface
// ============================================================================

// Bounds determines the well-definedness bounds for this assignment for both
// the negative (left) or positive (right) directions.  For example, consider an
// expression such as "(shift X -1)".  This is technically undefined for the
// first row of any trace and, by association, any constraint evaluating this
// expression on that first row is also undefined (and hence must pass).
func (p *ComputedColumn) Bounds() util.Bounds {
	return p.expr.Bounds()
}

// ComputeColumns computes the values of columns defined by this assignment.
// Specifically, this creates a new column which contains the result of
// evaluating a given expression on each row.
func (p *ComputedColumn) ComputeColumns(tr trace.Trace) ([]trace.ArrayColumn, error) {
	// Determine multiplied height
	height := tr.Height(p.target.Context)
	// Make space for computed data
	data := field.NewFrArray(height, p.target.DataType.BitWidth())
	// Expand the trace
	for i := uint(0); i < data.Len(); i++ {
		val, err := p.expr.EvalAt(int(i), tr)
		// error check
		if err != nil {
			return nil, err
		}
		//
		data.Set(i, val)
	}
	// Determine padding value.  A negative row index is used here to ensure
	// that all columns return their padding value which is then used to compute
	// the padding value for *this* column.
	padding, err := p.expr.EvalAt(-1, tr)
	// Construct column
	col := trace.NewArrayColumn(p.target.Context, p.Name(), data, padding)
	// Done
	return []trace.ArrayColumn{col}, err
}

// Dependencies returns the set of columns that this assignment depends upon.
// That can include both input columns, as well as other computed columns.
func (p *ComputedColumn) Dependencies() []uint {
	return *p.expr.RequiredColumns()
}

// ============================================================================
// Lispify Interface
// ============================================================================

// Lisp converts this schema element into a simple S-Expression, for example
// so it can be printed.
func (p *ComputedColumn) Lisp(schema sc.Schema) sexp.SExp {
	col := sexp.NewSymbol("computed")
	name := sexp.NewSymbol(p.Columns().Next().QualifiedName(schema))
	datatype := sexp.NewSymbol(p.target.DataType.String())
	multiplier := sexp.NewSymbol(fmt.Sprintf("x%d", p.target.Context.LengthMultiplier()))
	def := sexp.NewList([]sexp.SExp{name, datatype, multiplier})
	expr := p.expr.Lisp(schema)

	return sexp.NewList([]sexp.SExp{col, def, expr})
}
