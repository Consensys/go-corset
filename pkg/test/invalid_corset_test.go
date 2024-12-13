package test

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"testing"

	"github.com/consensys/go-corset/pkg/corset"
	"github.com/consensys/go-corset/pkg/sexp"
	"github.com/consensys/go-corset/pkg/util"
)

// Determines the (relative) location of the test directory.  That is
// where the corset test files (lisp) and the corresponding traces
// (accepts/rejects) are found.
const InvalidTestDir = "../../testdata"

// ===================================================================
// Basic Tests
// ===================================================================

func Test_Invalid_Basic_01(t *testing.T) {
	CheckInvalid(t, "basic_invalid_01")
}

func Test_Invalid_Basic_02(t *testing.T) {
	CheckInvalid(t, "basic_invalid_02")
}

func Test_Invalid_Basic_03(t *testing.T) {
	CheckInvalid(t, "basic_invalid_03")
}

func Test_Invalid_Basic_04(t *testing.T) {
	CheckInvalid(t, "basic_invalid_04")
}

func Test_Invalid_Basic_05(t *testing.T) {
	CheckInvalid(t, "basic_invalid_05")
}

func Test_Invalid_Basic_06(t *testing.T) {
	CheckInvalid(t, "basic_invalid_06")
}

func Test_Invalid_Basic_07(t *testing.T) {
	CheckInvalid(t, "basic_invalid_07")
}

func Test_Invalid_Basic_08(t *testing.T) {
	CheckInvalid(t, "basic_invalid_08")
}

func Test_Invalid_Basic_09(t *testing.T) {
	CheckInvalid(t, "basic_invalid_09")
}

func Test_Invalid_Basic_10(t *testing.T) {
	CheckInvalid(t, "basic_invalid_10")
}

func Test_Invalid_Basic_11(t *testing.T) {
	CheckInvalid(t, "basic_invalid_11")
}

func Test_Invalid_Basic_12(t *testing.T) {
	CheckInvalid(t, "basic_invalid_12")
}

// ===================================================================
// Constant Tests
// ===================================================================
func Test_Invalid_Constant_01(t *testing.T) {
	CheckInvalid(t, "constant_invalid_01")
}

func Test_Invalid_Constant_02(t *testing.T) {
	CheckInvalid(t, "constant_invalid_02")
}

func Test_Invalid_Constant_03(t *testing.T) {
	CheckInvalid(t, "constant_invalid_03")
}

func Test_Invalid_Constant_04(t *testing.T) {
	CheckInvalid(t, "constant_invalid_04")
}

func Test_Invalid_Constant_05(t *testing.T) {
	CheckInvalid(t, "constant_invalid_05")
}

/* Recursive --- #406
  func Test_Invalid_Constant_06(t *testing.T) {
	CheckInvalid(t, "constant_invalid_06")
} */

/* Recursive --- #406
  func Test_Invalid_Constant_07(t *testing.T) {
	CheckInvalid(t, "constant_invalid_07")
}
*/
/* Recursive --- #406
  func Test_Invalid_Constant_08(t *testing.T) {
	CheckInvalid(t, "constant_invalid_08")
} */

func Test_Invalid_Constant_09(t *testing.T) {
	CheckInvalid(t, "constant_invalid_09")
}

func Test_Invalid_Constant_10(t *testing.T) {
	CheckInvalid(t, "constant_invalid_10")
}

func Test_Invalid_Constant_11(t *testing.T) {
	CheckInvalid(t, "constant_invalid_11")
}

func Test_Invalid_Constant_12(t *testing.T) {
	CheckInvalid(t, "constant_invalid_12")
}

func Test_Invalid_Constant_13(t *testing.T) {
	CheckInvalid(t, "constant_invalid_13")
}

func Test_Invalid_Constant_14(t *testing.T) {
	CheckInvalid(t, "constant_invalid_14")
}

func Test_Invalid_Constant_15(t *testing.T) {
	CheckInvalid(t, "constant_invalid_15")
}

func Test_Invalid_Constant_16(t *testing.T) {
	CheckInvalid(t, "constant_invalid_16")
}

func Test_Invalid_Constant_17(t *testing.T) {
	CheckInvalid(t, "constant_invalid_17")
}

