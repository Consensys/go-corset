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
	// Returns the context associated with this binding.
	IsBinding()
}

// ColumnBinding represents something bound to a given column.
type ColumnBinding struct {
	// Column's allocated identifier
	cid uint
	// Column's enclosing module
	module string
	// Determines whether this is a computed column, or not.
	computed bool
	// Column's length multiplier
	multiplier uint
	// Column's datatype
	datatype sc.Type
}

// NewColumnBinding constructs a new column binding in a given module.
func NewColumnBinding(module string, computed bool, multiplier uint, datatype sc.Type) *ColumnBinding {
	return &ColumnBinding{math.MaxUint, module, computed, multiplier, datatype}
}

// IsBinding ensures this is an instance of Binding.
func (p *ColumnBinding) IsBinding() {}

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

// ParameterBinding represents something bound to a given column.
type ParameterBinding struct {
	// Identifies the variable or column index (as appropriate).
	index uint
}

// IsBinding ensures this is an instance of Binding.
func (p *ParameterBinding) IsBinding() {}

// FunctionBinding represents the binding of a function application to its
// physical definition.
type FunctionBinding struct {
	// arity determines the number of arguments this function takes.
	arity uint
	// body of the function in question.
	body Expr
}

// IsBinding ensures this is an instance of Binding.
func (p *FunctionBinding) IsBinding() {}

// Apply a given set of arguments to this function binding.
func (p *FunctionBinding) Apply(args []Expr) Expr {
	return p.body.Substitute(args)
}
