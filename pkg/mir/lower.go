package mir

import (
	"github.com/consensys/go-corset/pkg/air"
	air_gadgets "github.com/consensys/go-corset/pkg/air/gadgets"
)

// LowerTo lowers a sum expression to the AIR level by lowering the arguments.
func (e *Add) LowerTo(schema *air.Schema) air.Expr {
	return &air.Add{Args: lowerExprs(e.Args, schema)}
}

// LowerTo lowers a subtract expression to the AIR level by lowering the arguments.
func (e *Sub) LowerTo(schema *air.Schema) air.Expr {
	return &air.Sub{Args: lowerExprs(e.Args, schema)}
}

// LowerTo lowers a product expression to the AIR level by lowering the arguments.
func (e *Mul) LowerTo(schema *air.Schema) air.Expr {
	return &air.Mul{Args: lowerExprs(e.Args, schema)}
}

// LowerTo lowers a normalise expression to the AIR level by "compiling it out"
// using a computed column.
func (p *Normalise) LowerTo(schema *air.Schema) air.Expr {
	// Lower the expression being normalised
	e := p.Arg.LowerTo(schema)
	// Construct an expression representing the normalised value of e.  That is,
	// an expression which is 0 when e is 0, and 1 when e is non-zero.
	return air_gadgets.Normalise(e, schema)
}

// LowerTo lowers a column access to the AIR level.  This is straightforward as
// it is already in the correct form.
func (e *ColumnAccess) LowerTo(schema *air.Schema) air.Expr {
	return &air.ColumnAccess{Column: e.Column, Shift: e.Shift}
}

// LowerTo lowers a constant to the AIR level.  This is straightforward as it is
// already in the correct form.
func (e *Constant) LowerTo(schema *air.Schema) air.Expr {
	return &air.Constant{Value: e.Value}
}

// Lower a set of zero or more MIR expressions.
func lowerExprs(exprs []Expr, schema *air.Schema) []air.Expr {
	n := len(exprs)
	nexprs := make([]air.Expr, n)

	for i := 0; i < n; i++ {
		nexprs[i] = exprs[i].LowerTo(schema)
	}

	return nexprs
}
