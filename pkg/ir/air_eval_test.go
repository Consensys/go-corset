package ir

import (
	"math/big"
	"testing"
)

func TestEvalConst_1(t *testing.T) {
	CheckEval(t, 1, "1")
}

func TestEvalAdd_1(t *testing.T) {
	CheckEval(t,1,"(+ 1)")
}

func TestEvalAdd_2(t *testing.T) {
	CheckEval(t,3,"(+ 1 2)")
}

func TestEvalAdd_3(t *testing.T) {
	CheckEval(t,6,"(+ 1 2 3)")
}

func TestEvalSub_1(t *testing.T) {
	CheckEval(t,1,"(- 1)")
}

func TestEvalSub_2(t *testing.T) {
	CheckEval(t,4,"(- 6 2)")
}

func TestEvalSub_3(t *testing.T) {
	CheckEval(t,3,"(- 6 2 1)")
}

func TestEvalMul_1(t *testing.T) {
	CheckEval(t,1,"(* 1)")
}

func TestEvalMul_2(t *testing.T) {
	CheckEval(t,12,"(* 6 2)")
}

func TestEvalMul_3(t *testing.T) {
	CheckEval(t,36,"(* 6 2 3)")
}

// ===================================================================

func CheckEval(t *testing.T, val int64, str string) {
	CheckEvalBig(t, big.NewInt(val), str)
}

func CheckEvalBig(t *testing.T, val *big.Int, str string) {
	sexp,err := ParseSExpToAir(str)
	if err != nil {
		t.Error(err)
	} else if sexp.EvalAt().Cmp(val) != 0 {
		t.Errorf("evaluation failed")
	}
}
