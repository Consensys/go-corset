package corset

import (
	"github.com/consensys/go-corset/pkg/hir"
	sc "github.com/consensys/go-corset/pkg/schema"
)

type Module struct {
	Name         string
	Declarations []Declaration
}

type Declaration interface {
	LowerToHir(schema *hir.Schema)
}

// DefColumns captures a set of one or more columns being declared.
type DefColumns struct {
	Columns []DefColumn
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
