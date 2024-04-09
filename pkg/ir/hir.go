package ir

import (
	"math/big"
)

// HirExpr An expression in the High-Level Intermediate Representation (HIR).
type HirExpr interface {
	// LowerToMir Lower this expression into the Mid-Level Intermediate
	// Representation.  Observe that a single expression at this
	// level can expand into *multiple* expressions at the MIR
	// level.
	LowerToMir() []MirExpr
}

// ============================================================================
// Definitions
// ============================================================================

type HirAdd Add[HirExpr]
type HirSub Sub[HirExpr]
type HirMul Mul[HirExpr]
type HirConstant = Constant
type HirIfZero IfZero[HirExpr]
type HirList List[HirExpr]

// ============================================================================
// Lowering
// ============================================================================

func (e *HirAdd) LowerToMir() []MirExpr {
	return LowerWithNaryConstructor(e.arguments, func(nargs []MirExpr) MirExpr {
		return &MirAdd{nargs}
	})
}

// LowerToMir Lowering a constant is straightforward as it is already in the correct form.
func (e *HirConstant) LowerToMir() []MirExpr {
	return []MirExpr{e}
}

func (e *HirMul) LowerToMir() []MirExpr {
	return LowerWithNaryConstructor(e.arguments, func(nargs []MirExpr) MirExpr {
		return &MirMul{nargs}
	})
}

// A list is lowered by eliminating it altogether.
func (e *HirList) LowerToMir() []MirExpr {
	var res []MirExpr
	for i := 0; i < len(e.elements); i++ {
		// Lower ith argument
		iths := e.elements[i].LowerToMir()
		// Append all as one
		res = append(res, iths...)
	}
	return res
}

func (e *HirIfZero) LowerToMir() []MirExpr {
	var res []MirExpr
	// Lower required condition
	c := e.condition
	// Lower optional true branch
	t := e.trueBranch
	// Lower optional false branch
	f := e.falseBranch
	// Add constraints arising from true branch
	if t != nil {
		// (1 - NORM(x)) * y for true branch
		ts := LowerWithBinaryConstructor(c, t, func(x MirExpr, y MirExpr) MirExpr {
			one := &MirConstant{big.NewInt(1)}
			norm_x := &MirNormalise{x}
			one_minus_norm_x := &MirSub{[]MirExpr{one, norm_x}}
			return &MirMul{[]MirExpr{one_minus_norm_x, y}}
		})
		res = append(res, ts...)
	}
	// Add constraints arising from false branch
	if f != nil {
		// x * y for false branch
		fs := LowerWithBinaryConstructor(c, f, func(x MirExpr, y MirExpr) MirExpr {
			return &MirMul{[]MirExpr{x, y}}
		})
		res = append(res, fs...)
	}
	// Done
	return res
}

func (e *HirSub) LowerToMir() []MirExpr {
	return LowerWithNaryConstructor(e.arguments, func(nargs []MirExpr) MirExpr {
		return &MirSub{nargs}
	})
}

// ============================================================================
// Helpers
// ============================================================================

type BinaryConstructor func(MirExpr, MirExpr) MirExpr
type NaryConstructor func([]MirExpr) MirExpr

// A generic mechanism for lowering down to a binary expression.
func LowerWithBinaryConstructor(lhs HirExpr, rhs HirExpr, create BinaryConstructor) []MirExpr {
	var res []MirExpr
	// Lower all three expressions
	is := lhs.LowerToMir()
	js := rhs.LowerToMir()
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

// LowerWithNaryConstructor Perform the cross-product expansion of an nary HIR expression.
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
func LowerWithNaryConstructor(args []HirExpr, constructor NaryConstructor) []MirExpr {
	// Accumulator is initially empty
	acc := make([]MirExpr, len(args))
	// Start from the first argument
	return LowerWithNaryConstructorHelper(0, acc, args, constructor)
}

// LowerWithNaryConstructorHelper This manages progress through the cross-product expansion.
// Specifically, "i" determines how much of args has been lowered thus
// far, whilst "acc" represents the current array being generated.
func LowerWithNaryConstructorHelper(i int, acc []MirExpr, args []HirExpr, constructor NaryConstructor) []MirExpr {
	if i == len(acc) {
		// Base Case
		nacc := make([]MirExpr, len(acc))
		// Clone the slice because it is used as a temporary
		// working storage during the expansion.
		copy(nacc, acc)
		// Apply the constructor to produce the appropriate
		// MirExpr.
		return []MirExpr{constructor(nacc)}
	} else {
		// Recursive Case
		nargs := []MirExpr{}
		for _, ith := range args[i].LowerToMir() {
			acc[i] = ith
			iths := LowerWithNaryConstructorHelper(i+1, acc, args, constructor)
			nargs = append(nargs, iths...)
		}
		return nargs
	}
}
