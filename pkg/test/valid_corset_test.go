package test

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
	"testing"

	"github.com/consensys/go-corset/pkg/corset"
	"github.com/consensys/go-corset/pkg/hir"
	sc "github.com/consensys/go-corset/pkg/schema"
	"github.com/consensys/go-corset/pkg/sexp"
	"github.com/consensys/go-corset/pkg/trace"
	"github.com/consensys/go-corset/pkg/trace/json"
)

// Determines the (relative) location of the test directory.  That is
// where the corset test files (lisp) and the corresponding traces
// (accepts/rejects) are found.
const TestDir = "../../testdata"

// ===================================================================
// Basic Tests
// ===================================================================

func Test_Basic_01(t *testing.T) {
	Check(t, "basic_01")
}

func Test_Basic_02(t *testing.T) {
	Check(t, "basic_02")
}

func Test_Basic_03(t *testing.T) {
	Check(t, "basic_03")
}

func Test_Basic_04(t *testing.T) {
	Check(t, "basic_04")
}

func Test_Basic_05(t *testing.T) {
	Check(t, "basic_05")
}

func Test_Basic_06(t *testing.T) {
	Check(t, "basic_06")
}

func Test_Basic_07(t *testing.T) {
	Check(t, "basic_07")
}

func Test_Basic_08(t *testing.T) {
	Check(t, "basic_08")
}

func Test_Basic_09(t *testing.T) {
	Check(t, "basic_09")
}

func Test_Basic_10(t *testing.T) {
	Check(t, "basic_10")
}

// ===================================================================
// Constants Tests
// ===================================================================
func Test_Constant_01(t *testing.T) {
	Check(t, "constant_01")
}

func Test_Constant_02(t *testing.T) {
	Check(t, "constant_02")
}

func Test_Constant_03(t *testing.T) {
	Check(t, "constant_03")
}

func Test_Constant_04(t *testing.T) {
	Check(t, "constant_04")
}

func Test_Constant_05(t *testing.T) {
	Check(t, "constant_05")
}

func Test_Constant_06(t *testing.T) {
	Check(t, "constant_06")
}

func Test_Constant_07(t *testing.T) {
	Check(t, "constant_07")
}

// ===================================================================
// Alias Tests
// ===================================================================
func Test_Alias_01(t *testing.T) {
	Check(t, "alias_01")
}
func Test_Alias_02(t *testing.T) {
	Check(t, "alias_02")
}
func Test_Alias_03(t *testing.T) {
	Check(t, "alias_03")
}
func Test_Alias_04(t *testing.T) {
	Check(t, "alias_04")
}
func Test_Alias_05(t *testing.T) {
	Check(t, "alias_05")
}
func Test_Alias_06(t *testing.T) {
	Check(t, "alias_06")
}

// ===================================================================
// Function Alias Tests
// ===================================================================
func Test_FunAlias_01(t *testing.T) {
	Check(t, "funalias_01")
}

func Test_FunAlias_02(t *testing.T) {
	Check(t, "funalias_02")
}

func Test_FunAlias_03(t *testing.T) {
	Check(t, "funalias_03")
}

// ===================================================================
// Domain Tests
// ===================================================================

func Test_Domain_01(t *testing.T) {
	Check(t, "domain_01")
}

func Test_Domain_02(t *testing.T) {
	Check(t, "domain_02")
}

func Test_Domain_03(t *testing.T) {
	Check(t, "domain_03")
}

// ===================================================================
// Block Tests
// ===================================================================

func Test_Block_01(t *testing.T) {
	Check(t, "block_01")
}

func Test_Block_02(t *testing.T) {
	Check(t, "block_02")
}

func Test_Block_03(t *testing.T) {
	Check(t, "block_03")
}

// ===================================================================
// Property Tests
// ===================================================================

func Test_Property_01(t *testing.T) {
	Check(t, "property_01")
}

// ===================================================================
// Shift Tests
// ===================================================================

func Test_Shift_01(t *testing.T) {
	Check(t, "shift_01")
}

func Test_Shift_02(t *testing.T) {
	Check(t, "shift_02")
}

func Test_Shift_03(t *testing.T) {
	Check(t, "shift_03")
}

func Test_Shift_04(t *testing.T) {
	Check(t, "shift_04")
}

func Test_Shift_05(t *testing.T) {
	Check(t, "shift_05")
}

func Test_Shift_06(t *testing.T) {
	Check(t, "shift_06")
}

