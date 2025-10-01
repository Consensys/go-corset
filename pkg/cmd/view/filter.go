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
	"github.com/consensys/go-corset/pkg/util/collection/bit"
)

// Filter is used to filter a given TraceView to focus on key aspects of the
// view, as required for the task at hand.
type Filter interface {
	// Determine filter for columns in the given module; if nil is returned,
	// then module is ignored.
	Module(sc.ModuleId) ColumnFilter
}

// ColumnFilter is used to focus on a subset of columns in a given module.
type ColumnFilter interface {
	// Determine filter for cells of this column; if nil is returned, then
	// column is ignored.
	Column(sc.LimbId) CellFilter
}

// CellFilter is used to focus on a subset of cells in a given column.
type CellFilter interface {
	Cell(uint) bool
}

// ============================================================================
// Default Filter
// ============================================================================

type defaultFilter struct{}

// DefaultFilter constructs a default filter which filters nothing.
func DefaultFilter() Filter {
	return &defaultFilter{}
}

// DefaultColumnFilter constructs a default column filter which filters nothing.
func DefaultColumnFilter() ColumnFilter {
	return &defaultFilter{}
}

// DefaultCellFilter constructs a default cell filter which filters nothing.
func DefaultCellFilter() CellFilter {
	return &defaultFilter{}
}

func (p *defaultFilter) Module(sc.ModuleId) ColumnFilter {
	return p
}

func (p *defaultFilter) Column(sc.LimbId) CellFilter {
	return p
}

func (p *defaultFilter) Cell(uint) bool {
	return true
}

// ============================================================================
// Module Filter
// ============================================================================

// FilterForModules constructs a filter from a given predicate.
func FilterForModules(fn func(sc.ModuleId) bool) Filter {
	return &moduleFilter{fn}
}

type moduleFilter struct {
	fn func(sc.ModuleId) bool
}

func (p *moduleFilter) Module(mid sc.ModuleId) ColumnFilter {
	if p.fn(mid) {
		return DefaultColumnFilter()
	}
	//
	return nil
}

// ============================================================================
// Cell Filter
// ============================================================================

// FilterForCells returns a filter which focuses specifically on the given
// cells.
func FilterForCells(cells CellRefSet) Filter {
	var filter moduleCellFilter
	//
	for iter := cells.Iter(); iter.HasNext(); {
		filter.addCell(iter.Next())
	}
	//
	return &filter
}

type moduleCellFilter []columnCellFilter
type columnCellFilter []cellFilter
type cellFilter bit.Set

func (p *moduleCellFilter) Module(mid sc.ModuleId) ColumnFilter {
	var n = uint(len((*p)[mid]))
	//
	if n <= mid || (*p)[mid] == nil {
		return nil
	}
	//
	return &(*p)[mid]
}

func (p *moduleCellFilter) addCell(cell tr.CellRef) {
	var (
		bits *bit.Set
		m    = cell.Column.Module()
		n    = cell.Column.Register().Unwrap()
	)
	// First ensure enough modules
	if uint(len(*p)) <= m {
		q := make([]columnCellFilter, m+1)
		copy(q, *p)
		*p = q
	}
	// Second ensure enough columns
	if uint(len((*p)[m])) <= n {
		q := make([]cellFilter, n+1)
		copy(q, (*p)[m])
		(*p)[m] = q
	}
	// Finally, add the cell
	bits = (*bit.Set)(&(*p)[m][n])
	bits.Insert(uint(cell.Row))
}

func (p *columnCellFilter) Column(cid sc.LimbId) CellFilter {
	var bits bit.Set = bit.Set((*p)[cid.Unwrap()])
	//
	if bits.Count() == 0 {
		return nil
	}
	//
	return &(*p)[cid.Unwrap()]
}

func (p *cellFilter) Cell(row uint) bool {
	var bits bit.Set = bit.Set(*p)
	return bits.Contains(row)
}
