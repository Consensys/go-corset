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

func Test_ZkcUnit_Basic_01(t *testing.T) {
	checkZkcUnit(t, "zkc/unit/basic_01")
}

func Test_ZkcUnit_Basic_02(t *testing.T) {
	checkZkcUnit(t, "zkc/unit/basic_02")
}

func Test_ZkcUnit_Basic_03(t *testing.T) {
	checkZkcUnit(t, "zkc/unit/basic_03")
}
func Test_ZkcUnit_Basic_04(t *testing.T) {
	checkZkcUnit(t, "zkc/unit/basic_04")
}

func Test_ZkcUnit_Basic_05(t *testing.T) {
	checkZkcUnit(t, "zkc/unit/basic_05")
}

func Test_ZkcUnit_Basic_06(t *testing.T) {
	checkZkcUnit(t, "zkc/unit/basic_06")
}
func Test_ZkcUnit_Basic_07(t *testing.T) {
	checkZkcUnit(t, "zkc/unit/basic_07")
}

func Test_ZkcUnit_Basic_08(t *testing.T) {
	checkZkcUnit(t, "zkc/unit/basic_08")
}

func Test_ZkcUnit_Basic_09(t *testing.T) {
	checkZkcUnit(t, "zkc/unit/basic_09")
}

// ===================================================================
// Constant Tests
// ===================================================================

func Test_ZkcUnit_Constant_01(t *testing.T) {
	checkZkcUnit(t, "zkc/unit/const_01")
}

func Test_ZkcUnit_Constant_02(t *testing.T) {
	checkZkcUnit(t, "zkc/unit/const_02")
}


// ===================================================================
// Type Tests
// ===================================================================

func Test_ZkcUnit_Type_01(t *testing.T) {
	checkZkcUnit(t, "zkc/unit/type_valid_01")
}

func Test_ZkcUnit_Type_02(t *testing.T) {
	checkZkcUnit(t, "zkc/unit/type_valid_02")
}

func Test_ZkcUnit_Type_03(t *testing.T) {
	checkZkcUnit(t, "zkc/unit/type_valid_03")
}

// ===================================================================
// Loop Tests
// ===================================================================

func Test_ZkcUnit_While_01(t *testing.T) {
	checkZkcUnit(t, "zkc/unit/while_01")
}

func Test_ZkcUnit_While_02(t *testing.T) {
	checkZkcUnit(t, "zkc/unit/while_02")
}

func Test_ZkcUnit_While_03(t *testing.T) {
	checkZkcUnit(t, "zkc/unit/while_03")
}

func Test_ZkcUnit_For_01(t *testing.T) {
	checkZkcUnit(t, "zkc/unit/for_01")
}

func Test_ZkcUnit_For_02(t *testing.T) {
	checkZkcUnit(t, "zkc/unit/for_02")
}

func Test_ZkcUnit_For_03(t *testing.T) {
	checkZkcUnit(t, "zkc/unit/for_03")
}

// ===================================================================
// Break Tests
// ===================================================================

func Test_ZkcUnit_Break_01(t *testing.T) {
	checkZkcUnit(t, "zkc/unit/break_01")
}

// ===================================================================
// Continue Tests
// ===================================================================

func Test_ZkcUnit_Continue_01(t *testing.T) {
	checkZkcUnit(t, "zkc/unit/continue_01")
}

// ===================================================================
// Bitwise Tests
// ===================================================================

func Test_ZkcUnit_Bitwise_01(t *testing.T) {
	checkZkcUnit(t, "zkc/unit/bitwise_01")
}

func Test_ZkcUnit_Bitwise_02(t *testing.T) {
	checkZkcUnit(t, "zkc/unit/bitwise_02")
}

func Test_ZkcUnit_Bitwise_03(t *testing.T) {
	checkZkcUnit(t, "zkc/unit/bitwise_03")
}

