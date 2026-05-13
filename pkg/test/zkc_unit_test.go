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
	"github.com/consensys/go-corset/pkg/zkc/compiler/codegen"
)

// ===================================================================
// Basic Tests
// ===================================================================

func Test_ZkcUnit_Basic_01(t *testing.T) {
	checkZkcUnit(t, "zkc/unit/basic_01", field.BLS12_377, field.KOALABEAR_16)
}

func Test_ZkcUnit_Basic_02(t *testing.T) {
	checkZkcUnit(t, "zkc/unit/basic_02", field.BLS12_377, field.KOALABEAR_16)
}

func Test_ZkcUnit_Basic_03(t *testing.T) {
	checkZkcUnit(t, "zkc/unit/basic_03", field.BLS12_377, field.KOALABEAR_16)
}
func Test_ZkcUnit_Basic_04(t *testing.T) {
	checkZkcUnit(t, "zkc/unit/basic_04", field.BLS12_377, field.KOALABEAR_16)
}

func Test_ZkcUnit_Basic_05(t *testing.T) {
	checkZkcUnit(t, "zkc/unit/basic_05", field.BLS12_377, field.KOALABEAR_16)
}

func Test_ZkcUnit_Basic_06(t *testing.T) {
	checkZkcUnit(t, "zkc/unit/basic_06", field.BLS12_377, field.KOALABEAR_16)
}
func Test_ZkcUnit_Basic_07(t *testing.T) {
	checkZkcUnit(t, "zkc/unit/basic_07", field.BLS12_377, field.KOALABEAR_16)
}

func Test_ZkcUnit_Basic_08(t *testing.T) {
	checkZkcUnit(t, "zkc/unit/basic_08", field.BLS12_377, field.KOALABEAR_16)
}

func Test_ZkcUnit_Basic_09(t *testing.T) {
	checkZkcUnit(t, "zkc/unit/basic_09", field.BLS12_377, field.KOALABEAR_16)
}

func Test_ZkcUnit_Basic_10(t *testing.T) {
	checkZkcUnit(t, "zkc/unit/basic_10", field.BLS12_377, field.KOALABEAR_16)
}

func Test_ZkcUnit_Basic_11(t *testing.T) {
	checkZkcUnit(t, "zkc/unit/basic_11", field.BLS12_377, field.KOALABEAR_16)
}

func Test_ZkcUnit_Basic_12(t *testing.T) {
	checkZkcUnit(t, "zkc/unit/basic_12", field.BLS12_377, field.KOALABEAR_16)
}

func Test_ZkcUnit_Basic_13(t *testing.T) {
	checkZkcUnit(t, "zkc/unit/basic_13", field.BLS12_377, field.KOALABEAR_16)
}

func Test_ZkcUnit_Basic_14(t *testing.T) {
	checkZkcUnit(t, "zkc/unit/basic_14", field.BLS12_377, field.KOALABEAR_16)
}

func Test_ZkcUnit_Basic_15(t *testing.T) {
	checkZkcUnit(t, "zkc/unit/basic_15", field.BLS12_377, field.KOALABEAR_16)
}

func Test_ZkcUnit_Basic_16(t *testing.T) {
	checkZkcUnit(t, "zkc/unit/basic_16", field.BLS12_377, field.KOALABEAR_16)
}

func Test_ZkcUnit_Basic_17(t *testing.T) {
	checkZkcUnit(t, "zkc/unit/basic_17", field.BLS12_377, field.KOALABEAR_16)
}

func Test_ZkcUnit_Basic_18(t *testing.T) {
	checkZkcUnit(t, "zkc/unit/basic_18", field.BLS12_377, field.KOALABEAR_16)
}

func Test_ZkcUnit_Basic_19(t *testing.T) {
	checkZkcUnit(t, "zkc/unit/basic_19", field.BLS12_377, field.KOALABEAR_16)
}

func Test_ZkcUnit_Basic_20(t *testing.T) {
	checkZkcUnit(t, "zkc/unit/basic_20", field.BLS12_377, field.KOALABEAR_16)
}
func Test_ZkcUnit_Basic_21(t *testing.T) {
	checkZkcUnit(t, "zkc/unit/basic_21", field.BLS12_377, field.KOALABEAR_16)
}

