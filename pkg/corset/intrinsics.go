package corset

import (
	"fmt"
	"math"

	"github.com/consensys/go-corset/pkg/sexp"
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

// IsPure checks whether this pure (which intrinsics always are).
func (p *IntrinsicDefinition) IsPure() bool {
	return true
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

// ReturnType gets the declared return type of this function, or nil if no
// return type was declared.
func (p *IntrinsicDefinition) ReturnType() Type {
	return nil
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

// Apply a given set of arguments to this function binding.
func (p *IntrinsicDefinition) Apply(args []Expr) Expr {
	// First construct the body
	body := p.constructor(uint(len(args)))
	// Then, substitute through.
	mapping := make(map[uint]Expr)
	// Setup the mapping
	for i, e := range args {
		mapping[uint(i)] = e
	}
	// Substitute through
	return body.Substitute(mapping)
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
		binding := &LocalVariableBinding{name, nil, i}
		args[i] = &VariableAccess{nil, name, true, binding}
	}
	//
	return args
}
