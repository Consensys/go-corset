package corset

import (
	"github.com/consensys/gnark-crypto/ecc/bls12-377/fr"
	sc "github.com/consensys/go-corset/pkg/schema"
	"github.com/consensys/go-corset/pkg/sexp"
	tr "github.com/consensys/go-corset/pkg/trace"
	"github.com/consensys/go-corset/pkg/util"
)

// Circuit represents the root of the Abstract Syntax Tree.  This is also
// referred to as the "prelude".  All modules are contained within the root, and
// declarations can also be declared here as well.
type Circuit struct {
	Modules      []Module
	Declarations []Declaration
}

// Module represents a top-level module declaration.  This corresponds to a
// table in the final constraint set.
type Module struct {
	Name         string
	Declarations []Declaration
}

// Node provides common functionality across all elements of the Abstract Syntax
// Tree.  For example, it ensures every element can converted back into Lisp
// form for debugging.  Furthermore, it provides a reference point for
// constructing a suitable source map for reporting syntax errors.
type Node interface {
	// Convert this node into its lisp representation.  This is primarily used
	// for debugging purposes.
	Lisp() sexp.SExp
}

// ColumnAssignment provides a schematic for describing a column arising from an
// assignment.
type ColumnAssignment struct {
	// Name of defined column
	Name string
	// Length multiplier for defined column
	LengthMultiplier uint
	// Type of defined column
	Type sc.Type
}

// Symbol represents a variable or function access within a declaration.
// Initially, such the proper interpretation of such accesses is unclear and it
// is only later when we can distinguish them (e.g. whether its a column access,
// a constant access, etc).
type Symbol interface {
	Node
	// Determines whether this symbol is qualfied or not (i.e. has an explicitly
	// module specifier).
	IsQualified() bool
	// Indicates whether or not this is a function.
	IsFunction() bool
	// Checks whether this symbol has been resolved already, or not.
	IsResolved() bool
	// Optional module qualification
	Module() string
	// Name of the symbol
	Name() string
	// Get binding associated with this interface.  This will panic if this
	// symbol is not yet resolved.
	Binding() Binding
	// Resolve this symbol by associating it with the binding associated with
	// the definition of the symbol to which this refers.
	Resolve(Binding)
}

// SymbolDefinition represents a declaration (or part thereof) which defines a
// particular symbol.  For example, "defcolumns" will define one or more symbols
// representing columns, etc.
type SymbolDefinition interface {
	Node
	// Name of symbol being defined
	Name() string
	// Indicates whether or not this is a function definition.
	IsFunction() bool
	// Allocated binding for the symbol which may or may not be finalised.
	Binding() Binding
}

// Declaration represents a top-level declaration in a Corset source file (e.g.
// defconstraint, defcolumns, etc).
type Declaration interface {
	Node
	// Returns the set of symbols being defined this declaration.  Observe that
	// these may not yet have been finalised.
	Definitions() util.Iterator[SymbolDefinition]
	// Return set of columns on which this declaration depends.
	Dependencies() util.Iterator[Symbol]
}

// Assignment is a declaration which introduces one (or more) computed columns.
type Assignment interface {
	Declaration

	// Return the set of columns which are declared by this assignment.
	Targets() []string

	// Return the set of column assignments, or nil if the assignments cannot yet
	// be determined (i.e. because the environment doesn't have complete
	// information for one or more dependent columns).  This can also fail for
	// other reasons, such as when two columns in an interleaving have different
	// length multipliers.
	Resolve(*Environment) ([]ColumnAssignment, []SyntaxError)
}

// ColumnName represents a name within some syntactic item.  Essentially this wraps a
// string and provides a mechanism for it to be associated with source line
// information.
type ColumnName struct {
	name    string
	binding Binding
}

// IsQualified determines whether this symbol is qualfied or not (i.e. has an
// explicit module specifier).  Column names are never qualified.
func (e *ColumnName) IsQualified() bool {
	return false
}

// IsFunction indicates whether or not this symbol refers to a function (which
// of course it never does).
func (e *ColumnName) IsFunction() bool {
	return false
}

// IsResolved checks whether this symbol has been resolved already, or not.
func (e *ColumnName) IsResolved() bool {
	return e.binding != nil
}

// Module returns the optional module qualification.  This always panics because
// column name's are never qualified.
func (e *ColumnName) Module() string {
	panic("undefined")
}

// Name returns the (unqualified) name of the column to which this symbol
// refers.
func (e *ColumnName) Name() string {
	return e.name
}

// Binding gets binding associated with this interface.  This will panic if this
// symbol is not yet resolved.
func (e *ColumnName) Binding() Binding {
	if e.binding == nil {
		panic("name not yet resolved")
	}
	//
	return e.binding
}