func Test_ZkcUnit_Basic_22(t *testing.T) {
	checkZkcUnit(t, "zkc/unit/basic_22", field.BLS12_377, field.KOALABEAR_16)
}

func Test_ZkcUnit_Basic_23(t *testing.T) {
	checkZkcUnit(t, "zkc/unit/basic_23", field.BLS12_377, field.KOALABEAR_16)
}

func Test_ZkcUnit_Basic_24(t *testing.T) {
	checkZkcUnit(t, "zkc/unit/basic_24", field.BLS12_377, field.KOALABEAR_16)
}

func Test_ZkcUnit_Basic_25(t *testing.T) {
	checkZkcUnit(t, "zkc/unit/basic_25", field.BLS12_377, field.KOALABEAR_16)
}
func Test_ZkcUnit_Basic_26(t *testing.T) {
	checkZkcUnit(t, "zkc/unit/basic_26", field.BLS12_377, field.KOALABEAR_16)
}

func Test_ZkcUnit_Basic_27(t *testing.T) {
	checkZkcUnit(t, "zkc/unit/basic_27", field.BLS12_377, field.KOALABEAR_16)
}

func Test_ZkcUnit_Basic_28(t *testing.T) {
	checkZkcUnit(t, "zkc/unit/basic_28", field.BLS12_377, field.KOALABEAR_16)
}

func Test_ZkcUnit_Basic_29(t *testing.T) {
	checkZkcUnit(t, "zkc/unit/basic_29", field.BLS12_377, field.KOALABEAR_16)
}

func Test_ZkcUnit_Basic_30(t *testing.T) {
	checkZkcUnit(t, "zkc/unit/basic_30", field.BLS12_377, field.KOALABEAR_16)
}

func Test_ZkcUnit_Basic_31(t *testing.T) {
	checkZkcUnit(t, "zkc/unit/basic_31", field.BLS12_377, field.KOALABEAR_16)
}

// ===================================================================
// If-Else-If Tests
// ===================================================================

func Test_ZkcUnit_IfElse_01(t *testing.T) {
	checkZkcUnit(t, "zkc/unit/ifelse_01", field.BLS12_377, field.KOALABEAR_16)
}

func Test_ZkcUnit_IfElse_02(t *testing.T) {
	checkZkcUnit(t, "zkc/unit/ifelse_02", field.BLS12_377, field.KOALABEAR_16)
}

func Test_ZkcUnit_IfElse_03(t *testing.T) {
	checkZkcUnit(t, "zkc/unit/ifelse_03", field.BLS12_377, field.KOALABEAR_16)
}

func Test_ZkcUnit_IfElse_04(t *testing.T) {
	checkZkcUnit(t, "zkc/unit/ifelse_04", field.BLS12_377, field.KOALABEAR_16)
}

func Test_ZkcUnit_IfElse_05(t *testing.T) {
	checkZkcUnit(t, "zkc/unit/ifelse_05", field.BLS12_377, field.KOALABEAR_16)
}

func Test_ZkcUnit_IfElse_06(t *testing.T) {
	checkZkcUnit(t, "zkc/unit/ifelse_06", field.BLS12_377, field.KOALABEAR_16)
}

func Test_ZkcUnit_IfElse_07(t *testing.T) {
	checkZkcUnit(t, "zkc/unit/ifelse_07", field.BLS12_377, field.KOALABEAR_16)
}

// ===================================================================
// Constant Tests
// ===================================================================

func Test_ZkcUnit_Const_01(t *testing.T) {
	checkZkcUnit(t, "zkc/unit/const_01", field.BLS12_377, field.KOALABEAR_16)
}

func Test_ZkcUnit_Const_02(t *testing.T) {
	checkZkcUnit(t, "zkc/unit/const_02", field.BLS12_377, field.KOALABEAR_16)
}

func Test_ZkcUnit_Const_03(t *testing.T) {
	checkZkcUnit(t, "zkc/unit/const_03", field.BLS12_377, field.KOALABEAR_16)
}

func Test_ZkcUnit_Const_04(t *testing.T) {
	checkZkcUnit(t, "zkc/unit/const_04", field.BLS12_377, field.KOALABEAR_16)
}

