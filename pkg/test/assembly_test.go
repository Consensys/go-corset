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
	"testing"

	sc "github.com/consensys/go-corset/pkg/schema"
	test_util "github.com/consensys/go-corset/pkg/test/util"
)

func Test_Asm_Add(t *testing.T) {
	test_util.CheckWithFields(t, false, "asm/add", sc.BLS12_377, sc.KOALABEAR_16)
}

// Recusion
//
//	func Test_Asm_Byte(t *testing.T) {
//		test_util.Check(t, false, "asm/byte")
//	}
func Test_Asm_Dec4(t *testing.T) {
	test_util.CheckWithFields(t, false, "asm/dec4", sc.BLS12_377)
}

// See #1081
// func Test_Asm_ParseNonDecimal(t *testing.T) {
// 	test_util.CheckWithFields(t, false, "asm/parse_nondecimal", sc.BLS12_377, sc.GF_8209, sc.GF_251)
// }

func Test_Asm_Counter(t *testing.T) {
	test_util.CheckWithFields(t, false, "asm/counter", sc.BLS12_377, sc.GF_8209, sc.GF_251)
}

func Test_Asm_Counter256(t *testing.T) {
	test_util.CheckWithFields(t, false, "asm/counter256", sc.BLS12_377, sc.KOALABEAR_16)
}

func Test_Asm_FastPow(t *testing.T) {
	test_util.Check(t, false, "asm/fast_pow")
}

func Test_Asm_Gas(t *testing.T) {
	test_util.CheckWithFields(t, false, "asm/gas", sc.BLS12_377, sc.GF_8209, sc.GF_251)
}

func Test_Asm_Inc(t *testing.T) {
	test_util.CheckWithFields(t, false, "asm/inc", sc.BLS12_377, sc.GF_8209, sc.GF_251)
}

func Test_Asm_Log256(t *testing.T) {
	test_util.Check(t, false, "asm/log256")
}

func Test_Asm_Max14(t *testing.T) {
	test_util.CheckWithFields(t, false, "asm/max14", sc.BLS12_377, sc.GF_8209, sc.GF_251)
}
func Test_Asm_Max15(t *testing.T) {
	test_util.CheckWithFields(t, false, "asm/max15", sc.BLS12_377, sc.GF_8209, sc.GF_251)
}

func Test_Asm_Max16(t *testing.T) {
	test_util.CheckWithFields(t, false, "asm/max16", sc.BLS12_377, sc.GF_8209, sc.GF_251)
}

func Test_Asm_Max256(t *testing.T) {
	test_util.CheckWithFields(t, false, "asm/max256", sc.BLS12_377, sc.KOALABEAR_16)
}

func Test_Asm_MixedSmall(t *testing.T) {
	test_util.Check(t, false, "asm/mixed_small")
}

func Test_Asm_MixedLarge(t *testing.T) {
	test_util.Check(t, false, "asm/mixed_large")
}

func Test_Asm_SlowPow(t *testing.T) {
	test_util.Check(t, false, "asm/slow_pow")
}

func Test_Asm_Sub(t *testing.T) {
	test_util.CheckWithFields(t, false, "asm/sub", sc.BLS12_377, sc.GF_8209, sc.GF_251)
}

// See #1081
// func Test_Asm_SimpleOnCurve(t *testing.T) {
// 	// Check(t, false, "asm/simple_on_curve")
// 	// To be replaced once splitting algorithm is available
// 	test_util.Check(t, false, "asm/simple_on_curve_u16")
// }

func Test_Asm_Trim(t *testing.T) {
	test_util.CheckWithFields(t, false, "asm/trim", sc.BLS12_377, sc.KOALABEAR_16)
}

// Recursion
//
// func Test_Asm_RecPow(t *testing.T) {
// 	test_util.Check(t, false, "asm/rec_pow")
// }

// Recursion
//
// func Test_Asm_Shift(t *testing.T) {
// 	test_util.Check(t, false, "asm/shift")
// }

// Field Element Out-Of-Bounds
func Test_Asm_Wcp(t *testing.T) {
	test_util.CheckWithFields(t, false, "asm/wcp", sc.BLS12_377, sc.KOALABEAR_16)
}
