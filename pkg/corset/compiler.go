package corset

import "github.com/consensys/go-corset/pkg/hir"

// Compile one or more source files into a schema.
func CompileSourceFiles(files []string) (*hir.Schema, []error) {
	_, errs := ParseSourceFiles(files)
	// Check for parsing errors
	if errs != nil {
		return nil, errs
	}
	// Compile each module into the schema
	panic("TODO")
}

// Compile exactly one source file into a schema.  This is really helper
// function for e.g. the testing environment.
func CompileSourceFile(file string) (*hir.Schema, error) {
	schema, errs := CompileSourceFiles([]string{file})
	// Check for errors
	if errs != nil {
		return nil, errs[0]
	}
	//
	return schema, nil
}
