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
package source

import (
	"slices"
	"testing"
)

func TestLexer_00(t *testing.T) {
	var tokens []Token = []Token{
		{END_OF, NewSpan(0, 0)},
	}

	checkLexer(t, "", 0, tokens...)
}

func TestLexer_01(t *testing.T) {
	var tokens []Token = []Token{
		{LBRACE, NewSpan(0, 1)},
		{END_OF, NewSpan(1, 1)},
	}

	checkLexer(t, "(", 0, tokens...)
}

func TestLexer_02(t *testing.T) {
	var tokens []Token = []Token{
		{LBRACE, NewSpan(0, 1)},
		{RBRACE, NewSpan(1, 2)},
		{END_OF, NewSpan(2, 2)},
	}

	checkLexer(t, "()", 0, tokens...)
}

func TestLexer_03(t *testing.T) {
	var tokens []Token = []Token{}

	checkLexer(t, "x", 1, tokens...)
}

func TestLexer_04(t *testing.T) {
	var tokens []Token = []Token{
		{LBRACE, NewSpan(0, 1)},
		{WSPACE, NewSpan(1, 2)},
		{RBRACE, NewSpan(2, 3)},
		{END_OF, NewSpan(3, 3)},
	}

	checkLexer(t, "( )", 0, tokens...)
}

func TestLexer_05(t *testing.T) {
	var tokens []Token = []Token{
		{LBRACE, NewSpan(0, 1)},
		{WSPACE, NewSpan(1, 3)},
		{RBRACE, NewSpan(3, 4)},
		{END_OF, NewSpan(4, 4)},
	}

	checkLexer(t, "(  )", 0, tokens...)
}

func TestLexer_06(t *testing.T) {
	var tokens []Token = []Token{
		{NUMBER, NewSpan(0, 1)},
		{END_OF, NewSpan(1, 1)},
	}

	checkLexer(t, "1", 0, tokens...)
}

func TestLexer_07(t *testing.T) {
	var tokens []Token = []Token{
		{NUMBER, NewSpan(0, 2)},
		{END_OF, NewSpan(2, 2)},
	}

	checkLexer(t, "12", 0, tokens...)
}
func TestLexer_08(t *testing.T) {
	var tokens []Token = []Token{
		{NUMBER, NewSpan(0, 3)},
		{END_OF, NewSpan(3, 3)},
	}

	checkLexer(t, "123", 0, tokens...)
}
func TestLexer_09(t *testing.T) {
	var tokens []Token = []Token{
		{LBRACE, NewSpan(0, 1)},
		{NUMBER, NewSpan(1, 3)},
		{RBRACE, NewSpan(3, 4)},
		{END_OF, NewSpan(4, 4)},
	}

	checkLexer(t, "(90)", 0, tokens...)
}

// ==================================================================
// Framework
// ==================================================================

const END_OF uint = 0
const WSPACE uint = 1
const LBRACE uint = 2
const RBRACE uint = 3
const NUMBER uint = 4

var scanner Scanner[rune] = Or(
	One(LBRACE, '('),
	One(RBRACE, ')'),
	Many(WSPACE, ' ', '\t'),
	ManyWith(NUMBER, '0', '9'),
	Eof[rune](END_OF))

func checkLexer(t *testing.T, input string, remainder uint, expected ...Token) {
	items := []rune(input)
	// Construct text lexer
	lexer := NewLexer[rune](items, scanner)
	// Apply lexer
	tokens := lexer.Collect()
	// Keep scanning
	if !slices.Equal(tokens, expected) {
		t.Errorf("got %v, expected %v", tokens, expected)
	} else if lexer.Remaining() != remainder {
		n := len(items) - int(lexer.Remaining())
		t.Errorf("unmatched items: %v", items[n:])
	}
}
