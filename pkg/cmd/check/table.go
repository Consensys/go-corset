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

	"github.com/consensys/go-corset/pkg/cmd/inspector"
	"github.com/consensys/go-corset/pkg/corset"
	tr "github.com/consensys/go-corset/pkg/trace"
	"github.com/consensys/go-corset/pkg/util"
	"github.com/consensys/go-corset/pkg/util/collection/bit"
	"github.com/consensys/go-corset/pkg/util/collection/set"
	"github.com/consensys/go-corset/pkg/util/field/bls12_377"
)

// CellRefSet defines a type for sets of cell references.
type CellRefSet = *set.AnySortedSet[tr.CellRef]

// SourceColumn provides information about a source-level column and its mapping
// to the underlying registers of a trace.
type SourceColumn = inspector.SourceColumn

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
func NewTraceWindow(
	cells CellRefSet,
	trace tr.Trace[bls12_377.Element],
	padding uint,
	srcmap *corset.SourceMap,
) TraceWindow {
	// Determine corset-level columns to show in this window.  Observe that
	// registers do not directly correspond with columns at the corset level, as
	// one register can represent multiple corset columns (e.g. in different
	// perspectives).
	columns := determineSourceColumns(cells, trace, srcmap)
	// Determine row bounds
	start, end := determineWindowBounds(cells, trace, padding)
	//
	return &traceWindow{
		rows:       determineWindowRows(start, end),
		columns:    determineWindowColumns(columns),
		data:       determineWindowData(start, end, columns, trace),
		highlights: determineWindowHighlights(start, end, cells, columns),
	}
}

// Determine complete set of source columns.
func determineSourceColumns(
	cells CellRefSet,
	trace tr.Trace[bls12_377.Element],
	srcmap *corset.SourceMap,
) []SourceColumn {
	var (
		ncolumns []SourceColumn
		seen     bit.Set
	)
	//
	if srcmap == nil {
		// Fall back when corset source mapping unavailable.
		return determineSourceColumnsFromTrace(cells, trace)
	}
	//
	mod := determineEnclosingModule(cells, trace, srcmap.Root)
	// Reuse existing functionality from inspector to determine set of all modules.
	columns := inspector.ExtractSourceColumns(util.NewAbsolutePath(""), mod.Selector, mod.Columns, mod.Submodules)
	//
	for _, c := range cells.ToArray() {
		index := c.Column.Index(trace.Width())
		if !seen.Contains(index) {
			column := determineSourceColumn(c, trace, columns)
			ncolumns = append(ncolumns, column)
			// Don't include column more than once.
			seen.Insert(index)
		}
	}
	//
	return ncolumns
}

// Determine complete set of source columns using only a trace as the source of
// truth.  This means, for example, that perspectives are not properly accounted
// for.  Likewise, any display information given on the original column
// definition is ignored.
func determineSourceColumnsFromTrace(cells CellRefSet, trace tr.Trace[bls12_377.Element]) []SourceColumn {
	var (
		columns   []SourceColumn
		registers bit.Set
	)
	// Compute relevant registers
	for iter := cells.Iter(); iter.HasNext(); {
		// Determine unique column index
		index := iter.Next().Column.Index(trace.Width())
		// Store it as seen.
		registers.Insert(index)
	}
	// Include all relevant registers, using defaults as necessary to fill the
	// missing gaps.
	for iter := registers.Iter(); iter.HasNext(); {
		cell := iter.Next()
		// Reconstruct the column reference
		cref := tr.NewIndexedColumnRef(cell, trace.Width())
		// Extract column in question
		column := trace.Column(cref)
		// Append it
		columns = append(columns, SourceColumn{
			Name:     column.Name(),
			Computed: false,
			Selector: util.None[string](),
			Display:  corset.DISPLAY_HEX,
			Register: cref,
		})
	}
	//
	return columns
}

