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
	"math/big"
	"strings"

	"github.com/consensys/go-corset/pkg/corset"
	sc "github.com/consensys/go-corset/pkg/schema"
	"github.com/consensys/go-corset/pkg/schema/register"
	tr "github.com/consensys/go-corset/pkg/trace"
	"github.com/consensys/go-corset/pkg/util"
	"github.com/consensys/go-corset/pkg/util/field"
)

// SourceColumnId abstracts the idea of a source-level column declaration.  Due
// to register allocation, we can have multiple source-level columns mapped to
// the same register; likewise, due to register splitting, we can have one
// register mapping to multiple limbs.
type SourceColumnId = register.Id

// SourceColumn provides key information to the inspector about source-level
// columns and their mapping to registers at the MIR/AIR levels (i.e. columns we
// would find in the trace).
type SourceColumn struct {
	// column Name
	Name string
	// Display modifier
	Display uint
	// Determines whether this is a Computed column.
	Computed bool
	// Selector determines when column active.
	Selector util.Option[string]
	// RegisterId to which this column was allocated.
	Register register.Id
	// Limbs making up the register to which this column is allocated.
	Limbs []register.Id
	// rendered column data
	data []string
}

// ModuleData abstracts the raw data of a module.
type ModuleData interface {
	// Id returns the module identifier
	Id() sc.ModuleId
	// Access abtract data for given register
	DataOf(register.Id) RegisterView
	// Dimensions returns width and height of data
	Dimensions() (uint, uint)
	// Determine whether a given source column is active on a given row.  A
	// source column declared within a perspective will only be active when the
	// given perspective's selector is enabled.
	IsActive(SourceColumn, uint) bool
	// Determines whether or not this module is externally visible.
	IsPublic() bool
	// Mapping returns the register limbs map being used by this module view.
	Mapping() register.LimbsMap
	// Name returns the name of the given module
	Name() string
	// SourceColumn returns the source column associated with a given id.
	SourceColumn(col SourceColumnId) SourceColumn
	// SourceColumnOf returns the source column associated with a given name.
	SourceColumnOf(name string) SourceColumn
}

// ============================================================================
// Module Data
// ============================================================================

type moduleData[F field.Element[F]] struct {
	// Module identifier
	id sc.ModuleId
	// Height of module
	height uint
	// Mapping registers <-> limbs
	mapping register.LimbsMap
	// Enumeration values
	enumerations []corset.Enumeration
	// public modifier
	public bool
	// Trace provides the raw data for this view
	trace tr.Module[F]
	// Set of column titles
	columns []string
	// Set of rows in this window
	rows []SourceColumn
}

func newModuleData[F field.Element[F]](id sc.ModuleId, mapping register.LimbsMap, trace tr.Module[F], public bool,
	enums []corset.Enumeration, rows []SourceColumn) *moduleData[F] {
	//
	return &moduleData[F]{id, trace.Height(), mapping, enums, public, trace, nil, rows}
}

// CellAt returns the contents of a specific cell in this table.
func (p *moduleData[F]) CellAt(col uint, row uint) string {
	// Ensure enough space
	p.expand(col, row)
	//
	return p.rows[row].data[col]
}

// ColumnTitle returns the title for a given data column
func (p *moduleData[F]) ColumnTitle(col uint) string {
	// Construct titles lazily
	if col >= uint(len(p.columns)) {
		ncols := make([]string, col+1)
		copy(ncols, p.columns)
		//
		for i := len(p.columns); i < len(ncols); i++ {
			ncols[i] = fmt.Sprintf("#%d", i)
		}
		//
		p.columns = ncols
	}
	//
	return p.columns[col]
}

// IsActive determines whether a given cell is active, or not.  A cell can be
// inactive, for example, if its part of a perspective which is not active (on
// the given row).
func (p *moduleData[F]) IsActive(col SourceColumn, row uint) bool {
	// Santity check whether actually need to do anything
	if col.Selector.IsEmpty() {
		return true
	}
	// Extract relevant selector
	selector := p.SourceColumnOf(col.Selector.Unwrap())
	// Extract selector's value on this row
	val := p.DataOf(selector.Register).Get(row)
	// Check whether selector is active (or not)
	return val.BitLen() != 0
}

