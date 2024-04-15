package trace

// An abstract notion of a constraint which must hold true for a given
// table.
type Constraint interface {
	// Get the handle for this constraint (i.e. its name).
	GetHandle() string
	// Check whether or not this constraint holds on a particular
	// table.
	Check(Table) error
}
