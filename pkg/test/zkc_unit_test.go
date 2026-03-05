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
// Test Helpers
// ===================================================================

func checkZkcUnit(t *testing.T, test string) {
	util.CheckValid(t, test, "zkc", util.CompileZkc)
}
