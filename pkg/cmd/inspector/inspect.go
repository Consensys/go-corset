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
	"time"

	"github.com/consensys/go-corset/pkg/corset"
	sc "github.com/consensys/go-corset/pkg/schema"
	tr "github.com/consensys/go-corset/pkg/trace"
	"github.com/consensys/go-corset/pkg/util/termio"
	"github.com/consensys/go-corset/pkg/util/termio/widget"
)

// ==================================================================
// Inspector
// ==================================================================

// DEFAULT_MODE sets the default command bar, and allows the user to navigate
// the trace.
const DEFAULT_MODE = 0

// NUMERIC_INPUT_MODE is where the user is entering a numberic value (e.g. to
// specify the row for a goto command).
const NUMERIC_INPUT_MODE = 1

// TEXT_INPUT_MODE is where the user is entering a text value (e.g. for a column
// filter).
const TEXT_INPUT_MODE = 2

// STATUS_MODE means the commandbar is notifying the user with a message for a
// short period of time.
const STATUS_MODE = 3

// Inspector provides the necessary package
type Inspector struct {
	width  uint
	height uint
	//
	term  *termio.Terminal
	trace tr.Trace
	// Module states
	modules []ModuleState
	// Widgets
	tabs      *widget.Tabs
	table     *widget.Table
	cmdBar    *widget.TextLine
	statusBar *widget.TextLine
	statusClk uint
	// The stack of "modes" in which the inspector is operating.  The root modes
	// is the first in the stack.  When this is terminated, then the inspector
	// closes.
	modes []Mode
}

// Mode identifies a mode in which the inspector is operating.  The
// default mode is for navigating the trace, but other modes are available for
// receiving input from the user or displaying error messages, etc.
type Mode interface {
	// Activate is called when this mode becomes active.  This happens when the
	// mode is first entered, but can also happen subsequently when a child mode
	// exits and results in this mode being reactivated.
	Activate(*Inspector)
	// Clock is called on every clock tick.  This gives the mode an opportunity
	// to do something if it wishes to.
	Clock(*Inspector)
	// KeyPressed in the inspector and received by this mode.
	KeyPressed(*Inspector, uint16) bool
}

// NewInspector constructs a new inspector on given terminal.
func NewInspector(term *termio.Terminal, schema sc.Schema, trace tr.Trace, srcmap *corset.SourceMap) *Inspector {
	states := make([]ModuleState, 0)
	//
	for _, module := range srcmap.Flattern(concreteModules) {
		// only consider modules which actually have columns.
		if len(module.Columns) > 0 {
			states = append(states, newModuleState(&module, trace, srcmap.Enumerations, true))
		}
	}
	//
	tabs, table, cmdbar, statusbar := initInspectorWidgets(term, states)
	//
	inspector := &Inspector{0, 0, term, trace, states, tabs, table, cmdbar, statusbar, 0, nil}
	table.SetSource(inspector)
	// Put the inspector into default mode.
	inspector.EnterMode(&NavigationMode{})
	//
	return inspector
}

// Clock the inspector
func (p *Inspector) Clock() error {
	dirty := false
	mode := len(p.modes) - 1
	nWidth, nHeight := p.term.GetSize()
	// Pass on clock
	p.modes[mode].Clock(p)
	//
	if p.statusClk != 0 {
		p.statusClk = p.statusClk - 1
		// Clear status when clock expired
		if p.statusClk == 0 {
			p.statusClk = 0
			p.statusBar.Clear()
			// Force render
			dirty = true
		}
	}
	// Only force rerender if dimensions have changed.
	if dirty || nWidth != p.width || nHeight != p.height {
		// Update cached dimensions
		p.width, p.height = nWidth, nHeight
		// Render
		return p.Render()
	}
	//
	return nil
}

// Render the inspector to the given terminal
func (p *Inspector) Render() error {
	return p.term.Render()
}

// Close the inspector.
func (p *Inspector) Close() error {
	return p.term.Restore()
}

// CurrentModule returns the currently selected module
func (p *Inspector) CurrentModule() *ModuleState {
	module := p.tabs.Selected()
	//
	return &p.modules[module]
}

// EnterMode pushes a new mode onto the mode stack.
func (p *Inspector) EnterMode(mode Mode) {
	// Append mode to stack of active modes
	p.modes = append(p.modes, mode)
	// Activate mode
	mode.Activate(p)
}

// KeyPressed allows the inspector to react to a key being pressed by the user.
func (p *Inspector) KeyPressed(key uint16) bool {
	var n = len(p.modes) - 1
	//
	if p.modes[n].KeyPressed(p, key) {
		p.modes = p.modes[0:n]
		//
		if n > 0 {
			// Reactivate mode
			p.modes[n-1].Activate(p)
		}
	}
	// Exit when the mode stack is empty.
	return len(p.modes) == 0
}

