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
	tr "github.com/consensys/go-corset/pkg/trace"
	"github.com/consensys/go-corset/pkg/util"
	"github.com/consensys/go-corset/pkg/util/collection/iter"
	"github.com/consensys/go-corset/pkg/util/field"
	"github.com/consensys/go-corset/pkg/util/sexp"
)

// Interleaving generates a new column by interleaving two or more existing
// colummns.  For example, say Z interleaves X and Y (in that order) and we have
// a trace X=[1,2], Y=[3,4].  Then, the interleaved column Z has the values
// Z=[1,3,2,4].
type Interleaving struct {
	// The new (interleaved) column
	Target sc.Column
	// Sources are the columns used by this interleaving to define the new
	// (interleaved) column.
	Sources []uint
}

// NewInterleaving constructs a new interleaving assignment.
func NewInterleaving(context tr.Context, name string, sources []uint, datatype sc.Type) *Interleaving {
	if context.LengthMultiplier()%uint(len(sources)) != 0 {
		panic(fmt.Sprintf("length multiplier (%d) for column %s not divisible by number of columns (%d)",
			context.LengthMultiplier(), name, len(sources)))
	}
	// Fixme: determine interleaving type
	target := sc.NewColumn(context, name, datatype)

	return &Interleaving{target, sources}
}

// Module returns the module which encloses this interleaving.
func (p *Interleaving) Module() uint {
	return p.Target.Context.Module()
}

// ============================================================================
// Declaration Interface
// ============================================================================

// Context returns the evaluation context for this interleaving.
func (p *Interleaving) Context() tr.Context {
	return p.Target.Context
}

// Columns returns the column declared by this interleaving.
func (p *Interleaving) Columns() iter.Iterator[sc.Column] {
	return iter.NewUnitIterator(p.Target)
}

// IsComputed Determines whether or not this declaration is computed (which an
// interleaving column is by definition).
func (p *Interleaving) IsComputed() bool {
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
func (p *Interleaving) Bounds() util.Bounds {
	return util.EMPTY_BOUND
}

// ComputeColumns computes the values of columns defined by this assignment.
// This requires copying the data in the source columns to create the
// interleaved column.
func (p *Interleaving) ComputeColumns(trace tr.Trace) ([]tr.ArrayColumn, error) {
	ctx := p.Target.Context
	// Byte width records the largest width of any column.
	bit_width := uint(0)
	// Ensure target column doesn't exist
	for i := p.Columns(); i.HasNext(); {
		ith := i.Next()
		// Update byte width
		bit_width = max(bit_width, ith.DataType.BitWidth())
	}
	// Determine interleaving width
	width := uint(len(p.Sources))
	// Following division should always produce whole value because the length
	// multiplier already includes the width as a factor.
	height := trace.Height(ctx) / width
	// Construct empty array
	data := field.NewFrArray(height*width, bit_width)
	// Offset just gives the column index
	offset := uint(0)
	// Copy interleaved data
	for i := uint(0); i < width; i++ {
		// Lookup source column
		col := trace.Column(p.Sources[i])
		// Copy over
		for j := uint(0); j < height; j++ {
			data.Set(offset+(j*width), col.Get(int(j)))
		}

		offset++
	}
	// Padding for the entire column is determined by the padding for the first
	// column in the interleaving.
	padding := trace.Column(p.Sources[0]).Padding()
	// Colunm needs to be expanded.
	col := tr.NewArrayColumn(ctx, p.Target.Name, data, padding)
	//
	return []tr.ArrayColumn{col}, nil
}

// Dependencies returns the set of columns that this assignment depends upon.
// That can include both input columns, as well as other computed columns.
func (p *Interleaving) Dependencies() []uint {
	return p.Sources
}

// CheckConsistency performs some simple checks that the given schema is
// consistent.  This provides a double check of certain key properties, such as
// that registers used for assignments are large enough, etc.
func (p *Interleaving) CheckConsistency(schema sc.Schema) error {
	var datatype sc.Type = nil
	// Determine type of source registers
	for _, src := range p.Sources {
		// Determine src type
		srcType := schema.Columns().Nth(src).DataType
		//
		if datatype == nil {
			datatype = srcType
		} else {
			datatype = sc.Join(datatype, srcType)
		}
	}
	// Check type matches
	if datatype.Cmp(p.Target.DataType) != 0 {
		return fmt.Errorf("inconsistent interleaving type (was %s, expected %s)", p.Target.DataType, datatype)
	}
	//
	return nil
}

// ============================================================================
// Lispify Interface
// ============================================================================

// Lisp converts this schema element into a simple S-Expression, for example
// so it can be printed.
func (p *Interleaving) Lisp(schema sc.Schema) sexp.SExp {
	target := sexp.NewSymbol(p.Target.QualifiedName(schema))
	sources := sexp.EmptyList()
	// Convert source columns
	for _, src := range p.Sources {
		sources.Append(sexp.NewSymbol(sc.QualifiedName(schema, src)))
	}
	// Add datatype (if non-field)
	datatype := sexp.NewSymbol(p.Target.DataType.String())
	multiplier := sexp.NewSymbol(fmt.Sprintf("x%d", p.Target.Context.LengthMultiplier()))
	def := sexp.NewList([]sexp.SExp{target, datatype, multiplier})
	// Construct S-Expression
	return sexp.NewList([]sexp.SExp{
		sexp.NewSymbol("interleaved"),
		def,
		sources,
	})
}

// ============================================================================
// Encoding / Decoding
// ============================================================================

func init() {
	gob.Register(sc.Declaration(&Interleaving{}))
}
