package hir

import (
	"github.com/consensys/gnark-crypto/ecc/bls12-377/fr"
	"github.com/consensys/go-corset/pkg/table"
)

// EvalAllAt evaluates a column access at a given row in a trace, which returns the
// value at that row of the column in question or nil is that row is
// out-of-bounds.
func (e *ColumnAccess) EvalAllAt(k int, tbl table.Trace) []*fr.Element {
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
	return []*fr.Element{clone.Set(val)}
}

// EvalAllAt evaluates a constant at a given row in a trace, which simply returns
// that constant.
func (e *Constant) EvalAllAt(k int, tbl table.Trace) []*fr.Element {
	var clone fr.Element
	// Clone original value
	return []*fr.Element{clone.Set(e.Val)}
}

// EvalAllAt evaluates a sum at a given row in a trace by first evaluating all of
// its arguments at that row.
func (e *Add) EvalAllAt(k int, tbl table.Trace) []*fr.Element {
	fn := func(l *fr.Element, r *fr.Element) { l.Add(l, r) }
	return evalExprsAt(k, tbl, e.Args, fn)
}

// EvalAllAt evaluates a product at a given row in a trace by first evaluating all of
// its arguments at that row.
func (e *Mul) EvalAllAt(k int, tbl table.Trace) []*fr.Element {
	fn := func(l *fr.Element, r *fr.Element) { l.Mul(l, r) }
	return evalExprsAt(k, tbl, e.Args, fn)
}

// EvalAllAt evaluates a conditional at a given row in a trace by first evaluating
// its condition at that row.  If that condition is zero then the true branch
// (if applicable) is evaluated; otherwise if the condition is non-zero then
// false branch (if applicable) is evaluated).  If the branch to be evaluated is
// missing (i.e. nil), then nil is returned.
func (e *IfZero) EvalAllAt(k int, tbl table.Trace) []*fr.Element {
	vals := make([]*fr.Element, 0)
	// Evaluate condition
	conditions := e.Condition.EvalAllAt(k, tbl)
	// Check all results
	for _, cond := range conditions {
		if cond.IsZero() && e.TrueBranch != nil {
			vals = append(vals, e.TrueBranch.EvalAllAt(k, tbl)...)
		} else if !cond.IsZero() && e.FalseBranch != nil {
			vals = append(vals, e.FalseBranch.EvalAllAt(k, tbl)...)
		}
	}

	return vals
}

// EvalAllAt evaluates a list at a given row in a trace by evaluating each of its
// arguments at that row.
func (e *List) EvalAllAt(k int, tbl table.Trace) []*fr.Element {
	vals := make([]*fr.Element, 0)

	for _, e := range e.Args {
		vs := e.EvalAllAt(k, tbl)
		vals = append(vals, vs...)
	}

	return vals
}

// EvalAllAt evaluates the normalisation of some expression by first evaluating
// that expression.  Then, zero is returned if the result is zero; otherwise one
// is returned.
func (e *Normalise) EvalAllAt(k int, tbl table.Trace) []*fr.Element {
	// Check whether argument evaluates to zero or not.
	vals := e.Arg.EvalAllAt(k, tbl)
	// Normalise values (as necessary)
	for _, e := range vals {
		if !e.IsZero() {
			e.SetOne()
		}
	}

	return vals
}

// EvalAllAt evaluates a subtraction at a given row in a trace by first evaluating all of
// its arguments at that row.
func (e *Sub) EvalAllAt(k int, tbl table.Trace) []*fr.Element {
	fn := func(l *fr.Element, r *fr.Element) { l.Sub(l, r) }
	return evalExprsAt(k, tbl, e.Args, fn)
}

// EvalExprsAt evaluates all expressions in a given slice at a given row on the
// table, and fold their results together using a combinator.
func evalExprsAt(k int, tbl table.Trace, exprs []Expr, fn func(*fr.Element, *fr.Element)) []*fr.Element {
	// Evaluate first argument.
	vals := exprs[0].EvalAllAt(k, tbl)

	// Continue evaluating the rest.
	for i := 1; i < len(exprs); i++ {
		vs := exprs[i].EvalAllAt(k, tbl)
		vals = evalExprsAtApply(vals, vs, fn)
	}

	// Done.
	return vals
}

// Perform a vector operation using the given primitive operator "fn".
func evalExprsAtApply(lhs []*fr.Element, rhs []*fr.Element, fn func(*fr.Element, *fr.Element)) []*fr.Element {
	if len(rhs) == 1 {
		// Optimise for common case.
		for _, ith := range lhs {
			fn(ith, rhs[0])
		}

		return lhs
	}
	// Harder case
	vals := make([]*fr.Element, 0)
	// Perform n x m operations
	for _, ith := range lhs {
		for _, jth := range rhs {
			var clone fr.Element

			clone.Set(ith)
			fn(&clone, jth)
			vals = append(vals, &clone)
		}
	}

	return vals
}
