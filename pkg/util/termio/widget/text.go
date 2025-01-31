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
