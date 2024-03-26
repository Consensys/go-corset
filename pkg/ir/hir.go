package ir

// An expression in the High-Level Intermediate Representation (HIR).
type HirExpr interface {
	// Lower this expression into the Mid-Level Intermediate
	// Representation.  Observe that a single expression at this
	// level can expand into *multiple* expressions at the MIR
	// level.
	LowerToMir() []MirExpr
}

// ============================================================================
// Definitions
// ============================================================================

type HirAdd Add[HirExpr]
type HirConstant = Constant
type HirIfZero IfZero[HirExpr]
type HirList List[HirExpr]

// ============================================================================
// Lowering
// ============================================================================

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
		ts := LowerWithBinaryConstructor(c, t, func(x MirExpr, y MirExpr) MirExpr {
			panic("got here")
		})
		res = append(res, ts...)
	}
	// Add constraints arising from false branch
	if f != nil {
		fs := LowerWithBinaryConstructor(c, f, func(x MirExpr, y MirExpr) MirExpr {
			panic("got here")
		})
		res = append(res, fs...)
	}
	// Done
	return res
}

// Lowering a constant is straightforward as it is already in the correct form.
func (e *HirConstant) LowerToMir() []MirExpr {
	return []MirExpr{e}
}

// ============================================================================
// Helpers
// ============================================================================

type BinaryConstructor func(MirExpr, MirExpr) MirExpr

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
