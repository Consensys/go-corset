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

func Check(t *testing.T, stdlib bool, test string) {
	util.CheckWithFields(t, stdlib, test, util.CORSET_MAX_PADDING,
		field.BLS12_377,
		field.KOALABEAR_16,
		//field.GF_8209,
	)
}

// ===================================================================
// Basic Tests
// ===================================================================

func Test_Valid_Basic_01(t *testing.T) {
	Check(t, false, "corset/valid/basic_01")
}

func Test_Valid_Basic_02(t *testing.T) {
	Check(t, false, "corset/valid/basic_02")
}

func Test_Valid_Basic_03(t *testing.T) {
	Check(t, false, "corset/valid/basic_03")
}

func Test_Valid_Basic_04(t *testing.T) {
	Check(t, false, "corset/valid/basic_04")
}

// Ignored because uses a negative constant.
//
// func Test_Valid_Basic_05(t *testing.T) {
// 	util.Check(t, false, "corset/valid/basic_05")
// }

func Test_Valid_Basic_06(t *testing.T) {
	Check(t, false, "corset/valid/basic_06")
}

func Test_Valid_Basic_07(t *testing.T) {
	Check(t, false, "corset/valid/basic_07")
}

func Test_Valid_Basic_08(t *testing.T) {
	Check(t, false, "corset/valid/basic_08")
}

func Test_Valid_Basic_09(t *testing.T) {
	Check(t, false, "corset/valid/basic_09")
}

func Test_Valid_Basic_10(t *testing.T) {
	Check(t, false, "corset/valid/basic_10")
}

func Test_Valid_Basic_11(t *testing.T) {
	Check(t, false, "corset/valid/basic_11")
}
func Test_Valid_Basic_12(t *testing.T) {
	Check(t, false, "corset/valid/basic_12")
}
func Test_Valid_Basic_13(t *testing.T) {
	Check(t, false, "corset/valid/basic_13")
}
func Test_Valid_Basic_14(t *testing.T) {
	Check(t, false, "corset/valid/basic_14")
}
func Test_Valid_Basic_15(t *testing.T) {
	Check(t, false, "corset/valid/basic_15")
}

// ===================================================================
// Constants Tests
// ===================================================================
func Test_Valid_Constant_01(t *testing.T) {
	Check(t, false, "corset/valid/constant_01")
}

func Test_Valid_Constant_02(t *testing.T) {
	Check(t, false, "corset/valid/constant_02")
}

func Test_Valid_Constant_03(t *testing.T) {
	Check(t, false, "corset/valid/constant_03")
}

func Test_Valid_Constant_04(t *testing.T) {
	Check(t, false, "corset/valid/constant_04")
}

func Test_Valid_Constant_05(t *testing.T) {
	Check(t, false, "corset/valid/constant_05")
}

func Test_Valid_Constant_06(t *testing.T) {
	Check(t, false, "corset/valid/constant_06")
}

func Test_Valid_Constant_07(t *testing.T) {
	Check(t, false, "corset/valid/constant_07")
}

func Test_Valid_Constant_08(t *testing.T) {
	Check(t, false, "corset/valid/constant_08")
}

func Test_Valid_Constant_09(t *testing.T) {
	Check(t, false, "corset/valid/constant_09")
}

func Test_Valid_Constant_10(t *testing.T) {
	Check(t, false, "corset/valid/constant_10")
}

func Test_Valid_Constant_11(t *testing.T) {
	Check(t, false, "corset/valid/constant_11")
}

func Test_Valid_Constant_12(t *testing.T) {
	Check(t, false, "corset/valid/constant_12")
}

func Test_Valid_Constant_13(t *testing.T) {
	Check(t, false, "corset/valid/constant_13")
}

func Test_Valid_Constant_14(t *testing.T) {
	Check(t, false, "corset/valid/constant_14")
}

func Test_Valid_Constant_15(t *testing.T) {
	Check(t, false, "corset/valid/constant_15")
}

func Test_Valid_Constant_16(t *testing.T) {
	Check(t, false, "corset/valid/constant_16")
}

// ===================================================================
// Alias Tests
// ===================================================================
func Test_Valid_Alias_01(t *testing.T) {
	Check(t, false, "corset/valid/alias_01")
}
func Test_Valid_Alias_02(t *testing.T) {
	Check(t, false, "corset/valid/alias_02")
}
func Test_Valid_Alias_03(t *testing.T) {
	Check(t, false, "corset/valid/alias_03")
}
func Test_Valid_Alias_04(t *testing.T) {
	Check(t, false, "corset/valid/alias_04")
}
func Test_Valid_Alias_05(t *testing.T) {
	Check(t, false, "corset/valid/alias_05")
}
func Test_Valid_Alias_06(t *testing.T) {
	Check(t, false, "corset/valid/alias_06")
}

