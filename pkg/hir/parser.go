package hir

import (
	"errors"
	"fmt"
	"strconv"
	"github.com/consensys/go-corset/pkg/sexp"
	"github.com/consensys/go-corset/pkg/table"
	"github.com/consensys/gnark-crypto/ecc/bls12-377/fr"
)

// ===================================================================
// Public
// ===================================================================

// Parse a string representing an HIR expression formatted using
// S-expressions.
func ParseSExp(s string) (Expr,error) {
	p := newExprTranslator()
	// Parse string
	return p.ParseAndTranslate(s)
}

// Parse a string representing an HIR schema formatted using
// S-expressions.
func ParseSchemaSExp(s string) (*Schema,error) {
	t := newExprTranslator()
	p := sexp.NewParser(s)
	// Construct initially empty schema
	schema := table.EmptySchema[Column,Constraint]()
	// Continue parsing string until nothing remains.
	for {
		s,err := p.Parse()
		// Check for parsing error
		if err != nil { return nil, err }
		// Check whether complete
		if s == nil { return schema,nil }
		// Process declaration
		err = sexpDeclaration(s,schema,t)
		if err != nil { return nil, err }
	}
}

// ===================================================================
// Private
// ===================================================================

func newExprTranslator() *sexp.Translator[Expr] {
	p := sexp.NewTranslator[Expr]()
	// Configure translator
	p.AddSymbolRule(sexpConstant)
	p.AddSymbolRule(sexpColumn)
	p.AddBinaryRule("shift", sexpShift)
	p.AddRecursiveRule("+", sexpAdd)
	p.AddRecursiveRule("-", sexpSub)
	p.AddRecursiveRule("*", sexpMul)
	p.AddRecursiveRule("~", sexpNorm)
	p.AddRecursiveRule("if", sexpIf)
	//
	return &p
}

func sexpDeclaration(s sexp.SExp, schema *Schema, p *sexp.Translator[Expr]) error {
	switch e := s.(type) {
	case *sexp.List:
		if e.Len() == 2 && e.MatchSymbols(2,"column") {
			columnName := e.Elements[1].String()
			schema.AddColumn(&DataColumn{columnName})
			return nil
		} else if e.Len() == 3 && e.MatchSymbols(2,"vanishing") {
			handle := e.Elements[1].String()
			expr,err := p.Translate(e.Elements[2])
			if err != nil { return err }
			schema.AddConstraint(&VanishingConstraint{Handle: handle, Expr: expr})
			return nil
		}
	}
	msg := fmt.Sprintf("Unexpected declaration: %s",s)
	return errors.New(msg)
}

func sexpConstant(symbol string) (Expr,error) {
	num := new(fr.Element)
	// Attempt to parse
	c,err := num.SetString(symbol)
	// Check for errors
	if err != nil { return nil,err }
	// Done
	return &Constant{c},nil
}

func sexpColumn(col string) (Expr,error) {
	return &ColumnAccess{col,0},nil
}

func sexpAdd(args []Expr)(Expr,error) {
	return &Add{args},nil
}

func sexpSub(args []Expr)(Expr,error) {
	return &Sub{args},nil
}

func sexpMul(args []Expr)(Expr,error) {
	return &Mul{args},nil
}

func sexpIf(args []Expr)(Expr,error) {
	if len(args) == 2 {
		return &IfZero{args[0],args[1],nil},nil
	} else if len(args) == 3 {
		return &IfZero{args[0],args[1],args[2]},nil
	} else {
		msg := fmt.Sprintf("Incorrect number of arguments: {%d}",len(args))
		return nil, errors.New(msg)
	}
}

func sexpShift(col string, amt string) (Expr,error) {
	n,err1 := strconv.Atoi(amt)
	if err1 != nil { return nil,err1 }
	return &ColumnAccess{col,n},nil
}

func sexpNorm(args []Expr) (Expr,error) {
	if len(args) != 1 {
		msg := fmt.Sprintf("Incorrect number of arguments: {%d}",len(args))
		return nil, errors.New(msg)
	} else {
		return &Normalise{args[0]}, nil
	}
}
