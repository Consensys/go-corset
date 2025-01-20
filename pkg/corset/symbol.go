package corset

import (
	"github.com/consensys/go-corset/pkg/util"
	"github.com/consensys/go-corset/pkg/util/sexp"
)

// Symbol represents a variable or function access within a declaration.
// Initially, such the proper interpretation of such accesses is unclear and it
// is only later when we can distinguish them (e.g. whether its a column access,
// a constant access, etc).
type Symbol interface {
	Node
	// Path returns the given path of this symbol.
	Path() *util.Path
	// Indicates whether or not this is a function.
	IsFunction() bool
	// Checks whether this symbol has been resolved already, or not.
	IsResolved() bool
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

// TypedSymbol is an extended form of symbol which contains additional
// information about a given column access.
type TypedSymbol interface {
	Symbol
	// Type returns the type associated with this symbol.  If the type cannot be
	// determined, then nil is returned.
	Type() Type
}

// SymbolDefinition represents a declaration (or part thereof) which defines a
// particular symbol.  For example, "defcolumns" will define one or more symbols
// representing columns, etc.
type SymbolDefinition interface {
	Node
	// Name returns the (unqualified) name of this symbol.  For example, "X" for
	// a column X defined in a module m1.
	Name() string
	// Path returns the qualified name (i.e. absolute path) of this symbol.  For
	// example, "m1.X" for a column X defined in module m1.
	Path() *util.Path
	// Indicates whether or not this is a function definition.
	IsFunction() bool
	// Allocated binding for the symbol which may or may not be finalised.
	Binding() Binding
}

// FunctionName represents a name used in a position where it can only be
// resolved as a function.
type FunctionName = Name[*DefunBinding]

// NewFunctionName construct a new column name which is (initially) unresolved.
func NewFunctionName(path util.Path, binding *DefunBinding) *FunctionName {
	return &FunctionName{path, true, binding, true}
}

// PerspectiveName represents a name used in a position where it can only be
// resolved as a perspective.
type PerspectiveName = Name[*PerspectiveBinding]

// NewPerspectiveName construct a new column name which is (initially) unresolved.
func NewPerspectiveName(path util.Path, binding *PerspectiveBinding) *PerspectiveName {
	return &PerspectiveName{path, true, binding, true}
}

// Name represents a name within some syntactic item.  Essentially this wraps a
// string and provides a mechanism for it to be associated with source line
// information.
type Name[T Binding] struct {
	// Name of symbol
	path util.Path
	// Indicates whether represents function or something else.
	function bool
	// Binding constructed for symbol.
	binding T
	// Indicates whether resolved.
	resolved bool
}

// NewName construct a new name which is (initially) unresolved.
func NewName[T Binding](path util.Path, function bool) *Name[T] {
	// Default value for type T
	var empty T
	// Construct the name
	return &Name[T]{path, function, empty, false}
}

// Name returns the (unqualified) name of this symbol.  For example, "X" for
// a column X defined in a module m1.
func (e *Name[T]) Name() string {
	return e.path.Tail()
}

// Path returns the qualified name (i.e. absolute path) of this symbol.  For
// example, "m1.X" for a column X defined in module m1.
func (e *Name[T]) Path() *util.Path {
	return &e.path
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
	return sexp.NewSymbol(e.path.String())
}
