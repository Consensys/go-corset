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
package inspector

import (
	"fmt"
	"math/big"
	"strings"

	"github.com/consensys/gnark-crypto/ecc/bls12-377/fr"
	"github.com/consensys/go-corset/pkg/corset"
	tr "github.com/consensys/go-corset/pkg/trace"
	"github.com/consensys/go-corset/pkg/util/field/bls12_377"
	"github.com/consensys/go-corset/pkg/util/termio"
)

// ModuleView is responsible for generating the current window into a trace to
// be shown in the inspector.
type ModuleView struct {
	// Offsets into the module data.
	row, col uint
	// Maximum number of rows in any column
	height uint
	// Columns currently being shown in this view.  For example, only columns
	// matching the currently active filter would be in this array.
	columns []SourceColumn
	// RowWidths holds the maximum width of all rows in the module's trace.
	rowWidths []uint
	// Specifies the maximum width for any row
	maxRowWidth uint
	// ColTitleWidth holds the maximum width of any (active) column title in the
	// module's trace.
	colTitleWidth uint
	// Available enumerations
	enumerations []corset.Enumeration
}

// SetColumn sets the column offset if its valid (otherwise ignore).  This
// affects which columns are visible in the view.
func (p *ModuleView) SetColumn(col uint) {
	if col < uint(len(p.columns)) {
		p.col = col
	}
}

// SetRow the row offset if its valid (otherwise ignore).  This affects which
// rows are visible in the view.  Note: column titles are always visible though.
func (p *ModuleView) SetRow(row uint) uint {
	p.row = min(row, uint(len(p.rowWidths)-1))
	//
	return p.row
}

// SetMaxRowWidth sets the maximum display width of any row in the trace.  Cells
// whose contents are wider than this will be clipped accordingly.
func (p *ModuleView) SetMaxRowWidth(width uint, trace tr.Trace[bls12_377.Element]) {
	p.maxRowWidth = width
	// Action change
	p.rowWidths = p.recalculateRowWidths(trace)
}

// SetActiveColumns sets the currently active set of columns.  This updates the
// current column title width, as as well as the maximum width for every row.
func (p *ModuleView) SetActiveColumns(trace tr.Trace[bls12_377.Element], columns []SourceColumn) {
	p.columns = columns
	p.height = 0
	// Recalculate module height
	for _, c := range p.columns {
		p.height = max(p.height, trace.Column(c.Register).Data().Len())
	}
	// Recalculate maximum title width
	p.colTitleWidth = p.recalculateColumnTitleWidth()
	// Recalculate row widths
	p.rowWidths = p.recalculateRowWidths(trace)
}

// RowWidth returns the width of the largest element in a given row.  Observe
// that the first row is always reserved for the column titles.
func (p *ModuleView) RowWidth(row uint) uint {
	if row == 0 {
		return p.colTitleWidth
	}
	// Calculate actual row
	row = row - 1 + p.row
	// Sanity check is valid
	if row < uint(len(p.rowWidths)) {
		return p.rowWidths[row]
	}
	// Invalid row, so show nothing.
	return 0
}

// CellAt returns a textual representation of the data at a given column and row
// in the module's view.  Observe that the first row and column typically show
// titles.
func (p *ModuleView) CellAt(trace tr.Trace[bls12_377.Element], col, row uint) termio.FormattedText {
	if row == 0 && col == 0 {
		return termio.NewText("")
	}
	// Determine trace row / column indices
	trCol := col - 1 + p.col
	trRow := row - 1 + p.row
	//
	if row == 0 && trCol < uint(len(p.columns)) {
		// Column title
		name := p.columns[trCol].Name
		if p.columns[trCol].Computed {
			return termio.NewColouredText(name, termio.TERM_GREEN)
		}
		//
		return termio.NewColouredText(name, termio.TERM_BLUE)
	} else if col == 0 {
		// Row title
		val := fmt.Sprintf("%d", trRow)
		return termio.NewColouredText(val, termio.TERM_BLUE)
	} else if trRow >= uint(len(p.rowWidths)) || trCol >= uint(len(p.columns)) {
		// non-existent rows
		return termio.NewText("")
	}
	// Determine value at given trace column / row
	val := p.ValueAt(trace, trCol, trRow)
	// Generate textual representation of value, and clip accordingly.
	str := clipValue(p.display(trCol, val), p.rowWidths[trRow])
	//
	if p.IsActive(trace, trCol, trRow) {
		// Calculate appropriate colour for this cell.
		return termio.NewFormattedText(str, cellColour(val))
	} else {
		return termio.NewColouredText(str, termio.TERM_BLACK)
	}
}