// ===================================================================
// Alias Tests
// ===================================================================
func Test_Invalid_Alias_01(t *testing.T) {
	CheckInvalid(t, "alias_invalid_01")
}

func Test_Invalid_Alias_02(t *testing.T) {
	CheckInvalid(t, "alias_invalid_02")
}

func Test_Invalid_Alias_03(t *testing.T) {
	CheckInvalid(t, "alias_invalid_03")
}

func Test_Invalid_Alias_04(t *testing.T) {
	CheckInvalid(t, "alias_invalid_04")
}

func Test_Invalid_Alias_05(t *testing.T) {
	CheckInvalid(t, "alias_invalid_05")
}

func Test_Invalid_Alias_06(t *testing.T) {
	CheckInvalid(t, "alias_invalid_06")
}

func Test_Invalid_Alias_07(t *testing.T) {
	CheckInvalid(t, "alias_invalid_07")
}

// ===================================================================
// Function Alias Tests
// ===================================================================
func Test_Invalid_FunAlias_01(t *testing.T) {
	CheckInvalid(t, "funalias_invalid_01")
}

func Test_Invalid_FunAlias_02(t *testing.T) {
	CheckInvalid(t, "funalias_invalid_02")
}

func Test_Invalid_FunAlias_03(t *testing.T) {
	CheckInvalid(t, "funalias_invalid_03")
}

func Test_Invalid_FunAlias_04(t *testing.T) {
	CheckInvalid(t, "funalias_invalid_04")
}

func Test_Invalid_FunAlias_05(t *testing.T) {
	CheckInvalid(t, "funalias_invalid_05")
}

// ===================================================================
// Property Tests
// ===================================================================
func Test_Invalid_Property_01(t *testing.T) {
	CheckInvalid(t, "property_invalid_01")
}

func Test_Invalid_Property_02(t *testing.T) {
	CheckInvalid(t, "property_invalid_02")
}

// ===================================================================
// Shift Tests
// ===================================================================

func Test_Invalid_Shift_01(t *testing.T) {
	CheckInvalid(t, "shift_invalid_01")
}

func Test_Invalid_Shift_02(t *testing.T) {
	CheckInvalid(t, "shift_invalid_02")
}

// ===================================================================
// Normalisation Tests
// ===================================================================

func Test_Invalid_Norm_01(t *testing.T) {
	CheckInvalid(t, "norm_invalid_01")
}

// ===================================================================
// If-Zero
// ===================================================================

func Test_Invalid_If_01(t *testing.T) {
	CheckInvalid(t, "if_invalid_01")
}

func Test_Invalid_If_02(t *testing.T) {
	CheckInvalid(t, "if_invalid_02")
}

// ===================================================================
// Types
// ===================================================================

func Test_Invalid_Type_01(t *testing.T) {
	CheckInvalid(t, "type_invalid_01")
}

func Test_Invalid_Type_02(t *testing.T) {
	CheckInvalid(t, "type_invalid_02")
}

func Test_Invalid_Type_03(t *testing.T) {
	CheckInvalid(t, "type_invalid_03")
}

func Test_Invalid_Type_04(t *testing.T) {
	CheckInvalid(t, "type_invalid_04")
}

func Test_Invalid_Type_05(t *testing.T) {
	CheckInvalid(t, "type_invalid_05")
}

func Test_Invalid_Type_06(t *testing.T) {
	CheckInvalid(t, "type_invalid_06")
}

func Test_Invalid_Type_07(t *testing.T) {
	CheckInvalid(t, "type_invalid_07")
}

func Test_Invalid_Type_08(t *testing.T) {
	CheckInvalid(t, "type_invalid_08")
}

// ===================================================================
// Range Constraints
// ===================================================================

func Test_Invalid_Range_01(t *testing.T) {
	CheckInvalid(t, "range_invalid_01")
}

func Test_Invalid_Range_02(t *testing.T) {
	CheckInvalid(t, "range_invalid_02")
}

func Test_Invalid_Range_03(t *testing.T) {
	CheckInvalid(t, "range_invalid_03")
}

func Test_Invalid_Range_04(t *testing.T) {
	CheckInvalid(t, "range_invalid_04")
}

