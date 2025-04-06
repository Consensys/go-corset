package poly

import "math/big"

// Term represents a product (or monomial) within a polynomial.
type Term[T any] interface {

	// Coefficient returns the coefficient of this term.
	Coefficient() big.Int

	// Len returns the number of variables in this polynomial term.
	Len() uint

	// Nth returns the nth variable in this polynomial term.
	Nth(uint) T

	// Matches determines whether or not the variables of this term match those
	// of the other.
	Matches(other Term[T]) bool

	// Add updates the cofficient for this term.
	Add(coeff big.Int)

	// Sub updates the cofficient for this term.
	Sub(coeff big.Int)

	// Neg negates the coefficient of this term
	Neg()

	// IsZero checks whether the coefficient for this term is zero or not.
	IsZero() bool
}
