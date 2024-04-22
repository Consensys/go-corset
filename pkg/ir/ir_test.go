package ir

import (
	"bufio"
	"encoding/json"
	"fmt"
	"math/big"
	"os"
	"testing"
	"github.com/consensys/go-corset/pkg/trace"
	"github.com/consensys/gnark-crypto/ecc/bls12-377/fr"
)

// Determines the (relative) location of the test directory.  That is
// where the corset test files (lisp) and the corresponding traces
// (accepts/rejects) are found.
const TestDir = "../../tests"

// Following definition is to improve readability.
type Trace = []trace.Column

// ===================================================================
// Basic Tests
// ===================================================================

func TestEval_Basic_01(t *testing.T) {
	Check(t,"basic_01")
}

func TestEval_Basic_02(t *testing.T) {
	Check(t,"basic_02")
}

func TestEval_Basic_03(t *testing.T) {
	Check(t,"basic_03")
}

func TestEval_Basic_04(t *testing.T) {
	Check(t,"basic_04")
}

func TestEval_Basic_05(t *testing.T) {
	Check(t,"basic_05")
}

func TestEval_Basic_06(t *testing.T) {
	Check(t,"basic_06")
}

func TestEval_Basic_07(t *testing.T) {
	Check(t,"basic_07")
}

func TestEval_Basic_08(t *testing.T) {
	Check(t,"basic_08")
}

func TestEval_Basic_09(t *testing.T) {
	Check(t,"basic_09")
}

// ===================================================================
// Shift Tests
// ===================================================================

func TestEval_Shift_01(t *testing.T) {
	Check(t,"shift_01")
}

func TestEval_Shift_02(t *testing.T) {
	Check(t,"shift_02")
}

func TestEval_Shift_03(t *testing.T) {
	Check(t,"shift_03")
}

func TestEval_Shift_04(t *testing.T) {
	Check(t,"shift_04")
}

func TestEval_Shift_05(t *testing.T) {
	Check(t,"shift_05")
}

func TestEval_Shift_06(t *testing.T) {
	Check(t,"shift_06")
}

// ===================================================================
// Normalisation Tests
// ===================================================================

func TestEval_Norm_01(t *testing.T) {
	Check(t,"norm_01")
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
	accepts := ReadTracesFile(test,"accepts")
	CheckTraces(t,test,true,accepts,constraints)
	// Check invalid traces are rejected
	rejects := ReadTracesFile(test,"rejects")
	CheckTraces(t,test,false,rejects,constraints)
}

// Check a given set of tests have an expected outcome (i.e. are
// either accepted or rejected) by a given set of constraints.
func CheckTraces(t *testing.T, test string, expected bool, traces []Trace, constraints []MirConstraint) {
	for i,tr := range traces {
		// Construct table for evaluation
		mir := trace.NewLazyTable(tr, constraints)
		air := trace.EmptyLazyTable[AirConstraint]()
		// Lower MIR => AIR
		LowerToAir(mir,air)
		// Check MIR table
		CheckTrace(t,"MIR",test,i+1,expected,mir)
		// Check AIR table
		CheckTrace(t,"AIR",test,i+1,expected,air)
	}
}

func CheckTrace[C trace.Constraint](t *testing.T, ir string, test string, row int, expected bool, tbl trace.Table[C]) {
	// Check whether constraints hold (or not)
	err := tbl.Check()
	// Process output
	if err != nil && expected {
		msg := fmt.Sprintf("Trace rejected incorrectly (%s, %s.accepts, line %d): %s",ir,test,row,err)
		t.Errorf(msg)
	} else if err == nil && !expected {
		msg := fmt.Sprintf("Trace accepted incorrectly (%s, %s.rejects, line %d)",ir,test,row)
		t.Errorf(msg)
	}
}

// Read in a sequence of constraints from a given file.  For now, the
// constraints are always assumed to be vanishing constraints.
func ReadConstraintsFile(name string) []MirConstraint {
	lines := ReadInputFile(name,"lisp")
	constraints := make([]MirConstraint,len(lines))
	// Read constraints line by line
	for i,line := range lines {
		air,err := ParseSExpToMir(line)
		if err != nil { panic(err) }
		constraints[i] = &trace.VanishingConstraint[MirExpr]{Handle: "tmp", Expr: air}
	}
	//
	return constraints
}

// Read a file containing zero or more traces expressed as JSON, where
// each trace is on a separate line.
func ReadTracesFile(name string, ext string) []Trace {
	lines := ReadInputFile(name,ext)
	traces := make([]Trace,len(lines))
	// Read constraints line by line
	for i,line := range lines {
		// Parse input line as JSON
		traces[i] = ParseJsonTrace(line,name,ext,i)
	}
	return traces
}

// Parse a trace expressed in JSON notation.  For example, {"X": [0],
// "Y": [1]} is a trace containing one row of data each for two
// columns "X" and "Y".
func ParseJsonTrace(jsn string, test string, ext string, row int) Trace {
	var raw_data map[string][]*big.Int
	// Unmarshall
	json_err := json.Unmarshal([]byte(jsn), &raw_data)
	if json_err != nil {
		msg := fmt.Sprintf("%s.%s:%d: %s",test,ext,row+1,json_err)
		panic(msg)
	}
	//
	var columns Trace = make([]trace.Column,0)
	//
	for name,raw_ints := range raw_data {
		raw_elements := ToFieldElements(raw_ints)
		columns = append(columns,trace.NewDataColumn(name,raw_elements))
	}
	// Done
	return columns
}

// Read an input file as a sequence of lines.
func ReadInputFile(name string, ext string) []string {
	name = fmt.Sprintf("%s/%s.%s",TestDir,name,ext)
	file, err := os.Open(name)
	if err != nil { panic(err) }
	defer file.Close()
	scanner := bufio.NewScanner(file)
	lines := make([]string,0)
	// Read file line-by-line
	for scanner.Scan() {
		lines = append(lines,scanner.Text())
	}
	// Sanity check we read everything
	if err := scanner.Err(); err != nil { panic(err) }
	// Done
	return lines
}

// Convert an array of big integers into an array of field elements.
func ToFieldElements(ints []*big.Int) []*fr.Element {
	elements := make([]*fr.Element,len(ints))
	// Convert each integer in turn
	for i,v := range ints {
		element := new(fr.Element)
		element.SetBigInt(v)
		elements[i] = element
	}
	// Done
	return elements
}
