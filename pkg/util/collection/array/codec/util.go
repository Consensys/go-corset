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
package codec

import (
	"encoding/binary"
	"fmt"
)

func countSparseRows8(byteWidth uint, bytes []byte) uint64 {
	var (
		count uint64
		n     = 1 + byteWidth
	)
	// Sanity check
	if uint(len(bytes))%n != 0 {
		panic(fmt.Sprintf("invalid length for sparse data (%d)", len(bytes)))
	}
	// Count data
	for i := byteWidth; i < uint(len(bytes)); i += n {
		// Read ith count
		count += uint64(bytes[i])
	}
	//
	return count
}

func countSparseRows16(byteWidth uint, bytes []byte) uint64 {
	var (
		count uint64
		n     = 2 + byteWidth
	)
	// Sanity check
	if uint(len(bytes))%n != 0 {
		panic(fmt.Sprintf("invalid length for sparse data (%d)", len(bytes)))
	}
	// Count data
	for i := byteWidth; i < uint(len(bytes)); i += n {
		// Read ith count
		count += uint64(binary.BigEndian.Uint16(bytes[i : i+2]))
	}
	//
	return count
}

func countSparseRows32(byteWidth uint, bytes []byte) uint64 {
	var (
		count uint64
		n     = 4 + byteWidth
	)
	// Sanity check
	if uint(len(bytes))%n != 0 {
		panic(fmt.Sprintf("invalid length for sparse data (%d)", len(bytes)))
	}
	// Count data
	for i := byteWidth; i < uint(len(bytes)); i += n {
		// Read ith count
		count += uint64(binary.BigEndian.Uint32(bytes[i : i+4]))
	}
	//
	return count
}
