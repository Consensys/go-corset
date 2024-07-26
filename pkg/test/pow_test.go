package test

import (
	"math/big"
	"testing"

	"github.com/consensys/gnark-crypto/ecc/bls12-377/fr"
	"github.com/consensys/go-corset/pkg/util"
)

const POW_BASE_MAX uint = 65536
const POW_BASE_INC uint = 8

func Test_Pow_01(t *testing.T) {
	PowCheckLoop(t, 0)
}

func Test_Pow_02(t *testing.T) {
	PowCheckLoop(t, 1)
}

func Test_Pow_03(t *testing.T) {
	PowCheckLoop(t, 2)
}

func Test_Pow_04(t *testing.T) {
	PowCheckLoop(t, 3)
}

func Test_Pow_05(t *testing.T) {
	PowCheckLoop(t, 4)
}

func Test_Pow_06(t *testing.T) {
	PowCheckLoop(t, 5)
}

func Test_Pow_07(t *testing.T) {
	PowCheckLoop(t, 6)
}

func Test_Pow_08(t *testing.T) {
	PowCheckLoop(t, 7)
}

func PowCheckLoop(t *testing.T, first uint) {
	// Enable parallel testing
	t.Parallel()
	// Run through the loop
	for i := first; i < POW_BASE_MAX; i += POW_BASE_INC {
		for j := uint64(0); j < 256; j++ {
			PowCheck(t, i, j)
		}
	}
}

// Check pow computed correctly.  This is done by comparing against the existing
// gnark function.
func PowCheck(t *testing.T, base uint, pow uint64) {
	k := big.NewInt(int64(pow))
	v1 := fr.NewElement(uint64(base))
	v2 := fr.NewElement(uint64(base))
	// V1 computed using our optimised method
	util.Pow(&v1, pow)
	// V2 computed using existing gnark function
	v2.Exp(v2, k)
	// Final sanity check
	if v1.Cmp(&v2) != 0 {
		t.Errorf("Pow(%d,%d)=%s (not %s)", base, pow, v1.String(), v2.String())
	}
}
