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
)

// ===================================================================
// Basic Tests
// ===================================================================

func Test_ZkcInvalid_Basic_01(t *testing.T) {
	checkZkcInvalid(t, "zkc/invalid/basic_01")
}

func Test_ZkcInvalid_Basic_02(t *testing.T) {
	checkZkcInvalid(t, "zkc/invalid/basic_02")
}

func Test_ZkcInvalid_Basic_03(t *testing.T) {
	checkZkcInvalid(t, "zkc/invalid/basic_03")
}

func Test_ZkcInvalid_Basic_04(t *testing.T) {
	checkZkcInvalid(t, "zkc/invalid/basic_04")
}

func Test_ZkcInvalid_Basic_05(t *testing.T) {
	checkZkcInvalid(t, "zkc/invalid/basic_05")
}

func Test_ZkcInvalid_Basic_06(t *testing.T) {
	checkZkcInvalid(t, "zkc/invalid/basic_06")
}

func Test_ZkcInvalid_Basic_07(t *testing.T) {
	checkZkcInvalid(t, "zkc/invalid/basic_07")
}

func Test_ZkcInvalid_Basic_08(t *testing.T) {
	checkZkcInvalid(t, "zkc/invalid/basic_08")
}

func Test_ZkcInvalid_Basic_09(t *testing.T) {
	checkZkcInvalid(t, "zkc/invalid/basic_09")
}

func Test_ZkcInvalid_Basic_10(t *testing.T) {
	checkZkcInvalid(t, "zkc/invalid/basic_10")
}

func Test_ZkcInvalid_Basic_11(t *testing.T) {
	checkZkcInvalid(t, "zkc/invalid/basic_11")
}

// ===================================================================
// If Tests
// ===================================================================

func Test_ZkcInvalid_If_01(t *testing.T) {
	checkZkcInvalid(t, "zkc/invalid/if_01")
}

func Test_ZkcInvalid_If_02(t *testing.T) {
	checkZkcInvalid(t, "zkc/invalid/if_02")
}

func Test_ZkcInvalid_If_03(t *testing.T) {
	checkZkcInvalid(t, "zkc/invalid/if_03")
}

func Test_ZkcInvalid_If_04(t *testing.T) {
	checkZkcInvalid(t, "zkc/invalid/if_04")
}

// ===================================================================
// Constant Tests
// ===================================================================

func Test_ZkcInvalid_Constant_01(t *testing.T) {
	checkZkcInvalid(t, "zkc/invalid/constant_01")
}

// should fail
// func Test_ZkcInvalid_Constant_02(t *testing.T) {
// 	checkZkcInvalid(t, "zkc/invalid/constant_02")
// }

func Test_ZkcInvalid_Constant_03(t *testing.T) {
	checkZkcInvalid(t, "zkc/invalid/constant_03")
}

// should fail
// func Test_ZkcInvalid_Constant_04(t *testing.T) {
// 	checkZkcInvalid(t, "zkc/invalid/constant_04")
// }

func Test_ZkcInvalid_Constant_05(t *testing.T) {
	checkZkcInvalid(t, "zkc/invalid/constant_05")
}

// ===================================================================
// While Tests
// ===================================================================

func Test_ZkcInvalid_While_01(t *testing.T) {
	checkZkcInvalid(t, "zkc/invalid/while_01")
}

func Test_ZkcInvalid_While_02(t *testing.T) {
	checkZkcInvalid(t, "zkc/invalid/while_02")
}

func Test_ZkcInvalid_While_03(t *testing.T) {
	checkZkcInvalid(t, "zkc/invalid/while_03")
}

// ===================================================================
// For Tests
// ===================================================================

func Test_ZkcInvalid_For_01(t *testing.T) {
	checkZkcInvalid(t, "zkc/invalid/for_01")
}

func Test_ZkcInvalid_For_02(t *testing.T) {
	checkZkcInvalid(t, "zkc/invalid/for_02")
}

func Test_ZkcInvalid_For_03(t *testing.T) {
	checkZkcInvalid(t, "zkc/invalid/for_03")
}

// ===================================================================
// Break Tests
// ===================================================================

func Test_ZkcInvalid_Break_01(t *testing.T) {
	checkZkcInvalid(t, "zkc/invalid/break_01")
}

func Test_ZkcInvalid_Break_02(t *testing.T) {
	checkZkcInvalid(t, "zkc/invalid/break_02")
}

func Test_ZkcInvalid_Break_03(t *testing.T) {
	checkZkcInvalid(t, "zkc/invalid/break_03")
}

func Test_ZkcInvalid_Break_04(t *testing.T) {
	checkZkcInvalid(t, "zkc/invalid/break_04")
}

// ===================================================================
// Continue Tests
// ===================================================================

func Test_ZkcInvalid_Continue_01(t *testing.T) {
	checkZkcInvalid(t, "zkc/invalid/continue_01")
}

