package hir

import (
	"github.com/consensys/gnark-crypto/ecc/bls12-377/fr"
	"github.com/consensys/go-corset/pkg/mir"
)

// LowerTo lowers to Mid-Level Intermediate Representation (MIR).
func (e *Add) LowerTo() []mir.Expr {
	return LowerWithNaryConstructor(e.arguments, func(nargs []mir.Expr) mir.Expr {
		return &mir.Add{
			Arguments: nargs,
		}
	})
}

// LowerTo lowering a constant is straightforward as it is already in the correct form.
func (e *Constant) LowerTo() []mir.Expr {
	c := mir.Constant{Value: e.Val}
	return []mir.Expr{&c}
}

// LowerTo lowers to Mid-Level Intermediate Representation (MIR).
func (e *ColumnAccess) LowerTo() []mir.Expr {
	return []mir.Expr{&mir.ColumnAccess{Column: e.Column(), Shift: e.Shift()}}
}

// LowerTo lowers to Mid-Level Intermediate Representation (MIR).
func (e *Mul) LowerTo() []mir.Expr {
	return LowerWithNaryConstructor(e.arguments, func(nargs []mir.Expr) mir.Expr {
		return &mir.Mul{Arguments: nargs}
	})
}

// LowerTo lowers to Mid-Level Intermediate Representation (MIR).
func (e *Normalise) LowerTo() []mir.Expr {
	mirEs := e.expr.LowerTo()
	for i, mirE := range mirEs {
		mirEs[i] = &mir.Normalise{Expr: mirE}
	}

	return mirEs
}

// LowerTo lowers a list by eliminating it altogether.
func (e *List) LowerTo() []mir.Expr {
	var res []mir.Expr

	for i := 0; i < len(e.arguments); i++ {
		// Lower ith argument
		iths := e.arguments[i].LowerTo()
		// Append all as one
		res = append(res, iths...)
	}

	return res
}

// LowerTo lowers to Mid-Level Intermediate Representation (MIR).
func (e *IfZero) LowerTo() []mir.Expr {
	var res []mir.Expr
	// Lower required condition
	c := e.condition
	// Lower optional true branch
	t := e.trueBranch
	// Lower optional false branch
	f := e.falseBranch
	// Add constraints arising from true branch
	if t != nil {
		// (1 - NORM(x)) * y for true branch
		ts := LowerWithBinaryConstructor(c, t, func(x mir.Expr, y mir.Expr) mir.Expr {
			one := new(fr.Element)
			one.SetOne()

			normX := &mir.Normalise{Expr: x}
			oneMinusNormX := &mir.Sub{
				Arguments: []mir.Expr{
					&mir.Constant{Value: one},
					normX,
				},
			}

			return &mir.Mul{
				Arguments: []mir.Expr{
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
		fs := LowerWithBinaryConstructor(c, f, func(x mir.Expr, y mir.Expr) mir.Expr {
			return &mir.Mul{
				Arguments: []mir.Expr{x, y},
			}
		})

		res = append(res, fs...)
	}

	// Done
	return res
}

// LowerTo lowers to Mid-Level Intermediate Representation (MIR).
func (e *Sub) LowerTo() []mir.Expr {
	return LowerWithNaryConstructor(e.arguments, func(nargs []mir.Expr) mir.Expr {
		return &mir.Sub{Arguments: nargs}
	})
}

// ============================================================================
// Helpers
// ============================================================================

type BinaryConstructor func(mir.Expr, mir.Expr) mir.Expr
type NaryConstructor func([]mir.Expr) mir.Expr

// LowerWithBinaryConstructor is a generic mechanism for lowering down to a binary expression.
func LowerWithBinaryConstructor(lhs Expr, rhs Expr, create BinaryConstructor) []mir.Expr {
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
func LowerWithNaryConstructor(args []Expr, constructor NaryConstructor) []mir.Expr {
	// Accumulator is initially empty
	acc := make([]mir.Expr, len(args))
	// Start from the first argument
	return LowerWithNaryConstructorHelper(0, acc, args, constructor)
}

// LowerWithNaryConstructorHelper manages progress through the cross-product expansion.
// Specifically, "i" determines how much of args has been lowered thus
// far, whilst "acc" represents the current array being generated.
func LowerWithNaryConstructorHelper(i int, acc []mir.Expr, args []Expr, constructor NaryConstructor) []mir.Expr {
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
		iths := LowerWithNaryConstructorHelper(i+1, acc, args, constructor)
		nargs = append(nargs, iths...)
	}

	return nargs
}
