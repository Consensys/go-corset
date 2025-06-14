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
	"bytes"
	"encoding/gob"
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/consensys/go-corset/pkg/asm"
	"github.com/consensys/go-corset/pkg/corset"
	"github.com/consensys/go-corset/pkg/ir/mir"
	"github.com/consensys/go-corset/pkg/schema"
	sc "github.com/consensys/go-corset/pkg/schema"
	"github.com/consensys/go-corset/pkg/trace"
	"github.com/consensys/go-corset/pkg/trace/json"
	"github.com/consensys/go-corset/pkg/util"
	"github.com/consensys/go-corset/pkg/util/source"
)

// Determines the (relative) location of the test directory.  That is
// where the corset test files (lisp) and the corresponding traces
// (accepts/rejects) are found.
const TestDir = "../../testdata"

// ===================================================================
// Basic Tests
// ===================================================================

func Test_Basic_01(t *testing.T) {
	Check(t, false, "basic_01")
}

func Test_Basic_02(t *testing.T) {
	Check(t, false, "basic_02")
}

func Test_Basic_03(t *testing.T) {
	Check(t, false, "basic_03")
}

func Test_Basic_04(t *testing.T) {
	Check(t, false, "basic_04")
}

func Test_Basic_05(t *testing.T) {
	Check(t, false, "basic_05")
}

func Test_Basic_06(t *testing.T) {
	Check(t, false, "basic_06")
}

func Test_Basic_07(t *testing.T) {
	Check(t, false, "basic_07")
}

func Test_Basic_08(t *testing.T) {
	Check(t, false, "basic_08")
}

func Test_Basic_09(t *testing.T) {
	Check(t, false, "basic_09")
}

func Test_Basic_10(t *testing.T) {
	Check(t, false, "basic_10")
}

func Test_Basic_11(t *testing.T) {
	Check(t, false, "basic_11")
}

// ===================================================================
// Constants Tests
// ===================================================================
func Test_Constant_01(t *testing.T) {
	Check(t, false, "constant_01")
}

func Test_Constant_02(t *testing.T) {
	Check(t, false, "constant_02")
}

func Test_Constant_03(t *testing.T) {
	Check(t, false, "constant_03")
}

func Test_Constant_04(t *testing.T) {
	Check(t, false, "constant_04")
}

func Test_Constant_05(t *testing.T) {
	Check(t, false, "constant_05")
}

func Test_Constant_06(t *testing.T) {
	Check(t, false, "constant_06")
}

func Test_Constant_07(t *testing.T) {
	Check(t, false, "constant_07")
}

func Test_Constant_08(t *testing.T) {
	Check(t, false, "constant_08")
}

func Test_Constant_09(t *testing.T) {
	Check(t, false, "constant_09")
}

func Test_Constant_10(t *testing.T) {
	Check(t, false, "constant_10")
}

func Test_Constant_11(t *testing.T) {
	Check(t, false, "constant_11")
}

func Test_Constant_12(t *testing.T) {
	Check(t, false, "constant_12")
}

func Test_Constant_13(t *testing.T) {
	Check(t, false, "constant_13")
}

func Test_Constant_14(t *testing.T) {
	Check(t, false, "constant_14")
}

func Test_Constant_15(t *testing.T) {
	Check(t, false, "constant_15")
}

func Test_Constant_16(t *testing.T) {
	Check(t, false, "constant_16")
}

// ===================================================================
// Alias Tests
// ===================================================================
func Test_Alias_01(t *testing.T) {
	Check(t, false, "alias_01")
}
func Test_Alias_02(t *testing.T) {
	Check(t, false, "alias_02")
}
func Test_Alias_03(t *testing.T) {
	Check(t, false, "alias_03")
}
func Test_Alias_04(t *testing.T) {
	Check(t, false, "alias_04")
}
func Test_Alias_05(t *testing.T) {
	Check(t, false, "alias_05")
}
func Test_Alias_06(t *testing.T) {
	Check(t, false, "alias_06")
}

// ===================================================================
// Domain Tests
// ===================================================================

func Test_Domain_01(t *testing.T) {
	Check(t, false, "domain_01")
}

func Test_Domain_02(t *testing.T) {
	Check(t, false, "domain_02")
}

func Test_Domain_03(t *testing.T) {
	Check(t, false, "domain_03")
}

// ===================================================================
// Block Tests
// ===================================================================

func Test_Block_01(t *testing.T) {
	Check(t, false, "block_01")
}

func Test_Block_02(t *testing.T) {
	Check(t, false, "block_02")
}

func Test_Block_03(t *testing.T) {
	Check(t, false, "block_03")
}

func Test_Block_04(t *testing.T) {
	Check(t, false, "block_04")
}

