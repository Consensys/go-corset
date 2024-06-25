package hir

import "fmt"

// ===================================================================
// Environment
// ===================================================================

// Identifies a specific column within the environment.
type columnRef struct {
	module uint
	column string
}

// Environment maps module and column names to their (respective) module and
// column indices.  The environment also keeps trace of which modules / columns
// are declared so we can sanity check them when they are referred to (e.g. in a
// constraint).
type Environment struct {
	// Maps module names to their module indices.
	modules map[string]uint
	// Maps column references to their column indices.
	columns map[columnRef]uint
}

// EmptyEnvironment constructs an empty environment.
func EmptyEnvironment() *Environment {
	modules := make(map[string]uint)
	columns := make(map[columnRef]uint)

	return &Environment{modules, columns}
}

// RegisterModule registers a new module within this environment.  Observe that
// this will panic if the module already exists.
func (p *Environment) RegisterModule(module string) uint {
	if p.HasModule(module) {
		panic(fmt.Sprintf("module %s already exists", module))
	}

	mid := uint(len(p.modules))
	p.modules[module] = mid

	return mid
}

// RegisterColumn registesr a new column within a given module.  Observe that
// this will panic if the column already exists.
func (p *Environment) RegisterColumn(module uint, column string) uint {
	if p.HasColumn(module, column) {
		panic(fmt.Sprintf("column %d:%s already exists", module, column))
	}

	cid := uint(len(p.columns))
	cref := columnRef{module, column}
	p.columns[cref] = cid

	return cid
}

// LookupModule determines the module index for a given named module, or return
// false if no such module exists.
func (p *Environment) LookupModule(module string) (uint, bool) {
	mid, ok := p.modules[module]
	return mid, ok
}

// LookupColumn determines the column index for a given named column in a given
// module, or return false if no such column exists.
func (p *Environment) LookupColumn(module uint, column string) (uint, bool) {
	cref := columnRef{module, column}
	cid, ok := p.columns[cref]

	return cid, ok
}

// HasModule checks whether a given module exists, or not.
func (p *Environment) HasModule(module string) bool {
	_, ok := p.LookupModule(module)
	// Discard column index
	return ok
}

// HasColumn checks whether a given module has a given column, or not.
func (p *Environment) HasColumn(module uint, column string) bool {
	_, ok := p.LookupColumn(module, column)
	// Discard column index
	return ok
}
