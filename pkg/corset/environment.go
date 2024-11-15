package corset

import (
	"fmt"

	"github.com/consensys/go-corset/pkg/schema"
	"github.com/consensys/go-corset/pkg/trace"
)

// ===================================================================
// Environment
// ===================================================================

// Identifies a specific column within the environment.
type colRef struct {
	module uint
	column string
}

// Packages up information about a declared column (either input or assignment).
type colInfo struct {
	// Column index
	cid uint
	// Length multiplier
	multiplier uint
	// Datatype
	datatype schema.Type
}

// Environment maps module and column names to their (respective) module and
// column indices.  The environment separates input columns from assignment
// columns because they are disjoint in the schema being constructed (i.e. input
// columns always have a lower index than assignments).
type Environment struct {
	// Maps module names to their module indices.
	modules map[string]uint
	// Maps input columns to their column indices.
	columns map[colRef]colInfo
}

// EmptyEnvironment constructs an empty environment.
func EmptyEnvironment() *Environment {
	modules := make(map[string]uint)
	columns := make(map[colRef]colInfo)
	//
	return &Environment{modules, columns}
}

// RegisterModule registers a new module within this environment.  Observe that
// this will panic if the module already exists.  Furthermore, the module
// identifier is always determined as the next available identifier.
func (p *Environment) RegisterModule(module string) trace.Context {
	if p.HasModule(module) {
		panic(fmt.Sprintf("module %s already exists", module))
	}
	// Update schema
	mid := uint(len(p.modules))
	// Update cache
	p.modules[module] = mid
	// Done
	return trace.NewContext(mid, 1)
}

// RegisterColumn registers a new column (input or assignment) within a given
// module.  Observe that this will panic if the column already exists.
// Furthermore, the column identifier is always determined as the next available
// identifier.  Hence, care must be taken when declaring columns to ensure they
// are allocated in the right order.
func (p *Environment) RegisterColumn(context trace.Context, column string, datatype schema.Type) uint {
	if p.HasColumn(context.Module(), column) {
		panic(fmt.Sprintf("column %d:%s already exists", context.Module(), column))
	}
	// Update cache
	cid := uint(len(p.columns))
	cref := colRef{context.Module(), column}
	p.columns[cref] = colInfo{cid, context.LengthMultiplier(), datatype}
	// Done
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
	cref := colRef{module, column}
	cinfo, ok := p.columns[cref]

	return cinfo.cid, ok
}

// Module determines the module index for a given module.  This assumes the
// module exists, and will panic otherwise.
func (p *Environment) Module(module string) uint {
	ctx, ok := p.LookupModule(module)
	// Sanity check we found something
	if !ok {
		panic(fmt.Sprintf("unknown module %s", module))
	}
	// Discard column index
	return ctx
}

// Column determines the column index for a given column declared in a given
// module.  This assumes the column / module exist, and will panic otherwise.
func (p *Environment) Column(module uint, column string) uint {
	// FIXME: doesn't make sense using context here.
	cid, ok := p.LookupColumn(module, column)
	// Sanity check we found something
	if !ok {
		panic(fmt.Sprintf("unknown column %s", column))
	}
	// Discard column index
	return cid
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