// ===================================================================
// Modules
// ===================================================================

func Test_Invalid_Module_01(t *testing.T) {
	CheckInvalid(t, "module_invalid_01")
}

// ===================================================================
// Permutations
// ===================================================================

func Test_Invalid_Permute_01(t *testing.T) {
	CheckInvalid(t, "permute_invalid_01")
}

func Test_Invalid_Permute_02(t *testing.T) {
	CheckInvalid(t, "permute_invalid_02")
}

func Test_Invalid_Permute_03(t *testing.T) {
	CheckInvalid(t, "permute_invalid_03")
}

func Test_Invalid_Permute_04(t *testing.T) {
	CheckInvalid(t, "permute_invalid_04")
}

func Test_Invalid_Permute_05(t *testing.T) {
	CheckInvalid(t, "permute_invalid_05")
}

func Test_Invalid_Permute_06(t *testing.T) {
	CheckInvalid(t, "permute_invalid_06")
}
func Test_Invalid_Permute_07(t *testing.T) {
	CheckInvalid(t, "permute_invalid_07")
}

func Test_Invalid_Permute_08(t *testing.T) {
	CheckInvalid(t, "permute_invalid_08")
}

// ===================================================================
// Lookups
// ===================================================================

func Test_Invalid_Lookup_01(t *testing.T) {
	CheckInvalid(t, "lookup_invalid_01")
}

func Test_Invalid_Lookup_02(t *testing.T) {
	CheckInvalid(t, "lookup_invalid_02")
}
func Test_Invalid_Lookup_03(t *testing.T) {
	CheckInvalid(t, "lookup_invalid_03")
}

func Test_Invalid_Lookup_04(t *testing.T) {
	CheckInvalid(t, "lookup_invalid_04")
}

func Test_Invalid_Lookup_05(t *testing.T) {
	CheckInvalid(t, "lookup_invalid_05")
}
func Test_Invalid_Lookup_06(t *testing.T) {
	CheckInvalid(t, "lookup_invalid_06")
}
func Test_Invalid_Lookup_07(t *testing.T) {
	CheckInvalid(t, "lookup_invalid_07")
}
func Test_Invalid_Lookup_08(t *testing.T) {
	CheckInvalid(t, "lookup_invalid_08")
}
func Test_Invalid_Lookup_09(t *testing.T) {
	CheckInvalid(t, "lookup_invalid_09")
}

// ===================================================================
// Interleavings
// ===================================================================

func Test_Invalid_Interleave_01(t *testing.T) {
	CheckInvalid(t, "interleave_invalid_01")
}

func Test_Invalid_Interleave_02(t *testing.T) {
	CheckInvalid(t, "interleave_invalid_02")
}

func Test_Invalid_Interleave_03(t *testing.T) {
	CheckInvalid(t, "interleave_invalid_03")
}

func Test_Invalid_Interleave_04(t *testing.T) {
	CheckInvalid(t, "interleave_invalid_04")
}

func Test_Invalid_Interleave_05(t *testing.T) {
	CheckInvalid(t, "interleave_invalid_05")
}

func Test_Invalid_Interleave_06(t *testing.T) {
	CheckInvalid(t, "interleave_invalid_06")
}

func Test_Invalid_Interleave_07(t *testing.T) {
	CheckInvalid(t, "interleave_invalid_07")
}

func Test_Invalid_Interleave_08(t *testing.T) {
	CheckInvalid(t, "interleave_invalid_08")
}

func Test_Invalid_Interleave_09(t *testing.T) {
	CheckInvalid(t, "interleave_invalid_09")
}

func Test_Invalid_Interleave_10(t *testing.T) {
	CheckInvalid(t, "interleave_invalid_10")
}

func Test_Invalid_Interleave_11(t *testing.T) {
	CheckInvalid(t, "interleave_invalid_11")
}

func Test_Invalid_Interleave_12(t *testing.T) {
	CheckInvalid(t, "interleave_invalid_12")
}

// ===================================================================
// Functions
// ===================================================================

func Test_Invalid_Fun_01(t *testing.T) {
	CheckInvalid(t, "fun_invalid_01")
}

