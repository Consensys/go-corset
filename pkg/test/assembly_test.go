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
package test

import (
	"fmt"
	"math/big"
	"testing"

	"github.com/consensys/gnark-crypto/ecc/bls12-377/fr"
)

func Test_01(t *testing.T) {
	var (
		x        *big.Int = fr.Modulus()
		c        *big.Int = big.NewInt(2)
		msb, lsb big.Int
	)
	// Determine 2^128
	c.Exp(c, big.NewInt(128), nil)
	//
	msb.Div(x, c)
	lsb.Mod(x, c)
	//
	fmt.Printf("MOD=0x%s (%s)\n", x.Text(16), x.String())
	fmt.Printf("MSB=0x%s (%s)\n", msb.Text(16), msb.String())
	fmt.Printf("LSB=0x%s (%s)\n", lsb.Text(16), lsb.String())
}

// Recusion
//
// func Test_Asm_Byte(t *testing.T) {
// 	Check(t, false, "asm/byte")
// }

func Test_Asm_Counter(t *testing.T) {
	Check(t, false, "asm/counter")
}

func Test_Asm_FastPow(t *testing.T) {
	Check(t, false, "asm/fast_pow")
}

func Test_Asm_Max(t *testing.T) {
	Check(t, false, "asm/max")
}

func Test_Asm_Max256(t *testing.T) {
	Check(t, false, "asm/max256")
}
func Test_Asm_MixedSmall(t *testing.T) {
	Check(t, false, "asm/mixed_small")
}

func Test_Asm_MixedLarge(t *testing.T) {
	Check(t, false, "asm/mixed_large")
}

func Test_Asm_SlowPow(t *testing.T) {
	Check(t, false, "asm/slow_pow")
}

// Recusion
//
// func Test_Asm_RecPow(t *testing.T) {
// 	Check(t, false, "asm/rec_pow")
// }

// Recusion
//
// func Test_Asm_Shift(t *testing.T) {
// 	Check(t, false, "asm/shift")
// }

// Field Element Out-Of-Bounds
//
// func Test_Asm_Wcp(t *testing.T) {
// 	Check(t, false, "asm/wcp")
// }
