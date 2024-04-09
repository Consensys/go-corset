package constraint

// Domain represents the set of rows for which a given constraint is active.
type Domain interface {
	// Iterator get an iterator over the values of this domain.
	Iterator() func() (int, bool)

	// Contains check whether or not this domain contains a given row.
	Contains(int) bool
}

// RangeDomain a domain which contains all rows in the range start..end
// (exclusive).
type RangeDomain struct {
	// Start of range
	start int
	// end of range
	end int
}

func (r *RangeDomain) Iterator() func() (int, bool) {
	// Not sure about the implementation here.
	panic("implement me")
}
