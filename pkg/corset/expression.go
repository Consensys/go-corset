package corset

import (
	"fmt"
	"math/big"

	"github.com/consensys/go-corset/pkg/sexp"
	tr "github.com/consensys/go-corset/pkg/trace"
)

// Expr represents an arbitrary expression over the columns of a given context
// (or the parameters of an enclosing function).  Such expressions are pitched
// at a higher-level than those of the underlying constraint system.  For
// example, they can contain conditionals (i.e. if expressions) and
// normalisations, etc.  During the lowering process down to the underlying
// constraints level (AIR), such expressions are "compiled out" using various
// techniques (such as introducing computed columns where necessary).
type Expr interface {
	Node
	// Evaluates this expression as a constant (signed) value.  If this
	// expression is not constant, then nil is returned.
	AsConstant() *big.Int
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

// AsConstant attempts to evaluate this expression as a constant (signed) value.
// If this expression is not constant, then nil is returned.
func (e *Add) AsConstant() *big.Int {
	fn := func(l *big.Int, r *big.Int) *big.Int { l.Add(l, r); return l }
	return AsConstantOfExpressions(e.Args, fn)
}

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
	return ListOfExpressions(sexp.NewSymbol("+"), e.Args)
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
type Constant struct{ Val big.Int }

// AsConstant attempts to evaluate this expression as a constant (signed) value.
// If this expression is not constant, then nil is returned.
func (e *Constant) AsConstant() *big.Int {
	return &e.Val
}

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
	Pow Expr
}

// AsConstant attempts to evaluate this expression as a constant (signed) value.
// If this expression is not constant, then nil is returned.
func (e *Exp) AsConstant() *big.Int {
	arg := e.Arg.AsConstant()
	pow := e.Pow.AsConstant()
	// Check if can evaluate
	if arg != nil && pow != nil {
		return arg.Exp(arg, pow, nil)
	}
	//
	return nil
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
	return ContextOfExpressions([]Expr{e.Arg, e.Pow})
}

// Lisp converts this schema element into a simple S-Expression, for example
// so it can be printed.
func (e *Exp) Lisp() sexp.SExp {
	return sexp.NewList([]sexp.SExp{
		sexp.NewSymbol("^"),
		e.Arg.Lisp(),
		e.Pow.Lisp()})
}

// Substitute all variables (such as for function parameters) arising in
// this expression.
func (e *Exp) Substitute(args []Expr) Expr {
	return &Exp{e.Arg.Substitute(args), e.Pow}
}

// Dependencies needed to signal declaration.
func (e *Exp) Dependencies() []Symbol {
	return DependenciesOfExpressions([]Expr{e.Arg, e.Pow})
}

// ============================================================================
// IfZero
// ============================================================================

// If returns the (optional) true branch when the condition evaluates to zero, and
// the (optional false branch otherwise.
type If struct {
	// Indicates whether this is an if-zero (kind==1) or an if-notzero
	// (kind==2).  Any other kind value implies this has not yet been
	// determined.
	kind uint8
	// Elements contained within this list.
	Condition Expr
	// True branch (optional).
	TrueBranch Expr
	// False branch (optional).
	FalseBranch Expr
}

// IsIfZero determines whether or not this has been determined as an IfZero
// condition.
func (e *If) IsIfZero() bool {
	return e.kind == 1
}

// IsIfNotZero determines whether or not this has been determined as an
// IfNotZero condition.
func (e *If) IsIfNotZero() bool {
	return e.kind == 2
}

// FixSemantics fixes the semantics for this condition to be either "if-zero" or
// "if-notzero".
func (e *If) FixSemantics(ifzero bool) {
	if ifzero {
		e.kind = 1
	} else {
		e.kind = 2
	}
}

// AsConstant attempts to evaluate this expression as a constant (signed) value.
// If this expression is not constant, then nil is returned.
func (e *If) AsConstant() *big.Int {
	if condition := e.Condition.AsConstant(); condition != nil {
		// Determine whether condition holds true (or not).
		holds := condition.Cmp(big.NewInt(0)) == 0
		//
		if holds && e.TrueBranch != nil {
			return e.TrueBranch.AsConstant()
		} else if !holds && e.FalseBranch != nil {
			return e.FalseBranch.AsConstant()
		}
	}
	//
	return nil
}

