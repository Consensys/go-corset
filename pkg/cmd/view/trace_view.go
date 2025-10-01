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
	"math/big"

	"github.com/consensys/go-corset/pkg/corset"
	"github.com/consensys/go-corset/pkg/schema"
	tr "github.com/consensys/go-corset/pkg/trace"
	"github.com/consensys/go-corset/pkg/util"
	"github.com/consensys/go-corset/pkg/util/field"
	"github.com/consensys/go-corset/pkg/util/file"
)

// ReadColumn reads the value of a given column abstraction from a given trace,
// converting it into a big integer.
func ReadColumn[F field.Element[F]](row uint, view ColumnView, trace tr.Trace[F]) big.Int {
	var result big.Int
	//
	if len(view.Register) > 1 {
		panic("todo")
	}
	//
	result.SetBytes(trace.Column(view.Register[0]).Data().Get(row).Bytes())
	//
	return result
}

// ModuleView abstracts the underlying modules in a trace in such a way as to
// produce human-readable output.
type ModuleView struct {
	// Module name
	Name string
	// Indicates whether externally visible
	Public bool
	// Indicates whether artificial or not.
	Synthetic bool
	// Columns making up this view.
	Columns []ColumnView
}

// ColumnView abstracts the underlying columns in a trace in such a way as to
// produce human-readable output.
type ColumnView struct {
	// Column name
	Name string
	// Determines whether this is a Computed column.
	Computed bool
	// Selector determines when column active.
	Selector util.Option[string]
	// Display modifier
	Display uint
	// Register(s) to which this column is allocated
	Register []schema.RegisterRef
}

// Includes checks whether a given register reference forms part of this column.
func (p *ColumnView) Includes(reg schema.RegisterRef) bool {
	for _, r := range p.Register {
		if r == reg {
			return true
		}
	}
	//
	return false
}

// Module returns the enclosing module of this abstract column.
func (p *ColumnView) Module() schema.ModuleId {
	return p.Register[0].Module()
}

// ExtractSourceColumns extracts source column descriptions for a given module
// based on the corset source mapping.  This is particularly useful when you
// want to show the original name for a column (e.g. when its in a perspective),
// rather than the raw register name.
func ExtractSourceColumns(path file.Path, selector util.Option[string], columns []corset.SourceColumn,
	submodules []corset.SourceModule) []ColumnView {
	//
	var srcColumns []ColumnView
	//
	for _, col := range columns {
		name := path.Extend(col.Name).String()[1:]
		srcRegs := []schema.RegisterRef{col.Register}
		srcCol := ColumnView{name, col.Computed, selector, col.Display, srcRegs}
		srcColumns = append(srcColumns, srcCol)
	}
	//
	for _, submod := range submodules {
		subpath := path.Extend(submod.Name)
		subSrcColumns := ExtractSourceColumns(*subpath, submod.Selector, submod.Columns, submod.Submodules)
		srcColumns = append(srcColumns, subSrcColumns...)
	}
	//
	return srcColumns
}