// ValueAt extracts the data point at a given rol and column in the trace.
func (p *ModuleView) ValueAt(trace tr.Trace[bls12_377.Element], trCol, trRow uint) fr.Element {
	// Determine underlying register for the given column.
	ref := p.columns[trCol].Register
	// Extract cell value from register
	return trace.Column(ref).Get(int(trRow)).Element
}

// IsActive determines whether a given cell is active, or not.  A cell can be
// inactive, for example, if its part of a perspective which is not active.
func (p *ModuleView) IsActive(trace tr.Trace[bls12_377.Element], trCol, trRow uint) bool {
	// Determine enclosing module
	module := trace.Module(p.columns[trCol].Register.Module())
	// Extract relevant selector
	selector := p.columns[trCol].Selector
	// Santity check whether actually need to do anything
	if selector.IsEmpty() {
		return true
	}
	// Check selector value
	val := module.ColumnOf(selector.Unwrap()).Get(int(trRow))
	//
	return !val.IsZero()
}

// ============================================================================
// Helpers
// ============================================================================

// This algorithm is based on that used in the original tool.  To understand
// this algorithm, you need to look at the 256 colour table for ANSI escape
// codes.  It actually does make sense, even if it doesn't appear to.
func cellColour(val fr.Element) termio.AnsiEscape {
	if val.IsZero() {
		return termio.NewAnsiEscape().FgColour(termio.TERM_WHITE)
	}
	// Compute a simple hash of the bytes making up the value in question.
	col := uint(0)
	for _, b := range val.Bytes() {
		col = col ^ uint(b)
	}
	// Select suitable background colour based on hash, whilst also ensuring
	// contrast with the foreground colour.
	bg_col := (col % (213 - 16))
	escape := termio.NewAnsiEscape().Bg256Colour(16 + bg_col)
	//
	if bg_col%36 > 18 {
		escape = escape.FgColour(termio.TERM_BLACK)
	}
	//
	return escape
}

// Determine the maximum width of any column name in the given set of columns.
func (p *ModuleView) recalculateColumnTitleWidth() uint {
	maxWidth := 0

	for _, col := range p.columns {
		maxWidth = max(maxWidth, len(col.Name))
	}

	return uint(maxWidth)
}

func (p *ModuleView) recalculateRowWidths(module tr.Trace[bls12_377.Element]) []uint {
	widths := make([]uint, p.height)
	//
	for row := uint(0); row < uint(len(widths)); row++ {
		maxWidth := uint(0)
		//
		for col := uint(0); col < uint(len(p.columns)); col++ {
			val := p.ValueAt(module, col, row)
			width := len(p.display(col, val))
			maxWidth = max(maxWidth, uint(width))
		}
		//
		widths[row] = min(p.maxRowWidth, maxWidth)
	}
	//
	return widths
}

// Determine the (unclipped) string value at a given column and row in a given
// trace.
func (p *ModuleView) display(col uint, val fr.Element) string {
	if col < uint(len(p.columns)) {
		disp := p.columns[col].Display
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
			if enumID < len(p.enumerations) {
				// Check whether value covered by enumeration.
				if lab, ok := p.enumerations[enumID][val]; ok {
					return lab
				}
			}
		}
	}
	// Default:
	return fmt.Sprintf("0x%s", val.Text(16))
}

// Format a field element according to the ":bytes" directive.
func displayBytes(val fr.Element) string {
	var (
		builder strings.Builder
		bival   big.Int
	)
	// Handle zero case specifically.
	if val.IsZero() {
		return "00"
	}
	// assign as big integer
	val.BigInt(&bival)
	//
	for i, b := range bival.Bytes() {
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
	if len(runes) > int(maxWidth) {
		runes := runes[0:maxWidth]
		runes[maxWidth-1] = '.'
		runes[maxWidth-2] = '.'
		// done
		return string(runes)
	}
	// No clipping required
	return str
}
