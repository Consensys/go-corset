package corset

import (
	"errors"
	"math/big"
	"sort"
	"strconv"
	"strings"
	"unicode"

	"github.com/consensys/gnark-crypto/ecc/bls12-377/fr"
	sc "github.com/consensys/go-corset/pkg/schema"
	"github.com/consensys/go-corset/pkg/sexp"
)

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
func ParseSourceFiles(files []*sexp.SourceFile) (Circuit, *sexp.SourceMaps[Node], []SyntaxError) {
	var circuit Circuit
	// (for now) at most one error per source file is supported.
	var errors []SyntaxError = make([]SyntaxError, len(files))
	// Construct an initially empty source map
	srcmaps := sexp.NewSourceMaps[Node]()
	// num_errs counts the number of errors reported
	var num_errs uint
	// Contents map holds the combined fragments of each module.
	contents := make(map[string]Module, 0)
	// Names identifies the names of each unique module.
	names := make([]string, 0)
	//
	for i, file := range files {
		c, srcmap, err := ParseSourceFile(file)
		// Handle errors
		if err != nil {
			num_errs++
			// Report any errors encountered
			errors[i] = *err
		} else {
			// Combine source maps
			srcmaps.Join(srcmap)
		}
		// Update top-level declarations
		circuit.Declarations = append(circuit.Declarations, c.Declarations...)
		// Allocate any module fragments
		for _, m := range c.Modules {
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
	circuit.Modules = make([]Module, len(names))
	// Sort module names to ensure that compilation is always deterministic.
	sort.Strings(names)
	// Finalise every module
	for i, n := range names {
		// Assume this cannot fail as every module in names has been assigned at
		// least one fragment.
		circuit.Modules[i] = contents[n]
	}
	// Done
	if num_errs > 0 {
		return circuit, srcmaps, errors
	}
	// no errors
	return circuit, srcmaps, nil
}

// ParseSourceFile parses the contents of a single lisp file into one or more
// modules.  Observe that every lisp file starts in the "prelude" or "root"
// module, and may declare items for additional modules as necessary.
func ParseSourceFile(srcfile *sexp.SourceFile) (Circuit, *sexp.SourceMap[Node], *SyntaxError) {
	var circuit Circuit
	// Parse bytes into an S-Expression
	terms, srcmap, err := srcfile.ParseAll()
	// Check test file parsed ok
	if err != nil {
		return circuit, nil, err
	}
	// Construct parser for corset syntax
	p := NewParser(srcfile, srcmap)
	// Parse whatever is declared at the beginning of the file before the first
	// module declaration.  These declarations form part of the "prelude".
	if circuit.Declarations, terms, err = p.parseModuleContents(terms); err != nil {
		return circuit, nil, err
	}
	// Continue parsing string until nothing remains.
	for len(terms) != 0 {
		var (
			name  string
			decls []Declaration
		)
		// Extract module name
		if name, err = p.parseModuleStart(terms[0]); err != nil {
			return circuit, nil, err
		}
		// Parse module contents
		if decls, terms, err = p.parseModuleContents(terms[1:]); err != nil {
			return circuit, nil, err
		} else if len(decls) != 0 {
			circuit.Modules = append(circuit.Modules, Module{name, decls})
		}
	}
	// Done
	return circuit, p.NodeMap(), nil
}

// Parser implements a simple parser for the Corset language.  The parser itself
// is relatively simplistic and simply packages up the relevant lisp constructs
// into their corresponding AST forms.  This can fail in various ways, such as
// e.g. a "defconstraint" not having exactly three arguments, etc.  However, the
// parser does not attempt to perform more complex forms of validation (e.g.
// ensuring that expressions are well-typed, etc) --- that is left up to the
// compiler.
type Parser struct {
	// Translator used for recursive expressions.
	translator *sexp.Translator[Expr]
	// Mapping from constructed S-Expressions to their spans in the original text.
	nodemap *sexp.SourceMap[Node]
}

// NewParser constructs a new parser using a given mapping from S-Expressions to
// spans in the underlying source file.
func NewParser(srcfile *sexp.SourceFile, srcmap *sexp.SourceMap[sexp.SExp]) *Parser {
	p := sexp.NewTranslator[Expr](srcfile, srcmap)
	// Construct (initially empty) node map
	nodemap := sexp.NewSourceMap[Node](srcmap.Source())
	// Construct parser
	parser := &Parser{p, nodemap}
	// Configure expression translator
	p.AddSymbolRule(constantParserRule)
	p.AddSymbolRule(varAccessParserRule)
	p.AddBinaryRule("shift", shiftParserRule)
	p.AddRecursiveRule("+", addParserRule)
	p.AddRecursiveRule("-", subParserRule)
	p.AddRecursiveRule("*", mulParserRule)
	p.AddRecursiveRule("~", normParserRule)
	p.AddRecursiveRule("^", powParserRule)
	p.AddRecursiveRule("if", ifParserRule)
	p.AddRecursiveRule("begin", beginParserRule)
	//
	return parser
}

// NodeMap extract the node map constructec by this parser.  A key task here is
// to copy all mappings from the expression translator, which maintains its own
// map.
func (p *Parser) NodeMap() *sexp.SourceMap[Node] {
	// Copy all mappings from translator's source map into this map.  A mapping
	// function is required to coerce the types.
	sexp.JoinMaps(p.nodemap, p.translator.SourceMap(), func(e Expr) Node { return e })
	// Done
	return p.nodemap
}

// Register a source mapping from a given S-Expression to a given target node.
func (p *Parser) mapSourceNode(from sexp.SExp, to Node) {
	span := p.translator.SpanOf(from)
	p.nodemap.Put(to, span)
}

// Extract all declarations associated with a given module and package them up.
func (p *Parser) parseModuleContents(terms []sexp.SExp) ([]Declaration, []sexp.SExp, *SyntaxError) {
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
func (p *Parser) parseModuleStart(s sexp.SExp) (string, *SyntaxError) {
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

func (p *Parser) parseDeclaration(s *sexp.List) (Declaration, *SyntaxError) {
	var (
		decl  Declaration
		error *SyntaxError
	)
	//
	if s.MatchSymbols(1, "defcolumns") {
		decl, error = p.parseDefColumns(s)
	} else if s.Len() == 4 && s.MatchSymbols(2, "defconstraint") {
		decl, error = p.parseDefConstraint(s.Elements)
	} else if s.Len() == 3 && s.MatchSymbols(1, "definrange") {
		decl, error = p.parseDefInRange(s.Elements)
	} else if s.Len() == 3 && s.MatchSymbols(1, "definterleaved") {
		decl, error = p.parseDefInterleaved(s.Elements)
	} else if s.Len() == 4 && s.MatchSymbols(1, "deflookup") {
		decl, error = p.parseDefLookup(s.Elements)
	} else if s.Len() == 3 && s.MatchSymbols(2, "defpermutation") {
		decl, error = p.parseDefPermutation(s.Elements)
	} else if s.Len() == 3 && s.MatchSymbols(2, "defproperty") {
		decl, error = p.parseDefProperty(s.Elements)
	} else {
		error = p.translator.SyntaxError(s, "malformed declaration")
	}
	// Register node if appropriate
	if decl != nil {
		p.mapSourceNode(s, decl)
	}
	// done
	return decl, error
}

// Parse a column declaration
func (p *Parser) parseDefColumns(l *sexp.List) (*DefColumns, *SyntaxError) {
	columns := make([]*DefColumn, l.Len()-1)
	// Sanity check declaration
	if len(l.Elements) == 1 {
		return nil, p.translator.SyntaxError(l, "malformed column declaration")
	}
	// Process column declarations one by one.
	for i := 1; i < len(l.Elements); i++ {
		decl, err := p.parseColumnDeclaration(l.Elements[i])
		// Extract column name
		if err != nil {
			return nil, err
		}
		// Assign the declaration
		columns[i-1] = decl
	}
	// Done
	return &DefColumns{columns}, nil
}

func (p *Parser) parseColumnDeclaration(e sexp.SExp) (*DefColumn, *SyntaxError) {
	defcolumn := &DefColumn{"", nil, false, 1}
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
			var err *SyntaxError
			if defcolumn.DataType, defcolumn.MustProve, err = p.parseType(l.Elements[1]); err != nil {
				return defcolumn, err
			}
		} else if len(l.Elements) > 2 {
			// For now.
			return defcolumn, p.translator.SyntaxError(l, "unknown column declaration attributes")
		}
	} else {
		defcolumn.Name = e.String(false)
	}
	// Update source mapping
	p.mapSourceNode(e, defcolumn)
	//
	return defcolumn, nil
}

// Parse a vanishing declaration
func (p *Parser) parseDefConstraint(elements []sexp.SExp) (*DefConstraint, *SyntaxError) {
	// Initial sanity checks
	if elements[1].AsSymbol() == nil {
		return nil, p.translator.SyntaxError(elements[1], "expected constraint handle")
	}
	//
	handle := elements[1].AsSymbol().Value
	// Vanishing constraints do not have global scope, hence qualified column
	// accesses are not permitted.
	domain, guard, err := p.parseConstraintAttributes(elements[2])
	// Check for error
	if err != nil {
		return nil, err
	}
	// Translate expression
	expr, err := p.translator.Translate(elements[3])
	if err != nil {
		return nil, err
	}
	// Done
	return &DefConstraint{handle, domain, guard, expr}, nil
}

// Parse a interleaved declaration
func (p *Parser) parseDefInterleaved(elements []sexp.SExp) (*DefInterleaved, *SyntaxError) {
	// Initial sanity checks
	if elements[1].AsSymbol() == nil {
		return nil, p.translator.SyntaxError(elements[1], "malformed target column")
	} else if elements[2].AsList() == nil {
		return nil, p.translator.SyntaxError(elements[2], "malformed source columns")
	}
	// Extract target and source columns
	target := elements[1].AsSymbol().Value
	sexpSources := elements[2].AsList()
	sources := make([]string, sexpSources.Len())
	//
	for i := 0; i != sexpSources.Len(); i++ {
		ith := sexpSources.Get(i)
		if ith.AsSymbol() == nil {
			return nil, p.translator.SyntaxError(ith, "malformed source column")
		}
		// Extract column name
		sources[i] = ith.AsSymbol().Value
	}
	// Done
	return &DefInterleaved{target, sources}, nil
}

// Parse a lookup declaration
func (p *Parser) parseDefLookup(elements []sexp.SExp) (*DefLookup, *SyntaxError) {
	// Initial sanity checks
	if elements[1].AsSymbol() == nil {
		return nil, p.translator.SyntaxError(elements[1], "malformed handle")
	} else if elements[2].AsList() == nil {
		return nil, p.translator.SyntaxError(elements[2], "malformed target columns")
	} else if elements[3].AsList() == nil {
		return nil, p.translator.SyntaxError(elements[3], "malformed source columns")
	}
	// Extract items
	handle := elements[1].AsSymbol().Value
	sexpTargets := elements[2].AsList()
	sexpSources := elements[3].AsList()
	// Sanity check number of columns matches
	if sexpTargets.Len() != sexpSources.Len() {
		return nil, p.translator.SyntaxError(elements[3], "incorrect number of columns")
	}

	sources := make([]Expr, sexpSources.Len())
	targets := make([]Expr, sexpTargets.Len())
	// Translate source & target expressions
	for i := 0; i < sexpTargets.Len(); i++ {
		var err *SyntaxError
		// Translate source expressions
		if sources[i], err = p.translator.Translate(sexpSources.Get(i)); err != nil {
			return nil, err
		}
		// Translate target expressions
		if targets[i], err = p.translator.Translate(sexpTargets.Get(i)); err != nil {
			return nil, err
		}
	}
	// Done
	return &DefLookup{handle, sources, targets}, nil
}

// Parse a permutation declaration
func (p *Parser) parseDefPermutation(elements []sexp.SExp) (*DefPermutation, *SyntaxError) {
	var err *SyntaxError
	//
	sexpTargets := elements[1].AsList()
	sexpSources := elements[2].AsList()
	// Initial sanity checks
	if sexpTargets == nil {
		return nil, p.translator.SyntaxError(elements[1], "malformed target columns")
	} else if sexpSources == nil {
		return nil, p.translator.SyntaxError(elements[2], "malformed source columns")
	} else if sexpTargets.Len() < sexpSources.Len() {
		return nil, p.translator.SyntaxError(elements[1], "too few target columns")
	} else if sexpTargets.Len() > sexpSources.Len() {
		return nil, p.translator.SyntaxError(elements[1], "too many target columns")
	}
	//
	targets := make([]*DefColumn, sexpTargets.Len())
	sources := make([]*DefPermutedColumn, sexpSources.Len())
	//
	for i := 0; i < len(targets); i++ {
		// Parse target column
		if targets[i], err = p.parseColumnDeclaration(sexpTargets.Get(i)); err != nil {
			return nil, err
		}
		// Parse source column
		if sources[i], err = p.parsePermutedColumnDeclaration(sexpSources.Get(i)); err != nil {
			return nil, err
		}
	}
	//
	return &DefPermutation{targets, sources}, nil
}

func (p *Parser) parsePermutedColumnDeclaration(e sexp.SExp) (*DefPermutedColumn, *SyntaxError) {
	var err *SyntaxError
	//
	defcolumn := &DefPermutedColumn{"", false}
	// Check whether extended declaration or not.
	if l := e.AsList(); l != nil {
		// Check at least the name provided.
		if len(l.Elements) == 0 {
			return defcolumn, p.translator.SyntaxError(l, "empty permutation column")
		} else if len(l.Elements) != 2 {
			return defcolumn, p.translator.SyntaxError(l, "malformed permutation column")
		} else if l.Get(0).AsSymbol() == nil || l.Get(1).AsSymbol() == nil {
			return defcolumn, p.translator.SyntaxError(l, "empty permutation column")
		}
		// Parse sign
		if defcolumn.Sign, err = p.parsePermutedColumnSign(l.Get(0).AsSymbol()); err != nil {
			return nil, err
		}
		// Parse column name
		defcolumn.Name = l.Get(1).AsSymbol().Value
	} else {
		defcolumn.Name = e.String(false)
	}
	// Update source mapping
	p.mapSourceNode(e, defcolumn)
	//
	return defcolumn, nil
}

func (p *Parser) parsePermutedColumnSign(sign *sexp.Symbol) (bool, *SyntaxError) {
	switch sign.Value {
	case "+", "↓":
		return true, nil
	case "-", "↑":
		return false, nil
	default:
		return false, p.translator.SyntaxError(sign, "malformed sort direction")
	}
}

// Parse a property assertion
func (p *Parser) parseDefProperty(elements []sexp.SExp) (*DefProperty, *SyntaxError) {
	// Initial sanity checks
	if elements[1].AsSymbol() == nil {
		return nil, p.translator.SyntaxError(elements[1], "expected constraint handle")
	}
	//
	handle := elements[1].AsSymbol().Value
	// Translate expression
	expr, err := p.translator.Translate(elements[2])
	if err != nil {
		return nil, err
	}
	// Done
	return &DefProperty{handle, expr}, nil
}

// Parse a range declaration
func (p *Parser) parseDefInRange(elements []sexp.SExp) (*DefInRange, *SyntaxError) {
	var bound fr.Element
	// Translate expression
	expr, err := p.translator.Translate(elements[1])
	if err != nil {
		return nil, err
	}
	// Check & parse bound
	if elements[2].AsSymbol() == nil {
		return nil, p.translator.SyntaxError(elements[2], "malformed bound")
	} else if _, err := bound.SetString(elements[2].AsSymbol().Value); err != nil {
		return nil, p.translator.SyntaxError(elements[2], "malformed bound")
	}
	// Done
	return &DefInRange{Expr: expr, Bound: bound}, nil
}

func (p *Parser) parseConstraintAttributes(attributes sexp.SExp) (domain *int, guard Expr, err *SyntaxError) {
	// Check attribute list is a list
	if attributes.AsList() == nil {
		return nil, nil, p.translator.SyntaxError(attributes, "expected attribute list")
	}
	// Deconstruct as list
	attrs := attributes.AsList()
	// Process each attribute in turn
	for i := 0; i < attrs.Len(); i++ {
		ith := attrs.Get(i)
		// Check start of attribute
		if ith.AsSymbol() == nil {
			return nil, nil, p.translator.SyntaxError(ith, "malformed attribute")
		}
		// Check what we've got
		switch ith.AsSymbol().Value {
		case ":domain":
			i++
			if domain, err = p.parseDomainAttribute(attrs.Get(i)); err != nil {
				return nil, nil, err
			}
		case ":guard":
			i++
			if guard, err = p.translator.Translate(attrs.Get(i)); err != nil {
				return nil, nil, err
			}
		default:
			return nil, nil, p.translator.SyntaxError(ith, "unknown attribute")
		}
	}
	// Done
	return domain, guard, nil
}

func (p *Parser) parseDomainAttribute(attribute sexp.SExp) (domain *int, err *SyntaxError) {
	if attribute.AsSet() == nil {
		return nil, p.translator.SyntaxError(attribute, "malformed domain set")
	}
	// Sanity check
	set := attribute.AsSet()
	// Check all domain elements well-formed.
	for i := 0; i < set.Len(); i++ {
		ith := set.Get(i)
		if ith.AsSymbol() == nil {
			return nil, p.translator.SyntaxError(ith, "malformed domain")
		}
	}
	// Currently, only support domains of size 1.
	if set.Len() == 1 {
		first, err := strconv.Atoi(set.Get(0).AsSymbol().Value)
		// Check for parse error
		if err != nil {
			return nil, p.translator.SyntaxError(set.Get(0), "malformed domain element")
		}
		// Done
		return &first, nil
	}
	// Fail
	return nil, p.translator.SyntaxError(attribute, "multiple values not supported")
}

func (p *Parser) parseType(term sexp.SExp) (sc.Type, bool, *SyntaxError) {
	symbol := term.AsSymbol()
	if symbol == nil {
		return nil, false, p.translator.SyntaxError(term, "malformed type")
	}
	// Access string of symbol
	parts := strings.Split(symbol.Value, "@")
	if len(parts) > 2 || (len(parts) == 2 && parts[1] != "prove") {
		return nil, false, p.translator.SyntaxError(term, "malformed type")
	}
	// Determine whether type should be proven or not.
	proven := len(parts) == 2
	// See what we've got.
	switch parts[0] {
	case ":binary":
		return sc.NewUintType(1), proven, nil
	case ":byte":
		return sc.NewUintType(8), proven, nil
	default:
		// Handle generic types like i16, i128, etc.
		str := parts[0]
		if strings.HasPrefix(str, ":i") {
			n, err := strconv.Atoi(str[2:])
			if err != nil {
				return nil, false, p.translator.SyntaxError(symbol, err.Error())
			}
			// Done
			return sc.NewUintType(uint(n)), proven, nil
		}
	}
	// Error
	return nil, false, p.translator.SyntaxError(symbol, "unknown type")
}

func beginParserRule(_ string, args []Expr) (Expr, error) {
	return &List{args}, nil
}

func constantParserRule(symbol string) (Expr, bool, error) {
	if symbol[0] >= '0' && symbol[0] < '9' {
		var num fr.Element
		// Attempt to parse
		_, err := num.SetString(symbol)
		// Check for errors
		if err != nil {
			return nil, true, err
		}
		// Done
		return &Constant{Val: num}, true, nil
	}
	// Not applicable
	return nil, false, nil
}

func varAccessParserRule(col string) (Expr, bool, error) {
	// Sanity check what we have
	if !unicode.IsLetter(rune(col[0])) {
		return nil, false, nil
	}
	// Handle qualified accesses (where permitted)
	// Attempt to split column name into module / column pair.
	split := strings.Split(col, ".")
	if len(split) == 2 {
		return &VariableAccess{&split[0], split[1], 0, nil}, true, nil
	} else if len(split) > 2 {
		return nil, true, errors.New("malformed column access")
	}
	// Done
	return &VariableAccess{nil, col, 0, nil}, true, nil
}

func addParserRule(_ string, args []Expr) (Expr, error) {
	return &Add{args}, nil
}

func subParserRule(_ string, args []Expr) (Expr, error) {
	return &Sub{args}, nil
}

func mulParserRule(_ string, args []Expr) (Expr, error) {
	return &Mul{args}, nil
}

func ifParserRule(_ string, args []Expr) (Expr, error) {
	if len(args) == 2 {
		return &IfZero{args[0], args[1], nil}, nil
	} else if len(args) == 3 {
		return &IfZero{args[0], args[1], args[2]}, nil
	}

	return nil, errors.New("incorrect number of arguments")
}

func shiftParserRule(col string, amt string) (Expr, error) {
	n, err := strconv.Atoi(amt)

	if err != nil {
		return nil, err
	}
	// Sanity check what we have
	if !unicode.IsLetter(rune(col[0])) {
		return nil, nil
	}
	// Handle qualified accesses (where appropriate)
	split := strings.Split(col, ".")
	if len(split) == 2 {
		return &VariableAccess{&split[0], split[1], n, nil}, nil
	} else if len(split) > 2 {
		return nil, errors.New("malformed column access")
	}
	// Done
	return &VariableAccess{nil, col, n, nil}, nil
}

func powParserRule(_ string, args []Expr) (Expr, error) {
	var k big.Int

	if len(args) != 2 {
		return nil, errors.New("incorrect number of arguments")
	}

	c, ok := args[1].(*Constant)
	if !ok {
		return nil, errors.New("expected constant power")
	} else if !c.Val.IsUint64() {
		return nil, errors.New("constant power too large")
	}
	// Convert power to uint64
	c.Val.BigInt(&k)
	// Done
	return &Exp{Arg: args[0], Pow: k.Uint64()}, nil
}

func normParserRule(_ string, args []Expr) (Expr, error) {
	if len(args) != 1 {
		return nil, errors.New("incorrect number of arguments")
	}

	return &Normalise{Arg: args[0]}, nil
}
