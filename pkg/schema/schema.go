package schema

import tr "github.com/consensys/go-corset/pkg/trace"

// Schema represents a schema which can be used to manipulate a trace.
// Specifically, a schema can determine whether or not a trace is accepted;
// likewise, a schema can expand a trace according to its internal computation.
type Schema interface {
	Accepts(tr.Trace) error
	// Expandtr.Trace expands a given trace to include "computed
	// columns".  These are columns which do not exist in the
	// original trace, but are added during trace expansion to
	// form the final trace.
	ExpandTrace(tr.Trace) error

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

	// Determine the index of a named column in this schema, or return false if
	// no matching column exists.
	ColumnIndex(string) (uint, bool)

	// Access information about the ith column group in this schema.
	ColumnGroup(uint) ColumnGroup

	// Access information about the ith column in this schema.
	Column(uint) ColumnSchema
}

// Assignment represents a schema element which declares one or more columns
// whose values are "assigned" from the results of a computation.  An assignment
// is a column group which, additionally, can provide information about the
// computation (e.g. which columns it depends upon, etc).
type Assignment interface {
	ColumnGroup
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

// ColumnSchema provides information about a specific column in the schema.
type ColumnSchema interface {
	// Returns the name of this column
	Name() string
}

// Acceptable represents an element which can "accept" a trace, or either reject
// with an error (or eventually perhaps report a warning).
type Acceptable interface {
	Accepts(tr.Trace) error
}

// Declaration represents a declared element of a schema.  For example, a column
// declaration or a vanishing constraint declaration.  The purpose of this
// interface is to provide some generic interactions that are available
// regardless of the IR level.
type Declaration interface {
	// Return a human-readable string for this declaration.
	String() string
}

// ConstraintsAcceptTrace determines whether or not one or more groups of
// constraints accept a given trace.  It returns the first error or warning
// encountered.
func ConstraintsAcceptTrace[T Acceptable](trace tr.Trace, constraints []T) error {
	for _, c := range constraints {
		err := c.Accepts(trace)
		if err != nil {
			return err
		}
	}
	//
	return nil
}
