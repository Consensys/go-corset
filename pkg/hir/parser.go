package hir

import (
	"errors"
	"fmt"
	"math/big"
	"strconv"
	"strings"
	"unicode"

	"github.com/consensys/gnark-crypto/ecc/bls12-377/fr"
	sc "github.com/consensys/go-corset/pkg/schema"
	"github.com/consensys/go-corset/pkg/schema/assignment"
	"github.com/consensys/go-corset/pkg/sexp"
	"github.com/consensys/go-corset/pkg/trace"
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
	module trace.Context
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
	p.AddRecursiveRule("^", powParserRule)
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
		} else if e.MatchSymbols(1, "defcolumns") {
			return p.parseColumnDeclarations(e)
		} else if e.Len() == 4 && e.MatchSymbols(2, "defconstraint") {
			return p.parseConstraintDeclaration(e.Elements)
		} else if e.Len() == 3 && e.MatchSymbols(2, "assert") {
			return p.parseAssertionDeclaration(e.Elements)
		} else if e.Len() == 3 && e.MatchSymbols(1, "defpermutation") {
			return p.parsePermutationDeclaration(e)
		} else if e.Len() == 4 && e.MatchSymbols(1, "deflookup") {
			return p.parseLookupDeclaration(e)
		} else if e.Len() == 3 && e.MatchSymbols(1, "definterleaved") {
			return p.parseInterleavingDeclaration(e)
		}
	}
	// Error
	return p.translator.SyntaxError(s, "unexpected or malformed declaration")
}

