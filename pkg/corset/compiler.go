package corset

import (
	_ "embed"

	"github.com/consensys/go-corset/pkg/binfile"
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
func CompileSourceFiles(stdlib bool, debug bool, srcfiles []*sexp.SourceFile) (*binfile.BinaryFile, []SyntaxError) {
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
func CompileSourceFile(stdlib bool, debug bool, srcfile *sexp.SourceFile) (*binfile.BinaryFile, []SyntaxError) {
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
func (p *Compiler) Compile() (*binfile.BinaryFile, []SyntaxError) {
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
	// Translate everything and add it to the schema.
	schema, errs := compiler.TranslateCircuit(environment, p.srcmap, &p.circuit)
	// Sanity check for errors
	if len(errs) > 0 {
		return nil, errs
	}
	// Construct source map
	source_map := constructSourceMap(scope, environment)
	// Extract key attributes (for debugging purposes)
	attributes := []binfile.Attribute{source_map}
	// Construct binary file
	return binfile.NewBinaryFile(nil, attributes, schema), errs
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

func constructSourceMap(scope *compiler.ModuleScope, env compiler.GlobalEnvironment) *SourceMap {
	return &SourceMap{constructSourceModule(scope, env)}
}

func constructSourceModule(scope *compiler.ModuleScope, env compiler.GlobalEnvironment) SourceModule {
	var (
		columns    []SourceColumn
		submodules []SourceModule
	)
	//
	for _, col := range scope.DestructuredColumns() {
		// Determine register allocated to this (destructured) column.
		regId := env.RegisterOf(&col.Name)
		// Determine (unqualified) column name
		name := col.Name.Tail()
		//
		display := constructDisplayModifier(col.Display)
		// Translate register source into source column
		srcCol := SourceColumn{name, col.Multiplier, col.DataType, col.MustProve, col.Computed, display, regId}
		columns = append(columns, srcCol)
	}
	//
	for _, child := range scope.Children() {
		submodules = append(submodules, constructSourceModule(child, env))
	}
	//
	return SourceModule{
		Name:       scope.Name(),
		Synthetic:  false,
		Virtual:    scope.Virtual(),
		Selector:   compileSelector(env, scope.Selector()),
		Submodules: submodules,
		Columns:    columns,
	}
}

func constructDisplayModifier(modifier string) uint {
	switch modifier {
	case "hex":
		return DISPLAY_HEX
	case "dec":
		return DISPLAY_DEC
	case "bytes":
		return DISPLAY_BYTES
	case "opcode":
		return DISPLAY_CUSTOM
	}
	// unknown, so default to hex
	return DISPLAY_HEX
}

// This is really broken.  The problem is that we need to translate the selector
// expression within the translator.  But, setting that all up is not
// straightforward.  This should be done in the future!
func compileSelector(env compiler.Environment, selector ast.Expr) *hir.UnitExpr {
	if selector == nil {
		return nil
	}
	//
	if e, ok := selector.(*ast.VariableAccess); ok {
		if binding, ok := e.Binding().(*ast.ColumnBinding); ok {
			// Lookup column binding
			register_id := env.RegisterOf(binding.AbsolutePath())
			// Done
			expr := &hir.ColumnAccess{Column: register_id, Shift: 0}
			//
			return &hir.UnitExpr{Expr: expr}
		}
	}
	// FIXME: #630
	panic("unsupported selector")
}
