package air

import (
	"fmt"
	"reflect"

	"github.com/consensys/gnark-crypto/ecc/bls12-377/fr"
	"github.com/consensys/go-corset/pkg/trace"
)

func evalAtTerm(e Term, k int, tr trace.Trace) fr.Element {
	switch e := e.(type) {
	case *Add:
		return evalAtAdd(e, k, tr)
	case *Constant:
		return e.Value
	case *ColumnAccess:
		return tr.Column(e.Column).Get(k + e.Shift)
	case *Sub:
		return evalAtSub(e, k, tr)
	case *Mul:
		return evalAtMul(e, k, tr)
	default:
		name := reflect.TypeOf(e).Name()
		panic(fmt.Sprintf("unknown AIR expression \"%s\"", name))
	}
}

func evalAtAdd(e *Add, k int, tr trace.Trace) fr.Element {
	// Evaluate first argument
	val := evalAtTerm(e.Args[0], k, tr)
	// Continue evaluating the rest
	for i := 1; i < len(e.Args); i++ {
		ith := evalAtTerm(e.Args[i], k, tr)
		val.Add(&val, &ith)
	}
	// Done
	return val
}

func evalAtMul(e *Mul, k int, tr trace.Trace) fr.Element {
	// Evaluate first argument
	val := evalAtTerm(e.Args[0], k, tr)
	// Continue evaluating the rest
	for i := 1; i < len(e.Args); i++ {
		// Can short-circuit evaluation?
		if val.IsZero() {
			break
		}
		// No
		ith := evalAtTerm(e.Args[i], k, tr)
		val.Mul(&val, &ith)
	}
	// Done
	return val
}

func evalAtSub(e *Sub, k int, tr trace.Trace) fr.Element {
	// Evaluate first argument
	val := evalAtTerm(e.Args[0], k, tr)
	// Continue evaluating the rest
	for i := 1; i < len(e.Args); i++ {
		ith := evalAtTerm(e.Args[i], k, tr)
		val.Sub(&val, &ith)
	}
	// Done
	return val
}
