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

func Test_Asm_Dec4(t *testing.T) {
	test_util.CheckWithFields(t, false, "asm/unit/dec4", ASM_MAX_PADDING, sc.BLS12_377)
}

func Test_Asm_Dec251(t *testing.T) {
	test_util.CheckWithFields(t, false, "asm/unit/dec251", ASM_MAX_PADDING, sc.BLS12_377, sc.KOALABEAR_16)
}

func Test_Asm_ParseNonDecimal(t *testing.T) {
	test_util.CheckWithFields(t, false, "asm/unit/parse_nondecimal", ASM_MAX_PADDING, sc.BLS12_377, sc.GF_8209, sc.GF_251)
}

func Test_Asm_Counter(t *testing.T) {
	test_util.CheckWithFields(t, false, "asm/unit/counter", ASM_MAX_PADDING, sc.BLS12_377, sc.GF_8209, sc.GF_251)
}

func Test_Asm_Counter256(t *testing.T) {
	test_util.CheckWithFields(t, false, "asm/unit/counter256", ASM_MAX_PADDING, sc.BLS12_377, sc.KOALABEAR_16)
}

func Test_Asm_FastPow(t *testing.T) {
	test_util.CheckWithFields(t, false, "asm/unit/fast_pow", ASM_MAX_PADDING, sc.BLS12_377)
}

func Test_Asm_Inc(t *testing.T) {
	test_util.CheckWithFields(t, false, "asm/unit/inc", ASM_MAX_PADDING, sc.BLS12_377, sc.GF_8209, sc.GF_251)
}

func Test_Asm_Max14(t *testing.T) {
	test_util.CheckWithFields(t, false, "asm/unit/max14", ASM_MAX_PADDING, sc.BLS12_377, sc.GF_8209, sc.GF_251)
}
func Test_Asm_Max15(t *testing.T) {
	test_util.CheckWithFields(t, false, "asm/unit/max15", ASM_MAX_PADDING, sc.BLS12_377, sc.GF_8209, sc.GF_251)
}

func Test_Asm_Max16(t *testing.T) {
	test_util.CheckWithFields(t, false, "asm/unit/max16", ASM_MAX_PADDING, sc.BLS12_377, sc.GF_8209, sc.GF_251)
}

func Test_Asm_Max256(t *testing.T) {
	test_util.CheckWithFields(t, false, "asm/unit/max256", ASM_MAX_PADDING, sc.BLS12_377, sc.KOALABEAR_16)
}

func Test_Asm_MixedSmall(t *testing.T) {
	test_util.CheckWithFields(t, false, "asm/unit/mixed_small", ASM_MAX_PADDING, sc.BLS12_377)
}
func Test_Asm_MultiLine(t *testing.T) {
	test_util.CheckWithFields(t, false, "asm/unit/multiline", ASM_MAX_PADDING, sc.BLS12_377)
}

func Test_Asm_MixedLarge(t *testing.T) {
	test_util.CheckWithFields(t, false, "asm/unit/mixed_large", ASM_MAX_PADDING, sc.BLS12_377)
}

func Test_Asm_SlowPow(t *testing.T) {
	test_util.CheckWithFields(t, false, "asm/unit/slow_pow", ASM_MAX_PADDING, sc.BLS12_377)
}

func Test_Asm_Sub(t *testing.T) {
	test_util.CheckWithFields(t, false, "asm/unit/sub", ASM_MAX_PADDING, sc.BLS12_377, sc.GF_8209, sc.GF_251)
}

func Test_Asm_SimpleOnCurve(t *testing.T) {
	// Check(t, false, "asm/unit/simple_on_curve")
	// To be replaced once splitting algorithm is available
	test_util.CheckWithFields(t, false, "asm/unit/simple_on_curve_u16", ASM_MAX_PADDING, sc.BLS12_377)
}

func Test_Asm_RecPow(t *testing.T) {
	test_util.CheckWithFields(t, false, "asm/unit/rec_pow", ASM_MAX_PADDING, sc.BLS12_377)
}

func Test_Asm_Gf251(t *testing.T) {
	test_util.Check(t, false, "asm/unit/gf251")
}
