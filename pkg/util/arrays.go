package util

// FlatArrayIndexOf_2 returns the ith element of the flattened form of a 2d
// array. Consider the array "[[0,7],[4]]".  Then its flattened form is
// "[0,7,4]" and, for example, the element at index 1 is "7".
func FlatArrayIndexOf_2[A any, B any](index int, as []A, bs []B) any {
	if index < len(as) {
		return as[index]
	}
	// Otherwise bs
	return bs[index-len(as)]
}

// FlatArrayIndexOf_3 returns the ith element of the flattened form of a 2d
// array. Consider the array "[[0,7],[4]]".  Then its flattened form is
// "[0,7,4]" and, for example, the element at index 1 is "7".
func FlatArrayIndexOf_3[A any, B any, C any](index int, as []A, bs []B, cs []C) any {
	if index < len(as) {
		return as[index]
	}

	return FlatArrayIndexOf_2(index-len(as), bs, cs)
}

// FlatArrayIndexOf_4 returns the ith element of the flattened form of a 2d
// array. Consider the array "[[0,7],[4]]".  Then its flattened form is
// "[0,7,4]" and, for example, the element at index 1 is "7".
func FlatArrayIndexOf_4[A any, B any, C any, D any](index int, as []A, bs []B, cs []C, ds []D) any {
	if index < len(as) {
		return as[index]
	}

	return FlatArrayIndexOf_3(index-len(as), bs, cs, ds)
}

// FlatArrayIndexOf_5 returns the ith element of the flattened form of a 2d
// array. Consider the array "[[0,7],[4]]".  Then its flattened form is
// "[0,7,4]" and, for example, the element at index 1 is "7".
func FlatArrayIndexOf_5[A any, B any, C any, D any, E any](index int, as []A, bs []B, cs []C, ds []D, es []E) any {
	if index < len(as) {
		return as[index]
	}

	return FlatArrayIndexOf_4(index-len(as), bs, cs, ds, es)
}

// FlatArrayIndexOf_6 returns the ith element of the flattened form of a 2d
// array. Consider the array "[[0,7],[4]]".  Then its flattened form is
// "[0,7,4]" and, for example, the element at index 1 is "7".
func FlatArrayIndexOf_6[A any, B any, C any, D any, E any, F any](index int, as []A, bs []B, cs []C, ds []D, es []E,
	fs []F) any {
	if index < len(as) {
		return as[index]
	}

	return FlatArrayIndexOf_5(index-len(as), bs, cs, ds, es, fs)
}
