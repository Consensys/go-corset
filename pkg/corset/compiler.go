package corset

import (
	_ "embed"

	"github.com/consensys/go-corset/pkg/corset/ast"
	"github.com/consensys/go-corset/pkg/corset/compiler"
	"github.com/consensys/go-corset/pkg/hir"
	"github.com/consensys/go-corset/pkg/util/sexp"
)

// STDLIB is an import of the standard library.
//
//go:embed stdlib.lisp
var STDLIB []byte

// SyntaxError defines the kind of errors that can be reported by this compiler.
// Syntax errors are always associated with some line in one of the original
// source files.  For simplicity, we reuse existing notion of syntax error from
// the S-Expression library.
type SyntaxError = sexp.SyntaxError

// CompileSourceFiles compiles one or more source files into a schema.  This
// process can fail if the source files are mal-formed, or contain syntax errors
// or other forms of error (e.g. type errors).
func CompileSourceFiles(stdlib bool, debug bool, srcfiles []*sexp.SourceFile) (*hir.Schema, []SyntaxError) {
	// Include the standard library (if requested)
	srcfiles = includeStdlib(stdlib, srcfiles)
	// Parse all source files (inc stdblib if applicable).
	circuit, srcmap, errs := compiler.ParseSourceFiles(srcfiles)
	// Check for parsing errors
	if errs != nil {
		return nil, errs
	}
	// Compile each module into the schema
	return NewCompiler(circuit, srcmap).SetDebug(debug).Compile()
}

// CompileSourceFile compiles exactly one source file into a schema.  This is
// really helper function for e.g. the testing environment.   This process can
// fail if the source file is mal-formed, or contains syntax errors or other
// forms of error (e.g. type errors).
func CompileSourceFile(stdlib bool, debug bool, srcfile *sexp.SourceFile) (*hir.Schema, []SyntaxError) {
	schema, errs := CompileSourceFiles(stdlib, debug, []*sexp.SourceFile{srcfile})
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
	// The register allocation algorithm to be used by this compiler.
	allocator func(compiler.RegisterAllocation)
	// A high-level definition of a Corset circuit.
	circuit ast.Circuit
	// Determines whether debug
	debug bool
	// Source maps nodes in the circuit back to the spans in their original
	// source files.  This is needed when reporting syntax errors to generate
	// highlights of the relevant source line(s) in question.
	srcmap *sexp.SourceMaps[ast.Node]
}

// NewCompiler constructs a new compiler for a given set of modules.
func NewCompiler(circuit ast.Circuit, srcmaps *sexp.SourceMaps[ast.Node]) *Compiler {
	return &Compiler{compiler.DEFAULT_ALLOCATOR, circuit, false, srcmaps}
}

// SetDebug enables or disables debug mode.  In debug mode, debug constraints
// will be compiled in.
func (p *Compiler) SetDebug(flag bool) *Compiler {
	p.debug = flag
	return p
}

// SetAllocator overides the default register allocator.
func (p *Compiler) SetAllocator(allocator func(compiler.RegisterAllocation)) *Compiler {
	p.allocator = allocator
	return p
}

// Compile is the top-level function for the corset compiler which actually
// compiles the given modules down into a schema.  This can fail in a variety of
// ways if the given modules are malformed in some way.  For example, if some
// expression refers to a non-existent module or column, or is not well-typed,
// etc.
func (p *Compiler) Compile() (*hir.Schema, []SyntaxError) {
	// Resolve variables (via nested scopes)
	scope, res_errs := compiler.ResolveCircuit(p.srcmap, &p.circuit)
	// Type check circuit.
	type_errs := compiler.TypeCheckCircuit(p.srcmap, &p.circuit)
	// Don't proceed if errors at this point.
	if len(res_errs) > 0 || len(type_errs) > 0 {
		return nil, append(res_errs, type_errs...)
	}
	// Preprocess circuit to remove invocations, reductions, etc.
	if errs := compiler.PreprocessCircuit(p.debug, p.srcmap, &p.circuit); len(errs) > 0 {
		return nil, errs
	}
	// Convert global scope into an environment by allocating all columns.
	environment := compiler.NewGlobalEnvironment(scope, p.allocator)
	// Finally, translate everything and add it to the schema.
	return compiler.TranslateCircuit(environment, p.srcmap, &p.circuit)
}

func includeStdlib(stdlib bool, srcfiles []*sexp.SourceFile) []*sexp.SourceFile {
	if stdlib {
		// Include stdlib file
		srcfile := sexp.NewSourceFile("stdlib.lisp", STDLIB)
		// Append to srcfile list
		srcfiles = append(srcfiles, srcfile)
	}
	// Not included
	return srcfiles
}
