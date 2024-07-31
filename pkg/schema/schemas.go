package schema

import (
	tr "github.com/consensys/go-corset/pkg/trace"
)

// JoinContexts combines one or more evaluation contexts together.  If all
// expressions have the void context, then this is returned.  Likewise, if any
// expression has a conflicting context then this is returned.  Finally, if any
// two expressions have conflicting contexts between them, then the conflicting
// context is returned.  Otherwise, the common context to all expressions is
// returned.
func JoinContexts[E Contextual](args []E, schema Schema) tr.Context {
	ctx := tr.VoidContext()
	//
	for _, e := range args {
		ctx = ctx.Join(e.Context(schema))
	}
	// If we get here, then no conflicts were detected.
	return ctx
}

// ContextOfColumns determines the enclosing context for a given set of columns.
// If all columns have the void context, then this is returned.  Likewise,
// if any column has a conflicting context then this is returned.  Finally,
// if any two columns have conflicting contexts between them, then the
// conflicting context is returned.  Otherwise, the common context to all
// columns is returned.
func ContextOfColumns(cols []uint, schema Schema) tr.Context {
	ctx := tr.VoidContext()
	//
	for i := 0; i < len(cols); i++ {
		col := schema.Columns().Nth(cols[i])
		ctx = ctx.Join(col.Context())
	}
	// Done
	return ctx
}

// Accepts determines whether this schema will accept a given trace.  That
// is, whether or not the given trace adheres to the schema.  A trace can fail
// to adhere to the schema for a variety of reasons, such as having a constraint
// which does not hold.
//
//nolint:revive
func Accepts(schema Schema, trace tr.Trace) error {
	var err error
	// Determine how many constraints
	n := schema.Constraints().Count()
	// Construct a communication channel for errors.
	c := make(chan error, 100)
	// Check each constraint in turn
	for i := schema.Constraints(); i.HasNext(); {
		// Get ith constraint
		ith := i.Next()
		// Launch checker for constraint
		go func() {
			// Send outcome back
			c <- ith.Accepts(trace)
		}()
	}
	// Read responses back from each constraint.
	for i := uint(0); i < n; i++ {
		// Read from channel
		if e := <-c; e != nil {
			err = e
		}
	}
	// Success
	return err
}

// ColumnIndexOf returns the column index of the column with the given name, or
// returns false if no matching column exists.
func ColumnIndexOf(schema Schema, module uint, name string) (uint, bool) {
	return schema.Columns().Find(func(c Column) bool {
		return c.Context().Module() == module && c.Name() == name
	})
}
