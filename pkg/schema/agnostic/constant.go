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
package agnostic

import (
	"math/big"
)

var (
	zero big.Int
	one  big.Int
)

// SplitConstant splits a given constant into a number of "limbs". For example,
// consider splitting the constant 0x7b2d into 8bit limbs.  Then, this function
// returns the array [0x2d,0x7b].  Observe that the least significant limb is
// always returned first (i.e. at index zero in the resulting array).
func SplitConstant(constant big.Int, limbWidths ...uint) []big.Int {
	var (
		acc   big.Int
		limbs []big.Int = make([]big.Int, len(limbWidths))
	)
	// Clone constant
	acc.Set(&constant)
	//
	for i := 0; acc.Cmp(&zero) != 0; i++ {
		var (
			width = limbWidths[i]
			limb  big.Int
			bound *big.Int = big.NewInt(2)
		)
		// Determine upper bound
		bound.Exp(bound, big.NewInt(int64(width)), nil)
		// Pull of excess
		limb.Mod(&acc, bound)
		limbs[i] = limb
		// Shift down
		acc.Rsh(&acc, width)
	}
	//
	return limbs
}

func init() {
	zero = *big.NewInt(0)
	one = *big.NewInt(1)
}
