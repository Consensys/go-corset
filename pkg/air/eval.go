package air

import (
	"github.com/consensys/gnark-crypto/ecc/bls12-377/fr"
	"github.com/consensys/go-corset/pkg/table"
)

// EvalAt evaluates a column access at a given row in a trace, which returns the
// value at that row of the column in question or nil is that row is
// out-of-bounds.
func (e *ColumnAccess) EvalAt(k int, tbl table.Trace) *fr.Element {
	val := tbl.GetByName(e.Column, k+e.Shift)

	var clone fr.Element
	// Clone original value
	return clone.Set(val)
}

// EvalAt evaluates a constant at a given row in a trace, which simply returns
// that constant.
func (e *Constant) EvalAt(k int, tbl table.Trace) *fr.Element {
	var clone fr.Element
	// Clone original value
	return clone.Set(e.Value)
}

// EvalAt evaluates a sum at a given row in a trace by first evaluating all of
// its arguments at that row.
func (e *Add) EvalAt(k int, tbl table.Trace) *fr.Element {
	fn := func(l *fr.Element, r *fr.Element) { l.Add(l, r) }
	return evalExprsAt(k, tbl, e.Args, fn)
}

// EvalAt evaluates a product at a given row in a trace by first evaluating all of
// its arguments at that row.
func (e *Mul) EvalAt(k int, tbl table.Trace) *fr.Element {
	fn := func(l *fr.Element, r *fr.Element) { l.Mul(l, r) }
	return evalExprsAt(k, tbl, e.Args, fn)
}

// EvalAt evaluates a subtraction at a given row in a trace by first evaluating all of
// its arguments at that row.
func (e *Sub) EvalAt(k int, tbl table.Trace) *fr.Element {
	fn := func(l *fr.Element, r *fr.Element) { l.Sub(l, r) }
	return evalExprsAt(k, tbl, e.Args, fn)
}

// EvalExprsAt evaluates all expressions in a given slice at a given row on the
// table, and fold their results together using a combinator.
func evalExprsAt(k int, tbl table.Trace, exprs []Expr, fn func(*fr.Element, *fr.Element)) *fr.Element {
	// Evaluate first argument
	val := exprs[0].EvalAt(k, tbl)
	if val == nil {
		return nil
	}

	// Continue evaluating the rest
	for i := 1; i < len(exprs); i++ {
		ith := exprs[i].EvalAt(k, tbl)
		fn(val, ith)
	}

	// Done
	return val
}