// ===================================================================
// Domain Tests
// ===================================================================

func Test_Valid_Domain_01(t *testing.T) {
	Check(t, false, "corset/valid/domain_01")
}

func Test_Valid_Domain_02(t *testing.T) {
	Check(t, false, "corset/valid/domain_02")
}

func Test_Valid_Domain_03(t *testing.T) {
	Check(t, false, "corset/valid/domain_03")
}

// ===================================================================
// Block Tests
// ===================================================================

func Test_Valid_Block_01(t *testing.T) {
	Check(t, false, "corset/valid/block_01")
}

func Test_Valid_Block_02(t *testing.T) {
	Check(t, false, "corset/valid/block_02")
}

func Test_Valid_Block_03(t *testing.T) {
	Check(t, false, "corset/valid/block_03")
}

func Test_Valid_Block_04(t *testing.T) {
	Check(t, false, "corset/valid/block_04")
}

// ===================================================================
// Logical Tests
// ===================================================================

func Test_Valid_Logic_01(t *testing.T) {
	Check(t, false, "corset/valid/logic_01")
}

func Test_Valid_Logic_02(t *testing.T) {
	// Performance
	Check(t, false, "corset/valid/logic_02")
}

// ===================================================================
// Property Tests
// ===================================================================

func Test_Valid_Property_01(t *testing.T) {
	Check(t, false, "corset/valid/property_01")
}

func Test_Valid_Property_02(t *testing.T) {
	Check(t, false, "corset/valid/property_02")
}

func Test_Valid_Property_03(t *testing.T) {
	Check(t, false, "corset/valid/property_03")
}
func Test_Valid_Property_04(t *testing.T) {
	Check(t, false, "corset/valid/property_04")
}

// ===================================================================
// Shift Tests
// ===================================================================

func Test_Valid_Shift_01(t *testing.T) {
	Check(t, false, "corset/valid/shift_01")
}

func Test_Valid_Shift_02(t *testing.T) {
	Check(t, false, "corset/valid/shift_02")
}

func Test_Valid_Shift_03(t *testing.T) {
	Check(t, false, "corset/valid/shift_03")
}

func Test_Valid_Shift_04(t *testing.T) {
	Check(t, false, "corset/valid/shift_04")
}

func Test_Valid_Shift_05(t *testing.T) {
	Check(t, false, "corset/valid/shift_05")
}

func Test_Valid_Shift_06(t *testing.T) {
	Check(t, false, "corset/valid/shift_06")
}

func Test_Valid_Shift_07(t *testing.T) {
	Check(t, false, "corset/valid/shift_07")
}

func Test_Valid_Shift_08(t *testing.T) {
	Check(t, false, "corset/valid/shift_08")
}
func Test_Valid_Shift_09(t *testing.T) {
	Check(t, false, "corset/valid/shift_09")
}

// ===================================================================
// Spillage Tests
// ===================================================================

func Test_Valid_Spillage_01(t *testing.T) {
	Check(t, false, "corset/valid/spillage_01")
}

func Test_Valid_Spillage_02(t *testing.T) {
	Check(t, false, "corset/valid/spillage_02")
}

func Test_Valid_Spillage_03(t *testing.T) {
	Check(t, false, "corset/valid/spillage_03")
}

func Test_Valid_Spillage_04(t *testing.T) {
	Check(t, false, "corset/valid/spillage_04")
}

func Test_Valid_Spillage_05(t *testing.T) {
	Check(t, false, "corset/valid/spillage_05")
}

func Test_Valid_Spillage_06(t *testing.T) {
	Check(t, false, "corset/valid/spillage_06")
}

func Test_Valid_Spillage_07(t *testing.T) {
	Check(t, false, "corset/valid/spillage_07")
}

func Test_Valid_Spillage_08(t *testing.T) {
	Check(t, false, "corset/valid/spillage_08")
}

func Test_Valid_Spillage_09(t *testing.T) {
	Check(t, false, "corset/valid/spillage_09")
}

// ===================================================================
// Normalisation Tests
// ===================================================================

func Test_Valid_Norm_01(t *testing.T) {
	Check(t, false, "corset/valid/norm_01")
}

func Test_Valid_Norm_02(t *testing.T) {
	Check(t, false, "corset/valid/norm_02")
}

func Test_Valid_Norm_03(t *testing.T) {
	Check(t, false, "corset/valid/norm_03")
}

func Test_Valid_Norm_04(t *testing.T) {
	Check(t, false, "corset/valid/norm_04")
}