func Test_Shift_07(t *testing.T) {
	Check(t, "shift_07")
}

func Test_Shift_08(t *testing.T) {
	Check(t, "shift_08")
}

// ===================================================================
// Spillage Tests
// ===================================================================

func Test_Spillage_01(t *testing.T) {
	Check(t, "spillage_01")
}

func Test_Spillage_02(t *testing.T) {
	Check(t, "spillage_02")
}

func Test_Spillage_03(t *testing.T) {
	Check(t, "spillage_03")
}

func Test_Spillage_04(t *testing.T) {
	Check(t, "spillage_04")
}

func Test_Spillage_05(t *testing.T) {
	Check(t, "spillage_05")
}

func Test_Spillage_06(t *testing.T) {
	Check(t, "spillage_06")
}

func Test_Spillage_07(t *testing.T) {
	Check(t, "spillage_07")
}

func Test_Spillage_08(t *testing.T) {
	Check(t, "spillage_08")
}

func Test_Spillage_09(t *testing.T) {
	Check(t, "spillage_09")
}

// ===================================================================
// Normalisation Tests
// ===================================================================

func Test_Norm_01(t *testing.T) {
	Check(t, "norm_01")
}

func Test_Norm_02(t *testing.T) {
	Check(t, "norm_02")
}

func Test_Norm_03(t *testing.T) {
	Check(t, "norm_03")
}

func Test_Norm_04(t *testing.T) {
	Check(t, "norm_04")
}

func Test_Norm_05(t *testing.T) {
	Check(t, "norm_05")
}

func Test_Norm_06(t *testing.T) {
	Check(t, "norm_06")
}

func Test_Norm_07(t *testing.T) {
	Check(t, "norm_07")
}

// ===================================================================
// If-Zero
// ===================================================================

func Test_If_01(t *testing.T) {
	Check(t, "if_01")
}

func Test_If_02(t *testing.T) {
	Check(t, "if_02")
}

func Test_If_03(t *testing.T) {
	Check(t, "if_03")
}

func Test_If_04(t *testing.T) {
	Check(t, "if_04")
}

func Test_If_05(t *testing.T) {
	Check(t, "if_05")
}

func Test_If_06(t *testing.T) {
	Check(t, "if_06")
}

func Test_If_07(t *testing.T) {
	Check(t, "if_07")
}

func Test_If_08(t *testing.T) {
	Check(t, "if_08")
}

func Test_If_09(t *testing.T) {
	Check(t, "if_09")
}

// ===================================================================
// Guards
// ===================================================================

func Test_Guard_01(t *testing.T) {
	Check(t, "guard_01")
}

func Test_Guard_02(t *testing.T) {
	Check(t, "guard_02")
}

func Test_Guard_03(t *testing.T) {
	Check(t, "guard_03")
}

func Test_Guard_04(t *testing.T) {
	Check(t, "guard_04")
}

func Test_Guard_05(t *testing.T) {
	Check(t, "guard_05")
}

// ===================================================================
// Types
// ===================================================================

func Test_Type_01(t *testing.T) {
	Check(t, "type_01")
}

func Test_Type_02(t *testing.T) {
	Check(t, "type_02")
}

func Test_Type_03(t *testing.T) {
	Check(t, "type_03")
}

func Test_Type_04(t *testing.T) {
	Check(t, "type_04")
}

func Test_Type_05(t *testing.T) {
	Check(t, "type_04")
}

func Test_Type_06(t *testing.T) {
	Check(t, "type_06")
}

func Test_Type_07(t *testing.T) {
	Check(t, "type_07")
}

func Test_Type_08(t *testing.T) {
	Check(t, "type_08")
}

// ===================================================================
// Range Constraints
// ===================================================================

func Test_Range_01(t *testing.T) {
	Check(t, "range_01")
}

func Test_Range_02(t *testing.T) {
	Check(t, "range_02")
}

func Test_Range_03(t *testing.T) {
	Check(t, "range_03")
}

func Test_Range_04(t *testing.T) {
	Check(t, "range_04")
}

func Test_Range_05(t *testing.T) {
	Check(t, "range_05")
}

// ===================================================================
// Constant Propagation
// ===================================================================

func Test_ConstExpr_01(t *testing.T) {
	Check(t, "constexpr_01")
}

func Test_ConstExpr_02(t *testing.T) {
	Check(t, "constexpr_02")
}

