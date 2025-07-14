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
package bit

import "fmt"

func bitCopy(src []byte, srcOffset uint, dst []byte, dstOffset uint, nbits uint) {
	fmt.Printf("BIT COPY: %d => %d (%d bits)\n", srcOffset, dstOffset, nbits)
	// TODO: this could be significantly optimised.
	for i := range nbits {
		ith := readBit(src, srcOffset+i)
		writeBit(ith, dst, dstOffset+i)
	}
}

func readBit(src []byte, bitoffset uint) bool {
	var (
		byte = bitoffset / 8
		bit  = bitoffset % 8
		mask = uint8(1) << bit
	)
	//
	return src[byte]&mask != 0
}

func writeBit(val bool, src []byte, bitoffset uint) {
	var (
		byte = bitoffset / 8
		bit  = bitoffset % 8
		mask = uint8(1) << bit
	)
	//
	if val {
		// set bit
		src[byte] = src[byte] | mask
	} else {
		// Clear bit
		src[byte] = src[byte] & ^mask
	}
}