func Test_ZkcInvalid_Continue_02(t *testing.T) {
	checkZkcInvalid(t, "zkc/invalid/continue_02")
}

// ===================================================================
// Bitwise Tests
// ===================================================================

func Test_ZkcInvalid_Bitwise_01(t *testing.T) {
	checkZkcInvalid(t, "zkc/invalid/bitwise_01")
}

func Test_ZkcInvalid_Bitwise_02(t *testing.T) {
	checkZkcInvalid(t, "zkc/invalid/bitwise_02")
}

func Test_ZkcInvalid_Bitwise_03(t *testing.T) {
	checkZkcInvalid(t, "zkc/invalid/bitwise_03")
}

func Test_ZkcInvalid_Bitwise_04(t *testing.T) {
	checkZkcInvalid(t, "zkc/invalid/bitwise_04")
}

func Test_ZkcInvalid_Bitwise_05(t *testing.T) {
	checkZkcInvalid(t, "zkc/invalid/bitwise_05")
}

func Test_ZkcInvalid_Bitwise_06(t *testing.T) {
	checkZkcInvalid(t, "zkc/invalid/bitwise_06")
}

func Test_ZkcInvalid_Bitwise_07(t *testing.T) {
	checkZkcInvalid(t, "zkc/invalid/bitwise_07")
}

func Test_ZkcInvalid_Bitwise_08(t *testing.T) {
	checkZkcInvalid(t, "zkc/invalid/bitwise_08")
}

func Test_ZkcInvalid_Bitwise_09(t *testing.T) {
	checkZkcInvalid(t, "zkc/invalid/bitwise_09")
}

func Test_ZkcInvalid_Bitwise_10(t *testing.T) {
	checkZkcInvalid(t, "zkc/invalid/bitwise_10")
}

func Test_ZkcInvalid_Bitwise_11(t *testing.T) {
	checkZkcInvalid(t, "zkc/invalid/bitwise_11")
}

// ===================================================================
// Shift Tests
// ===================================================================

func Test_ZkcInvalid_Shift_01(t *testing.T) {
	checkZkcInvalid(t, "zkc/invalid/shift_01")
}

func Test_ZkcInvalid_Shift_02(t *testing.T) {
	checkZkcInvalid(t, "zkc/invalid/shift_02")
}

func Test_ZkcInvalid_Shift_03(t *testing.T) {
	checkZkcInvalid(t, "zkc/invalid/shift_03")
}

func Test_ZkcInvalid_Shift_04(t *testing.T) {
	checkZkcInvalid(t, "zkc/invalid/shift_04")
}

func Test_ZkcInvalid_Shift_05(t *testing.T) {
	checkZkcInvalid(t, "zkc/invalid/shift_05")
}

func Test_ZkcInvalid_Shift_06(t *testing.T) {
	checkZkcInvalid(t, "zkc/invalid/shift_06")
}

func Test_ZkcInvalid_Shift_07(t *testing.T) {
	checkZkcInvalid(t, "zkc/invalid/shift_07")
}

func Test_ZkcInvalid_Shift_08(t *testing.T) {
	checkZkcInvalid(t, "zkc/invalid/shift_08")
}

// ===================================================================
// Memory Tests
// ===================================================================

func Test_ZkcInvalid_Memory_01(t *testing.T) {
	checkZkcInvalid(t, "zkc/invalid/memory_01")
}

func Test_ZkcInvalid_Memory_02(t *testing.T) {
	checkZkcInvalid(t, "zkc/invalid/memory_02")
}

func Test_ZkcInvalid_Memory_03(t *testing.T) {
	checkZkcInvalid(t, "zkc/invalid/memory_03")
}

func Test_ZkcInvalid_Memory_04(t *testing.T) {
	checkZkcInvalid(t, "zkc/invalid/memory_04")
}

func Test_ZkcInvalid_Memory_05(t *testing.T) {
	checkZkcInvalid(t, "zkc/invalid/memory_05")
}

func Test_ZkcInvalid_Memory_06(t *testing.T) {
	checkZkcInvalid(t, "zkc/invalid/memory_06")
}

func Test_ZkcInvalid_Memory_07(t *testing.T) {
	checkZkcInvalid(t, "zkc/invalid/memory_07")
}

func Test_ZkcInvalid_Memory_08(t *testing.T) {
	checkZkcInvalid(t, "zkc/invalid/memory_08")
}

func Test_ZkcInvalid_Memory_09(t *testing.T) {
	checkZkcInvalid(t, "zkc/invalid/memory_09")
}

func Test_ZkcInvalid_Memory_10(t *testing.T) {
	checkZkcInvalid(t, "zkc/invalid/memory_10")
}

func Test_ZkcInvalid_Memory_11(t *testing.T) {
	checkZkcInvalid(t, "zkc/invalid/memory_11")
}