// ===================================================================
// Fixed-size array Tests
// ===================================================================

func Test_ZkcUnit_FixedArray_01(t *testing.T) {
	checkZkcUnit(t, "zkc/unit/fixed_array_01", field.BLS12_377, field.KOALABEAR_16)
}

func Test_ZkcUnit_FixedArray_02(t *testing.T) {
	checkZkcUnit(t, "zkc/unit/fixed_array_02", field.BLS12_377, field.KOALABEAR_16)
}

func Test_ZkcUnit_FixedArray_03(t *testing.T) {
	checkZkcUnit(t, "zkc/unit/fixed_array_03", field.BLS12_377, field.KOALABEAR_16)
}

func Test_ZkcUnit_FixedArray_04(t *testing.T) {
	checkZkcUnit(t, "zkc/unit/fixed_array_04", field.BLS12_377, field.KOALABEAR_16)
}

func Test_ZkcUnit_FixedArray_05(t *testing.T) {
	checkZkcUnit(t, "zkc/unit/fixed_array_05", field.BLS12_377, field.KOALABEAR_16)
}

func Test_ZkcUnit_FixedArray_06(t *testing.T) {
	checkZkcUnit(t, "zkc/unit/fixed_array_06", field.BLS12_377, field.KOALABEAR_16)
}

func Test_ZkcUnit_FixedArray_07(t *testing.T) {
	checkZkcUnit(t, "zkc/unit/fixed_array_07", field.BLS12_377, field.KOALABEAR_16)
}

func Test_ZkcUnit_FixedArray_08(t *testing.T) {
	checkZkcUnit(t, "zkc/unit/fixed_array_08", field.BLS12_377, field.KOALABEAR_16)
}

func Test_ZkcUnit_FixedArray_09(t *testing.T) {
	checkZkcUnit(t, "zkc/unit/fixed_array_09", field.BLS12_377, field.KOALABEAR_16)
}

// see #1711
// func Test_ZkcUnit_FixedArray_10(t *testing.T) {
// 	checkZkcUnit(t, "zkc/unit/fixed_array_10", field.BLS12_377, field.KOALABEAR_16)
// }

// see #1711
// func Test_ZkcUnit_FixedArray_11(t *testing.T) {
// 	checkZkcUnit(t, "zkc/unit/fixed_array_11", field.BLS12_377, field.KOALABEAR_16)
// }

// see #1711
// func Test_ZkcUnit_FixedArray_12(t *testing.T) {
// 	checkZkcUnit(t, "zkc/unit/fixed_array_12", field.BLS12_377, field.KOALABEAR_16)
// }

// ===================================================================
// Type Tests
// ===================================================================

func Test_ZkcUnit_Type_01(t *testing.T) {
	checkZkcUnit(t, "zkc/unit/type_01", field.BLS12_377, field.KOALABEAR_16)
}

func Test_ZkcUnit_Type_02(t *testing.T) {
	checkZkcUnit(t, "zkc/unit/type_02", field.BLS12_377, field.KOALABEAR_16)
}

func Test_ZkcUnit_Type_03(t *testing.T) {
	checkZkcUnit(t, "zkc/unit/type_03", field.BLS12_377, field.KOALABEAR_16)
}

func Test_ZkcUnit_Type_04(t *testing.T) {
	checkZkcUnit(t, "zkc/unit/type_04", field.BLS12_377, field.KOALABEAR_16)
}

func Test_ZkcUnit_Type_05(t *testing.T) {
	checkZkcUnit(t, "zkc/unit/type_05", field.BLS12_377, field.KOALABEAR_16)
}

func Test_ZkcUnit_Type_06(t *testing.T) {
	checkZkcUnit(t, "zkc/unit/type_06", field.BLS12_377, field.KOALABEAR_16)
}

func Test_ZkcUnit_Type_07(t *testing.T) {
	checkZkcUnit(t, "zkc/unit/type_07", field.BLS12_377, field.KOALABEAR_16)
}

func Test_ZkcUnit_Type_08(t *testing.T) {
	checkZkcUnit(t, "zkc/unit/type_08", field.BLS12_377, field.KOALABEAR_16)
}

