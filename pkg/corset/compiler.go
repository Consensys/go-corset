package corset

import (
	"github.com/consensys/go-corset/pkg/hir"
	"github.com/consensys/go-corset/pkg/sexp"
)

// SyntaxError defines the kind of errors that can be reported by this compiler.
// Syntax errors are always associated with some line in one of the original
// source files.  For simplicity, we reuse existing notion of syntax error from
// the S-Expression library.
type SyntaxError = sexp.SyntaxError

// CompileSourceFiles compiles one or more source files into a schema.  This
// process can fail if the source files are mal-formed, or contain syntax errors
// or other forms of error (e.g. type errors).
func CompileSourceFiles(srcfiles []*sexp.SourceFile) (*hir.Schema, []SyntaxError) {
	circuit, srcmap, errs := ParseSourceFiles(srcfiles)
	// Check for parsing errors
	if errs != nil {
		return nil, errs
	}
	// Compile each module into the schema
	return NewCompiler(circuit, srcmap).Compile()
}

// CompileSourceFile compiles exactly one source file into a schema.  This is
// really helper function for e.g. the testing environment.   This process can
// fail if the source file is mal-formed, or contains syntax errors or other
// forms of error (e.g. type errors).
func CompileSourceFile(srcfile *sexp.SourceFile) (*hir.Schema, []SyntaxError) {
	schema, errs := CompileSourceFiles([]*sexp.SourceFile{srcfile})
	// Check for errors
	if errs != nil {
		return nil, errs
	}
	//
	return schema, nil
}

// Compiler packages up everything needed to compile a given set of module
// definitions down into an HIR schema.  Observe that the compiler may fail if
// the modules definitions are malformed in some way (e.g. fail type checking).
type Compiler struct {
	// A high-level definition of a Corset circuit.
	circuit Circuit
	// Source maps nodes in the circuit back to the spans in their original
	// source files.  This is needed when reporting syntax errors to generate
	// highlights of the relevant source line(s) in question.
	srcmap *sexp.SourceMaps[Node]
}

// NewCompiler constructs a new compiler for a given set of modules.
func NewCompiler(circuit Circuit, srcmaps *sexp.SourceMaps[Node]) *Compiler {
	return &Compiler{circuit, srcmaps}
}

// Compile is the top-level function for the corset compiler which actually
// compiles the given modules down into a schema.  This can fail in a variety of
// ways if the given modules are malformed in some way.  For example, if some
// expression refers to a non-existent module or column, or is not well-typed,
// etc.
func (p *Compiler) Compile() (*hir.Schema, []SyntaxError) {
	schema := hir.EmptySchema()
	// Allocate columns?
	//
	// Resolve variables (via nested scopes)
	// Check constraint contexts (e.g. for constraints, lookups, etc)
	// Type check constraints
	// Finally, translate everything and add it to the schema.
	errors := translateCircuit(&p.circuit, schema)
	// Done
	return schema, errors
}