func Test_ConstExpr_03(t *testing.T) {
	Check(t, "constexpr_03")
}

func Test_ConstExpr_04(t *testing.T) {
	Check(t, "constexpr_04")
}

func Test_ConstExpr_05(t *testing.T) {
	Check(t, "constexpr_05")
}

// ===================================================================
// Modules
// ===================================================================

func Test_Module_01(t *testing.T) {
	Check(t, "module_01")
}

func Test_Module_02(t *testing.T) {
	Check(t, "module_02")
}

func Test_Module_03(t *testing.T) {
	Check(t, "module_03")
}

func Test_Module_04(t *testing.T) {
	Check(t, "module_04")
}

func Test_Module_05(t *testing.T) {
	Check(t, "module_05")
}

func Test_Module_06(t *testing.T) {
	Check(t, "module_06")
}

func Test_Module_07(t *testing.T) {
	Check(t, "module_07")
}

func Test_Module_08(t *testing.T) {
	Check(t, "module_08")
}

func Test_Module_09(t *testing.T) {
	Check(t, "module_09")
}

func Test_Module_10(t *testing.T) {
	Check(t, "module_10")
}

// ===================================================================
// Permutations
// ===================================================================

func Test_Permute_01(t *testing.T) {
	Check(t, "permute_01")
}

func Test_Permute_02(t *testing.T) {
	Check(t, "permute_02")
}

func Test_Permute_03(t *testing.T) {
	Check(t, "permute_03")
}

func Test_Permute_04(t *testing.T) {
	Check(t, "permute_04")
}

func Test_Permute_05(t *testing.T) {
	Check(t, "permute_05")
}

func Test_Permute_06(t *testing.T) {
	Check(t, "permute_06")
}

func Test_Permute_07(t *testing.T) {
	Check(t, "permute_07")
}

func Test_Permute_08(t *testing.T) {
	Check(t, "permute_08")
}

func Test_Permute_09(t *testing.T) {
	Check(t, "permute_09")
}

// ===================================================================
// Lookups
// ===================================================================

func Test_Lookup_01(t *testing.T) {
	Check(t, "lookup_01")
}

func Test_Lookup_02(t *testing.T) {
	Check(t, "lookup_02")
}

func Test_Lookup_03(t *testing.T) {
	Check(t, "lookup_03")
}

func Test_Lookup_04(t *testing.T) {
	Check(t, "lookup_04")
}

func Test_Lookup_05(t *testing.T) {
	Check(t, "lookup_05")
}

func Test_Lookup_06(t *testing.T) {
	Check(t, "lookup_06")
}

func Test_Lookup_07(t *testing.T) {
	Check(t, "lookup_07")
}

func Test_Lookup_08(t *testing.T) {
	Check(t, "lookup_08")
}

// ===================================================================
// Interleaving
// ===================================================================

func Test_Interleave_01(t *testing.T) {
	Check(t, "interleave_01")
}

func Test_Interleave_02(t *testing.T) {
	Check(t, "interleave_02")
}

func Test_Interleave_03(t *testing.T) {
	Check(t, "interleave_03")
}

func Test_Interleave_04(t *testing.T) {
	Check(t, "interleave_04")
}

// ===================================================================
// Functions
// ===================================================================

func Test_Fun_01(t *testing.T) {
	Check(t, "fun_01")
}

func Test_Fun_02(t *testing.T) {
	Check(t, "fun_02")
}

func Test_Fun_03(t *testing.T) {
	Check(t, "fun_03")
}

func Test_Fun_04(t *testing.T) {
	Check(t, "fun_04")
}

func Test_Fun_05(t *testing.T) {
	Check(t, "fun_05")
}

func Test_Fun_06(t *testing.T) {
	Check(t, "fun_06")
}

// ===================================================================
// Pure Functions
// ===================================================================

func Test_PureFun_01(t *testing.T) {
	Check(t, "purefun_01")
}

func Test_PureFun_02(t *testing.T) {
	Check(t, "purefun_02")
}

/*
	func Test_PureFun_03(t *testing.T) {
		Check(t, "purefun_03")
	}
*/

func Test_PureFun_04(t *testing.T) {
	Check(t, "purefun_04")
}

// ===================================================================
// Complex Tests
// ===================================================================

func Test_Counter(t *testing.T) {
	Check(t, "counter")
}

func Test_ByteDecomp(t *testing.T) {
	Check(t, "byte_decomposition")
}

