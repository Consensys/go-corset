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
	"fmt"
	"os"
	"strconv"
	"strings"
	"testing"

	"github.com/consensys/go-corset/pkg/corset"
	"github.com/consensys/go-corset/pkg/util/sexp"
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

func Test_Invalid_Basic_13(t *testing.T) {
	CheckInvalid(t, "basic_invalid_13")
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

func Test_Invalid_If_03(t *testing.T) {
	CheckInvalid(t, "if_invalid_03")
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

func Test_Invalid_Type_09(t *testing.T) {
	CheckInvalid(t, "type_invalid_09")
}

func Test_Invalid_Type_10(t *testing.T) {
	CheckInvalid(t, "type_invalid_10")
}

func Test_Invalid_Type_11(t *testing.T) {
	CheckInvalid(t, "type_invalid_11")
}

func Test_Invalid_Type_12(t *testing.T) {
	CheckInvalid(t, "type_invalid_12")
}

func Test_Invalid_Type_13(t *testing.T) {
	CheckInvalid(t, "type_invalid_13")
}

func Test_Invalid_Type_14(t *testing.T) {
	CheckInvalid(t, "type_invalid_14")
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

/* func Test_Invalid_Permute_05(t *testing.T) {
	CheckInvalid(t, "permute_invalid_05")
} */

func Test_Invalid_Permute_06(t *testing.T) {
	CheckInvalid(t, "permute_invalid_06")
}
func Test_Invalid_Permute_07(t *testing.T) {
	CheckInvalid(t, "permute_invalid_07")
}

func Test_Invalid_Permute_08(t *testing.T) {
	CheckInvalid(t, "permute_invalid_08")
}
func Test_Invalid_Permute_09(t *testing.T) {
	CheckInvalid(t, "permute_invalid_09")
}
func Test_Invalid_Permute_10(t *testing.T) {
	CheckInvalid(t, "permute_invalid_10")
}

// ===================================================================
// Sortings
// ===================================================================

func Test_Invalid_Sorted_01(t *testing.T) {
	CheckInvalid(t, "sorted_invalid_01")
}

func Test_Invalid_Sorted_02(t *testing.T) {
	CheckInvalid(t, "sorted_invalid_02")
}

func Test_Invalid_Sorted_03(t *testing.T) {
	CheckInvalid(t, "sorted_invalid_03")
}
func Test_Invalid_Sorted_04(t *testing.T) {
	CheckInvalid(t, "sorted_invalid_04")
}
func Test_Invalid_Sorted_05(t *testing.T) {
	CheckInvalid(t, "sorted_invalid_05")
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

func Test_Invalid_PureFun_14(t *testing.T) {
	CheckInvalid(t, "purefun_invalid_14")
}
func Test_Invalid_PureFun_15(t *testing.T) {
	CheckInvalid(t, "purefun_invalid_15")
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

func Test_Invalid_Array_06(t *testing.T) {
	CheckInvalid(t, "array_invalid_06")
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
// Perspectives
// ===================================================================
func Test_Invalid_Perspective_01(t *testing.T) {
	CheckInvalid(t, "perspective_invalid_01")
}

func Test_Invalid_Perspective_02(t *testing.T) {
	CheckInvalid(t, "perspective_invalid_02")
}

func Test_Invalid_Perspective_03(t *testing.T) {
	CheckInvalid(t, "perspective_invalid_03")
}

func Test_Invalid_Perspective_05(t *testing.T) {
	CheckInvalid(t, "perspective_invalid_05")
}

func Test_Invalid_Perspective_06(t *testing.T) {
	CheckInvalid(t, "perspective_invalid_06")
}

func Test_Invalid_Perspective_08(t *testing.T) {
	CheckInvalid(t, "perspective_invalid_08")
}

// ===================================================================
// Perspectives
// ===================================================================
func Test_Invalid_Let_01(t *testing.T) {
	CheckInvalid(t, "let_invalid_01")
}

func Test_Invalid_Let_02(t *testing.T) {
	CheckInvalid(t, "let_invalid_02")
}
func Test_Invalid_Let_03(t *testing.T) {
	CheckInvalid(t, "let_invalid_03")
}
func Test_Invalid_Let_04(t *testing.T) {
	CheckInvalid(t, "let_invalid_04")
}
func Test_Invalid_Let_05(t *testing.T) {
	CheckInvalid(t, "let_invalid_05")
}
func Test_Invalid_Let_06(t *testing.T) {
	CheckInvalid(t, "let_invalid_06")
}

func Test_Invalid_Let_07(t *testing.T) {
	CheckInvalid(t, "let_invalid_07")
}

// ===================================================================
// Computed Columns
// ===================================================================

func Test_Invalid_Compute_01(t *testing.T) {
	CheckInvalid(t, "compute_invalid_01")
}

func Test_Invalid_Compute_02(t *testing.T) {
	CheckInvalid(t, "compute_invalid_02")
}

func Test_Invalid_Compute_03(t *testing.T) {
	CheckInvalid(t, "compute_invalid_03")
}

func Test_Invalid_Compute_04(t *testing.T) {
	CheckInvalid(t, "compute_invalid_04")
}

func Test_Invalid_Compute_05(t *testing.T) {
	CheckInvalid(t, "compute_invalid_05")
}

func Test_Invalid_Compute_06(t *testing.T) {
	CheckInvalid(t, "compute_invalid_06")
}

func Test_Invalid_Compute_07(t *testing.T) {
	CheckInvalid(t, "compute_invalid_07")
}

// ===================================================================
// Test Helpers
// ===================================================================

// Check that a given source file fails to compiler.
// nolint
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
	expectedErrs, lineOffsets := extractExpectedErrors(bytes)
	// Check program did not compile!
	if len(errs) == 0 {
		t.Fatalf("Error %s should not have compiled\n", filename)
	} else {
		error := false
		// Construct initial message
		msg := fmt.Sprintf("Error %s\n", filename)
		// Pad out with what received
		for i := 0; i < max(len(errs), len(expectedErrs)); i++ {
			if i < len(errs) && i < len(expectedErrs) {
				expected := expectedErrs[i]
				actual := errs[i]
				// Check whether message OK
				if expected.msg == actual.Message() && expected.span == actual.Span() {
					continue
				}
			}
			// Indicate error arose
			error = true
			// actual
			if i < len(errs) {
				actual := errs[i]
				msg = fmt.Sprintf("%s unexpected error %s:%s\n", msg, spanToString(actual.Span(), lineOffsets), actual.Message())
			}
			// expected
			if i < len(expectedErrs) {
				expected := expectedErrs[i]
				msg = fmt.Sprintf("%s   expected error %s:%s\n", msg, spanToString(expected.span, lineOffsets), expected.msg)
			}
		}
		//
		if error {
			t.Fatalf(msg)
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

func extractExpectedErrors(bytes []byte) ([]SyntaxError, []int) {
	// Calcuate the character offset of each line
	offsets, lines := splitFileLines(bytes)
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

// Split out a given file into the line contents and the line offsets.  This
// needs to be done carefully to ensure that these both align properly,
// otherwise error messages tend to have the wrong column numbers, etc.
func splitFileLines(bytes []byte) ([]int, []string) {
	contents := []rune(string(bytes))
	// Calcuate the character offset of each line
	offsets := make([]int, 1)
	lines := make([]string, 0)
	start := 0
	// Iterate each byte
	for i := 0; i <= len(contents); i++ {
		if i == len(contents) || contents[i] == '\n' {
			line := string(contents[start:i])
			offsets = append(offsets, i+1)
			lines = append(lines, line)
			//
			start = i + 1
		}
	}
	// Done
	return offsets, lines
}

// Extract the syntax error from a given line in the source file, or return nil
// if it does not describe an error.
func extractSyntaxError(line string, offsets []int) *SyntaxError {
	if strings.HasPrefix(line, ";;error") {
		splits := strings.Split(line, ":")
		span := determineFileSpan(splits[1], splits[2], offsets)
		msg := strings.Join(splits[3:], ":")
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
	} else if start == 0 {
		panic("columns numbered from 1")
	}
	// Parse span end as integer
	end, err := strconv.Atoi(span_splits[1])
	if err != nil {
		panic(err)
	}
	// Add line offset
	start += offsets[line-1]
	end += offsets[line-1]
	// Sanity check
	if start >= offsets[line] || end > offsets[line] {
		panic("span overflows to following line")
	}
	// Create span, recalling that span's start from zero whereas column numbers
	// start from 1.
	return sexp.NewSpan(start-1, end-1)
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
	return fmt.Sprintf("%d:%d-%d", line, 1+start-last, 1+end-last)
}
