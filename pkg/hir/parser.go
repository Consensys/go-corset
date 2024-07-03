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
	return p.env.schema, nil
}

// ===================================================================
// Private
// ===================================================================

type hirParser struct {
	// Translator used for recursive expressions.
	translator *sexp.Translator[Expr]
	// Current module being parsed.
	module uint
	// Environment used during parsing to resolve column names into column
	// indices.
	env *Environment
	// Global is used exclusively when parsing expressions to signal whether or
	// not qualified column accesses are permitted (i.e. which include a
	// module).
	global bool
}

func newHirParser(srcmap *sexp.SourceMap[sexp.SExp]) *hirParser {
	p := sexp.NewTranslator[Expr](srcmap)
	// Initialise empty environment
	env := EmptyEnvironment()
	// Register top-level module (aka the prelude)
	prelude := env.RegisterModule("")
	// Construct parser
	parser := &hirParser{p, prelude, env, false}
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
		if e.MatchSymbols(2, "module") {
			return p.parseModuleDeclaration(e)
		} else if e.MatchSymbols(2, "column") {
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
		} else if e.Len() == 4 && e.MatchSymbols(1, "lookup") {
			return p.parseLookupDeclaration(e)
		}
	}
	// Error
	return p.translator.SyntaxError(s, "unexpected declaration")
}

// Parse a column declaration
func (p *hirParser) parseModuleDeclaration(l *sexp.List) error {
	// Sanity check declaration
	if len(l.Elements) > 2 {
		return p.translator.SyntaxError(l, "malformed module declaration")
	}
	// Extract column name
	moduleName := l.Elements[1].String()
	// Sanity check doesn't already exist
	if p.env.HasModule(moduleName) {
		return p.translator.SyntaxError(l, "duplicate module declaration")
	}
	// Register module
	mid := p.env.RegisterModule(moduleName)
	// Set current module
	p.module = mid
	//
	return nil
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
	// Register column
	cid := p.env.AddDataColumn(p.module, columnName, columnType)
	p.env.schema.AddTypeConstraint(cid, columnType)

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
	sources := make([]uint, sexpSources.Len())
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

		sourceName := sortName[1:]
		targetName := target.String()
		// Determine index for source column
		sourceIndex, ok := p.env.LookupColumn(p.module, sourceName)
		if !ok {
			// Column doesn't exist!
			return p.translator.SyntaxError(sexpSources.Get(i), fmt.Sprintf("unknown column %s", sourceName))
		}
		// Sanity check that target column *doesn't* exist.
		if p.env.HasColumn(p.module, targetName) {
			// No, it doesn't.
			return p.translator.SyntaxError(sexpTargets.Get(i), fmt.Sprintf("duplicate column %s", targetName))
		}
		// Copy over column name
		sources[i] = sourceIndex
		// FIXME: determine source column type
		targets[i] = schema.NewColumn(p.module, targetName, &schema.FieldType{})
	}
	//
	p.env.AddPermutationColumns(p.module, targets, signs, sources)
	//
	return nil
}

