package corset

import (
	"sort"
	"strconv"
	"strings"

	sc "github.com/consensys/go-corset/pkg/schema"
	"github.com/consensys/go-corset/pkg/sexp"
)

// Void type represents an empty struct.
type Void = struct{}

// ===================================================================
// Public
// ===================================================================

// ParseSourceFiles parses zero or more source files producing zero or more
// modules.  Observe that, since a given module can be spread over multiple
// files, there can be far few modules created than there are source files. This
// function does more than just parse the individual files, because it
// additional combines all fragments of the same module together into one place.
// Thus, you should never expect to see duplicate module names in the returned
// array.
func ParseSourceFiles(files []string) ([]Module, []error) {
	var errors []error = make([]error, len(files))
	var num_errs uint
	// Contents map holds the combined fragments of each module.
	contents := make(map[string]Module, 0)
	// Names identifies the names of each unique module.
	names := make([]string, 0)
	//
	for i, file := range files {
		mods, err := ParseSourceFile(file)
		// Handle errors
		if err != nil {
			num_errs++
		}
		// Report any errors encountered
		errors[i] = err
		// Allocate any module fragments
		for _, m := range mods {
			if om, ok := contents[m.Name]; !ok {
				contents[m.Name] = m
				names = append(names, m.Name)
			} else {
				om.Declarations = append(om.Declarations, m.Declarations...)
				contents[m.Name] = om
			}
		}
	}
	// Bring all fragmenmts together
	modules := make([]Module, len(names))
	// Sort module names to ensure that compilation is always deterministic.
	sort.Strings(names)
	// Finalise every module
	for i, n := range names {
		// Assume this cannot fail as every module in names has been assigned at
		// least one fragment.
		modules[i] = contents[n]
	}
	// Done
	if num_errs > 0 {
		return modules, errors
	}
	// no errors
	return modules, nil
}

// ParseSourceFile parses the contents of a single lisp file into one or more
// modules.  Observe that every lisp file starts in the "prelude" or "root"
// module, and may declare items for additional modules as necessary.
func ParseSourceFile(file string) ([]Module, error) {
	parser := sexp.NewParser(file)
	// Parse bytes into an S-Expression
	terms, err := parser.ParseAll()
	// Check test file parsed ok
	if err != nil {
		return nil, err
	}
	// Construct parser for corset syntax
	p := NewCorsetParser(parser.SourceMap())
	// Initially empty set of modules
	var modules []Module

	var contents []Declaration
	// Parse whatever is declared at the beginning of the file before the first
	// module declaration.  These declarations form part of the "prelude".
	if contents, terms, err = p.parseModuleContents("", terms); err != nil {
		return nil, err
	} else if len(contents) != 0 {
		modules = append(modules, Module{"", contents})
	}
	// Continue parsing string until nothing remains.
	for len(terms) != 0 {
		var name string
		// Extract module name
		if name, err = p.parseModuleStart(terms[0]); err != nil {
			return nil, err
		}
		// Parse module contents
		if contents, terms, err = p.parseModuleContents(name, terms[1:]); err != nil {
			return nil, err
		} else if len(contents) != 0 {
			modules = append(modules, Module{"", contents})
		}
	}
	// Done
	return modules, nil
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

// Extract all declarations associated with a given module and package them up.
func (p *CorsetParser) parseModuleContents(name string, terms []sexp.SExp) ([]Declaration, []sexp.SExp, error) {
	//
	decls := make([]Declaration, 0)
	//
	for i, s := range terms {
		e, ok := s.(*sexp.List)
		// Check for error
		if !ok {
			return nil, nil, p.translator.SyntaxError(s, "unexpected or malformed declaration")
		}
		// Check for end-of-module
		if e.MatchSymbols(2, "module") {
			return decls, terms[i:], nil
		}
		// Parse the declaration
		if decl, err := p.parseDeclaration(e); err != nil {
			return nil, nil, err
		} else {
			// Continue accumulating declarations for this module.
			decls = append(decls, decl)
		}
	}
	// End-of-file signals end-of-module.
	return decls, make([]sexp.SExp, 0), nil
}

// Parse a module declaration of the form "(module m1)" which indicates the
// start of module m1.
func (p *CorsetParser) parseModuleStart(s sexp.SExp) (string, error) {
	l, ok := s.(*sexp.List)
	// Check for error
	if !ok {
		return "", p.translator.SyntaxError(s, "unexpected or malformed declaration")
	}
	// Sanity check declaration
	if len(l.Elements) > 2 {
		return "", p.translator.SyntaxError(l, "malformed module declaration")
	}
	// Extract column name
	name := l.Elements[1].AsSymbol().Value
	//
	return name, nil
}

func (p *CorsetParser) parseDeclaration(s *sexp.List) (Declaration, error) {
	if s.MatchSymbols(1, "defcolumns") {
		return p.parseColumnDeclarations(s)
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
	return nil, p.translator.SyntaxError(s, "malformed module declaration")
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
		columns[i-1] = decl
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
		defcolumn.Name = l.Elements[0].String(false)
		//	Parse type (if applicable)
		if len(l.Elements) == 2 {
			var err error
			if defcolumn.DataType, err = p.parseType(l.Elements[1]); err != nil {
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

func (p *CorsetParser) parseType(term sexp.SExp) (sc.Type, error) {
	symbol := term.AsSymbol()
	if symbol == nil {
		return nil, p.translator.SyntaxError(term, "malformed column")
	}
	// Access string of symbol
	str := symbol.Value
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