func Test_Invalid_Fun_02(t *testing.T) {
	CheckInvalid(t, "fun_invalid_02")
}

func Test_Invalid_Fun_03(t *testing.T) {
	CheckInvalid(t, "fun_invalid_03")
}

/*
func Test_Invalid_Fun_04(t *testing.T) {
	CheckInvalid(t, "fun_invalid_04")
} */

// ===================================================================
// Pure Functions
// ===================================================================

func Test_Invalid_PureFun_01(t *testing.T) {
	CheckInvalid(t, "purefun_invalid_01")
}

func Test_Invalid_PureFun_02(t *testing.T) {
	CheckInvalid(t, "purefun_invalid_02")
}

func Test_Invalid_PureFun_03(t *testing.T) {
	CheckInvalid(t, "purefun_invalid_03")
}

func Test_Invalid_PureFun_04(t *testing.T) {
	CheckInvalid(t, "purefun_invalid_04")
}

func Test_Invalid_PureFun_05(t *testing.T) {
	CheckInvalid(t, "purefun_invalid_05")
}

/*
	func Test_Invalid_PureFun_06(t *testing.T) {
		CheckInvalid(t, "purefun_invalid_06")
	}
*/

func Test_Invalid_PureFun_07(t *testing.T) {
	CheckInvalid(t, "purefun_invalid_07")
}

func Test_Invalid_PureFun_08(t *testing.T) {
	// tricky one
	CheckInvalid(t, "purefun_invalid_08")
}

func Test_Invalid_PureFun_09(t *testing.T) {
	// tricky one
	CheckInvalid(t, "purefun_invalid_09")
}

func Test_Invalid_PureFun_10(t *testing.T) {
	CheckInvalid(t, "purefun_invalid_10")
}

func Test_Invalid_PureFun_11(t *testing.T) {
	CheckInvalid(t, "purefun_invalid_11")
}

func Test_Invalid_PureFun_12(t *testing.T) {
	CheckInvalid(t, "purefun_invalid_12")
}

func Test_Invalid_PureFun_13(t *testing.T) {
	CheckInvalid(t, "purefun_invalid_13")
}

// ===================================================================
// For Loops
// ===================================================================
func Test_Invalid_For_01(t *testing.T) {
	CheckInvalid(t, "for_invalid_01")
}

func Test_Invalid_For_02(t *testing.T) {
	CheckInvalid(t, "for_invalid_02")
}

func Test_Invalid_For_03(t *testing.T) {
	CheckInvalid(t, "for_invalid_03")
}

// ===================================================================
// Arrays
// ===================================================================
func Test_Invalid_Array_01(t *testing.T) {
	CheckInvalid(t, "array_invalid_01")
}

func Test_Invalid_Array_02(t *testing.T) {
	CheckInvalid(t, "array_invalid_02")
}

func Test_Invalid_Array_03(t *testing.T) {
	CheckInvalid(t, "array_invalid_03")
}

func Test_Invalid_Array_04(t *testing.T) {
	CheckInvalid(t, "array_invalid_04")
}

func Test_Invalid_Array_05(t *testing.T) {
	CheckInvalid(t, "array_invalid_05")
}

// ===================================================================
// Reduce
// ===================================================================

func Test_Invalid_Reduce_01(t *testing.T) {
	CheckInvalid(t, "reduce_invalid_01")
}

func Test_Invalid_Reduce_02(t *testing.T) {
	CheckInvalid(t, "reduce_invalid_02")
}

func Test_Invalid_Reduce_03(t *testing.T) {
	CheckInvalid(t, "reduce_invalid_03")
}

func Test_Invalid_Reduce_04(t *testing.T) {
	CheckInvalid(t, "reduce_invalid_04")
}

func Test_Invalid_Reduce_05(t *testing.T) {
	CheckInvalid(t, "reduce_invalid_05")
}

// ===================================================================
// Debug
// ===================================================================

func Test_Invalid_Debug_01(t *testing.T) {
	CheckInvalid(t, "debug_invalid_01")
}

func Test_Invalid_Debug_02(t *testing.T) {
	CheckInvalid(t, "debug_invalid_02")
}

