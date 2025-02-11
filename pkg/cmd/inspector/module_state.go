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
	"fmt"
	"regexp"
	"slices"
	"strings"

	"github.com/consensys/go-corset/pkg/corset"
	"github.com/consensys/go-corset/pkg/hir"
	tr "github.com/consensys/go-corset/pkg/trace"
	"github.com/consensys/go-corset/pkg/util"
)

// ModuleState provides state regarding how to display the trace for a given
// module, including related aspects like filter histories, etc.
type ModuleState struct {
	// Module name
	name string
	// Identifies trace columns in this module.
	columns []SourceColumn
	// Active module view
	view ModuleView
	// History for goto row commands
	targetRowHistory []string
	// Active column filter
	columnFilter string
	// Set of column filters used.
	columnFilterHistory []string
}

// SourceColumn provides key information to the inspector about source-level
// columns and their mapping to registers at the HIR level (i.e. columns we
// would find in the trace).
type SourceColumn struct {
	// Column name
	Name string
	// Selector determines when column active.
	Selector *hir.UnitExpr
	// Display modifier
	Display uint
	// Register to which this column is allocated
	Register uint
}

func newModuleState(module *corset.SourceModule, trace tr.Trace, enums []corset.Enumeration,
	recurse bool) ModuleState {
	//
	var (
		state      ModuleState
		submodules []corset.SourceModule
	)
	// Handle non-root modules
	if recurse {
		submodules = module.Submodules
	}
	//
	state.name = module.Name
	// Extract source columns from module tree
	state.columns = extractSourceColumns(util.NewAbsolutePath(""), module.Selector, module.Columns, submodules)
	// Sort all column names so that, for example, columns in the same
	// perspective are grouped together.
	slices.SortFunc(state.columns, func(l SourceColumn, r SourceColumn) int {
		return strings.Compare(l.Name, r.Name)
	})
	// Configure view
	state.view.maxRowWidth = 16
	state.view.enumerations = enums
	// Finalise view
	state.view.SetActiveColumns(trace, state.columns)
	//
	return state
}

func (p *ModuleState) setColumnOffset(colOffset uint) {
	p.view.SetColumn(colOffset)
}

func (p *ModuleState) setRowOffset(rowOffset uint) bool {
	if p.view.SetRow(rowOffset) {
		// Update history
		rowOffsetStr := fmt.Sprintf("%d", rowOffset)
		p.targetRowHistory = history_append(p.targetRowHistory, rowOffsetStr)
		//
		return true
	}
	// failed
	return false
}

// Apply a new column filter to the module view.  This determines which columns
// are currently visible.
func (p *ModuleState) applyColumnFilter(trace tr.Trace, regex *regexp.Regexp, history bool) {
	filteredColumns := make([]SourceColumn, 0)
	// Apply filter
	for _, col := range p.columns {
		// Check whether it matches the regex or not.
		if name := col.Name; regex.MatchString(name) {
			filteredColumns = append(filteredColumns, col)
		}
	}
	// Update the view
	p.view.SetActiveColumns(trace, filteredColumns)
	// Update selection and history
	p.columnFilter = regex.String()
	//
	if history {
		p.columnFilterHistory = history_append(p.columnFilterHistory, regex.String())
	}
}

// History append will append a given item to the end of the history.  However,
// if that item already existed in the history, then that is removed.  This is
// to avoid duplicates in the history.
func history_append[T comparable](history []T, item T) []T {
	// Remove previous entry (if applicable)
	history = util.RemoveMatching(history, func(ith T) bool { return ith == item })
	// Add item to end
	return append(history, item)
}

func extractSourceColumns(path util.Path, selector *hir.UnitExpr, columns []corset.SourceColumn,
	submodules []corset.SourceModule) []SourceColumn {
	//
	var srcColumns []SourceColumn
	//
	for _, col := range columns {
		name := path.Extend(col.Name).String()[1:]
		srcCol := SourceColumn{name, selector, col.Display, col.Register}
		srcColumns = append(srcColumns, srcCol)
	}
	//
	for _, submod := range submodules {
		subpath := path.Extend(submod.Name)
		subSrcColumns := extractSourceColumns(*subpath, submod.Selector, submod.Columns, submod.Submodules)
		srcColumns = append(srcColumns, subSrcColumns...)
	}
	//
	return srcColumns
}
