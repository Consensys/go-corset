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
package parser

import (
	"github.com/consensys/go-corset/pkg/util/collection/array"
	"github.com/consensys/go-corset/pkg/util/source"
	"github.com/consensys/go-corset/pkg/util/source/lex"
)

const (
	// END_OF signals "end of file"
	END_OF uint = iota
	// WHITESPACE signals whitespace
	WHITESPACE
	// COMMENT signals "// ... \n"
	COMMENT
	// LBRACE signals "("
	LBRACE
	// RBRACE signals ")"
	RBRACE
	// LCURLY signals "{"
	LCURLY
	// RCURLY signals "}"
	RCURLY
	// LSQUARE signals "["
	LSQUARE
	// RSQUARE signals "]"
	RSQUARE
	// COMMA signals ","
	COMMA
	// COLON signals ":"
	COLON
	// COLONCOLON signals "::"
	COLONCOLON
	// SEMICOLON signals ":"
	SEMICOLON
	// NUMBER signals an integer number
	NUMBER
	// STRING signals a quoted string
	STRING
	// IDENTIFIER signals a column variable
	IDENTIFIER
	// KEYWORD_AS signals a type cast expression (e.g. "x as u8")
	KEYWORD_AS
	// KEYWORD_BREAK signals a break statement
	KEYWORD_BREAK
	// KEYWORD_CONTINUE signals a continue statement
	KEYWORD_CONTINUE
	// KEYWORD_CONST signals a constant declaration
	KEYWORD_CONST
	// KEYWORD_ELSE signals an else branch
	KEYWORD_ELSE
	// KEYWORD_FAIL signals a return statement
	KEYWORD_FAIL
	// KEYWORD_FN signals a function declaration
	KEYWORD_FN
	// KEYWORD_FOR signals a for loop
	KEYWORD_FOR
	// KEYWORD_IF signals a return statement
	KEYWORD_IF
	// KEYWORD_INCLUDE signals an include declaration
	KEYWORD_INCLUDE
	// KEYWORD_INPUT signals a read-only memory
	KEYWORD_INPUT
	// KEYWORD_MEMORY signals a random-access memory declaration
	KEYWORD_MEMORY
	// KEYWORD_RETURN signals a return statement
	KEYWORD_RETURN
	// KEYWORD_STATIC signals a static read-only memory
	KEYWORD_STATIC
	// KEYWORD_OUTPUT signals a write-once memory
	KEYWORD_OUTPUT
	// KEYWORD_PRINTF signals a printf statement
	KEYWORD_PRINTF
	// KEYWORD_PUB signals a public input / output
	KEYWORD_PUB
	// KEYWORD_WHILE signals a while loop
	KEYWORD_WHILE
	// KEYWORD_VAR signals a local variable declaration
	KEYWORD_VAR
	// KEYWORD_TYPE signals a type alias declaration
	KEYWORD_TYPE
	// RIGHTARROW signals "->"
	RIGHTARROW
	// EQUALS signals "="
	EQUALS
	// EQUALS_EQUALS signals "=="
	EQUALS_EQUALS
	// NOT_EQUALS signals "!="
	NOT_EQUALS
	// LESS_THAN signals "<"
	LESS_THAN
	// LESS_THAN_EQUALS signals "<="
	LESS_THAN_EQUALS
	// GREATER_THAN signals ">"
	GREATER_THAN
	// GREATER_THAN_EQUALS signals ">="
	GREATER_THAN_EQUALS
	// LOGICAL_AND signals "&&"
	LOGICAL_AND
	// LOGICAL_OR signals "||"
	LOGICAL_OR
	// LOGICAL_NOT signals "!"
	LOGICAL_NOT
	// ADD signals "+"
	ADD
	// SUB signals "-"
	SUB
	// MUL signals "*"
	MUL
	// DIV signals "/"
	DIV
	// BITWISE_AND signals "&"
	BITWISE_AND
	// BITWISE_OR signals "|"
	BITWISE_OR
	// BITWISE_XOR signals "^"
	BITWISE_XOR
	// BITWISE_NOT signals "~"
	BITWISE_NOT
	// BITWISE_SHL signals "<<"
	BITWISE_SHL
	// BITWISE_SHR signals ">>"
	BITWISE_SHR
	// REM signals "%"
	REM
	// QMARK signals "?"
	QMARK
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

// Comments start with '//'
var commentStart lex.Scanner[rune] = lex.Unit('/', '/')

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
	lex.Rule(lex.Unit('['), LSQUARE),
	lex.Rule(lex.Unit(']'), RSQUARE),
	lex.Rule(lex.Unit(','), COMMA),
	lex.Rule(lex.Unit(':', ':'), COLONCOLON),
	lex.Rule(lex.Unit(':'), COLON),
	lex.Rule(lex.Unit(';'), SEMICOLON),
	lex.Rule(lex.Unit('-', '>'), RIGHTARROW),
	lex.Rule(lex.Unit('=', '='), EQUALS_EQUALS),
	lex.Rule(lex.Unit('!', '='), NOT_EQUALS),
	lex.Rule(lex.Unit('<', '='), LESS_THAN_EQUALS),
	lex.Rule(lex.Unit('>', '='), GREATER_THAN_EQUALS),
	lex.Rule(lex.Unit('<', '<'), BITWISE_SHL),
	lex.Rule(lex.Unit('>', '>'), BITWISE_SHR),
	lex.Rule(lex.Unit('<'), LESS_THAN),
	lex.Rule(lex.Unit('>'), GREATER_THAN),
	lex.Rule(lex.Unit('='), EQUALS),
	lex.Rule(lex.Unit('+'), ADD),
	lex.Rule(lex.Unit('-'), SUB),
	lex.Rule(lex.Unit('*'), MUL),
	lex.Rule(lex.Unit('/'), DIV),
	lex.Rule(lex.Unit('%'), REM),
	lex.Rule(lex.Unit('!'), LOGICAL_NOT),
	lex.Rule(lex.Unit('&', '&'), LOGICAL_AND),
	lex.Rule(lex.Unit('|', '|'), LOGICAL_OR),
	lex.Rule(lex.Unit('&'), BITWISE_AND),
	lex.Rule(lex.Unit('|'), BITWISE_OR),
	lex.Rule(lex.Unit('^'), BITWISE_XOR),
	lex.Rule(lex.Unit('~'), BITWISE_NOT),
	lex.Rule(lex.Unit('?'), QMARK),
	lex.Rule(whitespace, WHITESPACE),
	lex.Rule(number, NUMBER),
	lex.Rule(strung, STRING),
	lex.Rule(lex.String("as"), KEYWORD_AS),
	lex.Rule(lex.String("break"), KEYWORD_BREAK),
	lex.Rule(lex.String("const"), KEYWORD_CONST),
	lex.Rule(lex.String("continue"), KEYWORD_CONTINUE),
	lex.Rule(lex.String("else"), KEYWORD_ELSE),
	lex.Rule(lex.String("fail"), KEYWORD_FAIL),
	lex.Rule(lex.String("for"), KEYWORD_FOR),
	lex.Rule(lex.String("fn"), KEYWORD_FN),
	lex.Rule(lex.String("if"), KEYWORD_IF),
	lex.Rule(lex.String("include"), KEYWORD_INCLUDE),
	lex.Rule(lex.String("input"), KEYWORD_INPUT),
	lex.Rule(lex.String("memory"), KEYWORD_MEMORY),
	lex.Rule(lex.String("output"), KEYWORD_OUTPUT),
	lex.Rule(lex.String("printf"), KEYWORD_PRINTF),
	lex.Rule(lex.String("pub"), KEYWORD_PUB),
	lex.Rule(lex.String("return"), KEYWORD_RETURN),
	lex.Rule(lex.String("static"), KEYWORD_STATIC),
	lex.Rule(lex.String("type"), KEYWORD_TYPE),
	lex.Rule(lex.String("var"), KEYWORD_VAR),
	lex.Rule(lex.String("while"), KEYWORD_WHILE),
	lex.Rule(identifier, IDENTIFIER),
	lex.Rule(lex.Eof[rune](), END_OF),
}

// Lex a given source file into a sequence of zero or more tokens, along with
// any syntax errors arising. When includeComments is true, COMMENT tokens are
// retained in the output (e.g. for syntax highlighting); otherwise they are
// removed along with whitespace.
func Lex(srcfile source.File, includeComments bool) ([]lex.Token, []source.SyntaxError) {
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
	tokens = array.RemoveMatching(tokens, func(t lex.Token) bool { return t.Kind == WHITESPACE })
	// Remove comments unless the caller wants them (e.g. for syntax highlighting)
	if !includeComments {
		tokens = array.RemoveMatching(tokens, func(t lex.Token) bool { return t.Kind == COMMENT })
	}
	// Done
	return tokens, nil
}
