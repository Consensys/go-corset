package air

import (
	"github.com/consensys/gnark-crypto/ecc/bls12-377/fr"
	"github.com/consensys/go-corset/pkg/trace"
)

// EvalAt evaluates a column access at a given row in a trace, which returns the
// value at that row of the column in question or nil is that row is
// out-of-bounds.
func (e *ColumnAccess) EvalAt(k int, tr trace.Trace) fr.Element {
	return tr.Column(e.Column).Get(k + e.Shift)
}

// EvalAt evaluates a constant at a given row in a trace, which simply returns
// that constant.
func (e *Constant) EvalAt(k int, tr trace.Trace) fr.Element {
	return e.Value
}

// EvalAt evaluates a sum at a given row in a trace by first evaluating all of
// its arguments at that row.
func (e *Add) EvalAt(k int, tr trace.Trace) fr.Element {
	// Evaluate first argument
	val := e.Args[0].EvalAt(k, tr)
	// Continue evaluating the rest
	for i := 1; i < len(e.Args); i++ {
		ith := e.Args[i].EvalAt(k, tr)
		val.Add(&val, &ith)
	}
	// Done
	return val
}

// EvalAt evaluates a product at a given row in a trace by first evaluating all of
// its arguments at that row.
func (e *Mul) EvalAt(k int, tr trace.Trace) fr.Element {
	// Evaluate first argument
	val := e.Args[0].EvalAt(k, tr)
	// Continue evaluating the rest
	for i := 1; i < len(e.Args); i++ {
		// Can short-circuit evaluation?
		if val.IsZero() {
			break
		}
		// No
		ith := e.Args[i].EvalAt(k, tr)
		val.Mul(&val, &ith)
	}
	// Done
	return val
}

// EvalAt evaluates a subtraction at a given row in a trace by first evaluating all of
// its arguments at that row.
func (e *Sub) EvalAt(k int, tr trace.Trace) fr.Element {
	// Evaluate first argument
	val := e.Args[0].EvalAt(k, tr)
	// Continue evaluating the rest
	for i := 1; i < len(e.Args); i++ {
		ith := e.Args[i].EvalAt(k, tr)
		val.Sub(&val, &ith)
	}
	// Done
	return val
}
