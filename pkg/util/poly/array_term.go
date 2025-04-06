package poly

import "math/big"

type ArrayTerm[S comparable] struct {
	coefficient big.Int
	vars        []S
}

var _ Term[string] = &ArrayTerm[string]{}

// NewArrayTerm constructs a new polynomial term.
func NewArrayTerm[S comparable](coefficient big.Int, vars []S) *ArrayTerm[S] {
	return &ArrayTerm[S]{coefficient, vars}
}

// Coefficient returns the coefficient of this term.
func (p *ArrayTerm[S]) Coefficient() big.Int {
	return p.coefficient
}

// Len returns the number of variables in this polynomial term.
func (p *ArrayTerm[S]) Len() uint {
	return uint(len(p.vars))
}

// Nth returns the nth variable in this polynomial term.
func (p *ArrayTerm[S]) Nth(index uint) S {
	return p.vars[index]
}

// Matches determines whether or not the variables of this term match those
// of the other.
func (p *ArrayTerm[S]) Matches(other Term[S]) bool {
	if p.Len() != other.Len() {
		return false
	}
	//
	for i := uint(0); i < p.Len(); i++ {
		if p.vars[i] != other.Nth(i) {
			return false
		}
	}
	//
	return true
}

// Add updates the cofficient for this term.
func (p *ArrayTerm[S]) Add(coeff big.Int) {
	p.coefficient.Add(&p.coefficient, &coeff)
}

// Sub updates the cofficient for this term.
func (p *ArrayTerm[S]) Sub(coeff big.Int) {
	p.coefficient.Sub(&p.coefficient, &coeff)
}

// Neg negates the coefficient for this term.
func (p *ArrayTerm[S]) Neg() {
	p.coefficient.Neg(&p.coefficient)
}

// IsZero checks whether the coefficient for this term is zero or not.
func (p *ArrayTerm[S]) IsZero() bool {
	var zero = big.NewInt(0)
	return p.coefficient.Cmp(zero) == 0
}