// Resolve this symbol by associating it with the binding associated with
// the definition of the symbol to which this refers.
func (e *ColumnName) Resolve(binding Binding) {
	if e.binding != nil {
		panic("name already resolved")
	}
	//
	e.binding = binding
}

// Lisp converts this node into its lisp representation.  This is primarily used
// for debugging purposes.
func (e *ColumnName) Lisp() sexp.SExp {
	return sexp.NewSymbol(e.name)
}

// DefColumns captures a set of one or more columns being declared.
type DefColumns struct {
	Columns []*DefColumn
}

// Dependencies needed to signal declaration.
func (p *DefColumns) Dependencies() util.Iterator[Symbol] {
	return util.NewArrayIterator[Symbol](nil)
}

// Definitions returns the set of symbols defined by this declaration.  Observe
// that these may not yet have been finalised.
func (p *DefColumns) Definitions() util.Iterator[SymbolDefinition] {
	iter := util.NewArrayIterator(p.Columns)
	return util.NewCastIterator[*DefColumn, SymbolDefinition](iter)
}

// Lisp converts this node into its lisp representation.  This is primarily used
// for debugging purposes.
func (p *DefColumns) Lisp() sexp.SExp {
	panic("got here")
}

// DefColumn packages together those piece relevant to declaring an individual
// column, such its name and type.
type DefColumn struct {
	// Column name
	name string
	// Binding of this column (which may or may not be finalised).
	binding ColumnBinding
}

// IsFunction is never true for a column definition.
func (e *DefColumn) IsFunction() bool {
	return false
}

// Binding returns the allocated binding for this symbol (which may or may not
// be finalised).
func (e *DefColumn) Binding() Binding {
	return &e.binding
}

// Name of symbol being defined
func (e *DefColumn) Name() string {
	return e.name
}

// DataType returns the type of this column.  If this column have not yet been
// finalised, then this will panic.
func (e *DefColumn) DataType() sc.Type {
	if !e.binding.IsFinalised() {
		panic("unfinalised column")
	}
	//
	return e.binding.dataType
}

// LengthMultiplier returns the length multiplier of this column (where the
// height of this column is determined as the product of the enclosing module's
// height and this length multiplier).  If this column have not yet been
// finalised, then this will panic.
func (e *DefColumn) LengthMultiplier() uint {
	if !e.binding.IsFinalised() {
		panic("unfinalised column")
	}
	//
	return e.binding.multiplier
}

// MustProve determines whether or not the type of this column must be
// established by the prover (e.g. a range constraint or similar).
func (e *DefColumn) MustProve() bool {
	if !e.binding.IsFinalised() {
		panic("unfinalised column")
	}
	//
	return e.binding.mustProve
}

// Lisp converts this node into its lisp representation.  This is primarily used
// for debugging purposes.
func (e *DefColumn) Lisp() sexp.SExp {
	panic("got here")
}

// DefConstraint represents a vanishing constraint, which is either "local" or
// "global".  A local constraint applies either to the first or last rows,
// whilst a global constraint applies to all rows.  For a constraint to hold,
// its expression must evaluate to zero for the rows on which it is active.  A
// constraint may also have a "guard" which is an expression that must evaluate
// to a non-zero value for the constraint to be considered active.  The
// expression for a constraint must have a single context.  That is, it can only
// be applied to columns within the same module (i.e. to ensure they have the
// same height).  Furthermore, within a given module, we require that all
// columns accessed by the constraint have the same length multiplier.
type DefConstraint struct {
	// Unique handle given to this constraint.  This is primarily useful for
	// debugging (i.e. so we know which constaint failed, etc).
	Handle string
	// Domain of this constraint, where nil indicates a global constraint.
	// Otherwise, a given value indicates a single row on which this constraint
	// should apply (where negative values are taken from the end, meaning that
	// -1 represents the last row of a given module).
	Domain *int
	// A selector which determines for which rows this constraint is active.
	// Specifically, when the expression evaluates to a non-zero value then the
	// constraint is active; otherwiser, its inactive. Nil is permitted to
	// indicate no guard is present.
	Guard Expr
	// The constraint itself which (when active) should evaluate to zero for the
	// relevant set of rows.
	Constraint Expr
}

// Definitions returns the set of symbols defined by this declaration.  Observe
// that these may not yet have been finalised.
func (p *DefConstraint) Definitions() util.Iterator[SymbolDefinition] {
	return util.NewArrayIterator[SymbolDefinition](nil)
}

