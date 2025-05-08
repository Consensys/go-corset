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
package check

import (
	"fmt"
	"math"

	"github.com/consensys/go-corset/pkg/util/termio"
)

// Printer encapsulates various configuration options useful for printing out
// traces in human-readable forms.
type Printer struct {
	// Determine maximum width to print
	maxCellWidth uint
	// Enable ANSI
	ansiEscapes bool
}

// NewPrinter constructs a default printer
func NewPrinter() *Printer {
	// Return an empty printer
	return &Printer{math.MaxUint, true}
}

// AnsiEscapes can be used to enable or disable the use of ANSI escape sequences
// (e.g. for showing colour in a terminal, etc)
func (p *Printer) AnsiEscapes(enable bool) *Printer {
	p.ansiEscapes = enable
	return p
}

// MaxCellWidth sets the maximum width to use for the cell data.
func (p *Printer) MaxCellWidth(width uint) *Printer {
	p.maxCellWidth = width
	return p
}

// Print a given trace using the configured printer
func (p *Printer) Print(trace TraceWindow) {
	var height = trace.Height()
	// Construct table
	tp := termio.NewTablePrinter(1+height, 1+trace.Width())
	// Initialise row titles
	for j := uint(0); j < height; j++ {
		title := trace.Row(j)
		escape := termio.NewAnsiEscape().FgColour(termio.TERM_WHITE)
		text := termio.NewFormattedText(title, escape)
		tp.Set(1+j, 0, text)
	}
	// Construct suitable highlighting escape
	highlightEscape := termio.BoldAnsiEscape().FgColour(termio.TERM_RED)
	// Fill table
	for col := uint(0); col != trace.Width(); col++ {
		title := trace.Column(col)
		// Set columns names
		escape := termio.NewAnsiEscape().FgColour(termio.TERM_WHITE)
		text := termio.NewFormattedText(title, escape)
		tp.Set(0, col+1, text)
		//
		for row := uint(0); row < height; row++ {
			var text termio.FormattedText
			// Extract contents of cell
			contents := trace.CellAt(col, row)
			// Determine text of cell
			highlight := trace.Highlighted(col, row)
			//
			if highlight && !p.ansiEscapes {
				// In a non-ANSI environment, use a marker "*" to identify which cells were depended upon.
				text = termio.NewText(fmt.Sprintf("*0x%s", contents))
			} else if highlight {
				hex := fmt.Sprintf("0x%s", contents)
				text = termio.NewFormattedText(hex, highlightEscape)
			} else {
				text = termio.NewText(fmt.Sprintf("0x%s", contents))
			}
			//
			tp.Set(1+row, col+1, text)
		}
	}
	// Cap cell widths
	for j := uint(0); j < height; j++ {
		tp.SetMaxWidth(1+j, p.maxCellWidth)
	}
	// Done
	tp.Print(p.ansiEscapes)
}
