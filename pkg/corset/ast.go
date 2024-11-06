package corset

import (
	sc "github.com/consensys/go-corset/pkg/schema"
)

type Module struct {
	Name         string
	Declarations []Declaration
}

type Declaration interface {
	Resolve()
}

// DefColumns captures a set of one or more columns being declared.
type DefColumns struct {
	Columns []DefColumn
}

func (p *DefColumns) Resolve() {
	panic("got here")
}

// DefColumn packages together those piece relevant to declaring an individual
// column, such its name and type.
type DefColumn struct {
	Name     string
	DataType sc.Type
}

type DefConstraint struct {
}

type DefLookup struct {
}

type DefPermutation struct {
}

type DefPureFun struct {
}

type Expr interface {
	// Resolve resolves this expression in a given scope and constructs a fully
	// resolved HIR expression.
	Resolve()
}
