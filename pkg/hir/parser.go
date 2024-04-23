package hir

import (
	"errors"
	"fmt"
	"strconv"
	"github.com/consensys/go-corset/pkg/sexp"
	"github.com/consensys/gnark-crypto/ecc/bls12-377/fr"
)

// ===================================================================
// Public
// ===================================================================

// Parse a string representing an HIR expression formatted using
// S-expressions.
func ParseSExp(s string) (Expr,error) {
	p := sexp.NewTranslator[Expr]()
	// Configure translator
	p.AddSymbolRule(sexpConstantToHir)
	p.AddSymbolRule(sexpColumnToHir)
	p.AddBinaryRule("shift", sexpShiftToHir)
	p.AddRecursiveRule("+", sexpAddToHir)
	p.AddRecursiveRule("-", sexpSubToHir)
	p.AddRecursiveRule("*", sexpMulToHir)
	p.AddRecursiveRule("~", sexpNormToHir)
	p.AddRecursiveRule("if", sexpIfToHir)
	// Parse string
	return p.Translate(s)
}

// ===================================================================
// Private
// ===================================================================

func sexpConstantToHir(symbol string) (Expr,error) {
	num := new(fr.Element)
	// Attempt to parse
	c,err := num.SetString(symbol)
	// Check for errors
	if err != nil { return nil,err }
	// Done
	return &Constant{c},nil
}

func sexpColumnToHir(col string) (Expr,error) {
	return &ColumnAccess{col,0},nil
}

func sexpAddToHir(args []Expr)(Expr,error) {
	return &Add{args},nil
}

func sexpSubToHir(args []Expr)(Expr,error) {
	return &Sub{args},nil
}

func sexpMulToHir(args []Expr)(Expr,error) {
	return &Mul{args},nil
}

func sexpIfToHir(args []Expr)(Expr,error) {
	if len(args) == 2 {
		return &IfZero{args[0],args[1],nil},nil
	} else if len(args) == 3 {
		return &IfZero{args[0],args[1],args[2]},nil
	} else {
		msg := fmt.Sprintf("Incorrect number of arguments: {%d}",len(args))
		return nil, errors.New(msg)
	}
}

func sexpShiftToHir(col string, amt string) (Expr,error) {
	n,err1 := strconv.Atoi(amt)
	if err1 != nil { return nil,err1 }
	return &ColumnAccess{col,n},nil
}

func sexpNormToHir(args []Expr) (Expr,error) {
	if len(args) != 1 {
		msg := fmt.Sprintf("Incorrect number of arguments: {%d}",len(args))
		return nil, errors.New(msg)
	} else {
		return &Normalise{args[0]}, nil
	}
}
