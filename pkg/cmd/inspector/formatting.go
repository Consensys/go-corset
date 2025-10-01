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
	"math/big"

	"github.com/consensys/go-corset/pkg/cmd/view"
	sc "github.com/consensys/go-corset/pkg/schema"
	tr "github.com/consensys/go-corset/pkg/trace"
	"github.com/consensys/go-corset/pkg/util"
	"github.com/consensys/go-corset/pkg/util/termio"
)

// NewFormatter constructs a new cell formatter
func NewFormatter() view.TraceFormatting {
	return &inspectorFormatter{nil}
}

type inspectorFormatter struct {
	module view.ModuleData
}

// ColumnTitle implementation for ModuleFormatting interface
func (p *inspectorFormatter) ColumnTitle(uint) util.Option[termio.AnsiEscape] {
	return util.Some(termio.NewAnsiEscape().FgColour(termio.TERM_BLUE))
}

// ColumnTitle implementation for ModuleFormatting interface
func (p *inspectorFormatter) RowTitle(reg sc.RegisterId) util.Option[termio.AnsiEscape] {
	if p.module.DataOf(reg).IsComputed() {
		return util.Some(termio.NewAnsiEscape().FgColour(termio.TERM_GREEN))
	}
	//
	return util.Some(termio.NewAnsiEscape().FgColour(termio.TERM_BLUE))
}

// Cell implementation for ModuleFormatting interface
func (p *inspectorFormatter) Cell(col tr.ColumnId, row uint) util.Option[termio.AnsiEscape] {
	val := p.module.DataOf(col).Get(row)
	//return util.None[termio.AnsiEscape]()
	return util.Some(cellColour(val))
}

// Module implementation for TraceFormatting interface
func (p *inspectorFormatter) Module(mod view.ModuleData) view.ModuleFormatting {
	var q = *p
	//
	q.module = mod
	//
	return &q
}

// This algorithm is based on that used in the original tool.  To understand
// this algorithm, you need to look at the 256 colour table for ANSI escape
// codes.  It actually does make sense, even if it doesn't appear to.
func cellColour(val big.Int) termio.AnsiEscape {
	if val.Cmp(biZero) == 0 {
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
