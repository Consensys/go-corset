package poly

import "math/big"

type ArrayTerm[S comparable] struct {
	coefficient big.Int
	vars        []S
}

var _ Term[string] = &ArrayTerm[string]{}

// Coefficient returns the coefficient of this term.
func (p *ArrayTerm[S]) Coefficient() big.Int {
	return p.coefficient
}

// Len returns the number of variables in this polynomial term.
func (p *ArrayTerm[S]) Len() uint {
	return uint(len(p.vars))
}

// Nth returns the nth variable in this polynomial term.
func (p *ArrayTerm[S]) Nth(uint) S {
	panic("todo")
}

// Matches determines whether or not the variables of this term match those
// of the other.
func (p *ArrayTerm[S]) Matches(other Term[S]) bool {
	panic("todo")
}

// Add updates the cofficient for this term.
func (p *ArrayTerm[S]) Add(coeff big.Int) {
	panic("todo")
}

// IsZero checks whether the coefficient for this term is zero or not.
func (p *ArrayTerm[S]) IsZero() bool {
	panic("todo")
}
