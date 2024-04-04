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

// ===================================================================

func CheckEval(t *testing.T, val int64, str string) {
	CheckEvalBig(t, big.NewInt(val), str)
}

func CheckEvalBig(t *testing.T, val *big.Int, str string) {
	sexp,err := ParseToAir(str)
	if err != nil {
		t.Error(err)
	} else if sexp.EvalAt().Cmp(val) != 0 {
		t.Errorf("evaluation failed")
	}
}
