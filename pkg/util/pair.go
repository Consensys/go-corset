package util

// Pair provides a simple encapsulation of two items paired together.
type Pair[S any, T any] struct {
	Left  S
	Right T
}

// NewPair returns a new instance of Pair by value.
func NewPair[S any, T any](left S, right T) Pair[S, T] {
	return Pair[S, T]{left, right}
}

// NewPairRef returns a reference to a new instance of Pair.
func NewPairRef[S any, T any](left S, right T) *Pair[S, T] {
	var p Pair[S, T] = NewPair(left, right)
	return &p
}
