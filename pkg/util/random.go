package util

import "math/rand/v2"

// GenerateRandomInputs generates n random inputs in the range 0..m.
func GenerateRandomInputs(n, m uint) []uint {
	items := make([]uint, n)

	for i := uint(0); i < n; i++ {
		items[i] = rand.UintN(m)
	}

	return items
}