func Test_ZkcUnit_Type_09(t *testing.T) {
	checkZkcUnit(t, "zkc/unit/type_09", field.BLS12_377, field.KOALABEAR_16)
}

func Test_ZkcUnit_Type_10(t *testing.T) {
	checkZkcUnit(t, "zkc/unit/type_10", field.BLS12_377, field.KOALABEAR_16)
}

// ===================================================================
// Control Flow Tests
// ===================================================================

func Test_ZkcUnit_Cfg_01(t *testing.T) {
	checkZkcUnit(t, "zkc/unit/cfg_01", field.BLS12_377, field.KOALABEAR_16)
}

func Test_ZkcUnit_Cfg_02(t *testing.T) {
	checkZkcUnit(t, "zkc/unit/cfg_02", field.BLS12_377, field.KOALABEAR_16)
}

// ===================================================================
// Loop Tests
// ===================================================================

func Test_ZkcUnit_While_01(t *testing.T) {
	checkZkcUnit(t, "zkc/unit/while_01", field.BLS12_377, field.KOALABEAR_16)
}

func Test_ZkcUnit_While_02(t *testing.T) {
	checkZkcUnit(t, "zkc/unit/while_02", field.BLS12_377, field.KOALABEAR_16)
}

func Test_ZkcUnit_While_03(t *testing.T) {
	checkZkcUnit(t, "zkc/unit/while_03", field.BLS12_377, field.KOALABEAR_16)
}

func Test_ZkcUnit_For_01(t *testing.T) {
	checkZkcUnit(t, "zkc/unit/for_01", field.BLS12_377, field.KOALABEAR_16)
}

func Test_ZkcUnit_For_02(t *testing.T) {
	checkZkcUnit(t, "zkc/unit/for_02", field.BLS12_377, field.KOALABEAR_16)
}

func Test_ZkcUnit_For_03(t *testing.T) {
	checkZkcUnit(t, "zkc/unit/for_03", field.BLS12_377, field.KOALABEAR_16)
}

func Test_ZkcUnit_For_04(t *testing.T) {
	checkZkcUnit(t, "zkc/unit/for_04", field.BLS12_377, field.KOALABEAR_16)
}

// ===================================================================
// Break Tests
// ===================================================================

func Test_ZkcUnit_Break_01(t *testing.T) {
	checkZkcUnit(t, "zkc/unit/break_01", field.BLS12_377, field.KOALABEAR_16)
}

// ===================================================================
// Continue Tests
// ===================================================================

func Test_ZkcUnit_Continue_01(t *testing.T) {
	checkZkcUnit(t, "zkc/unit/continue_01", field.BLS12_377, field.KOALABEAR_16)
}

// ===================================================================
// Bitwise Tests
// ===================================================================

func Test_ZkcUnit_Bitwise_01(t *testing.T) {
	checkZkcUnitWithLowerZkcNative(t, "zkc/unit/bitwise_01", field.BLS12_377, field.KOALABEAR_16)
}

func Test_ZkcUnit_Bitwise_02(t *testing.T) {
	checkZkcUnitWithLowerZkcNative(t, "zkc/unit/bitwise_02", field.BLS12_377, field.KOALABEAR_16)
}

func Test_ZkcUnit_Bitwise_03(t *testing.T) {
	checkZkcUnitWithLowerZkcNative(t, "zkc/unit/bitwise_03", field.BLS12_377, field.KOALABEAR_16)
}

func Test_ZkcUnit_Bitwise_04(t *testing.T) {
	checkZkcUnitWithLowerZkcNative(t, "zkc/unit/bitwise_04", field.BLS12_377, field.KOALABEAR_16)
}

func Test_ZkcUnit_Bitwise_05(t *testing.T) {
	checkZkcUnitWithLowerZkcNative(t, "zkc/unit/bitwise_05", field.BLS12_377, field.KOALABEAR_16)
}

func Test_ZkcUnit_Bitwise_06(t *testing.T) {
	checkZkcUnitWithLowerZkcNative(t, "zkc/unit/bitwise_06", field.BLS12_377, field.KOALABEAR_16)
}

func Test_ZkcUnit_Bitwise_07(t *testing.T) {
	checkZkcUnitWithLowerZkcNative(t, "zkc/unit/bitwise_07", field.BLS12_377, field.KOALABEAR_16)
}

