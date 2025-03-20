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
	"encoding/gob"
	"fmt"

	sc "github.com/consensys/go-corset/pkg/schema"
	"github.com/consensys/go-corset/pkg/trace"
	"github.com/consensys/go-corset/pkg/util/collection/iter"
	"github.com/consensys/go-corset/pkg/util/source/sexp"
)

// DataColumn represents a column of user-provided values.
type DataColumn struct {
	// Context where this data column is located.
	TraceContext trace.Context
	// Name of this datacolumn
	ColumnName string
	// Expected type of values held in this column.  Observe that this should be
	// true for the input columns for any valid trace and, furthermore, every
	// computed column should have values of this type.
	DataType sc.Type
}

// NewDataColumn constructs a new data column with a given name.
func NewDataColumn(context trace.Context, name string, base sc.Type) *DataColumn {
	return &DataColumn{context, name, base}
}

// Context returns the evaluation context for this column.
func (p *DataColumn) Context() trace.Context {
	return p.TraceContext
}

// Module identifies the module which encloses this column.
func (p *DataColumn) Module() uint {
	return p.TraceContext.Module()
}

// Name provides access to information about the ith column in a schema.
func (p *DataColumn) Name() string {
	return p.ColumnName
}

// Type Returns the expected type of data in this column
func (p *DataColumn) Type() sc.Type {
	return p.DataType
}

// ============================================================================
// Declaration Interface
// ============================================================================

// Columns returns the columns declared by this computed column.
func (p *DataColumn) Columns() iter.Iterator[sc.Column] {
	// Datacolumns always have a multiplier of 1.
	column := sc.NewColumn(p.TraceContext, p.ColumnName, p.DataType)
	return iter.NewUnitIterator[sc.Column](column)
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
	col := sexp.NewSymbol("column")
	name := sexp.NewSymbol(p.Columns().Next().QualifiedName(schema))
	//
	datatype := sexp.NewSymbol(p.DataType.String())
	multiplier := sexp.NewSymbol(fmt.Sprintf("x%d", p.TraceContext.LengthMultiplier()))
	def := sexp.NewList([]sexp.SExp{name, datatype, multiplier})
	//
	return sexp.NewList([]sexp.SExp{col, def})
}

// ============================================================================
// Encoding / Decoding
// ============================================================================

func init() {
	gob.Register(sc.Declaration(&DataColumn{}))
}