// Dependencies needed to signal declaration.
func (p *DefConstraint) Dependencies() util.Iterator[Symbol] {
	var guard_deps []Symbol
	// Extract guard's dependencies (if applicable)
	if p.Guard != nil {
		guard_deps = p.Guard.Dependencies()
	}
	// Extract bodies dependencies
	body_deps := p.Constraint.Dependencies()
	// Done
	return util.NewArrayIterator[Symbol](append(guard_deps, body_deps...))
}

// Lisp converts this node into its lisp representation.  This is primarily used
// for debugging purposes.
func (p *DefConstraint) Lisp() sexp.SExp {
	panic("got here")
}

// DefInRange restricts all values for a given expression to be within a range
// [0..n) for some bound n.  Any bound is supported, and the system will choose
// the best underlying implementation as needed.
type DefInRange struct {
	// The expression whose values are being constrained to within the given
	// bound.
	Expr Expr
	// The upper bound for this constraint.  Specifically, every evaluation of
	// the expression should produce a value strictly below this bound.  NOTE:
	// an fr.Element is used here to store the bound simply to make the
	// necessary comparison against table data more direct.
	Bound fr.Element
}

// Definitions returns the set of symbols defined by this declaration.  Observe
// that these may not yet have been finalised.
func (p *DefInRange) Definitions() util.Iterator[SymbolDefinition] {
	return util.NewArrayIterator[SymbolDefinition](nil)
}

// Dependencies needed to signal declaration.
func (p *DefInRange) Dependencies() util.Iterator[Symbol] {
	return util.NewArrayIterator[Symbol](p.Expr.Dependencies())
}

// Lisp converts this node into its lisp representation.  This is primarily used
// for debugging purposes.
func (p *DefInRange) Lisp() sexp.SExp {
	panic("got here")
}

// DefInterleaved generates a new column by interleaving two or more existing
// colummns.  For example, say Z interleaves X and Y (in that order) and we have
// a trace X=[1,2], Y=[3,4].  Then, the interleaved column Z has the values
// Z=[1,3,2,4].  All columns must be defined within the same context.  Finally,
// the type of the interleaved column is the widest type of any source columns.
// For example, consider an interleaving of two columns X and Y with types i16
// and i8 respectively.  Then, the type of the resulting column is i16 (as this
// is required to hold an element from any source column).
type DefInterleaved struct {
	// The target column being defined
	Target *DefColumn
	// The source columns used to define the interleaved target column.
	Sources []Symbol
}

// Definitions returns the set of symbols defined by this declaration.  Observe
// that these may not yet have been finalised.
func (p *DefInterleaved) Definitions() util.Iterator[SymbolDefinition] {
	iter := util.NewUnitIterator(p.Target)
	return util.NewCastIterator[*DefColumn, SymbolDefinition](iter)
}

// Dependencies needed to signal declaration.
func (p *DefInterleaved) Dependencies() util.Iterator[Symbol] {
	return util.NewArrayIterator(p.Sources)
}

// Lisp converts this node into its lisp representation.  This is primarily used
// for debugging purposes.
func (p *DefInterleaved) Lisp() sexp.SExp {
	panic("got here")
}

// DefLookup represents a lookup constraint between a set N of source
// expressions and a set of N target expressions.  The source expressions must
// have a single context (i.e. all be in the same module) and likewise for the
// target expressions (though the source and target contexts can differ).  The
// constraint can be viewed as a "subset constraint".  Let the set of "source
// tuples" be those obtained by evaluating the source expressions over all rows
// in the source context, and likewise the "target tuples" those for the target
// expressions in the target context.  Then the lookup constraint holds if the
// set of source tuples is a subset of the target tuples.  This does not need to
// be a strict subset, so the two sets can be identical.  Furthermore, these are
// not treated as multi-sets, hence the number of occurrences of a given tuple
// is not relevant.
type DefLookup struct {
	// Unique handle given to this constraint.  This is primarily useful for
	// debugging (i.e. so we know which constaint failed, etc).
	Handle string
	// Source expressions for lookup (i.e. these values must all be contained
	// within the targets).
	Sources []Expr
	// Target expressions for lookup (i.e. these values must contain all of the
	// source values, but may contain more).
	Targets []Expr
}

// Definitions returns the set of symbols defined by this declaration.  Observe
// that these may not yet have been finalised.
func (p *DefLookup) Definitions() util.Iterator[SymbolDefinition] {
	return util.NewArrayIterator[SymbolDefinition](nil)
}

// Dependencies needed to signal declaration.
func (p *DefLookup) Dependencies() util.Iterator[Symbol] {
	sourceDeps := DependenciesOfExpressions(p.Sources)
	targetDeps := DependenciesOfExpressions(p.Targets)
	// Combine deps
	return util.NewArrayIterator(append(sourceDeps, targetDeps...))
}

