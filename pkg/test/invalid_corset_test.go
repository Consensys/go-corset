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
