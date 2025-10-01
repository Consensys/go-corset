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
	tr "github.com/consensys/go-corset/pkg/trace"
	"github.com/consensys/go-corset/pkg/util/field"
)

// ModuleData abstracts the raw data of a module.
type ModuleData interface {
	// Id returns the module identifier
	Id() sc.ModuleId
	// Access abtract data for given register
	DataOf(sc.RegisterId) RegisterView
	// Dimensions returns width and height of data
	Dimensions() (uint, uint)
	// Determines whether or not this module is externally visible.
	IsPublic() bool
	// Mapping returns the register limbs map being used by this module view.
	Mapping() sc.RegisterLimbsMap
	// Name returns the name of the given module
	Name() string
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
	mapping sc.RegisterLimbsMap
	// Enumeration values
	enumerations []corset.Enumeration
	// public modifier
	public bool
	// Trace provides the raw data for this view
	trace tr.Module[F]
	// Set of column titles
	columns []string
	// Set of rows in this window
	rows []rowData
}

func newModuleData[F field.Element[F]](id sc.ModuleId, mapping sc.RegisterLimbsMap, trace tr.Module[F], public bool,
	display []uint, enums []corset.Enumeration) *moduleData[F] {
	//
	var data []rowData
	// Iterate source-level registers
	for i, reg := range mapping.Registers() {
		// construct source-level register id
		rid := sc.NewRegisterId(uint(i))
		// determine corresponding limbs
		limbs := mapping.LimbIds(rid)
		//
		data = append(data, rowData{
			limbs: limbs,
			// Determine column name
			name: reg.Name,
			// Display info
			display: display[i],
			// Render column data from all limbs
			data: nil,
		})
	}
	//
	return &moduleData[F]{id, trace.Height(), mapping, enums, public, trace, nil, data}
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

// Data returns an abtract view of the data for given register
func (p *moduleData[F]) DataOf(reg sc.RegisterId) RegisterView {
	return &registerView[F]{
		p.trace, reg, p.mapping,
	}
}

func (p *moduleData[F]) Dimensions() (uint, uint) {
	return p.height, uint(len(p.rows))
}

func (p *moduleData[F]) Id() sc.ModuleId {
	return p.id
}

func (p *moduleData[F]) IsPublic() bool {
	return p.public
}

// Mapping returns the register-limbs mapping used within this view.
func (p *moduleData[F]) Mapping() sc.RegisterLimbsMap {
	return p.mapping
}

// Name return name of this module
func (p *moduleData[F]) Name() string {
	return p.trace.Name()
}

// RowTitle returns the title for a given data row
func (p *moduleData[F]) RowTitle(row sc.RegisterId) string {
	return p.rows[row.Unwrap()].name
}

func (p *moduleData[F]) expand(col, row uint) {
	var (
		rowData = p.rows[row]
		n       = uint(len(rowData.data))
	)
	// Check whether expansion required
	if col >= n {
		// Yes
		ndata := make([]string, col+1)
		//
		view := p.DataOf(sc.NewRegisterId(row))
		// Copy existing data
		copy(ndata, rowData.data)
		// Construct new data
		for i := n; i <= col; i++ {
			ith := view.Get(i)
			ndata[i] = renderCellValue(rowData.display, ith, p.enumerations)
		}
		//
		p.rows[row].data = ndata
	}
}

type rowData struct {
	// column name
	name string
	// display modifier
	display uint
	// limbs making up this row
	limbs []sc.RegisterId
	// rendered column data
	data []string
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
