package test

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/consensys/go-corset/pkg/corset"
	"github.com/consensys/go-corset/pkg/hir"
	sc "github.com/consensys/go-corset/pkg/schema"
	"github.com/consensys/go-corset/pkg/sexp"
	"github.com/consensys/go-corset/pkg/trace"
	"github.com/consensys/go-corset/pkg/trace/json"
	"github.com/consensys/go-corset/pkg/util"
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
// Function Alias Tests
// ===================================================================
func Test_FunAlias_01(t *testing.T) {
	Check(t, false, "funalias_01")
}

func Test_FunAlias_02(t *testing.T) {
	Check(t, false, "funalias_02")
}

func Test_FunAlias_03(t *testing.T) {
	Check(t, false, "funalias_03")
}

func Test_FunAlias_04(t *testing.T) {
	Check(t, false, "funalias_04")
}

func Test_FunAlias_05(t *testing.T) {
	Check(t, false, "funalias_05")
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

/* #479
func Test_PureFun_09(t *testing.T) {
	Check(t, false, "purefun_0")
}
*/
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

// NOTE: this test could not currently pass because it results in a name clash
// between a symbol in the enclosing module, and one defined in a perspective.
// This is because of the weak naming scheme currently used for perspective
// columns to maintain backwards compatibility.  When this restriction is
// lifted, this test can pass.
//
// func Test_Perspective_25(t *testing.T) {
//   Check(t, false, "perspective_25")
// }

func Test_Perspective_26(t *testing.T) {
	Check(t, false, "perspective_26")
}

func Test_Perspective_27(t *testing.T) {
	Check(t, false, "perspective_27")
}

func Test_Perspective_28(t *testing.T) {
	Check(t, false, "perspective_28")
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

func Test_Memory(t *testing.T) {
	Check(t, true, "memory")
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

func TestSlow_Rom(t *testing.T) {
	Check(t, true, "rom")
}

func TestSlow_Mmu(t *testing.T) {
	Check(t, true, "mmu")
}

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
	filename := fmt.Sprintf("%s.lisp", test)
	// Enable testing each trace in parallel
	t.Parallel()
	// Read constraints file
	bytes, err := os.ReadFile(fmt.Sprintf("%s/%s", TestDir, filename))
	// Check test file read ok
	if err != nil {
		t.Fatal(err)
	}
	// Package up as source file
	srcfile := sexp.NewSourceFile(filename, bytes)
	// Parse terms into an HIR schema
	schema, errs := corset.CompileSourceFile(stdlib, false, srcfile)
	// Check terms parsed ok
	if len(errs) > 0 {
		t.Fatalf("Error parsing %s: %v\n", filename, errs)
	}

	// Check valid traces are accepted
	accepts_file := fmt.Sprintf("%s/%s.%s", TestDir, test, "accepts")
	accepts := ReadTracesFile(accepts_file)
	ntests := len(accepts)
	BinCheckTraces(t, accepts_file, true, true, accepts, schema)
	// Check invalid traces are rejected
	rejects_file := fmt.Sprintf("%s/%s.%s", TestDir, test, "rejects")
	rejects := ReadTracesFile(rejects_file)
	ntests += len(rejects)
	BinCheckTraces(t, rejects_file, false, true, rejects, schema)
	// Check expanded traces are rejected
	expands_file := fmt.Sprintf("%s/%s.%s", TestDir, test, "expanded")
	expands := ReadTracesFile(expands_file)
	ntests += len(expands)
	BinCheckTraces(t, expands_file, false, false, expands, schema)
	// Check auto-generated valid traces (if applicable)
	auto_accepts_file := fmt.Sprintf("%s/%s.%s", TestDir, test, "auto.accepts")
	if auto_accepts := ReadTracesFileIfExists(auto_accepts_file); auto_accepts != nil {
		BinCheckTraces(t, auto_accepts_file, true, true, auto_accepts, schema)
	}
	// Check auto-generated invalid traces (if applicable)
	auto_rejects_file := fmt.Sprintf("%s.%s", test, "auto.rejects")
	if auto_rejects := ReadTracesFileIfExists(auto_rejects_file); auto_rejects != nil {
		BinCheckTraces(t, auto_rejects_file, false, true, auto_rejects, schema)
	}
	//
	if ntests == 0 {
		panic(fmt.Sprintf("missing any tests for %s", test))
	}
}

func BinCheckTraces(t *testing.T, test string, expected bool, expand bool,
	traces [][]trace.RawColumn, srcSchema *hir.Schema) {
	// Run checks using schema compiled from source
	CheckTraces(t, test, MAX_PADDING, expected, expand, traces, srcSchema)
	// Construct binary schema
	if binSchema := encodeDecodeSchema(t, srcSchema); binSchema != nil {
		// Run checks using schema from binary file.  Observe, to try and reduce
		// overhead of repeating all the tests we don't consider padding.
		CheckTraces(t, test, 0, expected, expand, traces, binSchema)
	}
}

// Check a given set of tests have an expected outcome (i.e. are
// either accepted or rejected) by a given set of constraints.
func CheckTraces(t *testing.T, test string, maxPadding uint, expected bool, expand bool,
	traces [][]trace.RawColumn, hirSchema *hir.Schema) {
	for i, tr := range traces {
		if tr != nil {
			// Lower HIR => MIR
			mirSchema := hirSchema.LowerToMir()
			// Lower MIR => AIR
			airSchema := mirSchema.LowerToAir()
			// Align trace with schema, and check whether expanded or not.
			for padding := uint(0); padding <= maxPadding; padding++ {
				// Construct trace identifiers
				hirID := traceId{"HIR", test, expected, i + 1, padding}
				mirID := traceId{"MIR", test, expected, i + 1, padding}
				airID := traceId{"AIR", test, expected, i + 1, padding}
				//
				if expand {
					// Only HIR / MIR constraints for traces which must be
					// expanded.  They don't really make sense otherwise.
					checkTrace(t, tr, expand, hirID, hirSchema)
					checkTrace(t, tr, expand, mirID, mirSchema)
				}
				// Always check AIR constraints
				checkTrace(t, tr, expand, airID, airSchema)
			}
		}
	}
}

func checkTrace(t *testing.T, inputs []trace.RawColumn, expand bool, id traceId, schema sc.Schema) {
	// Construct the trace
	tr, errs := sc.NewTraceBuilder(schema).Expand(expand).Padding(id.padding).Parallel(true).Build(inputs)
	// Sanity check construction
	if len(errs) > 0 {
		for _, err := range errs {
			t.Error(err)
		}
	} else {
		// Check Constraints
		errs := sc.Accepts(100, schema, tr)
		// Check assertions
		errs = append(errs, sc.Asserts(100, schema, tr)...)
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

// A trace identifier uniquely identifies a specific trace within a given test.
// This is used to provide debug information about a trace failure.
// Specifically, so the user knows which line in which file caused the problem.
type traceId struct {
	// Identifies the Intermediate Representation tested against.
	ir string
	// Identifies the test name.  From this, the test filename can be determined
	// in conjunction with the expected outcome.
	test string
	// Identifiers whether this trace should be accepted (true) or rejected
	// (false).
	expected bool
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

func ReadTracesFileIfExists(name string) [][]trace.RawColumn {
	filename := fmt.Sprintf("%s/%s", TestDir, name)
	// Check whether it exists or not
	if _, err := os.Stat(filename); err != nil {
		return nil
	}
	// Yes it does
	return ReadTracesFile(filename)
}

// This is a little test to ensure the binary file format (specifically the
// binary encoder / decoder) works as expected.
func encodeDecodeSchema(t *testing.T, schema *hir.Schema) *hir.Schema {
	var (
		buffer     bytes.Buffer
		gobEncoder *gob.Encoder = gob.NewEncoder(&buffer)
		binSchema  hir.Schema
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
