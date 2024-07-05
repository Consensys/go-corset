package hir

import (
	"github.com/consensys/gnark-crypto/ecc/bls12-377/fr"
	sc "github.com/consensys/go-corset/pkg/schema"
	"github.com/consensys/go-corset/pkg/trace"
	"github.com/consensys/go-corset/pkg/util"
)

// ============================================================================
// ZeroArrayTest
// ============================================================================

// ZeroArrayTest is a wrapper which converts an array of expressions into a
// Testable constraint.  Specifically, by checking whether or not the each
// expression vanishes (i.e. evaluates to zero).
type ZeroArrayTest struct {
	Expr Expr
}

// TestAt determines whether or not every element from a given array of
// expressions evaluates to zero. Observe that any expressions which are
// undefined are assumed to hold.
func (p ZeroArrayTest) TestAt(row int, tr trace.Trace) bool {
	// Evalues expression yielding zero or more values.
	vals := p.Expr.EvalAllAt(row, tr)
	// Check each value in turn against zero.
	for _, val := range vals {
		if val != nil && !val.IsZero() {
			// This expression does not evaluat to zero, hence failure.
			return false
		}
	}
	// Success
	return true
}

func (p ZeroArrayTest) String() string {
	return p.Expr.String()
}

// Bounds determines the bounds for this zero test.
func (p ZeroArrayTest) Bounds() util.Bounds {
	return p.Expr.Bounds()
}

// Context determines the evaluation context (i.e. enclosing module) for this
// expression.
func (p ZeroArrayTest) Context(schema sc.Schema) trace.Context {
	return p.Expr.Context(schema)
}

// ============================================================================
// UnitExpr
// ============================================================================

// UnitExpr is an adaptor for a general expression which can be used in
// situations where an Evaluable expression is required.  This performs a
// similar function to the ZeroArrayTest, but actually produces a value.  A
// strict requirement is placed that the given expression always returns (via
// EvalAll) exactly one result.  This means the presence of certain constructs,
// such as lists and if conditions can result in Eval causing a panic.
type UnitExpr struct {
	//
	expr Expr
}

// NewUnitExpr constructs a unit wrapper around an HIR expression.  In essence,
// this introduces a runtime check that the given expression only every reduces
// to a single value.  Evaluation of this expression will panic if that
// condition does not hold.  The intention is that this error is checked for
// upstream (e.g. as part of the compiler front end).
func NewUnitExpr(expr Expr) UnitExpr {
	return UnitExpr{expr}
}

// EvalAt evaluates a column access at a given row in a trace, which returns the
// value at that row of the column in question or nil is that row is
// out-of-bounds.
func (e UnitExpr) EvalAt(k int, tr trace.Trace) *fr.Element {
	vals := e.expr.EvalAllAt(k, tr)
	// Check we got exactly one thing
	if len(vals) == 1 {
		return vals[0]
	}
	// Fail
	panic("invalid unitary expression")
}

// Bounds returns max shift in either the negative (left) or positive
// direction (right).
func (e UnitExpr) Bounds() util.Bounds {
	return e.expr.Bounds()
}

// Context determines the evaluation context (i.e. enclosing module) for this
// expression.
func (e UnitExpr) Context(schema sc.Schema) trace.Context {
	return e.expr.Context(schema)
}
