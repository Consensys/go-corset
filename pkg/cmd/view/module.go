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
	"math/big"

	sc "github.com/consensys/go-corset/pkg/schema"
	tr "github.com/consensys/go-corset/pkg/trace"
	"github.com/consensys/go-corset/pkg/util/field"
	"github.com/consensys/go-corset/pkg/util/math"
	"github.com/consensys/go-corset/pkg/util/termio"
	"github.com/consensys/go-corset/pkg/util/termio/widget"
)

// ModuleView abstracts an underlying trace module.  For example, it manages the
// way that column data is displayed (e.g. in hex or in decimal, etc), whether
// or not register limbs are shown, and provides a mechanism for querying
// whether a given cell is "active" or not.  Specifically, cells in a
// perspective are not active when that perspective is not active.
type ModuleView interface {
	widget.TableSource
	// Highlighted determines whether a given cell should be highlighted or not.
	Highlighted(col uint, row uint) bool
	// Name returns the name of the given module
	Name() string
}

// ============================================================================
// View Implementation
// ============================================================================

type moduleView[F field.Element[F]] struct {
	// Module identifier
	id uint
	// Trace provides the raw data for this view
	trace tr.Module[F]
	// Mapping describes how source-level registers are mapped into columns
	// (i.e. limbs) as found in the trace.
	mapping sc.RegisterLimbsMap
	// Padding to use
	padding uint
	// Limbs indicates whether or not to show the raw limbs, or the combined
	// source-level register.
	limbs bool
	// Filter determines which bits of this view are shown
	filter ColumnFilter
	// Data provides the data.  If this is nil, then it needs to be recomputed.
	data *moduleData
}

// CellAt returns the contents of a specific cell in this table.
func (p *moduleView[F]) CellAt(col uint, row uint) termio.FormattedText {
	return p.get().CellAt(col, row)
}

func (p *moduleView[F]) ColumnWidth(col uint) uint {
	return p.get().ColumnWidth(col)
}

