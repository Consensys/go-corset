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
package widget

import "github.com/consensys/go-corset/pkg/util/termio"

// Tabs is a simple widget which shows a bunch of titles in a bar, and
// highlights a selected one.
type Tabs struct {
	tabs     []string
	selected uint
}

// NewTabs constructs a new tabs widget with the given titles.
func NewTabs(tabs ...string) *Tabs {
	return &Tabs{tabs, 0}
}

// Selected returns the currently selected tab.
func (p *Tabs) Selected() uint {
	return p.selected
}

// Select sets the given selected tab.  If the index is greater than the
// available tabs, then it automatically "wraps around".
func (p *Tabs) Select(tab uint) {
	p.selected = tab % uint(len(p.tabs))
}

// Render the tabs widget to a given canvas.
func (p *Tabs) Render(canvas termio.Canvas) {
	w, _ := canvas.GetDimensions()
	//
	x := uint(1)
	//
	for i := 0; i < len(p.tabs) && x < w; i++ {
		if i != 0 {
			// Write out separator
			canvas.Write(x, 0, termio.NewText(" | "))
			x += 3
		}
		// Extract title
		cell := termio.NewText(p.tabs[i])
		// Check for selected
		if uint(i) == p.selected {
			cell.Format(termio.UnderlineAnsiEscape())
		}
		// Write out title
		canvas.Write(x, 0, cell)
		x += cell.Len()
	}
}

// GetHeight of this widget, where MaxUint indicates widget expands to take as
// much as it can.
func (p *Tabs) GetHeight() uint {
	return 1
}
