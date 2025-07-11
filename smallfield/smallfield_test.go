package smallfield

import (
	"math/big"
	"math/rand/v2"
	"testing"

	"github.com/consensys/go-corset/pkg/util/assert"
)

func TestField_Mul(t *testing.T) {
	f := New(1<<31 - 1) // Mersenne31

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

func TestField_Inverse(t *testing.T) {
	f := New(1<<31 - 1) // Mersenne31

	var i, m big.Int

	m.SetUint64(uint64(f.modulus))

	for range 1000000 {
		a := rand.Uint32N(f.modulus)

		i.SetUint64(uint64(a)).
			ModInverse(&i, &m).
			Lsh(&i, 32). // Montgomery form
			Mod(&i, &m)

		x := f.NewElement(a)
		x = f.Inverse(x)

		assert.Equal(t, i.Uint64(), x[0], "inverse of %d", a)
	}
}

func TestField_Halve(t *testing.T) {
	f := New(1<<31 - 1) // Mersenne31

	var i, j, m big.Int

	m.SetUint64(uint64(f.modulus))

	for range 1000000 {
		a := rand.Uint32N(f.modulus)
		x := f.NewElement(a)
		x = f.Half(x)

		i.SetUint64(uint64(x[0])).Add(&i, &i).Mod(&i, &m) // (a/2) as computed, multiplied by 2
		j.SetUint64(uint64(a)).Lsh(&j, 32).Mod(&j, &m)    // Montgomery representation of a

		assert.Equal(t, j.Uint64(), i.Uint64(), "halving of %d", a)
	}
}

func TestField_InverseNonMont(t *testing.T) {
	f := New(1<<31 - 1) // Mersenne31

	var i, m big.Int

	a := rand.Uint32N(f.modulus)
	m.SetUint64(uint64(f.modulus))

	for range 1000000 {
		i.SetUint64(uint64(a)).
			ModInverse(&i, &m)

		x := Element{a}
		x = f.inverse(x, Element{1})

		assert.Equal(t, i.Uint64(), x[0], "inverse of %d", a)
	}
}

func TestField_rSq(t *testing.T) {
	for _, p := range []uint64{3, 5, 7, 11, 1<<31 - 1} {
		assert.Equal(t, (((1<<63)%p)*2)%p, New(uint32(p)).rSq()[0], "modulus %d", p)
	}
}

func TestField_Sub(t *testing.T) {
	f := New(1<<31 - 1) // Mersenne31

	var i, j, m big.Int

	m.SetUint64(uint64(f.modulus))

	for range 100000 {
		a := rand.Uint32N(f.modulus)
		b := rand.Uint32N(f.modulus)

		i.SetUint64(uint64(a)).
			Sub(&i, j.SetUint64(uint64(b))).
			Lsh(&i, 32).
			Mod(&i, &m)

		x := f.NewElement(a)
		y := f.NewElement(b)

		x = f.Sub(x, y)

		assert.Equal(t, i.Uint64(), x[0])
	}
}
