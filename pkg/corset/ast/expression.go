// Copyright Consensys Software Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License"); you may not use this file except in compliance with
// the License. You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
// an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
// specific language governing permissions and limitations under the License.
//
// SPDX-License-Identifier: Apache-2.0
package ast

import (
	"fmt"
	"math/big"
	"reflect"

	tr "github.com/consensys/go-corset/pkg/trace"
	"github.com/consensys/go-corset/pkg/util"
	"github.com/consensys/go-corset/pkg/util/source"
	"github.com/consensys/go-corset/pkg/util/source/sexp"
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
	fn := func(l *big.Int, r *big.Int) { l.Add(l, r) }
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
	ctx, _ := ContextOfExpressions(e.Args...)
	return ctx
}

// Lisp converts this schema element into a simple S-Expression, for example
// so it can be printed.
func (e *Add) Lisp() sexp.SExp {
	return ListOfExpressions(sexp.NewSymbol("+"), e.Args)
}

// Dependencies needed to signal declaration.
func (e *Add) Dependencies() []Symbol {
	return DependenciesOfExpressions(e.Args)
}

// ============================================================================
// ArrayAccess
// ============================================================================

// ArrayAccess represents the a given value taken to a power.
type ArrayAccess struct {
	Name         util.Path
	Arg          Expr
	ArrayBinding Binding
}

// IsFunction indicates whether or not this symbol refers to a function (which
// of course it always does).
func (e *ArrayAccess) IsFunction() bool {
	return false
}

// IsResolved checks whether this symbol has been resolved already, or not.
func (e *ArrayAccess) IsResolved() bool {
	return e.ArrayBinding != nil
}

// AsConstant attempts to evaluate this expression as a constant (signed) value.
// If this expression is not constant, then nil is returned.
func (e *ArrayAccess) AsConstant() *big.Int {
	return nil
}

// Multiplicity determines the number of values that evaluating this expression
// can generate.
func (e *ArrayAccess) Multiplicity() uint {
	return determineMultiplicity([]Expr{e.Arg})
}

// Path returns the given path of this symbol.
func (e *ArrayAccess) Path() *util.Path {
	return &e.Name
}

// Binding gets binding associated with this interface.  This will panic if this
// symbol is not yet resolved.
func (e *ArrayAccess) Binding() Binding {
	return e.ArrayBinding
}

// Type returns the type associated with this symbol.  If the type cannot be
// determined, then nil is returned.
func (e *ArrayAccess) Type() Type {
	if binding, ok := e.ArrayBinding.(*ColumnBinding); !ok {
		return nil
	} else if arr_t, ok := binding.DataType.(*ArrayType); ok {
		return arr_t.element
	}
	// Cannot be typed
	return nil
}

// Context returns the context for this expression.  Observe that the
// expression must have been resolved for this to be defined (i.e. it may
// panic if it has not been resolved yet).
func (e *ArrayAccess) Context() Context {
	return e.Arg.Context()
}

// Lisp converts this schema element into a simple S-Expression, for example
// so it can be printed.
func (e *ArrayAccess) Lisp() sexp.SExp {
	return sexp.NewArray([]sexp.SExp{
		sexp.NewSymbol(e.Name.String()),
		e.Arg.Lisp(),
	})
}

// Resolve this symbol by associating it with the binding associated with
// the definition of the symbol to which this refers.
func (e *ArrayAccess) Resolve(binding Binding) bool {
	if binding == nil {
		panic("empty binding")
	} else if e.ArrayBinding != nil {
		panic("already resolved")
	}
	//
	e.ArrayBinding = binding
	//
	return true
}

// Dependencies needed to signal declaration.
func (e *ArrayAccess) Dependencies() []Symbol {
	deps := e.Arg.Dependencies()
	return append(deps, e)
}

// ============================================================================
// Cast
// ============================================================================

// Cast represents a user-supplied annotation indicating the given expression
// has the given type.  This is only sound upto the user.
type Cast struct {
	Arg      Expr
	BitWidth uint
}

