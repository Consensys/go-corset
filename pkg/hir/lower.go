package hir

import (
	"github.com/consensys/gnark-crypto/ecc/bls12-377/fr"
	"github.com/consensys/go-corset/pkg/mir"
)

// LowerTo lowers a sum expression to the MIR level.  This requires lowering all
// of the arguments, and then computing "cross-product" of all combinations.
func (e *Add) LowerTo() []mir.Expr {
	return lowerWithNaryConstructor(e.Args, func(nargs []mir.Expr) mir.Expr {
		return &mir.Add{Args: nargs}
	})
}

// LowerTo lowers a constant to the MIR level.  This is straightforward as it is
// already in the correct form.
func (e *Constant) LowerTo() []mir.Expr {
	c := mir.Constant{Value: e.Val}
	return []mir.Expr{&c}
}

// LowerTo lowers a column access to the MIR level.  This is straightforward as
// it is already in the correct form.
func (e *ColumnAccess) LowerTo() []mir.Expr {
	return []mir.Expr{&mir.ColumnAccess{Column: e.Column, Shift: e.Shift}}
}

// LowerTo lowers a product expression to the MIR level.  This requires lowering all
// of the arguments, and then computing "cross-product" of all combinations.
func (e *Mul) LowerTo() []mir.Expr {
	return lowerWithNaryConstructor(e.Args, func(nargs []mir.Expr) mir.Expr {
		return &mir.Mul{Args: nargs}
	})
}

// LowerTo lowers a normalise expression to the MIR level by first lowering its
// argument.
func (e *Normalise) LowerTo() []mir.Expr {
	mirEs := e.Arg.LowerTo()
	for i, mir_e := range mirEs {
		mirEs[i] = &mir.Normalise{Arg: mir_e}
	}

	return mirEs
}

// LowerTo lowers a list to the MIR level by eliminating it altogether.
func (e *List) LowerTo() []mir.Expr {
	var res []mir.Expr

	for i := 0; i < len(e.Args); i++ {
		// Lower ith argument
		iths := e.Args[i].LowerTo()
		// Append all as one
		res = append(res, iths...)
	}

	return res
}

// LowerTo lowers an if expression to the MIR level by "compiling out" the
// expression using normalisation at the MIR level.
func (e *IfZero) LowerTo() []mir.Expr {
	var res []mir.Expr
	// Lower required condition
	c := e.Condition
	// Lower optional true branch
	t := e.TrueBranch
	// Lower optional false branch
	f := e.FalseBranch
	// Add constraints arising from true branch
	if t != nil {
		// (1 - NORM(x)) * y for true branch
		ts := lowerWithBinaryConstructor(c, t, func(x mir.Expr, y mir.Expr) mir.Expr {
			one := new(fr.Element)
			one.SetOne()

			normX := &mir.Normalise{Arg: x}
			oneMinusNormX := &mir.Sub{
				Args: []mir.Expr{
					&mir.Constant{Value: one},
					normX,
				},
			}

			return &mir.Mul{
				Args: []mir.Expr{
					oneMinusNormX,
					y,
				},
			}
		})

		res = append(res, ts...)
	}
	// Add constraints arising from false branch
	if f != nil {
		// x * y for false branch
		fs := lowerWithBinaryConstructor(c, f, func(x mir.Expr, y mir.Expr) mir.Expr {
			return &mir.Mul{
				Args: []mir.Expr{x, y},
			}
		})

		res = append(res, fs...)
	}

	// Done
	return res
}

// LowerTo lowers a subtract expression to the MIR level.  This requires lowering all
// of the arguments, and then computing "cross-product" of all combinations.
func (e *Sub) LowerTo() []mir.Expr {
	return lowerWithNaryConstructor(e.Args, func(nargs []mir.Expr) mir.Expr {
		return &mir.Sub{Args: nargs}
	})
}

// ============================================================================
// Helpers
// ============================================================================

type binaryConstructor func(mir.Expr, mir.Expr) mir.Expr
type naryConstructor func([]mir.Expr) mir.Expr

// LowerWithBinaryConstructor is a generic mechanism for lowering down to a binary expression.
func lowerWithBinaryConstructor(lhs Expr, rhs Expr, create binaryConstructor) []mir.Expr {
	var res []mir.Expr
	// Lower all three expressions
	is := lhs.LowerTo()
	js := rhs.LowerTo()

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

// LowerWithNaryConstructor performs the cross-product expansion of an nary HIR expression.
// This is necessary because each argument of that expression will
// itself turn into one or more MIR expressions.  For example,
// consider lowering the following HIR expression:
//
// > (if X Y Z) + 10
//
// Here, (if X Y Z) will lower into two MIR expressions: (1-NORM(X))*Y
// and X*Z.  Thus, we need to generate two MIR expressions for our
// example:
//
// > ((1 - NORM(X)) * Y) + 10
// > (X * Y) + 10
//
// Finally, consider an expression such as the following:
//
// > (if X Y Z) + (if A B C)
//
// This will expand into *four* MIR expressions (i.e. the cross
// product of the left and right ifs).
func lowerWithNaryConstructor(args []Expr, constructor naryConstructor) []mir.Expr {
	// Accumulator is initially empty
	acc := make([]mir.Expr, len(args))
	// Start from the first argument
	return lowerWithNaryConstructorHelper(0, acc, args, constructor)
}

// LowerWithNaryConstructorHelper manages progress through the cross-product expansion.
// Specifically, "i" determines how much of args has been lowered thus
// far, whilst "acc" represents the current array being generated.
func lowerWithNaryConstructorHelper(i int, acc []mir.Expr, args []Expr, constructor naryConstructor) []mir.Expr {
	if i == len(acc) {
		// Base Case
		nacc := make([]mir.Expr, len(acc))
		// Clone the slice because it is used as a temporary
		// working storage during the expansion.
		copy(nacc, acc)
		// Apply the constructor to produce the appropriate
		// mir.Expr.
		return []mir.Expr{constructor(nacc)}
	}

	// Recursive Case
	var nargs []mir.Expr

	for _, ith := range args[i].LowerTo() {
		acc[i] = ith
		iths := lowerWithNaryConstructorHelper(i+1, acc, args, constructor)
		nargs = append(nargs, iths...)
	}

	return nargs
}
