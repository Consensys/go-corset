package schema

import (
	"fmt"

	tr "github.com/consensys/go-corset/pkg/trace"
)

// RequiredSpillage returns the minimum amount of spillage required to ensure
// valid traces are accepted in the presence of arbitrary padding.  Spillage can
// only arise from computations as this is where values outside of the user's
// control are determined.
func RequiredSpillage(schema Schema) uint {
	// Ensures always at least one row of spillage (referred to as the "initial
	// padding row")
	mx := uint(1)
	// Determine if any more spillage required
	for i := schema.Assignments(); i.HasNext(); {
		// Get ith assignment
		ith := i.Next()
		// Incorporate its spillage requirements
		mx = max(mx, ith.RequiredSpillage())
	}

	return mx
}

// ExpandTrace expands a given trace according to this schema.  More
// specifically, that means computing the actual values for any assignments.
// Observe that assignments have to be computed in the correct order.
func ExpandTrace(schema Schema, trace tr.Trace) error {
	// Compute each assignment in turn
	for i := schema.Assignments(); i.HasNext(); {
		// Get ith assignment
		ith := i.Next()
		// Compute ith assignment(s)
		if err := ith.ExpandTrace(trace); err != nil {
			return err
		}
	}
	// Done
	return nil
}

// Accepts determines whether this schema will accept a given trace.  That
// is, whether or not the given trace adheres to the schema.  A trace can fail
// to adhere to the schema for a variety of reasons, such as having a constraint
// which does not hold.
//
//nolint:revive
func Accepts(schema Schema, trace tr.Trace) error {
	// Check each constraint in turn
	for i := schema.Constraints(); i.HasNext(); {
		// Get ith constraint
		ith := i.Next()
		// Check it holds (or report an error)
		if err := ith.Accepts(trace); err != nil {
			return err
		}
	}
	// Success
	return nil
}

// ColumnIndexOf returns the column index of the column with the given name, or
// returns false if no matching column exists.
func ColumnIndexOf(schema Schema, name string) (uint, bool) {
	return schema.Columns().Find(func(c Column) bool {
		return c.Name() == name
	})
}

// ColumnByName returns the column with the matching name, or panics if no such
// column exists.
func ColumnByName(schema Schema, name string) Column {
	var col Column
	// Attempt to determine the index of this column
	_, ok := schema.Columns().Find(func(c Column) bool {
		col = c
		return c.Name() == name
	})
	// If we found it, then done.
	if ok {
		return col
	}
	// Otherwise panic.
	panic(fmt.Sprintf("unknown column %s", name))
}

// HasColumn checks whether a column of the given name is declared within the schema.
func HasColumn(schema Schema, name string) bool {
	_, ok := ColumnIndexOf(schema, name)
	return ok
}
