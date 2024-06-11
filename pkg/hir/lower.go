package hir

import (
	"fmt"

	"github.com/consensys/gnark-crypto/ecc/bls12-377/fr"
	"github.com/consensys/go-corset/pkg/mir"
)

// LowerTo lowers a sum expression to the MIR level.  This requires expanding
// the arguments, then lowering them.  Furthermore, conditionals are "lifted" to
// the top.
func (e *Add) LowerTo() []mir.Expr {
	return lowerTo(e)
}

// LowerTo lowers a constant to the MIR level.   This requires expanding the
// arguments, then lowering them.  Furthermore, conditionals are "lifted" to the
// top.
func (e *Constant) LowerTo() []mir.Expr {
	return lowerTo(e)
}

// LowerTo lowers a column access to the MIR level.  This requires expanding
// the arguments, then lowering them.  Furthermore, conditionals are "lifted" to
// the top.
func (e *ColumnAccess) LowerTo() []mir.Expr {
	return lowerTo(e)
}

// LowerTo lowers a product expression to the MIR level.  This requires expanding
// the arguments, then lowering them.  Furthermore, conditionals are "lifted" to
// the top.
func (e *Mul) LowerTo() []mir.Expr {
	return lowerTo(e)
}

// LowerTo lowers a list expression to the MIR level by eliminating it
// altogether.  This still requires expanding the arguments, then lowering them.
// Furthermore, conditionals are "lifted" to the top..
func (e *List) LowerTo() []mir.Expr {
	return lowerTo(e)
}

// LowerTo lowers a normalise expression to the MIR level.  This requires
// expanding the arguments, then lowering them.  Furthermore, conditionals are
// "lifted" to the top..
func (e *Normalise) LowerTo() []mir.Expr {
	return lowerTo(e)
}

// LowerTo lowers an if expression to the MIR level by "compiling out" the
// expression using normalisation at the MIR level.  This also requires
// expanding the arguments, then lowering them.  Furthermore, conditionals are
// "lifted" to the top.
func (e *IfZero) LowerTo() []mir.Expr {
	return lowerTo(e)
}

// LowerTo lowers a subtract expression to the MIR level. This also requires
// expanding the arguments, then lowering them.  Furthermore, conditionals are
// "lifted" to the top.
func (e *Sub) LowerTo() []mir.Expr {
	return lowerTo(e)
}

// ============================================================================
// expandedLowerTo
// ============================================================================

// Lowers a given expression to the MIR level.  The expression is first expanded
// into one or more target expressions. Furthermore, conditions must be "lifted"
// to the root.
func lowerTo(e Expr) []mir.Expr {
	// First expand expression
	es := expand(e)
	// Now lower each one (carefully)
	mes := make([]mir.Expr, len(es))
	//
	for i, e := range es {
		c := lowerCondition(e)
		b := lowerBody(e)
		mes[i] = mul2(c, b)
	}
	// Done
	return mes
}

// Lower the "condition" of an expression.  Every expression can be view as a
// conditional constraint of the form "if c then e", where "c" is the condition.
// This is allowed to return nil if the body is unconditional.
func lowerCondition(e Expr) mir.Expr {
	if p, ok := e.(*Add); ok {
		return lowerConditions(p.Args)
	} else if _, ok := e.(*Constant); ok {
		return nil
	} else if _, ok := e.(*ColumnAccess); ok {
		return nil
	} else if p, ok := e.(*Mul); ok {
		return lowerConditions(p.Args)
	} else if p, ok := e.(*Normalise); ok {
		return lowerCondition(p.Arg)
	} else if p, ok := e.(*IfZero); ok {
		return lowerIfZeroCondition(p)
	} else if p, ok := e.(*Sub); ok {
		return lowerConditions(p.Args)
	}
	// Should be unreachable
	panic(fmt.Sprintf("unknown expression: %s", e.String()))
}

func lowerConditions(es []Expr) mir.Expr {
	var r mir.Expr = nil
	for _, e := range es {
		r = mul2(r, lowerCondition(e))
	}

	return r
}