func Test_ZkcUnit_Bitwise_08(t *testing.T) {
	checkZkcUnitWithLowerZkcNative(t, "zkc/unit/bitwise_08", field.BLS12_377, field.KOALABEAR_16)
}

func Test_ZkcUnit_Bitwise_09(t *testing.T) {
	checkZkcUnitWithLowerZkcNative(t, "zkc/unit/bitwise_09", field.BLS12_377, field.KOALABEAR_16)
}

func Test_ZkcUnit_Bitwise_10(t *testing.T) {
	checkZkcUnitWithLowerZkcNative(t, "zkc/unit/bitwise_10", field.BLS12_377, field.KOALABEAR_16)
}

func Test_ZkcUnit_Bitwise_11(t *testing.T) {
	checkZkcUnitWithLowerZkcNative(t, "zkc/unit/bitwise_11", field.BLS12_377, field.KOALABEAR_16)
}

func Test_ZkcUnit_Bitwise_12(t *testing.T) {
	checkZkcUnitWithLowerZkcNative(t, "zkc/unit/bitwise_12", field.BLS12_377, field.KOALABEAR_16)
}

// ===================================================================
// Shift Tests
// ===================================================================

func Test_ZkcUnit_Shift_01(t *testing.T) {
	checkZkcUnitWithLowerZkcNative(t, "zkc/unit/shift_01", field.BLS12_377, field.KOALABEAR_16)
}

func Test_ZkcUnit_Shift_02(t *testing.T) {
	checkZkcUnitWithLowerZkcNative(t, "zkc/unit/shift_02", field.BLS12_377, field.KOALABEAR_16)
}

func Test_ZkcUnit_Shift_03(t *testing.T) {
	checkZkcUnitWithLowerZkcNative(t, "zkc/unit/shift_03", field.BLS12_377, field.KOALABEAR_16)
}

func Test_ZkcUnit_Shift_04(t *testing.T) {
	checkZkcUnitWithLowerZkcNative(t, "zkc/unit/shift_04", field.BLS12_377, field.KOALABEAR_16)
}

func Test_ZkcUnit_Shift_05(t *testing.T) {
	checkZkcUnitWithLowerZkcNative(t, "zkc/unit/shift_05", field.BLS12_377, field.KOALABEAR_16)
}

func Test_ZkcUnit_Shift_06(t *testing.T) {
	checkZkcUnitWithLowerZkcNative(t, "zkc/unit/shift_06", field.BLS12_377, field.KOALABEAR_16)
}

func Test_ZkcUnit_Shift_07(t *testing.T) {
	checkZkcUnitWithLowerZkcNative(t, "zkc/unit/shift_07", field.BLS12_377, field.KOALABEAR_16)
}

func Test_ZkcUnit_Shift_08(t *testing.T) {
	checkZkcUnitWithLowerZkcNative(t, "zkc/unit/shift_08", field.BLS12_377, field.KOALABEAR_16)
}

func Test_ZkcUnit_Shift_09(t *testing.T) {
	checkZkcUnitWithLowerZkcNative(t, "zkc/unit/shift_09", field.BLS12_377, field.KOALABEAR_16)
}

func Test_ZkcUnit_Shift_10(t *testing.T) {
	checkZkcUnitWithLowerZkcNative(t, "zkc/unit/shift_10", field.BLS12_377, field.KOALABEAR_16)
}

func Test_ZkcUnit_Shift_11(t *testing.T) {
	checkZkcUnitWithLowerZkcNative(t, "zkc/unit/shift_11", field.BLS12_377, field.KOALABEAR_16)
}

// ===================================================================
// Static Initialiser Tests
// ===================================================================

func Test_ZkcUnit_Static_01(t *testing.T) {
	checkZkcUnit(t, "zkc/unit/static_01", field.BLS12_377, field.KOALABEAR_16)
}

func Test_ZkcUnit_Static_02(t *testing.T) {
	checkZkcUnit(t, "zkc/unit/static_02", field.BLS12_377, field.KOALABEAR_16)
}

func Test_ZkcUnit_SwitchEndian(t *testing.T) {
	checkZkcUnit(t, "zkc/unit/switch_endian", field.BLS12_377, field.KOALABEAR_16)
}

