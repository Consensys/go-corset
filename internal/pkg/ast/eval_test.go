package ast

import (
	"math/big"
	"testing"
)

func TestEvalConst(t *testing.T) {
	one := big.NewInt(1)
	e := Const{one}
	if e.eval_at() != one {
		t.Errorf("evaluation failed")
	}
}
