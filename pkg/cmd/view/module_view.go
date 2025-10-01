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
	sc "github.com/consensys/go-corset/pkg/schema"
	"github.com/consensys/go-corset/pkg/util"
	"github.com/consensys/go-corset/pkg/util/field"
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
	// Data abstracts the raw data of the underlying module.
	Data() ModuleData
	// Filter this view to produce a more focused view.
	Filter(ColumnFilter) ModuleView
	// Return offset position within module
	Offset() (col uint, row uint)
	// Set offset position within module
	Goto(col, row uint)
}

// ============================================================================
// View Implementation
// ============================================================================

type moduleView[F field.Element[F]] struct {
	// Offset position within module
	x, y uint
	// width of all cells / titles
	cellWidth, titleWidth uint
	// Padding to use
	padding uint
	// Limbs indicates whether or not to show the raw limbs, or the combined
	// source-level register.
	limbs bool
	// Formatting for this module
	formatting ModuleFormatting
	// Active set of rows
	active []sc.RegisterId
	// Data provides the raw underlying data which can be shared between
	// multiple views.
	data *moduleData[F]
}

// Offset returns the current offset position within module
func (p *moduleView[F]) Config() ModuleConfig {
	return p
}

// Data returns an abtract view of the data for given register
func (p *moduleView[F]) Data() ModuleData {
	return p.data
}

// Filter columns in this module
func (p *moduleView[F]) Filter(filter ColumnFilter) ModuleView {
	var (
		mapping = p.data.Mapping()
		q       = p
	)
	// Reset filter
	q.active = nil
	//
	for i := range uint(len(mapping.Registers())) {
		rid := sc.NewRegisterId(i)
		// If any limb is included, the whole limb is included.
		if columnIncluded(filter, mapping.LimbIds(rid)) {
			q.active = append(q.active, rid)
		}
	}
	//
	return q
}

// Offset returns the current offset position within module
func (p *moduleView[F]) Offset() (col uint, row uint) {
	return p.x, p.y
}

// Goto a specific offset within module
func (p *moduleView[F]) Goto(col, row uint) {
	width, height := p.data.Dimensions()

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
}

// ============================================================================
// TableSource
// ============================================================================

// CellAt returns the contents of a specific cell in this table.
func (p *moduleView[F]) CellAt(col uint, row uint) termio.FormattedText {
	var (
		text       string
		formatting util.Option[termio.AnsiEscape]
	)
	//
	if col == 0 && row == 0 {
		return termio.NewText("")
	} else if col == 0 {
		reg := p.active[row+p.y-1]
		text = p.data.RowTitle(reg)
		formatting = p.formatting.RowTitle(reg)
	} else if row == 0 {
		text = p.data.ColumnTitle(col + p.x - 1)
		formatting = p.formatting.ColumnTitle(col + p.x - 1)
	} else {
		reg := p.active[row+p.y-1]
		row = col + p.x - 1
		//
		text = p.data.CellAt(row, reg.Unwrap())
		formatting = p.formatting.Cell(reg, row)
		// Clip the cell value
		text = clipValue(text, p.cellWidth)
	}
	// apply formatting (if applicable)
	if formatting.HasValue() {
		return termio.NewFormattedText(text, formatting.Unwrap())
	}
	// no formatting
	return termio.NewText(text)
}

func (p *moduleView[F]) ColumnWidth(col uint) uint {
	var w, _ = p.data.Dimensions()
	//
	if col == 0 {
		width := uint(0)
		//
		for _, row := range p.active {
			text := p.data.RowTitle(row)
			width = max(width, uint(len(text)))
		}
		//
		return min(p.titleWidth, width)
	} else if col+p.x-1 >= w {
		return 0
	}
	//
	col = col + p.x - 1
	//
	width := uint(len(p.data.ColumnTitle(col)))
	//
	for _, row := range p.active {
		text := p.data.CellAt(col, row.Unwrap())
		width = max(width, uint(len(text)))
	}
	//
	return min(p.cellWidth, width)
}

func (p *moduleView[F]) Dimensions() (uint, uint) {
	var (
		width, _ = p.data.Dimensions()
		height   = uint(len(p.active))
	)
	// Account for title rows
	width, height = width+1, height+1
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

// Check whether the given filter includes any if the limbs (or not).
func columnIncluded(filter ColumnFilter, limbs []sc.RegisterId) bool {
	for _, lid := range limbs {
		if filter.Column(lid) != nil {
			return true
		}
	}
	//
	return false
}

// func minWindowRow(width, height uint, filter ColumnFilter) int {
// 	// Find column with least cell in the filter
// 	for i := range height {
// 		for j := range width {
// 			if f := filter.Column(j); f != nil && f.Cell(i) {
// 				return int(i)
// 			}
// 		}
// 	}
// 	// Suggests an empty filter
// 	return 0
// }

// func maxWindowRow(width, height uint, filter ColumnFilter) int {
// 	// Find column with greatest cell in the filter
// 	for i := height; i > 0; {
// 		i = i - 1
// 		//
// 		for j := range width {
// 			if f := filter.Column(j); f != nil && f.Cell(i) {
// 				return int(i)
// 			}
// 		}
// 	}
// 	// Suggests an empty filter
// 	return 0
// }

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
