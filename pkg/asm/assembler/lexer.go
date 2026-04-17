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
	"fmt"

	"github.com/consensys/go-corset/pkg/util/collection/array"
	"github.com/consensys/go-corset/pkg/util/source"
	"github.com/consensys/go-corset/pkg/util/source/lex"
)

// AsmTokenKind identifies the kind of a lexer token in the assembler parser.
type AsmTokenKind uint

// String returns the name of the token kind for display and debugging.
func (k AsmTokenKind) String() string {
	switch k {
	case END_OF:
		return "END_OF"
	case WHITESPACE:
		return "WHITESPACE"
	case COMMENT:
		return "COMMENT"
	case LBRACE:
		return "LBRACE"
	case RBRACE:
		return "RBRACE"
	case LCURLY:
		return "LCURLY"
	case RCURLY:
		return "RCURLY"
	case COMMA:
		return "COMMA"
	case COLON:
		return "COLON"
	case SEMICOLON:
		return "SEMICOLON"
	case NUMBER:
		return "NUMBER"
	case STRING:
		return "STRING"
	case IDENTIFIER:
		return "IDENTIFIER"
	case KEYWORD_CONST:
		return "KEYWORD_CONST"
	case KEYWORD_INCLUDE:
		return "KEYWORD_INCLUDE"
	case KEYWORD_FN:
		return "KEYWORD_FN"
	case KEYWORD_PUB:
		return "KEYWORD_PUB"
	case RIGHTARROW:
		return "RIGHTARROW"
	case EQUALS:
		return "EQUALS"
	case EQUALS_EQUALS:
		return "EQUALS_EQUALS"
	case NOT_EQUALS:
		return "NOT_EQUALS"
	case LESS_THAN:
		return "LESS_THAN"
	case LESS_THAN_EQUALS:
		return "LESS_THAN_EQUALS"
	case GREATER_THAN:
		return "GREATER_THAN"
	case GREATER_THAN_EQUALS:
		return "GREATER_THAN_EQUALS"
	case ADD:
		return "ADD"
	case SUB:
		return "SUB"
	case MUL:
		return "MUL"
	case DIV:
		return "DIV"
	case QMARK:
		return "QMARK"
	default:
		return fmt.Sprintf("AsmTokenKind(%d)", uint(k))
	}
}

// Primitive tokens (0–11)
const (
	END_OF     AsmTokenKind = 0
	WHITESPACE AsmTokenKind = 1
	COMMENT    AsmTokenKind = 2
	LBRACE     AsmTokenKind = 3
	RBRACE     AsmTokenKind = 4
	LCURLY     AsmTokenKind = 5
	RCURLY     AsmTokenKind = 6
	COMMA      AsmTokenKind = 7
	COLON      AsmTokenKind = 8
	SEMICOLON  AsmTokenKind = 9
	NUMBER     AsmTokenKind = 10
	STRING     AsmTokenKind = 11
)

// Identifiers and keywords (20–24)
const (
	IDENTIFIER      AsmTokenKind = 20
	KEYWORD_CONST   AsmTokenKind = 21
	KEYWORD_INCLUDE AsmTokenKind = 22
	KEYWORD_FN      AsmTokenKind = 23
	KEYWORD_PUB     AsmTokenKind = 24
)

// Operators (30–50)
const (
	RIGHTARROW          AsmTokenKind = 30
	EQUALS              AsmTokenKind = 31
	EQUALS_EQUALS       AsmTokenKind = 32
	NOT_EQUALS          AsmTokenKind = 33
	LESS_THAN           AsmTokenKind = 34
	LESS_THAN_EQUALS    AsmTokenKind = 35
	GREATER_THAN        AsmTokenKind = 36
	GREATER_THAN_EQUALS AsmTokenKind = 37
	ADD                 AsmTokenKind = 38
	SUB                 AsmTokenKind = 39
	MUL                 AsmTokenKind = 40
	DIV                 AsmTokenKind = 41
	QMARK               AsmTokenKind = 50
)

// Rule for describing whitespace
var whitespace lex.Scanner[rune] = lex.Many(lex.Or(lex.Unit(' '), lex.Unit('\t'), lex.Unit('\n')))

// Rule for describing numbers
// A number is either a hexadecimal, binary, or decimal one.
// Allowing (and ignoring) '_' in the middle of a number for readability.
var (
	binaryStart  = lex.Sequence(lex.String("0b"), lex.Within('0', '1'))
	binaryRest   = lex.Or(lex.Within('0', '1'), lex.Unit('_'))
	decimalStart = lex.Within('0', '9')
	decimalRest  = lex.Or(lex.Within('0', '9'), lex.Unit('_'), lex.Unit('^'))
	hexDigit     = lex.Or(lex.Within('0', '9'), lex.Within('A', 'F'), lex.Within('a', 'f'))
	hexStart     = lex.Sequence(lex.String("0x"), hexDigit)
	hexRest      = lex.Or(hexDigit, lex.Unit('_'))

	number = lex.Or(
		lex.SequenceNullableLast(binaryStart, lex.Many(binaryRest)),
		lex.SequenceNullableLast(hexStart, lex.Many(hexRest)),
		lex.SequenceNullableLast(decimalStart, lex.Many(decimalRest)),
	)
)

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

// Rule for describing strings in quotes
var strung lex.Scanner[rune] = lex.Sequence(lex.Unit('"'), lex.Many(lex.Not('"')), lex.Unit('"'))

// Comments start with ';;'
var commentStart lex.Scanner[rune] = lex.Unit(';', ';')

// Comments continue until a newline or EOF.
var commentRest lex.Scanner[rune] = lex.Until('\n')

var comment lex.Scanner[rune] = lex.And(commentStart, commentRest)

// lexing rules
var rules []lex.LexRule[rune, AsmTokenKind] = []lex.LexRule[rune, AsmTokenKind]{
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
	lex.Rule(lex.Unit('/'), DIV),
	lex.Rule(lex.Unit('?'), QMARK),
	lex.Rule(whitespace, WHITESPACE),
	lex.Rule(number, NUMBER),
	lex.Rule(strung, STRING),
	lex.Rule(lex.String("pub"), KEYWORD_PUB),
	lex.Rule(lex.String("const"), KEYWORD_CONST),
	lex.Rule(lex.String("include"), KEYWORD_INCLUDE),
	lex.Rule(lex.String("fn"), KEYWORD_FN),
	lex.Rule(identifier, IDENTIFIER),
	lex.Rule(lex.Eof[rune](), END_OF),
}

// Lex a given source file into a sequence of zero or more tokens, along with
// any syntax errors arising.
func Lex(srcfile source.File) ([]lex.Token[AsmTokenKind], []source.SyntaxError) {
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
	tokens = array.RemoveMatching(tokens, func(t lex.Token[AsmTokenKind]) bool { return t.Kind == WHITESPACE })
	// Remove any comments (for not)
	tokens = array.RemoveMatching(tokens, func(t lex.Token[AsmTokenKind]) bool { return t.Kind == COMMENT })
	// Done
	return tokens, nil
}
