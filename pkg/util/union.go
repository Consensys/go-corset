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
package util

import (
	"bytes"
	"encoding/gob"
)

// Comparable interface which can be implemented by non-primitive types.
type Comparable[T any] interface {
	// Cmp returns < 0 if this is less than other, or 0 if they are equal, or >
	// 0 if this is greater than other.
	Cmp(other T) int
}

// Union represents a value which is either of the first type or of the second
// type.
type Union[S, T any] struct {
	// Indicates first present
	Sign bool
	// Left value
	first S
	// Right value
	second T
}

// Union1 constructs a union holding a value of the first type.
func Union1[S, T any](value S) Union[S, T] {
	var empty T
	//
	return Union[S, T]{true, value, empty}
}

// Union2 constructs a union holding a value of the second type.
func Union2[S, T any](value T) Union[S, T] {
	var empty S
	//
	return Union[S, T]{false, empty, value}
}

// HasFirst indicates whether this union holds a value of the first type (or
// not).
func (u Union[S, T]) HasFirst() bool {
	return u.Sign
}

// HasSecond indicates whether this union holds a value of the second type (or
// not).
func (u Union[S, T]) HasSecond() bool {
	return !u.Sign
}

// First returns the contained value of the first type.  If the union does not
// hold a value of the first type, then this will panic.
func (u Union[S, T]) First() S {
	if u.Sign {
		return u.first
	}
	//
	panic("cannot take first item, as union holds second")
}

// Second returns the contained value of the second type.  If the union does not
// hold a value of the second type, then this will panic.
func (u Union[S, T]) Second() T {
	if !u.Sign {
		return u.second
	}
	//
	panic("cannot take second item, as union holds first")
}

// ============================================================================
// Encoding / Decoding
// ============================================================================

// GobEncode an option.  This allows it to be marshalled into a binary form.
func (u Union[S, T]) GobEncode() (data []byte, err error) {
	var (
		buffer     bytes.Buffer
		gobEncoder = gob.NewEncoder(&buffer)
	)
	// Some
	if err := gobEncoder.Encode(&u.Sign); err != nil {
		return nil, err
	}
	// Decide whether need anything else.
	if u.Sign {
		// FirstValue
		if err := gobEncoder.Encode(&u.first); err != nil {
			return nil, err
		}
	} else {
		// SecondValue
		if err := gobEncoder.Encode(&u.second); err != nil {
			return nil, err
		}
	}
	// Success
	return buffer.Bytes(), nil
}

// GobDecode a previously encoded option
func (u *Union[S, T]) GobDecode(data []byte) error {
	buffer := bytes.NewBuffer(data)
	gobDecoder := gob.NewDecoder(buffer)
	// Some
	if err := gobDecoder.Decode(&u.Sign); err != nil {
		return err
	}
	// Check whether value provided
	if u.Sign {
		// FirstValue
		if err := gobDecoder.Decode(&u.first); err != nil {
			return err
		}
	} else {
		// SecondValue
		if err := gobDecoder.Decode(&u.second); err != nil {
			return err
		}
	}
	// Success!
	return nil
}
