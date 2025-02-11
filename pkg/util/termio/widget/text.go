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

// TextLine displays a line of formatted text.
type TextLine struct {
	// Contents of this line
	contents []termio.FormattedText
}

// NewText constructs a new text widget which is initially empty.
func NewText() *TextLine {
	return &TextLine{nil}
}

// GetHeight of this widget, where MaxUint indicates widget expands to take as
// much as it can.
func (p *TextLine) GetHeight() uint {
	return 1
}

// Clear contents of this text line.
func (p *TextLine) Clear() {
	p.contents = nil
}

// Add a new chunk of formatted text.
func (p *TextLine) Add(txt termio.FormattedText) {
	p.contents = append(p.contents, txt)
}

// Render the tabs widget to a given canvas.
func (p *TextLine) Render(canvas termio.Canvas) {
	xpos := uint(0)

	for _, txt := range p.contents {
		canvas.Write(xpos, 0, txt)
		xpos += txt.Len() + 0
	}
}
