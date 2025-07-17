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
package field

import (
	"encoding/binary"

	"github.com/consensys/gnark-crypto/ecc/bls12-377/fr"
	"github.com/consensys/go-corset/pkg/util/collection/array"
	"github.com/consensys/go-corset/pkg/util/word"
)

// ToBigEndianByteArray converts an array of field elements into an array of
// byte chunks in big endian form.
func ToBigEndianByteArray[P word.Pool[uint, word.BigEndian]](arr FrArray, pool P) array.Array[word.BigEndian] {
	var builder = word.NewArray(arr.Len(), arr.BitWidth(), pool)
	//
	for i := range arr.Len() {
		var (
			ith       = arr.Get(i)
			ith_bytes = ith.Bytes()
			trimmed   = ith_bytes[:]
		)
		//
		builder.Set(i, word.NewBigEndian(trimmed))
	}
	//
	return builder.Build()
}

// Pow takes a given value to the power n.
func Pow(val *fr.Element, n uint64) {
	if n == 0 {
		val.SetOne()
	} else if n > 1 {
		m := n / 2
		// Check for odd case
		if n%2 == 1 {
			var tmp fr.Element
			// Clone value
			tmp.Set(val)
			Pow(val, m)
			val.Square(val)
			val.Mul(val, &tmp)
		} else {
			// Even case is easy
			Pow(val, m)
			val.Square(val)
		}
	}
}

// FrElementToBytes converts a given field element into a slice of 32 bytes.
func FrElementToBytes(element fr.Element) [32]byte {
	// Each fr.Element is 4 x 64bit words.
	var bytes [32]byte
	// Copy over each element
	binary.BigEndian.PutUint64(bytes[:], element[0])
	binary.BigEndian.PutUint64(bytes[8:], element[1])
	binary.BigEndian.PutUint64(bytes[16:], element[2])
	binary.BigEndian.PutUint64(bytes[24:], element[3])
	// Done
	return bytes
}
