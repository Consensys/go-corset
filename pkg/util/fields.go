package util

import (
	"math/big"

	"github.com/consensys/gnark-crypto/ecc/bls12-377/fr"
)

// ToFieldElements converts an array of big integers into an array of field elements.
func ToFieldElements(ints []*big.Int) []*fr.Element {
	elements := make([]*fr.Element, len(ints))
	// Convert each integer in turn.
	for i, v := range ints {
		element := new(fr.Element)
		element.SetBigInt(v)
		elements[i] = element
	}

	// Done.
	return elements
}

// Pow takes a given value to the power n.
func Pow(val *fr.Element, n uint64) {
	if n == 0 {
		val.SetOne()
	} else if n > 1 {
		m := n / 2
		// Check for odd case
		if n%2 == 1 {
			var tmp fr.Element
			// Clone value
			tmp.Set(val)
			Pow(val, m)
			val.Square(val)
			val.Mul(val, &tmp)
		} else {
			// Even case is easy
			Pow(val, m)
			val.Square(val)
		}
	}
}
