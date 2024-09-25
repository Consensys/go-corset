package test

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
	"testing"

	"github.com/consensys/go-corset/pkg/hir"
	sc "github.com/consensys/go-corset/pkg/schema"
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
// Column Types
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

// ===================================================================
// Constant Propagation
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
	// Enable testing each trace in parallel
	t.Parallel()
	// Read constraints file
	bytes, err := os.ReadFile(fmt.Sprintf("%s/%s.lisp", TestDir, test))
	// Check test file read ok
	if err != nil {
		t.Fatal(err)
	}
	// Parse terms into an HIR schema
	schema, err := hir.ParseSchemaString(string(bytes))
	// Check terms parsed ok
	if err != nil {
		t.Fatalf("Error parsing %s.lisp: %s\n", test, err)
	}
	// Check valid traces are accepted
	accepts := ReadTracesFile(test, "accepts")
	CheckTraces(t, test, true, true, accepts, schema)
	// Check invalid traces are rejected
	rejects := ReadTracesFile(test, "rejects")
	CheckTraces(t, test, false, true, rejects, schema)
	// Check expanded traces are rejected
	expands := ReadTracesFile(test, "expanded")
	CheckTraces(t, test, false, false, expands, schema)
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
			t.Errorf("Trace rejected incorrectly (%s, %s.accepts, line %d with padding %d): %s",
				id.ir, id.test, id.line, id.padding, errs)
		} else if accepted && !id.expected {
			//printTrace(tr)
			t.Errorf("Trace accepted incorrectly (%s, %s.rejects, line %d with padding %d)",
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
func ReadTracesFile(name string, ext string) [][]trace.RawColumn {
	lines := ReadInputFile(name, ext)
	traces := make([][]trace.RawColumn, len(lines))
	// Read constraints line by line
	for i, line := range lines {
		// Parse input line as JSON
		if line != "" && !strings.HasPrefix(line, ";;") {
			tr, err := json.FromBytes([]byte(line))
			if err != nil {
				msg := fmt.Sprintf("%s.%s:%d: %s", name, ext, i+1, err)
				panic(msg)
			}

			traces[i] = tr
		}
	}

	return traces
}

// ReadInputFile reads an input file as a sequence of lines.
func ReadInputFile(name string, ext string) []string {
	name = fmt.Sprintf("%s/%s.%s", TestDir, name, ext)
	file, err := os.Open(name)
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