func Test_BitDecomp(t *testing.T) {
	Check(t, "bit_decomposition")
}

func Test_ByteSorting(t *testing.T) {
	Check(t, "byte_sorting")
}

func Test_WordSorting(t *testing.T) {
	Check(t, "word_sorting")
}

func Test_Memory(t *testing.T) {
	Check(t, "memory")
}

func TestSlow_Add(t *testing.T) {
	Check(t, "add")
}

func TestSlow_BinStatic(t *testing.T) {
	Check(t, "bin-static")
}

func TestSlow_BinDynamic(t *testing.T) {
	Check(t, "bin-dynamic")
}

func TestSlow_Wcp(t *testing.T) {
	Check(t, "wcp")
}

func TestSlow_Mxp(t *testing.T) {
	Check(t, "mxp")
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
func Check(t *testing.T, test string) {
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
	schema, errs := corset.CompileSourceFile(false, srcfile)
	// Check terms parsed ok
	if len(errs) > 0 {
		t.Fatalf("Error parsing %s: %v\n", filename, errs)
	}
	// Check valid traces are accepted
	accepts_file := fmt.Sprintf("%s.%s", test, "accepts")
	accepts := ReadTracesFile(accepts_file)
	CheckTraces(t, accepts_file, true, true, accepts, schema)
	// Check invalid traces are rejected
	rejects_file := fmt.Sprintf("%s.%s", test, "rejects")
	rejects := ReadTracesFile(rejects_file)
	CheckTraces(t, rejects_file, false, true, rejects, schema)
	// Check expanded traces are rejected
	expands_file := fmt.Sprintf("%s.%s", test, "expanded")
	expands := ReadTracesFile(expands_file)
	CheckTraces(t, expands_file, false, false, expands, schema)
	// Check auto-generated valid traces (if applicable)
	auto_accepts_file := fmt.Sprintf("%s.%s", test, "auto.accepts")
	if auto_accepts := ReadTracesFileIfExists(auto_accepts_file); auto_accepts != nil {
		CheckTraces(t, auto_accepts_file, true, true, auto_accepts, schema)
	}
	// Check auto-generated invalid traces (if applicable)
	auto_rejects_file := fmt.Sprintf("%s.%s", test, "auto.rejects")
	if auto_rejects := ReadTracesFileIfExists(auto_rejects_file); auto_rejects != nil {
		CheckTraces(t, auto_rejects_file, false, true, auto_rejects, schema)
	}
}

// Check a given set of tests have an expected outcome (i.e. are
// either accepted or rejected) by a given set of constraints.
func CheckTraces(t *testing.T, test string, expected bool, expand bool,
	traces [][]trace.RawColumn, hirSchema *hir.Schema) {
	for i, tr := range traces {
		if tr != nil {
			// Lower HIR => MIR
			mirSchema := hirSchema.LowerToMir()
			// Lower MIR => AIR
			airSchema := mirSchema.LowerToAir()
			// Align trace with schema, and check whether expanded or not.
			for padding := uint(0); padding <= MAX_PADDING; padding++ {
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
		// Check
		errs := sc.Accepts(100, schema, tr)
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
	lines := ReadInputFile(filename)
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
	return ReadTracesFile(name)
}

// ReadInputFile reads an input file as a sequence of lines.
func ReadInputFile(filename string) []string {
	filename = fmt.Sprintf("%s/%s", TestDir, filename)
	file, err := os.Open(filename)
	// Check whether file exists
	if errors.Is(err, os.ErrNotExist) {
		return []string{}
	} else if err != nil {
		panic(err)
	}

	reader := bufio.NewReaderSize(file, 1024*128)
	lines := make([]string, 0)
	// Read file line-by-line
	for {
		// Read the next line
		line := readLine(reader)
		// Check whether for EOF
		if line == nil {
			if err = file.Close(); err != nil {
				panic(err)
			}

			return lines
		}

		lines = append(lines, *line)
	}
}

// Read a single line
func readLine(reader *bufio.Reader) *string {
	var (
		bytes []byte
		bit   []byte
		err   error
	)
	//
	cont := true
	//
	for cont {
		bit, cont, err = reader.ReadLine()
		if err == io.EOF {
			return nil
		} else if err != nil {
			panic(err)
		}

		bytes = append(bytes, bit...)
	}
	// Convert to string
	str := string(bytes)
	// Done
	return &str
}
