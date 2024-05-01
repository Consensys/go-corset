package hir

import (
	"github.com/consensys/gnark-crypto/ecc/bls12-377/fr"
	"github.com/consensys/go-corset/pkg/table"
)

// EvalAt evaluates a column access at a given row in a trace, which returns the
// value at that row of the column in question or nil is that row is
// out-of-bounds.
func (e *ColumnAccess) EvalAt(k int, tbl table.Trace) *fr.Element {
	val, err := tbl.GetByName(e.Column, k+e.Shift)
	// We can ignore err as val is always nil when err != nil.
	// Furthermore, as stated in the documentation for this
	// method, we return nil upon error.
	if err != nil || val == nil {
		// Indicates an out-of-bounds access of some kind.
		return nil
	}

	var clone fr.Element
	// Clone original value
	return clone.Set(val)
}

// EvalAt evaluates a constant at a given row in a trace, which simply returns
// that constant.
func (e *Constant) EvalAt(k int, tbl table.Trace) *fr.Element {
	var clone fr.Element
	// Clone original value
	return clone.Set(e.Val)
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

// EvalAt evaluates a conditional at a given row in a trace by first evaluating
// its condition at that row.  If that condition is zero then the true branch
// (if applicable) is evaluated; otherwise if the condition is non-zero then
// false branch (if applicable) is evaluated).  If the branch to be evaluated is
// missing (i.e. nil), then nil is returned.
func (e *IfZero) EvalAt(k int, tbl table.Trace) *fr.Element {
	// Evaluate condition
	cond := e.Condition.EvalAt(k, tbl)
	// Check whether zero or not
	if cond.IsZero() && e.TrueBranch != nil {
		return e.TrueBranch.EvalAt(k, tbl)
	} else if !cond.IsZero() && e.FalseBranch != nil {
		return e.FalseBranch.EvalAt(k, tbl)
	}

	// If either true / false branch undefined.
	return nil
}

// EvalAt evaluates a list at a given row in a trace by evaluating each of its
// arguments at that row.
func (e *List) EvalAt(k int, tbl table.Trace) *fr.Element {
	panic("Implement hir.List.EvalAt()")
}

// EvalAt evaluates the normalisation of some expression by first evaluating
// that expression.  Then, zero is returned if the result is zero; otherwise one
// is returned.
func (e *Normalise) EvalAt(k int, tbl table.Trace) *fr.Element {
	// Check whether argument evaluates to zero or not.
	val := e.Arg.EvalAt(k, tbl)
	// Normalise value (if necessary)
	if !val.IsZero() {
		val.SetOne()
	}

	return val
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
	// Evaluate first argument.
	val := exprs[0].EvalAt(k, tbl)
	if val == nil {
		return nil
	}

	// Continue evaluating the rest.
	for i := 1; i < len(exprs); i++ {
		ith := exprs[i].EvalAt(k, tbl)
		if ith == nil {
			return ith
		}

		fn(val, ith)
	}

	// Done.
	return val
}
