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
package instruction

import "math/big"

var zero = *big.NewInt(0)
var one = *big.NewInt(1)

// Write the value to a given set of target registers, splitting its bits as
// necessary.  The target registers are given with the least significant first.
func writeTargetRegisters(targets []uint, regs []big.Int, widths []uint, value big.Int) {
	var (
		offset uint = 0
	)
	//
	for _, reg := range targets {
		width := widths[reg]
		regs[reg] = readBitSlice(offset, width, value)
		offset += width
	}
}

func readBitSlice(offset uint, width uint, value big.Int) big.Int {
	var slice big.Int
	//
	for i := 0; uint(i) < width; i++ {
		// Read appropriate bit
		bit := value.Bit(i + int(offset))
		// set appropriate bit
		slice.SetBit(&slice, i, bit)
	}
	//
	return slice
}
