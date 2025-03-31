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
package trace

import (
	"fmt"
	"math"
	"unicode/utf8"

	"github.com/consensys/go-corset/pkg/util/termio"
)

// ColumnFilter is a predicate which determines whether a given column should be
// included in the print out, or not.
type ColumnFilter = func(uint, Trace) bool

// Highlighter identifies cells which should be highlighted.
type Highlighter = func(CellRef, Trace) bool

// Printer encapsulates various configuration options useful for printing out
// traces in human-readable forms.
type Printer struct {
	// First row to print
	startRow uint
	// Last row to print
	endRow uint
	// Additional rows either side
	padding uint
	// Which columns to include
	colFilter ColumnFilter
	// Which columns to highlight
	highlighter Highlighter
	// Determine maximum width to print
	maxCellWidth uint
	// Enable ANSI
	ansiEscapes bool
}

// NewPrinter constructs a default printer
func NewPrinter() *Printer {
	// Include all colunms by default
	emptyFilter := func(row uint, t Trace) bool {
		return true
	}
	// Highlight nothing by default
	emptyHighlighter := func(cell CellRef, t Trace) bool {
		return false
	}
	// Return an empty printer
	return &Printer{0, math.MaxInt, 2, emptyFilter, emptyHighlighter, math.MaxUint, true}
}

// Start configures the starting row for this printer.
func (p *Printer) Start(start uint) *Printer {
	p.startRow = start
	return p
}

// End configures the ending row (inclusive) for this printer.
func (p *Printer) End(end uint) *Printer {
	p.endRow = end
	return p
}

// Padding configures the number of padding rows (i.e. rows outside the affected
// area) to include for additional context.
func (p *Printer) Padding(padding uint) *Printer {
	p.padding = padding
	return p
}

// Columns configures a filter which selects columns to be included in the final
// print out.
func (p *Printer) Columns(filter ColumnFilter) *Printer {
	p.colFilter = filter
	return p
}

// AnsiEscapes can be used to enable or disable the use of ANSI escape sequences
// (e.g. for showing colour in a terminal, etc)
func (p *Printer) AnsiEscapes(enable bool) *Printer {
	p.ansiEscapes = enable
	return p
}

// Highlight configures a filter for cells which should be highlighted.  By
// default, no cells are highlighted.
func (p *Printer) Highlight(highlighter Highlighter) *Printer {
	p.highlighter = highlighter
	return p
}

// MaxCellWidth sets the maximum width to use for the cell data.
func (p *Printer) MaxCellWidth(width uint) *Printer {
	p.maxCellWidth = width
	return p
}

// Print a given trace using the configured printer
func (p *Printer) Print(trace Trace) {
	var start uint
	if p.startRow >= p.padding {
		start = p.startRow - p.padding
	} else if p.padding > 0 {
		start = 0
	} else {
		start = p.startRow
	}
	//
	end := min(MaxHeight(trace), p.endRow+p.padding+1)
	//
	columns := make([]uint, 0)
	width := 1 + end - start
	// Filter columns
	for i := uint(0); i < trace.Width(); i++ {
		if p.colFilter(i, trace) {
			columns = append(columns, i)
		}
	}
	// Construct table
	tp := termio.NewTablePrinter(width, uint(1+len(columns)))
	// Initialise row indices
	for j := start; j < end; j++ {
		escape := termio.NewAnsiEscape().FgColour(termio.TERM_WHITE)
		text := termio.NewFormattedText(fmt.Sprintf("%d", j), escape)
		tp.Set(1+j-start, 0, text)
	}
	// Construct suitable highlighting escape
	highlightEscape := termio.BoldAnsiEscape().FgColour(termio.TERM_RED)
	// Fill table
	for i, col := range columns {
		column := trace.Column(col)
		maxRow := min(end, column.Data().Len())
		// Set columns names
		escape := termio.NewAnsiEscape().FgColour(termio.TERM_WHITE)
		text := termio.NewFormattedText(column.Name(), escape)
		tp.Set(0, uint(i+1), text)
		//
		for row := start; row < maxRow; row++ {
			var hex string
			// Extract data for cell
			jth := column.Data().Get(row)
			// Determine text of cell
			highlight := p.highlighter(NewCellRef(col, int(row)), trace)
			//
			if highlight && !p.ansiEscapes {
				// In a non-ANSI environment, use a marker "*" to identify which cells were depended upon.
				hex = fmt.Sprintf("*0x%s", jth.Text(16))
			} else {
				hex = fmt.Sprintf("0x%s", jth.Text(16))
			}
			//
			tp.Set(1+row-start, uint(i+1), termio.NewFormattedText(hex, highlightEscape))
		}
	}
	// Cap cells
	for j := start; j < end; j++ {
		tp.SetMaxWidth(1+j-start, p.maxCellWidth)
	}
	// Done
	tp.Print(p.ansiEscapes)
}

// PrintTrace prints a trace in a more human-friendly fashion.
func PrintTrace(tr Trace) {
	n := tr.Width()
	m := MaxHeight(tr)
	//
	rows := make([][]string, n)
	for i := uint(0); i < n; i++ {
		rows[i] = traceColumnData(tr, i)
	}
	//
	widths := traceRowWidths(m, rows)
	//
	printHorizontalRule(widths)
	//
	for _, r := range rows {
		printTraceRow(r, widths)
		printHorizontalRule(widths)
	}
}

func traceColumnData(tr Trace, col uint) []string {
	n := MaxHeight(tr)
	data := make([]string, n+2)
	data[0] = fmt.Sprintf("#%d", col)
	data[1] = tr.Column(col).Name()

	for row := 0; row < int(n); row++ {
		ith := tr.Column(col).Get(row)
		data[row+2] = ith.String()
	}

	return data
}

func traceRowWidths(height uint, rows [][]string) []int {
	widths := make([]int, height+2)

	for _, row := range rows {
		for i, col := range row {
			w := utf8.RuneCountInString(col)
			widths[i] = max(w, widths[i])
		}
	}

	return widths
}

func printTraceRow(row []string, widths []int) {
	for i, col := range row {
		fmt.Printf(" %*s |", widths[i], col)
	}

	fmt.Println()
}

func printHorizontalRule(widths []int) {
	for _, w := range widths {
		fmt.Print("-")

		for i := 0; i < w; i++ {
			fmt.Print("-")
		}

		fmt.Print("-+")
	}

	fmt.Println()
}
