package testA

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"
	"testing"
	"unicode/utf8"

	"github.com/consensys/go-corset/pkg/hir"
	"github.com/consensys/go-corset/pkg/table"
)

// Determines the (relative) location of the test directory.  That is
// where the corset test files (lisp) and the corresponding traces
// (accepts/rejects) are found.
const TestDir = "../../testdata"

// ===================================================================
// Basic Tests
// ===================================================================

func TestEval_Basic_01(t *testing.T) {
	Check(t, "basic_01")
}

func TestEval_Basic_02(t *testing.T) {
	Check(t, "basic_02")
}

func TestEval_Basic_03(t *testing.T) {
	Check(t, "basic_03")
}

func TestEval_Basic_04(t *testing.T) {
	Check(t, "basic_04")
}

func TestEval_Basic_05(t *testing.T) {
	Check(t, "basic_05")
}

func TestEval_Basic_06(t *testing.T) {
	Check(t, "basic_06")
}

func TestEval_Basic_07(t *testing.T) {
	Check(t, "basic_07")
}

func TestEval_Basic_08(t *testing.T) {
	Check(t, "basic_08")
}

func TestEval_Basic_09(t *testing.T) {
	Check(t, "basic_09")
}

// ===================================================================
// Domain Tests
// ===================================================================

func TestEval_Domain_01(t *testing.T) {
	Check(t, "domain_01")
}

func TestEval_Domain_02(t *testing.T) {
	Check(t, "domain_02")
}

func TestEval_Domain_03(t *testing.T) {
	Check(t, "domain_03")
}

// ===================================================================
// Block Tests
// ===================================================================

func TestEval_Block_01(t *testing.T) {
	Check(t, "block_01")
}

func TestEval_Block_02(t *testing.T) {
	Check(t, "block_02")
}

func TestEval_Block_03(t *testing.T) {
	Check(t, "block_03")
}

// ===================================================================
// Property Tests
// ===================================================================

func TestEval_Property_01(t *testing.T) {
	Check(t, "property_01")
}

// ===================================================================
// Shift Tests
// ===================================================================

func TestEval_Shift_01(t *testing.T) {
	Check(t, "shift_01")
}

func TestEval_Shift_02(t *testing.T) {
	Check(t, "shift_02")
}

func TestEval_Shift_03(t *testing.T) {
	Check(t, "shift_03")
}

func TestEval_Shift_04(t *testing.T) {
	Check(t, "shift_04")
}

func TestEval_Shift_05(t *testing.T) {
	Check(t, "shift_05")
}

func TestEval_Shift_06(t *testing.T) {
	Check(t, "shift_06")
}

// ===================================================================
// Normalisation Tests
// ===================================================================

func TestEval_Norm_01(t *testing.T) {
	Check(t, "norm_01")
}

func TestEval_Norm_02(t *testing.T) {
	Check(t, "norm_02")
}

func TestEval_Norm_03(t *testing.T) {
	Check(t, "norm_03")
}

func TestEval_Norm_04(t *testing.T) {
	Check(t, "norm_04")
}

func TestEval_Norm_05(t *testing.T) {
	Check(t, "norm_05")
}

func TestEval_Norm_06(t *testing.T) {
	Check(t, "norm_06")
}

func TestEval_Norm_07(t *testing.T) {
	Check(t, "norm_07")
}

// ===================================================================
// If-Zero
// ===================================================================

func TestEval_If_01(t *testing.T) {
	Check(t, "if_01")
}

func TestEval_If_02(t *testing.T) {
	Check(t, "if_02")
}

func TestEval_If_03(t *testing.T) {
	Check(t, "if_03")
}

// SEE: #152
// func TestEval_If_04(t *testing.T) {
// 	Check(t, "if_04")
// }

// ===================================================================
// Column Types
// ===================================================================

func TestEval_Type_01(t *testing.T) {
	Check(t, "type_01")
}

func TestEval_Type_02(t *testing.T) {
	Check(t, "type_02")
}

func TestEval_Type_03(t *testing.T) {
	Check(t, "type_03")
}

// ===================================================================
// Permutations
// ===================================================================

func TestEval_Permute_01(t *testing.T) {
	Check(t, "permute_01")
}

func TestEval_Permute_02(t *testing.T) {
	Check(t, "permute_02")
}

func TestEval_Permute_03(t *testing.T) {
	Check(t, "permute_03")
}

func TestEval_Permute_04(t *testing.T) {
	Check(t, "permute_04")
}

func TestEval_Permute_05(t *testing.T) {
	Check(t, "permute_05")
}

func TestEval_Permute_06(t *testing.T) {
	Check(t, "permute_06")
}

func TestEval_Permute_07(t *testing.T) {
	Check(t, "permute_07")
}

func TestEval_Permute_08(t *testing.T) {
	Check(t, "permute_08")
}

func TestEval_Permute_09(t *testing.T) {
	Check(t, "permute_09")
}

// ===================================================================
// Complex Tests
// ===================================================================

func TestEval_Counter(t *testing.T) {
	Check(t, "counter")
}

func TestEval_ByteDecomp(t *testing.T) {
	Check(t, "byte_decomposition")
}

func TestEval_BitDecomp(t *testing.T) {
	Check(t, "bit_decomposition")
}

func TestEval_ByteSorting(t *testing.T) {
	Check(t, "byte_sorting")
}

