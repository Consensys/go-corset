package constraint

// Represents the set of rows for which a given constraint is active.
type Domain interface {
	// Get an iterator over the values of this domain.
	Iterator() func() (int,bool)

	// Check whether or not this domain contains a given row.
	Contains(int) bool
}

// A domain which contains all rows in the range start..end
// (exclusive).
type RangeDomain struct {
	// Start of range
	start int
	// end of range
	end int
}

func (r RangeDomain) Iterator() func() (int,bool) {
	// Not sure about the implementation here.
	panic("imeplement me")
}
