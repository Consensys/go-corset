package corset

import (
	"fmt"
	"math"

	"github.com/consensys/go-corset/pkg/util"
	"github.com/consensys/go-corset/pkg/util/sexp"
)

// IntrinsicDefinition is a SymbolDefinition for an intrinsic (i.e. built-in)
// operation, such as "+", "-", etc.  These are needed for two reasons: firstly,
// so we can alias them; secondly, so they can be used in reductions.
type IntrinsicDefinition struct {
	// Name of the intrinsic (e.g. "+")
	name string
	// Minimum number of arguments this intrinsic can accept.
	min_arity uint
	// Maximum number of arguments this intrinsic can accept.
	max_arity uint
	// Construct an instance of this intrinsic for a given arity (i.e. number of
	// arguments).
	constructor func(uint) Expr
}

var _ FunctionBinding = &IntrinsicDefinition{}

// Name returns the name of the intrinsic being defined.
func (p *IntrinsicDefinition) Name() string {
	return p.name
}

// Path returns the qualified name (i.e. absolute path) of this symbol.  For
// example, "m1.X" for a column X defined in module m1.
func (p *IntrinsicDefinition) Path() *util.Path {
	path := util.NewAbsolutePath(p.name)
	return &path
}

// IsPure checks whether this pure (which intrinsics always are).
func (p *IntrinsicDefinition) IsPure() bool {
	return true
}

// IsNative checks whether this function binding is native (or not).
func (p *IntrinsicDefinition) IsNative() bool {
	return false
}

// IsFunction identifies whether or not the intrinsic being defined is a
// function.  At this time, all intrinsics are functions.
func (p *IntrinsicDefinition) IsFunction() bool {
	return true
}

// IsFinalised checks whether this binding has been finalised yet or not.
func (p *IntrinsicDefinition) IsFinalised() bool {
	return true
}

// Binding returns the binding associated with this intrinsic.
func (p *IntrinsicDefinition) Binding() Binding {
	return p
}

// Lisp returns a lisp representation of this intrinsic.
func (p *IntrinsicDefinition) Lisp() sexp.SExp {
	panic("unreacahble")
}

// HasArity checks whether this function accepts a given number of arguments (or
// not).
func (p *IntrinsicDefinition) HasArity(arity uint) bool {
	return arity >= p.min_arity && arity <= p.max_arity
}

// Select the best fit signature based on the available parameter types.
// Observe that, for valid arities, this always returns a signature.
// However, that signature may not actually accept the provided parameters
// (in which case, an error should be reported).  Furthermore, if no
// appropriate signature exists then this will return nil.
func (p *IntrinsicDefinition) Select(args []Type) *FunctionSignature {
	// construct the body
	body := p.constructor(uint(len(args)))
	types := make([]Type, len(args))
	//
	for i := 0; i < len(types); i++ {
		types[i] = NewFieldType()
	}
	// Allow return type to be inferred.
	return &FunctionSignature{true, types, nil, body}
}

// Overload (a.k.a specialise) this function binding to incorporate another
// function binding.  This can fail for a few reasons: (1) some bindings
// (e.g. intrinsics) cannot be overloaded; (2) duplicate overloadings are
// not permitted; (3) combinding pure and impure overloadings is also not
// permitted.
func (p *IntrinsicDefinition) Overload(binding *DefunBinding) (FunctionBinding, bool) {
	// Easy case, as intrinsics cannot be overloaded.
	return nil, false
}

// ============================================================================
// Intrinsic Definitions
// ============================================================================

// INTRINSICS identifies all of the built-in functions used within the corset
// language, such as "+", "-", etc.  This is needed for two reasons: firstly, so
// we can alias them; secondly, so they can be used in reductions.
var INTRINSICS []IntrinsicDefinition = []IntrinsicDefinition{
	// Addition
	{"+", 1, math.MaxUint, intrinsicAdd},
	// Subtraction
	{"-", 1, math.MaxUint, intrinsicSub},
	// Multiplication
	{"*", 1, math.MaxUint, intrinsicMul},
}

func intrinsicAdd(arity uint) Expr {
	return &Add{intrinsicNaryBody(arity)}
}

func intrinsicSub(arity uint) Expr {
	return &Sub{intrinsicNaryBody(arity)}
}

func intrinsicMul(arity uint) Expr {
	return &Mul{intrinsicNaryBody(arity)}
}

func intrinsicNaryBody(arity uint) []Expr {
	args := make([]Expr, arity)
	//
	for i := uint(0); i != arity; i++ {
		name := fmt.Sprintf("x%d", i)
		path := util.NewAbsolutePath(name)
		binding := &LocalVariableBinding{name, nil, i}
		args[i] = &VariableAccess{path, true, binding}
	}
	//
	return args
}
