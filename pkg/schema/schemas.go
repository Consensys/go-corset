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
	index := schema.InputColumns().Count()
	m := schema.Assignments().Count()
	batchjobs := make([]expandTraceJob, m)
	// Compute each assignment in turn
	for i, j := schema.Assignments(), uint(0); i.HasNext(); j++ {
		// Get ith assignment
		ith := i.Next()
		// Compute ith assignment(s)
		batchjobs[j] = expandTraceJob{index, ith, trace}
		// Update index
		index += ith.Columns().Count()
	}
	//
	return util.ParExec[expandTraceJob](batchjobs)
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

// ----------------------------------------------------------------------------

// ExpandTraceJob represents a unit of work which can be parallelised during
// trace expansion. N Specifically, the unit of work is a single assignment.  In
// the terminology of ParExec, an assignment is a batch of jobs, each of which
// assigns values to a given column.
type expandTraceJob struct {
	// Index of first column in the assignment
	index uint
	// Assignment itself being computed
	assignment Assignment
	// Trace being expanded
	trace tr.Trace
}

// Jobs returns the underlying jobs that this "trace job" computes.
// Specifically, it identifies the columns that the trace job completes.
func (p expandTraceJob) Jobs() []uint {
	n := p.assignment.Columns().Count()
	cols := make([]uint, n)
	// TODO: this is really a hack for now.  I think we should use some kind of
	// range iterator.
	for i := uint(0); i < n; i++ {
		cols[i] = p.index + i
	}

	return cols
}

// Dependencies returns the columns that this batch job depends upon.  That is
// the set of source columns (if any) required before this job can be completed.
func (p expandTraceJob) Dependencies() []uint {
	return p.assignment.Dependencies()
}

// Run computes the coumns values for a given assignment.
func (p expandTraceJob) Run() error {
	n := p.trace.Columns().Len()
	if n != p.index {
		panic(fmt.Sprintf("internal failure (%d trace columns, versus index %d)", n, p.index))
	}
	//
	return p.assignment.ExpandTrace(p.trace)
}
