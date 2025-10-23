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

func Test_Agnostic_Padding_01(t *testing.T) {
	util.CheckWithFields(t, false, "corset/agnostic/padding_01", util.CORSET_MAX_PADDING, field.BLS12_377,
		field.KOALABEAR_16)
}

// func Test_Agnostic_Vanish_01(t *testing.T) {
// 	test_util.Check(t, false, "corset/agnostic/vanish_01")
// }

// func Test_Vanish_02(t *testing.T) {
// 	test_util.Check(t, false, "corset/agnostic/vanish_02")
// }

func Test_Agnostic_Lookup_01(t *testing.T) {
	util.CheckWithFields(t, false, "corset/agnostic/lookup_01", util.CORSET_MAX_PADDING, field.BLS12_377,
		field.KOALABEAR_16)
}

// #1247
// func Test_Agnostic_Lookup_02(t *testing.T) {
// 	util.CheckWithFields(t, false, "corset/agnostic/lookup_02", util.CORSET_MAX_PADDING, field.BLS12_377,
//  field.KOALABEAR_16)
// }

// #1247
// func Test_Agnostic_Lookup_03(t *testing.T) {
// 	// NOTE: BLS12_377 generates an irregular lookup (which, at the time of
// 	// writing, are not supported).
// 	util.CheckWithFields(t, false, "corset/agnostic/lookup_03", util.CORSET_MAX_PADDING, field.KOALABEAR_16)
// }

// #1247
// func Test_Agnostic_Lookup_04(t *testing.T) {
// 	util.CheckWithFields(t, false, "corset/agnostic/lookup_04", util.CORSET_MAX_PADDING, field.BLS12_377,
// field.KOALABEAR_16)
// }

// #1247
//
//	func Test_Agnostic_Lookup_05(t *testing.T) {
//		util.CheckWithFields(t, false, "corset/agnostic/lookup_05", util.CORSET_MAX_PADDING, field.BLS12_377,
// field.KOALABEAR_16)
//	}

// #1247
//
//	func Test_Agnostic_Lookup_06(t *testing.T) {
//		util.CheckWithFields(t, false, "corset/agnostic/lookup_06", util.CORSET_MAX_PADDING, field.BLS12_377,
// field.KOALABEAR_16)
//	}

// #1247
//
//	func Test_Agnostic_Lookup_07(t *testing.T) {
//		util.CheckWithFields(t, false, "corset/agnostic/lookup_07", util.CORSET_MAX_PADDING, field.BLS12_377,
// field.KOALABEAR_16)
//	}

// #1247
// func Test_Agnostic_Lookup_08(t *testing.T) {
// 	util.CheckWithFields(t, false, "corset/agnostic/lookup_08", util.CORSET_MAX_PADDING, field.BLS12_377,
// field.KOALABEAR_16)
// }
