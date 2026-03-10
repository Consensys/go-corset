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
	checkZkcInvalid(t, "zkc/invalid/basic_invalid_01")
}

func Test_ZkcInvalid_Basic_02(t *testing.T) {
	checkZkcInvalid(t, "zkc/invalid/basic_invalid_02")
}

func Test_ZkcInvalid_Basic_03(t *testing.T) {
	checkZkcInvalid(t, "zkc/invalid/basic_invalid_03")
}

func Test_ZkcInvalid_Basic_04(t *testing.T) {
	checkZkcInvalid(t, "zkc/invalid/basic_invalid_04")
}

func Test_ZkcInvalid_Basic_05(t *testing.T) {
	checkZkcInvalid(t, "zkc/invalid/basic_invalid_05")
}

func Test_ZkcInvalid_Basic_06(t *testing.T) {
	checkZkcInvalid(t, "zkc/invalid/basic_invalid_06")
}

func Test_ZkcInvalid_Basic_07(t *testing.T) {
	checkZkcInvalid(t, "zkc/invalid/basic_invalid_07")
}

func Test_ZkcInvalid_Basic_08(t *testing.T) {
	checkZkcInvalid(t, "zkc/invalid/basic_invalid_08")
}

func Test_ZkcInvalid_Basic_09(t *testing.T) {
	checkZkcInvalid(t, "zkc/invalid/basic_invalid_09")
}

func Test_ZkcInvalid_Basic_10(t *testing.T) {
	checkZkcInvalid(t, "zkc/invalid/basic_invalid_10")
}

// ===================================================================
// If Tests
// ===================================================================

func Test_ZkcInvalid_If_01(t *testing.T) {
	checkZkcInvalid(t, "zkc/invalid/if_invalid_01")
}

func Test_ZkcInvalid_If_02(t *testing.T) {
	checkZkcInvalid(t, "zkc/invalid/if_invalid_02")
}

func Test_ZkcInvalid_If_03(t *testing.T) {
	checkZkcInvalid(t, "zkc/invalid/if_invalid_03")
}

func Test_ZkcInvalid_If_04(t *testing.T) {
	checkZkcInvalid(t, "zkc/invalid/if_invalid_04")
}

// ===================================================================
// Constant Tests
// ===================================================================

func Test_ZkcInvalid_Constant_01(t *testing.T) {
	checkZkcInvalid(t, "zkc/invalid/constant_invalid_01")
}

// should fail
// func Test_ZkcInvalid_Constant_02(t *testing.T) {
// 	checkZkcInvalid(t, "zkc/invalid/constant_invalid_02")
// }

func Test_ZkcInvalid_Constant_03(t *testing.T) {
	checkZkcInvalid(t, "zkc/invalid/constant_invalid_03")
}

// should fail
// func Test_ZkcInvalid_Constant_04(t *testing.T) {
// 	checkZkcInvalid(t, "zkc/invalid/constant_invalid_04")
// }

func Test_ZkcInvalid_Constant_05(t *testing.T) {
	checkZkcInvalid(t, "zkc/invalid/constant_invalid_05")
}

// ===================================================================
// While Tests
// ===================================================================

func Test_ZkcInvalid_While_01(t *testing.T) {
	checkZkcInvalid(t, "zkc/invalid/while_invalid_01")
}

func Test_ZkcInvalid_While_02(t *testing.T) {
	checkZkcInvalid(t, "zkc/invalid/while_invalid_02")
}

func Test_ZkcInvalid_While_03(t *testing.T) {
	checkZkcInvalid(t, "zkc/invalid/while_invalid_03")
}

// ===================================================================
// For Tests
// ===================================================================

func Test_ZkcInvalid_For_01(t *testing.T) {
	checkZkcInvalid(t, "zkc/invalid/for_invalid_01")
}

func Test_ZkcInvalid_For_02(t *testing.T) {
	checkZkcInvalid(t, "zkc/invalid/for_invalid_02")
}

func Test_ZkcInvalid_For_03(t *testing.T) {
	checkZkcInvalid(t, "zkc/invalid/for_invalid_03")
}

// ===================================================================
// Bitwise Tests
// ===================================================================

func Test_ZkcInvalid_Bitwise_01(t *testing.T) {
	checkZkcInvalid(t, "zkc/invalid/bitwise_invalid_01")
}

func Test_ZkcInvalid_Bitwise_02(t *testing.T) {
	checkZkcInvalid(t, "zkc/invalid/bitwise_invalid_02")
}

func Test_ZkcInvalid_Bitwise_03(t *testing.T) {
	checkZkcInvalid(t, "zkc/invalid/bitwise_invalid_03")
}

func Test_ZkcInvalid_Bitwise_04(t *testing.T) {
	checkZkcInvalid(t, "zkc/invalid/bitwise_invalid_04")
}

// ===================================================================
// Memory Tests
// ===================================================================

func Test_ZkcInvalid_Memory_01(t *testing.T) {
	checkZkcInvalid(t, "zkc/invalid/memory_invalid_01")
}

func Test_ZkcInvalid_Memory_02(t *testing.T) {
	checkZkcInvalid(t, "zkc/invalid/memory_invalid_02")
}

func Test_ZkcInvalid_Memory_03(t *testing.T) {
	checkZkcInvalid(t, "zkc/invalid/memory_invalid_03")
}

func Test_ZkcInvalid_Memory_04(t *testing.T) {
	checkZkcInvalid(t, "zkc/invalid/memory_invalid_04")
}

func Test_ZkcInvalid_Memory_05(t *testing.T) {
	checkZkcInvalid(t, "zkc/invalid/memory_invalid_05")
}

func Test_ZkcInvalid_Memory_06(t *testing.T) {
	checkZkcInvalid(t, "zkc/invalid/memory_invalid_06")
}

// ===================================================================
// Test Helpers
// ===================================================================

func checkZkcInvalid(t *testing.T, test string) {
	util.CheckInvalid(t, test, "zkc", "//error", util.CompileZkc)
}