// Lisp converts this node into its lisp representation.  This is primarily used
// for debugging purposes.
func (p *DefLookup) Lisp() sexp.SExp {
	panic("got here")
}

// DefPermutation represents a (lexicographically sorted) permutation of a set
// of source columns in a given source context, manifested as an assignment to a
// corresponding set of target columns.  The sort direction for each of the
// source columns can be specified as increasing or decreasing.
type DefPermutation struct {
	Targets []*DefColumn
	Sources []Symbol
	Signs   []bool
}

// Definitions returns the set of symbols defined by this declaration.  Observe
// that these may not yet have been finalised.
func (p *DefPermutation) Definitions() util.Iterator[SymbolDefinition] {
	iter := util.NewArrayIterator(p.Targets)
	return util.NewCastIterator[*DefColumn, SymbolDefinition](iter)
}

// Dependencies needed to signal declaration.
func (p *DefPermutation) Dependencies() util.Iterator[Symbol] {
	return util.NewArrayIterator(p.Sources)
}

// Lisp converts this node into its lisp representation.  This is primarily used
// for debugging purposes.
func (p *DefPermutation) Lisp() sexp.SExp {
	panic("got here")
}

// DefProperty represents an assertion to be used only for debugging / testing /
// verification.  Unlike vanishing constraints, property assertions do not
// represent something that the prover can enforce.  Rather, they represent
// properties which are expected to hold for every valid trace. That is, they
// should be implied by the actual constraints.  Thus, whilst the prover cannot
// enforce such properties, external tools (such as for formal verification) can
// attempt to ensure they do indeed always hold.
type DefProperty struct {
	// Unique handle given to this constraint.  This is primarily useful for
	// debugging (i.e. so we know which constaint failed, etc).
	Handle string
	// The assertion itself which (when active) should evaluate to zero for the
	// relevant set of rows.
	Assertion Expr
}

// Definitions returns the set of symbols defined by this declaration.  Observe that
// these may not yet have been finalised.
func (p *DefProperty) Definitions() util.Iterator[SymbolDefinition] {
	return util.NewArrayIterator[SymbolDefinition](nil)
}

// Dependencies needed to signal declaration.
func (p *DefProperty) Dependencies() util.Iterator[Symbol] {
	return util.NewArrayIterator(p.Assertion.Dependencies())
}

// Lisp converts this node into its lisp representation.  This is primarily used
// for debugging purposes.
func (p *DefProperty) Lisp() sexp.SExp {
	panic("got here")
}

// DefFun represents defines a (possibly pure) "function" (which, in actuality,
// is more like a macro).  Specifically, whenever an invocation of this function
// is encountered we can imagine that, in the final constraint set, the body of
// the function is inlined at the point of the call.  A pure function is not
// permitted to access any columns in scope (i.e. it can only manipulate its
// parameters).  In contrast, an impure function can access those columns
// defined within its enclosing context.
type DefFun struct {
	name string
	// Parameters
	parameters []*DefParameter
	//
	binding FunctionBinding
}

// IsFunction is always true for a function definition!
func (p *DefFun) IsFunction() bool {
	return true
}

// IsPure indicates whether or not this is a pure function.  That is, a function
// which is not permitted to access any columns from the enclosing environment
// (either directly itself, or indirectly via functions it calls).
func (p *DefFun) IsPure() bool {
	return p.binding.pure
}

// Parameters returns information about the parameters defined by this
// declaration.
func (p *DefFun) Parameters() []*DefParameter {
	return p.parameters
}

// Body Access information about the parameters defined by this declaration.
func (p *DefFun) Body() Expr {
	return p.binding.body
}

// Binding returns the allocated binding for this symbol (which may or may not
// be finalised).
func (p *DefFun) Binding() Binding {
	return &p.binding
}

// Name of symbol being defined
func (p *DefFun) Name() string {
	return p.name
}

// Definitions returns the set of symbols defined by this declaration.  Observe
// that these may not yet have been finalised.
func (p *DefFun) Definitions() util.Iterator[SymbolDefinition] {
	iter := util.NewUnitIterator(p)
	return util.NewCastIterator[*DefFun, SymbolDefinition](iter)
}

// Dependencies needed to signal declaration.
func (p *DefFun) Dependencies() util.Iterator[Symbol] {
	deps := p.binding.body.Dependencies()
	ndeps := make([]Symbol, 0)
	// Filter out all parameters declared in this function, since these are not
	// external dependencies.
	for _, d := range deps {
		if d.IsQualified() || d.IsFunction() || !p.hasParameter(d.Name()) {
			ndeps = append(ndeps, d)
		}
	}
	// Done
	return util.NewArrayIterator(ndeps)
}

