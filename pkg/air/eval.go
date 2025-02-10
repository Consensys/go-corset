package air

import (
	"fmt"
	"reflect"

	"github.com/consensys/gnark-crypto/ecc/bls12-377/fr"
	sc "github.com/consensys/go-corset/pkg/schema"
	tr "github.com/consensys/go-corset/pkg/trace"
)

func evalAtTerm[T sc.Metric[T]](e Term, k int, trace tr.Trace) (fr.Element, T) {
	var id T
	//
	switch e := e.(type) {
	case *Add:
		return evalAtAdd[T](e, k, trace)
	case *Constant:
		return e.Value, id.Empty()
	case *ColumnAccess:
		return trace.Column(e.Column).Get(k + e.Shift), id.Empty()
	case *Sub:
		return evalAtSub[T](e, k, trace)
	case *Mul:
		return evalAtMul[T](e, k, trace)
	default:
		name := reflect.TypeOf(e).Name()
		panic(fmt.Sprintf("unknown AIR expression \"%s\"", name))
	}
}

func evalAtAdd[T sc.Metric[T]](e *Add, k int, trace tr.Trace) (fr.Element, T) {
	// Evaluate first argument
	val, metric := evalAtTerm[T](e.Args[0], k, trace)
	// Continue evaluating the rest
	for i := 1; i < len(e.Args); i++ {
		ith, ithmetric := evalAtTerm[T](e.Args[i], k, trace)
		val.Add(&val, &ith)
		// update metric
		metric = metric.Join(ithmetric)
	}
	// Done
	return val, metric
}

func evalAtMul[T sc.Metric[T]](e *Mul, k int, trace tr.Trace) (fr.Element, T) {
	n := uint(len(e.Args))
	// Evaluate first argument
	val, metric := evalAtTerm[T](e.Args[0], k, trace)
	// Continue evaluating the rest
	for i := uint(1); i < n; i++ {
		var ith fr.Element
		// Can short-circuit evaluation?
		if val.IsZero() {
			return val, metric.Mark(i, n)
		}
		// No
		ith, metric = evalAtTerm[T](e.Args[i], k, trace)
		//
		val.Mul(&val, &ith)
	}
	// Done
	return val, metric.Mark(n-1, n)
}

func evalAtSub[T sc.Metric[T]](e *Sub, k int, trace tr.Trace) (fr.Element, T) {
	// Evaluate first argument
	val, metric := evalAtTerm[T](e.Args[0], k, trace)
	// Continue evaluating the rest
	for i := 1; i < len(e.Args); i++ {
		ith, ithmetric := evalAtTerm[T](e.Args[i], k, trace)
		val.Sub(&val, &ith)
		// update metric
		metric = metric.Join(ithmetric)
	}
	// Done
	return val, metric
}

// Determine the number of distinct evaluation paths through a given term.
func pathsOfTerm(e Term) uint {
	switch e := e.(type) {
	case *Add:
		count := uint(1)
		//
		for _, arg := range e.Args {
			count *= pathsOfTerm(arg)
		}
		//
		return count
	case *Constant:
		return 1
	case *ColumnAccess:
		return 1
	case *Sub:
		count := uint(1)
		//
		for _, arg := range e.Args {
			count *= pathsOfTerm(arg)
		}
		//
		return count
	case *Mul:
		count := uint(0)
		//
		for _, arg := range e.Args {
			count += pathsOfTerm(arg)
		}
		//
		return count
	default:
		name := reflect.TypeOf(e).Name()
		panic(fmt.Sprintf("unknown AIR expression \"%s\"", name))
	}
}
