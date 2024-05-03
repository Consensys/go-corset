package testA

import (
	"bufio"
	"encoding/json"
	"fmt"
	"math/big"
	"os"
	"strings"
	"testing"

	"github.com/consensys/gnark-crypto/ecc/bls12-377/fr"
	"github.com/consensys/go-corset/pkg/air"
	"github.com/consensys/go-corset/pkg/hir"
	"github.com/consensys/go-corset/pkg/mir"
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
// Property Tests
// ===================================================================

// func TestEval_Property_01(t *testing.T) {
// 	Check(t, "property_01")
// }

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

// ===================================================================
// Column Types
// ===================================================================

func TestEval_Type_01(t *testing.T) {
	Check(t, "type_01")
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
	// Parse as a schema
	schema, err := hir.ParseSchemaSExp(string(bytes))
	// Check test file parsed ok
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
		mirSchema := table.EmptySchema[mir.Column, mir.Constraint]()
		airSchema := air.EmptySchema()
		// Lower HIR => MIR
		hir.LowerToMir(hirSchema, mirSchema)
		// Lower MIR => AIR
		mir.LowerToAir(mirSchema, airSchema)
		// Check HIR/MIR trace (if applicable)
		if ValidHirMirTrace(tr) {
			checkTrace(t, "HIR", test, i+1, expected, hirSchema.AcceptsTrace(tr))
			checkTrace(t, "MIR", test, i+1, expected, mirSchema.AcceptsTrace(tr))
		}
		// Perform trace expansion
		airSchema.ExpandTrace(tr)
		// Check AIR trace
		err, _ := airSchema.AcceptsTrace(tr)
		checkTrace(t, "AIR", test, i+1, expected, err == nil)
	}
}

func checkTrace(t *testing.T, ir string, test string, line int, expected bool, accepted bool) {
	// Process what happened versus what was supposed to happen.
	if !accepted && expected {
		msg := fmt.Sprintf("Trace rejected incorrectly (%s, %s.accepts, line %d)", ir, test, line)
		t.Errorf(msg)
	} else if accepted && !expected {
		msg := fmt.Sprintf("Trace accepted incorrectly (%s, %s.rejects, line %d)", ir, test, line)
		t.Errorf(msg)
	}
}

// ValidHirMirTrace In some circumstances there are traces which should not be considered
// at the MIR level.  The reason for this is that they contain manual
// entries for computed columns (e.g. in an effort to prevent a trace
// from being rejected).  As such, the MIR level does not see those
// columns and, hence, cannot always know the trace should have been
// rejected.
//
// For now, we simply say that any trace containing a column whose
// name suggests it is (or represents) a computed column is not a
// valid MIR table.
func ValidHirMirTrace(tbl *table.ArrayTrace) bool {
	for _, col := range tbl.Columns() {
		if strings.Contains(col.Name(), "(") {
			return false
		}
	}

	return true
}

// ReadTracesFile reads a file containing zero or more traces expressed as JSON, where
// each trace is on a separate line.
func ReadTracesFile(name string, ext string) []*table.ArrayTrace {
	lines := ReadInputFile(name, ext)
	traces := make([]*table.ArrayTrace, len(lines))
	// Read constraints line by line
	for i, line := range lines {
		// Parse input line as JSON
		traces[i] = ParseJsonTrace(line, name, ext, i)
	}

	return traces
}

// Parse a trace expressed in JSON notation.  For example, {"X": [0],
// "Y": [1]} is a trace containing one row of data each for two
// columns "X" and "Y".
func ParseJsonTrace(jsn string, test string, ext string, row int) *table.ArrayTrace {
	var rawData map[string][]*big.Int
	// Unmarshall
	jsonErr := json.Unmarshal([]byte(jsn), &rawData)
	if jsonErr != nil {
		msg := fmt.Sprintf("%s.%s:%d: %s", test, ext, row+1, jsonErr)
		panic(msg)
	}

	trace := table.EmptyArrayTrace()

	for name, rawInts := range rawData {
		// Translate raw bigints into raw field elements
		rawElements := ToFieldElements(rawInts)
		// Add new column to the trace
		trace.AddColumn(name, rawElements)
	}

	// Done.
	return trace
}

// ReadInputFile reads an input file as a sequence of lines.
func ReadInputFile(name string, ext string) []string {
	name = fmt.Sprintf("%s/%s.%s", TestDir, name, ext)
	file, err := os.Open(name)

	if err != nil {
		panic(err)
	}

	scanner := bufio.NewScanner(file)
	lines := make([]string, 0)
	// Read file line-by-line
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	// Sanity check we read everything
	if err := scanner.Err(); err != nil {
		panic(err)
	}
	// Close file as complete
	if err = file.Close(); err != nil {
		panic(err)
	}

	// Done
	return lines
}

// ToFieldElements converts an array of big integers into an array of field elements.
func ToFieldElements(ints []*big.Int) []*fr.Element {
	elements := make([]*fr.Element, len(ints))
	// Convert each integer in turn.
	for i, v := range ints {
		element := new(fr.Element)
		element.SetBigInt(v)
		elements[i] = element
	}

	// Done.
	return elements
}
