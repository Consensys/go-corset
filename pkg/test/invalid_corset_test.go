package test

import (
	"fmt"
	"os"
	"testing"

	"github.com/consensys/go-corset/pkg/corset"
	"github.com/consensys/go-corset/pkg/sexp"
)

// Determines the (relative) location of the test directory.  That is
// where the corset test files (lisp) and the corresponding traces
// (accepts/rejects) are found.
const InvalidTestDir = "../../testdata"

// ===================================================================
// Basic Tests
// ===================================================================

func Test_Basic_Invalid_01(t *testing.T) {
	CheckInvalid(t, "basic_invalid_01")
}

func Test_Basic_Invalid_02(t *testing.T) {
	CheckInvalid(t, "basic_invalid_02")
}

func Test_Basic_Invalid_03(t *testing.T) {
	CheckInvalid(t, "basic_invalid_03")
}

func Test_Basic_Invalid_04(t *testing.T) {
	CheckInvalid(t, "basic_invalid_04")
}

func Test_Basic_Invalid_05(t *testing.T) {
	CheckInvalid(t, "basic_invalid_05")
}

func Test_Basic_Invalid_06(t *testing.T) {
	CheckInvalid(t, "basic_invalid_06")
}

func Test_Basic_Invalid_07(t *testing.T) {
	CheckInvalid(t, "basic_invalid_07")
}

func Test_Basic_Invalid_08(t *testing.T) {
	CheckInvalid(t, "basic_invalid_08")
}

func Test_Basic_Invalid_09(t *testing.T) {
	CheckInvalid(t, "basic_invalid_09")
}

func Test_Basic_Invalid_10(t *testing.T) {
	CheckInvalid(t, "basic_invalid_10")
}

func Test_Basic_Invalid_11(t *testing.T) {
	CheckInvalid(t, "basic_invalid_11")
}

func Test_Basic_Invalid_12(t *testing.T) {
	CheckInvalid(t, "basic_invalid_12")
}

// ===================================================================
// Property Tests
// ===================================================================
func Test_Property_Invalid_01(t *testing.T) {
	CheckInvalid(t, "property_invalid_01")
}

func Test_Property_Invalid_02(t *testing.T) {
	CheckInvalid(t, "property_invalid_02")
}

// ===================================================================
// Shift Tests
// ===================================================================

func Test_Shift_Invalid_01(t *testing.T) {
	CheckInvalid(t, "shift_invalid_01")
}

func Test_Shift_Invalid_02(t *testing.T) {
	CheckInvalid(t, "shift_invalid_02")
}

// ===================================================================
// Normalisation Tests
// ===================================================================

func Test_Norm_Invalid_01(t *testing.T) {
	CheckInvalid(t, "norm_invalid_01")
}

// ===================================================================
// If-Zero
// ===================================================================

func Test_If_Invalid_01(t *testing.T) {
	CheckInvalid(t, "if_invalid_01")
}

func Test_If_Invalid_02(t *testing.T) {
	CheckInvalid(t, "if_invalid_02")
}

// ===================================================================
// Range Constraints
// ===================================================================

func Test_Range_Invalid_01(t *testing.T) {
	CheckInvalid(t, "range_invalid_01")
}

func Test_Range_Invalid_02(t *testing.T) {
	CheckInvalid(t, "range_invalid_02")
}

func Test_Range_Invalid_03(t *testing.T) {
	CheckInvalid(t, "range_invalid_03")
}

func Test_Range_Invalid_04(t *testing.T) {
	CheckInvalid(t, "range_invalid_04")
}

// ===================================================================
// Modules
// ===================================================================

func Test_Module_Invalid_01(t *testing.T) {
	CheckInvalid(t, "module_invalid_01")
}

// ===================================================================
// Permutations
// ===================================================================

func Test_Permute_Invalid_01(t *testing.T) {
	CheckInvalid(t, "permute_invalid_01")
}

func Test_Permute_Invalid_02(t *testing.T) {
	CheckInvalid(t, "permute_invalid_02")
}

func Test_Permute_Invalid_03(t *testing.T) {
	CheckInvalid(t, "permute_invalid_03")
}

func Test_Permute_Invalid_04(t *testing.T) {
	CheckInvalid(t, "permute_invalid_04")
}

func Test_Permute_Invalid_05(t *testing.T) {
	CheckInvalid(t, "permute_invalid_05")
}

func Test_Permute_Invalid_06(t *testing.T) {
	CheckInvalid(t, "permute_invalid_06")
}
func Test_Permute_Invalid_07(t *testing.T) {
	CheckInvalid(t, "permute_invalid_07")
}

// ===================================================================
// Test Helpers
// ===================================================================

// Check that a given source file fails to compiler.
func CheckInvalid(t *testing.T, test string) {
	filename := fmt.Sprintf("%s.lisp", test)
	// Enable testing each trace in parallel
	t.Parallel()
	// Read constraints file
	bytes, err := os.ReadFile(fmt.Sprintf("%s/%s", InvalidTestDir, filename))
	// Check test file read ok
	if err != nil {
		t.Fatal(err)
	}
	// Package up as source file
	srcfile := sexp.NewSourceFile(filename, bytes)
	// Parse terms into an HIR schema
	_, errs := corset.CompileSourceFile(srcfile)
	// Check program did not compile!
	if len(errs) == 0 {
		t.Fatalf("Error %s should not have compiled\n", filename)
	}
}
