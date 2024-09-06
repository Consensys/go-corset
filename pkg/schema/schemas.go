package schema

import (
	"fmt"

	tr "github.com/consensys/go-corset/pkg/trace"
	"github.com/consensys/go-corset/pkg/util"
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
func Accepts(batchsize uint, schema Schema, trace tr.Trace) []Failure {
	errors := make([]Failure, 0)
	iter := schema.Constraints()
	// Initialise batch number (for debugging purposes)
	batch := uint(0)
	// Process constraints in batches
	for iter.HasNext() {
		errs := processConstraintBatch(batch, batchsize, iter, trace)
		errors = append(errors, errs...)
		// Increment batch number
		batch++
	}
	// Success
	return errors
}

// Process a given set of constraints in a single batch whilst recording all constraint failures.
func processConstraintBatch(batch uint, batchsize uint, iter util.Iterator[Constraint], trace tr.Trace) []Failure {
	n := uint(0)
	c := make(chan Failure, 1024)
	errors := make([]Failure, 0)
	stats := util.NewPerfStats()
	// Launch at most 100 go-routines.
	for ; n < batchsize && iter.HasNext(); n++ {
		// Get ith constraint
		ith := iter.Next()
		// Launch checker for constraint
		go func() {
			// Send outcome back
			c <- ith.Accepts(trace)
		}()
	}
	//
	for i := uint(0); i < n; i++ {
		// Read from channel
		if e := <-c; e != nil {
			errors = append(errors, e)
		}
	}
	// Log stats about this batch
	stats.Log(fmt.Sprintf("Constraint batch %d", batch))
	//
	return errors
}

// ColumnIndexOf returns the column index of the column with the given name, or
// returns false if no matching column exists.
func ColumnIndexOf(schema Schema, module uint, name string) (uint, bool) {
	return schema.Columns().Find(func(c Column) bool {
		return c.Context().Module() == module && c.Name() == name
	})
}
