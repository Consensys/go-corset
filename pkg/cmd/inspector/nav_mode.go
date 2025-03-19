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

import "github.com/consensys/go-corset/pkg/util/termio"

// NavigationMode is the default mode of the inspector.  In this mode, the user
// is navigating the trace in the normal fashion.
type NavigationMode struct {
}

// Activate navigation mode by setting the command bar to show the navigation
// commands.
func (p *NavigationMode) Activate(parent *Inspector) {
	parent.cmdBar.Clear()
	parent.cmdBar.Add(termio.NewColouredText("[g]", termio.TERM_YELLOW))
	parent.cmdBar.Add(termio.NewText("oto :: "))
	parent.cmdBar.Add(termio.NewColouredText("[f]", termio.TERM_YELLOW))
	parent.cmdBar.Add(termio.NewText("ilter :: "))
	parent.cmdBar.Add(termio.NewColouredText("[#]", termio.TERM_YELLOW))
	parent.cmdBar.Add(termio.NewText("clear filter :: "))
	parent.cmdBar.Add(termio.NewColouredText("[s]", termio.TERM_YELLOW))
	parent.cmdBar.Add(termio.NewText("scan :: "))
	//p.cmdbar.Add(termio.NewFormattedText("[p]erspectives"))
	parent.cmdBar.Add(termio.NewColouredText("[q]", termio.TERM_RED))
	parent.cmdBar.Add(termio.NewText("uit"))
	//
	//parent.statusBar.Clear()
}

// Clock navitation mode, which does nothing at this time.
func (p *NavigationMode) Clock(parent *Inspector) {

}

// KeyPressed in navigation mode, which either adjusts our view of the trace
// table or fires off some command.
func (p *NavigationMode) KeyPressed(parent *Inspector, key uint16) bool {
	module := parent.tabs.Selected()
	//
	switch key {
	case termio.TAB:
		parent.tabs.Select(module + 1)
	case termio.BACKTAB:
		parent.tabs.Select(module - 1)
	case termio.CURSOR_UP:
		col := parent.modules[module].view.col
		parent.modules[module].setColumnOffset(col - 1)
	case termio.CURSOR_DOWN:
		col := parent.modules[module].view.col
		parent.modules[module].setColumnOffset(col + 1)
	case termio.CURSOR_LEFT:
		row := parent.modules[module].view.row
		parent.modules[module].setRowOffset(row - 1)
	case termio.CURSOR_RIGHT:
		row := parent.modules[module].view.row
		parent.modules[module].setRowOffset(row + 1)
	// quit
	case 'q':
		return true
	// goto command
	case 'g':
		parent.EnterMode(p.gotoInputMode(parent))
	case 'f':
		parent.EnterMode(p.filterInputMode(parent))
	case 's':
		parent.EnterMode(p.scanInputMode(parent))
	case '#':
		parent.clearColumnFilter()
	}
	//
	return false
}

func (p *NavigationMode) gotoInputMode(parent *Inspector) Mode {
	prompt := termio.NewColouredText("[history ↑/↓] row? ", termio.TERM_YELLOW)
	history := parent.currentView().targetRowHistory
	history_index := uint(len(history))
	//
	return newInputMode(prompt, history_index, history, newUintHandler(parent.gotoRow))
}

func (p *NavigationMode) filterInputMode(parent *Inspector) Mode {
	prompt := termio.NewColouredText("[history ↑/↓] regex? ", termio.TERM_YELLOW)
	// Determine current active filter
	filter := parent.currentView().columnFilter
	history := parent.currentView().columnFilterHistory
	history_index := uint(len(history))
	//
	if filter != "" {
		history_index--
	}
	//
	return newInputMode(prompt, history_index, history, newRegexHandler(parent.filterColumns))
}

func (p *NavigationMode) scanInputMode(parent *Inspector) Mode {
	prompt := termio.NewColouredText("[history ↑/↓] expression? ", termio.TERM_YELLOW)
	history := parent.currentView().scanHistory
	history_index := uint(len(history))
	//
	return newInputMode(prompt, history_index, history, newQueryHandler(parent.matchQuery))
}
