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
	"github.com/consensys/go-corset/pkg/util/termio"
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
	columnFilter SourceColumnFilter
	// Set of column filters used (regexes only).
	columnFilterHistory []string
	// Histor for scan commands
	scanHistory []string
}

// SourceColumn provides key information to the inspector about source-level
// columns and their mapping to registers at the HIR level (i.e. columns we
// would find in the trace).
type SourceColumn struct {
	// Column name
	Name string
	// Determines whether this is a Computed column.
	Computed bool
	// Selector determines when column active.
	Selector *hir.Expr
	// Display modifier
	Display uint
	// Register to which this column is allocated
	Register uint
}

// SourceColumnFilter packages up everything needed for filtering columns in a
// given module.
type SourceColumnFilter struct {
	// Regex filters columns based on whether their name matches the regex or
	// not.
	Regex *regexp.Regexp
	// Computed filters columns based on whether they are computed.
	Computed bool
	// UserDefined filters columns based on whether they are non-computed columns.
	UserDefined bool
}

// Match this filter against a given column.
func (p *SourceColumnFilter) Match(col SourceColumn) bool {
	if p.Regex == nil || p.Regex.MatchString(col.Name) {
		if p.Computed && col.Computed {
			return true
		} else if p.UserDefined && !col.Computed {
			return true
		}
	}
	// failed
	return false
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
	// Include all columns initially
	state.columnFilter.Computed = true
	state.columnFilter.UserDefined = true
	// Extract source columns from module tree
	state.columns = ExtractSourceColumns(util.NewAbsolutePath(""), module.Selector, module.Columns, submodules)
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

func (p *ModuleState) height() uint {
	return uint(len(p.view.rowWidths))
}

func (p *ModuleState) setColumnOffset(colOffset uint) {
	p.view.SetColumn(colOffset)
}

func (p *ModuleState) setRowOffset(rowOffset uint) uint {
	row := p.view.SetRow(rowOffset)
	//
	if row != rowOffset {
		// Update history
		rowOffsetStr := fmt.Sprintf("%d", rowOffset)
		p.targetRowHistory = history_append(p.targetRowHistory, rowOffsetStr)
	}
	// failed
	return row
}

// Apply a new column filter to the module view.  This determines which columns
// are currently visible.
func (p *ModuleState) applyColumnFilter(trace tr.Trace, filter SourceColumnFilter, history bool) {
	filteredColumns := make([]SourceColumn, 0)
	// Apply filter
	for _, col := range p.columns {
		// Check whether it matches the regex or not.
		if filter.Match(col) {
			filteredColumns = append(filteredColumns, col)
		}
	}
	// Update the view
	p.view.SetActiveColumns(trace, filteredColumns)
	// Save active filter
	p.columnFilter = filter
	// Update selection and history
	if filter.Regex != nil {
		//
		if history {
			regex_string := filter.Regex.String()
			p.columnFilterHistory = history_append(p.columnFilterHistory, regex_string)
		}
	}
}

// Evaluate a query on the current module using those values from the given
// trace, looking for the first row where the query holds.
func (p *ModuleState) matchQuery(query *Query, trace tr.Trace) termio.FormattedText {
	// Always update history
	p.scanHistory = history_append(p.scanHistory, query.String())
	// Proceed
	env := make(map[string]tr.Column)
	// construct environment
	for _, col := range p.columns {
		env[col.Name] = trace.Column(col.Register)
	}
	// evaluate forward
	for i := uint(0); i < p.height(); i++ {
		val := query.Eval(i, env)
		//
		if val.IsZero() {
			r := p.setRowOffset(i)
			return termio.NewColouredText(fmt.Sprintf("Matched row %d", r), termio.TERM_GREEN)
		}
	}
	//
	return termio.NewColouredText("Matched nothing", termio.TERM_YELLOW)
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

// ExtractSourceColumns extracts source column descriptions for a given module
// based on the corset source mapping.  This is particularly useful when you
// want to show the original name for a column (e.g. when its in a perspective),
// rather than the raw register name.
func ExtractSourceColumns(path util.Path, selector *hir.Expr, columns []corset.SourceColumn,
	submodules []corset.SourceModule) []SourceColumn {
	//
	var srcColumns []SourceColumn
	//
	for _, col := range columns {
		name := path.Extend(col.Name).String()[1:]
		srcCol := SourceColumn{name, col.Computed, selector, col.Display, col.Register}
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
