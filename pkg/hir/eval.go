package hir

import (
	"fmt"
	"reflect"

	"github.com/consensys/gnark-crypto/ecc/bls12-377/fr"
	"github.com/consensys/go-corset/pkg/trace"
	tr "github.com/consensys/go-corset/pkg/trace"
	"github.com/consensys/go-corset/pkg/util"
)

func evalAtTerm(e Term, k int, trace tr.Trace) []fr.Element {
	switch e := e.(type) {
	case *Add:
		return evalAtAdd(e, k, trace)
	case *Constant:
		return []fr.Element{e.Value}
	case *ColumnAccess:
		val := trace.Column(e.Column).Get(k + e.Shift)
		return []fr.Element{val}
	case *Exp:
		return evalAtExp(e, k, trace)
	case *IfZero:
		return evalAtIfZero(e, k, trace)
	case *List:
		return evalAtList(e, k, trace)
	case *Mul:
		return evalAtMul(e, k, trace)
	case *Norm:
		return evalAtNormalise(e, k, trace)
	case *Sub:
		return evalAtSub(e, k, trace)
	default:
		name := reflect.TypeOf(e).Name()
		panic(fmt.Sprintf("unknown HIR expression \"%s\"", name))
	}
}

func evalAtAdd(e *Add, k int, tr trace.Trace) []fr.Element {
	fn := func(l fr.Element, r fr.Element) fr.Element { l.Add(&l, &r); return l }
	return evalAtTerms(k, tr, e.Args, fn)
}

func evalAtExp(e *Exp, k int, tr trace.Trace) []fr.Element {
	// Check whether argument evaluates to zero or not.
	vals := evalAtTerm(e.Arg, k, tr)
	//
	for i := range vals {
		// Compute exponent
		util.Pow(&vals[i], e.Pow)
	}
	// Done
	return vals
}

func evalAtIfZero(e *IfZero, k int, tr trace.Trace) []fr.Element {
	vals := make([]fr.Element, 0)
	// Evaluate condition
	conditions := evalAtTerm(e.Condition, k, tr)
	// Check all results
	for _, cond := range conditions {
		var vs []fr.Element
		//q
		if cond.IsZero() && e.TrueBranch != nil {
			vs = evalAtTerm(e.TrueBranch, k, tr)
		} else if !cond.IsZero() && e.FalseBranch != nil {
			vs = evalAtTerm(e.FalseBranch, k, tr)
		}
		//
		vals = append(vals, vs...)
	}

	return vals
}

func evalAtMul(e *Mul, k int, tr trace.Trace) []fr.Element {
	fn := func(l fr.Element, r fr.Element) fr.Element { l.Mul(&l, &r); return l }
	return evalAtTerms(k, tr, e.Args, fn)
}

func evalAtNormalise(e *Norm, k int, tr trace.Trace) []fr.Element {
	// Check whether argument evaluates to zero or not.
	vals := evalAtTerm(e.Arg, k, tr)
	// Normalise value (if necessary)
	for i := range vals {
		if !vals[i].IsZero() {
			vals[i].SetOne()
		}
	}
	// Done
	return vals
}

func evalAtList(e *List, k int, tr trace.Trace) []fr.Element {
	var vals []fr.Element
	//
	for _, arg := range e.Args {
		vs := evalAtTerm(arg, k, tr)
		vals = append(vals, vs...)
	}
	//
	return vals
}

func evalAtSub(e *Sub, k int, tr trace.Trace) []fr.Element {
	fn := func(l fr.Element, r fr.Element) fr.Element { l.Sub(&l, &r); return l }
	return evalAtTerms(k, tr, e.Args, fn)
}

// EvalExprsAt evaluates all expressions in a given slice at a given row on the
// table, and fold their results together using a combinator.
func evalAtTerms(k int, tr trace.Trace, terms []Term, fn func(fr.Element, fr.Element) fr.Element) []fr.Element {
	// Evaluate first argument.
	vals := evalAtTerm(terms[0], k, tr)
	// Continue evaluating the rest.
	for i := 1; i < len(terms); i++ {
		vs := evalAtTerm(terms[i], k, tr)
		vals = evalAtTermsApply(vals, vs, fn)
	}
	// Done.
	return vals
}

// Perform a vector operation using the given primitive operator "fn".
func evalAtTermsApply(lhs []fr.Element, rhs []fr.Element, fn func(fr.Element, fr.Element) fr.Element) []fr.Element {
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
