package ir

import (
	"fmt"
	"math/big"
	"testing"
	"github.com/Consensys/go-corset/pkg/trace"
)

// ===================================================================
// Pure Expressions
// ===================================================================

func TestEvalConst_1(t *testing.T) {
	CheckPureEval(t, big.NewInt(1), "1")
}

func TestEvalAdd_1(t *testing.T) {
	CheckPureEval(t,big.NewInt(1),"(+ 1)")
}

func TestEvalAdd_2(t *testing.T) {
	CheckPureEval(t,big.NewInt(3),"(+ 1 2)")
}

func TestEvalAdd_3(t *testing.T) {
	CheckPureEval(t,big.NewInt(6),"(+ 1 2 3)")
}

func TestEvalSub_1(t *testing.T) {
	CheckPureEval(t,big.NewInt(1),"(- 1)")
}

func TestEvalSub_2(t *testing.T) {
	CheckPureEval(t,big.NewInt(4),"(- 6 2)")
}

func TestEvalSub_3(t *testing.T) {
	CheckPureEval(t,big.NewInt(3),"(- 6 2 1)")
}

func TestEvalMul_1(t *testing.T) {
	CheckPureEval(t,big.NewInt(1),"(* 1)")
}

func TestEvalMul_2(t *testing.T) {
	CheckPureEval(t,big.NewInt(12),"(* 6 2)")
}

func TestEvalMul_3(t *testing.T) {
	CheckPureEval(t,big.NewInt(36),"(* 6 2 3)")
}

func TestEvalAddMul_1(t *testing.T) {
	CheckPureEval(t,big.NewInt(22),"(+ (* 4 5) 2)")
}

func TestEvalAddMul_2(t *testing.T) {
	CheckPureEval(t,big.NewInt(23),"(+ 3 (* 4 5))")
}

// ===================================================================
// Impure Expressions
// ===================================================================

func TestEvalColumnAccess_1(t *testing.T) {
	results := []*big.Int{big.NewInt(1),big.NewInt(2),big.NewInt(3),big.NewInt(4)}
	CheckTable(t,DataSet_1(),results,"X")
}

func TestEvalColumnAccess_2(t *testing.T) {
	results := []*big.Int{big.NewInt(5),big.NewInt(6),big.NewInt(7),big.NewInt(8)}
	CheckTable(t,DataSet_1(),results,"Y")
}

func TestEvalColumnAccess_3(t *testing.T) {
	results := []*big.Int{big.NewInt(6),big.NewInt(8),big.NewInt(10),big.NewInt(12)}
	CheckTable(t,DataSet_1(),results,"(+ X Y)")
	CheckTable(t,DataSet_1(),results,"(+ Y X)")
}

func TestEvalColumnAccess_4(t *testing.T) {
	results := []*big.Int{big.NewInt(-4),big.NewInt(-4),big.NewInt(-4),big.NewInt(-4)}
	CheckTable(t,DataSet_1(),results,"(- X Y)")
}

func TestEvalColumnAccess_5(t *testing.T) {
	results := []*big.Int{big.NewInt(11),big.NewInt(14),big.NewInt(17),big.NewInt(20)}
	CheckTable(t,DataSet_1(),results,"(+ X (* 2 Y))")
}

func TestEvalShiftAccess_1(t *testing.T) {
	results := []*big.Int{big.NewInt(2),big.NewInt(3),big.NewInt(4),nil}
	CheckTable(t,DataSet_1(),results,"(shift X 1)")
}

func TestEvalShiftAccess_2(t *testing.T) {
	results := []*big.Int{nil,big.NewInt(1),big.NewInt(2),big.NewInt(3)}
	CheckTable(t,DataSet_1(),results,"(shift X -1)")
}

func TestEvalShiftAccess_3(t *testing.T) {
	results := []*big.Int{big.NewInt(7),big.NewInt(9),big.NewInt(11),nil}
	CheckTable(t,DataSet_1(),results,"(+ (shift X 1) Y)")
}

func TestEvalShiftAccess_4(t *testing.T) {
	results := []*big.Int{big.NewInt(-3),big.NewInt(-3),big.NewInt(-3),nil}
	CheckTable(t,DataSet_1(),results,"(- (shift X 1) Y)")
}

func TestEvalShiftAccess_5(t *testing.T) {
	results := []*big.Int{big.NewInt(10),big.NewInt(18),big.NewInt(28),nil}
	CheckTable(t,DataSet_1(),results,"(* (shift X 1) Y)")
}

// NOTE: this test requires #19 before it could pass.
//
// func TestEvalNormalise(t *testing.T) {
// 	results := []*big.Int{big.NewInt(2),big.NewInt(3),big.NewInt(4),nil}
// 	CheckTable(t,DataSet_1(),results,"(norm X)")
// }

// ===================================================================
// Data Sets
// ===================================================================

func DataSet_1() trace.Table {
	schema := []string{"X","Y"}
	x_data := []*big.Int{big.NewInt(1),big.NewInt(2),big.NewInt(3),big.NewInt(4)}
	y_data := []*big.Int{big.NewInt(5),big.NewInt(6),big.NewInt(7),big.NewInt(8)}
	tbl,_ := trace.NewLazyTable(schema,x_data,y_data)
	return tbl
}

// ===================================================================
// Test Helpers
// ===================================================================

// Check that evaluating a pure expression yields a specific result.
func CheckPureEval(t *testing.T, val *big.Int, str string) {
	mir,err := ParseSExpToMir(str)
	// Construct empty table for the evaluation context.
	tbl := trace.EmptyLazyTable()
	//
	if err != nil {
		t.Error(err)
	} else {
		// Lower
		air := mir.LowerToAir()
		// Evaluate
		if air.EvalAt(0,tbl).Cmp(val) != 0 {
			t.Errorf("evaluation failed")
		}
	}
}

// Check that evaluating a given (vanishing) constraint on all rows of
// a table yields the expected results.
func CheckTable(t *testing.T, tbl trace.Table, data []*big.Int, str string) {
	// Parse string as MIR
	mir,err := ParseSExpToMir(str)
	//
	if err != nil {
		t.Error(err)
	} else if tbl.Height() != len(data) {
		t.Errorf("incorrect number of data points")
	} else {
		// Lower
		air := mir.LowerToAir()
		// Evaluate
		for i,expected := range data {
			// Compute evaluation point (MIR)
			mir_actual := mir.EvalAt(i,tbl)
			// Compute evaluation point (AIR)
			air_actual := air.EvalAt(i,tbl)
			// Check evaluation yields expected outcome
			if !Equal(mir_actual,air_actual) {
				// MIR and AIR evaluation differs.
				msg := fmt.Sprintf("Evaluation MIR/AIR differs on row %d: %s != %s",i,mir_actual,air_actual)
				t.Errorf(msg)
			} else if !Equal(air_actual,expected) {
				msg := fmt.Sprintf("Evaluation failed on row %d: was %s, expected %s",i,air_actual,expected)
				t.Errorf(msg)
			}
		}
	}
}

// Check whether two evaluation points match.
func Equal(lhs *big.Int, rhs *big.Int) bool {
	if lhs != rhs && (lhs == nil || rhs == nil || lhs.Cmp(rhs) != 0) {
		return false
	} else {
		return true
	}
}
