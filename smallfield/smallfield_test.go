package smallfield

import (
	"math/big"
	"math/rand/v2"
	"testing"

	"github.com/consensys/go-corset/pkg/util/assert"
)

func TestField_Mul(t *testing.T) {
	f := NewField(1<<31 - 1) // Mersenne31

	var i, j, m big.Int

	m.SetUint64(uint64(f.modulus))

	for range 10000 {
		a := rand.Uint32N(f.modulus)
		b := rand.Uint32N(f.modulus)

		i.SetUint64(uint64(a)).
			Mul(&i, j.SetUint64(uint64(b))).
			Lsh(&i, 32).
			Mod(&i, &m)

		x := f.NewElement(a)
		y := f.NewElement(b)

		x = f.Mul(x, y)

		assert.Equal(t, i.Uint64(), x[0])
	}
}