func Test_ZkcInvalid_Memory_12(t *testing.T) {
	checkZkcInvalid(t, "zkc/invalid/memory_12")
}

// ===================================================================
// Static Tests
// ===================================================================

func Test_ZkcInvalid_Static_01(t *testing.T) {
	checkZkcInvalid(t, "zkc/invalid/static_01")
}

func Test_ZkcInvalid_Static_02(t *testing.T) {
	checkZkcInvalid(t, "zkc/invalid/static_02")
}

func Test_ZkcInvalid_Static_03(t *testing.T) {
	checkZkcInvalid(t, "zkc/invalid/static_03")
}

func Test_ZkcInvalid_Static_04(t *testing.T) {
	checkZkcInvalid(t, "zkc/invalid/static_04")
}

// ===================================================================
// Call Tests
// ===================================================================

func Test_ZkcInvalid_Call_01(t *testing.T) {
	checkZkcInvalid(t, "zkc/invalid/call_01")
}

func Test_ZkcInvalid_Call_02(t *testing.T) {
	checkZkcInvalid(t, "zkc/invalid/call_02")
}

func Test_ZkcInvalid_Call_03(t *testing.T) {
	checkZkcInvalid(t, "zkc/invalid/call_03")
}
func Test_ZkcInvalid_Call_04(t *testing.T) {
	checkZkcInvalid(t, "zkc/invalid/call_04")
}

func Test_ZkcInvalid_Call_05(t *testing.T) {
	checkZkcInvalid(t, "zkc/invalid/call_05")
}

func Test_ZkcInvalid_Call_06(t *testing.T) {
	checkZkcInvalid(t, "zkc/invalid/call_06")
}

// ===================================================================
// Division Tests
// ===================================================================

func Test_ZkcInvalid_Div_01(t *testing.T) {
	checkZkcInvalid(t, "zkc/invalid/div_01")
}

func Test_ZkcInvalid_Div_02(t *testing.T) {
	checkZkcInvalid(t, "zkc/invalid/div_02")
}

func Test_ZkcInvalid_Div_03(t *testing.T) {
	checkZkcInvalid(t, "zkc/invalid/div_03")
}

func Test_ZkcInvalid_Div_04(t *testing.T) {
	checkZkcInvalid(t, "zkc/invalid/div_04")
}

func Test_ZkcInvalid_Div_05(t *testing.T) {
	checkZkcInvalid(t, "zkc/invalid/div_05")
}

// ===================================================================
// Remainder Tests
// ===================================================================

func Test_ZkcInvalid_Rem_01(t *testing.T) {
	checkZkcInvalid(t, "zkc/invalid/rem_01")
}

func Test_ZkcInvalid_Rem_02(t *testing.T) {
	checkZkcInvalid(t, "zkc/invalid/rem_02")
}

func Test_ZkcInvalid_Rem_03(t *testing.T) {
	checkZkcInvalid(t, "zkc/invalid/rem_03")
}

func Test_ZkcInvalid_Rem_04(t *testing.T) {
	checkZkcInvalid(t, "zkc/invalid/rem_04")
}

func Test_ZkcInvalid_Rem_05(t *testing.T) {
	checkZkcInvalid(t, "zkc/invalid/rem_05")
}

// ===================================================================
// Cast Tests
// ===================================================================

func Test_ZkcInvalid_Cast_01(t *testing.T) {
	checkZkcInvalid(t, "zkc/invalid/cast_01")
}

func Test_ZkcInvalid_Cast_02(t *testing.T) {
	checkZkcInvalid(t, "zkc/invalid/cast_02")
}

// ===================================================================
// Type Tests
// ===================================================================

// runtime: goroutine stack exceeds 1000000000-byte limit
// Address with cycle_detection.go
/*func Test_ZkcInvalid_Type_01(t *testing.T) {
	checkZkcInvalid(t, "zkc/invalid/type_01")
}
*/
func Test_ZkcInvalid_Type_02(t *testing.T) {
	checkZkcInvalid(t, "zkc/invalid/type_02")
}

func Test_ZkcInvalid_Type_03(t *testing.T) {
	checkZkcInvalid(t, "zkc/invalid/type_03")
}

func Test_ZkcInvalid_Type_04(t *testing.T) {
	checkZkcInvalid(t, "zkc/invalid/type_04")
}

func Test_ZkcInvalid_Type_05(t *testing.T) {
	checkZkcInvalid(t, "zkc/invalid/type_05")
}

func Test_ZkcInvalid_Type_03(t *testing.T) {
	checkZkcInvalid(t, "zkc/invalid/type_invalid_03")
}


func Test_ZkcInvalid_Type_04(t *testing.T) {
	checkZkcInvalid(t, "zkc/invalid/type_invalid_04")
}

// ===================================================================
// Test Helpers
// ===================================================================

func checkZkcInvalid(t *testing.T, test string) {
	util.CheckInvalid(t, test, "zkc", "//error", util.CompileZkc)
}
