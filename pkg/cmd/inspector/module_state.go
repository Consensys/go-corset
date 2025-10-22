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
	"math"
	"math/big"
	"regexp"

	"github.com/consensys/go-corset/pkg/cmd/view"
	"github.com/consensys/go-corset/pkg/schema/register"
	"github.com/consensys/go-corset/pkg/util/collection/array"
	"github.com/consensys/go-corset/pkg/util/termio"
)

// ModuleState provides state regarding how to display the trace for a given
// module, including related aspects like filter histories, etc.
type ModuleState struct {
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
	lastQuery *Query
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
	// Register mappin
	Mapping register.Map
}

// Column matches this filter against a given column.
func (p *SourceColumnFilter) Column(col view.SourceColumn) bool {
	//
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

// Range imoplementation for ModuleFilter interface.
func (p *SourceColumnFilter) Range() (start, end uint) {
	return 0, math.MaxUint
}

func newModuleState(view view.ModuleView, public bool) ModuleState {
	var state ModuleState
	//
	state.public = public
	state.view = view
	// Include all columns initially
	state.columnFilter.Computed = true
	state.columnFilter.UserDefined = true
	state.columnFilter.Mapping = view.Data().Mapping().LimbsMap()
	//
	return state
}

func (p *ModuleState) gotoRow(ncol uint) uint {
	col, row := p.view.Offset()
	//
	p.view.Goto(ncol, row)
	//
	if col != ncol {
		// Update history
		rowOffsetStr := fmt.Sprintf("%d", ncol)
		p.targetRowHistory = history_append(p.targetRowHistory, rowOffsetStr)
	}
	// failed
	return row
}

// Apply a new column filter to the module view.  This determines which columns
// are currently visible.
func (p *ModuleState) applyColumnFilter(filter SourceColumnFilter, history bool) {
	p.view = p.view.Filter(&filter)
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
func (p *ModuleState) matchQuery(col uint, forwards bool, query *Query) termio.FormattedText {
	var (
		mapping  = p.view.Data().Mapping()
		width, _ = p.view.Data().Dimensions()
		// Construct query environment
		env = func(col string, row uint) big.Int {
			id, ok := mapping.HasRegister(col)
			//
			if !ok {
				panic(fmt.Sprintf("unknown column \"%s\"", col))
			}
			//
			return p.view.Data().DataOf(id).Get(row)
		}
		// Direction text
		dir string
		// Determine current cursor offset
		_, row = p.view.Offset()
	)
	// set direction
	if forwards {
		dir = "forwards"
	} else {
		dir = "backwards"
	}
	// Always update history
	p.scanHistory = history_append(p.scanHistory, query.String())
	p.lastQuery = query
	// evaluate forward
	for i := col; i < width; {
		val := query.Eval(i, env)
		//
		if val.Cmp(biZero) == 0 {
			p.view.Goto(i, row)
			msg := fmt.Sprintf("%s from row %d, matched row %d", dir, col, i)

			return termio.NewColouredText(msg, termio.TERM_GREEN)
		}
		//
		if forwards {
			i++
		} else {
			i--
		}
	}
	//
	msg := fmt.Sprintf("%s from row %d, matched nothing", dir, col)
	//
	return termio.NewColouredText(msg, termio.TERM_YELLOW)
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
