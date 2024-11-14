package corset

import (
	"github.com/consensys/go-corset/pkg/sexp"
	tr "github.com/consensys/go-corset/pkg/trace"
)

// ResolveCircuit resolves all symbols declared and used within a circuit,
// producing an environment which can subsequently be used to look up the
// relevant module or column identifiers.  This process can fail, of course, it
// a symbol (e.g. a column) is referred to which doesn't exist.  Likewise, if
// two modules or columns with identical names are declared in the same scope,
// etc.
func ResolveCircuit(srcmap *sexp.SourceMaps[Node], circuit *Circuit) (*Environment, []SyntaxError) {
	r := resolver{EmptyEnvironment(), srcmap}
	// Allocate declared modules
	merrs := r.resolveModules(circuit)
	// Allocate declared input columns
	cerrs := r.resolveInputColumns(circuit)
	// Allocate declared assignments
	// Check expressions
	// Done
	return r.env, append(merrs, cerrs...)
}

// Resolver packages up information necessary for resolving a circuit and
// checking that everything makes sense.
type resolver struct {
	// Environment determines module and column indices, as needed for
	// translating the various constructs found in a circuit.
	env *Environment
	// Source maps nodes in the circuit back to the spans in their original
	// source files.  This is needed when reporting syntax errors to generate
	// highlights of the relevant source line(s) in question.
	srcmap *sexp.SourceMaps[Node]
}

// Process all module declarations, and allocating them into the environment.
// If any duplicates are found, one or more errors will be reported.  Note: it
// is important that this traverses the modules in an identical order to the
// translator.  This is to ensure that the relevant module identifiers line up.
func (r *resolver) resolveModules(circuit *Circuit) []SyntaxError {
	// Register the root module (which should always exist)
	r.env.RegisterModule("")
	//
	for _, m := range circuit.Modules {
		r.env.RegisterModule(m.Name)
	}
	// Done
	return nil
}

// Process all input (column) declarations.  These must be allocated before
// assignemts, since the hir.Schema separates these out.  Again, if any
// duplicates are found then one or more errors will be reported.
func (r *resolver) resolveInputColumns(circuit *Circuit) []SyntaxError {
	errs := r.resolveInputColumnsInModule(r.env.Module(""), circuit.Declarations)
	//
	for _, m := range circuit.Modules {
		// The module must exist given after resolveModules.
		ctx := r.env.Module(m.Name)
		// Process all declarations in the module
		merrs := r.resolveInputColumnsInModule(ctx, m.Declarations)
		// Package up all errors
		errs = append(errs, merrs...)
	}
	//
	return errs
}

func (r *resolver) resolveInputColumnsInModule(module uint, decls []Declaration) []SyntaxError {
	var errors []SyntaxError
	//
	for _, d := range decls {
		// Look for defcolumns decalarations only
		if dcols, ok := d.(*DefColumns); ok {
			// Found one.
			for _, col := range dcols.Columns {
				if r.env.HasColumn(module, col.Name) {
					errors = append(errors, *r.srcmap.SyntaxError(col, "duplicate declaration"))
				} else {
					context := tr.NewContext(module, col.LengthMultiplier)
					r.env.RegisterColumn(context, col.Name, col.DataType)
				}
			}
		}
	}
	//
	return errors
}
