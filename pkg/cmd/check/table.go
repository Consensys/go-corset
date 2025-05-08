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

	tr "github.com/consensys/go-corset/pkg/trace"
	"github.com/consensys/go-corset/pkg/util/collection/bit"
	"github.com/consensys/go-corset/pkg/util/collection/set"
)

// CellRefSet defines a type for sets of cell references.
type CellRefSet = *set.AnySortedSet[tr.CellRef]

// TraceWindow abstracts an underlying trace by accounting for perspectives at
// the corset-level (amongst other things).
type TraceWindow interface {
	// CellAt returns the contents of a specific cell in this table.
	CellAt(col uint, row uint) string
	// Column returns the title of the given column.
	Column(uint) string
	// Height returns the number of rows in this table.
	Height() uint
	// Highlighted determines whether a given cell should be highlighted or not.
	Highlighted(col uint, row uint) bool
	// Rowe returns the title of the given row
	Row(uint) string
	// Width returns the number of columns in this table.
	Width() uint
}

// NewTraceWindow constructs a window into a trace which includes all of the
// given cells, plus some amount of padding on either side (i.e. additional rows
// before and after to help with context).
func NewTraceWindow(cells CellRefSet, trace tr.Trace, padding uint) TraceWindow {
	var registers bit.Set
	// Determine underlying registers to be shown in this window.  Observe that
	// registers do not directly correspond with columns at the corset level, as
	// one register can represent multiple corset columns (e.g. in different
	// perspectives).
	for _, c := range cells.ToArray() {
		registers.Insert(c.Column)
	}
	// Determine row bounds
	start, end := determineWindowBounds(cells, trace, padding)
	//
	return &traceWindow{
		rows:       determineWindowRows(start, end),
		columns:    determineWindowColumns(registers, trace),
		data:       determineWindowData(start, end, registers, trace),
		highlights: determineWindowHighlights(start, end, cells, registers, trace),
	}
}

// Determine which rows to include in the given window.
func determineWindowBounds(cells CellRefSet, trace tr.Trace, padding uint) (uint, uint) {
	var (
		start  int = math.MaxInt
		end    int = 0
		height     = int(tr.MaxHeight(trace))
	)
	// Determine all (input) cells involved in evaluating the given constraint
	for _, c := range cells.ToArray() {
		start = min(start, c.Row)
		end = max(end, c.Row+1)
	}
	// apply padding
	start = max(start-int(padding), 0)
	end = min(end+int(padding), height)
	//
	return uint(start), uint(end)
}

func determineWindowColumns(regs bit.Set, trace tr.Trace) []string {
	var columns []string
	//
	for iter := regs.Iter(); iter.HasNext(); {
		ith := iter.Next()
		name := trace.Column(ith).Name()
		columns = append(columns, name)
	}
	//
	return columns
}

func determineWindowRows(start, end uint) []string {
	var rows = make([]string, end-start)

	for row := start; row < end; row++ {
		rows[row] = fmt.Sprintf("%d", row)
	}

	return rows
}

func determineWindowData(start, end uint, regs bit.Set, trace tr.Trace) [][]string {
	var data = make([][]string, end-start)
	//
	for r := start; r < end; r++ {
		var row []string
		//
		for iter := regs.Iter(); iter.HasNext(); {
			register := iter.Next()
			// extract value at given row in this register
			val := trace.Column(register).Data().Get(r)
			// convert value into string
			row = append(row, val.Text(16))
		}
		//
		data[r] = row
	}
	//
	return data
}

func determineWindowHighlights(start, end uint, cells CellRefSet, regs bit.Set, trace tr.Trace) [][]bool {
	var (
		highlights = make([][]bool, end-start)
		width      = regs.Count()
		mapping    = make([]uint, trace.Width())
	)
	// Initialise register mapping
	for iter, i := regs.Iter(), uint(0); iter.HasNext(); i++ {
		register := iter.Next()
		mapping[register] = i
	}
	// Initialise all highlights disabled
	for i := range highlights {
		highlights[i] = make([]bool, width)
	}
	// Enable highlights for affected cells
	for iter := cells.Iter(); iter.HasNext(); {
		cell := iter.Next()
		col := mapping[cell.Column]
		row := uint(cell.Row) - start
		highlights[row][col] = true
	}
	//
	return highlights
}

// ============================================================================
// Window Implementation
// ============================================================================

type traceWindow struct {
	// Set of rows in this window
	rows []string
	// Set of columns in this window
	columns []string
	// Data contained in this window
	data [][]string
	// Highlighted cells in this window
	highlights [][]bool
}

// CellAt returns the contents of a specific cell in this table.
func (p *traceWindow) CellAt(col uint, row uint) string {
	return p.data[row][col]
}

// Column returns the title of the given column.
func (p *traceWindow) Column(col uint) string {
	return p.columns[col]
}

// Height returns the number of rows in this table.
func (p *traceWindow) Height() uint {
	return uint(len(p.rows))
}

// Highlighted determines whether a given cell should be highlighted or not.
func (p *traceWindow) Highlighted(col uint, row uint) bool {
	return p.highlights[row][col]
}

// Rowe returns the title of the given row
func (p *traceWindow) Row(row uint) string {
	return p.rows[row]
}

// Width returns the number of columns in this table.
func (p *traceWindow) Width() uint {
	return uint(len(p.columns))
}