// Lisp converts this node into its lisp representation.  This is primarily used
// for debugging purposes.
func (p *DefFun) Lisp() sexp.SExp {
	panic("got here")
}

// hasParameter checks whether this function has a parameter with the given
// name, or not.
func (p *DefFun) hasParameter(name string) bool {
	for _, v := range p.parameters {
		if v.Name == name {
			return true
		}
	}
	//
	return false
}

// DefParameter packages together those piece relevant to declaring an individual
// parameter, such its name and type.
type DefParameter struct {
	// Column name
	Name string
	// The datatype which all values in this parameter should inhabit.
	DataType sc.Type
}

// Lisp converts this node into its lisp representation.  This is primarily used
// for debugging purposes.
func (p *DefParameter) Lisp() sexp.SExp {
	panic("got here")
}

// Expr represents an arbitrary expression over the columns of a given context
// (or the parameters of an enclosing function).  Such expressions are pitched
// at a higher-level than those of the underlying constraint system.  For
// example, they can contain conditionals (i.e. if expressions) and
// normalisations, etc.  During the lowering process down to the underlying
// constraints level (AIR), such expressions are "compiled out" using various
// techniques (such as introducing computed columns where necessary).
type Expr interface {
	Node
	// Multiplicity defines the number of values which will be returned when
	// evaluating this expression.  Due to the nature of expressions in Corset,
	// they can (perhaps surprisingly) return multiple values.  For example,
	// lists return one value for each element in the list.  Note, every
	// expression must return at least one value.
	Multiplicity() uint

	// Context returns the context for this expression.  Observe that the
	// expression must have been resolved for this to be defined (i.e. it may
	// panic if it has not been resolved yet).
	Context() Context

	// Substitute all variables (such as for function parameters) arising in
	// this expression.
	Substitute(args []Expr) Expr

	// Return set of columns on which this declaration depends.
	Dependencies() []Symbol
}

// Context represents the evaluation context for a given expression.
type Context = tr.RawContext[string]

// ============================================================================
// Addition
// ============================================================================

// Add represents the sum over zero or more expressions.
type Add struct{ Args []Expr }

// Multiplicity determines the number of values that evaluating this expression
// can generate.
func (e *Add) Multiplicity() uint {
	return determineMultiplicity(e.Args)
}

// Context returns the context for this expression.  Observe that the
// expression must have been resolved for this to be defined (i.e. it may
// panic if it has not been resolved yet).
func (e *Add) Context() Context {
	return ContextOfExpressions(e.Args)
}

// Lisp converts this schema element into a simple S-Expression, for example
// so it can be printed.
func (e *Add) Lisp() sexp.SExp {
	panic("todo")
}

// Substitute all variables (such as for function parameters) arising in
// this expression.
func (e *Add) Substitute(args []Expr) Expr {
	return &Add{SubstituteExpressions(e.Args, args)}
}

// Dependencies needed to signal declaration.
func (e *Add) Dependencies() []Symbol {
	return DependenciesOfExpressions(e.Args)
}

// ============================================================================
// Constants
// ============================================================================

// Constant represents a constant value within an expression.
type Constant struct{ Val fr.Element }

// Multiplicity determines the number of values that evaluating this expression
// can generate.
func (e *Constant) Multiplicity() uint {
	return 1
}

// Context returns the context for this expression.  Observe that the
// expression must have been resolved for this to be defined (i.e. it may
// panic if it has not been resolved yet).
func (e *Constant) Context() Context {
	return tr.VoidContext[string]()
}

// Lisp converts this schema element into a simple S-Expression, for example
// so it can be printed.
func (e *Constant) Lisp() sexp.SExp {
	return sexp.NewSymbol(e.Val.String())
}

// Substitute all variables (such as for function parameters) arising in
// this expression.
func (e *Constant) Substitute(args []Expr) Expr {
	return e
}

// Dependencies needed to signal declaration.
func (e *Constant) Dependencies() []Symbol {
	return nil
}

// ============================================================================
// Exponentiation
// ============================================================================

// Exp represents the a given value taken to a power.
type Exp struct {
	Arg Expr
	Pow uint64
}

// Multiplicity determines the number of values that evaluating this expression
// can generate.
func (e *Exp) Multiplicity() uint {
	return determineMultiplicity([]Expr{e.Arg})
}