// Parse a lookup declaration
func (p *hirParser) parseLookupDeclaration(l *sexp.List) error {
	handle := l.Elements[1].String()
	// Target columns are (sorted) permutations of source columns.
	sexpTargets := l.Elements[2].AsList()
	// Source columns.
	sexpSources := l.Elements[3].AsList()
	// Sanity check number of target colunms matches number of source columns.
	if sexpTargets.Len() != sexpSources.Len() {
		return p.translator.SyntaxError(l, "lookup constraint requires matching number of source and target columns")
	}
	// Sanity check expressions have unitary form.
	for i := 0; i < sexpTargets.Len(); i++ {
		// Sanity check source and target expressions do not contain expression
		// forms which are not permitted within a unitary expression.
		if err := p.checkUnitExpr(sexpTargets.Get(i)); err != nil {
			return err
		}

		if err := p.checkUnitExpr(sexpSources.Get(i)); err != nil {
			return err
		}
	}
	// Proceed with translation
	targets := make([]UnitExpr, sexpTargets.Len())
	sources := make([]UnitExpr, sexpSources.Len())
	// Lookup expressions are permitted to make fully qualified accesses.  This
	// is because inter-module lookups are supported.
	p.global = true
	// Parse source / target expressions
	for i := 0; i < len(targets); i++ {
		target, err1 := p.translator.Translate(sexpTargets.Get(i))
		source, err2 := p.translator.Translate(sexpSources.Get(i))

		if err1 != nil {
			return err1
		} else if err2 != nil {
			return err2
		}
		// Done
		targets[i] = UnitExpr{target}
		sources[i] = UnitExpr{source}
	}
	// Sanity check enclosing source and target modules
	source, err1 := schema.DetermineEnclosingModuleOfExpressions(sources, p.env.schema)
	target, err2 := schema.DetermineEnclosingModuleOfExpressions(targets, p.env.schema)
	// Propagate errors
	if err1 != nil {
		return p.translator.SyntaxError(sexpSources.Get(int(source)), err1.Error())
	} else if err2 != nil {
		return p.translator.SyntaxError(sexpTargets.Get(int(target)), err2.Error())
	}
	// Finally add constraint
	p.env.schema.AddLookupConstraint(handle, source, target, sources, targets)
	// DOne
	return nil
}

// Parse a property assertion
func (p *hirParser) parseAssertionDeclaration(elements []sexp.SExp) error {
	handle := elements[1].String()
	// Property assertions do not have global scope, hence qualified column
	// accesses are not permitted.
	p.global = false
	// Translate
	expr, err := p.translator.Translate(elements[2])
	if err != nil {
		return err
	}
	// Add assertion.
	p.env.schema.AddPropertyAssertion(p.module, handle, expr)

	return nil
}

// Parse a vanishing declaration
func (p *hirParser) parseVanishingDeclaration(elements []sexp.SExp, domain *int) error {
	handle := elements[1].String()
	// Vanishing constraints do not have global scope, hence qualified column
	// accesses are not permitted.
	p.global = false
	// Translate
	expr, err := p.translator.Translate(elements[2])
	if err != nil {
		return err
	}

	p.env.schema.AddVanishingConstraint(handle, p.module, domain, expr)

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

// Check that a given expression conforms to the requirements of a unitary
// expression.  That is, it cannot contain an "if", "ifnot" or "begin"
// expression form.
func (p *hirParser) checkUnitExpr(term sexp.SExp) error {
	l := term.AsList()

	if l != nil && l.Len() > 0 {
		if head := l.Get(0).AsSymbol(); head != nil {
			switch head.Value {
			case "if":
				fallthrough
			case "ifnot":
				fallthrough
			case "begin":
				return p.translator.SyntaxError(term, "not permitted in lookup")
			}
		}
		// Check arguments
		for i := 0; i < l.Len(); i++ {
			if err := p.checkUnitExpr(l.Get(i)); err != nil {
				return err
			}
		}
	}

	return nil
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
		var ok bool
		// Sanity check what we have
		if !unicode.IsLetter(rune(col[0])) {
			return nil, false, nil
		}
		// Handle qualified accesses (where permitted)
		module := parser.module
		colname := col
		// Attempt to split column name into module / column pair.
		split := strings.Split(col, ".")
		if parser.global && len(split) == 2 {
			// Lookup module
			if module, ok = parser.env.LookupModule(split[0]); !ok {
				return nil, true, errors.New("unknown module")
			}

			colname = split[1]
		} else if len(split) > 2 {
			return nil, true, errors.New("malformed column access")
		} else if len(split) == 2 {
			return nil, true, errors.New("qualified column access not permitted here")
		}
		// Now lookup column in the appropriate module.
		var cid uint
		// Look up column in the environment using local scope.
		cid, ok = parser.env.LookupColumn(module, colname)
		// Check column was found
		if !ok {
			return nil, true, errors.New("unknown column")
		}
		// Done
		return &ColumnAccess{cid, 0}, true, nil
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