// ===================================================================
// Inequality Tests
// ===================================================================

// func Test_Inequality_01(t *testing.T) {
// 	Check(t, false, "ieq_01")
// }

func Test_Inequality_02(t *testing.T) {
	Check(t, false, "ieq_02")
}

// ===================================================================
// Logical Tests
// ===================================================================

func Test_Logic_01(t *testing.T) {
	Check(t, false, "logic_01")
}

// ===================================================================
// Property Tests
// ===================================================================

func Test_Property_01(t *testing.T) {
	Check(t, false, "property_01")
}

// ===================================================================
// Shift Tests
// ===================================================================

func Test_Shift_01(t *testing.T) {
	Check(t, false, "shift_01")
}

func Test_Shift_02(t *testing.T) {
	Check(t, false, "shift_02")
}

func Test_Shift_03(t *testing.T) {
	Check(t, false, "shift_03")
}

func Test_Shift_04(t *testing.T) {
	Check(t, false, "shift_04")
}

func Test_Shift_05(t *testing.T) {
	Check(t, false, "shift_05")
}

func Test_Shift_06(t *testing.T) {
	Check(t, false, "shift_06")
}

func Test_Shift_07(t *testing.T) {
	Check(t, false, "shift_07")
}

func Test_Shift_08(t *testing.T) {
	Check(t, false, "shift_08")
}
func Test_Shift_09(t *testing.T) {
	Check(t, false, "shift_09")
}

// ===================================================================
// Spillage Tests
// ===================================================================

func Test_Spillage_01(t *testing.T) {
	Check(t, false, "spillage_01")
}

func Test_Spillage_02(t *testing.T) {
	Check(t, false, "spillage_02")
}

func Test_Spillage_03(t *testing.T) {
	Check(t, false, "spillage_03")
}

func Test_Spillage_04(t *testing.T) {
	Check(t, false, "spillage_04")
}

func Test_Spillage_05(t *testing.T) {
	Check(t, false, "spillage_05")
}

func Test_Spillage_06(t *testing.T) {
	Check(t, false, "spillage_06")
}

func Test_Spillage_07(t *testing.T) {
	Check(t, false, "spillage_07")
}

func Test_Spillage_08(t *testing.T) {
	Check(t, false, "spillage_08")
}

func Test_Spillage_09(t *testing.T) {
	Check(t, false, "spillage_09")
}

// ===================================================================
// Normalisation Tests
// ===================================================================

func Test_Norm_01(t *testing.T) {
	Check(t, false, "norm_01")
}

func Test_Norm_02(t *testing.T) {
	Check(t, false, "norm_02")
}

func Test_Norm_03(t *testing.T) {
	Check(t, false, "norm_03")
}

func Test_Norm_04(t *testing.T) {
	Check(t, false, "norm_04")
}

func Test_Norm_05(t *testing.T) {
	Check(t, false, "norm_05")
}

func Test_Norm_06(t *testing.T) {
	Check(t, false, "norm_06")
}

func Test_Norm_07(t *testing.T) {
	Check(t, false, "norm_07")
}

// ===================================================================
// If-Zero
// ===================================================================

func Test_If_01(t *testing.T) {
	Check(t, false, "if_01")
}

func Test_If_02(t *testing.T) {
	Check(t, false, "if_02")
}

func Test_If_03(t *testing.T) {
	Check(t, false, "if_03")
}

func Test_If_04(t *testing.T) {
	Check(t, false, "if_04")
}

func Test_If_05(t *testing.T) {
	Check(t, false, "if_05")
}

func Test_If_06(t *testing.T) {
	Check(t, false, "if_06")
}

func Test_If_07(t *testing.T) {
	Check(t, false, "if_07")
}

func Test_If_08(t *testing.T) {
	Check(t, false, "if_08")
}

func Test_If_09(t *testing.T) {
	Check(t, false, "if_09")
}

func Test_If_10(t *testing.T) {
	Check(t, false, "if_10")
}

func Test_If_11(t *testing.T) {
	Check(t, false, "if_11")
}
func Test_If_12(t *testing.T) {
	Check(t, false, "if_12")
}
func Test_If_13(t *testing.T) {
	Check(t, false, "if_13")
}

func Test_If_14(t *testing.T) {
	Check(t, false, "if_14")
}

func Test_If_15(t *testing.T) {
	Check(t, false, "if_15")
}

func Test_If_16(t *testing.T) {
	Check(t, false, "if_16")
}

func Test_If_17(t *testing.T) {
	Check(t, false, "if_17")
}

func Test_If_18(t *testing.T) {
	Check(t, false, "if_18")
}

func Test_If_19(t *testing.T) {
	Check(t, false, "if_19")
}