func Test_Valid_Norm_05(t *testing.T) {
	Check(t, false, "corset/valid/norm_05")
}

func Test_Valid_Norm_06(t *testing.T) {
	Check(t, false, "corset/valid/norm_06")
}

func Test_Valid_Norm_07(t *testing.T) {
	Check(t, false, "corset/valid/norm_07")
}

// ===================================================================
// If-Zero
// ===================================================================

func Test_Valid_If_01(t *testing.T) {
	Check(t, false, "corset/valid/if_01")
}

func Test_Valid_If_02(t *testing.T) {
	Check(t, false, "corset/valid/if_02")
}

func Test_Valid_If_03(t *testing.T) {
	Check(t, false, "corset/valid/if_03")
}

func Test_Valid_If_04(t *testing.T) {
	Check(t, false, "corset/valid/if_04")
}

func Test_Valid_If_05(t *testing.T) {
	Check(t, false, "corset/valid/if_05")
}

func Test_Valid_If_06(t *testing.T) {
	Check(t, false, "corset/valid/if_06")
}

func Test_Valid_If_07(t *testing.T) {
	Check(t, false, "corset/valid/if_07")
}

func Test_Valid_If_08(t *testing.T) {
	Check(t, false, "corset/valid/if_08")
}

func Test_Valid_If_09(t *testing.T) {
	Check(t, false, "corset/valid/if_09")
}

func Test_Valid_If_10(t *testing.T) {
	Check(t, false, "corset/valid/if_10")
}

func Test_Valid_If_11(t *testing.T) {
	Check(t, false, "corset/valid/if_11")
}
func Test_Valid_If_12(t *testing.T) {
	Check(t, false, "corset/valid/if_12")
}
func Test_Valid_If_13(t *testing.T) {
	Check(t, false, "corset/valid/if_13")
}

func Test_Valid_If_14(t *testing.T) {
	Check(t, false, "corset/valid/if_14")
}

func Test_Valid_If_15(t *testing.T) {
	Check(t, false, "corset/valid/if_15")
}

func Test_Valid_If_16(t *testing.T) {
	Check(t, false, "corset/valid/if_16")
}

func Test_Valid_If_17(t *testing.T) {
	Check(t, false, "corset/valid/if_17")
}

func Test_Valid_If_18(t *testing.T) {
	Check(t, false, "corset/valid/if_18")
}

func Test_Valid_If_19(t *testing.T) {
	Check(t, false, "corset/valid/if_19")
}

func Test_Valid_If_20(t *testing.T) {
	Check(t, false, "corset/valid/if_20")
}

func Test_Valid_If_21(t *testing.T) {
	Check(t, false, "corset/valid/if_21")
}

// ===================================================================
// Guards
// ===================================================================

func Test_Valid_Guard_01(t *testing.T) {
	Check(t, false, "corset/valid/guard_01")
}

func Test_Valid_Guard_02(t *testing.T) {
	Check(t, false, "corset/valid/guard_02")
}

func Test_Valid_Guard_03(t *testing.T) {
	Check(t, false, "corset/valid/guard_03")
}

func Test_Valid_Guard_04(t *testing.T) {
	Check(t, false, "corset/valid/guard_04")
}

func Test_Valid_Guard_05(t *testing.T) {
	Check(t, false, "corset/valid/guard_05")
}

// ===================================================================
// Types
// ===================================================================

func Test_Valid_Type_01(t *testing.T) {
	Check(t, false, "corset/valid/type_01")
}

func Test_Valid_Type_02(t *testing.T) {
	util.Check(t, false, "corset/valid/type_02")
}

func Test_Valid_Type_03(t *testing.T) {
	Check(t, false, "corset/valid/type_03")
}

func Test_Valid_Type_04(t *testing.T) {
	Check(t, false, "corset/valid/type_04")
}

func Test_Valid_Type_05(t *testing.T) {
	Check(t, false, "corset/valid/type_05")
}

func Test_Valid_Type_06(t *testing.T) {
	Check(t, false, "corset/valid/type_06")
}

func Test_Valid_Type_07(t *testing.T) {
	Check(t, false, "corset/valid/type_07")
}

func Test_Valid_Type_08(t *testing.T) {
	Check(t, false, "corset/valid/type_08")
}

func Test_Valid_Type_09(t *testing.T) {
	Check(t, false, "corset/valid/type_09")
}

func Test_Valid_Type_10(t *testing.T) {
	Check(t, false, "corset/valid/type_10")
}

func Test_Valid_Type_11(t *testing.T) {
	util.Check(t, false, "corset/valid/type_11")
}

func Test_Valid_Type_12(t *testing.T) {
	util.Check(t, false, "corset/valid/type_12")
}

