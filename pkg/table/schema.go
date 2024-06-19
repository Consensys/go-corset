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

	// Determine the number of column groups in this schema.
	Width() uint

	// Access information about the ith column group in this schema.
	ColumnGroup(uint) ColumnGroup
}

// ColumnGroup represents a group of related columns in the schema.  For
// example, a single data column is (for now) always a column group of size 1.
// Likewise, an array of size n is a column group of size n, etc.
type ColumnGroup interface {
	// Return the number of columns in this group.
	Width() uint

	// Returns the name of the ith column in this group.
	NameOf(uint) string

	// Determines whether or not this column group is synthetic.
	IsSynthetic() bool
}

// Declaration represents a declared element of a schema.  For example, a column
// declaration or a vanishing constraint declaration.  The purpose of this
// interface is to provide some generic interactions that are available
// regardless of the IR level.
type Declaration interface {
	// Return a human-readable string for this declaration.
	String() string
}
