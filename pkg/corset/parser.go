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
	p.AddRecursiveListRule("+", addParserRule)
	p.AddRecursiveListRule("-", subParserRule)
	p.AddRecursiveListRule("*", mulParserRule)
	p.AddRecursiveListRule("~", normParserRule)
	p.AddRecursiveListRule("^", powParserRule)
	p.AddRecursiveListRule("begin", beginParserRule)
	p.AddRecursiveListRule("debug", debugParserRule)
	p.AddListRule("for", forParserRule(parser))
	p.AddListRule("reduce", reduceParserRule(parser))
	p.AddRecursiveListRule("if", ifParserRule)
	p.AddRecursiveListRule("shift", shiftParserRule)
	p.AddDefaultListRule(invokeParserRule(parser))
	p.AddDefaultRecursiveArrayRule(arrayAccessParserRule)
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
			return decls, terms[i:], errors
		} else if decl, errs := p.parseDeclaration(module, e); len(errs) > 0 {
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
		decl, errors = p.parseDefInRange(s.Elements)
	} else if s.Len() == 3 && s.MatchSymbols(1, "definterleaved") {
		decl, errors = p.parseDefInterleaved(module, s.Elements)
	} else if s.Len() == 4 && s.MatchSymbols(1, "deflookup") {
		decl, errors = p.parseDefLookup(s.Elements)
	} else if s.Len() == 3 && s.MatchSymbols(2, "defpermutation") {
		decl, errors = p.parseDefPermutation(module, s.Elements)
	} else if s.Len() == 4 && s.MatchSymbols(2, "defperspective") {
		decl, errors = p.parseDefPerspective(module, s.Elements)
	} else if s.Len() == 3 && s.MatchSymbols(2, "defproperty") {
		decl, errors = p.parseDefProperty(s.Elements)
	} else {
		errors = p.translator.SyntaxErrors(s, "malformed declaration")
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
		} else if !isEitherOrIdentifier(elements[i], functions) {
			// Symbol expected!
			errors = append(errors, *p.translator.SyntaxError(elements[i], "invalid alias name"))
		} else if !isEitherOrIdentifier(elements[i+1], functions) {
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
		binding := NewInputColumnBinding(module, "", false, 1, NewFieldType())
		decl, err := p.parseColumnDeclaration(l.Elements[i], binding)
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

func (p *Parser) parseColumnDeclaration(e sexp.SExp, binding *ColumnBinding) (*DefColumn, *SyntaxError) {
	var (
		name  string
		error *SyntaxError
	)
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
		if binding.dataType, binding.mustProve, error = p.parseColumnDeclarationAttributes(l.Elements[1:]); error != nil {
			return nil, error
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

func (p *Parser) parseColumnDeclarationAttributes(attrs []sexp.SExp) (Type, bool, *SyntaxError) {
	var (
		dataType  Type = NewFieldType()
		mustProve bool = false
		array_min uint
		array_max uint
		err       *SyntaxError
	)

	for i := 0; i < len(attrs); i++ {
		ith := attrs[i]
		symbol := ith.AsSymbol()
		// Sanity check
		if symbol == nil {
			return nil, false, p.translator.SyntaxError(ith, "unknown column attribute")
		}
		//
		switch symbol.Value {
		case ":display":
			// skip these for now, as they are only relevant to the inspector.
			if i+1 == len(attrs) {
				return nil, false, p.translator.SyntaxError(ith, "incomplete display definition")
			} else if attrs[i+1].AsSymbol() == nil {
				return nil, false, p.translator.SyntaxError(ith, "malformed display definition")
			}
			// Check what display attribute we have
			switch attrs[i+1].AsSymbol().String(false) {
			case ":dec", ":hex", ":bytes", ":opcode":
				// all good
				i = i + 1
			default:
				// not good
				return nil, false, p.translator.SyntaxError(ith, "unknown display definition")
			}
		case ":array":
			if array_min, array_max, err = p.parseArrayDimension(attrs[i+1]); err != nil {
				return nil, false, err
			}
			// skip dimension
			i++
		default:
			if dataType, mustProve, err = p.parseType(ith); err != nil {
				return nil, false, err
			}
		}
	}
	// Done
	if array_max != 0 {
		return NewArrayType(dataType, array_min, array_max), mustProve, nil
	}
	//
	return dataType, mustProve, nil
}

func (p *Parser) parseArrayDimension(s sexp.SExp) (uint, uint, *SyntaxError) {
	dim := s.AsArray()
	//
	if dim == nil || dim.Get(0).AsSymbol() == nil || dim.Len() != 1 {
		return 0, 0, p.translator.SyntaxError(s, "invalid array dimension")
	} else {
		// Check for interval dimensions
		split := strings.Split(dim.Get(0).AsSymbol().Value, ":")
		//
		if len(split) == 0 || len(split) > 2 {
			return 0, 0, p.translator.SyntaxError(s, "invalid array dimension")
		} else if m, ok_m := strconv.Atoi(split[0]); len(split) == 1 && m >= 0 && ok_m == nil {
			return uint(1), uint(m), nil
		} else if ok_m != nil || m < 0 {
			//unlikely scenarios
		} else if n, ok_n := strconv.Atoi(split[1]); len(split) == 2 && n >= 0 && ok_n == nil {
			return uint(m), uint(n), nil
		}
	}
	//
	return 0, 0, p.translator.SyntaxError(s, "invalid array dimension")
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
	expr, errors := p.translator.Translate(value)
	// Check for errors
	if len(errors) != 0 {
		return nil, errors
	}
	// Looks good
	def := &DefConstUnit{name, NewConstantBinding(expr)}
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
		return nil, p.translator.SyntaxErrors(elements[1], "invalid constraint handle")
	}
	// Vanishing constraints do not have global scope, hence qualified column
	// accesses are not permitted.
	domain, guard, perspective, errs := p.parseConstraintAttributes(elements[2])
	errors = append(errors, errs...)
	// Translate expression
	expr, errs := p.translator.Translate(elements[3])
	errors = append(errors, errs...)
	// Error Check
	if len(errors) > 0 {
		return nil, errors
	}
	// Done
	return &DefConstraint{elements[1].AsSymbol().Value, domain, guard, perspective, expr, false}, nil
}

// Parse a interleaved declaration
func (p *Parser) parseDefInterleaved(module string, elements []sexp.SExp) (Declaration, []SyntaxError) {
	var (
		errors  []SyntaxError
		sources []Symbol
	)
	// Check target column
	if !isIdentifier(elements[1]) {
		errors = append(errors, *p.translator.SyntaxError(elements[1], "malformed target column"))
	}
	// Check source columns
	if elements[2].AsList() == nil {
		errors = append(errors, *p.translator.SyntaxError(elements[2], "malformed source columns"))
	} else {
		// Extract target and source columns
		sexpSources := elements[2].AsList()
		sources = make([]Symbol, sexpSources.Len())
		//
		for i := 0; i != sexpSources.Len(); i++ {
			ith := sexpSources.Get(i)
			if !isIdentifier(ith) {
				errors = append(errors, *p.translator.SyntaxError(ith, "malformed source column"))
			} else {
				// Extract column name
				sources[i] = NewColumnName(ith.AsSymbol().Value)
				p.mapSourceNode(ith, sources[i])
			}
		}
	}
	// Error Check
	if len(errors) != 0 {
		return nil, errors
	}
	//
	binding := NewComputedColumnBinding(module)
	target := &DefColumn{elements[1].AsSymbol().Value, *binding}
	// Updating mapping for target definition
	p.mapSourceNode(elements[1], target)
	// Done
	return &DefInterleaved{target, sources}, nil
}

// Parse a lookup declaration
func (p *Parser) parseDefLookup(elements []sexp.SExp) (Declaration, []SyntaxError) {
	var (
		errors  []SyntaxError
		sources []Expr
		targets []Expr
	)
	// Extract items
	handle := elements[1]
	sexpTargets := elements[2].AsList()
	sexpSources := elements[3].AsList()
	// Check Handle
	if !isIdentifier(handle) {
		errors = append(errors, *p.translator.SyntaxError(elements[1], "malformed handle"))
	}
	// Check target expressions
	if sexpTargets == nil {
		errors = append(errors, *p.translator.SyntaxError(elements[2], "malformed target columns"))
	}
	// Check source Expressions
	if sexpSources == nil {
		errors = append(errors, *p.translator.SyntaxError(elements[3], "malformed source columns"))
	}
	// Sanity check number of columns matches
	if sexpTargets != nil && sexpSources != nil {
		if sexpTargets.Len() != sexpSources.Len() {
			errors = append(errors, *p.translator.SyntaxError(elements[3], "incorrect number of columns"))
		} else {
			sources = make([]Expr, sexpSources.Len())
			targets = make([]Expr, sexpTargets.Len())
			// Translate source & target expressions
			for i := 0; i < sexpTargets.Len(); i++ {
				var errs []SyntaxError
				// Translate source expressions
				sources[i], errs = p.translator.Translate(sexpSources.Get(i))
				errors = append(errors, errs...)
				// Translate target expressions
				targets[i], errs = p.translator.Translate(sexpTargets.Get(i))
				errors = append(errors, errs...)
			}
		}
	}
	// Error check
	if len(errors) != 0 {
		return nil, errors
	}
	// Done
	return &DefLookup{handle.AsSymbol().Value, sources, targets, false}, nil
}

// Parse a permutation declaration
func (p *Parser) parseDefPermutation(module string, elements []sexp.SExp) (Declaration, []SyntaxError) {
	var (
		errors  []SyntaxError
		sources []Symbol
		signs   []bool
		targets []*DefColumn
	)
	//
	sexpTargets := elements[1].AsList()
	sexpSources := elements[2].AsList()
	// Check target columns
	if sexpTargets == nil {
		errors = append(errors, *p.translator.SyntaxError(elements[1], "malformed target columns"))
	}
	// Check source columns
	if sexpSources == nil {
		errors = append(errors, *p.translator.SyntaxError(elements[2], "malformed source columns"))
	}
	// Sanity check relative lengths
	if sexpTargets.Len() < sexpSources.Len() {
		errors = append(errors, *p.translator.SyntaxError(elements[1], "too few target columns"))
	}
	// Sanity check relative lengths
	if sexpTargets.Len() > sexpSources.Len() {
		errors = append(errors, *p.translator.SyntaxError(elements[1], "too many target columns"))
	}
	//
	if sexpTargets != nil && sexpSources != nil {
		targets = make([]*DefColumn, sexpTargets.Len())
		sources = make([]Symbol, sexpSources.Len())
		signs = make([]bool, sexpSources.Len())
		//
		for i := 0; i < min(len(sources), len(targets)); i++ {
			var err *SyntaxError
			// Parse target column
			binding := NewComputedColumnBinding(module)
			if targets[i], err = p.parseColumnDeclaration(sexpTargets.Get(i), binding); err != nil {
				errors = append(errors, *err)
			}
			// Parse source column
			if sources[i], signs[i], err = p.parsePermutedColumnDeclaration(i == 0, sexpSources.Get(i)); err != nil {
				errors = append(errors, *err)
			}
		}
	}
	// Error Check
	if len(errors) != 0 {
		return nil, errors
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

// Parse a perspective declaration
func (p *Parser) parseDefPerspective(module string, elements []sexp.SExp) (Declaration, []SyntaxError) {
	var (
		errors           []SyntaxError
		sexp_columns     *sexp.List = elements[3].AsList()
		columns          []*DefColumn
		perspective      *PerspectiveName
		perspective_name string
	)
	// Check for columns
	if sexp_columns == nil {
		errors = append(errors, *p.translator.SyntaxError(elements[3], "expected column declarations"))
	}
	// Translate selector
	selector, errs := p.translator.Translate(elements[2])
	errors = append(errors, errs...)
	// Parse perspective selector
	binding := NewPerspectiveBinding(selector)
	// Parse perspective name
	if perspective, errs = parseSymbolName(p, elements[1], false, binding); len(errs) == 0 {
		perspective_name = perspective.name
	} else {
		errors = append(errors, errs...)
	}
	// Process column declarations one by one.
	if sexp_columns != nil {
		columns = make([]*DefColumn, sexp_columns.Len())

		for i := 0; i < len(sexp_columns.Elements); i++ {
			binding := NewInputColumnBinding(module, perspective_name, false, 1, NewFieldType())
			decl, err := p.parseColumnDeclaration(sexp_columns.Elements[i], binding)
			// Extract column name
			if err != nil {
				errors = append(errors, *err)
			}
			// Assign the declaration
			columns[i] = decl
		}
	}
	// Error check
	if len(errors) != 0 {
		return nil, errors
	}
	//
	return &DefPerspective{perspective, selector, columns}, nil
}

// Parse a property assertion
func (p *Parser) parseDefProperty(elements []sexp.SExp) (Declaration, []SyntaxError) {
	var errors []SyntaxError
	// Initial sanity checks
	if !isIdentifier(elements[1]) {
		errors = p.translator.SyntaxErrors(elements[1], "expected constraint handle")
	}
	//
	handle := elements[1].AsSymbol()
	// Translate expression
	expr, errs := p.translator.Translate(elements[2])
	errors = append(errors, errs...)
	// Error Check
	if len(errors) != 0 {
		return nil, errors
	}
	// Done
	return &DefProperty{handle.Value, expr, false}, nil
}

// Parse a permutation declaration
func (p *Parser) parseDefFun(pure bool, elements []sexp.SExp) (Declaration, []SyntaxError) {
	var (
		name      *sexp.Symbol
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
	body, errs := p.translator.Translate(elements[2])
	errors = append(errors, errs...)
	// Check for errors
	if len(errors) > 0 {
		return nil, errors
	}
	// Extract parameter types
	paramTypes := make([]Type, len(params))
	for i, p := range params {
		paramTypes[i] = p.Binding.datatype
	}
	// Construct binding
	binding := NewDefunBinding(pure, paramTypes, ret, body)
	fn_name := NewFunctionName(name.Value, &binding)
	// Update source mapping
	p.mapSourceNode(name, fn_name)
	//
	return &DefFun{fn_name, params}, nil
}

func (p *Parser) parseFunctionSignature(elements []sexp.SExp) (*sexp.Symbol, Type, []*DefParameter, []SyntaxError) {
	var (
		params []*DefParameter = make([]*DefParameter, len(elements)-1)
	)
	// Parse name and (optional) return type
	name, ret, _, errors := p.parseFunctionNameReturn(elements[0])
	// Parse parameters
	for i := 0; i < len(params); i = i + 1 {
		var errs []SyntaxError

		if params[i], errs = p.parseFunctionParameter(elements[i+1]); len(errs) > 0 {
			errors = append(errors, errs...)
		}
	}
	// Check for any errors arising
	if len(errors) > 0 {
		return nil, nil, nil, errors
	}
	//
	return name, ret, params, nil
}

func (p *Parser) parseFunctionNameReturn(element sexp.SExp) (*sexp.Symbol, Type, bool, []SyntaxError) {
	var (
		err    *SyntaxError
		name   sexp.SExp
		ret    Type = nil
		forced bool
		symbol *sexp.Symbol = element.AsSymbol()
		list   *sexp.List   = element.AsList()
	)
	//
	if symbol != nil {
		name = symbol
	} else {
		// Check all modifiers
		for i, element := range list.Elements {
			symbol := element.AsSymbol()
			// Check what we have
			if symbol == nil {
				err := p.translator.SyntaxError(element, "modifier expected")
				return nil, nil, false, []SyntaxError{*err}
			} else if i == 0 {
				name = symbol
			} else {
				switch symbol.Value {
				case ":force":
					forced = true
				default:
					if ret, _, err = p.parseType(element); err != nil {
						return nil, nil, false, []SyntaxError{*err}
					}
				}
			}
		}
	}
	//
	if isFunIdentifier(name) {
		return name.AsSymbol(), ret, forced, nil
	} else {
		// Must be non-identifier symbol
		err = p.translator.SyntaxError(element, "invalid function name")
		return nil, nil, false, []SyntaxError{*err}
	}
}

func (p *Parser) parseFunctionParameter(element sexp.SExp) (*DefParameter, []SyntaxError) {
	list := element.AsList()
	//
	if isIdentifier(element) {
		binding := NewLocalVariableBinding(element.AsSymbol().Value, NewFieldType())
		return &DefParameter{binding}, nil
	} else if list == nil || list.Len() != 2 || !isIdentifier(list.Get(0)) {
		// Construct error message (for now)
		err := p.translator.SyntaxError(element, "malformed parameter declaration")
		//
		return nil, []SyntaxError{*err}
	}
	// Parse the type
	datatype, prove, err := p.parseType(list.Get(1))
	//
	if err != nil {
		return nil, []SyntaxError{*err}
	} else if prove {
		// Parameters cannot be marked @prove
		err := p.translator.SyntaxError(element, "malformed parameter declaration")
		//
		return nil, []SyntaxError{*err}
	}
	// Done
	binding := NewLocalVariableBinding(list.Get(0).AsSymbol().Value, datatype)
	//
	return &DefParameter{binding}, nil
}

// Parse a range declaration
func (p *Parser) parseDefInRange(elements []sexp.SExp) (Declaration, []SyntaxError) {
	var bound fr.Element
	// Translate expression
	expr, errors := p.translator.Translate(elements[1])
	// Check & parse bound
	if elements[2].AsSymbol() == nil {
		errors = append(errors, *p.translator.SyntaxError(elements[2], "malformed bound"))
	} else if _, err := bound.SetString(elements[2].AsSymbol().Value); err != nil {
		errors = append(errors, *p.translator.SyntaxError(elements[2], "malformed bound"))
	}
	// Error check
	if len(errors) != 0 {
		return nil, errors
	}
	// Done
	return &DefInRange{Expr: expr, Bound: bound}, nil
}

func (p *Parser) parseConstraintAttributes(attributes sexp.SExp) (domain *int, guard Expr,
	perspective *PerspectiveName, err []SyntaxError) {
	//
	var errors []SyntaxError
	// Check attribute list is a list
	if attributes.AsList() == nil {
		return nil, nil, nil, p.translator.SyntaxErrors(attributes, "expected attribute list")
	}
	// Deconstruct as list
	attrs := attributes.AsList()
	// Process each attribute in turn
	for i := 0; i < attrs.Len(); i++ {
		ith := attrs.Get(i)
		// Check start of attribute
		if ith.AsSymbol() == nil {
			errors = append(errors, *p.translator.SyntaxError(ith, "malformed attribute"))
		} else {
			var errs []SyntaxError
			// Check what we've got
			switch ith.AsSymbol().Value {
			case ":domain":
				i++
				domain, errs = p.parseDomainAttribute(attrs.Get(i))
			case ":guard":
				i++
				guard, errs = p.translator.Translate(attrs.Get(i))
			case ":perspective":
				i++
				//binding := NewPerspectiveBinding()
				perspective, errs = parseSymbolName[*PerspectiveBinding](p, attrs.Get(i), false, nil)
			default:
				errs = p.translator.SyntaxErrors(ith, "unknown attribute")
			}
			//
			if len(errs) != 0 {
				errors = append(errors, errs...)
			}
		}
	}
	// Error Check
	if len(errors) != 0 {
		return nil, nil, nil, errors
	}
	// Done
	return domain, guard, perspective, nil
}

// Parse a symbol name, which will include a binding.
func parseSymbolName[T Binding](p *Parser, symbol sexp.SExp, function bool, binding T) (*Name[T], []SyntaxError) {
	if !isEitherOrIdentifier(symbol, function) {
		return nil, p.translator.SyntaxErrors(symbol, "expected identifier")
	}
	// Extract
	name := &Name[T]{symbol.AsSymbol().Value, function, binding, false}
	// Update source mapping
	p.mapSourceNode(symbol, name)
	// Construct
	return name, nil
}

func (p *Parser) parseDomainAttribute(attribute sexp.SExp) (domain *int, err []SyntaxError) {
	if attribute.AsSet() == nil {
		return nil, p.translator.SyntaxErrors(attribute, "malformed domain set")
	}
	// Sanity check
	set := attribute.AsSet()
	// Check all domain elements well-formed.
	for i := 0; i < set.Len(); i++ {
		ith := set.Get(i)
		if ith.AsSymbol() == nil {
			return nil, p.translator.SyntaxErrors(ith, "malformed domain")
		}
	}
	// Currently, only support domains of size 1.
	if set.Len() == 1 {
		first, err := strconv.Atoi(set.Get(0).AsSymbol().Value)
		// Check for parse error
		if err != nil {
			return nil, p.translator.SyntaxErrors(set.Get(0), "malformed domain element")
		}
		// Done
		return &first, nil
	}
	// Fail
	return nil, p.translator.SyntaxErrors(attribute, "multiple values not supported")
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
	case ":":
		if len(parts) == 1 {
			return nil, false, p.translator.SyntaxError(symbol, "unknown type")
		}
		//
		datatype = NewFieldType()
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

func debugParserRule(_ string, args []Expr) (Expr, error) {
	if len(args) == 1 {
		return &Debug{args[0]}, nil
	}
	//
	return nil, errors.New("incorrect number of arguments")
}

func forParserRule(p *Parser) sexp.ListRule[Expr] {
	return func(list *sexp.List) (Expr, []SyntaxError) {
		var (
			errors   []SyntaxError
			indexVar *sexp.Symbol
		)
		// Check we've got the expected number
		if list.Len() != 4 {
			msg := fmt.Sprintf("expected 3 arguments, found %d", list.Len()-1)
			return nil, p.translator.SyntaxErrors(list, msg)
		}
		// Extract index variable
		if indexVar = list.Get(1).AsSymbol(); indexVar == nil {
			err := p.translator.SyntaxError(list.Get(1), "invalid index variable")
			errors = append(errors, *err)
		}
		// Parse range
		start, end, errs := parseForRange(p, list.Get(2))
		// Error Check
		errors = append(errors, errs...)
		// Parse body
		body, errs := p.translator.Translate(list.Get(3))
		errors = append(errors, errs...)
		// Error check
		if len(errors) > 0 {
			return nil, errors
		}
		// Construct binding.  At this stage, its unclear what the best type to
		// use for the index variable is here.  Potentially, it could be refined
		// based on the range of actual values, etc.
		binding := NewLocalVariableBinding(indexVar.Value, NewFieldType())
		// Done
		return &For{binding, start, end, body}, nil
	}
}

// Parse a range which, represented as a string is "[s:e]".
func parseForRange(p *Parser, interval sexp.SExp) (uint, uint, []SyntaxError) {
	var (
		start int
		end   int
		err1  error
		err2  error
	)
	// This is a bit dirty.  Essentially, we turn the sexp.Array back into a
	// string and then parse it from there.
	str := interval.String(false)
	// Strip out any whitespace (which is permitted)
	str = strings.ReplaceAll(str, " ", "")
	// Check has form "[...]"
	if !strings.HasPrefix(str, "[") || !strings.HasSuffix(str, "]") {
		// error
		return 0, 0, p.translator.SyntaxErrors(interval, "invalid interval")
	}
	// Split out components
	splits := strings.Split(str[1:len(str)-1], ":")
	// Error check
	if len(splits) == 0 || len(splits) > 2 {
		// error
		return 0, 0, p.translator.SyntaxErrors(interval, "invalid interval")
	} else if len(splits) == 1 {
		end, err1 = strconv.Atoi(splits[0])
		start = 1
	} else if len(splits) == 2 {
		start, err1 = strconv.Atoi(splits[0])
		end, err2 = strconv.Atoi(splits[1])
	}
	//
	if err1 != nil || err2 != nil {
		return 0, 0, p.translator.SyntaxErrors(interval, "invalid interval")
	}
	// Success
	return uint(start), uint(end), nil
}

func reduceParserRule(p *Parser) sexp.ListRule[Expr] {
	return func(list *sexp.List) (Expr, []SyntaxError) {
		var errors []SyntaxError
		// Check we've got the expected number
		if list.Len() != 3 {
			msg := fmt.Sprintf("expected 2 arguments, found %d", list.Len()-1)
			return nil, p.translator.SyntaxErrors(list, msg)
		}
		// function name
		name := list.Get(1).AsSymbol()
		//
		if name == nil {
			errors = append(errors, *p.translator.SyntaxError(list.Get(1), "invalid function"))
		}
		// Parse body
		body, errs := p.translator.Translate(list.Get(2))
		errors = append(errors, errs...)
		// Error check
		if len(errors) > 0 {
			return nil, errors
		}
		//
		varaccess := &VariableAccess{nil, name.Value, true, nil}
		p.mapSourceNode(name, varaccess)
		// Done
		return &Reduce{varaccess, nil, body}, nil
	}
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
	} else if (symbol[0] >= '0' && symbol[0] <= '9') || symbol[0] == '-' {
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
		return nil, false, errors.New("malformed column access")
	}
	// Handle qualified accesses (where permitted)
	// Attempt to split column name into module / column pair.
	split := strings.Split(col, ".")
	if len(split) == 2 {
		return &VariableAccess{&split[0], split[1], false, nil}, true, nil
	} else if len(split) > 2 {
		return nil, true, errors.New("malformed column access")
	} else {
		return &VariableAccess{nil, col, false, nil}, true, nil
	}
}

func arrayAccessParserRule(name string, args []Expr) (Expr, error) {
	if len(args) != 1 {
		return nil, errors.New("malformed array access")
	}
	//
	return &ArrayAccess{name, args[0], nil}, nil
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

func invokeParserRule(p *Parser) sexp.ListRule[Expr] {
	return func(list *sexp.List) (Expr, []SyntaxError) {
		var (
			varaccess *VariableAccess
			errors    []SyntaxError
		)
		//
		if list.Len() == 0 || list.Get(0).AsSymbol() == nil {
			return nil, p.translator.SyntaxErrors(list, "invalid invocation")
		}
		// Extract function name
		name := list.Get(0).AsSymbol()
		// Sanity check what we have
		if !isFunIdentifier(name) {
			errors = append(errors, *p.translator.SyntaxError(list.Get(0), "invalid function name"))
		}
		// Handle qualified accesses (where permitted)
		// Attempt to split column name into module / column pair.
		split := strings.Split(name.Value, ".")
		if len(split) == 2 {
			//
			varaccess = &VariableAccess{&split[0], split[1], true, nil}
		} else if len(split) > 2 {
			return nil, p.translator.SyntaxErrors(list.Get(0), "invalid function name")
		} else {
			varaccess = &VariableAccess{nil, split[0], true, nil}
		}
		// Parse arguments
		args := make([]Expr, list.Len()-1)
		for i := 0; i < len(args); i++ {
			var errs []SyntaxError
			args[i], errs = p.translator.Translate(list.Get(i + 1))
			errors = append(errors, errs...)
		}
		// Error check
		if len(errors) > 0 {
			return nil, errors
		}
		//
		p.mapSourceNode(list.Get(0), varaccess)
		// Done
		return &Invoke{varaccess, nil, args}, nil
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
// The function flag switches this to identifiers suitable for functions and
// invocations.
func isEitherOrIdentifier(sexp sexp.SExp, function bool) bool {
	if function {
		return isFunIdentifier(sexp)
	}
	//
	return isIdentifier(sexp)
}

// Attempt to parse an S-Expression as an identifier suitable for something
// which is not a function (e.g. column, constant, etc).
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

// Attempt to parse an S-Expression as an identifier suitable for something
// which is not a function (e.g. column, constant, etc).
func isFunIdentifier(sexp sexp.SExp) bool {
	if symbol := sexp.AsSymbol(); symbol != nil && len(symbol.Value) > 0 {
		runes := []rune(symbol.Value)
		if isFunctionSymbol(runes) {
			return true
		} else if isFunctionIdentifierStart(runes[0]) {
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
	return unicode.IsDigit(c) || isIdentifierStart(c) || c == '-' || c == '!' || c == '@'
}

func isFunctionIdentifierStart(c rune) bool {
	return isIdentifierStart(c) || c == '~'
}

func isFunctionSymbol(runes []rune) bool {
	return len(runes) == 1 && (runes[0] == '+' || runes[0] == '*' || runes[0] == '-' || runes[0] == '=')
}