// ===================================================================
// Test Helpers
// ===================================================================

// Check that a given source file fails to compiler.
func CheckInvalid(t *testing.T, test string) {
	filename := fmt.Sprintf("%s/%s.lisp", InvalidTestDir, test)
	// Enable testing each trace in parallel
	t.Parallel()
	// Read constraints file
	bytes, err := os.ReadFile(filename)
	// Check test file read ok
	if err != nil {
		t.Fatal(err)
	}
	// Package up as source file
	srcfile := sexp.NewSourceFile(filename, bytes)
	// Parse terms into an HIR schema
	_, errs := corset.CompileSourceFile(false, false, srcfile)
	// Extract expected errors for comparison
	expectedErrs, lineOffsets := extractExpectedErrors(filename)
	// Check program did not compile!
	if len(errs) == 0 {
		t.Fatalf("Error %s should not have compiled\n", filename)
	} else if len(errs) != len(expectedErrs) {
		t.Fatalf("Error %s reported incorrect number of errors (%d vs %d)\n", filename, len(errs), len(expectedErrs))
	} else {
		// Check errors match
		for i := 0; i < len(expectedErrs); i++ {
			expected := expectedErrs[i]
			actual := errs[i]
			//
			if expected.msg != actual.Message() {
				t.Fatalf("Error %s, got \"%s\" but wanted \"%s\"\n", filename, actual.Message(), expected.msg)
			} else if expected.span != actual.Span() {
				aSpan := spanToString(actual.Span(), lineOffsets)
				eSpan := spanToString(expected.span, lineOffsets)
				t.Fatalf("Error %s, span was %s but wanted %s\n", filename, aSpan, eSpan)
			}
		}
	}
}

// SyntaxError captures key information about an expected error
type SyntaxError struct {
	// The range of bytes in the original file to which this error is
	// associated.
	span sexp.Span
	// The error message reported.
	msg string
}

func extractExpectedErrors(filename string) ([]SyntaxError, []int) {
	// Read the source file and convert into one or more lines.
	lines := util.ReadInputFile(filename)
	// Calcuate the character offset of each line
	offsets := make([]int, len(lines))
	offset := 0
	//
	for i, line := range lines {
		offsets[i] = offset
		offset += len(line) + 1
	}
	// Now construct errors
	errors := make([]SyntaxError, 0)
	// scan file line-by-line until no more errors found
	for _, line := range lines {
		error := extractSyntaxError(line, offsets)
		// Keep going until no more errors
		if error == nil {
			return errors, offsets
		}

		errors = append(errors, *error)
	}
	//
	return errors, offsets
}

// Extract the syntax error from a given line in the source file, or return nil
// if it does not describe an error.
func extractSyntaxError(line string, offsets []int) *SyntaxError {
	if strings.HasPrefix(line, ";;error") {
		splits := strings.Split(line, ":")
		span := determineFileSpan(splits[1], splits[2], offsets)
		msg := splits[3]
		// Done
		return &SyntaxError{span, msg}
	}
	// No error
	return nil
}

// Determine the span that the the given line string and span string corresponds
// to.  We need the line offsets so that the computed span includes the starting
// offset of the relevant line.
func determineFileSpan(line_str string, span_str string, offsets []int) sexp.Span {
	line, err := strconv.Atoi(line_str)
	if err != nil {
		panic(err)
	}
	// Split the span
	span_splits := strings.Split(span_str, "-")
	// Parse span start as integer
	start, err := strconv.Atoi(span_splits[0])
	if err != nil {
		panic(err)
	}
	// Parse span end as integer
	end, err := strconv.Atoi(span_splits[1])
	if err != nil {
		panic(err)
	}
	// Add line offset
	start += offsets[line-1]
	end += offsets[line-1]
	// Done
	return sexp.NewSpan(start, end)
}

// Convert a span into a useful human readable string.
func spanToString(span sexp.Span, offsets []int) string {
	line := 0
	last := 0
	start := span.Start()
	end := span.End()
	//
	for i, o := range offsets {
		if o > start {
			break
		}
		// Update status
		last = o
		line = i + 1
	}
	//
	return fmt.Sprintf("%d:%d-%d", line, start-last, end-last)
}
