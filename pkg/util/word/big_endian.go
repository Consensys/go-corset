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
package word

import (
	"bytes"
	"hash/fnv"
	"math/big"

	"github.com/consensys/go-corset/pkg/util/collection/bit"
	"github.com/consensys/go-corset/pkg/util/collection/hash"
)

// BigEndian captures the notion of an array of bytes represented in big endian
// form.  This is really just a wrapper for convenience, and to help clarify the
// underlying byte order.
type BigEndian struct {
	bytes []byte
}

var _ hash.Hasher[BigEndian] = BigEndian{}

// NewBigEndian constructs a new big endian byte array.
func NewBigEndian(bytes []byte) BigEndian {
	return BigEndian{trim(bytes)}
}

// AsBigInt returns a freshly allocated big integer from the given bytes.
func (p BigEndian) AsBigInt() big.Int {
	var val big.Int
	return *val.SetBytes(p.bytes)
}

// Bit returnsthe bit at a given offset in this word, where offsets always start
// with the least-significant.
func (p BigEndian) Bit(offset uint) bool {
	var bitwidth = p.BitWidth()
	// If offset is past the end of the available bits, then it must have been
	// in the trimmed region and, therefore, was 0.
	if offset < bitwidth {
		return bit.ReadBigEndian(p.bytes, offset)
	}
	//
	return false
}

// BitWidth returns the actual bitwidth of this big endian.
func (p BigEndian) BitWidth() uint {
	return uint(len(p.bytes)) * 8
}

// Cmp implements a byte comparisong between two big endian instances.
func (p BigEndian) Cmp(o BigEndian) int {
	return bytes.Compare(p.bytes, o.bytes)
}

// Equals implementation for the hash.Hasher interface.
func (p BigEndian) Equals(o BigEndian) bool {
	return bytes.Equal(p.bytes, o.bytes)
}

// Hash implementation for the hash.Hasher interface.
func (p BigEndian) Hash() uint64 {
	hash := fnv.New64a()
	hash.Write(p.bytes)
	// Done
	return hash.Sum64()
}

// Put implementation for Word interface.
func (p BigEndian) Put(bytes []byte) []byte {
	var (
		n = uint(len(bytes))
		m = uint(len(p.bytes))
	)
	// Sanity check space
	if len(bytes) < len(p.bytes) {
		bytes = make([]byte, len(p.bytes))
	}
	//
	for m > 0 {
		m--
		n--
		bytes[n] = p.bytes[m]
	}
	//
	for n > 0 {
		n--
		bytes[n] = 0
	}
	//
	return bytes
}

// Set implementation for Word interface.
func (p BigEndian) Set(bytes []byte) BigEndian {
	return BigEndian{trim(bytes)}
}

// Bytes returns a direct access to the underlying byte array in big endian
// form.
func (p BigEndian) Bytes() []byte {
	return p.bytes
}

func (p BigEndian) String() string {
	bi := p.AsBigInt()
	//
	return bi.String()
}

func trim(bytes []byte) []byte {
	// trim any leading zeros to ensure words are in a canonical form.
	for len(bytes) > 0 && bytes[0] == 0 {
		bytes = bytes[1:]
	}
	//
	return bytes
}
