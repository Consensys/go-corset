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
	"github.com/consensys/go-corset/pkg/util/source"
	"github.com/consensys/go-corset/pkg/zkc/compiler"
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
// Test Helpers
// ===================================================================

func checkZkcInvalid(t *testing.T, test string) {
	util.CheckInvalid(t, test, "zkc", compileZkc)
}

func compileZkc(srcfile source.File) []source.SyntaxError {
	_, _, errors := compiler.Compile(srcfile)
	//
	return errors
}
