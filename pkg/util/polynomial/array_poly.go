package polynomial

// ArrayPoly is the simpliest (and least efficient) polynomial implementation.
// It provides a reference against which other (more efficient) implementations
// can be compared.
type ArrayPoly[S comparable, T Term[S]] struct {
	terms []T
}

// NewArrayPoly constructs a new array polynomial from a given initial term.
func NewArrayPoly[S comparable, T Term[S]](term T) *ArrayPoly[S, T] {
	return &ArrayPoly[S, T]{[]T{term}}
}

// Len returns the number of terms in this polynomial.
func (p *ArrayPoly[S, T]) Len() uint {
	return uint(len(p.terms))
}

// Term returns the ith term in this polynomial.
func (p *ArrayPoly[S, T]) Term(ith uint) *T {
	return &p.terms[ith]
}

// IsZero returns an indication as to whether this polynomial is equivalent
// to zero (or not).  This is a three valued logic system which can return
// either "yes", "no" or "maybe" where: (i) "yes" means the polynomial
// always evaluates to zero; (ii) "no" means the polynomial never evaluates
// to zero; (iii) "maybe" indicates the polynomial may sometimes evaluate to
// zero.  When the return ok holds then res indicates either yes or not.
// Otherwise, the result is maybe.
func (p *ArrayPoly[S, T]) IsZero() (res bool, ok bool) {
	panic("todo")
}

// Add another polynomial onto this polynomial.
func (p *ArrayPoly[S, T]) Add(other Polynomial[S, T]) {
	for i := uint(0); i < other.Len(); i++ {
		p.AddTerm(other.Term(i))
	}
}

// AddTerm adds a single term into this polynomial.
func (p *ArrayPoly[S, T]) AddTerm(other T) {
	for _, term := range p.terms {
		if term.Matches(other) {
			// Add term at this position
			term.Add(other.Coefficient())
			// Check whether its now zero (or not)
			if term.IsZero() {
				// Yes zero, so remove this term.
				panic("todo")
			}
		}
	}
}