// Determine a unique enclosing module for this report.  Note that this may not
// always exists.  For example, lookups involve two modules and, hence, there is
// no unique module we can refer to.  The purpose of identifying the enclosing
// module is simply to improve the column names reported (i.e. in some cases we
// can ommit the module itself from the name as this is just repeated).
func determineEnclosingModule(
	cells CellRefSet,
	trace tr.Trace[bls12_377.Element],
	root corset.SourceModule,
) corset.SourceModule {
	var name string
	// Attempt to reconstruct the enclosing modules name.
	for i, cr := range cells.ToArray() {
		// Determine enclosing column
		n := trace.Module(cr.Column.Module()).Name()
		// Convert module name back to a context
		context := corset.ContextOf(n)
		//
		if i == 0 {
			name = context.ModuleId
		} else if name != context.ModuleId {
			panic(fmt.Sprintf("conflicting submodules (%s vs %s)", name, context.ModuleId))
		}
	}
	//
	if name == "" {
		return root
	} else if mod := root.Submodule(name); mod != nil {
		return *mod
	}
	// Should be unreachable
	panic(fmt.Sprintf("unknown submodule %s", name))
}

// Find the unique source column to which a given cell references.
func determineSourceColumn(cell tr.CellRef, trace tr.Trace[bls12_377.Element], columns []SourceColumn) SourceColumn {
	for _, col := range columns {
		if col.Register == cell.Column && isActiveColumn(cell.Row, col, trace) {
			return col
		}
	}
	// In theory, this should be unreachable.  In practice, it can be triggered.
	// Therefore we use a suitable default instead.
	column := trace.Column(cell.Column)
	//
	return SourceColumn{
		Name:     column.Name(),
		Computed: false,
		Selector: util.None[string](),
		Display:  corset.DISPLAY_HEX,
		Register: cell.Column,
	}
}

// Determine whether a given source column is active on a given row of the
// trace.  A column which has no selector is always active.  Otherwise, the
// column is considered active if the given selector evaluates to a non-zero
// value on the given row.
func isActiveColumn(row int, col SourceColumn, trace tr.Trace[bls12_377.Element]) bool {
	if col.Selector.IsEmpty() {
		return true
	}
	// Lookup enclosing module
	module := trace.Module(col.Register.Module())
	// Check selector value
	val := module.ColumnOf(col.Selector.Unwrap()).Get(row)
	//
	return !val.IsZero()
}

// Determine which rows to include in the given window.
func determineWindowBounds(cells CellRefSet, trace tr.Trace[bls12_377.Element], padding uint) (uint, uint) {
	var (
		start  int  = math.MaxInt
		end    int  = 0
		height uint = 0
	)
	// Determine all (input) cells involved in evaluating the given constraint
	for _, c := range cells.ToArray() {
		start = min(start, c.Row)
		end = max(end, c.Row+1)
		height = max(height, trace.Column(c.Column).Data().Len())
	}
	// apply padding
	start = max(start-int(padding), 0)
	end = min(end+int(padding), int(height))
	//
	return uint(start), uint(end)
}

func determineWindowColumns(columns []SourceColumn) []string {
	columnTitles := make([]string, len(columns))
	//
	for i, col := range columns {
		columnTitles[i] = col.Name
	}
	//
	return columnTitles
}

func determineWindowRows(start, end uint) []string {
	var rows = make([]string, end-start)

	for row := start; row < end; row++ {
		rows[row-start] = fmt.Sprintf("%d", row)
	}

	return rows
}

func determineWindowData(start, end uint, columns []SourceColumn, trace tr.Trace[bls12_377.Element]) [][]string {
	var data = make([][]string, end-start)
	//
	for r := start; r < end; r++ {
		var row []string
		//
		for _, col := range columns {
			// extract value at given row in this register
			val := trace.Column(col.Register).Data().Get(r)
			// convert value into string
			row = append(row, val.Text(16))
		}
		//
		data[r-start] = row
	}
	//
	return data
}

func determineWindowHighlights(start, end uint, cells CellRefSet, columns []SourceColumn) [][]bool {
	var (
		highlights = make([][]bool, end-start)
		mapping    = make(map[tr.ColumnRef]uint, 0)
	)
	// Initialise register mapping
	for i, reg := range columns {
		mapping[reg.Register] = uint(i)
	}
	// Initialise all highlights disabled
	for i := range highlights {
		highlights[i] = make([]bool, len(columns))
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