func Test_ZkcUnit_Bitwise_04(t *testing.T) {
	checkZkcUnit(t, "zkc/unit/bitwise_04")
}

func Test_ZkcUnit_Bitwise_05(t *testing.T) {
	checkZkcUnit(t, "zkc/unit/bitwise_05")
}

func Test_ZkcUnit_Bitwise_06(t *testing.T) {
	checkZkcUnit(t, "zkc/unit/bitwise_06")
}

func Test_ZkcUnit_Bitwise_07(t *testing.T) {
	checkZkcUnit(t, "zkc/unit/bitwise_07")
}

func Test_ZkcUnit_Bitwise_08(t *testing.T) {
	checkZkcUnit(t, "zkc/unit/bitwise_08")
}

// ===================================================================
// Shift Tests
// ===================================================================

func Test_ZkcUnit_Shift_01(t *testing.T) {
	checkZkcUnit(t, "zkc/unit/shift_01")
}

func Test_ZkcUnit_Shift_02(t *testing.T) {
	checkZkcUnit(t, "zkc/unit/shift_02")
}

func Test_ZkcUnit_Shift_03(t *testing.T) {
	checkZkcUnit(t, "zkc/unit/shift_03")
}

func Test_ZkcUnit_Shift_04(t *testing.T) {
	checkZkcUnit(t, "zkc/unit/shift_04")
}

func Test_ZkcUnit_Shift_05(t *testing.T) {
	checkZkcUnit(t, "zkc/unit/shift_05")
}

func Test_ZkcUnit_Shift_06(t *testing.T) {
	checkZkcUnit(t, "zkc/unit/shift_06")
}

func Test_ZkcUnit_Shift_07(t *testing.T) {
	checkZkcUnit(t, "zkc/unit/shift_07")
}

// ===================================================================
// Static Initialiser Tests
// ===================================================================

func Test_ZkcUnit_Static_01(t *testing.T) {
	checkZkcUnit(t, "zkc/unit/static_01")
}

// ===================================================================
// Cast Tests
// ===================================================================

func Test_ZkcUnit_Cast_01(t *testing.T) {
	checkZkcUnit(t, "zkc/unit/cast_01")
}

func Test_ZkcUnit_Cast_02(t *testing.T) {
	checkZkcUnit(t, "zkc/unit/cast_02")
}

func Test_ZkcUnit_Cast_03(t *testing.T) {
	checkZkcUnit(t, "zkc/unit/cast_03")
}

func Test_ZkcUnit_Cast_04(t *testing.T) {
	checkZkcUnit(t, "zkc/unit/cast_04")
}

func Test_ZkcUnit_Cast_05(t *testing.T) {
	checkZkcUnit(t, "zkc/unit/cast_05")
}

// ===================================================================
// Division Tests
// ===================================================================

func Test_ZkcUnit_Div_01(t *testing.T) {
	checkZkcUnit(t, "zkc/unit/div_01")
}

func Test_ZkcUnit_Div_02(t *testing.T) {
	checkZkcUnit(t, "zkc/unit/div_02")
}

// ===================================================================
// Remainder Tests
// ===================================================================

func Test_ZkcUnit_Rem_01(t *testing.T) {
	checkZkcUnit(t, "zkc/unit/rem_01")
}

func Test_ZkcUnit_Rem_02(t *testing.T) {
	checkZkcUnit(t, "zkc/unit/rem_02")
}

// ===================================================================
// Call Tests
// ===================================================================

func Test_ZkcUnit_Call_01(t *testing.T) {
	checkZkcUnit(t, "zkc/unit/call_01")
}

func Test_ZkcUnit_Call_02(t *testing.T) {
	checkZkcUnit(t, "zkc/unit/call_02")
}

// ===================================================================
// Test Helpers
// ===================================================================

func checkZkcUnit(t *testing.T, test string) {
	util.CheckValid(t, test, "zkc", util.CompileZkc)
}