// ===================================================================
// Cast Tests
// ===================================================================

func Test_ZkcUnit_Cast_01(t *testing.T) {
	checkZkcUnit(t, "zkc/unit/cast_01", field.BLS12_377, field.KOALABEAR_16)
}

func Test_ZkcUnit_Cast_02(t *testing.T) {
	checkZkcUnit(t, "zkc/unit/cast_02", field.BLS12_377, field.KOALABEAR_16)
}

func Test_ZkcUnit_Cast_03(t *testing.T) {
	checkZkcUnit(t, "zkc/unit/cast_03", field.BLS12_377, field.KOALABEAR_16)
}

func Test_ZkcUnit_Cast_04(t *testing.T) {
	checkZkcUnit(t, "zkc/unit/cast_04", field.BLS12_377, field.KOALABEAR_16)
}

func Test_ZkcUnit_Cast_05(t *testing.T) {
	checkZkcUnit(t, "zkc/unit/cast_05", field.BLS12_377, field.KOALABEAR_16)
}

// ===================================================================
// Division Tests
// ===================================================================

func Test_ZkcUnit_Div_01(t *testing.T) {
	checkZkcUnit(t, "zkc/unit/div_01", field.BLS12_377, field.KOALABEAR_16)
}

func Test_ZkcUnit_Div_02(t *testing.T) {
	checkZkcUnit(t, "zkc/unit/div_02", field.BLS12_377, field.KOALABEAR_16)
}

// ===================================================================
// Remainder Tests
// ===================================================================

func Test_ZkcUnit_Rem_01(t *testing.T) {
	checkZkcUnit(t, "zkc/unit/rem_01", field.BLS12_377, field.KOALABEAR_16)
}

func Test_ZkcUnit_Rem_02(t *testing.T) {
	checkZkcUnit(t, "zkc/unit/rem_02", field.BLS12_377, field.KOALABEAR_16)
}

// ===================================================================
// Call Tests
// ===================================================================

func Test_ZkcUnit_Call_01(t *testing.T) {
	checkZkcUnit(t, "zkc/unit/call_01", field.BLS12_377, field.KOALABEAR_16)
}

func Test_ZkcUnit_Call_02(t *testing.T) {
	checkZkcUnit(t, "zkc/unit/call_02", field.BLS12_377, field.KOALABEAR_16)
}

// ===================================================================
// Ternary Tests
// ===================================================================

func Test_ZkcUnit_Ternary_01(t *testing.T) {
	checkZkcUnit(t, "zkc/unit/ternary_01", field.BLS12_377, field.KOALABEAR_16)
}

func Test_ZkcUnit_Ternary_02(t *testing.T) {
	checkZkcUnit(t, "zkc/unit/ternary_02", field.BLS12_377, field.KOALABEAR_16)
}

func Test_ZkcUnit_Ternary_03(t *testing.T) {
	checkZkcUnit(t, "zkc/unit/ternary_03", field.BLS12_377, field.KOALABEAR_16)
}

func Test_ZkcUnit_Ternary_04(t *testing.T) {
	checkZkcUnit(t, "zkc/unit/ternary_04", field.BLS12_377, field.KOALABEAR_16)
}

func Test_ZkcUnit_Ternary_05(t *testing.T) {
	checkZkcUnit(t, "zkc/unit/ternary_05", field.BLS12_377, field.KOALABEAR_16)
}

// ===================================================================
// Switch Tests
// ===================================================================

func Test_ZkcUnit_Switch_01(t *testing.T) {
	checkZkcUnit(t, "zkc/unit/switch_01", field.BLS12_377, field.KOALABEAR_16)
}

func Test_ZkcUnit_Switch_02(t *testing.T) {
	checkZkcUnit(t, "zkc/unit/switch_02", field.BLS12_377, field.KOALABEAR_16)
}

func Test_ZkcUnit_Switch_03(t *testing.T) {
	checkZkcUnit(t, "zkc/unit/switch_03", field.BLS12_377, field.KOALABEAR_16)
}

func Test_ZkcUnit_Switch_04(t *testing.T) {
	checkZkcUnit(t, "zkc/unit/switch_04", field.BLS12_377, field.KOALABEAR_16)
}