func TestEval_WordSorting(t *testing.T) {
	Check(t, "word_sorting")
}

func TestEval_Memory(t *testing.T) {
	Check(t, "memory")
}

func TestEval_Add(t *testing.T) {
	Check(t, "add")
}

func TestEval_BinStatic(t *testing.T) {
	Check(t, "bin-static")
}

func TestEval_BinDynamic(t *testing.T) {
	Check(t, "bin-dynamic")
}

func TestEval_Wcp(t *testing.T) {
	Check(t, "wcp")
}

// DISABLED: #152
// func TestEval_Mxp(t *testing.T) {
// 	Check(t, "mxp")
// }

// ===================================================================
// Test Helpers
// ===================================================================

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
func CheckTraces(t *testing.T, test string, expected bool, traces []*table.ArrayTrace, hirSchema *hir.Schema) {
	for i, tr := range traces {
		if tr != nil {
			// Lower HIR => MIR
			mirSchema := hirSchema.LowerToMir()
			// Lower MIR => AIR
			airSchema := mirSchema.LowerToAir()
			// Check HIR/MIR trace (if applicable)
			if airSchema.IsInputTrace(tr) == nil {
				// This is an unexpanded input trace.
				checkInputTrace(t, tr, traceId{"HIR", test, expected, i + 1}, hirSchema)
				checkInputTrace(t, tr, traceId{"MIR", test, expected, i + 1}, mirSchema)
				checkInputTrace(t, tr, traceId{"AIR", test, expected, i + 1}, airSchema)
			} else if airSchema.IsOutputTrace(tr) == nil {
				// This is an already expanded input trace.  Therefore, no need
				// to perform expansion.
				checkExpandedTrace(t, tr, traceId{"AIR", test, expected, i + 1}, airSchema)
			} else {
				// Trace appears to be malformed.
				err1 := airSchema.IsInputTrace(tr)
				err2 := airSchema.IsOutputTrace(tr)

				if expected {
					t.Errorf("Trace malformed (%s.accepts, line %d): [%s][%s]", test, i+1, err1, err2)
				} else {
					t.Errorf("Trace malformed (%s.rejects, line %d): [%s][%s]", test, i+1, err1, err2)
				}
			}
		}
	}
}

func checkInputTrace(t *testing.T, tr *table.ArrayTrace, id traceId, schema table.Schema) {
	// Clone trace (to ensure expansion does not affect subsequent tests)
	etr := tr.Clone()
	// Expand trace
	err := schema.ExpandTrace(etr)

	if err != nil {
		t.Error(err)
	} else {
		checkExpandedTrace(t, etr, id, schema)
	}
}

func checkExpandedTrace(t *testing.T, tr table.Trace, id traceId, schema table.Schema) {
	err := schema.Accepts(tr)
	// Determine whether trace accepted or not.
	accepted := (err == nil)
	// Process what happened versus what was supposed to happen.
	if !accepted && id.expected {
		//printTrace(tr)
		msg := fmt.Sprintf("Trace rejected incorrectly (%s, %s.accepts, line %d): %s", id.ir, id.test, id.line, err)
		t.Errorf(msg)
	} else if accepted && !id.expected {
		printTrace(tr)
		msg := fmt.Sprintf("Trace accepted incorrectly (%s, %s.rejects, line %d)", id.ir, id.test, id.line)
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
}

// ReadTracesFile reads a file containing zero or more traces expressed as JSON, where
// each trace is on a separate line.
func ReadTracesFile(name string, ext string) []*table.ArrayTrace {
	lines := ReadInputFile(name, ext)
	traces := make([]*table.ArrayTrace, len(lines))
	// Read constraints line by line
	for i, line := range lines {
		// Parse input line as JSON
		if line != "" && !strings.HasPrefix(line, ";;") {
			tr, err := table.ParseJsonTrace([]byte(line))
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

// Prints a trace in a more human-friendly fashion.
func printTrace(tr table.Trace) {
	n := tr.Width()
	//
	rows := make([][]string, n)
	for i := 0; i < n; i++ {
		rows[i] = traceColumnData(tr, i)
	}
	//
	widths := traceRowWidths(tr.Height(), rows)
	//
	printHorizontalRule(widths)
	//
	for _, r := range rows {
		printTraceRow(r, widths)
		printHorizontalRule(widths)
	}
}

func traceColumnData(tr table.Trace, col int) []string {
	n := tr.Height()
	data := make([]string, n+1)
	data[0] = tr.ColumnName(col)

	for row := 0; row < n; row++ {
		data[row+1] = tr.GetByIndex(col, row).String()
	}

	return data
}

func traceRowWidths(height int, rows [][]string) []int {
	widths := make([]int, height+1)

	for _, row := range rows {
		for i, col := range row {
			w := utf8.RuneCountInString(col)
			widths[i] = max(w, widths[i])
		}
	}

	return widths
}

func printTraceRow(row []string, widths []int) {
	for i, col := range row {
		fmt.Printf(" %*s |", widths[i], col)
	}

	fmt.Println()
}

func printHorizontalRule(widths []int) {
	for _, w := range widths {
		fmt.Print("-")

		for i := 0; i < w; i++ {
			fmt.Print("-")
		}
		fmt.Print("-+")
	}

	fmt.Println()
}
