package ir

import (
	"bufio"
	"encoding/json"
	"fmt"
	"math/big"
	"os"
	"testing"

	"github.com/Consensys/go-corset/pkg/trace"
)

// Determines the (relative) location of the test directory.  That is
// where the corset test files (lisp) and the corresponding traces
// (accepts/rejects) are found.
const TestDir = "../../testdata"

// Following definition is to improve readability.
type Trace = []trace.Column

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
// Test Helpers
// ===================================================================

// For a given set of constraints, check that all traces which we
// expect to be accepted are accepted, and all traces that we expect
// to be rejected are rejected.
func Check(t *testing.T, test string) {
	constraints := ReadConstraintsFile(test)
	// Check valid traces are accepted
	accepts := ReadTracesFile(test, "accepts")
	CheckTraces(t, test, true, accepts, constraints)
	// Check invalid traces are rejected
	rejects := ReadTracesFile(test, "rejects")
	CheckTraces(t, test, false, rejects, constraints)
}

// Check a given set of testdata have an expected outcome (i.e. are
// either accepted or rejected) by a given set of constraints.
func CheckTraces(t *testing.T, test string, expected bool, traces []Trace, constraints []trace.Constraint) {
	for i, tr := range traces {
		// Construct table for evaluation
		tbl := trace.NewLazyTable(tr, constraints)
		// Check whether constraints hold (or not)
		err := tbl.Check()
		// Process output
		if err != nil && expected {
			msg := fmt.Sprintf("Trace rejected incorrectly (%s.accepts, row %d): %s", test, i+1, err)
			t.Errorf(msg)
		} else if err == nil && !expected {
			msg := fmt.Sprintf("Trace accepted incorrectly (%s.rejects, row %d)", test, i+1)
			t.Errorf(msg)
		}
	}
}

// Read in a sequence of constraints from a given file.  For now, the
// constraints are always assumed to be vanishing constraints.
func ReadConstraintsFile(name string) []trace.Constraint {
	lines := ReadInputFile(name, "lisp")
	constraints := make([]trace.Constraint, len(lines))
	// Read constraints line by line
	for i, line := range lines {
		air, err := ParseSExpToAir(line)
		if err != nil {
			panic("error parsing constraint")
		}

		constraints[i] = &AirVanishingConstraint{"tmp", air}
	}
	//
	return constraints
}

// Read a file containing zero or more traces expressed as JSON, where
// each trace is on a separate line.
func ReadTracesFile(name string, ext string) []Trace {
	lines := ReadInputFile(name, ext)
	traces := make([]Trace, len(lines))
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
func ParseJsonTrace(jsn string, test string, ext string, row int) Trace {
	var data map[string][]*big.Int
	// Unmarshall
	jsonErr := json.Unmarshal([]byte(jsn), &data)
	if jsonErr != nil {
		msg := fmt.Sprintf("%s.%s:%d: %s", test, ext, row+1, jsonErr)
		panic(msg)
	}
	//
	var columns = make([]trace.Column, 0)
	//
	for name, raw := range data {
		columns = append(columns, trace.NewDataColumn(name, raw))
	}
	// Done
	return columns
}

// Read an input file as a sequence of lines.
func ReadInputFile(name string, ext string) []string {
	name = fmt.Sprintf("%s/%s.%s", TestDir, name, ext)

	file, err := os.Open(name)
	if err != nil {
		panic(err)
	}
	defer file.Close()

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
	// Done
	return lines
}
