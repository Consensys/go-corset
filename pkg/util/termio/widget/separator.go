package widget

import (
	"github.com/consensys/go-corset/pkg/util/termio"
)

// Separator is intended to be something like a horizontal rule, where the
// separator character can be specified.
type Separator struct {
	separator string
}

// NewSeparator constructs a new separator with a given separator character.
func NewSeparator(separator string) termio.Widget {
	return &Separator{separator}
}

// GetHeight of this widget, where MaxUint indicates widget expands to take as
// much as it can.
func (p *Separator) GetHeight() uint {
	return 1
}

// Render this widget on the given canvas.
func (p *Separator) Render(canvas termio.Canvas) {
	w, _ := canvas.GetDimensions()
	//
	for i := uint(0); i < w; i++ {
		canvas.Write(i, 0, p.separator, nil)
	}
}
