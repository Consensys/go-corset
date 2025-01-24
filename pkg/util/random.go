package util

import "math/rand/v2"

// GenerateRandomUints generates n random unsigned integers in the range 0..m.
func GenerateRandomUints(n, m uint) []uint {
	items := make([]uint, n)

	for i := uint(0); i < n; i++ {
		items[i] = rand.UintN(m)
	}

	return items
}

// GenerateRandomInts generates n random unsigned integers in the range -m..m.
func GenerateRandomInts(n uint, m int) []int {
	items := make([]int, n)

	for i := uint(0); i < n; i++ {
		items[i] = rand.IntN(2*m) - m
	}

	return items
}

// GenerateRandomElements generates n elements selected at random from the given array.
func GenerateRandomElements[E any](n uint, elems []E) []E {
	items := make([]E, n)
	m := uint(len(elems))

	for i := uint(0); i < n; i++ {
		index := rand.UintN(m)
		items[i] = elems[index]
	}

	return items
}
