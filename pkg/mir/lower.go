package mir

import (
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
	// Construct an expression representing the normalised value of e.  That is,
	// an expression which is 0 when e is 0, and 1 when e is non-zero.
	return air.Norm(e, tbl)
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
