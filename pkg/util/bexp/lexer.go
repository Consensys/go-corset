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
	"github.com/consensys/go-corset/pkg/util/sexp"
)

// Token associates a piece of information with a given range of characters in
// the string being scanned.
type Token struct {
	Kind uint
	Span sexp.Span
}

type Scanner[T any] func([]T) util.Option[Token]

// Lexer provides a top-level construct for tokenising a given input string.
type Lexer[T any] struct {
	runes   []T
	index   uint
	scanner Scanner[T]
}

// NewLexer constructs a new lexer with a given scanner.
func NewLexer[T any](input []T, scanner Scanner[T]) *Lexer[T] {
	return &Lexer[T]{
		input,
		0,
		scanner,
	}
}

// Check whether or not there are any items remaining to visit.
func (p *Lexer[T]) HasNext() bool {
	panic("todo")
}

// Next returns the next item and advances the lexer.
func (p *Lexer[T]) Next() T {
	panic("todo")
}
