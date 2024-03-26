package ir

// An MirExpression in the Mid-Level Intermediate Representation (MIR).
type MirExpr interface {
	// Lower this MirExpression into the Arithmetic Intermediate
	// Representation.  Essentially, this means eliminating normalising
	// expressions by introducing new columns into the enclosing table (with
	// appropriate constraints).
	LowerToAir() AirExpr
}

// ============================================================================
// Definitions
// ============================================================================

type MirAdd Add[MirExpr]
type MirConstant = Constant
type MirNormalise Normalise[MirExpr]

// ============================================================================
// Lowering
// ============================================================================

func (e *MirAdd) LowerToAir() AirExpr {
	n := len(e.arguments)
	nargs := make([]AirExpr, n)
	for i := 0; i < n; i++ {
		nargs[i] = e.arguments[i].LowerToAir()
	}
	return &AirAdd{nargs}
}

func (e *MirNormalise) LowerToAir() AirExpr {
	panic("implement me!")
}

// Lowering a constant is straightforward as it is already in the correct form.
func (e *MirConstant) LowerToAir() AirExpr {
	return e
}