// ===================================================================
// Guards
// ===================================================================

func Test_Guard_01(t *testing.T) {
	Check(t, false, "guard_01")
}

func Test_Guard_02(t *testing.T) {
	Check(t, false, "guard_02")
}

func Test_Guard_03(t *testing.T) {
	Check(t, false, "guard_03")
}

func Test_Guard_04(t *testing.T) {
	Check(t, false, "guard_04")
}

func Test_Guard_05(t *testing.T) {
	Check(t, false, "guard_05")
}

// ===================================================================
// Types
// ===================================================================

func Test_Type_01(t *testing.T) {
	Check(t, false, "type_01")
}

func Test_Type_02(t *testing.T) {
	Check(t, false, "type_02")
}

func Test_Type_03(t *testing.T) {
	Check(t, false, "type_03")
}

func Test_Type_04(t *testing.T) {
	Check(t, false, "type_04")
}

func Test_Type_05(t *testing.T) {
	Check(t, false, "type_05")
}

func Test_Type_06(t *testing.T) {
	Check(t, false, "type_06")
}

func Test_Type_07(t *testing.T) {
	Check(t, false, "type_07")
}

func Test_Type_08(t *testing.T) {
	Check(t, false, "type_08")
}

func Test_Type_09(t *testing.T) {
	Check(t, false, "type_09")
}

func Test_Type_10(t *testing.T) {
	Check(t, false, "type_10")
}

func Test_Type_11(t *testing.T) {
	Check(t, false, "type_11")
}

// ===================================================================
// Range Constraints
// ===================================================================

func Test_Range_01(t *testing.T) {
	Check(t, false, "range_01")
}

func Test_Range_02(t *testing.T) {
	Check(t, false, "range_02")
}

func Test_Range_03(t *testing.T) {
	Check(t, false, "range_03")
}

func Test_Range_04(t *testing.T) {
	Check(t, false, "range_04")
}

func Test_Range_05(t *testing.T) {
	Check(t, false, "range_05")
}

// ===================================================================
// Constant Propagation
// ===================================================================

func Test_ConstExpr_01(t *testing.T) {
	Check(t, false, "constexpr_01")
}

func Test_ConstExpr_02(t *testing.T) {
	Check(t, false, "constexpr_02")
}

func Test_ConstExpr_03(t *testing.T) {
	Check(t, false, "constexpr_03")
}

func Test_ConstExpr_04(t *testing.T) {
	Check(t, false, "constexpr_04")
}

func Test_ConstExpr_05(t *testing.T) {
	Check(t, false, "constexpr_05")
}

// ===================================================================
// Modules
// ===================================================================

func Test_Module_01(t *testing.T) {
	Check(t, false, "module_01")
}

func Test_Module_02(t *testing.T) {
	Check(t, false, "module_02")
}

func Test_Module_03(t *testing.T) {
	Check(t, false, "module_03")
}

func Test_Module_04(t *testing.T) {
	Check(t, false, "module_04")
}

func Test_Module_05(t *testing.T) {
	Check(t, false, "module_05")
}

func Test_Module_06(t *testing.T) {
	Check(t, false, "module_06")
}

func Test_Module_07(t *testing.T) {
	Check(t, false, "module_07")
}

func Test_Module_08(t *testing.T) {
	Check(t, false, "module_08")
}

func Test_Module_09(t *testing.T) {
	Check(t, false, "module_09")
}

func Test_Module_10(t *testing.T) {
	Check(t, false, "module_10")
}

// NOTE: uses conditional module
//
// func Test_Module_11(t *testing.T) {
// 	Check(t, false, "module_11")
// }

// ===================================================================
// Permutations
// ===================================================================

func Test_Permute_01(t *testing.T) {
	Check(t, false, "permute_01")
}

func Test_Permute_02(t *testing.T) {
	Check(t, false, "permute_02")
}

func Test_Permute_03(t *testing.T) {
	Check(t, false, "permute_03")
}

func Test_Permute_04(t *testing.T) {
	Check(t, false, "permute_04")
}

func Test_Permute_05(t *testing.T) {
	Check(t, false, "permute_05")
}

func Test_Permute_06(t *testing.T) {
	Check(t, false, "permute_06")
}

func Test_Permute_07(t *testing.T) {
	Check(t, false, "permute_07")
}

func Test_Permute_08(t *testing.T) {
	Check(t, false, "permute_08")
}

func Test_Permute_09(t *testing.T) {
	Check(t, false, "permute_09")
}

func Test_Permute_10(t *testing.T) {
	Check(t, false, "permute_10")
}

func Test_Permute_11(t *testing.T) {
	Check(t, false, "permute_11")
}

