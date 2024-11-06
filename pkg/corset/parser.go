package corset

import (
	sc "github.com/consensys/go-corset/pkg/schema"
	"github.com/consensys/go-corset/pkg/sexp"
)

type Void = struct{}

// ===================================================================
// Public
// ===================================================================

// ParseSchemaString parses a sequence of zero or more HIR schema declarations
// represented as a string.  Internally, this uses sexp.ParseAll and
// ParseSchemaSExp to do the work.
func ParseSourceString(str string) (*Module, error) {
	parser := sexp.NewParser(str)
	// Parse bytes into an S-Expression
	terms, err := parser.ParseAll()
	// Check test file parsed ok
	if err != nil {
		return nil, err
	}
	// Parse terms into an HIR schema
	p, env := NewCorsetParser(parser.SourceMap())
	// Continue parsing string until nothing remains.
	for _, term := range terms {
		// Process declaration
		err2 := p.parseDeclaration(env, term)
		if err2 != nil {
			return nil, err2
		}
	}
	// Done
	return env.schema, nil
}

// ===================================================================
// Private
// ===================================================================

type CorsetParser struct {
	// Translator used for recursive expressions.
	translator *sexp.Translator[Void, Expr]
}

func NewCorsetParser(srcmap *sexp.SourceMap[sexp.SExp]) *CorsetParser {
	p := sexp.NewTranslator[Void, Expr](srcmap)
	// Construct parser
	parser := &CorsetParser{p}
	// Configure translator
	/* p.AddSymbolRule(constantParserRule)
	p.AddSymbolRule(varAccessParserRule(parser))
	p.AddSymbolRule(columnAccessParserRule(parser))
	p.AddBinaryRule("shift", shiftParserRule(parser))
	p.AddRecursiveRule("+", addParserRule)
	p.AddRecursiveRule("-", subParserRule)
	p.AddRecursiveRule("*", mulParserRule)
	p.AddRecursiveRule("~", normParserRule)
	p.AddRecursiveRule("^", powParserRule)
	p.AddRecursiveRule("if", ifParserRule)
	p.AddRecursiveRule("ifnot", ifNotParserRule)
	p.AddRecursiveRule("begin", beginParserRule)
	p.AddDefaultRecursiveRule(invokeParserRule) */
	//
	return parser
}

func (p *CorsetParser) parseDeclaration(env *Environment, s sexp.SExp) (Declaration, error) {
	if e, ok := s.(*sexp.List); ok {
		if e.MatchSymbols(2, "module") {
			return p.parseModuleDeclaration(env, e)
		} else if e.MatchSymbols(1, "defcolumns") {
			return p.parseColumnDeclarations(env, e)
		}
		/* else if e.Len() == 4 && e.MatchSymbols(2, "defconstraint") {
			return p.parseConstraintDeclaration(env, e.Elements)
		} else if e.Len() == 3 && e.MatchSymbols(2, "assert") {
			return p.parseAssertionDeclaration(env, e.Elements)
		} else if e.Len() == 3 && e.MatchSymbols(1, "defpermutation") {
			return p.parsePermutationDeclaration(env, e)
		} else if e.Len() == 4 && e.MatchSymbols(1, "deflookup") {
			return p.parseLookupDeclaration(env, e)
		} else if e.Len() == 3 && e.MatchSymbols(1, "definterleaved") {
			return p.parseInterleavingDeclaration(env, e)
		} else if e.Len() == 3 && e.MatchSymbols(1, "definrange") {
			return p.parseRangeDeclaration(env, e)
		} else if e.Len() == 3 && e.MatchSymbols(1, "defpurefun") {
			return p.parsePureFunDeclaration(env, e)
		} */
	}
	// Error
	return nil, p.translator.SyntaxError(s, "unexpected or malformed declaration")
}

// Parse a column declaration
func (p *CorsetParser) parseModuleDeclaration(env *Environment, l *sexp.List) (Declaration, error) {
	// Sanity check declaration
	if len(l.Elements) > 2 {
		return p.translator.SyntaxError(l, "malformed module declaration")
	}
	// Extract column name
	moduleName := l.Elements[1].AsSymbol().Value
	// Sanity check doesn't already exist
	if env.HasModule(moduleName) {
		return p.translator.SyntaxError(l, "duplicate module declaration")
	}
	// Register module
	mid := env.RegisterModule(moduleName)
	// Set current module
	p.module = mid
	//
	return nil
}

// Parse a column declaration
func (p *CorsetParser) parseColumnDeclarations(l *sexp.List) (*DefColumns, error) {
	// Sanity check declaration
	if len(l.Elements) == 1 {
		return nil, p.translator.SyntaxError(l, "malformed column declaration")
	}
	columns := make([]DefColumn, l.Len()-1)
	// Process column declarations one by one.
	for i := 1; i < len(l.Elements); i++ {
		decl, err := p.parseColumnDeclaration(l.Elements[i])
		// Extract column name
		if err != nil {
			return nil, err
		}
		columns[i] = decl
	}

	return &DefColumns{columns}, nil
}

func (p *CorsetParser) parseColumnDeclaration(e sexp.SExp) (DefColumn, error) {
	var defcolumn DefColumn
	// Default to field type
	defcolumn.DataType = &sc.FieldType{}
	// Check whether extended declaration or not.
	if l := e.AsList(); l != nil {
		// Check at least the name provided.
		if len(l.Elements) == 0 {
			return defcolumn, p.translator.SyntaxError(l, "empty column declaration")
		}
		// Column name is always first
		columnName = l.Elements[0].String(false)
		//	Parse type (if applicable)
		if len(l.Elements) == 2 {
			var err error
			if columnType, err = p.parseType(l.Elements[1]); err != nil {
				return defcolumn, err
			}
		} else if len(l.Elements) > 2 {
			// For now.
			return defcolumn, p.translator.SyntaxError(l, "unknown column declaration attributes")
		}
	} else {
		defcolumn.Name = e.String(false)
	}
	//
	return defcolumn, nil
}