func Test_Valid_Type_13(t *testing.T) {
	util.Check(t, false, "corset/valid/type_13")
}

// ===================================================================
// Range Constraints
// ===================================================================

func Test_Valid_Range_01(t *testing.T) {
	Check(t, false, "corset/valid/range_01")
}

func Test_Valid_Range_02(t *testing.T) {
	util.Check(t, false, "corset/valid/range_02")
}

func Test_Valid_Range_03(t *testing.T) {
	Check(t, false, "corset/valid/range_03")
}

func Test_Valid_Range_04(t *testing.T) {
	Check(t, false, "corset/valid/range_04")
}

// #1247 --- This is an issue as we have a potentially negative expression used
// in a range constraint.
//
// func Test_Valid_Range_05(t *testing.T) {
//  util.Check(t, false, "corset/valid/range_05")
// }

// ===================================================================
// Constant Propagation
// ===================================================================

func Test_Valid_ConstExpr_01(t *testing.T) {
	Check(t, false, "corset/valid/constexpr_01")
}

func Test_Valid_ConstExpr_02(t *testing.T) {
	Check(t, false, "corset/valid/constexpr_02")
}

func Test_Valid_ConstExpr_03(t *testing.T) {
	Check(t, false, "corset/valid/constexpr_03")
}

func Test_Valid_ConstExpr_04(t *testing.T) {
	Check(t, false, "corset/valid/constexpr_04")
}

func Test_Valid_ConstExpr_05(t *testing.T) {
	Check(t, false, "corset/valid/constexpr_05")
}

// ===================================================================
// Modules
// ===================================================================

func Test_Valid_Module_01(t *testing.T) {
	Check(t, false, "corset/valid/module_01")
}

func Test_Valid_Module_02(t *testing.T) {
	Check(t, false, "corset/valid/module_02")
}

func Test_Valid_Module_03(t *testing.T) {
	Check(t, false, "corset/valid/module_03")
}

func Test_Valid_Module_04(t *testing.T) {
	Check(t, false, "corset/valid/module_04")
}

func Test_Valid_Module_05(t *testing.T) {
	Check(t, false, "corset/valid/module_05")
}

func Test_Valid_Module_06(t *testing.T) {
	Check(t, false, "corset/valid/module_06")
}

func Test_Valid_Module_07(t *testing.T) {
	Check(t, false, "corset/valid/module_07")
}

func Test_Valid_Module_08(t *testing.T) {
	Check(t, false, "corset/valid/module_08")
}

func Test_Valid_Module_09(t *testing.T) {
	Check(t, false, "corset/valid/module_09")
}

func Test_Valid_Module_10(t *testing.T) {
	Check(t, false, "corset/valid/module_10")
}

// NOTE: uses conditional module
//
// func Test_Valid_Module_11(t *testing.T) {
// 	test_util.Check(t, false, "corset/valid/module_11")
// }

// ===================================================================
// Permutations
// ===================================================================

func Test_Valid_Permute_01(t *testing.T) {
	util.Check(t, false, "corset/valid/permute_01")
}

func Test_Valid_Permute_02(t *testing.T) {
	util.Check(t, false, "corset/valid/permute_02")
}

func Test_Valid_Permute_03(t *testing.T) {
	util.Check(t, false, "corset/valid/permute_03")
}

func Test_Valid_Permute_04(t *testing.T) {
	util.Check(t, false, "corset/valid/permute_04")
}

func Test_Valid_Permute_05(t *testing.T) {
	util.Check(t, false, "corset/valid/permute_05")
}

func Test_Valid_Permute_06(t *testing.T) {
	util.Check(t, false, "corset/valid/permute_06")
}

func Test_Valid_Permute_07(t *testing.T) {
	util.Check(t, false, "corset/valid/permute_07")
}

func Test_Valid_Permute_08(t *testing.T) {
	util.Check(t, false, "corset/valid/permute_08")
}

func Test_Valid_Permute_09(t *testing.T) {
	util.Check(t, false, "corset/valid/permute_09")
}

func Test_Valid_Permute_10(t *testing.T) {
	util.Check(t, false, "corset/valid/permute_10")
}

func Test_Valid_Permute_11(t *testing.T) {
	util.Check(t, false, "corset/valid/permute_11")
}

// ===================================================================
// Sorting Constraints
// ===================================================================