// ===================================================================
// Sorting Constraints
// ===================================================================

func Test_Sorted_01(t *testing.T) {
	Check(t, false, "sorted_01")
}
func Test_Sorted_02(t *testing.T) {
	Check(t, false, "sorted_02")
}
func Test_Sorted_03(t *testing.T) {
	Check(t, false, "sorted_03")
}
func Test_Sorted_04(t *testing.T) {
	Check(t, false, "sorted_04")
}
func Test_Sorted_05(t *testing.T) {
	Check(t, false, "sorted_05")
}
func Test_Sorted_06(t *testing.T) {
	Check(t, false, "sorted_06")
}

func Test_Sorted_07(t *testing.T) {
	Check(t, false, "sorted_07")
}
func Test_Sorted_08(t *testing.T) {
	Check(t, false, "sorted_08")
}

func Test_StrictSorted_01(t *testing.T) {
	Check(t, false, "strictsorted_01")
}

func Test_StrictSorted_02(t *testing.T) {
	Check(t, false, "strictsorted_02")
}

func Test_StrictSorted_03(t *testing.T) {
	Check(t, false, "strictsorted_03")
}

func Test_StrictSorted_04(t *testing.T) {
	Check(t, false, "strictsorted_04")
}

func Test_StrictSorted_05(t *testing.T) {
	Check(t, false, "strictsorted_05")
}

// ===================================================================
// Lookups
// ===================================================================

func Test_Lookup_01(t *testing.T) {
	Check(t, false, "lookup_01")
}

func Test_Lookup_02(t *testing.T) {
	Check(t, false, "lookup_02")
}

func Test_Lookup_03(t *testing.T) {
	Check(t, false, "lookup_03")
}

func Test_Lookup_04(t *testing.T) {
	Check(t, false, "lookup_04")
}

func Test_Lookup_05(t *testing.T) {
	Check(t, false, "lookup_05")
}

func Test_Lookup_06(t *testing.T) {
	Check(t, false, "lookup_06")
}

func Test_Lookup_07(t *testing.T) {
	Check(t, false, "lookup_07")
}

func Test_Lookup_08(t *testing.T) {
	Check(t, false, "lookup_08")
}

func Test_Lookup_09(t *testing.T) {
	Check(t, false, "lookup_09")
}

func Test_Lookup_10(t *testing.T) {
	Check(t, false, "lookup_10")
}

func Test_Lookup_11(t *testing.T) {
	Check(t, false, "lookup_11")
}

func Test_Lookup_12(t *testing.T) {
	Check(t, false, "lookup_12")
}

func Test_Lookup_13(t *testing.T) {
	Check(t, false, "lookup_13")
}

func Test_Lookup_14(t *testing.T) {
	Check(t, false, "lookup_14")
}

func Test_Lookup_15(t *testing.T) {
	Check(t, false, "lookup_15")
}

func Test_Lookup_16(t *testing.T) {
	Check(t, false, "lookup_16")
}

// ===================================================================
// Interleaving
// ===================================================================

func Test_Interleave_01(t *testing.T) {
	Check(t, false, "interleave_01")
}

func Test_Interleave_02(t *testing.T) {
	Check(t, false, "interleave_02")
}

func Test_Interleave_03(t *testing.T) {
	Check(t, false, "interleave_03")
}

func Test_Interleave_04(t *testing.T) {
	Check(t, false, "interleave_04")
}

func Test_Interleave_05(t *testing.T) {
	Check(t, false, "interleave_05")
}
func Test_Interleave_06(t *testing.T) {
	Check(t, false, "interleave_06")
}
func Test_Interleave_07(t *testing.T) {
	Check(t, false, "interleave_07")
}

// ===================================================================
// Functions
// ===================================================================

func Test_Fun_01(t *testing.T) {
	Check(t, false, "fun_01")
}

func Test_Fun_02(t *testing.T) {
	Check(t, false, "fun_02")
}

func Test_Fun_03(t *testing.T) {
	Check(t, false, "fun_03")
}

func Test_Fun_04(t *testing.T) {
	Check(t, false, "fun_04")
}

func Test_Fun_05(t *testing.T) {
	Check(t, false, "fun_05")
}

func Test_Fun_06(t *testing.T) {
	Check(t, false, "fun_06")
}

// ===================================================================
// Pure Functions
// ===================================================================

func Test_PureFun_01(t *testing.T) {
	Check(t, false, "purefun_01")
}

func Test_PureFun_02(t *testing.T) {
	Check(t, false, "purefun_02")
}

func Test_PureFun_03(t *testing.T) {
	Check(t, false, "purefun_03")
}

func Test_PureFun_04(t *testing.T) {
	Check(t, false, "purefun_04")
}

