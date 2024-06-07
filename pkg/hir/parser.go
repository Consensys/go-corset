package hir

import (
	"errors"
	"strconv"
	"strings"

	"github.com/consensys/gnark-crypto/ecc/bls12-377/fr"
	"github.com/consensys/go-corset/pkg/sexp"
	"github.com/consensys/go-corset/pkg/table"
)

// ===================================================================
// Public
// ===================================================================

// ParseSchemaString parses a sequence of zero or more HIR schema declarations
// represented as a string.  Internally, this uses sexp.ParseAll and
// ParseSchemaSExp to do the work.
func ParseSchemaString(str string) (*Schema, error) {
	parser := sexp.NewParser(str)
	// Parse bytes into an S-Expression
	terms, err := parser.ParseAll()
	// Check test file parsed ok
	if err != nil {
		return nil, err
	}
	// Parse terms into an HIR schema
	p := newHirParser(parser.SourceMap())
	// Continue parsing string until nothing remains.
	for _, term := range terms {
		// Process declaration
		err2 := p.parseDeclaration(term)
		if err2 != nil {
			return nil, err2
		}
	}
	// Done
	return p.schema, nil
}

// ===================================================================
// Private
// ===================================================================

type hirParser struct {
	// Translator used for recursive expressions.
	translator *sexp.Translator[Expr]
	// Schema being constructed
	schema *Schema
}

func newHirParser(srcmap *sexp.SourceMap[sexp.SExp]) *hirParser {
	p := sexp.NewTranslator[Expr](srcmap)
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
	p.AddRecursiveRule("begin", sexpBegin)
	//
	return &hirParser{p, EmptySchema()}
}

func (p *hirParser) parseDeclaration(s sexp.SExp) error {
	if e, ok := s.(*sexp.List); ok {
		if e.MatchSymbols(2, "column") {
			return p.parseColumnDeclaration(e)
		} else if e.Len() == 3 && e.MatchSymbols(2, "vanish") {
			return p.parseVanishingDeclaration(e.Elements, nil)
		} else if e.Len() == 3 && e.MatchSymbols(2, "vanish:last") {
			domain := -1
			return p.parseVanishingDeclaration(e.Elements, &domain)
		} else if e.Len() == 3 && e.MatchSymbols(2, "vanish:first") {
			domain := 0
			return p.parseVanishingDeclaration(e.Elements, &domain)
		} else if e.Len() == 3 && e.MatchSymbols(2, "assert") {
			return p.parseAssertionDeclaration(e.Elements)
		} else if e.Len() == 3 && e.MatchSymbols(1, "permute") {
			return p.parseSortedPermutationDeclaration(e.Elements)
		}
	}
	// Error
	return p.translator.SyntaxError(s, "unexpected declaration")
}

// Parse a column declaration
func (p *hirParser) parseColumnDeclaration(l *sexp.List) error {
	// Sanity check declaration
	if len(l.Elements) > 3 {
		return p.translator.SyntaxError(l, "malformed column declaration")
	}
	// Extract column name
	columnName := l.Elements[1].String()
	// Sanity check doesn't already exist
	if p.schema.HasColumn(columnName) {
		return p.translator.SyntaxError(l, "duplicate column declaration")
	}
	// Default to field type
	var columnType table.Type = &table.FieldType{}
	// Parse type (if applicable)
	if len(l.Elements) == 3 {
		var err error
		columnType, err = p.parseType(l.Elements[2])

		if err != nil {
			return err
		}
	}
	// Register column in Schema
	p.schema.AddDataColumn(columnName, columnType)

	return nil
}

// Parse a sorted permutation declaration
func (p *hirParser) parseSortedPermutationDeclaration(elements []sexp.SExp) error {
	// Target columns are (sorted) permutations of source columns.
	sexpTargets := elements[1].AsList()
	// Source columns.
	sexpSources := elements[2].AsList()
	// Convert into appropriate form.
	targets := make([]string, sexpTargets.Len())
	sources := make([]string, sexpSources.Len())
	signs := make([]bool, sexpSources.Len())
	//
	for i := 0; i < sexpTargets.Len(); i++ {
		target := sexpTargets.Get(i).AsSymbol()
		// Sanity check syntax as expected
		if target == nil {
			return p.translator.SyntaxError(sexpTargets.Get(i), "malformed column")
		}
		// Copy over
		targets[i] = target.String()
	}
	//
	for i := 0; i < sexpSources.Len(); i++ {
		source := sexpSources.Get(i).AsSymbol()
		// Sanity check syntax as expected
		if source == nil {
			return p.translator.SyntaxError(sexpSources.Get(i), "malformed column")
		}
		// Determine source column sign (i.e. sort direction)
		sortName := source.String()
		if strings.HasPrefix(sortName, "+") {
			signs[i] = true
		} else if strings.HasPrefix(sortName, "-") {
			signs[i] = false
		} else {
			return p.translator.SyntaxError(source, "malformed sort direction")
		}
		// Copy over column name
		sources[i] = sortName[1:]
	}
	//
	p.schema.AddPermutationColumns(targets, signs, sources)
	//
	return nil
}

// Parse a property assertion
func (p *hirParser) parseAssertionDeclaration(elements []sexp.SExp) error {
	handle := elements[1].String()

	expr, err := p.translator.Translate(elements[2])
	if err != nil {
		return err
	}
	// Add all assertions arising.
	for _, e := range expr.LowerTo() {
		p.schema.AddPropertyAssertion(handle, e)
	}

	return nil
}

// Parse a vanishing declaration
func (p *hirParser) parseVanishingDeclaration(elements []sexp.SExp, domain *int) error {
	handle := elements[1].String()

	expr, err := p.translator.Translate(elements[2])
	if err != nil {
		return err
	}

	p.schema.AddVanishingConstraint(handle, domain, expr)

	return nil
}

func (p *hirParser) parseType(term sexp.SExp) (table.Type, error) {
	symbol := term.AsSymbol()
	if symbol == nil {
		return nil, p.translator.SyntaxError(term, "malformed column")
	}
	// Access string of symbol
	str := symbol.String()
	if strings.HasPrefix(str, ":u") {
		n, err := strconv.Atoi(str[2:])
		if err != nil {
			return nil, err
		}
		// TODO: support @prove
		return table.NewUintType(uint(n), true), nil
	}
	// Error
	return nil, p.translator.SyntaxError(symbol, "unknown type")
}

func sexpBegin(args []Expr) (Expr, error) {
	return &List{args}, nil
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

	return nil, errors.New("incorrect number of arguments")
}

func sexpIfNot(args []Expr) (Expr, error) {
	if len(args) == 2 {
		return &IfZero{args[0], nil, args[1]}, nil
	}

	return nil, errors.New("incorrect number of arguments")
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
		return nil, errors.New("incorrect number of arguments")
	}

	return &Normalise{Arg: args[0]}, nil
}