func Test_Valid_Sorted_01(t *testing.T) {
	util.Check(t, false, "corset/valid/sorted_01")
}
func Test_Valid_Sorted_02(t *testing.T) {
	util.Check(t, false, "corset/valid/sorted_02")
}
func Test_Valid_Sorted_03(t *testing.T) {
	util.Check(t, false, "corset/valid/sorted_03")
}
func Test_Valid_Sorted_04(t *testing.T) {
	util.Check(t, false, "corset/valid/sorted_04")
}
func Test_Valid_Sorted_05(t *testing.T) {
	util.Check(t, false, "corset/valid/sorted_05")
}
func Test_Valid_Sorted_06(t *testing.T) {
	util.Check(t, false, "corset/valid/sorted_06")
}

func Test_Valid_Sorted_07(t *testing.T) {
	util.Check(t, false, "corset/valid/sorted_07")
}
func Test_Valid_Sorted_08(t *testing.T) {
	util.Check(t, false, "corset/valid/sorted_08")
}

func Test_Valid_StrictSorted_01(t *testing.T) {
	util.Check(t, false, "corset/valid/strictsorted_01")
}

func Test_Valid_StrictSorted_02(t *testing.T) {
	util.Check(t, false, "corset/valid/strictsorted_02")
}

func Test_Valid_StrictSorted_03(t *testing.T) {
	util.Check(t, false, "corset/valid/strictsorted_03")
}

func Test_Valid_StrictSorted_04(t *testing.T) {
	util.Check(t, false, "corset/valid/strictsorted_04")
}

func Test_Valid_StrictSorted_05(t *testing.T) {
	util.Check(t, false, "corset/valid/strictsorted_05")
}

// ===================================================================
// Lookups
// ===================================================================

func Test_Valid_Lookup_01(t *testing.T) {
	Check(t, false, "corset/valid/lookup_01")
}

func Test_Valid_Lookup_02(t *testing.T) {
	Check(t, false, "corset/valid/lookup_02")
}

func Test_Valid_Lookup_03(t *testing.T) {
	util.Check(t, false, "corset/valid/lookup_03")
}

func Test_Valid_Lookup_04(t *testing.T) {
	Check(t, false, "corset/valid/lookup_04")
}

func Test_Valid_Lookup_05(t *testing.T) {
	Check(t, false, "corset/valid/lookup_05")
}

func Test_Valid_Lookup_06(t *testing.T) {
	Check(t, false, "corset/valid/lookup_06")
}

// #1247
// func Test_Valid_Lookup_07(t *testing.T) {
// 	util.Check(t, false, "corset/valid/lookup_07")
// }

// #1247
// func Test_Valid_Lookup_08(t *testing.T) {
// 	util.Check(t, false, "corset/valid/lookup_08")
// }

func Test_Valid_Lookup_09(t *testing.T) {
	Check(t, false, "corset/valid/lookup_09")
}

func Test_Valid_Lookup_10(t *testing.T) {
	Check(t, false, "corset/valid/lookup_10")
}

func Test_Valid_Lookup_11(t *testing.T) {
	Check(t, false, "corset/valid/lookup_11")
}

func Test_Valid_Lookup_12(t *testing.T) {
	Check(t, false, "corset/valid/lookup_12")
}

func Test_Valid_Lookup_13(t *testing.T) {
	util.Check(t, false, "corset/valid/lookup_13")
}

func Test_Valid_Lookup_14(t *testing.T) {
	util.Check(t, false, "corset/valid/lookup_14")
}

func Test_Valid_Lookup_15(t *testing.T) {
	Check(t, false, "corset/valid/lookup_15")
}

func Test_Valid_Lookup_16(t *testing.T) {
	util.Check(t, false, "corset/valid/lookup_16")
}

func Test_Valid_Lookup_17(t *testing.T) {
	Check(t, false, "corset/valid/lookup_17")
}

func Test_Valid_Lookup_18(t *testing.T) {
	Check(t, false, "corset/valid/lookup_18")
}

func Test_Valid_Lookup_19(t *testing.T) {
	Check(t, false, "corset/valid/lookup_19")
}

func Test_Valid_Lookup_20(t *testing.T) {
	Check(t, false, "corset/valid/lookup_20")
}

func Test_Valid_Lookup_21(t *testing.T) {
	Check(t, false, "corset/valid/lookup_21")
}

func Test_Valid_Lookup_22(t *testing.T) {
	Check(t, false, "corset/valid/lookup_22")
}

func Test_Valid_Lookup_23(t *testing.T) {
	Check(t, false, "corset/valid/lookup_23")
}

func Test_Valid_Lookup_24(t *testing.T) {
	Check(t, false, "corset/valid/lookup_24")
}
func Test_Valid_Lookup_25(t *testing.T) {
	Check(t, false, "corset/valid/lookup_25")
}