// Context returns the context for this expression.  Observe that the
// expression must have been resolved for this to be defined (i.e. it may
// panic if it has not been resolved yet).
func (e *Exp) Context() Context {
	return ContextOfExpressions([]Expr{e.Arg})
}

// Lisp converts this schema element into a simple S-Expression, for example
// so it can be printed.
func (e *Exp) Lisp() sexp.SExp {
	panic("todo")
}

// Substitute all variables (such as for function parameters) arising in
// this expression.
func (e *Exp) Substitute(args []Expr) Expr {
	return &Exp{e.Arg.Substitute(args), e.Pow}
}

// Dependencies needed to signal declaration.
func (e *Exp) Dependencies() []Symbol {
	return e.Arg.Dependencies()
}

// ============================================================================
// IfZero
// ============================================================================

// IfZero returns the (optional) true branch when the condition evaluates to zero, and
// the (optional false branch otherwise.
type IfZero struct {
	// Elements contained within this list.
	Condition Expr
	// True branch (optional).
	TrueBranch Expr
	// False branch (optional).
	FalseBranch Expr
}

// Multiplicity determines the number of values that evaluating this expression
// can generate.
func (e *IfZero) Multiplicity() uint {
	return determineMultiplicity([]Expr{e.Condition, e.TrueBranch, e.FalseBranch})
}

// Context returns the context for this expression.  Observe that the
// expression must have been resolved for this to be defined (i.e. it may
// panic if it has not been resolved yet).
func (e *IfZero) Context() Context {
	return ContextOfExpressions([]Expr{e.Condition, e.TrueBranch, e.FalseBranch})
}

// Lisp converts this schema element into a simple S-Expression, for example
// so it can be printed.
func (e *IfZero) Lisp() sexp.SExp {
	panic("todo")
}

// Substitute all variables (such as for function parameters) arising in
// this expression.
func (e *IfZero) Substitute(args []Expr) Expr {
	return &IfZero{e.Condition.Substitute(args),
		SubstituteOptionalExpression(e.TrueBranch, args),
		SubstituteOptionalExpression(e.FalseBranch, args),
	}
}

// Dependencies needed to signal declaration.
func (e *IfZero) Dependencies() []Symbol {
	return DependenciesOfExpressions([]Expr{e.Condition, e.TrueBranch, e.FalseBranch})
}

// ============================================================================
// List
// ============================================================================

// List represents a block of zero or more expressions.
type List struct{ Args []Expr }

// Multiplicity determines the number of values that evaluating this expression
// can generate.
func (e *List) Multiplicity() uint {
	return determineMultiplicity(e.Args)
}

// Context returns the context for this expression.  Observe that the
// expression must have been resolved for this to be defined (i.e. it may
// panic if it has not been resolved yet).
func (e *List) Context() Context {
	return ContextOfExpressions(e.Args)
}

// Lisp converts this schema element into a simple S-Expression, for example
// so it can be printed.
func (e *List) Lisp() sexp.SExp {
	panic("todo")
}

// Substitute all variables (such as for function parameters) arising in
// this expression.
func (e *List) Substitute(args []Expr) Expr {
	return &List{SubstituteExpressions(e.Args, args)}
}

// Dependencies needed to signal declaration.
func (e *List) Dependencies() []Symbol {
	return DependenciesOfExpressions(e.Args)
}

// ============================================================================
// Multiplication
// ============================================================================

// Mul represents the product over zero or more expressions.
type Mul struct{ Args []Expr }

// Multiplicity determines the number of values that evaluating this expression
// can generate.
func (e *Mul) Multiplicity() uint {
	return determineMultiplicity(e.Args)
}

// Context returns the context for this expression.  Observe that the
// expression must have been resolved for this to be defined (i.e. it may
// panic if it has not been resolved yet).
func (e *Mul) Context() Context {
	return ContextOfExpressions(e.Args)
}

// Lisp converts this schema element into a simple S-Expression, for example
// so it can be printed.
func (e *Mul) Lisp() sexp.SExp {
	panic("todo")
}

// Substitute all variables (such as for function parameters) arising in
// this expression.
func (e *Mul) Substitute(args []Expr) Expr {
	return &Mul{SubstituteExpressions(e.Args, args)}
}

// Dependencies needed to signal declaration.
func (e *Mul) Dependencies() []Symbol {
	return DependenciesOfExpressions(e.Args)
}

// ============================================================================
// Normalise
// ============================================================================

// Normalise reduces the value of an expression to either zero (if it was zero)
// or one (otherwise).
type Normalise struct{ Arg Expr }

// Multiplicity determines the number of values that evaluating this expression
// can generate.
func (e *Normalise) Multiplicity() uint {
	return determineMultiplicity([]Expr{e.Arg})
}

