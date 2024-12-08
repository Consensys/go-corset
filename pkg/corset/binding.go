package corset

import (
	"math"

	sc "github.com/consensys/go-corset/pkg/schema"
	tr "github.com/consensys/go-corset/pkg/trace"
)

// BindingId is an identifier is used to distinguish different forms of binding,
// as some forms are known from their use.  Specifically, at the current time,
// only functions are distinguished from other categories (e.g. columns,
// parameters, etc).
type BindingId struct {
	// Name of the binding
	name string
	// Indicates whether function binding or other.
	fn bool
}

// Binding represents an association between a name, as found in a source file,
// and concrete item (e.g. a column, function, etc).
type Binding interface {
	// Determine whether this binding is finalised or not.
	IsFinalised() bool
}

// ============================================================================
// ColumnBinding
// ============================================================================

// ColumnBinding represents something bound to a given column.
type ColumnBinding struct {
	// Column's allocated identifier
	cid uint
	// Column's enclosing module
	module string
	// Determines whether this is a computed column, or not.
	computed bool
	// Determines whether this column must be proven (or not).
	mustProve bool
	// Column's length multiplier
	multiplier uint
	// Column's datatype
	dataType sc.Type
}

// NewColumnBinding constructs a new column binding in a given module.
func NewColumnBinding(module string, computed bool, mustProve bool, multiplier uint, datatype sc.Type) *ColumnBinding {
	return &ColumnBinding{math.MaxUint, module, computed, mustProve, multiplier, datatype}
}

// IsFinalised checks whether this binding has been finalised yet or not.
func (p *ColumnBinding) IsFinalised() bool {
	return p.multiplier != 0
}

// Context returns the of this column.  That is, the module in which this colunm
// was declared and also the length multiplier of that module it requires.
func (p *ColumnBinding) Context() Context {
	return tr.NewContext(p.module, p.multiplier)
}

// AllocateId allocates the column identifier for this column
func (p *ColumnBinding) AllocateId(cid uint) {
	p.cid = cid
}

// ColumnId returns the allocated identifier for this column.  NOTE: this will
// panic if this column has not yet been allocated an identifier.
func (p *ColumnBinding) ColumnId() uint {
	if p.cid == math.MaxUint {
		panic("column id not yet allocated")
	}
	//
	return p.cid
}

// ============================================================================
// ConstantBinding
// ============================================================================

// ConstantBinding represents a constant definition
type ConstantBinding struct {
	// Constant expression which, when evaluated, produces a constant value.
	value Expr
}

// IsFinalised checks whether this binding has been finalised yet or not.
func (p *ConstantBinding) IsFinalised() bool {
	return true
}

// Context returns the of this constant, noting that constants (by definition)
// do not have a context.
func (p *ConstantBinding) Context() Context {
	return tr.VoidContext[string]()
}

// ============================================================================
// ParameterBinding
// ============================================================================

// ParameterBinding represents something bound to a given column.
type ParameterBinding struct {
	// Identifies the variable or column index (as appropriate).
	index uint
}

// ============================================================================
// FunctionBinding
// ============================================================================

// IsFinalised checks whether this binding has been finalised yet or not.
func (p *ParameterBinding) IsFinalised() bool {
	panic("")
}

// FunctionBinding represents the binding of a function application to its
// physical definition.
type FunctionBinding struct {
	// Flag whether or not is pure function
	pure bool
	// Types of parameters
	paramTypes []sc.Type
	// Type of return
	returnType sc.Type
	// body of the function in question.
	body Expr
}

// NewFunctionBinding constructs a new function binding.
func NewFunctionBinding(pure bool, paramTypes []sc.Type, returnType sc.Type, body Expr) FunctionBinding {
	return FunctionBinding{pure, paramTypes, returnType, body}
}

// IsPure checks whether this is a defpurefun or not
func (p *FunctionBinding) IsPure() bool {
	return p.pure
}

// IsFinalised checks whether this binding has been finalised yet or not.
func (p *FunctionBinding) IsFinalised() bool {
	return p.returnType != nil
}

// Arity returns the number of parameters that this function accepts.
func (p *FunctionBinding) Arity() uint {
	return uint(len(p.paramTypes))
}

// Apply a given set of arguments to this function binding.
func (p *FunctionBinding) Apply(args []Expr) Expr {
	return p.body.Substitute(args)
}
