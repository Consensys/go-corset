package hir

import (
	"fmt"

	"github.com/consensys/go-corset/pkg/schema"
	sc "github.com/consensys/go-corset/pkg/schema"
)

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
	// Schema being constructed
	schema *Schema
}

// EmptyEnvironment constructs an empty environment.
func EmptyEnvironment() *Environment {
	modules := make(map[string]uint)
	columns := make(map[columnRef]uint)
	schema := EmptySchema()
	//
	return &Environment{modules, columns, schema}
}

// RegisterModule registers a new module within this environment.  Observe that
// this will panic if the module already exists.
func (p *Environment) RegisterModule(module string) uint {
	if p.HasModule(module) {
		panic(fmt.Sprintf("module %s already exists", module))
	}
	// Update schema
	mid := p.schema.AddModule(module)
	// Update cache
	p.modules[module] = mid
	// Done
	return mid
}

// AddDataColumn registers a new column within a given module.  Observe that
// this will panic if the column already exists.
func (p *Environment) AddDataColumn(module uint, column string, datatype sc.Type) uint {
	if p.HasColumn(module, column) {
		panic(fmt.Sprintf("column %d:%s already exists", module, column))
	}
	// Update schema
	p.schema.AddDataColumn(module, column, datatype)
	// Update cache
	cid := uint(len(p.columns))
	cref := columnRef{module, column}
	p.columns[cref] = cid
	// Done
	return cid
}

// AddAssignment appends a new assignment (i.e. set of computed columns) to be
// used during trace expansion for this schema.  Computed columns are introduced
// by the process of lowering from HIR / MIR to AIR.
func (p *Environment) AddAssignment(decl schema.Assignment) {
	// Update schema
	index := p.schema.AddAssignment(decl)
	// Update cache
	for i := decl.Columns(); i.HasNext(); {
		ith := i.Next()
		cref := columnRef{ith.Module(), ith.Name()}
		p.columns[cref] = index
		index++
	}
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
