package termio

import "fmt"

// TERM_BLACK represents black
const TERM_BLACK = uint(0)

// TERM_RED represents red
const TERM_RED = uint(1)

// TERM_GREEN represents green
const TERM_GREEN = uint(2)

// TERM_YELLOW represents yellow
const TERM_YELLOW = uint(3)

// TERM_BLUE represents blue
const TERM_BLUE = uint(4)

// TERM_MAGENTA represents magenta
const TERM_MAGENTA = uint(5)

// TERM_CYAN represents cyan
const TERM_CYAN = uint(6)

// TERM_WHITE represents white
const TERM_WHITE = uint(7)

// AnsiEscape represents an ANSI escape code used for formatting text in a terminal.
type AnsiEscape struct {
	escape string
	count  uint
}

// NewAnsiEscape construct an empty escape
func NewAnsiEscape() AnsiEscape {
	return AnsiEscape{"\033", 0}
}

// ResetAnsiEscape constructs a reset term.
func ResetAnsiEscape() AnsiEscape {
	return AnsiEscape{"\033[0", 1}
}

// BoldAnsiEscape constructs a reset term.
func BoldAnsiEscape() AnsiEscape {
	return AnsiEscape{"\033[1", 1}
}

// UnderlineAnsiEscape constructs a reset term.
func UnderlineAnsiEscape() AnsiEscape {
	return AnsiEscape{"\033[4", 1}
}

// FgColour sets the foreground colour
func (p AnsiEscape) FgColour(col uint) AnsiEscape {
	col += 30
	// Construct string
	var escape string
	if p.count > 0 {
		escape = fmt.Sprintf("%s;%d", p.escape, col)
	} else {
		escape = fmt.Sprintf("%s[%d", p.escape, col)
	}
	// Done
	return AnsiEscape{escape, p.count + 1}
}

// BgColour sets the foreground colour
func (p AnsiEscape) BgColour(col uint) AnsiEscape {
	col += 40
	// Construct string
	var escape string
	if p.count > 0 {
		escape = fmt.Sprintf("%s;%d", p.escape, col)
	} else {
		escape = fmt.Sprintf("%s[%d", p.escape, col)
	}
	// Done
	return AnsiEscape{escape, p.count + 1}
}

// Build constructs the final escape
func (p AnsiEscape) Build() string {
	return fmt.Sprintf("%sm", p.escape)
}
