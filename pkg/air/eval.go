package air

import (
	"fmt"
	"reflect"

	"github.com/consensys/gnark-crypto/ecc/bls12-377/fr"
	tr "github.com/consensys/go-corset/pkg/trace"
)

func evalAtTerm(e Term, k int, trace tr.Trace) (fr.Element, uint) {
	switch e := e.(type) {
	case *Add:
		return evalAtAdd(e, k, trace)
	case *Constant:
		return e.Value, 0
	case *ColumnAccess:
		return trace.Column(e.Column).Get(k + e.Shift), 0
	case *Sub:
		return evalAtSub(e, k, trace)
	case *Mul:
		return evalAtMul(e, k, trace)
	default:
		name := reflect.TypeOf(e).Name()
		panic(fmt.Sprintf("unknown AIR expression \"%s\"", name))
	}
}

func evalAtAdd(e *Add, k int, trace tr.Trace) (fr.Element, uint) {
	var n = uint(len(e.Args))
	// Evaluate first argument
	val, metric := evalAtTerm(e.Args[0], k, trace)
	// Continue evaluating the rest
	for i := 1; i < len(e.Args); i++ {
		ith, ithmetric := evalAtTerm(e.Args[i], k, trace)
		val.Add(&val, &ith)
		// update metric
		metric = (metric * n) + ithmetric
	}
	// Done
	return val, metric
}

func evalAtMul(e *Mul, k int, trace tr.Trace) (fr.Element, uint) {
	n := uint(len(e.Args))
	// Evaluate first argument
	val, metric := evalAtTerm(e.Args[0], k, trace)
	//
	metric = (metric * n) + uint(0)
	// Continue evaluating the rest
	for i := 1; i < len(e.Args); i++ {
		var ith fr.Element
		// Can short-circuit evaluation?
		if val.IsZero() {
			break
		}
		// No
		ith, metric = evalAtTerm(e.Args[i], k, trace)
		metric = (metric * n) + uint(i)
		//
		val.Mul(&val, &ith)
	}
	// Done
	return val, metric
}

func evalAtSub(e *Sub, k int, trace tr.Trace) (fr.Element, uint) {
	var n = uint(len(e.Args))
	// Evaluate first argument
	val, metric := evalAtTerm(e.Args[0], k, trace)
	// Continue evaluating the rest
	for i := 1; i < len(e.Args); i++ {
		ith, ithmetric := evalAtTerm(e.Args[i], k, trace)
		val.Sub(&val, &ith)
		// update metric
		metric = (metric * n) + ithmetric

	}
	// Done
	return val, metric
}
