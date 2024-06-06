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