// SetStatus puts a message on the status bar.  Messages remain visible for some
// number of clock cycles.
func (p *Inspector) SetStatus(msg termio.FormattedText) {
	p.statusBar.Clear()
	p.statusBar.Add(msg)
	p.statusClk = 5
}

// Access currently selected view
func (p *Inspector) currentView() *ModuleState {
	module := p.tabs.Selected()
	// Action change
	return &p.modules[module]
}

// Actions goto row mode
func (p *Inspector) gotoRow(row uint) termio.FormattedText {
	// Action change
	row = p.CurrentModule().setRowOffset(row)
	//
	return termio.NewColouredText(fmt.Sprintf("At row %d", row), termio.TERM_GREEN)
}

// filter columns based on a regex
func (p *Inspector) filterColumns(regex *regexp.Regexp) termio.FormattedText {
	filter := p.CurrentModule().columnFilter
	filter.Regex = regex
	p.CurrentModule().applyColumnFilter(p.trace, filter, true)
	// Success
	return termio.NewText("")
}

func (p *Inspector) clearColumnFilter() bool {
	filter := p.CurrentModule().columnFilter
	filter.Regex = nil
	p.CurrentModule().applyColumnFilter(p.trace, filter, false)
	// Success
	return true
}

func (p *Inspector) toggleColumnFilter() bool {
	var (
		filter = p.CurrentModule().columnFilter
		msg    string
	)
	// Implement toggle semantics
	switch {
	case !filter.Computed && filter.UserDefined:
		filter.Computed = true
		filter.UserDefined = false
		msg = "Showing computed columns only"
	case filter.Computed && !filter.UserDefined:
		filter.UserDefined = true
		msg = "Showing all columns"
	case filter.Computed && filter.UserDefined:
		filter.Computed = false
		msg = "Showing non-computed columns only"
	}
	//
	p.CurrentModule().applyColumnFilter(p.trace, filter, false)
	p.SetStatus(termio.NewColouredText(msg, termio.TERM_GREEN))
	// Success
	return true
}

func (p *Inspector) matchQuery(query *Query) termio.FormattedText {
	return p.CurrentModule().matchQuery(query, p.trace)
}

// ==================================================================
// TableSource
// ==================================================================

// ColumnWidth gets the width of a given column in the main table of the
// inspector.  Note that columns here are table columns, not trace columns.
func (p *Inspector) ColumnWidth(col uint) uint {
	module := p.tabs.Selected()
	state := p.modules[module]
	//
	return state.view.RowWidth(col)
}

// CellAt returns the contents of a given cell in the main table of the
// inspector.
func (p *Inspector) CellAt(col, row uint) termio.FormattedText {
	// Determine currently selected module
	module := p.tabs.Selected()
	state := &p.modules[module]
	// Get cell out of module view, noting that we are deliberately swapping row
	// and column.
	return state.view.CellAt(p.trace, row, col)
}

// Start provides a read / update / render loop.
func (p *Inspector) Start() []error {
	var errors []error
	// Start clock timer
	clk := time.NewTicker(500 * time.Millisecond)
	//
	go func() {
		for {
			// Receive clock signal
			<-clk.C
			// Force render
			//nolint:errcheck
			p.Clock()
		}
	}()
	//
	for {
		if key, err := p.term.ReadKey(); err != nil {
			errors = append(errors, err)
			break
		} else if exit := p.KeyPressed(key); exit {
			break
		}
		// Rerender window
		if err := p.Render(); err != nil {
			errors = append(errors, err)
			break
		}
	}
	// Attempt to restore terminal state
	if err := p.term.Restore(); err != nil {
		errors = append(errors, err)
	}
	// Done
	return errors
}

// ==================================================================
// Helpers
// ==================================================================

func initInspectorWidgets(term *termio.Terminal, states []ModuleState) (tabs *widget.Tabs,
	table *widget.Table, cmdbar *widget.TextLine, statusbar *widget.TextLine) {
	//
	tabs = initInspectorTabs(states)
	table = widget.NewTable(nil)
	cmdbar = widget.NewText()
	statusbar = widget.NewText()
	//
	term.Add(tabs)
	term.Add(widget.NewSeparator("⎯"))
	term.Add(table)
	term.Add(widget.NewSeparator("⎯"))
	term.Add(cmdbar)
	term.Add(statusbar)
	//
	return tabs, table, cmdbar, statusbar
}

func initInspectorTabs(states []ModuleState) *widget.Tabs {
	var titles []string
	for _, state := range states {
		titles = append(titles, state.name)
	}
	//
	return widget.NewTabs(titles...)
}

func concreteModules(m *corset.SourceModule) bool {
	return !m.Virtual
}
