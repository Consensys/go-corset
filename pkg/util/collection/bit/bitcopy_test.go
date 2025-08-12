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

import (
	"slices"
	"testing"
)

// dstOffset = 0

func Test_BitCopy_00(t *testing.T) {
	checkBitCopy(t, 0, 0, 8, []byte{0b1001_1111, 0b0000_0000}, []byte{0b1001_1111})
}
func Test_BitCopy_01(t *testing.T) {
	checkBitCopy(t, 1, 0, 8, []byte{0b1001_1111, 0b0000_0000}, []byte{0b0100_1111})
}
func Test_BitCopy_02(t *testing.T) {
	checkBitCopy(t, 2, 0, 8, []byte{0b1001_1111, 0b0000_0000}, []byte{0b0010_0111})
}
func Test_BitCopy_03(t *testing.T) {
	checkBitCopy(t, 3, 0, 8, []byte{0b1001_1111, 0b0000_0000}, []byte{0b0001_0011})
}
func Test_BitCopy_04(t *testing.T) {
	checkBitCopy(t, 4, 0, 10, []byte{0b1001_1111, 0b0000_0101}, []byte{0b0101_1001, 0b0000_0000})
}
func Test_BitCopy_05(t *testing.T) {
	checkBitCopy(t, 4, 0, 10, []byte{0b1001_1111, 0b0010_0101}, []byte{0b0101_1001, 0b0000_0010})
}
func Test_BitCopy_06(t *testing.T) {
	checkBitCopy(t, 8, 0, 8, []byte{0b0000_0000, 0b1001_1111}, []byte{0b1001_1111})
}
func Test_BitCopy_07(t *testing.T) {
	checkBitCopy(t, 8, 0, 10, []byte{0b0000_0000, 0b1001_1111, 0b0000_0010}, []byte{0b1001_1111, 0b0000_0010})
}
func Test_BitCopy_08(t *testing.T) {
	checkBitCopy(t, 8, 0, 18,
		[]byte{0b0000_0000, 0b1001_1111, 0b0101_0101, 0b0000_0010},
		[]byte{0b1001_1111, 0b0101_0101, 0b0000_0010})
}

// dstOffset = 1
func Test_BitCopy_10(t *testing.T) {
	checkBitCopy(t, 0, 1, 8, []byte{0b1001_1111, 0b0000_0000}, []byte{0b0011_1110, 0b0000_0001})
}
func Test_BitCopy_11(t *testing.T) {
	checkBitCopy(t, 1, 1, 8, []byte{0b1001_1111, 0b0000_0000}, []byte{0b1001_1110, 0b0000_0000})
}
func Test_BitCopy_12(t *testing.T) {
	checkBitCopy(t, 2, 1, 8, []byte{0b1001_1111, 0b0000_0000}, []byte{0b0100_1110, 0b0000_0000})
}

func checkBitCopy(t *testing.T, srcOffset uint, dstOffset uint, nbits uint, src []byte, expected []byte) {
	//
	t.Parallel()
	//
	var (
		buf = make([]byte, len(src))
		n   = (dstOffset + nbits) / 8
	)
	//
	if (dstOffset+nbits)%8 != 0 {
		n++
	}
	//
	Copy(src, srcOffset, buf, dstOffset, nbits)
	// Extract read bytes
	actual := buf[:n]
	//
	if !slices.Equal(expected, actual) {
		t.Errorf("expected %v, received %v", expected, actual)
	}
}