func Test_Valid_Lookup_26(t *testing.T) {
	util.Check(t, false, "corset/valid/lookup_26")
}
func Test_Valid_Lookup_27(t *testing.T) {
	util.Check(t, false, "corset/valid/lookup_27")
}

// ===================================================================
// Interleaving
// ===================================================================

func Test_Valid_Interleave_01(t *testing.T) {
	util.Check(t, false, "corset/valid/interleave_01")
}

func Test_Valid_Interleave_02(t *testing.T) {
	util.Check(t, false, "corset/valid/interleave_02")
}

func Test_Valid_Interleave_03(t *testing.T) {
	util.Check(t, false, "corset/valid/interleave_03")
}

func Test_Valid_Interleave_04(t *testing.T) {
	util.Check(t, false, "corset/valid/interleave_04")
}

func Test_Valid_Interleave_05(t *testing.T) {
	util.Check(t, false, "corset/valid/interleave_05")
}
func Test_Valid_Interleave_06(t *testing.T) {
	util.Check(t, false, "corset/valid/interleave_06")
}
func Test_Valid_Interleave_07(t *testing.T) {
	util.Check(t, false, "corset/valid/interleave_07")
}

// ===================================================================
// Functions
// ===================================================================

func Test_Valid_Fun_01(t *testing.T) {
	Check(t, false, "corset/valid/fun_01")
}

func Test_Valid_Fun_02(t *testing.T) {
	Check(t, false, "corset/valid/fun_02")
}

func Test_Valid_Fun_03(t *testing.T) {
	Check(t, false, "corset/valid/fun_03")
}

func Test_Valid_Fun_04(t *testing.T) {
	Check(t, false, "corset/valid/fun_04")
}

func Test_Valid_Fun_05(t *testing.T) {
	Check(t, false, "corset/valid/fun_05")
}

func Test_Valid_Fun_06(t *testing.T) {
	Check(t, false, "corset/valid/fun_06")
}

func Test_Valid_Fun_07(t *testing.T) {
	util.Check(t, false, "corset/valid/fun_07")
}

// ===================================================================
// Pure Functions
// ===================================================================

func Test_Valid_PureFun_01(t *testing.T) {
	Check(t, false, "corset/valid/purefun_01")
}

func Test_Valid_PureFun_02(t *testing.T) {
	Check(t, false, "corset/valid/purefun_02")
}

func Test_Valid_PureFun_03(t *testing.T) {
	Check(t, false, "corset/valid/purefun_03")
}

func Test_Valid_PureFun_04(t *testing.T) {
	Check(t, false, "corset/valid/purefun_04")
}

func Test_Valid_PureFun_05(t *testing.T) {
	Check(t, false, "corset/valid/purefun_05")
}

func Test_Valid_PureFun_06(t *testing.T) {
	Check(t, false, "corset/valid/purefun_06")
}

func Test_Valid_PureFun_07(t *testing.T) {
	Check(t, false, "corset/valid/purefun_07")
}

func Test_Valid_PureFun_08(t *testing.T) {
	Check(t, false, "corset/valid/purefun_08")
}

func Test_Valid_PureFun_09(t *testing.T) {
	Check(t, false, "corset/valid/purefun_09")
}
func Test_Valid_PureFun_10(t *testing.T) {
	Check(t, false, "corset/valid/purefun_10")
}

// ===================================================================
// For Loops
// ===================================================================

func Test_Valid_For_01(t *testing.T) {
	Check(t, false, "corset/valid/for_01")
}

func Test_Valid_For_02(t *testing.T) {
	Check(t, false, "corset/valid/for_02")
}

func Test_Valid_For_03(t *testing.T) {
	Check(t, false, "corset/valid/for_03")
}

func Test_Valid_For_04(t *testing.T) {
	Check(t, false, "corset/valid/for_04")
}

func Test_Valid_For_05(t *testing.T) {
	Check(t, false, "corset/valid/for_05")
}

func Test_Valid_For_06(t *testing.T) {
	Check(t, false, "corset/valid/for_06")
}

// ===================================================================
// Arrays
// ===================================================================

func Test_Valid_Array_01(t *testing.T) {
	Check(t, false, "corset/valid/array_01")
}

func Test_Valid_Array_02(t *testing.T) {
	Check(t, false, "corset/valid/array_02")
}

func Test_Valid_Array_03(t *testing.T) {
	Check(t, false, "corset/valid/array_03")
}

func Test_Valid_Array_04(t *testing.T) {
	Check(t, false, "corset/valid/array_04")
}

func Test_Valid_Array_05(t *testing.T) {
	Check(t, false, "corset/valid/array_05")
}

func Test_Valid_Array_06(t *testing.T) {
	Check(t, false, "corset/valid/array_06")
}

