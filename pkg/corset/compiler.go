package corset

import (
	"fmt"

	"github.com/consensys/go-corset/pkg/hir"
	"github.com/consensys/go-corset/pkg/sexp"
)

// CompileSourceFiles compiles one or more source files into a schema.
func CompileSourceFiles(srcfiles []*sexp.SourceFile) (*hir.Schema, []error) {
	circuit, srcmap, errs := ParseSourceFiles(srcfiles)
	// Check for parsing errors
	if errs != nil {
		return nil, errs
	}
	// Compile each module into the schema
	return NewCompiler(circuit, srcmap).Compile()
}

// CompileSourceFile compiles exactly one source file into a schema.  This is
// really helper function for e.g. the testing environment.
func CompileSourceFile(srcfile *sexp.SourceFile) (*hir.Schema, error) {
	schema, errs := CompileSourceFiles([]*sexp.SourceFile{srcfile})
	// Check for errors
	if errs != nil {
		return nil, errs[0]
	}
	//
	return schema, nil
}

// Compiler packages up everything needed to compiler a given set of
// module definitions down into an HIR schema.  Observe that the compiler may
// fail if the modules definitions are mal-formed in some way (e.g. fail type
// checking).
type Compiler struct {
	// A high-level definition of a Corset circuit.
	circuit Circuit
	// Source maps nodes in the circuit back to the spans in their original
	// source files.
	srcmap *sexp.SourceMaps[Node]
	// This schema is being constructed by the compiler from the circuit.
	schema *hir.Schema
}

// NewCompiler constructs a new compiler for a given set of modules.
func NewCompiler(circuit Circuit, srcmaps *sexp.SourceMaps[Node]) *Compiler {
	return &Compiler{circuit, srcmaps, hir.EmptySchema()}
}

// Compile is the top-level function for the corset compiler which actually
// compiles the given modules down into a schema.  This can fail in a variety of
// ways if the given modules are malformed in some way.  For example, if some
// expression refers to a non-existent module or column, or is not well-typed,
// etc.
func (p *Compiler) Compile() (*hir.Schema, []error) {
	fmt.Printf("MODULES: %v\n", p.circuit.Modules)
	panic("TODO")
}
