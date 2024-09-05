package mir

import (
	"github.com/consensys/gnark-crypto/ecc/bls12-377/fr"
	"github.com/consensys/go-corset/pkg/trace"
	"github.com/consensys/go-corset/pkg/util"
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

	return val
}

// EvalAt evaluates a product at a given row in a trace by first evaluating all of
// its arguments at that row.
func (e *Mul) EvalAt(k int, tr trace.Trace) fr.Element {
	// Evaluate first argument
	val := e.Args[0].EvalAt(k, tr)
	// Continue evaluating the rest
	for i := 1; i < len(e.Args); i++ {
		ith := e.Args[i].EvalAt(k, tr)
		val.Mul(&val, &ith)
	}

	return val
}

// EvalAt evaluates a product at a given row in a trace by first evaluating all of
// its arguments at that row.
func (e *Exp) EvalAt(k int, tr trace.Trace) fr.Element {
	// Check whether argument evaluates to zero or not.
	val := e.Arg.EvalAt(k, tr)
	// Compute exponent
	util.Pow(&val, e.Pow)
	// Done
	return val
}

// EvalAt evaluates the normalisation of some expression by first evaluating
// that expression.  Then, zero is returned if the result is zero; otherwise one
// is returned.
func (e *Normalise) EvalAt(k int, tr trace.Trace) fr.Element {
	// Check whether argument evaluates to zero or not.
	val := e.Arg.EvalAt(k, tr)
	// Normalise value (if necessary)
	if !val.IsZero() {
		val.SetOne()
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