func Test_Valid_Array_07(t *testing.T) {
	Check(t, false, "corset/valid/array_07")
}

func Test_Valid_Array_08(t *testing.T) {
	Check(t, false, "corset/valid/array_08")
}

// ===================================================================
// Reduce
// ===================================================================

func Test_Valid_Reduce_01(t *testing.T) {
	Check(t, false, "corset/valid/reduce_01")
}

func Test_Valid_Reduce_02(t *testing.T) {
	Check(t, false, "corset/valid/reduce_02")
}

func Test_Valid_Reduce_03(t *testing.T) {
	Check(t, false, "corset/valid/reduce_03")
}

func Test_Valid_Reduce_04(t *testing.T) {
	Check(t, false, "corset/valid/reduce_04")
}

func Test_Valid_Reduce_05(t *testing.T) {
	Check(t, false, "corset/valid/reduce_05")
}

// ===================================================================
// Debug
// ===================================================================

func Test_Valid_Debug_01(t *testing.T) {
	Check(t, false, "corset/valid/debug_01")
}

func Test_Valid_Debug_02(t *testing.T) {
	Check(t, false, "corset/valid/debug_02")
}

func Test_Valid_Debug_03(t *testing.T) {
	Check(t, false, "corset/valid/debug_03")
}

// ===================================================================
// Perspectives
// ===================================================================

func Test_Valid_Perspective_01(t *testing.T) {
	Check(t, false, "corset/valid/perspective_01")
}

func Test_Valid_Perspective_02(t *testing.T) {
	Check(t, false, "corset/valid/perspective_02")
}

func Test_Valid_Perspective_03(t *testing.T) {
	Check(t, false, "corset/valid/perspective_03")
}

func Test_Valid_Perspective_04(t *testing.T) {
	Check(t, false, "corset/valid/perspective_04")
}

func Test_Valid_Perspective_05(t *testing.T) {
	Check(t, false, "corset/valid/perspective_05")
}

func Test_Valid_Perspective_06(t *testing.T) {
	Check(t, false, "corset/valid/perspective_06")
}

func Test_Valid_Perspective_07(t *testing.T) {
	Check(t, false, "corset/valid/perspective_07")
}

func Test_Valid_Perspective_08(t *testing.T) {
	Check(t, false, "corset/valid/perspective_08")
}

func Test_Valid_Perspective_09(t *testing.T) {
	Check(t, false, "corset/valid/perspective_09")
}

func Test_Valid_Perspective_10(t *testing.T) {
	Check(t, false, "corset/valid/perspective_10")
}

func Test_Valid_Perspective_11(t *testing.T) {
	Check(t, false, "corset/valid/perspective_11")
}

func Test_Valid_Perspective_12(t *testing.T) {
	Check(t, false, "corset/valid/perspective_12")
}

func Test_Valid_Perspective_13(t *testing.T) {
	Check(t, false, "corset/valid/perspective_13")
}

func Test_Valid_Perspective_14(t *testing.T) {
	Check(t, false, "corset/valid/perspective_14")
}

func Test_Valid_Perspective_15(t *testing.T) {
	Check(t, false, "corset/valid/perspective_15")
}

func Test_Valid_Perspective_16(t *testing.T) {
	Check(t, false, "corset/valid/perspective_16")
}

func Test_Valid_Perspective_17(t *testing.T) {
	Check(t, false, "corset/valid/perspective_17")
}

func Test_Valid_Perspective_18(t *testing.T) {
	Check(t, false, "corset/valid/perspective_18")
}

func Test_Valid_Perspective_19(t *testing.T) {
	Check(t, false, "corset/valid/perspective_19")
}

func Test_Valid_Perspective_20(t *testing.T) {
	Check(t, false, "corset/valid/perspective_20")
}

func Test_Valid_Perspective_21(t *testing.T) {
	Check(t, false, "corset/valid/perspective_21")
}

func Test_Valid_Perspective_22(t *testing.T) {
	Check(t, false, "corset/valid/perspective_22")
}

func Test_Valid_Perspective_23(t *testing.T) {
	Check(t, false, "corset/valid/perspective_23")
}

func Test_Valid_Perspective_24(t *testing.T) {
	Check(t, false, "corset/valid/perspective_24")
}

func Test_Valid_Perspective_26(t *testing.T) {
	Check(t, false, "corset/valid/perspective_26")
}

func Test_Valid_Perspective_27(t *testing.T) {
	Check(t, false, "corset/valid/perspective_27")
}

func Test_Valid_Perspective_28(t *testing.T) {
	Check(t, false, "corset/valid/perspective_28")
}

func Test_Valid_Perspective_29(t *testing.T) {
	Check(t, false, "corset/valid/perspective_29")
}