// AsConstant attempts to evaluate this expression as a constant (signed) value.
// If this expression is not constant, then nil is returned.
func (e *Cast) AsConstant() *big.Int {
	return e.Arg.AsConstant()
}

// Multiplicity determines the number of values that evaluating this expression
// can generate.
func (e *Cast) Multiplicity() uint {
	return determineMultiplicity([]Expr{e.Arg})
}

// Context returns the context for this expression.  Observe that the
// expression must have been resolved for this to be defined (i.e. it may
// panic if it has not been resolved yet).
func (e *Cast) Context() Context {
	ctx, _ := ContextOfExpressions(e.Arg)
	return ctx
}

// Lisp converts this schema element into a simple S-Expression, for example
// so it can be printed.
func (e *Cast) Lisp() sexp.SExp {
	return sexp.NewList([]sexp.SExp{
		sexp.NewSymbol(fmt.Sprintf(":u%d", e.BitWidth)),
		e.Arg.Lisp()})
}

// Dependencies needed to signal declaration.
func (e *Cast) Dependencies() []Symbol {
	return e.Arg.Dependencies()
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

// Dependencies needed to signal declaration.
func (e *Constant) Dependencies() []Symbol {
	return nil
}

// ============================================================================
// Normalise
// ============================================================================

// Debug is an optional constraint which can be specifically enabled via the
// debug setting.  The intention of debug constraints is that they capture
// things which are implied by other constraints.  The ability to enable them
// can simply help with debugging, should it arise that they are not actually
// implied.
type Debug struct{ Arg Expr }

// AsConstant attempts to evaluate this expression as a constant (signed) value.
// If this expression is not constant, then nil is returned.
func (e *Debug) AsConstant() *big.Int {
	return nil
}

// Multiplicity determines the number of values that evaluating this expression
// can generate.
func (e *Debug) Multiplicity() uint {
	return determineMultiplicity([]Expr{e.Arg})
}

// Context returns the context for this expression.  Observe that the
// expression must have been resolved for this to be defined (i.e. it may
// panic if it has not been resolved yet).
func (e *Debug) Context() Context {
	return e.Arg.Context()
}

// Lisp converts this schema element into a simple S-Expression, for example
// so it can be printed.
func (e *Debug) Lisp() sexp.SExp {
	return sexp.NewList([]sexp.SExp{
		sexp.NewSymbol("debug"),
		e.Arg.Lisp()})
}

// Dependencies needed to signal declaration.
func (e *Debug) Dependencies() []Symbol {
	return e.Arg.Dependencies()
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
		var res big.Int
		// Compute exponent
		res.Exp(arg, pow, nil)
		// Done
		return &res
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
	ctx, _ := ContextOfExpressions(e.Arg, e.Pow)
	return ctx
}

// Lisp converts this schema element into a simple S-Expression, for example
// so it can be printed.
func (e *Exp) Lisp() sexp.SExp {
	return sexp.NewList([]sexp.SExp{
		sexp.NewSymbol("^"),
		e.Arg.Lisp(),
		e.Pow.Lisp()})
}

// Dependencies needed to signal declaration.
func (e *Exp) Dependencies() []Symbol {
	return DependenciesOfExpressions([]Expr{e.Arg, e.Pow})
}

// ============================================================================
// For
// ============================================================================

// For represents a for loop of a statically known range of values
type For struct {
	// Variable binding
	Binding LocalVariableBinding
	// Start value for Index
	Start uint
	// Last Value for Index
	End uint
	// Body of loop
	Body Expr
}

// NewFor constructs a new for-expression given a variable name, a static index
// range and a body.
func NewFor(name string, start uint, end uint, body Expr) *For {
	binding := NewLocalVariableBinding(name, NewFieldType())
	return &For{binding, start, end, body}
}

// AsConstant attempts to evaluate this expression as a constant (signed) value.
// If this expression is not constant, then nil is returned.
func (e *For) AsConstant() *big.Int {
	body := e.Body.AsConstant()
	// Check if can evaluate
	if body != nil {
		return body
	}
	//
	return nil
}

// Multiplicity determines the number of values that evaluating this expression
// can generate.
func (e *For) Multiplicity() uint {
	return e.End - e.Start + 1
}

// Context returns the context for this expression.  Observe that the
// expression must have been resolved for this to be defined (i.e. it may
// panic if it has not been resolved yet).
func (e *For) Context() Context {
	return e.Body.Context()
}

// Lisp converts this schema element into a simple S-Expression, for example
// so it can be printed.
func (e *For) Lisp() sexp.SExp {
	return sexp.NewList([]sexp.SExp{
		sexp.NewSymbol("for"),
		sexp.NewSymbol("..."),
		e.Body.Lisp(),
	})
}

// Dependencies needed to signal declaration.
func (e *For) Dependencies() []Symbol {
	// Remove occurrences of the index variable defined by this expression.  In
	// essence, we are capturing this occurrences of this symbol.
	var rest []Symbol
	//
	for _, s := range e.Body.Dependencies() {
		p := s.Path()
		if p.IsAbsolute() || p.Depth() != 1 || p.Head() != e.Binding.Name {
			rest = append(rest, s)
		}
	}
	//
	return rest
}

// ============================================================================
// IfZero
// ============================================================================

// If returns the (optional) true branch when the condition evaluates to zero, and
// the (optional false branch otherwise.
type If struct {
	// Indicates whether this is an if-zero (Kind==1) or an if-notzero
	// (Kind==2).  Any other Kind value implies this has not yet been
	// determined.
	Kind uint8
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
	return e.Kind == 1
}

// IsIfNotZero determines whether or not this has been determined as an
// IfNotZero condition.
func (e *If) IsIfNotZero() bool {
	return e.Kind == 2
}

// FixSemantics fixes the semantics for this condition to be either "if-zero" or
// "if-notzero".
func (e *If) FixSemantics(ifzero bool) {
	if ifzero {
		e.Kind = 1
	} else {
		e.Kind = 2
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
	ctx, _ := ContextOfExpressions(e.Condition, e.TrueBranch, e.FalseBranch)
	return ctx
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

// Dependencies needed to signal declaration.
func (e *If) Dependencies() []Symbol {
	return DependenciesOfExpressions([]Expr{e.Condition, e.TrueBranch, e.FalseBranch})
}

// ============================================================================
// Function Invocation
// ============================================================================

// Invoke represents an attempt to invoke a given function.
type Invoke struct {
	Name      *VariableAccess
	Signature *FunctionSignature
	Args      []Expr
}

// AsConstant attempts to evaluate this expression as a constant (signed) value.
// If this expression is not constant, then nil is returned.
func (e *Invoke) AsConstant() *big.Int {
	if e.Signature == nil {
		panic("unresolved invocation")
	}
	// Unroll body
	body := e.Signature.Apply(e.Args, nil)
	// Attempt to evaluate as constant
	return body.AsConstant()
}

// Context returns the context for this expression.  Observe that the
// expression must have been resolved for this to be defined (i.e. it may
// panic if it has not been resolved yet).
func (e *Invoke) Context() Context {
	ctx, _ := ContextOfExpressions(e.Args...)
	return ctx
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
	return ListOfExpressions(e.Name.Lisp(), e.Args)
}

// Finalise the signature for this invocation.
func (e *Invoke) Finalise(signature *FunctionSignature) {
	e.Signature = signature
}

// Dependencies needed to signal declaration.
func (e *Invoke) Dependencies() []Symbol {
	deps := DependenciesOfExpressions(e.Args)
	// Include this expression as a symbol (which must be bound to the function
	// being invoked)
	return append(deps, e.Name)
}

// ============================================================================
// Let
// ============================================================================

// Let is a common expression form used in programming languages, particularly
// functional languages.  It allows us to assign a "variable" to a given
// expression, such that we can reuse that variable in multiple places rather
// than repeat the entire expression.  Note, however, that such variables are
// functional in nature --- they cannot, for example, be mutated via assignment,
// etc.
type Let struct {
	// The set of variables defined by this expression.
	Vars []LocalVariableBinding
	// Identifies the assigned expression for each variable defined.
	Args []Expr
	// Body of the let expression (i.e. where the variables it defines can be
	// used).
	Body Expr
}

// NewLet constructs a new let-expression for a given number of bindings.
func NewLet(bindings []util.Pair[string, Expr], body Expr) *Let {
	vars := make([]LocalVariableBinding, len(bindings))
	exprs := make([]Expr, len(bindings))
	//
	for i, p := range bindings {
		vars[i] = NewLocalVariableBinding(p.Left, NewFieldType())
		exprs[i] = p.Right
	}
	//
	return &Let{vars, exprs, body}
}

// AsConstant attempts to evaluate this expression as a constant (signed) value.
// If this expression is not constant, then nil is returned.
func (e *Let) AsConstant() *big.Int {
	panic("todo")
}

// Multiplicity determines the number of values that evaluating this expression
// can generate.
func (e *Let) Multiplicity() uint {
	return e.Body.Multiplicity()
}

// Context returns the context for this expression.  Observe that the
// expression must have been resolved for this to be defined (i.e. it may
// panic if it has not been resolved yet).
func (e *Let) Context() Context {
	return e.Body.Context()
}

// Lisp converts this schema element into a simple S-Expression, for example
// so it can be printed.
func (e *Let) Lisp() sexp.SExp {
	bindings := make([]sexp.SExp, len(e.Args))
	//
	for i := range e.Args {
		bindings[i] = sexp.NewList([]sexp.SExp{
			sexp.NewSymbol(e.Vars[i].Name),
			e.Args[i].Lisp(),
		})
	}
	//
	return sexp.NewList([]sexp.SExp{
		sexp.NewSymbol("let"),
		sexp.NewList(bindings),
		e.Body.Lisp(),
	})
}

// Dependencies needed to signal declaration.
func (e *Let) Dependencies() []Symbol {
	// Remove occurrences of the let variables defined by this expression.  In
	// essence, we are capturing these occurrences in the body of this
	// expression.
	var rest []Symbol
	//
	for _, s := range e.Body.Dependencies() {
		p := s.Path()
		if p.IsAbsolute() || p.Depth() != 1 {
			rest = append(rest, s)
		} else {
			matched := false
			// Could be a variable defined here, so check variable names.
			for _, v := range e.Vars {
				if p.Head() == v.Name {
					matched = true
					break
				}
			}
			// Did we match anything?
			if !matched {
				rest = append(rest, s)
			}
		}
	}
	// Determine dependencies for assigned expressions
	return append(rest, DependenciesOfExpressions(e.Args)...)
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
	ctx, _ := ContextOfExpressions(e.Args...)
	return ctx
}

// Lisp converts this schema element into a simple S-Expression, for example
// so it can be printed.
func (e *List) Lisp() sexp.SExp {
	return ListOfExpressions(sexp.NewSymbol("begin"), e.Args)
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
	fn := func(l *big.Int, r *big.Int) { l.Mul(l, r) }
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
	ctx, _ := ContextOfExpressions(e.Args...)
	return ctx
}

// Lisp converts this schema element into a simple S-Expression, for example
// so it can be printed.
func (e *Mul) Lisp() sexp.SExp {
	return ListOfExpressions(sexp.NewSymbol("*"), e.Args)
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
	if arg := e.Arg.AsConstant(); arg != nil {
		if arg.Cmp(big.NewInt(0)) != 0 {
			return big.NewInt(1)
		}
		// zero
		return arg
	}
	//
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
	return e.Arg.Context()
}

// Lisp converts this schema element into a simple S-Expression, for example
// so it can be printed.
func (e *Normalise) Lisp() sexp.SExp {
	return sexp.NewList([]sexp.SExp{
		sexp.NewSymbol("~"),
		e.Arg.Lisp()})
}

// Dependencies needed to signal declaration.
func (e *Normalise) Dependencies() []Symbol {
	return e.Arg.Dependencies()
}

// ============================================================================
// Reduction
// ============================================================================

// Reduce reduces (i.e. folds) a list using a given binary function.
type Reduce struct {
	Name      *VariableAccess
	Signature *FunctionSignature
	Arg       Expr
}

// AsConstant attempts to evaluate this expression as a constant (signed) value.
// If this expression is not constant, then nil is returned.
func (e *Reduce) AsConstant() *big.Int {
	// TODO: potentially we can do better here.
	return nil
}

// Multiplicity determines the number of values that evaluating this expression
// can generate.
func (e *Reduce) Multiplicity() uint {
	return 1
}

// Context returns the context for this expression.  Observe that the
// expression must have been resolved for this to be defined (i.e. it may
// panic if it has not been resolved yet).
func (e *Reduce) Context() Context {
	return e.Arg.Context()
}

// Lisp converts this schema element into a simple S-Expression, for example
// so it can be printed.
func (e *Reduce) Lisp() sexp.SExp {
	return sexp.NewList([]sexp.SExp{
		sexp.NewSymbol("reduce"),
		sexp.NewSymbol(e.Name.Path().String()),
		e.Arg.Lisp()})
}

// Finalise the signature for this reduction.
func (e *Reduce) Finalise(signature *FunctionSignature) {
	if signature == nil {
		panic("cannot finalise with nil signature")
	} else if e.Signature != nil && !reflect.DeepEqual(e.Signature, signature) {
		panic("reduce has already been finalised")
	}

	e.Signature = signature
}

// Dependencies needed to signal declaration.
func (e *Reduce) Dependencies() []Symbol {
	deps := e.Arg.Dependencies()
	return append(deps, e.Name)
}

// ============================================================================
// Subtraction
// ============================================================================

// Sub represents the subtraction over zero or more expressions.
type Sub struct{ Args []Expr }

// AsConstant attempts to evaluate this expression as a constant (signed) value.
// If this expression is not constant, then nil is returned.
func (e *Sub) AsConstant() *big.Int {
	fn := func(l *big.Int, r *big.Int) { l.Sub(l, r) }
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
	ctx, _ := ContextOfExpressions(e.Args...)
	return ctx
}

// Lisp converts this schema element into a simple S-Expression, for example
// so it can be printed.
func (e *Sub) Lisp() sexp.SExp {
	return ListOfExpressions(sexp.NewSymbol("-"), e.Args)
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
	ctx, _ := ContextOfExpressions(e.Arg, e.Shift)
	return ctx
}

// Lisp converts this schema element into a simple S-Expression, for example
// so it can be printed.
func (e *Shift) Lisp() sexp.SExp {
	return sexp.NewList([]sexp.SExp{
		sexp.NewSymbol("shift"),
		e.Arg.Lisp(),
		e.Shift.Lisp()})
}

// Dependencies needed to signal declaration.
func (e *Shift) Dependencies() []Symbol {
	return DependenciesOfExpressions([]Expr{e.Arg, e.Shift})
}

// ============================================================================
// VariableAccess// ============================================================================

// VariableAccess represents reading the value of a given local variable (such
// as a function parameter).
type VariableAccess struct {
	Name     util.Path
	Function bool
	binding  Binding
}

// NewVariableAccess creates a new variable access with the given (optionally
// qualified) path that may (or may not) refer to a function, and which has a
// given initial binding (which can be nil).
func NewVariableAccess(path util.Path, isFunction bool, binding Binding) *VariableAccess {
	return &VariableAccess{path, isFunction, binding}
}

// AsConstant attempts to evaluate this expression as a constant (signed) value.
// If this expression is not constant, then nil is returned.
func (e *VariableAccess) AsConstant() *big.Int {
	if binding, ok := e.binding.(*ConstantBinding); ok && !binding.Extern {
		return binding.Value.AsConstant()
	}
	// not a constant
	return nil
}

// Path returns the given path of this symbol.
func (e *VariableAccess) Path() *util.Path {
	return &e.Name
}

// IsFunction determines whether this symbol refers to a function (which, of
// course, variable accesses never do).
func (e *VariableAccess) IsFunction() bool {
	return e.Function
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
	} else if _, ok := binding.(FunctionBinding); ok && !e.Function {
		return false
	} else if _, ok := binding.(FunctionBinding); !ok && e.Function {
		return false
	}
	//
	e.binding = binding
	//
	return true
}

// Binding gets binding associated with this interface.  This returns nil if the
// access has not already been resolved.
func (e *VariableAccess) Binding() Binding {
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
	// Check the expected options.
	if binding, ok := e.binding.(*ColumnBinding); ok {
		return binding.Context()
	} else if _, ok := e.Binding().(*ConstantBinding); ok {
		return tr.VoidContext[string]()
	} else if _, ok := e.Binding().(*LocalVariableBinding); ok {
		return tr.VoidContext[string]()
	}
	//
	panic("invalid column access")
}

// Lisp converts this schema element into a simple S-Expression, for example
// so it can be printed.a
func (e *VariableAccess) Lisp() sexp.SExp {
	return sexp.NewSymbol(e.Name.String())
}

// Dependencies needed to signal declaration.
func (e *VariableAccess) Dependencies() []Symbol {
	return []Symbol{e}
}

// Type returns the type associated with this symbol.  If the type cannot be
// determined, then nil is returned.
func (e *VariableAccess) Type() Type {
	if binding, ok := e.binding.(*ColumnBinding); ok {
		return binding.DataType
	}
	// Cannot be typed
	return nil
}

// ============================================================================
// Helpers
// ============================================================================

// ContextOfExpressions returns the context for a set of zero or more
// expressions.  Observe that, if there the expressions have no context (i.e.
// they are all constants) then the void context is returned.  Likewise, if
// there are expressions with different contexts then the conflicted context
// will be returned.  Otherwise, the one consistent context will be returned.
func ContextOfExpressions(exprs ...Expr) (Context, uint) {
	context := tr.VoidContext[string]()
	//
	for i, e := range exprs {
		context = context.Join(e.Context())
		//
		if context.IsConflicted() {
			return context, uint(i)
		}
	}
	//
	return context, uint(len(exprs))
}

// Substitute variables (such as for function parameters) in this expression
// based on a mapping of said variables to expressions.  Furthermore, an
// (optional) source map is provided which will be updated, such that the
// freshly created expressions are mapped to their corresponding nodes.
func Substitute(expr Expr, mapping map[uint]Expr, srcmap *source.Maps[Node]) Expr {
	var nexpr Expr
	//
	switch e := expr.(type) {
	case *ArrayAccess:
		arg := Substitute(e.Arg, mapping, srcmap)
		nexpr = &ArrayAccess{e.Name, arg, e.ArrayBinding}
	case *Add:
		args := SubstituteAll(e.Args, mapping, srcmap)
		nexpr = &Add{args}
	case *Cast:
		arg := Substitute(e.Arg, mapping, srcmap)
		nexpr = &Cast{arg, e.BitWidth}
	case *Constant:
		return e
	case *Debug:
		arg := Substitute(e.Arg, mapping, srcmap)
		nexpr = &Debug{arg}
	case *Exp:
		arg := Substitute(e.Arg, mapping, srcmap)
		pow := Substitute(e.Pow, mapping, srcmap)
		// Done
		nexpr = &Exp{arg, pow}
	case *For:
		body := Substitute(e.Body, mapping, srcmap)
		nexpr = &For{e.Binding, e.Start, e.End, body}
	case *If:
		condition := Substitute(e.Condition, mapping, srcmap)
		trueBranch := SubstituteOptional(e.TrueBranch, mapping, srcmap)
		falseBranch := SubstituteOptional(e.FalseBranch, mapping, srcmap)
		// Construct appropriate if form
		nexpr = &If{e.Kind, condition, trueBranch, falseBranch}
	case *Invoke:
		args := SubstituteAll(e.Args, mapping, srcmap)
		nexpr = &Invoke{e.Name, e.Signature, args}
	case *Let:
		args := SubstituteAll(e.Args, mapping, srcmap)
		body := Substitute(e.Body, mapping, srcmap)
		nexpr = &Let{e.Vars, args, body}
	case *List:
		args := SubstituteAll(e.Args, mapping, srcmap)
		nexpr = &List{args}
	case *Mul:
		args := SubstituteAll(e.Args, mapping, srcmap)
		nexpr = &Mul{args}
	case *Normalise:
		arg := Substitute(e.Arg, mapping, srcmap)
		nexpr = &Normalise{arg}
	case *Reduce:
		arg := Substitute(e.Arg, mapping, srcmap)
		nexpr = &Reduce{e.Name, e.Signature, arg}
	case *Sub:
		args := SubstituteAll(e.Args, mapping, srcmap)
		nexpr = &Sub{args}
	case *Shift:
		arg := Substitute(e.Arg, mapping, srcmap)
		shift := Substitute(e.Shift, mapping, srcmap)
		nexpr = &Shift{arg, shift}
	case *VariableAccess:
		//
		if b, ok1 := e.binding.(*LocalVariableBinding); !ok1 {
			return e
		} else if e2, ok2 := mapping[b.Index]; !ok2 {
			return e
		} else {
			// Shallow copy the node to ensure it is unique and, hence, can have
			// the source mapping associated with e.
			nexpr = ShallowCopy(e2)
		}
	default:
		panic(fmt.Sprintf("unknown expression (%s)", reflect.TypeOf(expr)))
	}
	//
	if srcmap != nil {
		// Copy over source information
		srcmap.Copy(expr, nexpr)
	}
	// Done
	return nexpr
}

// SubstituteAll substitutes all variables found in a given set of
// expressions.
func SubstituteAll(exprs []Expr, mapping map[uint]Expr, srcmap *source.Maps[Node]) []Expr {
	nexprs := make([]Expr, len(exprs))
	//
	for i := 0; i < len(nexprs); i++ {
		nexprs[i] = Substitute(exprs[i], mapping, srcmap)
	}
	//
	return nexprs
}

// SubstituteOptional substitutes through an expression which is
// optional (i.e. might be nil).  In such case, nil is returned.
func SubstituteOptional(expr Expr, mapping map[uint]Expr, srcmap *source.Maps[Node]) Expr {
	if expr != nil {
		expr = Substitute(expr, mapping, srcmap)
	}
	//
	return expr
}

// ShallowCopy creates a copy of the expression itself, but not those
// expressions it contains (if any).  This is useful in e.g. situations where we
// want to assocate different souce file information with a specific expression.
func ShallowCopy(expr Expr) Expr {
	//
	switch e := expr.(type) {
	case *ArrayAccess:
		return &ArrayAccess{e.Name, e.Arg, e.ArrayBinding}
	case *Add:
		return &Add{e.Args}
	case *Cast:
		return &Cast{e.Arg, e.BitWidth}
	case *Constant:
		return &Constant{e.Val}
	case *Debug:
		return &Debug{e.Arg}
	case *Exp:
		return &Exp{e.Arg, e.Pow}
	case *For:
		return &For{e.Binding, e.Start, e.End, e.Body}
	case *If:
		return &If{e.Kind, e.Condition, e.TrueBranch, e.FalseBranch}
	case *Invoke:
		return &Invoke{e.Name, e.Signature, e.Args}
	case *List:
		return &List{e.Args}
	case *Mul:
		return &Mul{e.Args}
	case *Normalise:
		return &Normalise{e.Arg}
	case *Reduce:
		return &Reduce{e.Name, e.Signature, e.Arg}
	case *Sub:
		return &Sub{e.Args}
	case *Shift:
		return &Shift{e.Arg, e.Shift}
	case *VariableAccess:
		return &VariableAccess{e.Name, e.Function, e.binding}
	default:
		panic(fmt.Sprintf("unknown expression (%s)", reflect.TypeOf(expr)))
	}
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
func AsConstantOfExpressions(exprs []Expr, fn func(*big.Int, *big.Int)) *big.Int {
	var val big.Int
	//
	for i, arg := range exprs {
		c := arg.AsConstant()
		if c == nil {
			return nil
		} else if i == 0 {
			// Must clone c
			val.Set(c)
		} else {
			fn(&val, c)
		}
	}
	//
	return &val
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