// Parse a column declaration
func (p *hirParser) parseModuleDeclaration(l *sexp.List) error {
	// Sanity check declaration
	if len(l.Elements) > 2 {
		return p.translator.SyntaxError(l, "malformed module declaration")
	}
	// Extract column name
	moduleName := l.Elements[1].AsSymbol().Value
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
func (p *hirParser) parseColumnDeclarations(l *sexp.List) error {
	// Sanity check declaration
	if len(l.Elements) == 1 {
		return p.translator.SyntaxError(l, "malformed column declaration")
	}
	// Process column declarations one by one.
	for i := 1; i < len(l.Elements); i++ {
		// Extract column name
		if err := p.parseColumnDeclaration(l.Elements[i]); err != nil {
			return err
		}
	}

	return nil
}

func (p *hirParser) parseColumnDeclaration(e sexp.SExp) error {
	var columnName string
	// Default to field type
	var columnType sc.Type = &sc.FieldType{}
	// Check whether extended declaration or not.
	if l := e.AsList(); l != nil {
		// Check at least the name provided.
		if len(l.Elements) == 0 {
			return p.translator.SyntaxError(l, "empty column declaration")
		}
		// Column name is always first
		columnName = l.Elements[0].String(false)
		//	Parse type (if applicable)
		if len(l.Elements) == 2 {
			var err error
			if columnType, err = p.parseType(l.Elements[1]); err != nil {
				return err
			}
		} else if len(l.Elements) > 2 {
			// For now.
			return p.translator.SyntaxError(l, "unknown column declaration attributes")
		}
	} else {
		columnName = e.String(false)
	}
	// Sanity check doesn't already exist
	if p.env.HasColumn(p.module, columnName) {
		return p.translator.SyntaxError(e, "duplicate column declaration")
	}
	// Register column
	cid := p.env.AddDataColumn(p.module, columnName, columnType)
	p.env.schema.AddTypeConstraint(cid, columnType)
	//
	return nil
}

// Parse a sorted permutation declaration
func (p *hirParser) parsePermutationDeclaration(l *sexp.List) error {
	// Target columns are (sorted) permutations of source columns.
	sexpTargets := l.Elements[1].AsList()
	// Source columns.
	sexpSources := l.Elements[2].AsList()
	// Sanity check
	if sexpTargets == nil {
		return p.translator.SyntaxError(l.Elements[1], "malformed target columns")
	} else if sexpSources == nil {
		return p.translator.SyntaxError(l.Elements[2], "malformed source columns")
	}
	// Convert into appropriate form.
	sources := make([]uint, sexpSources.Len())
	signs := make([]bool, sexpSources.Len())
	//
	if sexpTargets.Len() != sexpSources.Len() {
		return p.translator.SyntaxError(l, "sorted permutation requires matching number of source and target columns")
	}
	// initialise context
	ctx := trace.VoidContext()
	//
	for i := 0; i < sexpSources.Len(); i++ {
		sourceIndex, sourceSign, err := p.parsePermutationSource(sexpSources.Get(i))
		if err != nil {
			return err
		}
		// Check source context
		sourceCol := p.env.schema.Columns().Nth(sourceIndex)
		ctx = ctx.Join(sourceCol.Context())
		// Sanity check we have a sensible type here.
		if ctx.IsConflicted() {
			return p.translator.SyntaxError(sexpSources.Get(i), "conflicting evaluation context")
		} else if ctx.IsVoid() {
			return p.translator.SyntaxError(sexpSources.Get(i), "empty evaluation context")
		}
		// Copy over column name
		signs[i] = sourceSign
		sources[i] = sourceIndex
	}
	// Parse targets
	targets := make([]sc.Column, sexpTargets.Len())
	// Parse targets
	for i := 0; i < sexpTargets.Len(); i++ {
		targetName, err := p.parsePermutationTarget(sexpTargets.Get(i))
		//
		if err != nil {
			return err
		}
		// Lookup corresponding source
		source := p.env.schema.Columns().Nth(sources[i])
		// Done
		targets[i] = sc.NewColumn(ctx, targetName, source.Type())
	}
	//
	p.env.AddAssignment(assignment.NewSortedPermutation(ctx, targets, signs, sources))
	//
	return nil
}

func (p *hirParser) parsePermutationSource(source sexp.SExp) (uint, bool, error) {
	var (
		name string
		sign bool
		err  error
	)

	if source.AsList() != nil {
		l := source.AsList()
		// Check whether sort direction provided
		if l.Len() != 2 || l.Get(0).AsSymbol() == nil || l.Get(1).AsSymbol() == nil {
			return 0, false, p.translator.SyntaxError(source, "malformed column")
		}
		// Parser sorting direction
		if sign, err = p.parseSortDirection(l.Get(0).AsSymbol()); err != nil {
			return 0, false, err
		}
		// Extract column name
		name = l.Get(1).AsSymbol().Value
	} else {
		name = source.AsSymbol().Value
		sign = true // default
	}
	// Determine index for source column
	index, ok := p.env.LookupColumn(p.module, name)
	if !ok {
		// Column doesn't exist!
		return 0, false, p.translator.SyntaxError(source, "unknown column")
	}
	// Done
	return index, sign, nil
}

func (p *hirParser) parsePermutationTarget(target sexp.SExp) (string, error) {
	if target.AsSymbol() == nil {
		return "", p.translator.SyntaxError(target, "malformed target column")
	}
	//
	targetName := target.AsSymbol().Value
	// Sanity check that target column *doesn't* exist.
	if p.env.HasColumn(p.module, targetName) {
		// No, it doesn't.
		return "", p.translator.SyntaxError(target, "duplicate column")
	}
	// Done
	return targetName, nil
}

func (p *hirParser) parseSortDirection(l *sexp.Symbol) (bool, error) {
	switch l.Value {
	case "+", "↓":
		return true, nil
	case "-", "↑":
		return false, nil
	}
	// Unknown sort
	return false, p.translator.SyntaxError(l, "malformed sort direction")
}

// Parse a lookup declaration
func (p *hirParser) parseLookupDeclaration(l *sexp.List) error {
	handle := l.Elements[1].AsSymbol().Value
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
	sourceCtx := sc.JoinContexts(sources, p.env.schema)
	targetCtx := sc.JoinContexts(targets, p.env.schema)
	// Propagate errors
	if sourceCtx.IsConflicted() {
		return p.translator.SyntaxError(sexpSources, "conflicting evaluation context")
	} else if targetCtx.IsConflicted() {
		return p.translator.SyntaxError(sexpTargets, "conflicting evaluation context")
	} else if sourceCtx.IsVoid() {
		return p.translator.SyntaxError(sexpSources, "empty evaluation context")
	} else if targetCtx.IsVoid() {
		return p.translator.SyntaxError(sexpTargets, "empty evaluation context")
	}
	// Finally add constraint
	p.env.schema.AddLookupConstraint(handle, sourceCtx, targetCtx, sources, targets)
	// Done
	return nil
}

// Parse am interleaving declaration
func (p *hirParser) parseInterleavingDeclaration(l *sexp.List) error {
	// Target columns are (sorted) permutations of source columns.
	sexpTarget := l.Elements[1].AsSymbol()
	// Source columns.
	sexpSources := l.Elements[2].AsList()
	// Sanity checks.
	if sexpTarget == nil {
		return p.translator.SyntaxError(l, "column name expected")
	} else if sexpSources == nil {
		return p.translator.SyntaxError(l, "source column list expected")
	}
	// Construct and check source columns
	sources := make([]uint, sexpSources.Len())
	ctx := trace.VoidContext()

	for i := 0; i < sexpSources.Len(); i++ {
		ith := sexpSources.Get(i)
		col := ith.AsSymbol()
		// Sanity check a symbol was found
		if col == nil {
			return p.translator.SyntaxError(ith, "column name expected")
		}
		// Attempt to lookup the column
		cid, ok := p.env.LookupColumn(p.module, col.Value)
		// Check it exists
		if !ok {
			return p.translator.SyntaxError(ith, "unknown column")
		}
		// Check multiplier calculation
		sourceCol := p.env.schema.Columns().Nth(cid)
		ctx = ctx.Join(sourceCol.Context())
		// Sanity check we have a sensible context here.
		if ctx.IsConflicted() {
			return p.translator.SyntaxError(sexpSources.Get(i), "conflicting evaluation context")
		} else if ctx.IsVoid() {
			return p.translator.SyntaxError(sexpSources.Get(i), "empty evaluation context")
		}
		// Assign
		sources[i] = cid
	}
	// Add assignment
	p.env.AddAssignment(assignment.NewInterleaving(ctx, sexpTarget.Value, sources))
	// Done
	return nil
}

// Parse a property assertion
func (p *hirParser) parseAssertionDeclaration(elements []sexp.SExp) error {
	handle := elements[1].AsSymbol().Value
	// Property assertions do not have global scope, hence qualified column
	// accesses are not permitted.
	p.global = false
	// Translate
	expr, err := p.translator.Translate(elements[2])
	if err != nil {
		return err
	}
	// Determine evaluation context of assertion.
	ctx := expr.Context(p.env.schema)
	// Add assertion.
	p.env.schema.AddPropertyAssertion(handle, ctx, expr)

	return nil
}

// Parse a vanishing declaration
func (p *hirParser) parseConstraintDeclaration(elements []sexp.SExp) error {
	//
	handle := elements[1].AsSymbol().Value
	// Vanishing constraints do not have global scope, hence qualified column
	// accesses are not permitted.
	p.global = false
	attributes, err := p.parseConstraintAttributes(elements[2])
	// Check for error
	if err != nil {
		return err
	}
	// Translate expression
	expr, err := p.translator.Translate(elements[3])
	if err != nil {
		return err
	}
	// Determine evaluation context of expression.
	ctx := expr.Context(p.env.schema)
	// Sanity check we have a sensible context here.
	if ctx.IsConflicted() {
		return p.translator.SyntaxError(elements[3], "conflicting evaluation context")
	} else if ctx.IsVoid() {
		return p.translator.SyntaxError(elements[3], "empty evaluation context")
	}

	p.env.schema.AddVanishingConstraint(handle, ctx, attributes, expr)

	return nil
}

func (p *hirParser) parseConstraintAttributes(attributes sexp.SExp) (domain *int, err error) {
	var res *int = nil
	// Check attribute list is a list
	if attributes.AsList() == nil {
		return nil, p.translator.SyntaxError(attributes, "expected attribute list")
	}
	// Deconstruct as list
	attrs := attributes.AsList()
	// Process each attribute in turn
	for i := 0; i < attrs.Len(); i++ {
		ith := attrs.Get(i)
		// Check start of attribute
		if ith.AsSymbol() == nil {
			return nil, p.translator.SyntaxError(ith, "malformed attribute")
		}
		// Check what we've got
		switch ith.AsSymbol().Value {
		case ":domain":
			i++
			if res, err = p.parseDomainAttribute(attrs.Get(i)); err != nil {
				return nil, err
			}
		default:
			return nil, p.translator.SyntaxError(ith, "unknown attribute")
		}
	}
	// Done
	return res, nil
}

func (p *hirParser) parseDomainAttribute(attribute sexp.SExp) (domain *int, err error) {
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

func (p *hirParser) parseType(term sexp.SExp) (sc.Type, error) {
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

func columnAccessParserRule(parser *hirParser) func(col string) (Expr, bool, error) {
	// Returns a closure over the parser.
	return func(col string) (Expr, bool, error) {
		var ok bool
		// Sanity check what we have
		if !unicode.IsLetter(rune(col[0])) {
			return nil, false, nil
		}
		// Handle qualified accesses (where permitted)
		context := parser.module
		colname := col
		// Attempt to split column name into module / column pair.
		split := strings.Split(col, ".")
		if parser.global && len(split) == 2 {
			// Lookup module
			if context, ok = parser.env.LookupModule(split[0]); !ok {
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
		cid, ok = parser.env.LookupColumn(context, colname)
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

func powParserRule(args []Expr) (Expr, error) {
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

func normParserRule(args []Expr) (Expr, error) {
	if len(args) != 1 {
		return nil, errors.New("incorrect number of arguments")
	}

	return &Normalise{Arg: args[0]}, nil
}
