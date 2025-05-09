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
package assembler

import (
	"github.com/consensys/go-corset/pkg/util"
	"github.com/consensys/go-corset/pkg/util/source"
	"github.com/consensys/go-corset/pkg/util/source/lex"
)

// END_OF signals "end of file"
const END_OF uint = 0

// WHITESPACE signals whitespace
const WHITESPACE uint = 1

// COMMENT signals ";; ... \n"
const COMMENT uint = 2

// LBRACE signals "("
const LBRACE uint = 3

// RBRACE signals ")"
const RBRACE uint = 4

// LCURLY signals "{"
const LCURLY uint = 5

// RCURLY signals "}"
const RCURLY uint = 6

// COMMA signals ","
const COMMA uint = 7

// COLON signals ":"
const COLON uint = 8

// SEMICOLON signals ":"
const SEMICOLON uint = 9

// NUMBER signals an integer number
const NUMBER uint = 10

// IDENTIFIER signals a column variable.
const IDENTIFIER uint = 11

// RIGHTARROW signals "->"
const RIGHTARROW uint = 12

// EQUALS signals "="
const EQUALS uint = 13

// EQUALS_EQUALS signals "=="
const EQUALS_EQUALS uint = 14

// NOT_EQUALS signals "!="
const NOT_EQUALS uint = 15

// LESS_THAN signals "<"
const LESS_THAN uint = 16

// LESS_THAN_EQUALS signals "<="
const LESS_THAN_EQUALS uint = 17

// GREATER_THAN signals ">"
const GREATER_THAN uint = 18

// GREATER_THAN_EQUALS signals ">="
const GREATER_THAN_EQUALS uint = 19

// ADD signals "+"
const ADD uint = 20

// SUB signals "-"
const SUB uint = 21

// MUL signals "*"
const MUL uint = 22

// Rule for describing whitespace
var whitespace lex.Scanner[rune] = lex.Many(lex.Or(lex.Unit(' '), lex.Unit('\t'), lex.Unit('\n')))

// Rule for describing numbers
var number lex.Scanner[rune] = lex.Many(lex.Within('0', '9'))

var identifierStart lex.Scanner[rune] = lex.Or(
	lex.Unit('_'),
	lex.Unit('\''),
	lex.Within('a', 'z'),
	lex.Within('A', 'Z'))

var identifierRest lex.Scanner[rune] = lex.Many(lex.Or(
	lex.Unit('_'),
	lex.Unit('\''),
	lex.Within('0', '9'),
	lex.Within('a', 'z'),
	lex.Within('A', 'Z')))

// Rule for describing identifiers
var identifier lex.Scanner[rune] = lex.And(identifierStart, identifierRest)

// Comments start with ';;'
var commentStart lex.Scanner[rune] = lex.Unit(';', ';')

// Comments continue until a newline or EOF.
var commentRest lex.Scanner[rune] = lex.Until('\n')

var comment lex.Scanner[rune] = lex.And(commentStart, commentRest)

// lexing rules
var rules []lex.LexRule[rune] = []lex.LexRule[rune]{
	lex.Rule(comment, COMMENT),
	lex.Rule(lex.Unit('('), LBRACE),
	lex.Rule(lex.Unit(')'), RBRACE),
	lex.Rule(lex.Unit('{'), LCURLY),
	lex.Rule(lex.Unit('}'), RCURLY),
	lex.Rule(lex.Unit(','), COMMA),
	lex.Rule(lex.Unit(':'), COLON),
	lex.Rule(lex.Unit(';'), SEMICOLON),
	lex.Rule(lex.Unit('-', '>'), RIGHTARROW),
	lex.Rule(lex.Unit('=', '='), EQUALS_EQUALS),
	lex.Rule(lex.Unit('!', '='), NOT_EQUALS),
	lex.Rule(lex.Unit('<', '='), LESS_THAN_EQUALS),
	lex.Rule(lex.Unit('>', '='), GREATER_THAN_EQUALS),
	lex.Rule(lex.Unit('<'), LESS_THAN),
	lex.Rule(lex.Unit('>'), GREATER_THAN),
	lex.Rule(lex.Unit('='), EQUALS),
	lex.Rule(lex.Unit('+'), ADD),
	lex.Rule(lex.Unit('-'), SUB),
	lex.Rule(lex.Unit('*'), MUL),
	lex.Rule(whitespace, WHITESPACE),
	lex.Rule(number, NUMBER),
	lex.Rule(identifier, IDENTIFIER),
	lex.Rule(lex.Eof[rune](), END_OF),
}

// Lex a given source file into a sequence of zero or more tokens, along with
// any syntax errors arising.
func Lex(srcfile source.File) ([]lex.Token, []source.SyntaxError) {
	var (
		lexer = lex.NewLexer(srcfile.Contents(), rules...)
		// Lex as many tokens as possible
		tokens = lexer.Collect()
	)
	// Check whether anything was left (if so this is an error)
	if lexer.Remaining() != 0 {
		start, end := lexer.Index(), lexer.Index()+lexer.Remaining()
		err := srcfile.SyntaxError(source.NewSpan(int(start), int(end)), "unknown text encountered")
		// errors
		return nil, []source.SyntaxError{*err}
	}
	// Remove any whitespace
	tokens = util.RemoveMatching(tokens, func(t lex.Token) bool { return t.Kind == WHITESPACE })
	// Remove any comments (for not)
	tokens = util.RemoveMatching(tokens, func(t lex.Token) bool { return t.Kind == COMMENT })
	// Done
	return tokens, nil
}