func Test_PureFun_05(t *testing.T) {
	Check(t, false, "purefun_05")
}

func Test_PureFun_06(t *testing.T) {
	Check(t, false, "purefun_06")
}

func Test_PureFun_07(t *testing.T) {
	Check(t, false, "purefun_07")
}

func Test_PureFun_08(t *testing.T) {
	Check(t, false, "purefun_08")
}

func Test_PureFun_09(t *testing.T) {
	Check(t, false, "purefun_09")
}

// ===================================================================
// For Loops
// ===================================================================

func Test_For_01(t *testing.T) {
	Check(t, false, "for_01")
}

func Test_For_02(t *testing.T) {
	Check(t, false, "for_02")
}

func Test_For_03(t *testing.T) {
	Check(t, false, "for_03")
}

func Test_For_04(t *testing.T) {
	Check(t, false, "for_04")
}

func Test_For_05(t *testing.T) {
	Check(t, false, "for_05")
}

func Test_For_06(t *testing.T) {
	Check(t, false, "for_06")
}

// ===================================================================
// Arrays
// ===================================================================

func Test_Array_01(t *testing.T) {
	Check(t, false, "array_01")
}

func Test_Array_02(t *testing.T) {
	Check(t, false, "array_02")
}

func Test_Array_03(t *testing.T) {
	Check(t, false, "array_03")
}

func Test_Array_04(t *testing.T) {
	Check(t, false, "array_04")
}

func Test_Array_05(t *testing.T) {
	Check(t, false, "array_05")
}

func Test_Array_06(t *testing.T) {
	Check(t, false, "array_06")
}

func Test_Array_07(t *testing.T) {
	Check(t, false, "array_07")
}

func Test_Array_08(t *testing.T) {
	Check(t, false, "array_08")
}

// ===================================================================
// Reduce
// ===================================================================

func Test_Reduce_01(t *testing.T) {
	Check(t, false, "reduce_01")
}

func Test_Reduce_02(t *testing.T) {
	Check(t, false, "reduce_02")
}

func Test_Reduce_03(t *testing.T) {
	Check(t, false, "reduce_03")
}

func Test_Reduce_04(t *testing.T) {
	Check(t, false, "reduce_04")
}

func Test_Reduce_05(t *testing.T) {
	Check(t, false, "reduce_05")
}

// ===================================================================
// Debug
// ===================================================================

func Test_Debug_01(t *testing.T) {
	Check(t, false, "debug_01")
}

func Test_Debug_02(t *testing.T) {
	Check(t, false, "debug_02")
}

func Test_Debug_03(t *testing.T) {
	Check(t, false, "debug_03")
}

// ===================================================================
// Perspectives
// ===================================================================

func Test_Perspective_01(t *testing.T) {
	Check(t, false, "perspective_01")
}

func Test_Perspective_02(t *testing.T) {
	Check(t, false, "perspective_02")
}

func Test_Perspective_03(t *testing.T) {
	Check(t, false, "perspective_03")
}

func Test_Perspective_04(t *testing.T) {
	Check(t, false, "perspective_04")
}

func Test_Perspective_05(t *testing.T) {
	Check(t, false, "perspective_05")
}

func Test_Perspective_06(t *testing.T) {
	Check(t, false, "perspective_06")
}

func Test_Perspective_07(t *testing.T) {
	Check(t, false, "perspective_07")
}

func Test_Perspective_08(t *testing.T) {
	Check(t, false, "perspective_08")
}

func Test_Perspective_09(t *testing.T) {
	Check(t, false, "perspective_09")
}

func Test_Perspective_10(t *testing.T) {
	Check(t, false, "perspective_10")
}

func Test_Perspective_11(t *testing.T) {
	Check(t, false, "perspective_11")
}

func Test_Perspective_12(t *testing.T) {
	Check(t, false, "perspective_12")
}

func Test_Perspective_13(t *testing.T) {
	Check(t, false, "perspective_13")
}

func Test_Perspective_14(t *testing.T) {
	Check(t, false, "perspective_14")
}

func Test_Perspective_15(t *testing.T) {
	Check(t, false, "perspective_15")
}

func Test_Perspective_16(t *testing.T) {
	Check(t, false, "perspective_16")
}

func Test_Perspective_17(t *testing.T) {
	Check(t, false, "perspective_17")
}

func Test_Perspective_18(t *testing.T) {
	Check(t, false, "perspective_18")
}

func Test_Perspective_19(t *testing.T) {
	Check(t, false, "perspective_19")
}

func Test_Perspective_20(t *testing.T) {
	Check(t, false, "perspective_20")
}

func Test_Perspective_21(t *testing.T) {
	Check(t, false, "perspective_21")
}

