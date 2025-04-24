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
package lex

import (
	"cmp"
)

// Scanner is a function which accepts a given item or not.
type Scanner[T any] func(item []T) uint

// And combines zero or more scanners such that the resulting scanner succeeds if
// all of the scanners succeed.  Observe, however, that there is an implicit
// left-to-right order of evaluation.
func And[T any](scanners ...Scanner[T]) Scanner[T] {
	return func(items []T) uint {
		n := uint(0)

		for _, scanner := range scanners {
			m := scanner(items)
			if m == 0 {
				// fail
				return 0
			}
			//
			n = max(n, m)
		}
		// fail
		return n
	}
}

// Or combines zero or more scanners such that the resulting scanner succeeds if
// any of the scanners succeeds.  Observe, however, that there is an implicit
// left-to-right order of evaluation.
func Or[T any](scanners ...Scanner[T]) Scanner[T] {
	return func(items []T) uint {
		for _, scanner := range scanners {
			if n := scanner(items); n > 0 {
				return n
			}
		}
		// fail
		return 0
	}
}

// Unit accepts a given sequence of characters.  That is, for this scanner to
// match, it must match all the given characters (one after the other) in their given order.
func Unit[T comparable](chars ...T) Scanner[T] {
	return func(items []T) uint {
		if len(items) >= len(chars) {
			for i := 0; i < len(chars); i++ {
				if items[i] != chars[i] {
					// fail
					return 0
				}
			}
			// success
			return uint(len(chars))
		}
		// fail
		return 0
	}
}

// Within accepts any character within a given range.
func Within[T cmp.Ordered](lowest T, highest T) Scanner[T] {
	return func(items []T) uint {
		if len(items) != 0 && lowest <= items[0] && items[0] <= highest {
			return 1
		}
		// fail
		return 0
	}
}

// Many matches one or more of a given item.
func Many[T any](acceptor Scanner[T]) Scanner[T] {
	return func(items []T) uint {
		index := uint(0)
		//
		for index < uint(len(items)) {
			if n := acceptor(items[index:]); n != 0 {
				index += n
				continue
			}
			//
			break
		}
		// done
		return index
	}
}

// Until matches everything until a particular item is matched.
func Until[T comparable](item T) Scanner[T] {
	return func(items []T) uint {
		index := uint(0)
		//
		for index < uint(len(items)) {
			if items[index] == item {
				break
			}
			// continue match
			index = index + 1
		}
		// done
		return index
	}
}

// Eof matches the end of the input stream.
func Eof[T any]() Scanner[T] {
	return func(items []T) uint {
		if len(items) == 0 {
			return 1
		}
		//
		return 0
	}
}
