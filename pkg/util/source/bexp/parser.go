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
	"github.com/consensys/go-corset/pkg/util"
	"github.com/consensys/go-corset/pkg/util/source"
)

// Parse a given input string into logical proposition.
func Parse[T Proposition[T]](input string) (T, []source.SyntaxError) {
	var (
		prop    T
		srcfile = source.NewSourceFile("expr", []byte(input))
		lexer   = source.NewLexer[rune](srcfile.Contents(), scanner)
		// Lex as many tokens as possible
		tokens = lexer.Collect()
	)
	// Check whether anything was left (if so this is an error)
	if lexer.Remaining() != 0 {
		start, end := lexer.Index(), lexer.Index()+lexer.Remaining()
		err := srcfile.SyntaxError(source.NewSpan(int(start), int(end)), "unknown text encountered")

		return prop, []source.SyntaxError{*err}
	}
	// Remove any whitespace
	tokens = util.RemoveMatching(tokens, func(t source.Token) bool { return t.Kind == WHITESPACE })
	//
	parser := &Parser[T]{srcfile, tokens, 0}
	// For now, always fail.Hey
	p, errs := parser.ParseProp()
	// Check all parsed
	if len(errs) == 0 && !parser.Done() {
		//nolint
		token, _ := parser.lookahead()
		err := srcfile.SyntaxError(token.Span, "unknown token")
		//
		return p, []source.SyntaxError{*err}
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

var scanner source.Scanner[rune] = source.Or(
	source.One(LBRACE, '('),
	source.One(RBRACE, ')'),
	source.Many(WHITESPACE, ' ', '\t'),
	source.ManyWith(NUMBER, '0', '9'),
	source.ManyWith(IDENTIFIER, 'a', 'z'),
	source.Eof[rune](END_OF))

// Parser provides a general-purpose parser for propositions and arithmetic
// expressions.
type Parser[T Proposition[T]] struct {
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

// ParseProp parses a proposition
func (p *Parser[T]) ParseProp() (T, []source.SyntaxError) {
	var (
		empty      T
		token, err = p.lookahead()
	)
	//
	if err != nil {
		return empty, []source.SyntaxError{*err}
	}
	//
	switch token.Kind {
	case IDENTIFIER:
		return p.parseIdentifier()
	}
	//
	return empty, p.syntaxErrors(token, "proposition expected")
}

func (p *Parser[T]) parseIdentifier() (T, []source.SyntaxError) {
	var empty T
	//
	p.index++
	//
	return empty, nil
}

func (p *Parser[T]) syntaxErrors(token source.Token, msg string) []source.SyntaxError {
	return []source.SyntaxError{*p.srcfile.SyntaxError(token.Span, msg)}
}

func (p *Parser[T]) lookahead() (source.Token, *source.SyntaxError) {
	// NOTE: there is always a lookahead expression because EOF is always
	// appended at the end of the token stream.
	return p.tokens[p.index], nil
}
