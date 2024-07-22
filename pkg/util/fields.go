package util

import (
	"encoding/binary"

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

// FrElementToBytes converts a given field element into a slice of 32 bytes.
func FrElementToBytes(element *fr.Element) [32]byte {
	// Each fr.Element is 4 x 64bit words.
	var bytes [32]byte
	// Copy over each element
	binary.BigEndian.PutUint64(bytes[:], element[0])
	binary.BigEndian.PutUint64(bytes[8:], element[1])
	binary.BigEndian.PutUint64(bytes[16:], element[2])
	binary.BigEndian.PutUint64(bytes[24:], element[3])
	// Done
	return bytes
}