func Test_Perspective_22(t *testing.T) {
	Check(t, false, "perspective_22")
}

func Test_Perspective_23(t *testing.T) {
	Check(t, false, "perspective_23")
}

func Test_Perspective_24(t *testing.T) {
	Check(t, false, "perspective_24")
}

func Test_Perspective_26(t *testing.T) {
	Check(t, false, "perspective_26")
}

func Test_Perspective_27(t *testing.T) {
	Check(t, false, "perspective_27")
}

func Test_Perspective_28(t *testing.T) {
	Check(t, false, "perspective_28")
}

func Test_Perspective_29(t *testing.T) {
	Check(t, false, "perspective_29")
}

func Test_Perspective_30(t *testing.T) {
	Check(t, false, "perspective_30")
}

func Test_Perspective_31(t *testing.T) {
	Check(t, false, "perspective_31")
}

// ===================================================================
// Let
// ===================================================================

func Test_Let_01(t *testing.T) {
	Check(t, false, "let_01")
}

func Test_Let_02(t *testing.T) {
	Check(t, false, "let_02")
}

func Test_Let_03(t *testing.T) {
	Check(t, false, "let_03")
}

func Test_Let_04(t *testing.T) {
	Check(t, false, "let_04")
}

func Test_Let_05(t *testing.T) {
	Check(t, false, "let_05")
}

func Test_Let_06(t *testing.T) {
	Check(t, false, "let_06")
}

func Test_Let_07(t *testing.T) {
	Check(t, false, "let_07")
}

func Test_Let_08(t *testing.T) {
	Check(t, false, "let_08")
}
func Test_Let_09(t *testing.T) {
	Check(t, false, "let_09")
}

func Test_Let_10(t *testing.T) {
	Check(t, false, "let_10")
}

func Test_Let_11(t *testing.T) {
	Check(t, false, "let_11")
}

// ===================================================================
// Computed Columns
// ===================================================================

func Test_Compute_01(t *testing.T) {
	Check(t, false, "compute_01")
}

func Test_Compute_02(t *testing.T) {
	Check(t, false, "compute_02")
}

// ===================================================================
// Native computations
// ===================================================================

func Test_Native_01(t *testing.T) {
	Check(t, false, "native_01")
}
func Test_Native_02(t *testing.T) {
	Check(t, false, "native_02")
}
func Test_Native_03(t *testing.T) {
	Check(t, false, "native_03")
}
func Test_Native_04(t *testing.T) {
	Check(t, false, "native_04")
}

func Test_Native_05(t *testing.T) {
	Check(t, false, "native_05")
}

func Test_Native_06(t *testing.T) {
	Check(t, false, "native_06")
}

func Test_Native_07(t *testing.T) {
	Check(t, false, "native_07")
}

func Test_Native_08(t *testing.T) {
	Check(t, false, "native_08")
}

func Test_Native_09(t *testing.T) {
	Check(t, false, "native_09")
}

func Test_Native_10(t *testing.T) {
	Check(t, false, "native_10")
}

func Test_Native_11(t *testing.T) {
	Check(t, false, "native_11")
}

// ===================================================================
// Standard Library Tests
// ===================================================================

func Test_Stdlib_01(t *testing.T) {
	Check(t, true, "stdlib_01")
}

func Test_Stdlib_02(t *testing.T) {
	Check(t, true, "stdlib_02")
}

func Test_Stdlib_03(t *testing.T) {
	Check(t, true, "stdlib_03")
}

func Test_Stdlib_04(t *testing.T) {
	Check(t, true, "stdlib_04")
}

func Test_Stdlib_05(t *testing.T) {
	Check(t, true, "stdlib_05")
}

// ===================================================================
// Complex Tests
// ===================================================================

func Test_Counter(t *testing.T) {
	Check(t, true, "counter")
}

func Test_ByteDecomp(t *testing.T) {
	Check(t, true, "byte_decomposition")
}

func Test_BitDecomp(t *testing.T) {
	Check(t, true, "bit_decomposition")
}

func Test_ByteSorting(t *testing.T) {
	Check(t, true, "byte_sorting")
}

func Test_WordSorting(t *testing.T) {
	Check(t, true, "word_sorting")
}

func Test_Multiplier(t *testing.T) {
	Check(t, false, "multiplier")
}

func Test_Memory(t *testing.T) {
	Check(t, true, "memory")
}

func Test_Adder(t *testing.T) {
	Check(t, true, "adder")
}

func TestSlow_Fields(t *testing.T) {
	Check(t, true, "fields")
}

func TestSlow_Add(t *testing.T) {
	Check(t, true, "add")
}

