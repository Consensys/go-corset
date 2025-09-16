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

func Test_AsmUtil_Byte(t *testing.T) {
	test_util.CheckWithFields(t, false, "asm/util/byte", ASM_MAX_PADDING, sc.BLS12_377, sc.KOALABEAR_16)
}

func Test_AsmUtil_BitShift(t *testing.T) {
	test_util.CheckWithFields(t, false, "asm/util/bit_shift", ASM_MAX_PADDING, sc.BLS12_377)
}

func Test_AsmUtil_ByteShift(t *testing.T) {
	test_util.CheckWithFields(t, false, "asm/util/byte_shift", ASM_MAX_PADDING, sc.BLS12_377, sc.KOALABEAR_16)
}

func Test_AsmUtil_FillBytes(t *testing.T) {
	test_util.CheckWithFields(t, false, "asm/util/fill_bytes", ASM_MAX_PADDING, sc.BLS12_377, sc.KOALABEAR_16)
}

func Test_AsmUtil_Log2(t *testing.T) {
	test_util.CheckWithFields(t, false, "asm/util/log2", ASM_MAX_PADDING, sc.BLS12_377, sc.KOALABEAR_16)
}

func Test_AsmUtil_Log256(t *testing.T) {
	test_util.CheckWithFields(t, false, "asm/util/log256", ASM_MAX_PADDING, sc.BLS12_377, sc.KOALABEAR_16)
}
func Test_AsmUtil_Min(t *testing.T) {
	test_util.CheckWithFields(t, false, "asm/util/min", ASM_MAX_PADDING, sc.BLS12_377, sc.KOALABEAR_16)
}

func Test_AsmUtil_SetByte(t *testing.T) {
	test_util.CheckWithFields(t, false, "asm/util/set_byte", ASM_MAX_PADDING, sc.BLS12_377, sc.KOALABEAR_16)
}
