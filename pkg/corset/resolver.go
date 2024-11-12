package corset

// Resolve all symbols declared and used within a circuit, producing an
// environment which can subsequently be used to look up the relevant module or
// column identifiers.  This process can fail, of course, it a symbol (e.g. a
// column) is referred to which doesn't exist.  Likewise, if two modules or
// columns with identical names are declared in the same scope, etc.
func resolveCircuit(circuit *Circuit) (*Environment, []SyntaxError) {
	env := EmptyEnvironment()
	// Allocate declared modules
	merrs := resolveModules(env, circuit)
	// Allocate declared input columns
	cerrs := resolveInputColumns(env, circuit)
	// Allocate declared assignments
	// Check expressions
	// Done
	return env, append(merrs, cerrs...)
}

// Process all module declarations, and allocating them into the environment.
// If any duplicates are found, one or more errors will be reported.  Note: it
// is important that this traverses the modules in an identical order to the
// translator.  This is to ensure that the relevant module identifiers line up.
func resolveModules(env *Environment, circuit *Circuit) []SyntaxError {
	panic("todo")
}

// Process all input (column) declarations.  These must be allocated before
// assignemts, since the hir.Schema separates these out.  Again, if any
// duplicates are found then one or more errors will be reported.
func resolveInputColumns(env *Environment, circuit *Circuit) []SyntaxError {
	panic("todo")
}
