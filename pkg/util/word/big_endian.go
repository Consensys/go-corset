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
	"cmp"
	"encoding/binary"
	"hash/fnv"
	"math/big"

	"github.com/consensys/go-corset/pkg/util/collection/array"
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
	return BigEndian{array.TrimLeadingZeros(bytes)}
}

// AsBigInt returns a freshly allocated big integer from the given bytes.
func (p BigEndian) AsBigInt() big.Int {
	var val big.Int
	return *val.SetBytes(p.bytes)
}

// Bit returnsthe bit at a given offset in this word, where offsets always start
// with the least-significant.
func (p BigEndian) Bit(offset uint) bool {
	var bitwidth = p.ByteWidth()
	// If offset is past the end of the available bits, then it must have been
	// in the trimmed region and, therefore, was 0.
	if offset < bitwidth {
		return bit.ReadBigEndian(p.bytes, offset)
	}
	//
	return false
}

// ByteWidth implementation for the Word interface.
func (p BigEndian) ByteWidth() uint {
	return uint(len(p.bytes))
}

// Cmp64 implementation for Word interface.
func (p BigEndian) Cmp64(o uint64) int {
	switch p.ByteWidth() {
	case 0:
		return cmp.Compare(0, o)
	case 1:
		tmp := uint64(p.bytes[0])
		return cmp.Compare(tmp, o)
	case 2:
		tmp := uint64(p.bytes[1])
		tmp += uint64(p.bytes[0]) << 8
		//
		return cmp.Compare(tmp, o)
	case 3:
		tmp := uint64(p.bytes[2])
		tmp += uint64(p.bytes[1]) << 8
		tmp += uint64(p.bytes[0]) << 16
		//
		return cmp.Compare(tmp, o)
	case 4:
		tmp := uint64(p.bytes[3])
		tmp += uint64(p.bytes[2]) << 8
		tmp += uint64(p.bytes[1]) << 16
		tmp += uint64(p.bytes[0]) << 24
		//
		return cmp.Compare(tmp, o)
	case 5:
		tmp := uint64(p.bytes[4])
		tmp += uint64(p.bytes[3]) << 8
		tmp += uint64(p.bytes[2]) << 16
		tmp += uint64(p.bytes[1]) << 24
		tmp += uint64(p.bytes[0]) << 32
		//
		return cmp.Compare(tmp, o)
	case 6:
		tmp := uint64(p.bytes[5])
		tmp += uint64(p.bytes[4]) << 8
		tmp += uint64(p.bytes[3]) << 16
		tmp += uint64(p.bytes[2]) << 24
		tmp += uint64(p.bytes[1]) << 32
		tmp += uint64(p.bytes[0]) << 40
		//
		return cmp.Compare(tmp, o)
	case 7:
		tmp := uint64(p.bytes[6])
		tmp += uint64(p.bytes[5]) << 8
		tmp += uint64(p.bytes[4]) << 16
		tmp += uint64(p.bytes[3]) << 24
		tmp += uint64(p.bytes[2]) << 32
		tmp += uint64(p.bytes[1]) << 40
		tmp += uint64(p.bytes[0]) << 48
		//
		return cmp.Compare(tmp, o)
	case 8:
		tmp := uint64(p.bytes[7])
		tmp += uint64(p.bytes[6]) << 8
		tmp += uint64(p.bytes[5]) << 16
		tmp += uint64(p.bytes[4]) << 24
		tmp += uint64(p.bytes[3]) << 32
		tmp += uint64(p.bytes[2]) << 40
		tmp += uint64(p.bytes[1]) << 48
		tmp += uint64(p.bytes[0]) << 56
		//
		return cmp.Compare(tmp, o)
	default:
		return 1
	}
}

// Cmp implements a comparison by regarding the word as an unsigned integer.
func (p BigEndian) Cmp(o BigEndian) int {
	var (
		lp = len(p.bytes)
		op = len(o.bytes)
	)
	//
	if lp < op {
		return -1
	} else if lp > op {
		return 1
	}
	//
	for i := range lp {
		c := cmp.Compare(p.bytes[i], o.bytes[i])
		if c != 0 {
			return c
		}
	}
	//
	return 0
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

// IsZero implementation for the Word interface
func (p BigEndian) IsZero() bool {
	return len(p.bytes) == 0
}

// PutBytes implementation for Word interface.
func (p BigEndian) PutBytes(bytes []byte) []byte {
	var (
		n = uint(len(bytes))
		m = uint(len(p.bytes))
	)
	// Sanity check space
	if len(bytes) < len(p.bytes) {
		bytes = make([]byte, len(p.bytes))
		n = m
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

// SetBytes implementation for Word interface.
func (p BigEndian) SetBytes(bytes []byte) BigEndian {
	return BigEndian{array.TrimLeadingZeros(bytes)}
}

// SetUint64 implementation for Word interface.
func (p BigEndian) SetUint64(value uint64) BigEndian {
	var bytes [8]byte
	// Write big endian bytes
	binary.BigEndian.PutUint64(bytes[:], value)
	// Trim off leading zeros
	return BigEndian{array.TrimLeadingZeros(bytes[:])}
}

// Uint64 implementation for Word interface.
func (p BigEndian) Uint64() uint64 {
	var val uint64
	//
	switch p.ByteWidth() {
	case 0:
		return 0
	case 1:
		val = uint64(p.bytes[0])
	case 2:
		val = uint64(p.bytes[1])
		val += uint64(p.bytes[0]) << 8
		//
	case 3:
		val = uint64(p.bytes[2])
		val += uint64(p.bytes[1]) << 8
		val += uint64(p.bytes[0]) << 16
		//
	case 4:
		val = uint64(p.bytes[3])
		val += uint64(p.bytes[2]) << 8
		val += uint64(p.bytes[1]) << 16
		val += uint64(p.bytes[0]) << 24
		//
	case 5:
		val = uint64(p.bytes[4])
		val += uint64(p.bytes[3]) << 8
		val += uint64(p.bytes[2]) << 16
		val += uint64(p.bytes[1]) << 24
		val += uint64(p.bytes[0]) << 32
		//
	case 6:
		val = uint64(p.bytes[5])
		val += uint64(p.bytes[4]) << 8
		val += uint64(p.bytes[3]) << 16
		val += uint64(p.bytes[2]) << 24
		val += uint64(p.bytes[1]) << 32
		val += uint64(p.bytes[0]) << 40
		//
	case 7:
		val = uint64(p.bytes[6])
		val += uint64(p.bytes[5]) << 8
		val += uint64(p.bytes[4]) << 16
		val += uint64(p.bytes[3]) << 24
		val += uint64(p.bytes[2]) << 32
		val += uint64(p.bytes[1]) << 40
		val += uint64(p.bytes[0]) << 48
		//
	case 8:
		val = uint64(p.bytes[7])
		val += uint64(p.bytes[6]) << 8
		val += uint64(p.bytes[5]) << 16
		val += uint64(p.bytes[4]) << 24
		val += uint64(p.bytes[3]) << 32
		val += uint64(p.bytes[2]) << 40
		val += uint64(p.bytes[1]) << 48
		val += uint64(p.bytes[0]) << 56
		//
	default:
		// NOTE: we could do better here and return the truncated value.  Just
		// have to be careful to get the right bytes :)
		panic("not uint64")
	}
	//
	return val
}

// Bytes implementation for Word interface.
func (p BigEndian) Bytes() []byte {
	return p.bytes
}

func (p BigEndian) String() string {
	bi := p.AsBigInt()
	//
	return bi.String()
}