func (p *moduleView[F]) Dimensions() (uint, uint) {
	return p.get().Dimensions()
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

func (p *moduleView[F]) get() *moduleData {
	if p.data == nil {
		p.data = renderModule(p.filter, p.padding, p.limbs, p.trace, p.mapping)
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

// TableSource

// CellAt returns the contents of a specific cell in this table.
func (p *moduleData) CellAt(col uint, row uint) termio.FormattedText {
	if col == 0 && row == 0 {
		return termio.NewText("")
	} else if col == 0 {
		return termio.NewText(p.columns[row-1].name)
	} else if row == 0 {
		return termio.NewText(p.rows[col-1])
	}
	// switch col <-> row
	var text = p.columns[row-1].data[col-1]
	//
	return termio.NewText(text)
}

// ColumnWidth implementation for TableSource inteface
func (p *moduleData) ColumnWidth(col uint) uint {
	// FIXME!!
	return 10
}

// Dimensions implementation for TableSource inteface
func (p *moduleData) Dimensions() (uint, uint) {
	return uint(len(p.rows)) + 1, uint(len(p.columns)) + 1
}

// ============================================================================
// Column Data
// ============================================================================

type columnData struct {
	// column name
	name string
	// limbs making up this column
	limbs []sc.RegisterId
	// rendered column data
	data []string
}

// ============================================================================
// Helpers
// ============================================================================

func renderModule[F field.Element[F]](filter ColumnFilter, padding uint, limbs bool, trace tr.Module[F],
	mapping sc.RegisterLimbsMap) *moduleData {
	//
	var (
		first, last = boundWindowRows(trace.Width(), trace.Height(), filter, padding)
		rows        = buildWindowRows(first, last)
		cols        []columnData
	)
	// Render as limbs or as registers directly
	if limbs {
		cols = renderColumnsFromLimbs(first, last, filter, trace, mapping)
	} else {
		cols = renderColumnsFromRegisters(first, last, filter, trace, mapping)
	}
	// Determine window highlights
	highlights := buildWindowHighlights(first, last, cols, filter)
	//
	return &moduleData{rows, cols, highlights}
}

// Render columns as registers directly by reconstructing their values from that
// held in the trace for their limbs.
func renderColumnsFromRegisters[F field.Element[F]](first, last uint, filter ColumnFilter,
	trace tr.Module[F], mapping sc.RegisterLimbsMap) []columnData {
	//
	var data []columnData
	// Iterate source-level registers
	for i, reg := range mapping.Registers() {
		// construct source-level register id
		rid := sc.NewRegisterId(uint(i))
		// determine corresponding limbs
		limbs := mapping.LimbIds(rid)
		//
		if columnIncluded(filter, limbs) {
			//
			data = append(data, columnData{
				limbs: limbs,
				// Determine column name
				name: reg.Name,
				// Render column data from all limbs
				data: renderColumnData(first, last, limbs, mapping, trace),
			})
		}
	}
	//
	return data
}

// Render columns as register limbs by taking their values from the trace directly.
func renderColumnsFromLimbs[F field.Element[F]](first, last uint, filter ColumnFilter,
	trace tr.Module[F], mapping sc.RegisterLimbsMap) []columnData {
	//
	var data []columnData
	// Iterate source-level registers
	for i := range mapping.Registers() {
		// construct source-level register id
		rid := sc.NewRegisterId(uint(i))
		// iterate register limbs
		for _, lid := range mapping.LimbIds(rid) {
			//
			limbs := []sc.LimbId{lid}
			//
			if columnIncluded(filter, limbs) {
				//
				data = append(data, columnData{
					limbs: limbs,
					// Determine column name
					name: mapping.Limb(lid).Name,
					// Render column data only from this limb
					data: renderColumnData(first, last, limbs, mapping, trace),
				})
			}
		}
	}
	//
	return data
}

// Check whether the given filter includes any if the limbs (or not).
func columnIncluded(filter ColumnFilter, limbs []sc.RegisterId) bool {
	for _, lid := range limbs {
		if filter.Column(lid.Unwrap()) != nil {
			return true
		}
	}
	//
	return false
}

// Render data for a given source-level column.  This requires combining
// the actual column data for all limbs back together (i.e. undoing register
// splitting).
func renderColumnData[F field.Element[F]](first, last uint, limbs []sc.RegisterId, mapping sc.RegisterLimbsMap,
	trace tr.Module[F]) []string {
	//
	var data = make([]string, last-first)
	//
	for i := first; i < last; i++ {
		data[i-first] = renderColumnDataRow(i, limbs, mapping, trace)
	}
	//
	return data
}

// Render an individual row in a given source-level column.
func renderColumnDataRow[F field.Element[F]](row uint, limbs []sc.RegisterId, mapping sc.RegisterLimbsMap,
	trace tr.Module[F]) string {
	var (
		bits  = uint(0)
		value big.Int
	)
	//
	for _, lid := range limbs {
		var (
			data    = trace.Column(lid.Unwrap()).Data()
			element = data.Get(row)
			limb    = mapping.Limb(lid)
			val     big.Int
		)
		// Construct value from field element
		val.SetBytes(element.Bytes())
		// Shift and add
		value.Add(&value, val.Mul(&val, math.Pow2(bits)))
		//
		bits += limb.Width
	}
	// FIXME: for now, text is always rendered in hex.
	return value.Text(16)
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
			if cellIncluded(row, filter, col.limbs) {
				r := row - start
				rows[(r*ncols)+uint(c)] = true
			}
		}
	}
	//
	return rows
}

// Check whether the given filter includes any if the limbs (or not).
func cellIncluded(row uint, filter ColumnFilter, limbs []sc.RegisterId) bool {
	for _, lid := range limbs {
		if f := filter.Column(lid.Unwrap()); f != nil && f.Cell(row) {
			return true
		}
	}
	//
	return false
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
