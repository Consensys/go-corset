package poly

import "math/big"

// Polynomial represents a sum of terms of a type T of variables.
type Polynomial[S comparable, T Term[S]] interface {
	// Len returns the number of terms in this polynomial.
	Len() uint

	// Term returns the ith term in this polynomial.
	Term(uint) T

	// IsZero returns an indication as to whether this polynomial is equivalent
	// to zero (or not).  This is a three valued logic system which can return
	// either "yes", "no" or "maybe" where: (i) "yes" means the polynomial
	// always evaluates to zero; (ii) "no" means the polynomial never evaluates
	// to zero; (iii) "maybe" indicates the polynomial may sometimes evaluate to
	// zero.  When the return ok holds then res indicates either yes or not.
	// Otherwise, the result is maybe.
	IsZero() (res bool, ok bool)

	// Add another polynomial onto this polynomial, such that this polynomial is
	// updated in place.
	Add(Polynomial[S, T])

	// Subtract another polynomial from this polynomial, such that this
	// polynomial is updated in place.
	Sub(Polynomial[S, T])

	// Multiply this polynomial by another polynomial, such that this polynomial
	// is updated in place.
	Mul(Polynomial[S, T])
}

func Eval[S comparable, T Term[S]](poly Polynomial[S, T], env map[S]big.Int) *big.Int {
	val := big.NewInt(0)
	// Sum evaluated terms
	for i := uint(0); i < poly.Len(); i++ {
		val.Add(val, EvalTerm(poly.Term(i), env))
	}
	// Done
	return val
}

func EvalTerm[S comparable, T Term[S]](term T, env map[S]big.Int) *big.Int {
	var (
		acc   big.Int
		coeff big.Int = term.Coefficient()
	)
	// Initialise accumulator
	acc.Set(&coeff)
	//
	for j := uint(0); j < term.Len(); j++ {
		jth := env[term.Nth(j)]
		acc.Mul(&acc, &jth)
	}
	//
	return &acc
}
