package termio

import (
	"errors"
	"fmt"
	"io"
	"math"
	"os"
	"slices"

	"golang.org/x/term"
)

// ESC is the escape code.
const ESC uint16 = 0x1b

// TAB indicates the horizontal tab
const TAB uint16 = 0x09

// CARRIAGE_RETURN indicates "enter"
const CARRIAGE_RETURN uint16 = 0x0D

// BACKSPACE is the backspace
const BACKSPACE uint16 = 0x08

// DEL is the delete key
const DEL uint16 = 0x7f

// BACKTAB indicates shift + tab
const BACKTAB uint16 = 0x5b5a

// CURSOR_UP (up arrow)
const CURSOR_UP uint16 = 0x5b41

// CURSOR_DOWN (down arrow)
const CURSOR_DOWN uint16 = 0x5b42

// CURSOR_LEFT (left arrow)
const CURSOR_LEFT uint16 = 0x5b43

// CURSOR_RIGHT (left arrow)
const CURSOR_RIGHT uint16 = 0x5b44

// UNKNOWN is a fall-back for unknown escape sequences
const UNKNOWN uint16 = 0x5bff

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

// ReadKey returns a keyevent from the keyboard.  This is either an ASCII
// character, or an extended escape code.
func (t *Terminal) ReadKey() (uint16, error) {
	var key [3]byte
	//
	if n, err := os.Stdin.Read(key[:]); err != nil {
		return 0, err
	} else if n == 1 {
		return uint16(key[0]), nil
	} else if n != 3 || key[1] != '[' {
		// Unknown or malformed escape sequence.
		return UNKNOWN, nil
	}
	// Dispatch escape
	switch key[2] {
	case 'A':
		return CURSOR_UP, nil
	case 'B':
		return CURSOR_DOWN, nil
	case 'C':
		return CURSOR_RIGHT, nil
	case 'D':
		return CURSOR_LEFT, nil
	case 'Z':
		return BACKTAB, nil
	}
	//
	fmt.Printf("IGNORED %d\n", key[0])
	// unknown key
	return UNKNOWN, nil
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

func (p *terminalCanvas) Write(x, y uint, text FormattedText) {
	// Determine dimensions
	w, h := p.GetDimensions()
	// Clip chunk if necessary
	if x < w && y < h {
		mx := x + text.Len()
		if mx > w {
			text.Clip(0, w-x)
		}
		//
		p.lines[y] = append(p.lines[y], terminalChunk{x, text})
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
			c_text.Clip(diff, math.MaxUint)
		}
		// Construct
		bytes = append(bytes, c_text.Bytes()...)
		// Advance cursor
		xpos += c_text.Len()
	}
	// fill to end of line
	for ; xpos < p.width; xpos++ {
		bytes = append(bytes, ' ')
	}
	//
	return bytes
}

type terminalChunk struct {
	xpos uint
	text FormattedText
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