// Context returns the context for this expression.  Observe that the
// expression must have been resolved for this to be defined (i.e. it may
// panic if it has not been resolved yet).
func (e *Normalise) Context() Context {
	return ContextOfExpressions([]Expr{e.Arg})
}

// Lisp converts this schema element into a simple S-Expression, for example
// so it can be printed.
func (e *Normalise) Lisp() sexp.SExp {
	panic("todo")
}

// Substitute all variables (such as for function parameters) arising in
// this expression.
func (e *Normalise) Substitute(args []Expr) Expr {
	return &Normalise{e.Arg.Substitute(args)}
}

// Dependencies needed to signal declaration.
func (e *Normalise) Dependencies() []Symbol {
	return e.Arg.Dependencies()
}

// ============================================================================
// Subtraction
// ============================================================================

// Sub represents the subtraction over zero or more expressions.
type Sub struct{ Args []Expr }

// Multiplicity determines the number of values that evaluating this expression
// can generate.
func (e *Sub) Multiplicity() uint {
	return determineMultiplicity(e.Args)
}

// Context returns the context for this expression.  Observe that the
// expression must have been resolved for this to be defined (i.e. it may
// panic if it has not been resolved yet).
func (e *Sub) Context() Context {
	return ContextOfExpressions(e.Args)
}

// Lisp converts this schema element into a simple S-Expression, for example
// so it can be printed.
func (e *Sub) Lisp() sexp.SExp {
	panic("todo")
}

// Substitute all variables (such as for function parameters) arising in
// this expression.
func (e *Sub) Substitute(args []Expr) Expr {
	return &Sub{SubstituteExpressions(e.Args, args)}
}

// Dependencies needed to signal declaration.
func (e *Sub) Dependencies() []Symbol {
	return DependenciesOfExpressions(e.Args)
}

// ============================================================================
// Function Invocation
// ============================================================================

// Invoke represents an attempt to invoke a given function.
type Invoke struct {
	module  *string
	name    string
	args    []Expr
	binding *FunctionBinding
}

// IsQualified determines whether this symbol is qualfied or not (i.e. has an
// explicitly module specifier).
func (e *Invoke) IsQualified() bool {
	return e.module != nil
}

// IsFunction indicates whether or not this symbol refers to a function (which
// of course it always does).
func (e *Invoke) IsFunction() bool {
	return true
}

// IsResolved checks whether this symbol has been resolved already, or not.
func (e *Invoke) IsResolved() bool {
	return e.binding != nil
}

// Resolve this symbol by associating it with the binding associated with
// the definition of the symbol to which this refers.
func (e *Invoke) Resolve(binding Binding) {
	if fb, ok := binding.(*FunctionBinding); ok {
		e.binding = fb
		return
	}
	// Problem
	panic("cannot resolve function invocation with anything other than a function binding")
}

// Module returns the optional module qualification.  This will panic if this
// invocation is unqualified.
func (e *Invoke) Module() string {
	if e.module == nil {
		panic("invocation has no module qualifier")
	}

	return *e.module
}

// Name of the function being invoked.
func (e *Invoke) Name() string {
	return e.name
}

// Args returns the arguments provided by this invocation to the function being
// invoked.
func (e *Invoke) Args() []Expr {
	return e.args
}

// Binding gets binding associated with this interface.  This will panic if this
// symbol is not yet resolved.
func (e *Invoke) Binding() Binding {
	if e.binding == nil {
		panic("invocation not yet resolved")
	}

	return e.binding
}

// Context returns the context for this expression.  Observe that the
// expression must have been resolved for this to be defined (i.e. it may
// panic if it has not been resolved yet).
func (e *Invoke) Context() Context {
	if e.binding == nil {
		panic("unresolved expressions encountered whilst resolving context")
	}
	// TODO: impure functions can have their own context.
	return ContextOfExpressions(e.args)
}

// Multiplicity determines the number of values that evaluating this expression
// can generate.
func (e *Invoke) Multiplicity() uint {
	// FIXME: is this always correct?
	return 1
}

// Lisp converts this schema element into a simple S-Expression, for example
// so it can be printed.
func (e *Invoke) Lisp() sexp.SExp {
	panic("todo")
}

// Substitute all variables (such as for function parameters) arising in
// this expression.
func (e *Invoke) Substitute(args []Expr) Expr {
	return &Invoke{e.module, e.name, SubstituteExpressions(e.args, args), e.binding}
}

// Dependencies needed to signal declaration.
func (e *Invoke) Dependencies() []Symbol {
	deps := DependenciesOfExpressions(e.args)
	// Include this expression as a symbol (which must be bound to the function
	// being invoked)
	return append(deps, e)
}

