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
	"github.com/consensys/go-corset/pkg/util/collection/set"
	"github.com/consensys/go-corset/pkg/util/field"
	"github.com/consensys/go-corset/pkg/util/termio"
)

// NavigationMode is the default mode of the inspector.  In this mode, the user
// is navigating the trace in the normal fashion.
type NavigationMode[F field.Element[F]] struct {
}

// Activate navigation mode by setting the command bar to show the navigation
// commands.
func (p *NavigationMode[F]) Activate(parent *Inspector[F]) {
	parent.cmdBar.Clear()
	parent.cmdBar.AddLeft(termio.NewColouredText("[g]", termio.TERM_YELLOW))
	parent.cmdBar.AddLeft(termio.NewText("oto :: "))
	parent.cmdBar.AddLeft(termio.NewColouredText("[f]", termio.TERM_YELLOW))
	parent.cmdBar.AddLeft(termio.NewText("ilter :: "))
	parent.cmdBar.AddLeft(termio.NewColouredText("[#]", termio.TERM_YELLOW))
	parent.cmdBar.AddLeft(termio.NewText("clear filter :: "))
	parent.cmdBar.AddLeft(termio.NewColouredText("[t]", termio.TERM_YELLOW))
	parent.cmdBar.AddLeft(termio.NewText("oggle computed :: "))
	parent.cmdBar.AddLeft(termio.NewColouredText("[s]", termio.TERM_YELLOW))
	parent.cmdBar.AddLeft(termio.NewText("can :: "))
	parent.cmdBar.AddLeft(termio.NewColouredText("[n]", termio.TERM_YELLOW))
	parent.cmdBar.AddLeft(termio.NewText("ext match :: "))
	parent.cmdBar.AddLeft(termio.NewColouredText("[p]", termio.TERM_YELLOW))
	parent.cmdBar.AddLeft(termio.NewText("revious match :: "))
	//p.cmdbar.Add(termio.NewFormattedText("[p]erspectives"))
	parent.cmdBar.AddLeft(termio.NewColouredText("[q]", termio.TERM_RED))
	parent.cmdBar.AddLeft(termio.NewText("uit"))
	parent.cmdBar.AddRight(termio.NewText(" "))
	parent.cmdBar.AddRight(termio.NewColouredText("[-]", termio.TERM_YELLOW))
	parent.cmdBar.AddRight(termio.NewText(" / "))
	parent.cmdBar.AddRight(termio.NewColouredText("[+]", termio.TERM_YELLOW))
	parent.cmdBar.AddRight(termio.NewText("cell width "))
}

// Clock navitation mode, which does nothing at this time.
func (p *NavigationMode[F]) Clock(parent *Inspector[F]) {

}

// KeyPressed in navigation mode, which either adjusts our view of the trace
// table or fires off some command.
func (p *NavigationMode[F]) KeyPressed(parent *Inspector[F], key uint16) bool {
	module := parent.tabs.Selected()
	//
	switch key {
	case termio.TAB:
		parent.tabs.Select(int(module) + 1)
	case termio.BACKTAB:
		parent.tabs.Select(int(module) - 1)
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
	case termio.SCROLL_UP:
		n := parent.height / 2
		col := parent.modules[module].view.col
		parent.modules[module].setColumnOffset(col - n)
	case termio.SCROLL_DOWN:
		n := parent.height / 2
		col := parent.modules[module].view.col
		parent.modules[module].setColumnOffset(col + n)
	// quit
	case 'q':
		return true
	// goto command
	case 'g':
		parent.EnterMode(p.gotoInputMode(parent))
	case 'f':
		parent.EnterMode(p.filterInputMode(parent))
	case 't':
		parent.toggleColumnFilter()
	case 's':
		parent.EnterMode(p.scanInputMode(parent))
	case 'n':
		parent.nextScanResult(true)
	case 'p':
		parent.nextScanResult(false)
	case '#':
		parent.clearColumnFilter()
	case '+':
		parent.changeCellWidth(true)
	case '-':
		parent.changeCellWidth(false)
	}
	//
	return false
}

func (p *NavigationMode[F]) gotoInputMode(parent *Inspector[F]) Mode[F] {
	prompt := termio.NewColouredText("[history ↑/↓] row? ", termio.TERM_YELLOW)
	history := parent.currentView().targetRowHistory
	history_index := uint(len(history))
	//
	return newInputMode[F](prompt, history_index, history, newUintHandler(parent.gotoRow))
}

func (p *NavigationMode[F]) filterInputMode(parent *Inspector[F]) Mode[F] {
	prompt := termio.NewColouredText("[history ↑/↓] regex? ", termio.TERM_YELLOW)
	// Determine current active filter
	filter := parent.currentView().columnFilter
	history := parent.currentView().columnFilterHistory
	history_index := uint(len(history))
	//
	if filter.Regex != nil {
		history_index--
	}
	//
	return newInputMode[F](prompt, history_index, history, newRegexHandler(parent.filterColumns))
}

func (p *NavigationMode[F]) scanInputMode(parent *Inspector[F]) Mode[F] {
	var (
		prompt        = termio.NewColouredText("[history ↑/↓] where $=row, expression? ", termio.TERM_YELLOW)
		history       = parent.currentView().scanHistory
		history_index = uint(len(history))
		columns       set.SortedSet[string]
	)
	// Identify available columns
	for _, c := range parent.CurrentModule().columns {
		columns.Insert(c.Name)
	}
	// Construct environment
	env := func(col string) bool {
		return columns.Contains(col) || col == "$"
	}
	// Construct input mode
	return newInputMode[F](prompt, history_index, history, newQueryHandler(env, parent.matchQuery))
}
