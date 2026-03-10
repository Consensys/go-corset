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
	checkZkcUnit(t, "zkc/unit/basic_valid_01")
}

func Test_ZkcUnit_Basic_02(t *testing.T) {
	checkZkcUnit(t, "zkc/unit/basic_valid_02")
}

func Test_ZkcUnit_Basic_03(t *testing.T) {
	checkZkcUnit(t, "zkc/unit/basic_valid_03")
}
func Test_ZkcUnit_Basic_04(t *testing.T) {
	checkZkcUnit(t, "zkc/unit/basic_valid_04")
}

// ===================================================================
// Constant Tests
// ===================================================================

func Test_ZkcUnit_Constant_01(t *testing.T) {
	checkZkcUnit(t, "zkc/unit/const_valid_01")
}

func Test_ZkcUnit_Constant_02(t *testing.T) {
	checkZkcUnit(t, "zkc/unit/const_valid_02")
}

// ===================================================================
// Loop Tests
// ===================================================================

func Test_ZkcUnit_While_01(t *testing.T) {
	checkZkcUnit(t, "zkc/unit/while_valid_01")
}

func Test_ZkcUnit_While_02(t *testing.T) {
	checkZkcUnit(t, "zkc/unit/while_valid_02")
}

func Test_ZkcUnit_For_01(t *testing.T) {
	checkZkcUnit(t, "zkc/unit/for_valid_01")
}

func Test_ZkcUnit_For_02(t *testing.T) {
	checkZkcUnit(t, "zkc/unit/for_valid_02")
}

// ===================================================================
// Bitwise Tests
// ===================================================================

func Test_ZkcUnit_Bitwise_01(t *testing.T) {
	checkZkcUnit(t, "zkc/unit/bitwise_valid_01")
}

func Test_ZkcUnit_Bitwise_02(t *testing.T) {
	checkZkcUnit(t, "zkc/unit/bitwise_valid_02")
}

func Test_ZkcUnit_Bitwise_03(t *testing.T) {
	checkZkcUnit(t, "zkc/unit/bitwise_valid_03")
}

func Test_ZkcUnit_Bitwise_04(t *testing.T) {
	checkZkcUnit(t, "zkc/unit/bitwise_valid_04")
}

func Test_ZkcUnit_Bitwise_05(t *testing.T) {
	checkZkcUnit(t, "zkc/unit/bitwise_valid_05")
}

func Test_ZkcUnit_Bitwise_06(t *testing.T) {
	checkZkcUnit(t, "zkc/unit/bitwise_valid_06")
}

func Test_ZkcUnit_Bitwise_07(t *testing.T) {
	checkZkcUnit(t, "zkc/unit/bitwise_valid_07")
}

func Test_ZkcUnit_Bitwise_08(t *testing.T) {
	checkZkcUnit(t, "zkc/unit/bitwise_valid_08")
}

// ===================================================================
// Shift Tests
// ===================================================================

func Test_ZkcUnit_Shift_01(t *testing.T) {
	checkZkcUnit(t, "zkc/unit/shift_valid_01")
}

func Test_ZkcUnit_Shift_02(t *testing.T) {
	checkZkcUnit(t, "zkc/unit/shift_valid_02")
}

func Test_ZkcUnit_Shift_03(t *testing.T) {
	checkZkcUnit(t, "zkc/unit/shift_valid_03")
}

func Test_ZkcUnit_Shift_04(t *testing.T) {
	checkZkcUnit(t, "zkc/unit/shift_valid_04")
}

// ===================================================================
// Test Helpers
// ===================================================================

func checkZkcUnit(t *testing.T, test string) {
	util.CheckValid(t, test, "zkc", util.CompileZkc)
}
