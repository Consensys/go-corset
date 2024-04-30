package air

import (
	"github.com/consensys/gnark-crypto/ecc/bls12-377/fr"
	"github.com/consensys/go-corset/pkg/table"
)

func (e *ColumnAccess) EvalAt(k int, tbl table.Trace) *fr.Element {
	val, _ := tbl.GetByName(e.Column, k+e.Shift)
	// We can ignore err as val is always nil when err != nil.
	// Furthermore, as stated in the documentation for this
	// method, we return nil upon error.
	if val == nil {
		// Indicates an out-of-bounds access of some kind.
		return val
	}

	var clone fr.Element

	// Clone original value
	return clone.Set(val)

}

func (e *Constant) EvalAt(k int, tbl table.Trace) *fr.Element {
	var clone fr.Element
	// Clone original value
	return clone.Set(e.Value)
}

func (e *Add) EvalAt(k int, tbl table.Trace) *fr.Element {
	fn := func(l *fr.Element, r *fr.Element) { l.Add(l, r) }
	return EvalExprsAt(k, tbl, e.Arguments, fn)
}

func (e *Sub) EvalAt(k int, tbl table.Trace) *fr.Element {
	fn := func(l *fr.Element, r *fr.Element) { l.Sub(l, r) }
	return EvalExprsAt(k, tbl, e.Arguments, fn)
}

// EvalAt .
func (e *Mul) EvalAt(k int, tbl table.Trace) *fr.Element {
	fn := func(l *fr.Element, r *fr.Element) { l.Mul(l, r) }
	return EvalExprsAt(k, tbl, e.Arguments, fn)
}

func (e *Inverse) EvalAt(k int, tbl table.Trace) *fr.Element {
	inv := new(fr.Element)
	val := e.Expr.EvalAt(k, tbl)
	// Go syntax huh?
	return inv.Inverse(val)
}

// EvalExprsAt evaluates all expressions in a given slice at a given row on the
// table, and fold their results together using a combinator.
func EvalExprsAt(k int, tbl table.Trace, exprs []Expr, fn func(*fr.Element, *fr.Element)) *fr.Element {
	// Evaluate first argument
	val := exprs[0].EvalAt(k, tbl)
	if val == nil {
		return nil
	}

	// Continue evaluating the rest
	for i := 1; i < len(exprs); i++ {
		ith := exprs[i].EvalAt(k, tbl)
		if ith == nil {
			return ith
		}

		fn(val, ith)
	}

	// Done
	return val
}
