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

	"github.com/consensys/go-corset/pkg/cmd/view"
	"github.com/consensys/go-corset/pkg/schema"
	"github.com/consensys/go-corset/pkg/util"
	"github.com/consensys/go-corset/pkg/util/collection/array"
	"github.com/consensys/go-corset/pkg/util/field"
	"github.com/consensys/go-corset/pkg/util/termio"
)

// ModuleState provides state regarding how to display the trace for a given
// module, including related aspects like filter histories, etc.
type ModuleState[F field.Element[F]] struct {
	// public indicates whether or not this module is externally visible.
	public bool
	// Active module view
	view view.ModuleView
	// History for goto row commands
	targetRowHistory []string
	// Active column filter
	columnFilter SourceColumnFilter
	// Set of column filters used (regexes only).
	columnFilterHistory []string
	// History for scan commands
	scanHistory []string
	// Last executed query for next scan
	lastQuery *Query[F]
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
	Selector util.Option[string]
	// Display modifier
	Display uint
	// Register to which this column is allocated
	Register schema.RegisterRef
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

func newModuleState[F field.Element[F]](view view.ModuleView, public bool) ModuleState[F] {
	var state ModuleState[F]
	//
	state.public = public
	state.view = view
	// Include all columns initially
	state.columnFilter.Computed = true
	state.columnFilter.UserDefined = true
	//
	return state
}

func (p *ModuleState[F]) gotoRow(nrow uint) uint {
	col, row := p.view.Offset()
	//
	p.view.Goto(col, nrow)
	//
	if row != nrow {
		// Update history
		rowOffsetStr := fmt.Sprintf("%d", nrow)
		p.targetRowHistory = history_append(p.targetRowHistory, rowOffsetStr)
	}
	// failed
	return row

}

// Apply a new column filter to the module view.  This determines which columns
// are currently visible.
func (p *ModuleState[F]) applyColumnFilter(filter SourceColumnFilter, history bool) {
	// filteredColumns := make([]SourceColumn, 0)
	// // Apply filter
	// for _, col := range p.columns {
	// 	// Check whether it matches the regex or not.
	// 	if filter.Match(col) {
	// 		filteredColumns = append(filteredColumns, col)
	// 	}
	// }
	// // Update the view
	// p.view.SetActiveColumns(p.trace, filteredColumns)
	// // Save active filter
	// p.columnFilter = filter
	// // Update selection and history
	// if filter.Regex != nil {
	// 	//
	// 	if history {
	// 		regex_string := filter.Regex.String()
	// 		p.columnFilterHistory = history_append(p.columnFilterHistory, regex_string)
	// 	}
	// }
	panic("todo")
}

// Evaluate a query on the current module using those values from the given
// trace, looking for the first row where the query holds.
func (p *ModuleState[F]) matchQuery(row uint, forwards bool, query *Query[F]) termio.FormattedText {
	// var (
	// 	env = make(map[string]tr.Column[F])
	// 	dir string
	// )
	// // set direction
	// if forwards {
	// 	dir = "forwards"
	// } else {
	// 	dir = "backwards"
	// }
	// // Always update history
	// p.scanHistory = history_append(p.scanHistory, query.String())
	// p.lastQuery = query
	// // construct environment
	// for _, col := range p.columns {
	// 	env[col.Name] = p.trace.Column(col.Register)
	// }
	// // evaluate forward
	// for i := row; i < p.height(); {
	// 	val := query.Eval(i, env)
	// 	//
	// 	if val.IsZero() {
	// 		r := p.setRowOffset(i)
	// 		msg := fmt.Sprintf("%s from row %d, matched row %d", dir, row, r)

	// 		return termio.NewColouredText(msg, termio.TERM_GREEN)
	// 	}
	// 	//
	// 	if forwards {
	// 		i++
	// 	} else {
	// 		i--
	// 	}
	// }
	// //
	// msg := fmt.Sprintf("%s from row %d, matched nothing", dir, row)
	// //
	// return termio.NewColouredText(msg, termio.TERM_YELLOW)
	panic("todo")
}

// History append will append a given item to the end of the history.  However,
// if that item already existed in the history, then that is removed.  This is
// to avoid duplicates in the history.
func history_append[T comparable](history []T, item T) []T {
	// Remove previous entry (if applicable)
	history = array.RemoveMatching(history, func(ith T) bool { return ith == item })
	// Add item to end
	return append(history, item)
}
