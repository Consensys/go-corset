package inspector

import (
	"regexp"
	"strconv"

	"github.com/consensys/go-corset/pkg/util/termio"
)

// InputMode is where the user is entering some information (e.g. row for
// executing a goto command).
type InputMode[T any] struct {
	// prompt to show user
	prompt termio.FormattedText
	// input text being accumulated whilst in input mode.
	input []byte
	// history of options for this input mode.
	history []string
	// history index
	index uint
	// parser responsible for checking whether input is valid (or not).
	handler InputHandler[T]
}

// InputHandler provides a generic way of handling input, including a mechanism
// for checking that input is well formed.
type InputHandler[T any] interface {
	// Convert attempts to convert the input string into a valid value.
	Convert(string) (T, bool)
	// Apply the given input, which will activate some kind of callback.
	Apply(T)
}

func newInputMode[T any](prompt termio.FormattedText, index uint, history []string,
	handler InputHandler[T]) *InputMode[T] {
	var input []byte
	// Determine whether to show item from history
	if index >= uint(len(history)) {
		input = []byte{}
	} else {
		input = []byte(history[index])
	}
	// Done
	return &InputMode[T]{prompt, input, history, index, handler}
}

// Activate navigation mode by setting the command bar to show the navigation
// commands.
func (p *InputMode[T]) Activate(parent *Inspector) {
	parent.cmdBar.Clear()
	parent.cmdBar.Add(p.prompt)
	// Add current filter
	colour := termio.TERM_GREEN
	input := string(p.input)
	//
	if _, ok := p.handler.Convert(input); !ok {
		colour = termio.TERM_RED
	}
	//
	parent.cmdBar.Add(termio.NewColouredText(input, colour))
}

// Clock navitation mode, which does nothing at this time.
func (p *InputMode[T]) Clock(parent *Inspector) {
	// Nothing to do.
}

// KeyPressed in input mode simply updates the input, or exits the mode if
// either "ESC" or enter are pressed.
func (p *InputMode[T]) KeyPressed(parent *Inspector, key uint16) bool {
	switch {
	case key == termio.ESC:
		return true
	case key == termio.BACKSPACE || key == termio.DEL:
		if len(p.input) > 0 {
			n := len(p.input) - 1
			p.input = p.input[0:n]
		}
	case key == termio.CARRIAGE_RETURN:
		input := string(p.input)
		// Attempt conversion
		if val, ok := p.handler.Convert(input); ok {
			// Looks good, to fire the value
			p.handler.Apply(val)
		}
		// Success
		return true
	case key == termio.CURSOR_UP:
		if p.index > 0 {
			p.index--
			p.input = []byte(p.history[p.index])
		}
	case key == termio.CURSOR_DOWN:
		p.index++
		// Check for end-of-history
		if p.index >= uint(len(p.history)) {
			p.index = uint(len(p.history))
			p.input = []byte{}
		} else {
			p.input = []byte(p.history[p.index])
		}
	case key >= 32 && key <= 126:
		p.input = append(p.input, byte(key))
	}
	// Update displayed text
	p.Activate(parent)
	//
	return false
}

// ==================================================================
// UintHandler
// ==================================================================

type uintHandler struct {
	callback func(uint) bool
}

func newUintHandler(callback func(uint) bool) InputHandler[uint] {
	return &uintHandler{callback}
}

func (p *uintHandler) Convert(input string) (uint, bool) {
	val, err := strconv.Atoi(input)
	//
	if val < 0 || err != nil {
		return 0, false
	}
	//
	return uint(val), true
}

func (p *uintHandler) Apply(value uint) {
	p.callback(value)
}

// ==================================================================
// RegexHandler
// ==================================================================

type regexHandler struct {
	callback func(*regexp.Regexp) bool
}

func newRegexHandler(callback func(*regexp.Regexp) bool) InputHandler[*regexp.Regexp] {
	return &regexHandler{callback}
}

func (p *regexHandler) Convert(input string) (*regexp.Regexp, bool) {
	if regex, err := regexp.Compile(input); err == nil {
		return regex, true
	}

	return nil, false
}

func (p *regexHandler) Apply(regex *regexp.Regexp) {
	p.callback(regex)
}