func Test_ZkcUnit_Switch_05(t *testing.T) {
	checkZkcUnit(t, "zkc/unit/switch_05", field.BLS12_377, field.KOALABEAR_16)
}

func Test_ZkcUnit_Switch_06(t *testing.T) {
	checkZkcUnit(t, "zkc/unit/switch_06", field.BLS12_377, field.KOALABEAR_16)
}

func Test_ZkcUnit_Switch_07(t *testing.T) {
	checkZkcUnit(t, "zkc/unit/switch_07", field.BLS12_377, field.KOALABEAR_16)
}

func Test_ZkcUnit_Switch_08(t *testing.T) {
	checkZkcUnit(t, "zkc/unit/switch_08", field.BLS12_377, field.KOALABEAR_16)
}

func Test_ZkcUnit_Switch_09(t *testing.T) {
	checkZkcUnit(t, "zkc/unit/switch_09", field.BLS12_377, field.KOALABEAR_16)
}

// ===================================================================
// Printf Tests
// ===================================================================

func Test_ZkcUnit_Printf_01(t *testing.T) {
	checkZkcUnit(t, "zkc/unit/printf_01", field.BLS12_377, field.KOALABEAR_16)
}

func Test_ZkcUnit_Printf_02(t *testing.T) {
	checkZkcUnit(t, "zkc/unit/printf_02", field.BLS12_377, field.KOALABEAR_16)
}

func Test_ZkcUnit_Printf_03(t *testing.T) {
	checkZkcUnit(t, "zkc/unit/printf_03", field.BLS12_377, field.KOALABEAR_16)
}

func Test_ZkcUnit_Printf_04(t *testing.T) {
	checkZkcUnit(t, "zkc/unit/printf_04", field.BLS12_377, field.KOALABEAR_16)
}

// ===================================================================
// Include Tests
// ===================================================================

func Test_ZkcUnit_Include_01(t *testing.T) {
	checkZkcUnit(t, "zkc/unit/include_01", field.BLS12_377, field.KOALABEAR_16)
}

func Test_ZkcUnit_Include_02(t *testing.T) {
	checkZkcUnit(t, "zkc/unit/include_02", field.BLS12_377, field.KOALABEAR_16)
}

// ===================================================================
// Skip If (VM inst) Tests
// ===================================================================

func Test_ZkcUnit_SkipIf_01(t *testing.T) {
	checkZkcUnitWithLowerZkcNative(t, "zkc/unit/skip_if_01", field.BLS12_377, field.KOALABEAR_16)
}

//TODO: re-enable me when the ternary bandwidth allocator is fixed (issue 1758)
// func Test_ZkcUnit_SkipIf_02(t *testing.T) {
// 	checkZkcUnitWithLowerZkcNative(t, "zkc/unit/skip_if_02", field.BLS12_377, field.KOALABEAR_16)
// }
//
// func Test_ZkcUnit_SkipIf_03(t *testing.T) {
// 	checkZkcUnitWithLowerZkcNative(t, "zkc/unit/skip_if_03", field.BLS12_377, field.KOALABEAR_16)
// }
//
// func Test_ZkcUnit_SkipIf_04(t *testing.T) {
// 	checkZkcUnitWithLowerZkcNative(t, "zkc/unit/skip_if_04", field.BLS12_377, field.KOALABEAR_16)
// }
//
// func Test_ZkcUnit_SkipIf_05(t *testing.T) {
// 	checkZkcUnitWithLowerZkcNative(t, "zkc/unit/skip_if_05", field.BLS12_377, field.KOALABEAR_16)
// }

// ===================================================================
// Test Helpers
// ===================================================================

func checkZkcUnit(t *testing.T, test string, fields ...field.Config) {
	util.CheckValid(t, test, "zkc", fields...)
}

func checkZkcUnitWithLowerZkcNative(t *testing.T, test string, fields ...field.Config) {
	t.Parallel()
	util.CheckValidWithConfig(t, test, "zkc", codegen.DEFAULT_CONFIG.LowerZkcNative(false), fields...)
	util.CheckValidWithConfig(t, test, "zkc", codegen.DEFAULT_CONFIG.LowerZkcNative(true), fields...)
}
