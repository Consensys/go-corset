package termio

// Window provides an abstraction over an interactive terminal session.  This
// is fairly simplistic at this stage, and supports layout of widges in a
// vertical direction only.
type Window interface {
	// Add a new widget to this window.  Widgets will be laid out vertically in
	// the order they are added.
	Add(w Widget)
	// Render this window to the terminal.
	Render() error
}

// Canvas represents a surface on which widgets can draw.
type Canvas interface {
	// Get the dimensions of this canvas.
	GetDimensions() (uint, uint)
	// Write a chunk to the canvas.
	Write(x, y uint, text string, format *AnsiEscape)
}

// Widget is an abstract entity which can be displayed upon a terminal window.
type Widget interface {
	// Get height of this widget, where MaxUint indicates widget expands to take
	// as much as it can.
	GetHeight() uint
	// Render this widget on the given canvas.
	Render(canvas Canvas)
}
