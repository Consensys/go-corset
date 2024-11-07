package corset

import (
	sc "github.com/consensys/go-corset/pkg/schema"
	"github.com/consensys/go-corset/pkg/sexp"
)

// Circuit represents the root of the Abstract Syntax Tree.  This is also
// referred to as the "prelude".  All modules are contained within the root, and
// declarations can also be declared here as well.
type Circuit struct {
	Modules      []Module
	Declarations []Declaration
}

// Module represents a top-level module declaration.  This corresponds to a
// table in the final constraint set.
type Module struct {
	Name         string
	Declarations []Declaration
}

// Node provides common functionality across all elements of the Abstract Syntax
// Tree.  For example, it ensures every element can converted back into Lisp
// form for debugging.  Furthermore, it provides a reference point for
// constructing a suitable source map for reporting syntax errors.
type Node interface {
	// Convert this node into its lisp representation.  This is primarily used
	// for debugging purposes.
	Lisp() sexp.SExp
}

type Declaration interface {
	Node
	Resolve()
}

// ============================================================================
// DefColumns
// ============================================================================

// DefColumns captures a set of one or more columns being declared.
type DefColumns struct {
	Columns []DefColumn
}

func (p *DefColumns) Resolve() {
	panic("got here")
}

// Lisp converts this node into its lisp representation.  This is primarily used
// for debugging purposes.
func (p *DefColumns) Lisp() sexp.SExp {
	panic("got here")
}

// DefColumn packages together those piece relevant to declaring an individual
// column, such its name and type.
type DefColumn struct {
	Name     string
	DataType sc.Type
}

// Lisp converts this node into its lisp representation.  This is primarily used
// for debugging purposes.
func (p *DefColumn) Lisp() sexp.SExp {
	panic("got here")
}

// ============================================================================
// DefConstraint
// ============================================================================

type DefConstraint struct {
}

// ============================================================================
// DefLookup
// ============================================================================

type DefLookup struct {
}

// ============================================================================
// DefPermutation
// ============================================================================

type DefPermutation struct {
}

// ============================================================================
// DefPureFun
// ============================================================================

type DefPureFun struct {
}

// ============================================================================
// Expr
// ============================================================================

type Expr interface {
	Node
	// Resolve resolves this expression in a given scope and constructs a fully
	// resolved HIR expression.
	Resolve()
}
