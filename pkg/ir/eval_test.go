package ir

import (
	"math/big"
	"testing"
)

func TestEvalConst_1(t *testing.T) {
	e := Constant{big.NewInt(1)}
	CheckEval(t, &e, 1)
}

func TestEvalAdd_1(t *testing.T) {
	e1 := &Constant{big.NewInt(1)}
	e2 := AirAdd{[]AirExpr{e1}}
	CheckEval(t, &e2, 1)
}

func TestEvalAdd_2(t *testing.T) {
	e1 := Constant{big.NewInt(1)}
	e2 := Constant{big.NewInt(2)}
	e3 := AirAdd{[]AirExpr{&e1, &e2}}
	CheckEval(t, &e3, 3)
}

// ===================================================================

func CheckEval(t *testing.T, AirExpr AirExpr, val int64) {
	CheckEvalBig(t, AirExpr, big.NewInt(val))
}

func CheckEvalBig(t *testing.T, AirExpr AirExpr, val *big.Int) {
	if AirExpr.EvalAt().Cmp(val) != 0 {
		t.Errorf("evaluation failed")
	}
}
