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
package bexp

import (
	"math/big"

	"github.com/consensys/go-corset/pkg/util"
	"github.com/consensys/go-corset/pkg/util/source"
)

// Parse a given input string into logical proposition.
func Parse[T Term[T]](input string) (T, []source.SyntaxError) {
	var (
		empty   T
		srcfile = source.NewSourceFile("expr", []byte(input))
		lexer   = source.NewLexer[rune](srcfile.Contents(), scanner)
		// Lex as many tokens as possible
		tokens = lexer.Collect()
	)
	// Check whether anything was left (if so this is an error)
	if lexer.Remaining() != 0 {
		start, end := lexer.Index(), lexer.Index()+lexer.Remaining()
		err := srcfile.SyntaxError(source.NewSpan(int(start), int(end)), "unknown text encountered")

		return empty, []source.SyntaxError{*err}
	}
	// Remove any whitespace
	tokens = util.RemoveMatching(tokens, func(t source.Token) bool { return t.Kind == WHITESPACE })
	//
	parser := &Parser[T]{srcfile, tokens, 0}
	// Parse term
	p, errs := parser.parseTerm()
	// Check all parsed
	if len(errs) == 0 && !parser.Done() {
		return empty, parser.syntaxErrors(parser.lookahead(), "unknown token")
	}
	// All good!
	return p, errs
}

// END_OF signals "end of file"
const END_OF uint = 0

// WHITESPACE signals whitespace
const WHITESPACE uint = 1

// LBRACE signals "left brace"
const LBRACE uint = 2

// RBRACE signals "right brace"
const RBRACE uint = 3

// NUMBER signals an integer number
const NUMBER uint = 4

// IDENTIFIER signals a column variable.
const IDENTIFIER uint = 5

// EQUALS signals an equality
const EQUALS uint = 6

// NOT_EQUALS signals a non-equality
const NOT_EQUALS uint = 7

var scanner source.Scanner[rune] = source.Or(
	source.One(LBRACE, '('),
	source.One(RBRACE, ')'),
	source.Many(WHITESPACE, ' ', '\t'),
	source.ManyWith(NUMBER, '0', '9'),
	source.ManyWith(IDENTIFIER, 'a', 'z'),
	source.Eof[rune](END_OF))

// Parser provides a general-purpose parser for propositions and arithmetic
// expressions.
type Parser[T Term[T]] struct {
	srcfile *source.File
	tokens  []source.Token
	// Position within the tokens
	index int
}

// Done determines whether or not the parser has parsed all the available
// tokens.
func (p *Parser[T]) Done() bool {
	return p.index+1 >= len(p.tokens)
}

func (p *Parser[T]) parseTerm() (T, []source.SyntaxError) {
	term, errs := p.parseUnitTerm()
	// check for infix expression
	if len(errs) == 0 && !p.Done() {
		var token = p.lookahead()
		//
		switch token.Kind {
		case EQUALS, NOT_EQUALS:
			return p.parseEquality(token.Kind, term)
		default:
			var empty T
			return empty, p.syntaxErrors(token, "unknown expression")
		}
	}
	//
	return term, errs
}

func (p *Parser[T]) parseEquality(kind uint, lhs T) (T, []source.SyntaxError) {
	p.expect(kind)
	//
	rhs, errs := p.parseUnitTerm()
	//
	if len(errs) == 0 && kind == EQUALS {
		return lhs.Equals(rhs), nil
	} else if len(errs) == 0 {
		return lhs.NotEquals(rhs), nil
	}
	//
	return rhs, errs
}

// ParseTerm parses an expression.
func (p *Parser[T]) parseUnitTerm() (T, []source.SyntaxError) {
	var (
		empty T
		token = p.lookahead()
	)
	// Otherwise, assume connective
	switch token.Kind {
	case LBRACE:
		return p.parseBracketedTerm()
	case IDENTIFIER:
		return p.parseIdentifier(), nil
	case NUMBER:
		return p.parseNumber(), nil
	}
	//
	return empty, p.syntaxErrors(token, "unknown expression")
}

func (p *Parser[T]) parseBracketedTerm() (T, []source.SyntaxError) {
	var empty T
	//
	p.expect(LBRACE)
	//
	term, errs := p.parseTerm()
	//
	if len(errs) == 0 && !p.match(RBRACE) {
		return empty, p.syntaxErrors(p.lookahead(), "expected ')'")
	}
	//
	return term, errs
}

func (p *Parser[T]) parseIdentifier() T {
	var variable T
	//
	id := p.expect(IDENTIFIER)
	//
	return variable.Variable(p.string(id))
}

func (p *Parser[T]) parseNumber() T {
	var num T
	//
	id := p.expect(NUMBER)
	//
	return num.Number(p.number(id))
}

// Get the text representing the given token as a string.
func (p *Parser[T]) string(token source.Token) string {
	start, end := token.Span.Start(), token.Span.End()
	return string(p.srcfile.Contents()[start:end])
}

// Get the text representing the given token as a string.
func (p *Parser[T]) number(token source.Token) big.Int {
	var number big.Int
	//
	number.SetString(p.string(token), 0)
	//
	return number
}

func (p *Parser[T]) lookahead() source.Token {
	// NOTE: there is always a lookahead expression because EOF is always
	// appended at the end of the token stream.
	return p.tokens[p.index]
}

func (p *Parser[T]) expect(kind uint) source.Token {
	if p.lookahead().Kind != kind {
		panic("internal failure")
	}
	//
	token := p.tokens[p.index]
	p.index++
	//
	return token
}

func (p *Parser[T]) match(kind uint) bool {
	if p.lookahead().Kind == kind {
		p.index++
		return true
	}
	//
	return false
}

func (p *Parser[T]) syntaxErrors(token source.Token, msg string) []source.SyntaxError {
	return []source.SyntaxError{*p.srcfile.SyntaxError(token.Span, msg)}
}
