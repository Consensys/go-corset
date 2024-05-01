package mir

import (
	"fmt"

	"github.com/consensys/gnark-crypto/ecc/bls12-377/fr"
	"github.com/consensys/go-corset/pkg/air"
)

// LowerTo lowers a sum expression to the AIR level by lowering the arguments.
func (e *Add) LowerTo(tbl *air.Schema) air.Expr {
	return &air.Add{Args: lowerExprs(e.Args, tbl)}
}

// LowerTo lowers a subtract expression to the AIR level by lowering the arguments.
func (e *Sub) LowerTo(tbl *air.Schema) air.Expr {
	return &air.Sub{Args: lowerExprs(e.Args, tbl)}
}

// LowerTo lowers a product expression to the AIR level by lowering the arguments.
func (e *Mul) LowerTo(tbl *air.Schema) air.Expr {
	return &air.Mul{Args: lowerExprs(e.Args, tbl)}
}

// LowerTo lowers a normalise expression to the AIR level by "compiling it out"
// using a computed column.
func (p *Normalise) LowerTo(tbl *air.Schema) air.Expr {
	// Lower the expression being normalised
	e := p.Arg.LowerTo(tbl)
	// Invert expression
	ie := &air.Inverse{Expr: e}
	// Determine computed column name
	name := ie.String()
	// Add new column (if it does not already exist)
	if !tbl.HasColumn(name) {
		// Add computed column
		tbl.AddColumn(air.NewComputedColumn(name, ie))
	}

	one := fr.NewElement(1)
	// Construct 1/e
	inv_e := &air.ColumnAccess{Column: name, Shift: 0}
	// Construct e/e
	e_inv_e := &air.Mul{Args: []air.Expr{e, inv_e}}
	// Construct 1 == e/e
	one_e_e := &air.Sub{Args: []air.Expr{&air.Constant{Value: &one}, e_inv_e}}
	// Construct (e != 0) ==> (1 == e/e)
	e_implies_one_e_e := &air.Mul{Args: []air.Expr{e, one_e_e}}
	// Construct (1/e != 0) ==> (1 == e/e)
	inv_e_implies_one_e_e := &air.Mul{Args: []air.Expr{inv_e, one_e_e}}
	// Ensure (e != 0) ==> (1 == e/e)
	l_name := fmt.Sprintf("[%s <=]", ie.String())
	tbl.AddConstraint(&air.VanishingConstraint{Handle: l_name, Expr: e_implies_one_e_e})
	// Ensure (e/e != 0) ==> (1 == e/e)
	r_name := fmt.Sprintf("[%s =>]", ie.String())
	tbl.AddConstraint(&air.VanishingConstraint{Handle: r_name, Expr: inv_e_implies_one_e_e})
	// Done
	return &air.Mul{Args: []air.Expr{e, &air.ColumnAccess{Column: name, Shift: 0}}}
}

// LowerTo lowers a column access to the AIR level.  This is straightforward as
// it is already in the correct form.
func (e *ColumnAccess) LowerTo(tbl *air.Schema) air.Expr {
	return &air.ColumnAccess{Column: e.Column, Shift: e.Shift}
}

// LowerTo lowers a constant to the AIR level.  This is straightforward as it is
// already in the correct form.
func (e *Constant) LowerTo(tbl *air.Schema) air.Expr {
	return &air.Constant{Value: e.Value}
}

// Lower a set of zero or more MIR expressions.
func lowerExprs(exprs []Expr, tbl *air.Schema) []air.Expr {
	n := len(exprs)
	nexprs := make([]air.Expr, n)

	for i := 0; i < n; i++ {
		nexprs[i] = exprs[i].LowerTo(tbl)
	}

	return nexprs
}
