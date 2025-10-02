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
	// Trace provides the raw data for this view
	trace trace.Module[F]
	// Srcmap provides relevant display information.
	srcmap corset.SourceModule
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
		p.data = buildWindowData(p.trace, p.srcmap)
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
	name string
	data []string
}

// ============================================================================
// Helpers
// ============================================================================

func buildWindowData[F field.Element[F]](trace tr.Module[F], srcmap corset.SourceModule) *moduleData {
	var (
		rows = buildWindowRows(0, trace.Height())
		cols = buildWindowColumns(trace, srcmap)
	)

	return &moduleData{rows, cols, make([]bool, len(rows)*len(cols))}
}

func buildWindowColumns[F field.Element[F]](trace tr.Module[F], srcmap corset.SourceModule) []columnData {
	var data = make([]columnData, len(srcmap.Columns))
	//
	for i, c := range srcmap.Columns {
		// Dig out the column id
		cid := c.Register.Column().Unwrap()
		//
		data[i] = columnData{
			name: c.Name,
			data: buildWindowColumnData(trace.Column(cid)),
		}
	}
	//
	return data
}

func buildWindowColumnData[F field.Element[F]](column tr.Column[F]) []string {
	var data = make([]string, column.Data().Len())
	//
	for i := range column.Data().Len() {
		// FIXME: this is a very limited conversion at this time.
		ith := column.Data().Get(i)
		data[i] = ith.Text(16)
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
