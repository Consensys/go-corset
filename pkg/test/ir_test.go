package test

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"
	"testing"

	"github.com/consensys/go-corset/pkg/hir"
	"github.com/consensys/go-corset/pkg/schema"
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
// Modules
// ===================================================================

func Test_Module_01(t *testing.T) {
	Check(t, "module_01")
}

func Test_Module_02(t *testing.T) {
	Check(t, "module_02")
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
const MAX_PADDING uint = 1

// For a given set of constraints, check that all traces which we
// expect to be accepted are accepted, and all traces that we expect
// to be rejected are rejected.
func Check(t *testing.T, test string) {
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
	CheckTraces(t, test, true, accepts, schema)
	// Check invalid traces are rejected
	rejects := ReadTracesFile(test, "rejects")
	CheckTraces(t, test, false, rejects, schema)
}

// Check a given set of tests have an expected outcome (i.e. are
// either accepted or rejected) by a given set of constraints.
func CheckTraces(t *testing.T, test string, expected bool, traces []trace.Trace, hirSchema *hir.Schema) {
	for i, tr := range traces {
		if tr != nil {
			for padding := uint(0); padding <= MAX_PADDING; padding++ {
				// Lower HIR => MIR
				mirSchema := hirSchema.LowerToMir()
				// Lower MIR => AIR
				airSchema := mirSchema.LowerToAir()
				// Construct trace identifiers
				hirID := traceId{"HIR", test, expected, i + 1, padding, schema.RequiredSpillage(hirSchema)}
				mirID := traceId{"MIR", test, expected, i + 1, padding, schema.RequiredSpillage(mirSchema)}
				airID := traceId{"AIR", test, expected, i + 1, padding, schema.RequiredSpillage(airSchema)}
				// Check whether trace is input compatible with schema
				if err := sc.AlignInputs(tr, hirSchema); err != nil {
					// Alignment failed.  So, attempt alignment as expanded
					// trace instead.
					if err := sc.Align(tr, airSchema); err != nil {
						// Still failed, hence trace must be malformed in some way
						if expected {
							t.Errorf("Trace malformed (%s.accepts, line %d): [%s]", test, i+1, err)
						} else {
							t.Errorf("Trace malformed (%s.rejects, line %d): [%s]", test, i+1, err)
						}
					} else {
						// Aligned as expanded trace
						checkExpandedTrace(t, tr, airID, airSchema)
					}
				} else {
					// Aligned as unexpanded trace.
					checkInputTrace(t, tr, hirID, hirSchema)
					checkInputTrace(t, tr, mirID, mirSchema)
					checkInputTrace(t, tr, airID, airSchema)
				}
			}
		}
	}
}

func checkInputTrace(t *testing.T, tr trace.Trace, id traceId, schema sc.Schema) {
	// Clone trace (to ensure expansion does not affect subsequent tests)
	etr := tr.Clone()
	// Apply spillage
	trace.PadColumns(etr, id.spillage)
	// Expand trace
	err := sc.ExpandTrace(schema, etr)
	// Check
	if err != nil {
		t.Error(err)
	} else if err := sc.Align(etr, schema); err != nil {
		// Alignment problem
		t.Error(err)
	} else {
		checkExpandedTrace(t, etr, id, schema)
	}
}

func checkExpandedTrace(t *testing.T, tr trace.Trace, id traceId, schema sc.Schema) {
	// Apply padding
	trace.PadColumns(tr, id.padding)
	// Check
	err := sc.Accepts(schema, tr)
	// Determine whether trace accepted or not.
	accepted := (err == nil)
	// Process what happened versus what was supposed to happen.
	if !accepted && id.expected {
		//table.PrintTrace(tr)
		msg := fmt.Sprintf("Trace rejected incorrectly (%s, %s.accepts, line %d with spillage %d / padding %d): %s",
			id.ir, id.test, id.line, id.spillage, id.padding, err)
		t.Errorf(msg)
	} else if accepted && !id.expected {
		//printTrace(tr)
		msg := fmt.Sprintf("Trace accepted incorrectly (%s, %s.rejects, line %d with spillage %d / padding %d)",
			id.ir, id.test, id.line, id.spillage, id.padding)
		t.Errorf(msg)
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
	// Determines how much spillage was added to the original trace (prior to
	// expansion).
	spillage uint
}

// ReadTracesFile reads a file containing zero or more traces expressed as JSON, where
// each trace is on a separate line.
func ReadTracesFile(name string, ext string) []trace.Trace {
	lines := ReadInputFile(name, ext)
	traces := make([]trace.Trace, len(lines))
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

	if err != nil {
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
