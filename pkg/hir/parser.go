package hir

import (
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/consensys/gnark-crypto/ecc/bls12-377/fr"
	"github.com/consensys/go-corset/pkg/sexp"
	"github.com/consensys/go-corset/pkg/table"
)

// ===================================================================
// Public
// ===================================================================

// ParseSExp parses a string representing an HIR expression formatted using
// S-expressions.
func ParseSExp(s string) (Expr, error) {
	p := newExprTranslator()
	// Parse string
	return p.ParseAndTranslate(s)
}

// ParseSchemaSExp parses a string representing an HIR schema formatted using
// S-expressions.
func ParseSchemaSExp(s string) (*Schema, error) {
	t := newExprTranslator()
	p := sexp.NewParser(s)
	// Construct initially empty schema
	schema := EmptySchema()
	// Continue parsing string until nothing remains.
	for {
		sRest, err := p.Parse()
		// Check for parsing error
		if err != nil {
			return nil, err
		}
		// Check whether complete
		if sRest == nil {
			return schema, nil
		}
		// Process declaration
		err = sexpDeclaration(sRest, schema, t)
		if err != nil {
			return nil, err
		}
	}
}

// ===================================================================
// Private
// ===================================================================

func newExprTranslator() *sexp.Translator[Expr] {
	p := sexp.NewTranslator[Expr]()
	// Configure translator
	p.AddSymbolRule(sexpConstant)
	p.AddSymbolRule(sexpColumnAccess)
	p.AddBinaryRule("shift", sexpShift)
	p.AddRecursiveRule("+", sexpAdd)
	p.AddRecursiveRule("-", sexpSub)
	p.AddRecursiveRule("*", sexpMul)
	p.AddRecursiveRule("~", sexpNorm)
	p.AddRecursiveRule("if", sexpIf)
	p.AddRecursiveRule("ifnot", sexpIfNot)

	return p
}

func sexpDeclaration(s sexp.SExp, schema *Schema, p *sexp.Translator[Expr]) error {
	if e, ok := s.(*sexp.List); ok {
		if e.Len() >= 2 && e.Len() <= 3 && e.MatchSymbols(2, "column") {
			return sexpColumn(e.Elements, schema)
		} else if e.Len() == 3 && e.MatchSymbols(2, "vanish") {
			return sexpVanishing(e.Elements, nil, schema, p)
		} else if e.Len() == 3 && e.MatchSymbols(2, "vanish:last") {
			domain := -1
			return sexpVanishing(e.Elements, &domain, schema, p)
		} else if e.Len() == 3 && e.MatchSymbols(2, "vanish:first") {
			domain := 0
			return sexpVanishing(e.Elements, &domain, schema, p)
		} else if e.Len() == 3 && e.MatchSymbols(2, "assert") {
			return sexpAssertion(e.Elements, schema, p)
		}
	}

	return fmt.Errorf("unexpected declaration: %s", s)
}

// Parse a column declaration
func sexpColumn(elements []sexp.SExp, schema *Schema) error {
	columnName := elements[1].String()

	var columnType table.Type = &table.FieldType{}

	if len(elements) == 3 {
		var err error
		columnType, err = sexpType(elements[2].String())

		if err != nil {
			return err
		}
	}

	schema.AddDataColumn(columnName, columnType)

	return nil
}

// Parse a property assertion
func sexpAssertion(elements []sexp.SExp, schema *Schema, p *sexp.Translator[Expr]) error {
	handle := elements[1].String()

	expr, err := p.Translate(elements[2])
	if err != nil {
		return err
	}
	// Add all assertions arising.
	for _, e := range expr.LowerTo() {
		schema.AddPropertyAssertion(handle, e)
	}

	return nil
}

// Parse a vanishing declaration
func sexpVanishing(elements []sexp.SExp, domain *int, schema *Schema, p *sexp.Translator[Expr]) error {
	handle := elements[1].String()

	expr, err := p.Translate(elements[2])
	if err != nil {
		return err
	}

	schema.AddVanishingConstraint(handle, domain, expr)

	return nil
}

func sexpType(symbol string) (table.Type, error) {
	if strings.HasPrefix(symbol, ":u") {
		n, err := strconv.Atoi(symbol[2:])
		if err != nil {
			return nil, err
		}

		return table.NewUintType(uint(n)), nil
	}

	return nil, fmt.Errorf("unexpected type: %s", symbol)
}

func sexpConstant(symbol string) (Expr, error) {
	num := new(fr.Element)
	// Attempt to parse
	c, err := num.SetString(symbol)
	// Check for errors
	if err != nil {
		return nil, err
	}
	// Done
	return &Constant{Val: c}, nil
}

func sexpColumnAccess(col string) (Expr, error) {
	return &ColumnAccess{col, 0}, nil
}

func sexpAdd(args []Expr) (Expr, error) {
	return &Add{args}, nil
}

func sexpSub(args []Expr) (Expr, error) {
	return &Sub{args}, nil
}

func sexpMul(args []Expr) (Expr, error) {
	return &Mul{args}, nil
}

func sexpIf(args []Expr) (Expr, error) {
	if len(args) == 2 {
		return &IfZero{args[0], args[1], nil}, nil
	} else if len(args) == 3 {
		return &IfZero{args[0], args[1], args[2]}, nil
	}

	return nil, fmt.Errorf("incorrect number of arguments: {%d}", len(args))
}

func sexpIfNot(args []Expr) (Expr, error) {
	if len(args) == 2 {
		return &IfZero{args[0], nil, args[1]}, nil
	}

	return nil, fmt.Errorf("incorrect number of arguments: {%d}", len(args))
}

func sexpShift(col string, amt string) (Expr, error) {
	n, err := strconv.Atoi(amt)

	if err != nil {
		return nil, err
	}

	return &ColumnAccess{
		Column: col,
		Shift:  n,
	}, nil
}

func sexpNorm(args []Expr) (Expr, error) {
	if len(args) != 1 {
		msg := fmt.Sprintf("Incorrect number of arguments: {%d}", len(args))
		return nil, errors.New(msg)
	}

	return &Normalise{Arg: args[0]}, nil
}
