package table

type Constraint interface {
	// Get the handle for this constraint (i.e. its name).
	GetHandle() string
}