// SourceColumn returns the source column associated with the given source
// column id.
func (p *moduleData[F]) SourceColumn(col SourceColumnId) SourceColumn {
	return p.rows[col.Unwrap()]
}

// SourceColumnOf returns the source column associated with the given source
// column name.  This will panic if the given source column does not exist.
func (p *moduleData[F]) SourceColumnOf(name string) SourceColumn {
	for _, col := range p.rows {
		if col.Name == name {
			return col
		}
	}
	//
	panic(fmt.Sprintf("unknown source column %s", name))
}

// Data returns an abtract view of the data for given register
func (p *moduleData[F]) DataOf(reg register.Id) RegisterView {
	return &registerView[F]{
		p.trace, reg, p.mapping,
	}
}

func (p *moduleData[F]) Dimensions() (uint, uint) {
	return p.height, uint(len(p.rows))
}

// Window constructs a fresh window capturing this module data.
func (p *moduleData[F]) Window() Window {
	var (
		width, height = p.Dimensions()
		rows          = make([]SourceColumnId, height)
	)
	//
	for i := range height {
		rows[i] = register.NewId(i)
	}
	//
	return NewWindow(width, rows)
}

func (p *moduleData[F]) Filter(filter ModuleFilter) Window {
	var (
		q          Window = p.Window()
		nrows      []SourceColumnId
		width, _   = q.Dimensions()
		start, end = filter.Range()
	)
	//
	for i, ith := range p.rows {
		// Construct source column id
		sid := register.NewId(uint(i))
		// If any limb is included, the whole limb is included.
		if filter.Column(ith) {
			nrows = append(nrows, sid)
		}
	}
	// Finalise window
	q.rows = nrows
	q.startCol, q.endCol = min(width, start), min(width, end)
	//
	return q
}

func (p *moduleData[F]) Id() sc.ModuleId {
	return p.id
}

func (p *moduleData[F]) IsPublic() bool {
	return p.public
}

// Mapping returns the register-limbs mapping used within this view.
func (p *moduleData[F]) Mapping() register.LimbsMap {
	return p.mapping
}

// Name return name of this module
func (p *moduleData[F]) Name() string {
	return p.trace.Name()
}

// RowTitle returns the title for a given data row
func (p *moduleData[F]) RowTitle(row register.Id) string {
	return p.rows[row.Unwrap()].Name
}

func (p *moduleData[F]) expand(col, row uint) {
	var (
		srcColumn = p.rows[row]
		n         = uint(len(srcColumn.data))
	)
	// Check whether expansion required
	if col >= n {
		// Yes
		ndata := make([]string, col+1)
		//
		view := p.DataOf(srcColumn.Register)
		// Copy existing data
		copy(ndata, srcColumn.data)
		// Construct new data
		for i := n; i <= col; i++ {
			ith := view.Get(i)
			ndata[i] = renderCellValue(srcColumn.Display, ith, p.enumerations)
		}
		//
		p.rows[row].data = ndata
	}
}

// Determine the (unclipped) string value at a given column and row in a given
// trace.
func renderCellValue(disp uint, val big.Int, enums []corset.Enumeration) string {
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
		if enumID < len(enums) {
			var index big.Int
			//
			index.SetBytes(val.Bytes())
			//
			if index.IsUint64() {
				// Check whether value covered by enumeration.
				if lab, ok := enums[enumID][index.Uint64()]; ok {
					return lab
				}
			}
		}
	}
	// Default:
	return fmt.Sprintf("0x%s", val.Text(16))
}

// Format a field element according to the ":bytes" directive.
func displayBytes(val big.Int) string {
	var (
		builder strings.Builder
	)
	// Handle zero case specifically.
	if val.BitLen() == 0 {
		return "00"
	}
	//
	for i, b := range val.Bytes() {
		if i != 0 {
			builder.WriteString(" ")
		}
		//
		builder.WriteString(fmt.Sprintf("%02x", b))
	}
	//
	return builder.String()
}
