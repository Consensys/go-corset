package smallfield

import (
	"cmp"
	"math/big"
)

// Element of a prime order field, represented in Montgomery form to speed up multiplications.
type Element [1]uint32 // defined as an array to prevent mistaken use of arithmetic operators, or naive assignments.

// A Field of prime order, less than 2³¹.
type Field struct {
	modulus           uint32
	negModulusInvModR uint32
}

// NewField of the given order.
func NewField(modulus uint32) Field {
	if modulus >= 1<<31 {
		panic("modulus too large") // need at least one bit of "slack"
	}

	m := big.NewInt(int64(modulus))
	m.ModInverse(m, big.NewInt(1<<32))

	return Field{modulus: modulus, negModulusInvModR: uint32(1<<32 - m.Uint64())}
}

// Add x0 + x1 + xRest[0] + xRest[1] + ...
func (f Field) Add(x0, x1 Element, xRest ...Element) Element {
	res := Element{x0[0] + x1[0]}
	if res[0] >= f.modulus {
		res[0] -= f.modulus
	}

	for _, e := range xRest {
		res[0] += e[0]
		if res[0] >= f.modulus {
			res[0] -= f.modulus
		}
	}

	return res
}

// Sub x0 - x1 - xRest[0] - xRest[1] - ...
func (f Field) Sub(x0, x1 Element, xRest ...Element) Element {
	const negMask uint32 = 1 << 31

	res := Element{x0[0] - x1[0]}
	if res[0]&negMask != 0 {
		res[0] += f.modulus
	}

	for _, e := range xRest {
		res[0] -= e[0]
		if res[0]&negMask != 0 {
			res[0] += f.modulus
		}
	}

	return res
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

func (f Field) mul(a, b Element) Element {
	return f.montgomeryReduce(uint64(a[0]) * uint64(b[0]))
}

// Mul x0 * x1 * xRest[0] * xRest[1] * ...
func (f Field) Mul(x0, x1 Element, xRest ...Element) Element {
	res := f.mul(x0, x1)
	for _, e := range xRest {
		res = f.mul(res, e)
	}

	return res
}

// NewElement returns an element of the field f corresponding to the natural number x.
func (f Field) NewElement(x uint32) Element {
	return Element{uint32(uint64(x) << 32 % uint64(f.modulus))}
}

// Cmp compares the numerical values of x0 and x1.
func (f Field) Cmp(x0, x1 Element) int {
	return cmp.Compare(f.ToUint32(x0), f.ToUint32(x1))
}
