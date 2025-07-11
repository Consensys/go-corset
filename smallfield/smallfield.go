package smallfield

import (
	"cmp"
	"math/big"
	"math/bits"
)

// Element of a prime order field, represented in Montgomery form to speed up multiplications.
type Element [1]uint32 // defined as an array to prevent mistaken use of arithmetic operators, or naive assignments.

// A Field of prime order, less than 2³¹.
type Field struct {
	modulus           uint32
	negModulusInvModR uint32
}

// New field of the given order.
func New(modulus uint32) Field {
	if modulus >= 1<<31 {
		panic("modulus too large") // need at least one bit of "slack"
	}

	m := big.NewInt(int64(modulus))
	m.ModInverse(m, big.NewInt(1<<32))

	return Field{modulus: modulus, negModulusInvModR: uint32(1<<32 - m.Uint64())}
}

// Add x0 + x1
func (f Field) Add(x0, x1 Element) Element {
	res := x0[0] + x1[0]

	if reduced, borrow := bits.Sub32(res, f.modulus, 0); borrow == 0 {
		res = reduced
	}

	return Element{res}
}

// Sub x0 - x1
func (f Field) Sub(x0, x1 Element) Element {
	res, borrow := bits.Sub32(x0[0], x1[0], 0)
	if borrow != 0 {
		res += f.modulus
	}

	return Element{res}
}

// montgomeryReduce x -> x.R⁻¹ (mod m)
func (f Field) montgomeryReduce(x uint64) Element {
	// textbook Montgomery reduction
	const R = 1 << 32
	m := (x * uint64(f.negModulusInvModR)) % R // m = x * (-modulus⁻¹) (mod R)

	res := Element{uint32((x + m*uint64(f.modulus)) / R)}

	if res[0] >= f.modulus {
		res[0] -= f.modulus
	}

	return res
}

// ToUint32 returns the numerical (non-Montgomery)
// value of x.
func (f Field) ToUint32(x Element) uint32 {
	return f.montgomeryReduce(uint64(x[0]))[0]
}

// Mul x0 * x1
func (f Field) Mul(x0, x1 Element) Element {
	return f.montgomeryReduce(uint64(x0[0]) * uint64(x1[0]))
}

// NewElement returns an element of the field f corresponding to the natural number x.
func (f Field) NewElement(x uint32) Element {
	return Element{uint32(uint64(x) << 32 % uint64(f.modulus))}
}

// Cmp compares the numerical values of x0 and x1.
func (f Field) Cmp(x0, x1 Element) int {
	return cmp.Compare(f.ToUint32(x0), f.ToUint32(x1))
}

// Double x -> 2x
func (f Field) Double(x Element) Element {
	return f.Add(x, x)
}

// rSq returns R² (mod m), NOT IN MONTGOMERY FORM.
func (f Field) rSq() Element {
	// TODO decide whether or not to precompute this.
	// If we do inversions only rarely and in batched,
	// it may be worth it to absorb the penalty and keep
	// the entire field object fitting on one 64 bit word.
	// the largest exponent for 2 for which Montgomery reduction works.
	exponent := uint64(63 - bits.LeadingZeros32(f.modulus))

	x := Element{uint32((1 << exponent) % uint64(f.modulus))}
	for exponent < 64 {
		x = f.Double(x)
		exponent++
	}

	return x
}

// Half x -> x/2 (mod m).
func (f Field) Half(x Element) Element {
	if x[0]%2 == 0 {
		return Element{x[0] / 2}
	} else {
		return Element{(x[0] + f.modulus) / 2} // the modulus is less than 2³¹ so this is safe.
	}
}

// Inverse x -> x⁻¹ (mod m) or 0 if x = 0
func (f Field) Inverse(x Element) Element {
	// Since x actually contains x.R, we have to multiply the result by R² to get x⁻¹R⁻¹R² = x⁻¹R.
	return f.inverse(x, f.rSq())
}

// inverse x -> bias.x⁻¹ (mod m) in non-Montgomery form.
func (f Field) inverse(x, bias Element) Element {
	// Algorithm 16 in "Efficient Software-Implementation of Finite Fields with Applications to Cryptography"
	if x[0] == 0 {
		return Element{0}
	}

	u := x[0]
	v := f.modulus

	var c Element

	b := bias

	for (u != 1) && (v != 1) {
		for u%2 == 0 {
			u /= 2
			b = f.Half(b)
		}

		for v%2 == 0 {
			v /= 2
			c = f.Half(c)
		}

		if diff, borrow := bits.Sub32(u, v, 0); borrow == 0 {
			u = diff
			b = f.Sub(b, c)
		} else {
			v -= u
			c = f.Sub(c, b)
		}
	}

	if u == 1 {
		return b
	} else {
		return c
	}
}
