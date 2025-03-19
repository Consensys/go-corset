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
package inspector

import (
	"math/big"
	"regexp"
	"strconv"

	"github.com/consensys/go-corset/pkg/util"
	"github.com/consensys/go-corset/pkg/util/source/bexp"
	"github.com/consensys/go-corset/pkg/util/termio"
)

// InputMode is where the user is entering some information (e.g. row for
// executing a goto command).
type InputMode[T any] struct {
	// prompt to show user
	prompt termio.FormattedText
	// input text being accumulated whilst in input mode.
	input []byte
	// current cursor position
	cursor uint
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
	return &InputMode[T]{prompt, input, 0, history, index, handler}
}

// Activate navigation mode by setting the command bar to show the navigation
// commands.
func (p *InputMode[T]) Activate(parent *Inspector) {
	parent.cmdBar.Clear()
	parent.cmdBar.Add(p.prompt)
	// Add current filter
	colour := termio.TERM_GREEN
	invColour := termio.TERM_BLACK
	input := string(p.input)
	//
	if _, ok := p.handler.Convert(input); !ok {
		colour = termio.TERM_RED
	}
	// construct cursor escape code
	escape := termio.NewAnsiEscape().FgColour(invColour).BgColour(termio.TERM_YELLOW)
	// handle cursor
	if p.cursor < uint(len(p.input)) {
		// cursor behind text
		parent.cmdBar.Add(termio.NewColouredText(input[:p.cursor], colour))
		parent.cmdBar.Add(termio.NewFormattedText(input[p.cursor:p.cursor+1], escape))
		parent.cmdBar.Add(termio.NewColouredText(input[p.cursor+1:], colour))
	} else {
		// cursor leading text
		parent.cmdBar.Add(termio.NewColouredText(input, colour))
		parent.cmdBar.Add(termio.NewFormattedText(" ", escape))
	}
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
		p.deleteCharacterAtCursor()
	case key == termio.CARRIAGE_RETURN:
		input := string(p.input)
		// Attempt conversion
		if val, ok := p.handler.Convert(input); ok {
			// Looks good, to fire the value
			p.handler.Apply(val)
		}
		// Success
		return true
	case key == termio.CURSOR_LEFT:
		if p.cursor > 0 {
			p.cursor--
		}
	case key == termio.CURSOR_RIGHT:
		if p.cursor < uint(len(p.input)) {
			p.cursor++
		}
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
		p.insertCharacterAtCursor(byte(key))
	}
	// Update displayed text
	p.Activate(parent)
	//
	return false
}

// Delete character at cursor position
func (p *InputMode[T]) deleteCharacterAtCursor() {
	if p.cursor > 0 {
		p.cursor--
		p.input = util.RemoveAt(p.input, p.cursor)
	}
}

// Insert character at cursor position
func (p *InputMode[T]) insertCharacterAtCursor(char byte) {
	p.input = util.InsertAt(p.input, char, p.cursor)
	// advance cursor
	p.cursor++
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

// ==================================================================
// Proposition (i.e. Boolean Expression) Handler
// ==================================================================

type queryHandler struct {
	callback func(*Query) bool
}

func newQueryHandler(callback func(*Query) bool) InputHandler[*Query] {
	return &queryHandler{callback}
}

func (p *queryHandler) Convert(input string) (*Query, bool) {
	if prop, errs := bexp.Parse[*Query](input); len(errs) == 0 {
		return prop, true
	}

	return nil, false
}

func (p *queryHandler) Apply(query *Query) {
	p.callback(query)
}

// Query represents a boolean expression which can be evaluated over a
// given set of columns.
type Query struct {
	//
}

// Variable constructs a variable of the given name.
func (p *Query) Variable(name string) *Query {
	return &Query{}
}

// Number constructs a number with the given value
func (p *Query) Number(number big.Int) *Query {
	return &Query{}
}

// Or constructs a disjunction of the given proposition.
func (p *Query) Or(queries ...*Query) *Query {
	panic("todo")
}

// Equals constructs an equality between two queries.
func (p *Query) Equals(rhs *Query) *Query {
	return &Query{}
}

// NotEquals constructs a non-equality between two queries.
func (p *Query) NotEquals(rhs *Query) *Query {
	return &Query{}
}