func Test_Valid_Perspective_30(t *testing.T) {
	Check(t, false, "corset/valid/perspective_30")
}

func Test_Valid_Perspective_31(t *testing.T) {
	Check(t, false, "corset/valid/perspective_31")
}

// ===================================================================
// Let
// ===================================================================

func Test_Valid_Let_01(t *testing.T) {
	Check(t, false, "corset/valid/let_01")
}

func Test_Valid_Let_02(t *testing.T) {
	Check(t, false, "corset/valid/let_02")
}

func Test_Valid_Let_03(t *testing.T) {
	Check(t, false, "corset/valid/let_03")
}

func Test_Valid_Let_04(t *testing.T) {
	Check(t, false, "corset/valid/let_04")
}

func Test_Valid_Let_05(t *testing.T) {
	Check(t, false, "corset/valid/let_05")
}

func Test_Valid_Let_06(t *testing.T) {
	Check(t, false, "corset/valid/let_06")
}

func Test_Valid_Let_07(t *testing.T) {
	Check(t, false, "corset/valid/let_07")
}

func Test_Valid_Let_08(t *testing.T) {
	Check(t, false, "corset/valid/let_08")
}
func Test_Valid_Let_09(t *testing.T) {
	Check(t, false, "corset/valid/let_09")
}

func Test_Valid_Let_10(t *testing.T) {
	Check(t, false, "corset/valid/let_10")
}

func Test_Valid_Let_11(t *testing.T) {
	Check(t, false, "corset/valid/let_11")
}

// ===================================================================
// Computed Columns
// ===================================================================

func Test_Valid_Compute_01(t *testing.T) {
	util.Check(t, false, "corset/valid/compute_01")
}

func Test_Valid_Compute_02(t *testing.T) {
	util.Check(t, false, "corset/valid/compute_02")
}

// ===================================================================
// defcomputedcolumn
// ===================================================================

func Test_Valid_ComputedColumn_01(t *testing.T) {
	util.Check(t, false, "corset/valid/computedcolumn_01")
}
func Test_Valid_ComputedColumn_02(t *testing.T) {
	util.Check(t, false, "corset/valid/computedcolumn_02")
}
func Test_Valid_ComputedColumn_03(t *testing.T) {
	util.Check(t, false, "corset/valid/computedcolumn_03")
}

//	func Test_Valid_ComputedColumn_04(t *testing.T) {
//		test_util.Check(t, false, "corset/valid/computedcolumn_04")
//	}
func Test_Valid_ComputedColumn_05(t *testing.T) {
	util.Check(t, false, "corset/valid/computedcolumn_05")
}

// ===================================================================
// Native computations
// ===================================================================

func Test_Valid_Native_01(t *testing.T) {
	util.Check(t, false, "corset/valid/native_01")
}
func Test_Valid_Native_02(t *testing.T) {
	util.Check(t, false, "corset/valid/native_02")
}
func Test_Valid_Native_03(t *testing.T) {
	util.Check(t, false, "corset/valid/native_03")
}
func Test_Valid_Native_04(t *testing.T) {
	util.Check(t, false, "corset/valid/native_04")
}

func Test_Valid_Native_05(t *testing.T) {
	util.Check(t, false, "corset/valid/native_05")
}

func Test_Valid_Native_06(t *testing.T) {
	util.Check(t, false, "corset/valid/native_06")
}

func Test_Valid_Native_07(t *testing.T) {
	util.Check(t, false, "corset/valid/native_07")
}

func Test_Valid_Native_08(t *testing.T) {
	util.Check(t, false, "corset/valid/native_08")
}

func Test_Valid_Native_09(t *testing.T) {
	util.Check(t, false, "corset/valid/native_09")
}

func Test_Valid_Native_10(t *testing.T) {
	util.Check(t, false, "corset/valid/native_10")
}

func Test_Valid_Native_11(t *testing.T) {
	util.Check(t, false, "corset/valid/native_11")
}

// ===================================================================
// Standard Library Tests
// ===================================================================

func Test_Valid_Stdlib_01(t *testing.T) {
	Check(t, true, "corset/valid/stdlib_01")
}

func Test_Valid_Stdlib_02(t *testing.T) {
	Check(t, true, "corset/valid/stdlib_02")
}

func Test_Valid_Stdlib_03(t *testing.T) {
	Check(t, true, "corset/valid/stdlib_03")
}

func Test_Valid_Stdlib_04(t *testing.T) {
	Check(t, true, "corset/valid/stdlib_04")
}

func Test_Valid_Stdlib_05(t *testing.T) {
	Check(t, true, "corset/valid/stdlib_05")
}
