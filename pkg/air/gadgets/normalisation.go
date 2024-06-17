package gadgets

import (
	"fmt"

	"github.com/consensys/gnark-crypto/ecc/bls12-377/fr"
	"github.com/consensys/go-corset/pkg/air"
	"github.com/consensys/go-corset/pkg/table"
	"github.com/consensys/go-corset/pkg/util"
)

// Normalise constructs an expression representing the normalised value of e.
// That is, an expression which is 0 when e is 0, and 1 when e is non-zero.
// This is done by introducing a synthetic column to hold the (pseudo)
// mutliplicative inverse of e.
func Normalise(e air.Expr, tbl *air.Schema) air.Expr {
	// Construct pseudo multiplicative inverse of e.
	ie := ApplyPseudoInverseGadget(e, tbl)
	// Return e * e⁻¹.
	return e.Mul(ie)
}

// ApplyPseudoInverseGadget constructs an expression representing the
// (pseudo) multiplicative inverse of another expression.  Since this cannot be computed
// directly using arithmetic constraints, it is done by adding a new computed
// column which holds the multiplicative inverse.  Constraints are also added to
// ensure it really holds the inverted value.
func ApplyPseudoInverseGadget(e air.Expr, tbl *air.Schema) air.Expr {
	// Construct inverse computation
	ie := &Inverse{Expr: e}
	// Determine computed column name
	name := ie.String()
	// Add new column (if it does not already exist)
	if !tbl.HasColumn(name) {
		// Add (synthetic) computed column
		tbl.AddColumn(name, true)
		tbl.AddComputation(table.NewComputedColumn(name, ie))
	}

	// Construct 1/e
	inv_e := air.NewColumnAccess(name, 0)
	// Construct e/e
	e_inv_e := e.Mul(inv_e)
	// Construct 1 == e/e
	one_e_e := air.NewConst64(1).Equate(e_inv_e)
	// Construct (e != 0) ==> (1 == e/e)
	e_implies_one_e_e := e.Mul(one_e_e)
	// Construct (1/e != 0) ==> (1 == e/e)
	inv_e_implies_one_e_e := inv_e.Mul(one_e_e)
	// Ensure (e != 0) ==> (1 == e/e)
	l_name := fmt.Sprintf("[%s <=]", ie.String())
	tbl.AddVanishingConstraint(l_name, nil, e_implies_one_e_e)
	// Ensure (e/e != 0) ==> (1 == e/e)
	r_name := fmt.Sprintf("[%s =>]", ie.String())
	tbl.AddVanishingConstraint(r_name, nil, inv_e_implies_one_e_e)
	// Done
	return air.NewColumnAccess(name, 0)
}

// Inverse represents a computation which computes the multiplicative
// inverse of a given AIR expression.
type Inverse struct{ Expr air.Expr }

// EvalAt computes the multiplicative inverse of a given expression at a given
// row in the table.
func (e *Inverse) EvalAt(k int, tbl table.Trace) *fr.Element {
	inv := new(fr.Element)
	val := e.Expr.EvalAt(k, tbl)
	// Catch undefined case
	if val == nil {
		return nil
	}
	// Go syntax huh?
	return inv.Inverse(val)
}

// Bounds returns max shift in either the negative (left) or positive
// direction (right).
func (e *Inverse) Bounds() util.Bounds { return e.Expr.Bounds() }

func (e *Inverse) String() string {
	return fmt.Sprintf("(inv %s)", e.Expr)
}