// Lowering conditional expressions is slightly more complex than others, so it
// gets a case of its own.
func lowerIfZeroCondition(e *IfZero) mir.Expr {
	var bc mir.Expr
	// Lower condition
	cc := lowerCondition(e.Condition)
	cb := lowerBody(e.Condition)
	// Add conditions arising
	if e.TrueBranch != nil && e.FalseBranch != nil {
		// Expansion should ensure this case does not exist.  This is necessary
		// to ensure exactly one expression is generated from this expression.
		panic(fmt.Sprintf("unexpanded expression (%s)", e.String()))
	} else if e.TrueBranch != nil {
		// (1 - NORM(cb)) for true branch
		one := new(fr.Element)
		one.SetOne()

		normBody := &mir.Normalise{Arg: cb}
		oneMinusNormBody := &mir.Sub{
			Args: []mir.Expr{
				&mir.Constant{Value: one},
				normBody,
			},
		}

		cb = oneMinusNormBody
		// Lower conditional's arising from body
		bc = lowerCondition(e.TrueBranch)
	} else {
		// Lower conditional's arising from body
		bc = lowerCondition(e.FalseBranch)
	}
	//
	return mul3(cc, cb, bc)
}

// Translate the "body" of an expression.  Every expression can be view as a
// conditional constraint of the form "if c then e", where "e" is the
// constraint.
func lowerBody(e Expr) mir.Expr {
	if p, ok := e.(*Add); ok {
		return &mir.Add{Args: lowerBodies(p.Args)}
	} else if p, ok := e.(*Constant); ok {
		return &mir.Constant{Value: p.Val}
	} else if p, ok := e.(*ColumnAccess); ok {
		return &mir.ColumnAccess{Column: p.Column, Shift: p.Shift}
	} else if p, ok := e.(*Mul); ok {
		return &mir.Mul{Args: lowerBodies(p.Args)}
	} else if p, ok := e.(*Normalise); ok {
		return &mir.Normalise{Arg: lowerBody(p.Arg)}
	} else if p, ok := e.(*IfZero); ok {
		if p.TrueBranch != nil && p.FalseBranch != nil {
			// Expansion should ensure this case does not exist.  This is necessary
			// to ensure exactly one expression is generated from this expression.
			panic(fmt.Sprintf("unexpanded expression (%s)", e.String()))
		} else if p.TrueBranch != nil {
			return lowerBody(p.TrueBranch)
		}
		// Done
		return lowerBody(p.FalseBranch)
	} else if p, ok := e.(*Sub); ok {
		return &mir.Sub{Args: lowerBodies(p.Args)}
	}
	// Should be unreachable
	panic(fmt.Sprintf("unknown expression: %s", e.String()))
}

// Lower a vector of expanded expressions to the MIR level.
func lowerBodies(es []Expr) []mir.Expr {
	rs := make([]mir.Expr, len(es))
	for i, e := range es {
		rs[i] = lowerBody(e)
	}

	return rs
}

// ============================================================================
// expand
// ============================================================================

// Expand an expression into one or more expressions by eliminating lists and
// breaking down conditions.  For example, a list such as say "(begin (- X Y) (-
// Y Z))" is broken down into two distinct expressions "(- X Y)" and "(- Y Z)".
// Likewise, a condition such as "(if X Y Z)" is broken down into two
// expressions "(if X Y)" and "(ifnot X Z)".  These are necessary steps for the
// conversion into a lower-level form.
func expand(e Expr) []Expr {
	if p, ok := e.(*Add); ok {
		return expandWithNaryConstructor(p.Args, func(nargs []Expr) Expr {
			return &Add{Args: nargs}
		})
	} else if _, ok := e.(*Constant); ok {
		return []Expr{e}
	} else if _, ok := e.(*ColumnAccess); ok {
		return []Expr{e}
	} else if p, ok := e.(*Mul); ok {
		return expandWithNaryConstructor(p.Args, func(nargs []Expr) Expr {
			return &Mul{Args: nargs}
		})
	} else if p, ok := e.(*List); ok {
		ees := make([]Expr, 0)
		for _, arg := range p.Args {
			ees = append(ees, expand(arg)...)
		}

		return ees
	} else if p, ok := e.(*Normalise); ok {
		ees := expand(p.Arg)
		for i, ee := range ees {
			ees[i] = &Normalise{ee}
		}

		return ees
	} else if p, ok := e.(*IfZero); ok {
		ees := make([]Expr, 0)
		if p.TrueBranch != nil {
			// Expand true branch with condition
			ees = expandWithBinaryConstructor(p.Condition, p.TrueBranch, func(c Expr, tb Expr) Expr {
				return &IfZero{c, tb, nil}
			})
		}

		if p.FalseBranch != nil {
			// Expand false branch with condition
			fes := expandWithBinaryConstructor(p.Condition, p.FalseBranch, func(c Expr, fb Expr) Expr {
				return &IfZero{c, nil, fb}
			})
			ees = append(ees, fes...)
		}
		// Done
		return ees
	} else if p, ok := e.(*Sub); ok {
		return expandWithNaryConstructor(p.Args, func(nargs []Expr) Expr {
			return &Sub{Args: nargs}
		})
	}
	// Should be unreachable
	panic(fmt.Sprintf("unknown expression: %s", e.String()))
}

