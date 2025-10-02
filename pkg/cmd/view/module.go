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
package view

import (
	"fmt"

	"github.com/consensys/go-corset/pkg/corset"
	"github.com/consensys/go-corset/pkg/trace"
	tr "github.com/consensys/go-corset/pkg/trace"
	"github.com/consensys/go-corset/pkg/util/field"
)

// ModuleView abstracts an underlying trace module.  For example, it manages the
// way that column data is displayed (e.g. in hex or in decimal, etc), whether
// or not register limbs are shown, and provides a mechanism for querying
// whether a given cell is "active" or not.  Specifically, cells in a
// perspective are not active when that perspective is not active.
type ModuleView interface {
	// CellAt returns the contents of a specific cell in this table.
	CellAt(col uint, row uint) string
	// Column returns the title of the given column.
	Column(uint) string
	// Height returns the number of rows in this table.
	Height() uint
	// Highlighted determines whether a given cell should be highlighted or not.
	Highlighted(col uint, row uint) bool
	// Name returns the name of the given module
	Name() string
	// Rowe returns the title of the given row
	Row(uint) string
	// Width returns the number of columns in this table.
	Width() uint
}

// ============================================================================
// View Implementation
// ============================================================================

type moduleView[F field.Element[F]] struct {
	// Module identifier
	id uint
	// Trace provides the raw data for this view
	trace trace.Module[F]
	// Srcmap provides relevant display information.
	srcmap corset.SourceModule
	// Padding to use
	padding uint
	// Filter determines which bits of this view are shown
	filter ColumnFilter
	// Data provides the data.  If this is nil, then it needs to be recomputed.
	data *moduleData
}

// CellAt returns the contents of a specific cell in this table.
func (p *moduleView[F]) CellAt(col uint, row uint) string {
	return p.get().CellAt(col, row)
}

// Column returns the title of the given column.
func (p *moduleView[F]) Column(col uint) string {
	return p.get().Column(col)
}

// Filter columns in this module
func (p *moduleView[F]) Filter(filter ColumnFilter) moduleView[F] {
	var q = *p
	// NOTE: technically the following should conjunct the two filters together.
	// However, for now, we don't bother.
	q.filter = filter
	// Reset data to force it to be recomputed.
	q.data = nil
	//
	return q
}

// Height returns the number of rows in this table.
func (p *moduleView[F]) Height() uint {
	return p.get().Height()
}

// Highlighted determines whether a given cell should be highlighted or not.
func (p *moduleView[F]) Highlighted(col uint, row uint) bool {
	return p.get().Highlighted(col, row)
}

// Name return name of this module
func (p *moduleView[F]) Name() string {
	return p.trace.Name()
}

// Rowe returns the title of the given row
func (p *moduleView[F]) Row(row uint) string {
	return p.get().Row(row)
}

// Width returns the number of columns in this table.
func (p *moduleView[F]) Width() uint {
	return p.get().Width()
}

func (p *moduleView[F]) get() *moduleData {
	if p.data == nil {
		p.data = buildWindowData(p.filter, p.padding, p.trace, p.srcmap)
	}
	//
	return p.data
}

// ============================================================================
// Module Data
// ============================================================================

type moduleData struct {
	// Set of rows in this window
	rows []string
	// Set of columns in this window
	columns []columnData
	// Highlighted cells in this window
	highlights []bool
}

// CellAt returns the contents of a specific cell in this table.
func (p *moduleData) CellAt(col uint, row uint) string {
	return p.columns[col].data[row]
}

// Column returns the title of the given column.
func (p *moduleData) Column(col uint) string {
	return p.columns[col].name
}

// Height returns the number of rows in this table.
func (p *moduleData) Height() uint {
	return uint(len(p.rows))
}

// Highlighted determines whether a given cell should be highlighted or not.
func (p *moduleData) Highlighted(col uint, row uint) bool {
	var ncols = uint(len(p.columns))
	//
	return p.highlights[(ncols*row)+col]
}

// Rowe returns the title of the given row
func (p *moduleData) Row(row uint) string {
	return p.rows[row]
}

// Width returns the number of columns in this table.
func (p *moduleData) Width() uint {
	return uint(len(p.columns))
}

// ============================================================================
// Column Data
// ============================================================================

type columnData struct {
	// column identifier
	id uint
	// column name
	name string
	// rendered column data
	data []string
}

// ============================================================================
// Helpers
// ============================================================================

func buildWindowData[F field.Element[F]](filter ColumnFilter, padding uint, trace tr.Module[F],
	srcmap corset.SourceModule) *moduleData {
	//
	var (
		first, last = boundWindowRows(trace.Width(), trace.Height(), filter, padding)
		rows        = buildWindowRows(first, last)
		cols        = buildWindowColumns(first, last, filter, trace, srcmap)
		highlights  = buildWindowHighlights(first, last, cols, filter)
	)

	return &moduleData{rows, cols, highlights}
}

func buildWindowColumns[F field.Element[F]](first, last uint, filter ColumnFilter,
	trace tr.Module[F], srcmap corset.SourceModule) []columnData {
	//
	var data []columnData
	//
	for _, c := range srcmap.Columns {
		// Dig out the column id
		cid := c.Register.Column().Unwrap()
		//
		if filter.Column(c.Register.Column().Unwrap()) != nil {
			data = append(data, columnData{
				id:   cid,
				name: c.Name,
				data: buildWindowColumnData(first, last, trace.Column(cid)),
			})
		}
	}
	//
	return data
}

func buildWindowColumnData[F field.Element[F]](first, last uint, column tr.Column[F]) []string {
	var data = make([]string, last-first)
	//
	for i := first; i < last; i++ {
		// FIXME: this is a very limited conversion at this time.
		ith := column.Data().Get(i)
		data[i-first] = ith.Text(16)
	}
	//
	return data
}

func buildWindowRows(start, end uint) []string {
	var rows = make([]string, end-start)

	for row := start; row < end; row++ {
		rows[row-start] = fmt.Sprintf("%d", row)
	}

	return rows
}

func buildWindowHighlights(start, end uint, cols []columnData, filter ColumnFilter) []bool {
	var (
		ncols = uint(len(cols))
		nrows = end - start
		rows  = make([]bool, ncols*nrows)
	)
	//
	for row := start; row < end; row++ {
		for c, col := range cols {
			if f := filter.Column(col.id); f != nil && f.Cell(row) {
				r := row - start
				rows[(r*ncols)+uint(c)] = true
			}
		}
	}
	//
	return rows
}

func boundWindowRows(width, height uint, filter ColumnFilter, padding uint) (first, last uint) {
	var (
		m = minWindowRow(width, height, filter) - int(padding)
		n = maxWindowRow(width, height, filter) + 1 + int(padding)
	)
	//
	return uint(max(m, 0)), min(uint(n), height)
}

func minWindowRow(width, height uint, filter ColumnFilter) int {
	// Find column with least cell in the filter
	for i := range height {
		for j := range width {
			if f := filter.Column(j); f != nil && f.Cell(i) {
				return int(i)
			}
		}
	}
	// Suggests an empty filter
	panic("unreachable")
}

func maxWindowRow(width, height uint, filter ColumnFilter) int {
	// Find column with greatest cell in the filter
	for i := height; i > 0; {
		i = i - 1
		//
		for j := range width {
			if f := filter.Column(j); f != nil && f.Cell(i) {
				return int(i)
			}
		}
	}
	// Suggests an empty filter
	panic("unreachable")
}
