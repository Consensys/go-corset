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
	tr "github.com/consensys/go-corset/pkg/trace"
	"github.com/consensys/go-corset/pkg/util"
	"github.com/consensys/go-corset/pkg/util/termio"
)

// TraceFormatting provides a top-level notion of formatting
type TraceFormatting interface {
	Module(ModuleData) ModuleFormatting
}

// ModuleFormatting provides a generic mechanism for applying formatting to
// modules being refendered.
type ModuleFormatting interface {
	// ColumnTitle returns optional formating for a given column title.
	ColumnTitle(uint) util.Option[termio.AnsiEscape]
	// ColumnTitle returns optional formating for a given row title.
	RowTitle(sc.RegisterId) util.Option[termio.AnsiEscape]
	// Cell returns optional formating for a given cells in the trace.
	Cell(sc.RegisterId, uint) util.Option[termio.AnsiEscape]
}

// DefaultFormatter returns a default formatting
func DefaultFormatter() TraceFormatting {
	return &defaultFormatter{}
}

type defaultFormatter struct{}

// ColumnTitle implementation for Formatting interface
func (p *defaultFormatter) ColumnTitle(uint) util.Option[termio.AnsiEscape] {
	return util.None[termio.AnsiEscape]()
}

// ColumnTitle implementation for Formatting interface
func (p *defaultFormatter) Module(ModuleData) ModuleFormatting {
	return p
}

// ColumnTitle implementation for Formatting interface
func (p *defaultFormatter) RowTitle(sc.RegisterId) util.Option[termio.AnsiEscape] {
	return util.None[termio.AnsiEscape]()
}

// Cell implementation for Formatting interface
func (p *defaultFormatter) Cell(tr.ColumnId, uint) util.Option[termio.AnsiEscape] {
	return util.None[termio.AnsiEscape]()
}

// ============================================================================
// Cell Formatting
// ============================================================================

// NewCellFormatter constructs a new cell formatter
func NewCellFormatter(cells CellRefSet) TraceFormatting {
	return &cellFormatter{nil, cells}
}

type cellFormatter struct {
	data  ModuleData
	cells CellRefSet
}

// ColumnTitle implementation for Formatting interface
func (p *cellFormatter) ColumnTitle(uint) util.Option[termio.AnsiEscape] {
	return util.Some(termio.NewAnsiEscape().FgColour(termio.TERM_WHITE))
}

// ColumnTitle implementation for ModuleFormatting interface
func (p *cellFormatter) RowTitle(sc.RegisterId) util.Option[termio.AnsiEscape] {
	return util.Some(termio.NewAnsiEscape().FgColour(termio.TERM_WHITE))
}

// Module implementation for TraceFormatting interface
func (p *cellFormatter) Module(mod ModuleData) ModuleFormatting {
	var q = *p
	//
	q.data = mod
	//
	return &q
}

// Cell implementation for ModuleFormatting interface
func (p *cellFormatter) Cell(col tr.ColumnId, row uint) util.Option[termio.AnsiEscape] {
	var (
		// Extract limbs
		limbs = p.data.Mapping().LimbIds(col)
	)
	// Check each limb in turn
	for _, lid := range limbs {
		// Consrtuct column reference
		limbRef := tr.NewColumnRef(p.data.Id(), lid)
		// Construct cellRef reference
		cellRef := tr.NewCellRef(limbRef, int(row))
		// if any limb involved, entire column involved.
		if p.cells.Contains(cellRef) {
			return util.Some(termio.BoldAnsiEscape().FgColour(termio.TERM_RED))
		}
	}
	//
	return util.None[termio.AnsiEscape]()
}
