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
package data

import (
	"fmt"
	"math/big"

	"github.com/consensys/go-corset/pkg/util/collection/array"
	"github.com/consensys/go-corset/pkg/util/collection/bit"
	"github.com/consensys/go-corset/pkg/zkc/compiler/ast/symbol"
	"github.com/consensys/go-corset/pkg/zkc/vm/word"
)

// DecodeAll decodes the given set of bytes as big integer values according to
// the given data type(s).  Observe that values are assumed to be packed tightly
// (i.e. without any padding).  Consider the following input byte array:
//
// |  00  |  01  |  02  |  03  |
// +------+------+------+------+
// | 0x31 | 0xf0 | 0x0e | 0x1d |
//
// Then, decoding this into a u4 array will produce the following:
//
// |  0  |  1  |  2  |  3  |  4  |  5  |  6  |  7  |
// +-----+-----+-----+-----+-----+-----+-----+-----+
// | 0x3 | 0x1 | 0xf | 0x0 | 0x0 | 0xe | 0x1 | 0xd |
//
// If the input array is not a multiple of the bitwidth
func DecodeAll[S symbol.Symbol[S]](datatype Type[S], bytes []byte, env Environment[S]) []word.Uint {
	var (
		bitwidth = BitWidthOf(datatype, env)
		// Initially empty buffer which is expanded as necessary to accommodate
		// reading bits of the given data types.
		buffer []byte
	)
	// Decode array into
	values, _ := bit.DecodeArray[[]big.Int](bitwidth, bytes, func(bytes []byte) (ints []big.Int) {
		var reader = bit.NewReader(bytes)
		// Decode the type using the given buffer
		ints, buffer = decodeType(datatype, bitwidth, &reader, buffer, env)
		// Done
		return ints
	})
	// Flattern decoded tuples
	return array.FlatMap[[]big.Int, word.Uint](values, func(ints []big.Int) []word.Uint {
		var words = make([]word.Uint, len(ints))
		//
		for i, v := range ints {
			var ith word.Uint
			//
			words[i] = ith.SetBigInt(&v)
		}
		//
		return words
	})
}

func decodeType[S symbol.Symbol[S]](datatype Type[S], bitwidth uint, reader *bit.Reader, buffer []byte,
	env Environment[S]) ([]big.Int, []byte) {
	//
	switch datatype.(type) {
	case *UnsignedInt[S], *Alias[S]:
		return decodeUnsignedInt(bitwidth, reader, buffer)
	default:
		panic(fmt.Sprintf("unknown type \"%s\"", datatype.String(env)))
	}
}

func decodeUnsignedInt(bitwidth uint, reader *bit.Reader, buffer []byte) ([]big.Int, []byte) {
	var (
		val big.Int
		n   = bit.BytesRequiredFor(bitwidth)
	)
	// Expand buffer to ensure enough space
	buffer = expandBufferAsNeeded(bitwidth, buffer)
	// Read bitwidth bits out
	reader.BigEndianReadInto(bitwidth, buffer)
	// FIXME: broken!!!
	val.SetBytes(buffer[:n])
	//
	return []big.Int{val}, buffer
}

func expandBufferAsNeeded(bitwidth uint, buffer []byte) []byte {
	var n = bit.BytesRequiredFor(bitwidth)
	//
	if uint(len(buffer)) >= n {
		return buffer
	}
	//
	return make([]byte, n)
}
