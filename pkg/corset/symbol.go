package corset

import (
	"fmt"

	"github.com/consensys/go-corset/pkg/sexp"
)

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
	// the definition of the symbol to which this refers.  Observe that
	// resolution can fail if we cannot bind the symbol to the given binding
	// (e.g. a function binding was provided, but we're expecting a column
	// binding).
	Resolve(Binding) bool
}

// QualifiedName returns the qualified name of a given symbol
func QualifiedName(symbol Symbol) string {
	if symbol.IsQualified() {
		return fmt.Sprintf("%s.%s", symbol.Module(), symbol.Name())
	}
	//
	return symbol.Name()
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

// ColumnName represents a name used in a position where it can only be resolved
// against a column.
type ColumnName = Name[*ColumnBinding]

// NewColumnName construct a new column name which is (initially) unresolved.
func NewColumnName(name string) *ColumnName {
	return &ColumnName{name, false, nil, false}
}

// Name represents a name within some syntactic item.  Essentially this wraps a
// string and provides a mechanism for it to be associated with source line
// information.
type Name[T Binding] struct {
	// Name of symbol
	name string
	// Indicates whether represents function or something else.
	function bool
	// Binding constructed for symbol.
	binding T
	// Indicates whether resolved.
	resolved bool
}

// NewName construct a new name which is (initially) unresolved.
func NewName[T Binding](name string, function bool) *Name[T] {
	// Default value for type T
	var empty T
	// Construct the name
	return &Name[T]{name, function, empty, false}
}

// IsQualified determines whether this symbol is qualfied or not (i.e. has an
// explicit module specifier).  Column names are never qualified.
func (e *Name[T]) IsQualified() bool {
	return false
}

// IsFunction indicates whether or not this symbol refers to a function (which
// of course it never does).
func (e *Name[T]) IsFunction() bool {
	return e.function
}

// IsResolved checks whether this symbol has been resolved already, or not.
func (e *Name[T]) IsResolved() bool {
	return e.resolved
}

// Module returns the optional module qualification.  This always panics because
// column name's are never qualified.
func (e *Name[T]) Module() string {
	panic("undefined")
}

// Name returns the (unqualified) name of the column to which this symbol
// refers.
func (e *Name[T]) Name() string {
	return e.name
}

// Binding gets binding associated with this interface.  This will panic if this
// symbol is not yet resolved.
func (e *Name[T]) Binding() Binding {
	if !e.resolved {
		panic("name not yet resolved")
	}
	//
	return e.binding
}

// Resolve this symbol by associating it with the binding associated with
// the definition of the symbol to which this refers.
func (e *Name[T]) Resolve(binding Binding) bool {
	var ok bool
	//
	if e.resolved {
		panic("name already resolved")
	}
	// Attempt to assign binding.
	e.binding, ok = binding.(T)
	e.resolved = ok
	//
	return ok
}

// Lisp converts this node into its lisp representation.  This is primarily used
// for debugging purposes.
func (e *Name[T]) Lisp() sexp.SExp {
	return sexp.NewSymbol(e.name)
}
