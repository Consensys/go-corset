package table

// Schema represents a schema which can be used to manipulate a trace.
// Specifically, a schema can determine whether or not a trace is accepted;
// likewise, a schema can expand a trace according to its internal computation.
type Schema interface {
	Accepts(Trace) error
	// ExpandTrace expands a given trace to include "computed
	// columns".  These are columns which do not exist in the
	// original trace, but are added during trace expansion to
	// form the final trace.
	ExpandTrace(Trace) error

	// Size returns the number of declarations in this schema.
	Size() int

	// GetDeclaration returns the ith declaration in this schema.
	GetDeclaration(int) Declaration

	// RequiredSpillage returns the minimum amount of spillage required to
	// ensure valid traces are accepted in the presence of arbitrary padding.
	// Note: this is calculated on demand.
	RequiredSpillage() uint

	// ApplyPadding adds n items of padding to each column of the trace.
	// Padding values are placed either at the front or the back of a given
	// column, depending on their interpretation.
	ApplyPadding(uint, Trace)
}

// Declaration represents a declared element of a schema.  For example, a column
// declaration or a vanishing constraint declaration.  The purpose of this
// interface is to provide some generic interactions that are available
// regardless of the IR level.
type Declaration interface {
	// Return a human-readable string for this declaration.
	String() string
}
