package mir

import (
	"fmt"

	"github.com/consensys/gnark-crypto/ecc/bls12-377/fr"
	sc "github.com/consensys/go-corset/pkg/schema"
	"github.com/consensys/go-corset/pkg/util"
)

// ApplyConstantPropagation simply collapses constant expressions down to single
// values.  For example, "(+ 1 2)" would be collapsed down to "3".
func applyConstantPropagation(e Expr, schema sc.Schema) Expr {
	if p, ok := e.(*Add); ok {
		return applyConstantPropagationAdd(p.Args, schema)
	} else if _, ok := e.(*Constant); ok {
		return e
	} else if _, ok := e.(*ColumnAccess); ok {
		return e
	} else if p, ok := e.(*Mul); ok {
		return applyConstantPropagationMul(p.Args, schema)
	} else if p, ok := e.(*Exp); ok {
		return applyConstantPropagationExp(p.Arg, p.Pow, schema)
	} else if p, ok := e.(*Normalise); ok {
		return applyConstantPropagationNorm(p.Arg, schema)
	} else if p, ok := e.(*Sub); ok {
		return applyConstantPropagationSub(p.Args, schema)
	}
	// Should be unreachable
	panic(fmt.Sprintf("unknown expression: %s", e.Lisp(schema).String(true)))
}

func applyConstantPropagationAdd(es []Expr, schema sc.Schema) Expr {
	sum := fr.NewElement(0)
	count := 0
	rs := make([]Expr, len(es))
	//
	for i, e := range es {
		rs[i] = applyConstantPropagation(e, schema)
		// Check for constant
		c, ok := rs[i].(*Constant)
		// Try to continue sum
		if ok {
			sum.Add(&sum, &c.Value)
			// Increase count of constants
			count++
		}
	}
	// Check if constant
	if count == len(es) {
		// Propagate constant
		return &Constant{sum}
	} else if count > 1 {
		rs = mergeConstants(sum, rs)
	}
	// Done
	return &Add{rs}
}

func applyConstantPropagationSub(es []Expr, schema sc.Schema) Expr {
	var sum fr.Element

	is_const := true
	rs := make([]Expr, len(es))
	//
	for i, e := range es {
		rs[i] = applyConstantPropagation(e, schema)
		// Check for constant
		c, ok := rs[i].(*Constant)
		// Try to continue sum
		if ok && i == 0 {
			sum = c.Value
		} else if ok && is_const {
			sum.Sub(&sum, &c.Value)
		} else {
			is_const = false
		}
	}
	// Check if constant
	if is_const {
		// Propagate constant
		return &Constant{sum}
	}
	// Done
	return &Sub{rs}
}

func applyConstantPropagationMul(es []Expr, schema sc.Schema) Expr {
	one := fr.NewElement(1)
	prod := one
	rs := make([]Expr, len(es))
	ones := 0
	consts := 0
	//
	for i, e := range es {
		rs[i] = applyConstantPropagation(e, schema)
		// Check for constant
		c, ok := rs[i].(*Constant)
		//
		if ok && c.Value.IsZero() {
			// No matter what, outcome is zero.
			return &Constant{c.Value}
		} else if ok && c.Value.IsOne() {
			ones++
			consts++
			rs[i] = nil
		} else if ok {
			// Continue building constant
			prod.Mul(&prod, &c.Value)
			//
			consts++
		}
	}
	// Check if constant
	if consts == len(es) {
		return &Constant{prod}
	} else if ones > 0 {
		rs = util.RemoveMatching[Expr](rs, func(item Expr) bool { return item == nil })
	}
	// Sanity check what's left.
	if len(rs) == 1 {
		return rs[0]
	} else if consts-ones > 1 {
		// Combine constants
		rs = mergeConstants(prod, rs)
	}
	// Done
	return &Mul{rs}
}

func applyConstantPropagationExp(arg Expr, pow uint64, schema sc.Schema) Expr {
	arg = applyConstantPropagation(arg, schema)
	//
	if c, ok := arg.(*Constant); ok {
		var val fr.Element
		// Clone value
		val.Set(&c.Value)
		// Compute exponent (in place)
		util.Pow(&val, pow)
		// Done
		return &Constant{val}
	}
	//
	return &Exp{arg, pow}
}

func applyConstantPropagationNorm(arg Expr, schema sc.Schema) Expr {
	arg = applyConstantPropagation(arg, schema)
	//
	if c, ok := arg.(*Constant); ok {
		var val fr.Element
		// Clone value
		val.Set(&c.Value)
		// Normalise (in place)
		if !val.IsZero() {
			val.SetOne()
		}
		// Done
		return &Constant{val}
	}
	//
	return &Normalise{arg}
}

// Replace all constants within a given sequence of expressions with a single
// constant (whose value has been precomputed from those constants).  The new
// value replaces the first constant in the list.
func mergeConstants(constant fr.Element, es []Expr) []Expr {
	j := 0
	first := true
	//
	for i := range es {
		// Check for constant
		if _, ok := es[i].(*Constant); ok && first {
			es[j] = &Constant{constant}
			first = false
			j++
		} else if !ok {
			// Retain non-constant expression
			es[j] = es[i]
			j++
		}
	}
	// Return slice
	return es[0:j]
}
