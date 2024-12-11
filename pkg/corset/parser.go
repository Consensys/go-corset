package corset

import (
	"errors"
	"fmt"
	"math/big"
	"sort"
	"strconv"
	"strings"
	"unicode"

	"github.com/consensys/gnark-crypto/ecc/bls12-377/fr"
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
	var errors []SyntaxError
	// Construct an initially empty source map
	srcmaps := sexp.NewSourceMaps[Node]()
	// num_errs counts the number of errors reported
	var num_errs uint
	// Contents map holds the combined fragments of each module.
	contents := make(map[string]Module, 0)
	// Names identifies the names of each unique module.
	names := make([]string, 0)
	//
	for _, file := range files {
		c, srcmap, errs := ParseSourceFile(file)
		// Handle errors
		if len(errs) > 0 {
			num_errs += uint(len(errs))
			// Report any errors encountered
			errors = append(errors, errs...)
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
func ParseSourceFile(srcfile *sexp.SourceFile) (Circuit, *sexp.SourceMap[Node], []SyntaxError) {
	var (
		circuit Circuit
		errors  []SyntaxError
	)
	// Parse bytes into an S-Expression
	terms, srcmap, err := srcfile.ParseAll()
	// Check test file parsed ok
	if err != nil {
		return circuit, nil, []SyntaxError{*err}
	}
	// Construct parser for corset syntax
	p := NewParser(srcfile, srcmap)
	// Parse whatever is declared at the beginning of the file before the first
	// module declaration.  These declarations form part of the "prelude".
	if circuit.Declarations, terms, errors = p.parseModuleContents("", terms); len(errors) > 0 {
		return circuit, nil, errors
	}
	// Continue parsing string until nothing remains.
	for len(terms) != 0 {
		var (
			name  string
			decls []Declaration
		)
		// Extract module name
		if name, errors = p.parseModuleStart(terms[0]); len(errors) > 0 {
			return circuit, nil, errors
		}
		// Parse module contents
		if decls, terms, errors = p.parseModuleContents(name, terms[1:]); len(errors) > 0 {
			return circuit, nil, errors
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
	p.AddRecursiveRule("+", addParserRule)
	p.AddRecursiveRule("-", subParserRule)
	p.AddRecursiveRule("*", mulParserRule)
	p.AddRecursiveRule("~", normParserRule)
	p.AddRecursiveRule("^", powParserRule)
	p.AddRecursiveRule("begin", beginParserRule)
	p.AddRecursiveRule("if", ifParserRule)
	p.AddRecursiveRule("shift", shiftParserRule)
	p.AddDefaultRecursiveRule(invokeParserRule)
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
func (p *Parser) parseModuleContents(module string, terms []sexp.SExp) ([]Declaration, []sexp.SExp, []SyntaxError) {
	var errors []SyntaxError
	//
	decls := make([]Declaration, 0)
	//
	for i, s := range terms {
		e, ok := s.(*sexp.List)
		// Check for error
		if !ok {
			err := p.translator.SyntaxError(s, "unexpected or malformed declaration")
			errors = append(errors, *err)
		} else if e.MatchSymbols(2, "module") {
			return decls, terms[i:], nil
		} else if decl, errs := p.parseDeclaration(module, e); errs != nil {
			errors = append(errors, errs...)
		} else {
			// Continue accumulating declarations for this module.
			decls = append(decls, decl)
		}
	}
	// Sanity check errors
	if len(errors) > 0 {
		return nil, nil, errors
	}
	// End-of-file signals end-of-module.
	return decls, make([]sexp.SExp, 0), nil
}

// Parse a module declaration of the form "(module m1)" which indicates the
// start of module m1.
func (p *Parser) parseModuleStart(s sexp.SExp) (string, []SyntaxError) {
	l, ok := s.(*sexp.List)
	// Check for error
	if !ok {
		err := p.translator.SyntaxError(s, "unexpected or malformed declaration")
		return "", []SyntaxError{*err}
	}
	// Sanity check declaration
	if len(l.Elements) > 2 {
		err := p.translator.SyntaxError(l, "malformed module declaration")
		return "", []SyntaxError{*err}
	}
	// Extract column name
	name := l.Elements[1].AsSymbol().Value
	//
	return name, nil
}

func (p *Parser) parseDeclaration(module string, s *sexp.List) (Declaration, []SyntaxError) {
	var (
		decl   Declaration
		errors []SyntaxError
		err    *SyntaxError
	)
	//
	if s.MatchSymbols(1, "defalias") {
		decl, errors = p.parseDefAlias(false, s.Elements)
	} else if s.MatchSymbols(1, "defcolumns") {
		decl, errors = p.parseDefColumns(module, s)
	} else if s.Len() > 1 && s.MatchSymbols(1, "defconst") {
		decl, errors = p.parseDefConst(s.Elements)
	} else if s.Len() == 4 && s.MatchSymbols(2, "defconstraint") {
		decl, errors = p.parseDefConstraint(s.Elements)
	} else if s.MatchSymbols(1, "defunalias") {
		decl, errors = p.parseDefAlias(true, s.Elements)
	} else if s.Len() == 3 && s.MatchSymbols(1, "defpurefun") {
		decl, errors = p.parseDefFun(true, s.Elements)
	} else if s.Len() == 3 && s.MatchSymbols(1, "defun") {
		decl, errors = p.parseDefFun(false, s.Elements)
	} else if s.Len() == 3 && s.MatchSymbols(1, "definrange") {
		decl, err = p.parseDefInRange(s.Elements)
	} else if s.Len() == 3 && s.MatchSymbols(1, "definterleaved") {
		decl, err = p.parseDefInterleaved(module, s.Elements)
	} else if s.Len() == 4 && s.MatchSymbols(1, "deflookup") {
		decl, err = p.parseDefLookup(s.Elements)
	} else if s.Len() == 3 && s.MatchSymbols(2, "defpermutation") {
		decl, err = p.parseDefPermutation(module, s.Elements)
	} else if s.Len() == 3 && s.MatchSymbols(2, "defproperty") {
		decl, err = p.parseDefProperty(s.Elements)
	} else {
		err = p.translator.SyntaxError(s, "malformed declaration")
	}
	// Handle unit error case
	if err != nil {
		errors = append(errors, *err)
	}
	// Register node if appropriate
	if decl != nil {
		p.mapSourceNode(s, decl)
	}
	// done
	return decl, errors
}

// Parse an alias declaration
func (p *Parser) parseDefAlias(functions bool, elements []sexp.SExp) (Declaration, []SyntaxError) {
	var (
		errors  []SyntaxError
		aliases []*DefAlias
		names   []Symbol
	)

	for i := 1; i < len(elements); i += 2 {
		// Sanity check first
		if i+1 == len(elements) {
			// Uneven number of constant declarations!
			errors = append(errors, *p.translator.SyntaxError(elements[i], "missing alias definition"))
		} else if !isIdentifier(elements[i]) {
			// Symbol expected!
			errors = append(errors, *p.translator.SyntaxError(elements[i], "invalid alias name"))
		} else if !isIdentifier(elements[i+1]) {
			// Symbol expected!
			errors = append(errors, *p.translator.SyntaxError(elements[i+1], "invalid alias definition"))
		} else {
			alias := &DefAlias{elements[i].AsSymbol().Value}
			name := NewName[Binding](elements[i+1].AsSymbol().Value, functions)
			p.mapSourceNode(elements[i], alias)
			p.mapSourceNode(elements[i+1], name)
			//
			aliases = append(aliases, alias)
			names = append(names, name)
		}
	}
	// Done
	return &DefAliases{functions, aliases, names}, errors
}

// Parse a column declaration
func (p *Parser) parseDefColumns(module string, l *sexp.List) (Declaration, []SyntaxError) {
	columns := make([]*DefColumn, l.Len()-1)
	// Sanity check declaration
	if len(l.Elements) == 1 {
		err := p.translator.SyntaxError(l, "malformed column declaration")
		return nil, []SyntaxError{*err}
	}
	//
	var errors []SyntaxError
	// Process column declarations one by one.
	for i := 1; i < len(l.Elements); i++ {
		decl, err := p.parseColumnDeclaration(module, l.Elements[i])
		// Extract column name
		if err != nil {
			errors = append(errors, *err)
		}
		// Assign the declaration
		columns[i-1] = decl
	}
	// Sanity check errors
	if len(errors) > 0 {
		return nil, errors
	}
	// Done
	return &DefColumns{columns}, nil
}

func (p *Parser) parseColumnDeclaration(module string, e sexp.SExp) (*DefColumn, *SyntaxError) {
	var name string
	//
	binding := NewColumnBinding(module, false, false, 1, NewFieldType())
	// Check whether extended declaration or not.
	if l := e.AsList(); l != nil {
		// Check at least the name provided.
		if len(l.Elements) == 0 {
			return nil, p.translator.SyntaxError(l, "empty column declaration")
		} else if !isIdentifier(l.Elements[0]) {
			return nil, p.translator.SyntaxError(l.Elements[0], "invalid column name")
		}
		// Column name is always first
		name = l.Elements[0].String(false)
		//	Parse type (if applicable)
		if len(l.Elements) == 2 {
			var err *SyntaxError
			if binding.dataType, binding.mustProve, err = p.parseType(l.Elements[1]); err != nil {
				return nil, err
			}
		} else if len(l.Elements) > 2 {
			// For now.
			return nil, p.translator.SyntaxError(l, "unknown column declaration attributes")
		}
	} else {
		name = e.String(false)
	}
	//
	def := &DefColumn{name, *binding}
	// Update source mapping
	p.mapSourceNode(e, def)
	//
	return def, nil
}

// Parse a constant declaration
func (p *Parser) parseDefConst(elements []sexp.SExp) (Declaration, []SyntaxError) {
	var (
		errors    []SyntaxError
		constants []*DefConstUnit
	)

	for i := 1; i < len(elements); i += 2 {
		// Sanity check first
		if i+1 == len(elements) {
			// Uneven number of constant declarations!
			errors = append(errors, *p.translator.SyntaxError(elements[i], "missing constant definition"))
		} else if !isIdentifier(elements[i]) {
			// Symbol expected!
			errors = append(errors, *p.translator.SyntaxError(elements[i], "invalid constant name"))
		} else {
			// Attempt to parse definition
			constant, errs := p.parseDefConstUnit(elements[i].AsSymbol().Value, elements[i+1])
			errors = append(errors, errs...)
			constants = append(constants, constant)
		}
	}
	// Done
	return &DefConst{constants}, errors
}

func (p *Parser) parseDefConstUnit(name string, value sexp.SExp) (*DefConstUnit, []SyntaxError) {
	expr, err := p.translator.Translate(value)
	// Check for errors
	if err != nil {
		return nil, []SyntaxError{*err}
	}
	// Looks good
	def := &DefConstUnit{name, ConstantBinding{expr}}
	// Map to source node
	p.mapSourceNode(value, def)
	// Done
	return def, nil
}

// Parse a vanishing declaration
func (p *Parser) parseDefConstraint(elements []sexp.SExp) (Declaration, []SyntaxError) {
	var errors []SyntaxError
	// Initial sanity checks
	if !isIdentifier(elements[1]) {
		err := p.translator.SyntaxError(elements[1], "invalid constraint handle")
		return nil, []SyntaxError{*err}
	}
	// Vanishing constraints do not have global scope, hence qualified column
	// accesses are not permitted.
	domain, guard, err := p.parseConstraintAttributes(elements[2])
	// Check for error
	if err != nil {
		errors = append(errors, *err)
	}
	// Translate expression
	expr, err := p.translator.Translate(elements[3])
	if err != nil {
		errors = append(errors, *err)
	}
	//
	if len(errors) > 0 {
		return nil, errors
	}
	// Done
	return &DefConstraint{elements[1].AsSymbol().Value, domain, guard, expr}, nil
}

// Parse a interleaved declaration
func (p *Parser) parseDefInterleaved(module string, elements []sexp.SExp) (Declaration, *SyntaxError) {
	// Initial sanity checks
	if !isIdentifier(elements[1]) {
		return nil, p.translator.SyntaxError(elements[1], "malformed target column")
	} else if elements[2].AsList() == nil {
		return nil, p.translator.SyntaxError(elements[2], "malformed source columns")
	}
	// Extract target and source columns
	sexpSources := elements[2].AsList()
	sources := make([]Symbol, sexpSources.Len())
	//
	for i := 0; i != sexpSources.Len(); i++ {
		ith := sexpSources.Get(i)
		if !isIdentifier(ith) {
			return nil, p.translator.SyntaxError(ith, "malformed source column")
		}
		// Extract column name
		sources[i] = NewColumnName(ith.AsSymbol().Value)
		p.mapSourceNode(ith, sources[i])
	}
	//
	binding := NewColumnBinding(module, false, false, 1, NewFieldType())
	target := &DefColumn{elements[1].AsSymbol().Value, *binding}
	// Updating mapping for target definition
	p.mapSourceNode(elements[1], target)
	// Done
	return &DefInterleaved{target, sources}, nil
}

// Parse a lookup declaration
func (p *Parser) parseDefLookup(elements []sexp.SExp) (Declaration, *SyntaxError) {
	// Initial sanity checks
	if !isIdentifier(elements[1]) {
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
func (p *Parser) parseDefPermutation(module string, elements []sexp.SExp) (Declaration, *SyntaxError) {
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
	sources := make([]Symbol, sexpSources.Len())
	signs := make([]bool, sexpSources.Len())
	//
	for i := 0; i < len(targets); i++ {
		// Parse target column
		if targets[i], err = p.parseColumnDeclaration(module, sexpTargets.Get(i)); err != nil {
			return nil, err
		}
		// Parse source column
		if sources[i], signs[i], err = p.parsePermutedColumnDeclaration(i == 0, sexpSources.Get(i)); err != nil {
			return nil, err
		}
	}
	//
	return &DefPermutation{targets, sources, signs}, nil
}

func (p *Parser) parsePermutedColumnDeclaration(signRequired bool, e sexp.SExp) (*ColumnName, bool, *SyntaxError) {
	var (
		err  *SyntaxError
		name *ColumnName
		sign bool
	)
	// Check whether extended declaration or not.
	if l := e.AsList(); l != nil {
		// Check at least the name provided.
		if len(l.Elements) == 0 {
			return nil, false, p.translator.SyntaxError(l, "empty permutation column")
		} else if len(l.Elements) != 2 {
			return nil, false, p.translator.SyntaxError(l, "malformed permutation column")
		} else if l.Get(0).AsSymbol() == nil || l.Get(1).AsSymbol() == nil {
			return nil, false, p.translator.SyntaxError(l, "empty permutation column")
		}
		// Parse sign
		if sign, err = p.parsePermutedColumnSign(l.Get(0).AsSymbol()); err != nil {
			return nil, false, err
		}
		// Parse column name
		name = NewColumnName(l.Get(1).AsSymbol().Value)
	} else if signRequired {
		return nil, false, p.translator.SyntaxError(e, "missing sort direction")
	} else {
		name = NewColumnName(e.String(false))
	}
	// Update source mapping
	p.mapSourceNode(e, name)
	//
	return name, sign, nil
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
func (p *Parser) parseDefProperty(elements []sexp.SExp) (Declaration, *SyntaxError) {
	// Initial sanity checks
	if !isIdentifier(elements[1]) {
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

// Parse a permutation declaration
func (p *Parser) parseDefFun(pure bool, elements []sexp.SExp) (Declaration, []SyntaxError) {
	var (
		name      string
		ret       Type
		params    []*DefParameter
		errors    []SyntaxError
		signature *sexp.List = elements[1].AsList()
	)
	// Parse signature
	if signature == nil || signature.Len() == 0 {
		err := p.translator.SyntaxError(elements[1], "malformed function signature")
		errors = append(errors, *err)
	} else {
		name, ret, params, errors = p.parseFunctionSignature(signature.Elements)
	}
	// Translate expression
	body, err := p.translator.Translate(elements[2])
	if err != nil {
		errors = append(errors, *err)
	}
	// Check for errors
	if len(errors) > 0 {
		return nil, errors
	}
	// Extract parameter types
	paramTypes := make([]Type, len(params))
	for i, p := range params {
		paramTypes[i] = p.DataType
	}
	// Construct binding
	binding := NewFunctionBinding(pure, paramTypes, ret, body)
	//
	return &DefFun{name, params, binding}, nil
}

func (p *Parser) parseFunctionSignature(elements []sexp.SExp) (string, Type, []*DefParameter, []SyntaxError) {
	var (
		params []*DefParameter = make([]*DefParameter, len(elements)-1)
		ret    Type            = NewFieldType()
		errors []SyntaxError
	)
	// Parse name
	if !isIdentifier(elements[0]) {
		err := p.translator.SyntaxError(elements[1], "expected function name")
		errors = append(errors, *err)
	}
	// Parse parameters
	for i := 0; i < len(params); i = i + 1 {
		var errs []SyntaxError

		if params[i], errs = p.parseFunctionParameter(elements[i+1]); len(errs) > 0 {
			errors = append(errors, errs...)
		}
	}
	// Check for any errors arising
	if len(errors) > 0 {
		return "", nil, nil, errors
	}
	//
	return elements[0].AsSymbol().Value, ret, params, nil
}

func (p *Parser) parseFunctionParameter(element sexp.SExp) (*DefParameter, []SyntaxError) {
	if isIdentifier(element) {
		return &DefParameter{element.AsSymbol().Value, NewFieldType()}, nil
	}
	// Construct error message (for now)
	err := p.translator.SyntaxError(element, "malformed parameter declaration")
	//
	return nil, []SyntaxError{*err}
}

// Parse a range declaration
func (p *Parser) parseDefInRange(elements []sexp.SExp) (Declaration, *SyntaxError) {
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

func (p *Parser) parseType(term sexp.SExp) (Type, bool, *SyntaxError) {
	symbol := term.AsSymbol()
	if symbol == nil {
		return nil, false, p.translator.SyntaxError(term, "malformed type")
	}
	// Access string of symbol
	parts := strings.Split(symbol.Value, "@")
	// Determine whether type should be proven or not.
	var datatype Type
	// See what we've got.
	switch parts[0] {
	case ":binary":
		datatype = NewUintType(1)
	case ":byte":
		datatype = NewUintType(8)
	default:
		// Handle generic types like i16, i128, etc.
		str := parts[0]
		if !strings.HasPrefix(str, ":i") {
			return nil, false, p.translator.SyntaxError(symbol, "unknown type")
		}
		// Parse bitwidth
		n, err := strconv.Atoi(str[2:])
		if err != nil {
			return nil, false, p.translator.SyntaxError(symbol, err.Error())
		}
		// Done
		datatype = NewUintType(uint(n))
	}
	// Types not proven unless explicitly requested
	var proven bool = false
	// Process type modifiers
	for i := 1; i < len(parts); i++ {
		switch parts[i] {
		case "prove":
			proven = true
		case "loob":
			datatype = datatype.WithLoobeanSemantics()
		case "bool":
			datatype = datatype.WithBooleanSemantics()
		default:
			msg := fmt.Sprintf("unknown modifier \"%s\"", parts[i])
			return nil, false, p.translator.SyntaxError(symbol, msg)
		}
	}
	// Done
	return datatype, proven, nil
}

func beginParserRule(_ string, args []Expr) (Expr, error) {
	return &List{args}, nil
}

func constantParserRule(symbol string) (Expr, bool, error) {
	var (
		base int
		name string
		num  big.Int
	)
	//
	if strings.HasPrefix(symbol, "0x") {
		symbol = symbol[2:]
		base = 16
		name = "hexadecimal"
	} else if (symbol[0] >= '0' && symbol[0] < '9') || symbol[0] == '-' {
		base = 10
		name = "integer"
	} else {
		// Not applicable
		return nil, false, nil
	}
	// Attempt to parse
	if _, ok := num.SetString(symbol, base); !ok {
		err := fmt.Sprintf("invalid %s constant", name)
		return nil, true, errors.New(err)
	}
	// Done
	return &Constant{Val: num}, true, nil
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
		return &VariableAccess{&split[0], split[1], nil}, true, nil
	} else if len(split) > 2 {
		return nil, true, errors.New("malformed column access")
	} else {
		return &VariableAccess{nil, col, nil}, true, nil
	}
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
		return &If{0, args[0], args[1], nil}, nil
	} else if len(args) == 3 {
		return &If{0, args[0], args[1], args[2]}, nil
	}

	return nil, errors.New("incorrect number of arguments")
}

func invokeParserRule(name string, args []Expr) (Expr, error) {
	// Sanity check what we have
	if !unicode.IsLetter(rune(name[0])) {
		return nil, nil
	}
	// Handle qualified accesses (where permitted)
	// Attempt to split column name into module / column pair.
	split := strings.Split(name, ".")
	if len(split) == 2 {
		return &Invoke{&split[0], split[1], args, nil}, nil
	} else if len(split) > 2 {
		return nil, errors.New("malformed function invocation")
	} else {
		return &Invoke{nil, name, args, nil}, nil
	}
}

func shiftParserRule(_ string, args []Expr) (Expr, error) {
	if len(args) != 2 {
		return nil, errors.New("incorrect number of arguments")
	}
	// Done
	return &Shift{Arg: args[0], Shift: args[1]}, nil
}

func powParserRule(_ string, args []Expr) (Expr, error) {
	if len(args) != 2 {
		return nil, errors.New("incorrect number of arguments")
	}
	// Done
	return &Exp{Arg: args[0], Pow: args[1]}, nil
}

func normParserRule(_ string, args []Expr) (Expr, error) {
	if len(args) != 1 {
		return nil, errors.New("incorrect number of arguments")
	}

	return &Normalise{Arg: args[0]}, nil
}

// Attempt to parse an S-Expression as an identifier, return nil if this fails.
func isIdentifier(sexp sexp.SExp) bool {
	if symbol := sexp.AsSymbol(); symbol != nil && len(symbol.Value) > 0 {
		runes := []rune(symbol.Value)
		if isIdentifierStart(runes[0]) {
			for i := 1; i < len(runes); i++ {
				if !isIdentifierMiddle(runes[i]) {
					return false
				}
			}
			// Success
			return true
		}
	}
	// Fail
	return false
}

func isIdentifierStart(c rune) bool {
	return unicode.IsLetter(c) || c == '_' || c == '\''
}

func isIdentifierMiddle(c rune) bool {
	return unicode.IsDigit(c) || isIdentifierStart(c) || c == '-' || c == '!'
}
