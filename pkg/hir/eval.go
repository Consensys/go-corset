package hir

import (
	"github.com/consensys/gnark-crypto/ecc/bls12-377/fr"
	"github.com/consensys/go-corset/pkg/trace"
	"github.com/consensys/go-corset/pkg/util"
)

// EvalAllAt evaluates a column access at a given row in a trace, which returns the
// value at that row of the column in question or nil is that row is
// out-of-bounds.
func (e *ColumnAccess) EvalAllAt(k int, tr trace.Trace) []fr.Element {
	val := tr.Column(e.Column).Get(k + e.Shift)
	// Clone original value
	return []fr.Element{val}
}

// EvalAllAt evaluates a constant at a given row in a trace, which simply returns
// that constant.
func (e *Constant) EvalAllAt(k int, tr trace.Trace) []fr.Element {
	return []fr.Element{e.Val}
}

// EvalAllAt evaluates a sum at a given row in a trace by first evaluating all of
// its arguments at that row.
func (e *Add) EvalAllAt(k int, tr trace.Trace) []fr.Element {
	fn := func(l fr.Element, r fr.Element) fr.Element { l.Add(&l, &r); return l }
	return evalExprsAt(k, tr, e.Args, fn)
}

// EvalAllAt evaluates a product at a given row in a trace by first evaluating all of
// its arguments at that row.
func (e *Mul) EvalAllAt(k int, tr trace.Trace) []fr.Element {
	fn := func(l fr.Element, r fr.Element) fr.Element { l.Mul(&l, &r); return l }
	return evalExprsAt(k, tr, e.Args, fn)
}

// EvalAllAt evaluates a product at a given row in a trace by first evaluating all of
// its arguments at that row.
func (e *Exp) EvalAllAt(k int, tr trace.Trace) []fr.Element {
	vals := e.Arg.EvalAllAt(k, tr)
	for i := range vals {
		util.Pow(&vals[i], e.Pow)
	}
	// Done
	return vals
}

// EvalAllAt evaluates a conditional at a given row in a trace by first evaluating
// its condition at that row.  If that condition is zero then the true branch
// (if applicable) is evaluated; otherwise if the condition is non-zero then
// false branch (if applicable) is evaluated).  If the branch to be evaluated is
// missing (i.e. nil), then nil is returned.
func (e *IfZero) EvalAllAt(k int, tr trace.Trace) []fr.Element {
	vals := make([]fr.Element, 0)
	// Evaluate condition
	conditions := e.Condition.EvalAllAt(k, tr)
	// Check all results
	for _, cond := range conditions {
		if cond.IsZero() && e.TrueBranch != nil {
			vals = append(vals, e.TrueBranch.EvalAllAt(k, tr)...)
		} else if !cond.IsZero() && e.FalseBranch != nil {
			vals = append(vals, e.FalseBranch.EvalAllAt(k, tr)...)
		}
	}

	return vals
}

// EvalAllAt evaluates a list at a given row in a trace by evaluating each of its
// arguments at that row.
func (e *List) EvalAllAt(k int, tr trace.Trace) []fr.Element {
	vals := make([]fr.Element, 0)

	for _, e := range e.Args {
		vs := e.EvalAllAt(k, tr)
		vals = append(vals, vs...)
	}

	return vals
}

// EvalAllAt evaluates the normalisation of some expression by first evaluating
// that expression.  Then, zero is returned if the result is zero; otherwise one
// is returned.
func (e *Normalise) EvalAllAt(k int, tr trace.Trace) []fr.Element {
	// Check whether argument evaluates to zero or not.
	vals := e.Arg.EvalAllAt(k, tr)
	// Normalise values (as necessary)
	for i := range vals {
		if !vals[i].IsZero() {
			vals[i].SetOne()
		}
	}

	return vals
}

// EvalAllAt evaluates a subtraction at a given row in a trace by first evaluating all of
// its arguments at that row.
func (e *Sub) EvalAllAt(k int, tr trace.Trace) []fr.Element {
	fn := func(l fr.Element, r fr.Element) fr.Element { l.Sub(&l, &r); return l }
	return evalExprsAt(k, tr, e.Args, fn)
}

// EvalExprsAt evaluates all expressions in a given slice at a given row on the
// table, and fold their results together using a combinator.
func evalExprsAt(k int, tr trace.Trace, exprs []Expr, fn func(fr.Element, fr.Element) fr.Element) []fr.Element {
	// Evaluate first argument.
	vals := exprs[0].EvalAllAt(k, tr)

	// Continue evaluating the rest.
	for i := 1; i < len(exprs); i++ {
		vs := exprs[i].EvalAllAt(k, tr)
		vals = evalExprsAtApply(vals, vs, fn)
	}

	// Done.
	return vals
}

// Perform a vector operation using the given primitive operator "fn".
func evalExprsAtApply(lhs []fr.Element, rhs []fr.Element, fn func(fr.Element, fr.Element) fr.Element) []fr.Element {
	if len(rhs) == 1 {
		// Optimise for common case.
		for i, ith := range lhs {
			lhs[i] = fn(ith, rhs[0])
		}

		return lhs
	}
	// Harder case
	vals := make([]fr.Element, 0)
	// Perform n x m operations
	for _, ith := range lhs {
		for _, jth := range rhs {
			vals = append(vals, fn(ith, jth))
		}
	}

	return vals
}
