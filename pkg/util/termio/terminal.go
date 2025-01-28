package termio

import (
	"errors"
	"io"
	"math"
	"os"
	"slices"

	"golang.org/x/term"
)

// Terminal provides a simple top-level window.
type Terminal struct {
	// file descriptor for output.
	fd int
	// Underlying terminal
	xterm *term.Terminal
	// Stores original state of terminal so this can be restored.
	state *term.State
	// List of widgets to display
	widgets []Widget
}

// NewTerminal constructs a new terminal.
func NewTerminal() (*Terminal, error) {
	fd := int(os.Stdout.Fd())
	//
	if !term.IsTerminal(fd) {
		return nil, errors.New("invalid terminal")
	}
	// Move terminal into raw mode
	state, err := term.MakeRaw(0)
	if err != nil {
		return nil, err
	}
	// Construct "screen"
	screen := struct {
		io.Reader
		io.Writer
	}{os.Stdin, os.Stdout}
	// Grab terminal screen
	terminal := term.NewTerminal(screen, "")
	//
	return &Terminal{fd, terminal, state, nil}, nil
}

// GetSize returns the dimensions of the terminal.
func (t *Terminal) GetSize() (uint, uint) {
	w, h, err := term.GetSize(t.fd)
	// Sanity check for now
	if err != nil {
		panic(err)
	}
	//
	return uint(w), uint(h)
}

// Add a new widget to this window.  Widgets will be laid out vertically in
// the order they are added.
func (t *Terminal) Add(w Widget) {
	t.widgets = append(t.widgets, w)
}

// Render this window to the terminal.
func (t *Terminal) Render() error {
	var (
		taken uint
		nFlex uint
		flex  uint
	)
	// Determine dimensions
	width, height := t.GetSize()
	//
	for _, w := range t.widgets {
		if h := w.GetHeight(); h != math.MaxUint {
			taken += h
		} else {
			nFlex++
		}
	}
	// Determine flexible amount
	if nFlex > 0 {
		flex = (height - taken) / nFlex
	}
	// Reset taken
	taken = 0
	//
	for _, w := range t.widgets {
		var h uint
		//
		if h = w.GetHeight(); h == math.MaxUint {
			h = flex
		}
		// Construct canvas
		canvas := newTerminalCanvas(width, h)
		// Record how much taken
		taken += h
		// Render widget
		w.Render(canvas)
		// Render canvas
		if err := t.renderCanvas(canvas); err != nil {
			return err
		}
	}
	// Check whether any left
	if taken < height {
		blank := blankLine(width)
		// Fill out remainder with blanks
		for ; taken < height; taken++ {
			if _, err := t.xterm.Write(blank); err != nil {
				return err
			}
		}
	}
	//
	return nil
}

// Restore terminal to its original state.
func (t *Terminal) Restore() error {
	return term.Restore(t.fd, t.state)
}

func (t *Terminal) renderCanvas(canvas *terminalCanvas) error {
	for i := uint(0); i < uint(len(canvas.lines)); i++ {
		// Render the line
		line := canvas.renderLine(i)
		// Write the line
		if _, err := t.xterm.Write(line); err != nil {
			return err
		}
	}
	//
	return nil
}

// TerminalCanvas provides a Canvas which collects chunks to be written when
// rendering a given widget.
type terminalCanvas struct {
	width uint
	lines [][]terminalChunk
}

func newTerminalCanvas(width, height uint) *terminalCanvas {
	return &terminalCanvas{width, make([][]terminalChunk, height)}
}

func (p *terminalCanvas) GetDimensions() (uint, uint) {
	return p.width, uint(len(p.lines))
}

func (p *terminalCanvas) Write(x, y uint, str string, format *AnsiEscape) {
	text := []rune(str)
	// Determine dimensions
	w, h := p.GetDimensions()
	// Clip chunk if necessary
	if x < w && y < h {
		mx := (x + uint(len(text)))
		if mx > w {
			text = text[:w-x]
		}
		//
		p.lines[y] = append(p.lines[y], terminalChunk{x, format, text})
	}
}

func (p *terminalCanvas) renderLine(line uint) []byte {
	var (
		xpos   uint   = 0
		chunks        = p.lines[line]
		bytes  []byte = nil
	)
	// Sort by decreasing x value.
	slices.SortFunc(chunks, func(l terminalChunk, r terminalChunk) int {
		if l.xpos < r.xpos {
			return -1
		} else if l.xpos > r.xpos {
			return 1
		}
		//
		return 0
	})
	// Process each chunk in turn
	for _, c := range chunks {
		c_text := c.text
		// fill upto mark
		for ; xpos < c.xpos; xpos++ {
			bytes = append(bytes, ' ')
		}
		// Clip chunk if it overlaps
		if c.xpos < xpos {
			diff := xpos - c.xpos
			clip := min(len(c_text), int(diff))
			c_text = c_text[clip:]
		}
		// Construct string
		c_str := string(c_text)
		// Append bytes
		if c.escape != nil {
			// Apply formatting
			escape := c.escape.Build()
			bytes = append(bytes, []byte(escape)...)
			// Add content
			bytes = append(bytes, []byte(c_str)...)
			// Reset formatting
			escape = ResetAnsiEscape().Build()
			bytes = append(bytes, []byte(escape)...)
		} else {
			bytes = append(bytes, []byte(c_str)...)
		}
		// Advance cursor
		xpos += uint(len(c_text))
	}
	// fill to end of line
	for ; xpos < p.width; xpos++ {
		bytes = append(bytes, ' ')
	}
	//
	return bytes
}

type terminalChunk struct {
	xpos   uint
	escape *AnsiEscape
	text   []rune
}

// Construct a line full of blanks.
func blankLine(width uint) []byte {
	bytes := make([]byte, width)
	//
	for i := range bytes {
		bytes[i] = ' '
	}
	//
	return bytes
}
