package trace

// Constraint represents an abstract notion of a constraint which must hold true for a given
// table.
type Constraint interface {
	// GetHandle gets the handle for this constraint (i.e. its name).
	GetHandle() string
	// Check checks whether this constraint holds on a particular
	// table.
	Check(Table) error
}
