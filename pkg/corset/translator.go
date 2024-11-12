package corset

import (
	"github.com/consensys/go-corset/pkg/hir"
	tr "github.com/consensys/go-corset/pkg/trace"
)

// Translate the components of a Corset circuit and add them to the schema.  By
// the time we get to this point, all malformed source files should have been
// rejected already and the translation should go through easily.  Thus, whilst
// syntax errors can be returned here, this should never happen.  The mechanism
// is supported, however, to simplify development of new features, etc.
func translateCircuit(env *Environment, circuit *Circuit) (*hir.Schema, []SyntaxError) {
	schema := hir.EmptySchema()
	errors := []SyntaxError{}
	//
	context := env.Module("")
	// Translate root context
	for _, d := range circuit.Declarations {
		errs := translateDeclaration(d, context, schema)
		errors = append(errors, errs...)
	}
	// Translate submodules
	//
	return schema, errors
}

// Translate a Corset declaration and add it to the schema.  By the time we get
// to this point, all malformed source files should have been rejected already
// and the translation should go through easily.  Thus, whilst syntax errors can
// be returned here, this should never happen.  The mechanism is supported,
// however, to simplify development of new features, etc.
func translateDeclaration(decl Declaration, context tr.Context, schema *hir.Schema) []SyntaxError {
	if d, ok := decl.(*DefColumns); ok {
		translateDefColumns(d, context, schema)
	} else if d, ok := decl.(*DefConstraint); ok {
		translateDefConstraint(d, context, schema)
	}
	// Error handling
	panic("unknown declaration")
}

func translateDefColumns(decl *DefColumns, context tr.Context, schema *hir.Schema) {
	// Add each column to schema
	for _, c := range decl.Columns {
		schema.AddDataColumn(context, c.Name, c.DataType)
	}
}

func translateDefConstraint(decl *DefConstraint, context tr.Context, schema *hir.Schema) {
	panic("TODO")
}