// Multiplicity determines the number of values that evaluating this expression
// can generate.
func (e *If) Multiplicity() uint {
	return determineMultiplicity([]Expr{e.Condition, e.TrueBranch, e.FalseBranch})
}

// Context returns the context for this expression.  Observe that the
// expression must have been resolved for this to be defined (i.e. it may
// panic if it has not been resolved yet).
func (e *If) Context() Context {
	return ContextOfExpressions([]Expr{e.Condition, e.TrueBranch, e.FalseBranch})
}

// Lisp converts this schema element into a simple S-Expression, for example
// so it can be printed.
func (e *If) Lisp() sexp.SExp {
	if e.FalseBranch != nil {
		return sexp.NewList([]sexp.SExp{
			sexp.NewSymbol("if"),
			e.TrueBranch.Lisp(),
			e.FalseBranch.Lisp()})
	}
	//
	return sexp.NewList([]sexp.SExp{
		sexp.NewSymbol("if"),
		e.TrueBranch.Lisp()})
}

// Substitute all variables (such as for function parameters) arising in
// this expression.
func (e *If) Substitute(args []Expr) Expr {
	return &If{e.kind, e.Condition.Substitute(args),
		SubstituteOptionalExpression(e.TrueBranch, args),
		SubstituteOptionalExpression(e.FalseBranch, args),
	}
}

