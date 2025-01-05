package poly

import (
	"bytes"
	"fmt"
	"math/big"
)

var _ Polynomial[string, *ArrayTerm[string]] = &ArrayPoly[string]{}

// ArrayPoly is the simpliest (and least efficient) polynomial implementation.
// It provides a reference against which other (more efficient) implementations
// can be compared.
type ArrayPoly[S comparable] struct {
	terms []ArrayTerm[S]
}

// NewArrayPoly constructs a new array polynomial from a given initial term.
func NewArrayPoly[S comparable](term *ArrayTerm[S]) *ArrayPoly[S] {
	terms := []ArrayTerm[S]{*term}
	return &ArrayPoly[S]{terms}
}

// Len returns the number of terms in this polynomial.
func (p *ArrayPoly[S]) Len() uint {
	return uint(len(p.terms))
}

// Term returns the ith term in this polynomial.
func (p *ArrayPoly[S]) Term(ith uint) *ArrayTerm[S] {
	return &p.terms[ith]
}

// IsZero returns an indication as to whether this polynomial is equivalent
// to zero (or not).  This is a three valued logic system which can return
// either "yes", "no" or "maybe" where: (i) "yes" means the polynomial
// always evaluates to zero; (ii) "no" means the polynomial never evaluates
// to zero; (iii) "maybe" indicates the polynomial may sometimes evaluate to
// zero.  When the return ok holds then res indicates either yes or not.
// Otherwise, the result is maybe.
func (p *ArrayPoly[S]) IsZero() (res bool, ok bool) {
	panic("todo")
}

// Neg this polynomial, which is equivalent to multiplying each term by -1.
func (p *ArrayPoly[S]) Neg() {
	m_one := big.NewInt(-1)

	for i := 0; i < len(p.terms); i++ {
		p.terms[i].Mul(m_one)
	}
}

// Add another polynomial onto this polynomial.
func (p *ArrayPoly[S]) Add(other Polynomial[S, *ArrayTerm[S]]) {
	for i := uint(0); i < other.Len(); i++ {
		p.AddTerm(other.Term(i))
	}
}

// Sub another polynomial from this polynomil
func (p *ArrayPoly[S]) Sub(other Polynomial[S, *ArrayTerm[S]]) {
	for i := uint(0); i < other.Len(); i++ {
		p.SubTerm(other.Term(i))
	}
}

// Mul this polynomial by another polynomial.
func (p *ArrayPoly[S]) Mul(other Polynomial[S, *ArrayTerm[S]]) {
	panic("todo")
}

// AddTerm adds a single term into this polynomial.
func (p *ArrayPoly[S]) AddTerm(other *ArrayTerm[S]) {
	for _, term := range p.terms {
		if term.Matches(other) {
			coeff := other.Coefficient()
			// Add term at this position
			term.Add(&coeff)
			// Check whether its now zero (or not)
			if term.IsZero() {
				// Yes zero, so remove this term.
				panic("todo")
			}
			//
			return
		}
	}
	// Append to end
	p.terms = append(p.terms, *other.Clone())
	// Sort?
}

// SubTerm subtracts a single term into this polynomial.
func (p *ArrayPoly[S]) SubTerm(other *ArrayTerm[S]) {
	for _, term := range p.terms {
		if term.Matches(other) {
			coeff := other.Coefficient()
			// Subtract term at this position
			term.Sub(&coeff)
			// Check whether its now zero (or not)
			if term.IsZero() {
				// Yes zero, so remove this term.
				panic("todo")
			}
			//
			return
		}
	}
	// Clone & negate
	other = other.Clone()
	// Negate
	other.Mul(big.NewInt(-1))
	// Append to end
	p.terms = append(p.terms, *other)
	// Sort?
}

func (p *ArrayPoly[S]) String() string {
	var buf bytes.Buffer
	//
	for i := 0; i < len(p.terms); i++ {
		ith := p.terms[i]
		coeff := ith.Coefficient()
		//
		if i != 0 {
			buf.WriteString("+")
		}
		// Various cases to improve readability
		if ith.Len() == 0 {
			buf.WriteString(coeff.String())
		} else if coeff.Cmp(big.NewInt(1)) != 0 {
			buf.WriteString("(")
			buf.WriteString(coeff.String())
			//
			for j := uint(0); j < ith.Len(); j++ {
				buf.WriteString(fmt.Sprintf("*%v", ith.Nth(j)))
			}
			//
			buf.WriteString(")")
		} else if ith.Len() == 1 {
			buf.WriteString(fmt.Sprintf("%v", ith.Nth(0)))
		} else {
			buf.WriteString("(")
			//
			for j := uint(0); j < ith.Len(); j++ {
				if i == 0 {
					buf.WriteString("*")
				}
				//
				buf.WriteString(fmt.Sprintf("%v", ith.Nth(j)))
			}
			//
			buf.WriteString(")")
		}
	}
	//
	return buf.String()
}

// ============================================================================
// ArrayTerm
// ============================================================================

// ArrayTerm is the type of terms used within an array polynomial.
type ArrayTerm[S comparable] struct {
	coefficient big.Int
	vars        []S
}

var _ Term[string] = &ArrayTerm[string]{}

// NewArrayTerm constructs a new polynomial term.
func NewArrayTerm[S comparable](coefficient *big.Int, vars []S) *ArrayTerm[S] {
	return &ArrayTerm[S]{*coefficient, vars}
}

// Coefficient returns the coefficient of this term.
func (p *ArrayTerm[S]) Coefficient() big.Int {
	return p.coefficient
}

// Len returns the number of variables in this polynomial term.
//
//nolint:revive
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
//
//nolint:revive
func (p *ArrayTerm[S]) Add(coeff *big.Int) {
	p.coefficient.Add(&p.coefficient, coeff)
}

// Sub updates the cofficient for this term.
//
//nolint:revive
func (p *ArrayTerm[S]) Sub(coeff *big.Int) {
	p.coefficient.Sub(&p.coefficient, coeff)
}

// Mul updates the cofficient for this term.
//
//nolint:revive
func (p *ArrayTerm[S]) Mul(coeff *big.Int) {
	p.coefficient.Mul(&p.coefficient, coeff)
}

// IsZero checks whether the coefficient for this term is zero or not.
//
//nolint:revive
func (p *ArrayTerm[S]) IsZero() bool {
	panic("todo")
}

// Clone creates an identical copy of this term.
func (p *ArrayTerm[S]) Clone() *ArrayTerm[S] {
	var coeff big.Int
	// Clone coefficient
	coeff.Set(&p.coefficient)
	// Clone variables
	vars := make([]S, len(p.vars))
	copy(vars, p.vars)
	// Done
	return &ArrayTerm[S]{coeff, vars}
}
