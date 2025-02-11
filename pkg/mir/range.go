package mir

import (
	"fmt"
	"math/big"
	"reflect"

	sc "github.com/consensys/go-corset/pkg/schema"
	"github.com/consensys/go-corset/pkg/util"
)

func rangeOfTerm(e Term, schema sc.Schema) *util.Interval {
	switch e := e.(type) {
	case *Add:
		return rangeOfAdd(e.Args, schema)
	case *Constant:
		var c big.Int
		// Extract big integer from field element
		e.Value.BigInt(&c)
		// Return as interval
		return util.NewInterval(&c, &c)
	case *ColumnAccess:
		return rangeOfColumnAccess(e.Column, schema)
	case *Exp:
		bounds := rangeOfTerm(e.Arg, schema)
		bounds.Exp(uint(e.Pow))
		//
		return bounds
	case *Mul:
		return rangeOfMul(e.Args, schema)
	case *Norm:
		return util.NewInterval(big.NewInt(0), big.NewInt(1))
	case *Sub:
		return rangeOfSub(e.Args, schema)
	default:
		name := reflect.TypeOf(e).Name()
		panic(fmt.Sprintf("unknown MIR expression \"%s\"", name))
	}
}

func rangeOfAdd(args []Term, schema sc.Schema) *util.Interval {
	var res util.Interval

	for i, arg := range args {
		ith := rangeOfTerm(arg, schema)
		if i == 0 {
			res.Set(ith)
		} else {
			res.Add(ith)
		}
	}
	//
	return &res
}

func rangeOfColumnAccess(column uint, schema sc.Schema) *util.Interval {
	bound := big.NewInt(2)
	width := int64(schema.Columns().Nth(column).DataType.BitWidth())
	bound.Exp(bound, big.NewInt(width), nil)
	// Subtract 1 because interval is inclusive.
	bound.Sub(bound, big.NewInt(1))
	// Done
	return util.NewInterval(big.NewInt(0), bound)
}

func rangeOfMul(args []Term, schema sc.Schema) *util.Interval {
	var res util.Interval

	for i, arg := range args {
		ith := rangeOfTerm(arg, schema)
		if i == 0 {
			res.Set(ith)
		} else {
			res.Mul(ith)
		}
	}
	//
	return &res
}

func rangeOfSub(args []Term, schema sc.Schema) *util.Interval {
	var res util.Interval

	for i, arg := range args {
		ith := rangeOfTerm(arg, schema)
		if i == 0 {
			res.Set(ith)
		} else {
			res.Sub(ith)
		}
	}
	//
	return &res
}
