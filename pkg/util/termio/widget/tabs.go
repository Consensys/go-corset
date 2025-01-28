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

// Render the tabs widget to a given canvas.
func (p *Tabs) Render(canvas termio.Canvas) {
	w, _ := canvas.GetDimensions()
	//
	x := uint(1)
	//
	for i := 0; i < len(p.tabs) && x < w; i++ {
		var escape *termio.AnsiEscape = nil
		//
		if i != 0 {
			// Write out separator
			canvas.Write(x, 0, " | ", nil)
			x += 3
		}
		// Check for selected
		if uint(i) == p.selected {
			var esc termio.AnsiEscape = termio.UnderlineAnsiEscape()
			escape = &esc
		}
		// Extract title
		title := p.tabs[i]
		// Write out title
		canvas.Write(x, 0, title, escape)
		x += uint(len(title))
	}
}

// GetHeight of this widget, where MaxUint indicates widget expands to take as
// much as it can.
func (p *Tabs) GetHeight() uint {
	return 1
}
