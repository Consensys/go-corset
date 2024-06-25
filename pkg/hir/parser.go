package hir

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"unicode"

	"github.com/consensys/gnark-crypto/ecc/bls12-377/fr"
	"github.com/consensys/go-corset/pkg/schema"
	sc "github.com/consensys/go-corset/pkg/schema"
	"github.com/consensys/go-corset/pkg/sexp"
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
	// Current module being parsed.
	module uint
	// Environment used during parsing to resolve column names into column
	// indices.
	env *Environment
}

func newHirParser(srcmap *sexp.SourceMap[sexp.SExp]) *hirParser {
	p := sexp.NewTranslator[Expr](srcmap)
	// Initialise empty environment
	env := EmptyEnvironment()
	// Register top-level module (aka the prelude)
	prelude := env.RegisterModule("")
	// Construct parser
	parser := &hirParser{p, EmptySchema(), prelude, env}
	// Configure translator
	p.AddSymbolRule(constantParserRule)
	p.AddSymbolRule(columnAccessParserRule(parser))
	p.AddBinaryRule("shift", shiftParserRule(parser))
	p.AddRecursiveRule("+", addParserRule)
	p.AddRecursiveRule("-", subParserRule)
	p.AddRecursiveRule("*", mulParserRule)
	p.AddRecursiveRule("~", normParserRule)
	p.AddRecursiveRule("if", ifParserRule)
	p.AddRecursiveRule("ifnot", ifNotParserRule)
	p.AddRecursiveRule("begin", beginParserRule)
	//
	return parser
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
			return p.parseSortedPermutationDeclaration(e)
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
	if p.env.HasColumn(p.module, columnName) {
		return p.translator.SyntaxError(l, "duplicate column declaration")
	}
	// Register column
	cid := p.env.RegisterColumn(p.module, columnName)
	// Default to field type
	var columnType sc.Type = &sc.FieldType{}
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
	p.schema.AddTypeConstraint(cid, columnType)

	return nil
}

// Parse a sorted permutation declaration
func (p *hirParser) parseSortedPermutationDeclaration(l *sexp.List) error {
	// Target columns are (sorted) permutations of source columns.
	sexpTargets := l.Elements[1].AsList()
	// Source columns.
	sexpSources := l.Elements[2].AsList()
	// Convert into appropriate form.
	targets := make([]schema.Column, sexpTargets.Len())
	sources := make([]string, sexpSources.Len())
	signs := make([]bool, sexpSources.Len())
	//
	if sexpTargets.Len() != sexpSources.Len() {
		return p.translator.SyntaxError(l, "sorted permutation requires matching number of source and target columns")
	}
	//
	for i := 0; i < sexpSources.Len(); i++ {
		source := sexpSources.Get(i).AsSymbol()
		target := sexpTargets.Get(i).AsSymbol()
		// Sanity check syntax as expected
		if source == nil {
			return p.translator.SyntaxError(sexpSources.Get(i), "malformed column")
		} else if target == nil {
			return p.translator.SyntaxError(sexpTargets.Get(i), "malformed column")
		}
		// Determine source column sign (i.e. sort direction)
		sortName := source.String()
		if strings.HasPrefix(sortName, "+") {
			signs[i] = true
		} else if strings.HasPrefix(sortName, "-") {
			if i == 0 {
				return p.translator.SyntaxError(source, "sorted permutation requires ascending first column")
			}

			signs[i] = false
		} else {
			return p.translator.SyntaxError(source, "malformed sort direction")
		}
		// Copy over column name
		sources[i] = sortName[1:]
		// FIXME: determine source column type
		targets[i] = schema.NewColumn(target.String(), &schema.FieldType{})
		// Sanity check that source column exists
		if !p.env.HasColumn(p.module, sources[i]) {
			// No, it doesn't.
			return p.translator.SyntaxError(sexpSources.Get(i), fmt.Sprintf("unknown column %s", sources[i]))
		}
		// Sanity check that target column *doesn't* exist.
		if p.env.HasColumn(p.module, targets[i].Name()) {
			// No, it doesn't.
			return p.translator.SyntaxError(sexpTargets.Get(i), fmt.Sprintf("duplicate column %s", targets[i].Name()))
		}
		// Finally, register target column
		p.env.RegisterColumn(p.module, targets[i].Name())
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
	// Add assertion.
	p.schema.AddPropertyAssertion(handle, expr)

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

func (p *hirParser) parseType(term sexp.SExp) (sc.Type, error) {
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
		// Done
		return sc.NewUintType(uint(n)), nil
	}
	// Error
	return nil, p.translator.SyntaxError(symbol, "unknown type")
}

func beginParserRule(args []Expr) (Expr, error) {
	return &List{args}, nil
}

func constantParserRule(symbol string) (Expr, bool, error) {
	if symbol[0] >= '0' && symbol[0] < '9' {
		num := new(fr.Element)
		// Attempt to parse
		c, err := num.SetString(symbol)
		// Check for errors
		if err != nil {
			return nil, true, err
		}
		// Done
		return &Constant{Val: c}, true, nil
	}
	// Not applicable
	return nil, false, nil
}

func columnAccessParserRule(parser *hirParser) func(col string) (Expr, bool, error) {
	// Returns a closure over the parser.
	return func(col string) (Expr, bool, error) {
		// Sanity check what we have
		if !unicode.IsLetter(rune(col[0])) {
			return nil, false, nil
		}
		// Look up column in the environment
		i, ok := parser.env.LookupColumn(parser.module, col)
		// Check column was found
		if !ok {
			return nil, true, fmt.Errorf("unknown column %s", col)
		}
		// Done
		return &ColumnAccess{i, 0}, true, nil
	}
}

func addParserRule(args []Expr) (Expr, error) {
	return &Add{args}, nil
}

func subParserRule(args []Expr) (Expr, error) {
	return &Sub{args}, nil
}

func mulParserRule(args []Expr) (Expr, error) {
	return &Mul{args}, nil
}

func ifParserRule(args []Expr) (Expr, error) {
	if len(args) == 2 {
		return &IfZero{args[0], args[1], nil}, nil
	} else if len(args) == 3 {
		return &IfZero{args[0], args[1], args[2]}, nil
	}

	return nil, errors.New("incorrect number of arguments")
}

func ifNotParserRule(args []Expr) (Expr, error) {
	if len(args) == 2 {
		return &IfZero{args[0], nil, args[1]}, nil
	}

	return nil, errors.New("incorrect number of arguments")
}

func shiftParserRule(parser *hirParser) func(string, string) (Expr, error) {
	// Returns a closure over the parser.
	return func(col string, amt string) (Expr, error) {
		n, err := strconv.Atoi(amt)

		if err != nil {
			return nil, err
		}
		// Look up column in the environment
		i, ok := parser.env.LookupColumn(parser.module, col)
		// Check column was found
		if !ok {
			return nil, fmt.Errorf("unknown column %s", col)
		}
		// Done
		return &ColumnAccess{
			Column: i,
			Shift:  n,
		}, nil
	}
}

func normParserRule(args []Expr) (Expr, error) {
	if len(args) != 1 {
		return nil, errors.New("incorrect number of arguments")
	}

	return &Normalise{Arg: args[0]}, nil
}
