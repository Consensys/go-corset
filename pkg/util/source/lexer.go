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

// Token associates a piece of information with a given range of characters in
// the string being scanned.
type Token struct {
	Kind uint
	Span Span
}

// Lexer provides a top-level construct for tokenising a given input string.
type Lexer[T any] struct {
	items   []T
	index   int
	scanner Scanner[T]
	buffer  []Token
}

// NewLexer constructs a new lexer with a given scanner.
func NewLexer[T any](input []T, scanner Scanner[T]) *Lexer[T] {
	return &Lexer[T]{
		input,
		0,
		scanner,
		nil,
	}
}

// Remaining determines how many characters from the original sequence were
// left.
func (p *Lexer[T]) Remaining() uint {
	return uint(max(0, len(p.items)-p.index))
}

// HasNext checks whether or not there are any items remaining to visit.
func (p *Lexer[T]) HasNext() bool {
	p.scan()
	return len(p.buffer) > 0
}

// Next returns the next item and advances the lexer.
func (p *Lexer[T]) Next() Token {
	next := p.buffer[0]
	p.buffer = p.buffer[1:]
	//
	if p.index == len(p.items) {
		// EOF condition
		p.index++
	} else {
		p.index = next.Span.End()
	}
	//
	return next
}

// Collect is a convenience function which parses all remaining tokens in one
// go, producing an array of tokens.
func (p *Lexer[T]) Collect() []Token {
	var tokens []Token
	// Keep scanning
	for p.HasNext() {
		tokens = append(tokens, p.Next())
	}
	//
	return tokens
}

// internal scan functions.
func (p *Lexer[T]) scan() {
	if len(p.buffer) == 0 && p.index <= len(p.items) {
		// Look for item
		next := p.scanner.Scan(p.items[p.index:])
		// Check what we got
		if next.HasValue() {
			n := next.Unwrap()
			// Shift span into correct position
			n.Span = NewSpan(n.Span.Start()+p.index, n.Span.End()+p.index)
			// Insert into buffer
			p.buffer = append(p.buffer, n)
		}
	}
}
