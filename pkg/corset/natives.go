package corset

import (
	"math"

	"github.com/consensys/go-corset/pkg/corset/ast"
	"github.com/consensys/go-corset/pkg/util"
	"github.com/consensys/go-corset/pkg/util/sexp"
)

// NativeColumn provides information about a column acting as a parameter or
// return in a given native function.
type NativeColumn struct {
	// type of assigned column
	datatype ast.Type
	// multiplier for assigned column
	multiplier uint
}

// NativeDefinition describes a native function, such as specifying how many
// arguments it requires, etc.
type NativeDefinition struct {
	// Name of the intrinsic (e.g. "+")
	name string
	// Minimum number of arguments this native can accept.
	min_arity uint
	// Maximum number of arguments this native can accept.
	max_arity uint
	// Responsible for doing whatever the function does.
	constructor func([]NativeColumn) []NativeColumn
}

var _ ast.FunctionBinding = &NativeDefinition{}

// Name returns the name of the intrinsic being defined.
func (p *NativeDefinition) Name() string {
	return p.name
}

// Path returns the qualified name (i.e. absolute path) of this symbol.  For
// example, "m1.X" for a column X defined in module m1.
func (p *NativeDefinition) Path() *util.Path {
	path := util.NewAbsolutePath(p.name)
	return &path
}

// IsPure checks whether this pure (which intrinsics always are).
func (p *NativeDefinition) IsPure() bool {
	return false
}

// IsNative checks whether this function binding is native (or not).
func (p *NativeDefinition) IsNative() bool {
	return true
}

// IsFunction identifies whether or not the intrinsic being defined is a
// function.  At this time, all intrinsics are functions.
func (p *NativeDefinition) IsFunction() bool {
	return true
}

// IsFinalised checks whether this binding has been finalised yet or not.
func (p *NativeDefinition) IsFinalised() bool {
	return true
}

// Binding returns the binding associated with this intrinsic.
func (p *NativeDefinition) Binding() ast.Binding {
	return p
}

// Lisp returns a lisp representation of this intrinsic.
func (p *NativeDefinition) Lisp() sexp.SExp {
	panic("unreacahble")
}

// HasArity checks whether this function accepts a given number of arguments (or
// not).
func (p *NativeDefinition) HasArity(arity uint) bool {
	return arity >= p.min_arity && arity <= p.max_arity
}

// Select the best fit signature based on the available parameter types.
// Observe that, for valid arities, this always returns a signature.
// However, that signature may not actually accept the provided parameters
// (in which case, an error should be reported).  Furthermore, if no
// appropriate signature exists then this will return nil.
func (p *NativeDefinition) Select(args []ast.Type) *ast.FunctionSignature {
	// This is safe because natives can only (currently) be used in very
	// specific situations.
	return nil
}

// Apply returns the output columns given a set of input columns.
func (p *NativeDefinition) Apply(args []NativeColumn) []NativeColumn {
	return p.constructor(args)
}

// Overload (a.k.a specialise) this function binding to incorporate another
// function binding.  This can fail for a few reasons: (1) some bindings
// (e.g. intrinsics) cannot be overloaded; (2) duplicate overloadings are
// not permitted; (3) combinding pure and impure overloadings is also not
// permitted.
func (p *NativeDefinition) Overload(binding *ast.DefunBinding) (ast.FunctionBinding, bool) {
	// Easy case, as natives cannot be overloaded.
	return nil, false
}

// ============================================================================
// Native Definitions
// ============================================================================

// NATIVES identifies all built-in native computations which can be used in
// defcomputed assignments.
var NATIVES []NativeDefinition = []NativeDefinition{
	// Simple identity function.
	{"id", 1, 1, nativeId},
	// Filter based on second argument
	{"filter", 2, 2, nativeFilter},
	// Identify changes of a column within a given region (in forwards direction).
	{"fwd-changes-within", 2, math.MaxUint, nativeChangeWithin},
	// Identify rows which don't change within a given region (in forwards direction).
	{"fwd-unchanged-within", 2, math.MaxUint, nativeChangeWithin},
	// Identify changes of a column within a given region (in backwards direction).
	{"bwd-changes-within", 2, math.MaxUint, nativeChangeWithin},
	// Flood fill (forwards) within a given region
	{"fwd-fill-within", 3, 3, nativeFillWithin},
	// Flood fill (backwards) within a given region
	{"bwd-fill-within", 3, 3, nativeFillWithin},
}

func nativeId(inputs []NativeColumn) []NativeColumn {
	if len(inputs) != 1 {
		panic("unreachable")
	}

	return inputs
}

func nativeFilter(inputs []NativeColumn) []NativeColumn {
	if len(inputs) != 2 {
		panic("unreachable")
	}
	//
	return []NativeColumn{inputs[0]}
}

func nativeChangeWithin(inputs []NativeColumn) []NativeColumn {
	if len(inputs) <= 1 {
		panic("unreachable")
	}
	//
	result := NativeColumn{ast.NewUintType(1), inputs[0].multiplier}
	//
	return []NativeColumn{result}
}

func nativeFillWithin(inputs []NativeColumn) []NativeColumn {
	if len(inputs) <= 2 {
		panic("unreachable")
	}
	//
	return []NativeColumn{inputs[0]}
}
