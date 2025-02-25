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
	"github.com/consensys/go-corset/pkg/util/termio"
)

// ModuleView is responsible for generating the current window into a trace to
// be shown in the inspector.
type ModuleView struct {
	// Offsets into the module data.
	row, col uint
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
func (p *ModuleView) SetRow(row uint) bool {
	if row < uint(len(p.rowWidths)) {
		p.row = row
		return true
	}
	//
	return false
}

// SetActiveColumns sets the currently active set of columns.  This updates the
// current column title width, as as well as the maximum width for every row.
func (p *ModuleView) SetActiveColumns(trace tr.Trace, columns []SourceColumn) {
	p.columns = columns
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
func (p *ModuleView) CellAt(trace tr.Trace, col, row uint) termio.FormattedText {
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
func (p *ModuleView) ValueAt(trace tr.Trace, trCol, trRow uint) fr.Element {
	// Determine underlying register for the given column.
	regId := p.columns[trCol].Register
	// Extract cell value from register
	return trace.Column(regId).Get(int(trRow))
}

// IsActive determines whether a given cell is active, or not.  A cell can be
// inactive, for example, if its part of a perspective which is not active.
func (p *ModuleView) IsActive(trace tr.Trace, trCol, trRow uint) bool {
	selector := p.columns[trCol].Selector
	//
	if selector == nil {
		return true
	}
	//
	val, err := selector.EvalAt(int(trRow), trace)
	// error check
	if err != nil {
		panic(err.Error())
	}
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

func (p *ModuleView) recalculateRowWidths(trace tr.Trace) []uint {
	// Determine how many rows we have
	nrows := determineNumberOfRows(trace, p.columns)
	//
	widths := make([]uint, nrows)
	//
	for row := uint(0); row < uint(len(widths)); row++ {
		maxWidth := uint(0)
		//
		for col := uint(0); col < uint(len(p.columns)); col++ {
			val := p.ValueAt(trace, col, row)
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

// Determine the maximum number of rows whih can be displayed for a given set of
// columns.  Observe that this is not fully determined by the module height,
// since we have columns which may have length multipliers, etc.
func determineNumberOfRows(trace tr.Trace, columns []SourceColumn) uint {
	maxRows := uint(0)

	for _, col := range columns {
		nrows := trace.Column(col.Register).Data().Len()
		maxRows = max(maxRows, nrows)
	}
	//
	return maxRows
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
