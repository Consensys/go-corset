package util

import (
	"github.com/consensys/gnark-crypto/ecc/bls12-377/fr"
)

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
