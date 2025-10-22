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

	"github.com/consensys/go-corset/pkg/schema/register"
)

var (
	zero     big.Int
	one      big.Int
	minusOne big.Int
)

// CombinedWidthOfRegisters returns the combined bitwidth of all limbs.  For example,
// suppose we have three limbs: x:u8, y:u8, z:u11.  Then the combined width is
// 8+8+11=27.
func CombinedWidthOfRegisters(mapping register.Map, registers ...register.LimbId) uint {
	var (
		width uint
	)
	//
	for _, rid := range registers {
		width += mapping.Register(rid).Width
	}
	//
	return width
}

func init() {
	zero = *big.NewInt(0)
	one = *big.NewInt(1)
	minusOne = *big.NewInt(-1)
}
