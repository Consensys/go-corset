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
	"cmp"
	"slices"

	"github.com/consensys/go-corset/pkg/util"
)

// Scanner looks at a given sequence of items, starting from the beginning, and
// attempts to consume 1 or more of them.  If it cannot consume any, then None
// is returned.  Otherwise, it returns a Token which spans characters 0..n+1
// where n is the last character of the token.
type Scanner[T any] interface {
	Scan([]T) util.Option[Token]
}

// Eof adds a given tag to the end of the token stream.
func Eof[T comparable](tag uint) *eofScanner[T] {
	return &eofScanner[T]{tag}
}

// One creates a scanner responsible for associating a single item with a given
// tag.
func One[T comparable](tag uint, item T) *unitScanner[T] {
	return &unitScanner[T]{item, tag}
}

// Many creates a scanner responsible for associating zero or more items with a
// given tag.
func Many[T comparable](tag uint, items ...T) *manyScanner[T] {
	return &manyScanner[T]{tag, items}
}

// ManyWith creates a scanner responsible for associating items in a given range
// with a given tag.
func ManyWith[T cmp.Ordered](tag uint, first T, last T) *manyWithinScanner[T] {
	return &manyWithinScanner[T]{tag, first, last}
}

// Or constructs a scanner which accepts words accepts by any of the given
// scanners.
func Or[T comparable](scanners ...Scanner[T]) Scanner[T] {
	return &orScanner[T]{scanners}
}

// ============================================================================
// Eof Scanner
// ============================================================================

type eofScanner[T comparable] struct {
	tag uint
}

func (p *eofScanner[T]) Scan(items []T) util.Option[Token] {
	if len(items) == 0 {
		token := Token{p.tag, NewSpan(0, 0)}
		return util.Some(token)
	}
	//
	return util.None[Token]()
}

// ============================================================================
// Unit Scanner
// ============================================================================

type unitScanner[T comparable] struct {
	item T
	tag  uint
}

func (p *unitScanner[T]) Scan(items []T) util.Option[Token] {
	if len(items) > 0 && items[0] == p.item {
		token := Token{p.tag, NewSpan(0, 1)}
		return util.Some(token)
	}
	//
	return util.None[Token]()
}

// ============================================================================
// Many Scanner
// ============================================================================

type manyScanner[T comparable] struct {
	tag   uint
	items []T
}

func (p *manyScanner[T]) Scan(items []T) util.Option[Token] {
	i := 0
	//
	for i < len(items) && slices.Contains[[]T](p.items, items[i]) {
		i++
	}
	//
	if i != 0 {
		token := Token{p.tag, NewSpan(0, i)}
		return util.Some[Token](token)
	}
	//
	return util.None[Token]()
}

// ============================================================================
// ManyWithin Scanner
// ============================================================================

type manyWithinScanner[T cmp.Ordered] struct {
	tag   uint
	first T
	last  T
}

func (p *manyWithinScanner[T]) Scan(items []T) util.Option[Token] {
	i := 0
	//
	for i < len(items) && p.first <= items[i] && items[i] <= p.last {
		i++
	}
	//
	if i != 0 {
		token := Token{p.tag, NewSpan(0, i)}
		return util.Some[Token](token)
	}
	//
	return util.None[Token]()
}

// ============================================================================
// Or Scanner
// ============================================================================

type orScanner[T comparable] struct {
	scanners []Scanner[T]
}

func (p *orScanner[T]) Scan(items []T) util.Option[Token] {
	for _, scanner := range p.scanners {
		if res := scanner.Scan(items); res.HasValue() {
			return res
		}
	}
	// Failed
	return util.None[Token]()
}
