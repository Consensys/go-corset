package mir

import (
	"errors"
	"fmt"
	"strconv"

	"github.com/consensys/gnark-crypto/ecc/bls12-377/fr"
	"github.com/consensys/go-corset/pkg/sexp"
)

// ParseSExpToMir parses a string representing an MIR expression formatted using
// S-expressions.
func ParseSExpToMir(s string) (Expr, error) {
	p := sexp.NewTranslator[Expr]()
	// Configure
	p.AddSymbolRule(sexpConstantToMir)
	p.AddSymbolRule(sexpColumnToMir)
	p.AddRecursiveRule("+", sexpAddToMir)
	p.AddRecursiveRule("-", sexpSubToMir)
	p.AddRecursiveRule("*", sexpMulToMir)
	p.AddRecursiveRule("~", sexpNormToMir)
	p.AddBinaryRule("shift", sexpShiftToMir)
	// Parse string
	return p.ParseAndTranslate(s)
}

func sexpConstantToMir(symbol string) (Expr, error) {
	num := new(fr.Element)
	// Attempt to parse
	c, err := num.SetString(symbol)
	// Check for errors
	if err != nil {
		return nil, err
	}
	// Done
	return &Constant{c}, nil
}

func sexpColumnToMir(col string) (Expr, error) {
	return &ColumnAccess{col, 0}, nil
}

func sexpAddToMir(args []Expr) (Expr, error) {
	return &Add{args}, nil
}

func sexpSubToMir(args []Expr) (Expr, error) {
	return &Sub{args}, nil
}

func sexpMulToMir(args []Expr) (Expr, error) {
	return &Mul{args}, nil
}

func sexpShiftToMir(col string, amt string) (Expr, error) {
	n, err1 := strconv.Atoi(amt)
	if err1 != nil {
		return nil, err1
	}

	return &ColumnAccess{col, n}, nil
}

func sexpNormToMir(args []Expr) (Expr, error) {
	if len(args) != 1 {
		msg := fmt.Sprintf("Incorrect number of arguments: {%d}", len(args))
		return nil, errors.New(msg)
	}

	return &Normalise{args[0]}, nil
}
