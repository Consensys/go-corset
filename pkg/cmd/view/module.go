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
	"strings"

	"github.com/consensys/go-corset/pkg/corset"
	sc "github.com/consensys/go-corset/pkg/schema"
	tr "github.com/consensys/go-corset/pkg/trace"
	"github.com/consensys/go-corset/pkg/util/field"
	"github.com/consensys/go-corset/pkg/util/math"
	"github.com/consensys/go-corset/pkg/util/termio"
	"github.com/consensys/go-corset/pkg/util/termio/widget"
)

// ModuleConfig encapsulates various configuration settings for a ModuleView.
type ModuleConfig interface {
	// CellWidth reads the current maximum width of any cell in the table.
	CellWidth() uint

	// SetCellWidth determines the maximum width of any cells in the table.
	SetCellWidth(cellWidth uint)
}

// ModuleView abstracts an underlying trace module.  For example, it manages the
// way that column data is displayed (e.g. in hex or in decimal, etc), whether
// or not register limbs are shown, and provides a mechanism for querying
// whether a given cell is "active" or not.  Specifically, cells in a
// perspective are not active when that perspective is not active.
type ModuleView interface {
	widget.TableSource
	// Config returns config settings for this module.
	Config() ModuleConfig
	// Name returns the name of the given module
	Name() string
	// Return offset position within module
	Offset() (col uint, row uint)
	// Set offset position within module
	Goto(col, row uint)
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
	// Offset position within module
	x, y uint
	// display attribute
	display []uint
	// width of all cells
	cellWidth uint
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

// Offset returns the current offset position within module
func (p *moduleView[F]) Config() ModuleConfig {
	return p
}

// Offset returns the current offset position within module
func (p *moduleView[F]) Offset() (col uint, row uint) {
	return p.x, p.y
}

// Goto a specific offset within module
func (p *moduleView[F]) Goto(col, row uint) {
	width, height := p.get().Dimensions()

	p.x, p.y = min(width-1, col), min(height-1, row)
}

// ============================================================================
// ModuleConfig
// ============================================================================

// Offset returns the current offset position within module
func (p *moduleView[F]) CellWidth() uint {
	return p.cellWidth
}

// SetCellWidth determines the maximum width of any cells in the table.
func (p *moduleView[F]) SetCellWidth(cellWidth uint) {
	p.cellWidth = cellWidth
	// reset data to force recomputation
	p.data = nil
}

// ============================================================================
// TableSource
// ============================================================================

// CellAt returns the contents of a specific cell in this table.
func (p *moduleView[F]) CellAt(col uint, row uint) termio.FormattedText {
	return p.get().CellAt(col+p.x, row+p.y)
}

func (p *moduleView[F]) ColumnWidth(col uint) uint {
	var (
		data     = p.get()
		width, _ = data.Dimensions()
	)
	//
	if col+p.x >= width {
		return 10
	}
	//
	return p.get().ColumnWidth(col + p.x)
}

func (p *moduleView[F]) Dimensions() (uint, uint) {
	width, height := p.get().Dimensions()
	//
	if p.x < width {
		width -= p.x
	} else {
		width = 0
	}
	//
	if p.y < height {
		height -= p.y
	} else {
		height = 0
	}
	//
	return width, height
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

// Name return name of this module
func (p *moduleView[F]) Name() string {
	return p.trace.Name()
}

func (p *moduleView[F]) get() *moduleData {
	if p.data == nil {
		p.data = p.renderModule()
	}
	//
	return p.data
}

// ============================================================================
// Module Data
// ============================================================================

type moduleData struct {
	// Set of column titles
	columns []string
	// Max width of any column
	columnWidths []uint
	// Set of rows in this window
	rows []rowData
	// Highlighted cells in this window
	highlights []bool
}

// Highlighted determines whether a given cell should be highlighted or not.
func (p *moduleData) Highlighted(col uint, row uint) bool {
	var ncols = uint(len(p.rows))
	//
	return p.highlights[(ncols*row)+col]
}

// TableSource

// CellAt returns the contents of a specific cell in this table.
func (p *moduleData) CellAt(col uint, row uint) termio.FormattedText {
	if col == 0 && row == 0 {
		return termio.NewText("")
	} else if col == 0 {
		return termio.NewText(p.rows[row-1].name)
	} else if row == 0 {
		return termio.NewText(p.columns[col-1])
	}
	// switch col <-> row
	var text = p.rows[row-1].data[col-1]
	//
	return termio.NewText(text)
}

// ColumnWidth implementation for TableSource inteface
func (p *moduleData) ColumnWidth(col uint) uint {
	return p.columnWidths[col]
}

// Dimensions implementation for TableSource inteface
func (p *moduleData) Dimensions() (uint, uint) {
	return uint(len(p.columns)), uint(len(p.rows)) + 1
}

// ============================================================================
// Column Data
// ============================================================================

type rowData struct {
	// column name
	name string
	// limbs making up this row
	limbs []sc.RegisterId
	// rendered column data
	data []string
}

// ============================================================================
// Helpers
// ============================================================================

func (p *moduleView[F]) renderModule() *moduleData {
	//
	var (
		first, last = p.boundModuleRows()
		rows        = buildModuleColumns(first, last)
		cols        []rowData
	)
	// Render as limbs or as registers directly
	if p.limbs {
		cols = p.renderColumnsFromLimbs(first, last)
	} else {
		cols = p.renderColumnsFromRegisters(first, last)
	}
	// Determine row widths
	rowWidths := buildModuleColumnWidths(uint(len(rows)), cols)
	// Determine window highlights
	highlights := buildWindowHighlights(first, last, cols, p.filter)
	//
	return &moduleData{rows, rowWidths, cols, highlights}
}

// Render columns as registers directly by reconstructing their values from that
// held in the trace for their limbs.
func (p *moduleView[F]) renderColumnsFromRegisters(first, last uint) []rowData {
	//
	var data []rowData
	// Iterate source-level registers
	for i, reg := range p.mapping.Registers() {
		// construct source-level register id
		rid := sc.NewRegisterId(uint(i))
		// determine corresponding limbs
		limbs := p.mapping.LimbIds(rid)
		//
		if columnIncluded(p.filter, limbs) {
			//
			data = append(data, rowData{
				limbs: limbs,
				// Determine column name
				name: reg.Name,
				// Render column data from all limbs
				data: p.renderColumnData(first, last, rid, limbs),
			})
		}
	}
	//
	return data
}

// Render columns as register limbs by taking their values from the trace directly.
func (p *moduleView[F]) renderColumnsFromLimbs(first, last uint) []rowData {
	//
	var data []rowData
	// Iterate source-level registers
	for i := range p.mapping.Registers() {
		// construct source-level register id
		rid := sc.NewRegisterId(uint(i))
		// iterate register limbs
		for _, lid := range p.mapping.LimbIds(rid) {
			//
			limbs := []sc.LimbId{lid}
			//
			if columnIncluded(p.filter, limbs) {
				//
				data = append(data, rowData{
					limbs: limbs,
					// Determine column name
					name: p.mapping.Limb(lid).Name,
					// Render column data only from this limb
					data: p.renderColumnData(first, last, rid, limbs),
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
func (p *moduleView[F]) renderColumnData(first, last uint, rid sc.RegisterId, limbs []sc.LimbId) []string {
	//
	var data = make([]string, last-first)
	//
	for i := first; i < last; i++ {
		data[i-first] = p.renderColumnDataRow(i, rid, limbs)
	}
	//
	return data
}

// Render an individual row in a given source-level column.
func (p *moduleView[F]) renderColumnDataRow(column uint, rid sc.RegisterId, limbs []sc.LimbId) string {
	var (
		bits  = uint(0)
		value big.Int
	)
	//
	for _, lid := range limbs {
		var (
			data    = p.trace.Column(lid.Unwrap()).Data()
			element = data.Get(column)
			limb    = p.mapping.Limb(lid)
			val     big.Int
		)
		// Construct value from field element
		val.SetBytes(element.Bytes())
		// Shift and add
		value.Add(&value, val.Mul(&val, math.Pow2(bits)))
		//
		bits += limb.Width
	}
	//
	text := renderCellValue(p.display[rid.Unwrap()], value, nil)
	//
	return clipValue(text, p.cellWidth)
}

func (p *moduleView[F]) boundModuleRows() (first, last uint) {
	var (
		width, height = p.trace.Width(), p.trace.Height()
		m             = minWindowRow(width, height, p.filter) - int(p.padding)
		n             = maxWindowRow(width, height, p.filter) + 1 + int(p.padding)
	)
	//
	return uint(max(m, 0)), min(uint(n), height)
}

func buildModuleColumns(start, end uint) []string {
	var rows = make([]string, 1+end-start)

	rows[0] = ""
	for row := start; row < end; row++ {
		rows[1+row-start] = fmt.Sprintf("%d", row)
	}

	return rows
}

func buildModuleColumnWidths(ncolumns uint, rows []rowData) []uint {
	var colWidths = make([]uint, ncolumns)
	//
	for i := range ncolumns {
		for _, col := range rows {
			var ith uint
			//
			if i == 0 {
				ith = uint(len(col.name))
			} else {
				ith = uint(len(col.data[i-1]))
			}
			//
			colWidths[i] = max(colWidths[i], ith)
		}
	}

	return colWidths
}

func buildWindowHighlights(start, end uint, cols []rowData, filter ColumnFilter) []bool {
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

// Determine the (unclipped) string value at a given column and row in a given
// trace.
func renderCellValue(disp uint, val big.Int, enums []corset.Enumeration) string {
	//
	switch {
	case disp == corset.DISPLAY_HEX:
		// default
	case disp == corset.DISPLAY_DEC:
		return val.Text(10)
	case disp == corset.DISPLAY_BYTES:
		return displayBytes(val)
	case disp >= corset.DISPLAY_CUSTOM:
		enumID := int(disp - corset.DISPLAY_CUSTOM)
		// Check whether valid enumeration.
		if enumID < len(enums) {
			var index big.Int
			//
			index.SetBytes(val.Bytes())
			//
			if index.IsUint64() {
				// Check whether value covered by enumeration.
				if lab, ok := enums[enumID][index.Uint64()]; ok {
					return lab
				}
			}
		}
	}
	// Default:
	return fmt.Sprintf("0x%s", val.Text(16))
}

// Format a field element according to the ":bytes" directive.
func displayBytes(val big.Int) string {
	var (
		builder strings.Builder
	)
	// Handle zero case specifically.
	if val.BitLen() == 0 {
		return "00"
	}
	//
	for i, b := range val.Bytes() {
		if i != 0 {
			builder.WriteString(" ")
		}
		//
		builder.WriteString(fmt.Sprintf("%02x", b))
	}
	//
	return builder.String()
}

func clipValue(str string, maxWidth uint) string {
	runes := []rune(str)
	//
	if maxWidth > 2 && len(runes) > int(maxWidth) {
		runes := runes[:maxWidth]
		runes[maxWidth-1] = '.'
		runes[maxWidth-2] = '.'
		// done
		return string(runes)
	} else if len(runes) > int(maxWidth) {
		return string(runes[:maxWidth])
	}
	// No clipping required
	return str
}
