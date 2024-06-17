package util

// EMPTY_BOUND is the bound which overlaps exactly with the original range.  It
// represents the maximum possible bound.
var EMPTY_BOUND Bounds = Bounds{0, 0}

// Bounds captures the subrange of rows for which a computation is well-defined.
type Bounds struct {
	// Number of rows from first row where computation starts being defined.
	Start uint
	// Number of rows before last row where computation is no longer defined.
	End uint
}

// NewBounds constructs a new set of bounds.
func NewBounds(start uint, end uint) Bounds {
	return Bounds{start, end}
}

// Union merges one set of bounds into another.
func (p *Bounds) Union(q *Bounds) {
	p.Start = max(p.Start, q.Start)
	p.End = max(p.End, q.End)
}

// Boundable captures computations which are well-defined only for a specific
// subrange of rows (the bounds).
type Boundable interface {
	// Determine the well-definedness bounds for this expression for both the
	// negative (left) or positive (right) directions.  For example, consider an
	// expression such as "(shift X -1)".  This is technically undefined for the
	// first row of any trace and, by association, any constraint evaluating
	// this expression on that first row is also undefined (and hence must pass).
	Bounds() Bounds
}

// BoundsForArray determines the bounds for an array of expressions.
func BoundsForArray[E Boundable](args []E) Bounds {
	bounds := Bounds{0, 0}

	for _, e := range args {
		ith := e.Bounds()
		bounds.Union(&ith)
	}
	// Done
	return bounds
}
