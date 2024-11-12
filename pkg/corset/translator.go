package corset

import (
	"github.com/consensys/go-corset/pkg/hir"
)

// Translate the components of a Corset circuit and add them to the schema.  By
// the time we get to this point, all malformed source files should have been
// rejected already and the translation should go through easily.  Thus, whilst
// syntax errors can be returned here, this should never happen.  The mechanism
// is supported, however, to simplify development of new features, etc.
func translateCircuit(circuit *Circuit, schema *hir.Schema) []SyntaxError {
	panic("todo")
}

// Translate a Corset declaration and add it to the schema.  By the time we get
// to this point, all malformed source files should have been rejected already
// and the translation should go through easily.  Thus, whilst syntax errors can
// be returned here, this should never happen.  The mechanism is supported,
// however, to simplify development of new features, etc.
func translateDeclaration(decl Declaration, schema *hir.Schema) []SyntaxError {
	if d, ok := decl.(*DefColumns); ok {
		translateDefColumns(d, schema)
	} else if d, ok := decl.(*DefConstraint); ok {
		translateDefConstraint(d, schema)
	}
	// Error handling
	panic("unknown declaration")
}

func translateDefColumns(decl *DefColumns, schema *hir.Schema) {
	panic("TODO")
}

func translateDefConstraint(decl *DefConstraint, schema *hir.Schema) {
	panic("TODO")
}
