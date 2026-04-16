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
package debug

import (
	"time"

	"github.com/consensys/go-corset/pkg/util/termio"
	"github.com/consensys/go-corset/pkg/util/termio/widget"
)

// Debugger packages together the necessary components for the interactive
// debugger.  Since this runs in the terminal, it must manage the necessary
// termio widgets.
type Debugger struct {
	width  uint
	height uint
	//
	term *termio.Terminal
	//
	table     *widget.Table
	cmdBar    *widget.TextLine
	statusBar *widget.TextLine
	statusClk uint
}

// NewDebugger constructs a new debugger on the given terminal.
func NewDebugger(term *termio.Terminal, view *TraceView) *Debugger {
	//
	table, cmdbar, statusbar := initDebuggerWidgets(term)
	//
	table.SetSource(view)
	//
	return &Debugger{
		0, 0, term, table, cmdbar, statusbar, 0,
	}
}

// Clock the inspector
func (p *Debugger) Clock() error {
	dirty := false
	nWidth, nHeight := p.term.GetSize()
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

// KeyPressed allows the inspector to react to a key being pressed by the user.
func (p *Debugger) KeyPressed(key uint16) bool {
	return true
}

// Render the inspector to the given terminal
func (p *Debugger) Render() error {
	return p.term.Render()
}

// Start provides a read / update / render loop.
func (p *Debugger) Start() []error {
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

func initDebuggerWidgets(term *termio.Terminal) (table *widget.Table, cmdbar *widget.TextLine,
	statusbar *widget.TextLine) {
	//
	table = widget.NewTable(nil)
	cmdbar = widget.NewText()
	statusbar = widget.NewText()
	//
	term.Add(table)
	term.Add(widget.NewSeparator("⎯"))
	term.Add(cmdbar)
	term.Add(statusbar)
	//
	return table, cmdbar, statusbar
}
