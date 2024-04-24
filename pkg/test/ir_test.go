package testA

import (
	"bufio"
	"encoding/json"
	"fmt"
	"math/big"
	"os"
	//"strings"
	"testing"
	"github.com/consensys/go-corset/pkg/air"
	"github.com/consensys/go-corset/pkg/mir"
	"github.com/consensys/go-corset/pkg/hir"
	"github.com/consensys/go-corset/pkg/table"
	"github.com/consensys/gnark-crypto/ecc/bls12-377/fr"
)

// Determines the (relative) location of the test directory.  That is
// where the corset test files (lisp) and the corresponding traces
// (accepts/rejects) are found.
const TestDir = "../../tests"

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

func TestEval_Norm_02(t *testing.T) {
	Check(t,"norm_02")
}

func TestEval_Norm_03(t *testing.T) {
	Check(t,"norm_03")
}

func TestEval_Norm_04(t *testing.T) {
	Check(t,"norm_04")
}

func TestEval_Norm_05(t *testing.T) {
	Check(t,"norm_05")
}

func TestEval_Norm_06(t *testing.T) {
	Check(t,"norm_06")
}

func TestEval_Norm_07(t *testing.T) {
	Check(t,"norm_07")
}

// ===================================================================
// If-Zero
// ===================================================================

func TestEval_If_01(t *testing.T) {
	Check(t,"if_01")
}

func TestEval_If_02(t *testing.T) {
	Check(t,"if_02")
}

// ===================================================================
// Complex Tests
// ===================================================================

func TestEval_Counter(t *testing.T) {
	Check(t,"counter")
}

// ===================================================================
// Test Helpers
// ===================================================================

// For a given set of constraints, check that all traces which we
// expect to be accepted are accepted, and all traces that we expect
// to be rejected are rejected.
func Check(t *testing.T, test string) {
	schema := ReadSchemaFile(test)
	// Check valid traces are accepted
	accepts := ReadTracesFile(test,"accepts")
	CheckTraces(t,test,true,accepts,schema)
	// Check invalid traces are rejected
	rejects := ReadTracesFile(test,"rejects")
	CheckTraces(t,test,false,rejects,schema)
}

// Check a given set of tests have an expected outcome (i.e. are
// either accepted or rejected) by a given set of constraints.
func CheckTraces(t *testing.T, test string, expected bool, traces []*table.ArrayTrace, hirSchema *hir.Schema) {
	for i,tr := range traces {
		// Construct table for evaluation
		mirSchema := table.EmptySchema[mir.Column,mir.Constraint]()
		airSchema := table.EmptySchema[air.Column,air.Constraint]()
		// Lower HIR => MIR
		hir.LowerToMir(hirSchema,mirSchema)
		// Lower MIR => AIR
		mir.LowerToAir(mirSchema,airSchema)
		// Check HIR/MIR trace (if applicable)
		// if ValidHirMirTrace(mirTbl) {
		// CheckTrace(t,"HIR",test,i+1,expected,hirTbl)
		// CheckTrace(t,"MIR",test,i+1,expected,mirTbl)
		// }
		// Check AIR trace
		check(t,"AIR",test,i+1,expected,airSchema.Accepts(tr))
	}
}

func check(t *testing.T, ir string, test string, line int, expected bool, accepted bool) {
	// Process what happened versus what was supposed to happen.
	if !accepted && expected {
		msg := fmt.Sprintf("Trace rejected incorrectly (%s, %s.accepts, line %d)",ir,test,line)
		t.Errorf(msg)
	} else if accepted && !expected {
		msg := fmt.Sprintf("Trace accepted incorrectly (%s, %s.rejects, line %d)",ir,test,line)
		t.Errorf(msg)
	}
}

// In some circumstances there are traces which should not be considered
// at the MIR level.  The reason for this is that they contain manual
// entries for computed columns (e.g. in an effort to prevent a trace
// from being rejected).  As such, the MIR level does not see those
// columns and, hence, cannot always know the trace should have been
// rejected.
//
// For now, we simply say that any trace containing a column whose
// name suggests it is (or represents) a computed column is not a
// valid MIR table.
// func ValidHirMirTrace[C table.Column, R table.Constraint](tbl table.Table[C,R]) bool {
// 	for _,col := range tbl.Columns() {
// 		if strings.Contains(col.Name(),"(") {
// 			return false
// 		}
// 	}
// 	return true
// }

// Read in a sequence of constraints from a given file.  For now, the
// constraints are always assumed to be vanishing constraints.
func ReadSchemaFile(name string) *hir.Schema {
	lines := ReadInputFile(name,"lisp")
	// Construct (initially empty) schema
	schema := table.EmptySchema[hir.Column,hir.Constraint]()
	// Read constraints line by line
	for _,line := range lines {
		expr,err := hir.ParseSExp(line)
		if err != nil { panic(err) }
		schema.AddConstraint(&hir.VanishingConstraint{Handle: "tmp", Expr: expr})
	}
	// Done (for now)
	return schema
}

// Read a file containing zero or more traces expressed as JSON, where
// each trace is on a separate line.
func ReadTracesFile(name string, ext string) []*table.ArrayTrace {
	lines := ReadInputFile(name,ext)
	traces := make([]*table.ArrayTrace,len(lines))
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
func ParseJsonTrace(jsn string, test string, ext string, row int) *table.ArrayTrace {
	var raw_data map[string][]*big.Int
	// Unmarshall
	json_err := json.Unmarshal([]byte(jsn), &raw_data)
	if json_err != nil {
		msg := fmt.Sprintf("%s.%s:%d: %s",test,ext,row+1,json_err)
		panic(msg)
	}
	//
	trace := table.EmptyArrayTrace()
	//
	for name,raw_ints := range raw_data {
		// Translate raw bigints into raw field elements
		raw_elements := ToFieldElements(raw_ints)
		// Add new column to the trace
		trace.AddColumn(name,raw_elements)
	}
	// Done
	return trace
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