type binaryConstructor func(Expr, Expr) Expr
type naryConstructor func([]Expr) Expr

// LowerWithBinaryConstructor is a generic mechanism for lowering down to a binary expression.
func expandWithBinaryConstructor(lhs Expr, rhs Expr, create binaryConstructor) []Expr {
	var res []Expr
	// Lower all three expressions
	is := expand(lhs)
	js := expand(rhs)

	// Now construct
	for i := 0; i < len(is); i++ {
		for j := 0; j < len(js); j++ {
			// Construct binary expression
			expr := create(is[i], js[j])
			// Append to the end
			res = append(res, expr)
		}
	}

	return res
}

// LowerWithNaryConstructor performs the cross-product expansion of an nary HIR
// expression. This is necessary because each argument of that expression will
// itself turn into one or more MIR expressions.  For example, consider lowering
// the following HIR expression:
//
// > (if X Y Z) + 10
//
// Here, (if X Y Z) will lower into two MIR expressions: (1-NORM(X))*Y and X*Z.
// Thus, we need to generate two MIR expressions for our example:
//
// > ((1 - NORM(X)) * Y) + 10 > (X * Y) + 10
//
// Finally, consider an expression such as the following:
//
// > (if X Y Z) + (if A B C)
//
// This will expand into *four* MIR expressions (i.e. the cross product of the
// left and right ifs).
func expandWithNaryConstructor(args []Expr, constructor naryConstructor) []Expr {
	// Accumulator is initially empty
	acc := make([]Expr, len(args))
	// Start from the first argument
	return expandWithNaryConstructorHelper(0, acc, args, constructor)
}

// LowerWithNaryConstructorHelper manages progress through the cross-product expansion.
// Specifically, "i" determines how much of args has been lowered thus
// far, whilst "acc" represents the current array being generated.
func expandWithNaryConstructorHelper(i int, acc []Expr, args []Expr, constructor naryConstructor) []Expr {
	if i == len(acc) {
		// Base Case
		nacc := make([]Expr, len(acc))
		// Clone the slice because it is used as a temporary
		// working storage during the expansion.
		copy(nacc, acc)
		// Apply the constructor to produce the appropriate
		// mir.Expr.
		return []Expr{constructor(nacc)}
	}

	// Recursive Case
	var nargs []Expr

	for _, ith := range expand(args[i]) {
		acc[i] = ith
		iths := expandWithNaryConstructorHelper(i+1, acc, args, constructor)
		nargs = append(nargs, iths...)
	}

	return nargs
}

// Multiply three expressions together, any of which could be nil.
func mul3(lhs mir.Expr, mhs mir.Expr, rhs mir.Expr) mir.Expr {
	return mul2(lhs, mul2(mhs, rhs))
}

// Multiply two expressions together, where either could be nil.  This attempts
// to a little clever in that it combines products together.
func mul2(lhs mir.Expr, rhs mir.Expr) mir.Expr {
	// Check for short-circuit
	if lhs == nil {
		return rhs
	} else if rhs == nil {
		return lhs
	}
	// Look for optimisation
	l, lok := lhs.(*mir.Mul)
	r, rok := rhs.(*mir.Mul)
	//
	if lok && rok {
		l.Args = append(l.Args, r.Args...)
		return l
	} else if lok {
		l.Args = append(l.Args, rhs)
		return l
	} else if rok {
		r.Args = append(r.Args, lhs)
		return r
	}
	// Fall back
	return &mir.Mul{Args: []mir.Expr{lhs, rhs}}
}
