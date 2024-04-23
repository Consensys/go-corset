package air

import (
	"strconv"
	"github.com/consensys/go-corset/pkg/sexp"
	"github.com/consensys/gnark-crypto/ecc/bls12-377/fr"
)

// ============================================================================
// Public
// ============================================================================

// Parse a string representing an AIR expression formatted using
// S-expressions.
func ParseSExp(s string) (Expr, error) {
	p := sexp.NewTranslator[Expr]()
	// Configure
	p.AddSymbolRule(sexpConstantToAir)
	p.AddSymbolRule(sexpColumnToAir)
	p.AddRecursiveRule("+", sexpAddToAir)
	p.AddRecursiveRule("-", sexpSubToAir)
	p.AddRecursiveRule("*", sexpMulToAir)
	p.AddBinaryRule("shift", sexpShiftToAir)
	// Parse string
	return p.Translate(s)
}

// ============================================================================
// Private
// ============================================================================

func sexpConstantToAir(symbol string) (Expr, error) {
	num := new(fr.Element)
	// Attempt to parse
	c,err := num.SetString(symbol)
	// Check for errors
	if err != nil { return nil,err }
	// Done
	return &Constant{c},nil
}

func sexpColumnToAir(col string) (Expr, error) {
	return &ColumnAccess{col,0},nil
}

func sexpAddToAir(args []Expr) (Expr, error) {
	return &Add{args}, nil
}

func sexpSubToAir(args []Expr) (Expr, error) {
	return &Sub{args}, nil
}

func sexpMulToAir(args []Expr) (Expr, error) {
	return &Mul{args}, nil
}

func sexpShiftToAir(col string, amt string) (Expr, error) {
	n,err1 := strconv.Atoi(amt)
	if err1 != nil { return nil,err1 }
	return &ColumnAccess{col,n},nil
}
