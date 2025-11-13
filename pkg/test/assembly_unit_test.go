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
	"github.com/consensys/go-corset/pkg/test/util"
)

func Test_AsmUnit_BitShift(t *testing.T) {
	util.CheckWithFields(t, false, "asm/unit/bit_shift", util.ASM_MAX_PADDING, sc.BLS12_377)
}

func Test_AsmUnit_ByteShift(t *testing.T) {
	util.CheckWithFields(t, false, "asm/unit/byte_shift", util.ASM_MAX_PADDING, sc.BLS12_377, sc.KOALABEAR_16)
}

func Test_AsmUnit_Dec4(t *testing.T) {
	util.CheckWithFields(t, false, "asm/unit/dec4", util.ASM_MAX_PADDING, sc.BLS12_377)
}

func Test_AsmUnit_Dec251(t *testing.T) {
	util.CheckWithFields(t, false, "asm/unit/dec251", util.ASM_MAX_PADDING, sc.BLS12_377, sc.KOALABEAR_16)
}

func Test_AsmUnit_Diamond(t *testing.T) {
	util.CheckWithFields(t, false, "asm/unit/diamond", util.ASM_MAX_PADDING, sc.BLS12_377, sc.KOALABEAR_16)
}

func Test_AsmUnit_ParseNonDecimal(t *testing.T) {
	util.CheckWithFields(t, false, "asm/unit/parse_nondecimal", util.ASM_MAX_PADDING, sc.BLS12_377, sc.GF_8209, sc.GF_251)
}

func Test_AsmUnit_Counter(t *testing.T) {
	util.CheckWithFields(t, false, "asm/unit/counter", util.ASM_MAX_PADDING, sc.BLS12_377, sc.KOALABEAR_16, sc.GF_8209,
		sc.GF_251)
}

func Test_AsmUnit_Counter256(t *testing.T) {
	util.CheckWithFields(t, false, "asm/unit/counter256", util.ASM_MAX_PADDING, sc.BLS12_377)
}

func Test_AsmUnit_Diff(t *testing.T) {
	util.CheckWithFields(t, false, "asm/unit/diff", util.ASM_MAX_PADDING, sc.BLS12_377, sc.KOALABEAR_16,
		sc.GF_8209, sc.GF_251)
}

func Test_AsmUnit_FastPow(t *testing.T) {
	util.CheckWithFields(t, false, "asm/unit/fast_pow", util.ASM_MAX_PADDING, sc.BLS12_377)
}
func Test_AsmUnit_Gf251(t *testing.T) {
	util.Check(t, false, "asm/unit/gf251")
}

func Test_AsmUnit_Inc(t *testing.T) {
	util.CheckWithFields(t, false, "asm/unit/inc", util.ASM_MAX_PADDING, sc.BLS12_377, sc.GF_8209, sc.GF_251)
}

func Test_AsmUnit_Max(t *testing.T) {
	util.CheckWithFields(t, false, "asm/unit/max", util.ASM_MAX_PADDING, sc.BLS12_377, sc.KOALABEAR_16)
}

func Test_AsmUnit_MixedSmall(t *testing.T) {
	util.CheckWithFields(t, false, "asm/unit/mixed_small", util.ASM_MAX_PADDING, sc.BLS12_377)
}
func Test_AsmUnit_MultiLine(t *testing.T) {
	util.CheckWithFields(t, false, "asm/unit/multiline", util.ASM_MAX_PADDING, sc.BLS12_377)
}

func Test_AsmUnit_FnCallLarge(t *testing.T) {
	util.CheckWithFields(t, false, "asm/unit/fncall_large", util.ASM_MAX_PADDING, sc.BLS12_377)
}

func Test_AsmUnit_MixedLarge(t *testing.T) {
	util.CheckWithFields(t, false, "asm/unit/mixed_large", util.ASM_MAX_PADDING, sc.BLS12_377)
}

func Test_AsmUnit_SlowPow(t *testing.T) {
	util.CheckWithFields(t, false, "asm/unit/slow_pow", util.ASM_MAX_PADDING, sc.BLS12_377)
}

func Test_AsmUnit_Sub(t *testing.T) {
	util.CheckWithFields(t, false, "asm/unit/sub", util.ASM_MAX_PADDING, sc.BLS12_377, sc.KOALABEAR_16, sc.GF_8209,
		sc.GF_251)
}

func Test_AsmUnit_RecPow(t *testing.T) {
	util.CheckWithFields(t, false, "asm/unit/rec_pow", util.ASM_MAX_PADDING, sc.BLS12_377)
}

func Test_Asm_And(t *testing.T) {
	util.Check(t, false, "asm/unit/and")
}