func TestSlow_BinStatic(t *testing.T) {
	Check(t, true, "bin-static")
}

func TestSlow_Bin(t *testing.T) {
	Check(t, true, "bin")
}

func TestSlow_Wcp(t *testing.T) {
	Check(t, true, "wcp")
}

func TestSlow_Mxp(t *testing.T) {
	Check(t, true, "mxp")
}

func TestSlow_Shf(t *testing.T) {
	Check(t, true, "shf")
}

func TestSlow_Euc(t *testing.T) {
	Check(t, true, "euc")
}

func TestSlow_Oob(t *testing.T) {
	Check(t, true, "oob")
}

func TestSlow_Stp(t *testing.T) {
	Check(t, true, "stp")
}

func TestSlow_Mmio(t *testing.T) {
	Check(t, true, "mmio")
}

// func TestSlow_Rom(t *testing.T) {
// 	Check(t, true, "rom")
// }

func TestSlow_Gas(t *testing.T) {
	Check(t, true, "gas")
}

func TestSlow_Exp(t *testing.T) {
	Check(t, true, "exp")
}

func TestSlow_Mul(t *testing.T) {
	Check(t, true, "mul")
}

func TestSlow_Mod(t *testing.T) {
	Check(t, true, "mod")
}

// NOTE: uses invalid range bound
//
// func Test_TicTacToe(t *testing.T) {
// 	Check(t, true, "tic_tac_toe")
// }

// ===================================================================
// Test Helpers
// ===================================================================

// Determines the maximum amount of padding to use when testing.  Specifically,
// every trace is tested with varying amounts of padding upto this value.
const MAX_PADDING uint = 7

// For a given set of constraints, check that all traces which we
// expect to be accepted are accepted, and all traces that we expect
// to be rejected are rejected.
func Check(t *testing.T, stdlib bool, test string) {
	var (
		corsetConfig corset.CompilationConfig
		filename     = fmt.Sprintf("%s.lisp", test)
	)
	//
	corsetConfig.Legacy = true
	corsetConfig.Stdlib = stdlib
	// Enable testing each trace in parallel
	t.Parallel()
	// Read constraints file
	bytes, err := os.ReadFile(fmt.Sprintf("%s/%s", TestDir, filename))
	// Check test file read ok
	if err != nil {
		t.Fatal(err)
	}
	// Package up as source file
	srcfile := source.NewSourceFile(filename, bytes)
	// Parse terms into an HIR schema
	schema, _, errs := corset.CompileSourceFile[*asm.MacroFunction](corsetConfig, srcfile)
	// Check terms parsed ok
	if len(errs) > 0 {
		t.Fatalf("Error parsing %s: %v\n", filename, errs)
	}
	// Record how many tests executed.
	nTests := 0
	// Iterate possible testfile extensions
	for _, cfg := range TESTFILE_EXTENSIONS {
		var traces [][]trace.RawColumn
		// Construct test filename
		testFilename := fmt.Sprintf("%s/%s.%s", TestDir, test, cfg.extension)
		traces = ReadTracesFile(testFilename)
		// Run tests
		BinCheckTraces(t, testFilename, cfg, traces, schema)
		// Record how many tests we found
		nTests += len(traces)
	}
	// Sanity check at least one trace found.
	if nTests == 0 {
		panic(fmt.Sprintf("missing any tests for %s", test))
	}
}

func BinCheckTraces(t *testing.T, test string, cfg TestConfig,
	traces [][]trace.RawColumn, mixSchema asm.MixedMacroProgram) {
	// Strip off any externally defined modules (for which there should be
	// none).
	mirSchema := schema.NewUniformSchema(mixSchema.RightModules())
	// Run checks using schema compiled from source
	CheckTraces(t, test, MAX_PADDING, cfg, traces, mirSchema)
	// Construct binary schema
	if binSchema := encodeDecodeSchema(t, mirSchema); binSchema != nil {
		// Run checks using schema from binary file.  Observe, to try and reduce
		// overhead of repeating all the tests we don't consider padding.
		CheckTraces(t, test, 0, cfg, traces, *binSchema)
	}
}