// ============================================================================
// VariableAccess
// ============================================================================

// VariableAccess represents reading the value of a given local variable (such
// as a function parameter).
type VariableAccess struct {
	module  *string
	name    string
	shift   int
	binding Binding
}

// IsQualified determines whether this symbol is qualfied or not (i.e. has an
// explicitly module specifier).
func (e *VariableAccess) IsQualified() bool {
	return e.module != nil
}

// IsFunction determines whether this symbol refers to a function (which, of
// course, variable accesses never do).
func (e *VariableAccess) IsFunction() bool {
	return false
}

// IsResolved checks whether this symbol has been resolved already, or not.
func (e *VariableAccess) IsResolved() bool {
	return e.binding != nil
}

// Resolve this symbol by associating it with the binding associated with
// the definition of the symbol to which this refers.
func (e *VariableAccess) Resolve(binding Binding) {
	if binding == nil {
		panic("empty binding")
	} else if e.binding != nil {
		panic("already resolved")
	}

	e.binding = binding
}

// Module returns the optional module qualification.  This will panic if this
// invocation is unqualified.
func (e *VariableAccess) Module() string {
	return *e.module
}

// Name returns the (unqualified) name of this symbol
func (e *VariableAccess) Name() string {
	return e.name
}

// Binding gets binding associated with this interface.  This will panic if this
// symbol is not yet resolved.
func (e *VariableAccess) Binding() Binding {
	if e.binding == nil {
		panic("variable access is unresolved")
	}
	//
	return e.binding
}

// Shift returns the row shift (if any) associated with this variable access.
func (e *VariableAccess) Shift() int {
	return e.shift
}

// Multiplicity determines the number of values that evaluating this expression
// can generate.
func (e *VariableAccess) Multiplicity() uint {
	return 1
}

// Context returns the context for this expression.  Observe that the
// expression must have been resolved for this to be defined (i.e. it may
// panic if it has not been resolved yet).
func (e *VariableAccess) Context() Context {
	binding, ok := e.binding.(*ColumnBinding)
	//
	if ok {
		return binding.Context()
	}
	//
	panic("invalid column access")
}

// Lisp converts this schema element into a simple S-Expression, for example
// so it can be printed.a
func (e *VariableAccess) Lisp() sexp.SExp {
	panic("todo")
}

// Substitute all variables (such as for function parameters) arising in
// this expression.
func (e *VariableAccess) Substitute(args []Expr) Expr {
	if b, ok := e.binding.(*ParameterBinding); ok {
		// This is a variable to be substituted.
		if e.shift != 0 {
			panic("support variable shifts")
		}
		//
		return args[b.index]
	}
	// Nothing to do here
	return e
}

// Dependencies needed to signal declaration.
func (e *VariableAccess) Dependencies() []Symbol {
	return []Symbol{e}
}

// ============================================================================
// Helpers
// ============================================================================

// ContextOfExpressions returns the context for a set of zero or more
// expressions.  Observe that, if there the expressions have no context (i.e.
// they are all constants) then the void context is returned.  Likewise, if
// there are expressions with different contexts then the conflicted context
// will be returned.  Otherwise, the one consistent context will be returned.
func ContextOfExpressions(exprs []Expr) Context {
	context := tr.VoidContext[string]()
	//
	for _, e := range exprs {
		context = context.Join(e.Context())
	}
	//
	return context
}

// SubstituteExpressions substitutes all variables found in a given set of
// expressions.
func SubstituteExpressions(exprs []Expr, vars []Expr) []Expr {
	nexprs := make([]Expr, len(exprs))
	//
	for i := 0; i < len(nexprs); i++ {
		nexprs[i] = exprs[i].Substitute(vars)
	}
	//
	return nexprs
}

// SubstituteOptionalExpression substitutes through an expression which is
// optional (i.e. might be nil).  In such case, nil is returned.
func SubstituteOptionalExpression(expr Expr, vars []Expr) Expr {
	if expr != nil {
		expr = expr.Substitute(vars)
	}
	//
	return expr
}

// DependenciesOfExpressions determines the dependencies for a given set of zero
// or more expressions.
func DependenciesOfExpressions(exprs []Expr) []Symbol {
	var deps []Symbol
	//
	for _, e := range exprs {
		if e != nil {
			deps = append(deps, e.Dependencies()...)
		}
	}
	//
	return deps
}

func determineMultiplicity(exprs []Expr) uint {
	width := uint(1)
	//
	for _, e := range exprs {
		if e != nil {
			width *= e.Multiplicity()
		}
	}
	//
	return width
}