// Dependencies needed to signal declaration.
func (e *If) Dependencies() []Symbol {
	return DependenciesOfExpressions([]Expr{e.Condition, e.TrueBranch, e.FalseBranch})
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

// AsConstant attempts to evaluate this expression as a constant (signed) value.
// If this expression is not constant, then nil is returned.
func (e *Invoke) AsConstant() *big.Int {
	if e.binding == nil {
		panic("unresolved invocation")
	}
	// Unroll body
	body := e.binding.Apply(e.args)
	// Attempt to evaluate as constant
	return body.AsConstant()
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
func (e *Invoke) Resolve(binding Binding) bool {
	if fb, ok := binding.(*FunctionBinding); ok {
		e.binding = fb
		return true
	}
	// Problem
	return false
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
	var fn sexp.SExp
	if e.module != nil {
		fn = sexp.NewSymbol(fmt.Sprintf("%s.%s", *e.module, e.name))
	} else {
		fn = sexp.NewSymbol(e.name)
	}

	return ListOfExpressions(fn, e.args)
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
// List
// ============================================================================

// List represents a block of zero or more expressions.
type List struct{ Args []Expr }

// AsConstant attempts to evaluate this expression as a constant (signed) value.
// If this expression is not constant, then nil is returned.
func (e *List) AsConstant() *big.Int {
	// Potentially we could do better here, but its not clear we need to.
	return nil
}

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
	return ListOfExpressions(sexp.NewSymbol("begin"), e.Args)
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

// AsConstant attempts to evaluate this expression as a constant (signed) value.
// If this expression is not constant, then nil is returned.
func (e *Mul) AsConstant() *big.Int {
	fn := func(l *big.Int, r *big.Int) *big.Int { l.Mul(l, r); return l }
	return AsConstantOfExpressions(e.Args, fn)
}

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
	return ListOfExpressions(sexp.NewSymbol("*"), e.Args)
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

// AsConstant attempts to evaluate this expression as a constant (signed) value.
// If this expression is not constant, then nil is returned.
func (e *Normalise) AsConstant() *big.Int {
	// FIXME: we could do better here.
	return nil
}

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
	return sexp.NewList([]sexp.SExp{
		sexp.NewSymbol("~"),
		e.Arg.Lisp()})
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

// AsConstant attempts to evaluate this expression as a constant (signed) value.
// If this expression is not constant, then nil is returned.
func (e *Sub) AsConstant() *big.Int {
	fn := func(l *big.Int, r *big.Int) *big.Int { l.Sub(l, r); return l }
	return AsConstantOfExpressions(e.Args, fn)
}

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
	return ListOfExpressions(sexp.NewSymbol("-"), e.Args)
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
// Shift
// ============================================================================

// Shift represents the result of a given expression shifted by a certain
// amount.  In reality, the shift amount must be statically known.  However, it
// is represented here as an expression to allow for constants and the results
// of function invocations, etc to be used.  In all cases, these must still be
// eventually translated into constant values however.
type Shift struct {
	// The expression being shifted
	Arg Expr
	// The amount it is being shifted by.
	Shift Expr
}

// AsConstant attempts to evaluate this expression as a constant (signed) value.
// If this expression is not constant, then nil is returned.
func (e *Shift) AsConstant() *big.Int {
	// Observe the shift doesn't matter as, in the case that the argument is a
	// constant, then the shift has no effect anyway.
	return e.Arg.AsConstant()
}

// Multiplicity determines the number of values that evaluating this expression
// can generate.
func (e *Shift) Multiplicity() uint {
	return determineMultiplicity([]Expr{e.Arg})
}

// Context returns the context for this expression.  Observe that the
// expression must have been resolved for this to be defined (i.e. it may
// panic if it has not been resolved yet).
func (e *Shift) Context() Context {
	return ContextOfExpressions([]Expr{e.Arg, e.Shift})
}

// Lisp converts this schema element into a simple S-Expression, for example
// so it can be printed.
func (e *Shift) Lisp() sexp.SExp {
	return sexp.NewList([]sexp.SExp{
		sexp.NewSymbol("shift"),
		e.Arg.Lisp(),
		e.Shift.Lisp()})
}

// Substitute all variables (such as for function parameters) arising in
// this expression.
func (e *Shift) Substitute(args []Expr) Expr {
	return &Shift{e.Arg.Substitute(args), e.Shift.Substitute(args)}
}

// Dependencies needed to signal declaration.
func (e *Shift) Dependencies() []Symbol {
	return DependenciesOfExpressions([]Expr{e.Arg, e.Shift})
}

// ============================================================================
// VariableAccess
// ============================================================================

// VariableAccess represents reading the value of a given local variable (such
// as a function parameter).
type VariableAccess struct {
	module  *string
	name    string
	binding Binding
}

// AsConstant attempts to evaluate this expression as a constant (signed) value.
// If this expression is not constant, then nil is returned.
func (e *VariableAccess) AsConstant() *big.Int {
	if binding, ok := e.binding.(*ConstantBinding); ok {
		return binding.value.AsConstant()
	}
	// not a constant
	return nil
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
func (e *VariableAccess) Resolve(binding Binding) bool {
	if binding == nil {
		panic("empty binding")
	} else if e.binding != nil {
		panic("already resolved")
	}
	//
	e.binding = binding
	//
	return true
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
	var name string
	if e.module != nil {
		name = fmt.Sprintf("%s.%s", *e.module, e.name)
	} else {
		name = e.name
	}
	//
	return sexp.NewSymbol(name)
}

// Substitute all variables (such as for function parameters) arising in
// this expression.
func (e *VariableAccess) Substitute(args []Expr) Expr {
	if b, ok := e.binding.(*ParameterBinding); ok {
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

// ListOfExpressions converts an array of one or more expressions into a list of
// corresponding lisp expressions.
func ListOfExpressions(head sexp.SExp, exprs []Expr) *sexp.List {
	lisps := make([]sexp.SExp, len(exprs)+1)
	// Assign head
	lisps[0] = head
	//
	for i, e := range exprs {
		lisps[i+1] = e.Lisp()
	}
	//
	return sexp.NewList(lisps)
}

// AsConstantOfExpressions attempts to fold one or more expressions across a
// given operation (e.g. add, subtract, etc) to produce a constant value.  If
// any of the expressions are not themselves constant, then neither is the
// result.
func AsConstantOfExpressions(exprs []Expr, fn func(*big.Int, *big.Int) *big.Int) *big.Int {
	var val *big.Int = big.NewInt(0)
	//
	for _, arg := range exprs {
		c := arg.AsConstant()
		if c == nil {
			return nil
		}
		// Evaluate function
		val = fn(val, c)
	}
	//
	return val
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