// Check a given set of tests have an expected outcome (i.e. are
// either accepted or rejected) by a given set of constraints.
func CheckTraces(t *testing.T, test string, maxPadding uint, cfg TestConfig, traces [][]trace.RawColumn,
	mirSchema mir.Schema) {
	// For unexpected traces, we never want to explore padding (because that's
	// the whole point of unexpanded traces --- they are raw).
	if !cfg.expand {
		maxPadding = 0
	}
	//
	for i, tr := range traces {
		if tr != nil {
			// Lower MIR => AIR
			airSchema := mir.LowerToAir(mirSchema, mir.DEFAULT_OPTIMISATION_LEVEL)
			// Align trace with schema, and check whether expanded or not.
			for padding := uint(0); padding <= maxPadding; padding++ {
				// Construct trace identifiers
				mirID := traceId{"MIR", test, cfg.expected, cfg.expand, cfg.validate, i + 1, padding}
				airID := traceId{"AIR", test, cfg.expected, cfg.expand, cfg.validate, i + 1, padding}
				//
				if cfg.expand {
					// Only HIR / MIR constraints for traces which must be
					// expanded.  They don't really make sense otherwise.
					checkTrace(t, tr, mirID, mirSchema)
				}
				// Always check AIR constraints
				checkTrace(t, tr, airID, airSchema)
			}
		}
	}
}

func checkTrace[C sc.Constraint](t *testing.T, inputs []trace.RawColumn, id traceId, schema sc.Schema[C]) {
	// Construct the trace
	tr, errs := sc.NewTraceBuilder().
		WithExpansion(id.expand).
		WithValidation(id.validate).
		WithPadding(id.padding).
		WithParallelism(true).
		Build(sc.Any(schema), inputs)
	// Sanity check construction
	if len(errs) > 0 {
		t.Errorf("Trace expansion failed (%s, %s, line %d with padding %d): %s",
			id.ir, id.test, id.line, id.padding, errs)
	} else {
		// Check Constraints
		errs := sc.Accepts(true, 100, schema, tr)
		// Determine whether trace accepted or not.
		accepted := len(errs) == 0
		// Process what happened versus what was supposed to happen.
		if !accepted && id.expected {
			//table.PrintTrace(tr)
			t.Errorf("Trace rejected incorrectly (%s, %s, line %d with padding %d): %s",
				id.ir, id.test, id.line, id.padding, errs)
		} else if accepted && !id.expected {
			//printTrace(tr)
			t.Errorf("Trace accepted incorrectly (%s, %s, line %d with padding %d)",
				id.ir, id.test, id.line, id.padding)
		}
	}
}

// TestConfig provides a simple mechanism for searching for testfiles.
type TestConfig struct {
	extension string
	expected  bool
	expand    bool
	validate  bool
}

var TESTFILE_EXTENSIONS []TestConfig = []TestConfig{
	// should all pass
	{"accepts", true, true, true},
	{"accepts.bz2", true, true, true},
	{"auto.accepts", true, true, true},
	{"expanded.accepts", true, false, false},
	// should all fail
	{"rejects", false, true, false},
	{"rejects.bz2", false, true, false},
	{"auto.rejects", false, true, false},
	{"expanded.rejects", false, false, false},
}

// A trace identifier uniquely identifies a specific trace within a given test.
// This is used to provide debug information about a trace failure.
// Specifically, so the user knows which line in which file caused the problem.
type traceId struct {
	// Identifies the Intermediate Representation tested against.
	ir string
	// Identifies the test name.  From this, the test filename can be determined
	// in conjunction with the expected outcome.
	test string
	// Identifies whether this trace should be accepted (true) or rejected
	// (false).
	expected bool
	// Identifies whether this trace should be expanded (or not).
	expand bool
	// Identifies whether this trace should be validate (or not).
	validate bool
	// Identifies the line number within the test file that the failing trace
	// original.
	line int
	// Identifies how much padding has been added to the expanded trace.
	padding uint
}

// ReadTracesFile reads a file containing zero or more traces expressed as JSON, where
// each trace is on a separate line.
func ReadTracesFile(filename string) [][]trace.RawColumn {
	lines := util.ReadInputFile(filename)
	traces := make([][]trace.RawColumn, len(lines))
	// Read constraints line by line
	for i, line := range lines {
		// Parse input line as JSON
		if line != "" && !strings.HasPrefix(line, ";;") {
			tr, err := json.FromBytes([]byte(line))
			if err != nil {
				msg := fmt.Sprintf("%s:%d: %s", filename, i+1, err)
				panic(msg)
			}

			traces[i] = tr
		}
	}

	return traces
}

// This is a little test to ensure the binary file format (specifically the
// binary encoder / decoder) works as expected.
func encodeDecodeSchema(t *testing.T, schema mir.Schema) *mir.Schema {
	var (
		buffer     bytes.Buffer
		gobEncoder = gob.NewEncoder(&buffer)
		binSchema  mir.Schema
	)
	// Encode schema
	if err := gobEncoder.Encode(schema); err != nil {
		t.Error(err)
		return nil
	}
	// Decode schema
	decoder := gob.NewDecoder(&buffer)
	if err := decoder.Decode(&binSchema); err != nil {
		t.Error(err)
		return nil
	}
	//
	return &binSchema
}
