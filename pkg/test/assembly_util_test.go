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

	"github.com/consensys/go-corset/pkg/test/util"
	"github.com/consensys/go-corset/pkg/util/field"
)

func Test_AsmUtil_Byte(t *testing.T) {
	util.CheckWithFields(t, false, "asm/util/byte", util.ASM_MAX_PADDING, field.BLS12_377, field.KOALABEAR_16)
}

func Test_AsmUtil_BitShr(t *testing.T) {
	util.CheckWithFields(t, false, "asm/util/bit_shr", util.ASM_MAX_PADDING, field.BLS12_377)
}

func Test_AsmUtil_BitShl(t *testing.T) {
	util.CheckWithFields(t, false, "asm/util/bit_shl", util.ASM_MAX_PADDING, field.BLS12_377, field.KOALABEAR_16)
}

func Test_AsmUtil_FillBytes(t *testing.T) {
	util.CheckWithFields(t, false, "asm/util/fill_bytes", util.ASM_MAX_PADDING, field.BLS12_377, field.KOALABEAR_16)
}

func Test_AsmUtil_FirstByte(t *testing.T) {
	util.CheckWithFields(t, false, "asm/util/first_byte", util.ASM_MAX_PADDING, field.BLS12_377, field.KOALABEAR_16)
}

func Test_AsmUtil_Log2(t *testing.T) {
	util.CheckWithFields(t, false, "asm/util/log2", util.ASM_MAX_PADDING, field.BLS12_377, field.KOALABEAR_16)
}

func Test_AsmUtil_Log256(t *testing.T) {
	util.CheckWithFields(t, false, "asm/util/log256", util.ASM_MAX_PADDING, field.BLS12_377, field.KOALABEAR_16)
}
func Test_AsmUtil_Min(t *testing.T) {
	util.CheckWithFields(t, false, "asm/util/min", util.ASM_MAX_PADDING, field.BLS12_377, field.KOALABEAR_16)
}

func Test_AsmUtil_SetByte(t *testing.T) {
	util.CheckWithFields(t, false, "asm/util/set_byte", util.ASM_MAX_PADDING, field.BLS12_377, field.KOALABEAR_16)
}

func Test_AsmUtil_Signextend(t *testing.T) {
	util.CheckWithFields(t, false, "asm/util/signextend", util.ASM_MAX_PADDING, field.BLS12_377, field.KOALABEAR_16)
}
