package corset

import (
	tr "github.com/consensys/go-corset/pkg/trace"
)

// Binding represents an association between a name, as found in a source file,
// and concrete item (e.g. a column, function, etc).
type Binding interface {
	// Returns the context associated with this binding.
	Context() tr.Context
}

// ColumnBinding represents something bound to a given column.
type ColumnBinding struct {
	// For a column access, this identifies the enclosing context.
	context tr.Context
	// Identifies the variable or column index (as appropriate).
	index uint
}

// Context returns the enclosing context for this column access.
func (p *ColumnBinding) Context() tr.Context {
	return p.context
}

// ColumnID returns the column identifier that this column access refers to.
func (p *ColumnBinding) ColumnID() uint {
	return p.index
}

// ParameterBinding represents something bound to a given column.
type ParameterBinding struct {
	// Identifies the variable or column index (as appropriate).
	index uint
}

// Context for a parameter is always void, as it does not correspond to a column
// in given module.
func (p *ParameterBinding) Context() tr.Context {
	return tr.VoidContext()
}

// FunctionBinding represents the binding of a function application to its
// physical definition.
type FunctionBinding struct {
	// arity determines the number of arguments this function takes.
	arity uint
	// body of the function in question.
	body Expr
}

// Context for a parameter is always void, as it does not correspond to a column
// in given module.
func (p *FunctionBinding) Context() tr.Context {
	return tr.VoidContext()
}

// Apply a given set of arguments to this function binding.
func (p *FunctionBinding) Apply(args []Expr) Expr {
	return p.body.Substitute(args)
}
