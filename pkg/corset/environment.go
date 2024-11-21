package corset

import (
	"fmt"

	"github.com/consensys/go-corset/pkg/schema"
	sc "github.com/consensys/go-corset/pkg/schema"
	"github.com/consensys/go-corset/pkg/trace"
	tr "github.com/consensys/go-corset/pkg/trace"
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
type ColumnInfo struct {
	// Column index
	cid uint
	// Length multiplier
	multiplier uint
	// Datatype
	datatype schema.Type
}

// IsFinalised checks whether this column has been finalised already.
func (p ColumnInfo) IsFinalised() bool {
	return p.multiplier != 0
}

// Environment maps module and column names to their (respective) module and
// column indices.  The environment separates input columns from assignment
// columns because they are disjoint in the schema being constructed (i.e. input
// columns always have a lower index than assignments).
type Environment struct {
	// Maps module names to their module indices.
	modules map[string]uint
	// Maps input columns to their column indices.
	columns map[colRef]ColumnInfo
}

// EmptyEnvironment constructs an empty environment.
func EmptyEnvironment() *Environment {
	modules := make(map[string]uint)
	columns := make(map[colRef]ColumnInfo)
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

// RegisterColumn registers a new column within a given module. Observe that
// this will panic if the column already exists. Furthermore, the column
// identifier is always determined as the next available identifier. Hence, care
// must be taken when declaring columns to ensure they are allocated in the
// right order.
func (p *Environment) RegisterColumn(context trace.Context, column string, datatype schema.Type) uint {
	if p.HasColumn(context.Module(), column) {
		panic(fmt.Sprintf("column %d:%s already exists", context.Module(), column))
	} else if datatype == nil {
		panic(fmt.Sprintf("column %d:%s cannot have nil type", context.Module(), column))
	} else if context.LengthMultiplier() == 0 {
		panic(fmt.Sprintf("column %d:%s cannot have 0 length multiplier", context.Module(), column))
	}
	// Update cache
	cid := uint(len(p.columns))
	cref := colRef{context.Module(), column}
	p.columns[cref] = ColumnInfo{cid, context.LengthMultiplier(), datatype}
	// Done
	return cid
}

// PreRegisterColumn makes an initial recording of the column and allocates a
// column identifier.  A pre-registered column is a column who registration has
// not yet been finalised.  More specifically the column is not considered
// finalised (i.e. ready for use) until FinaliseColumn is called.
func (p *Environment) PreRegisterColumn(module uint, column string) uint {
	if p.HasColumn(module, column) {
		panic(fmt.Sprintf("column %d:%s already exists", module, column))
	}
	// Update cache
	cid := uint(len(p.columns))
	cref := colRef{module, column}
	p.columns[cref] = ColumnInfo{cid, 0, nil}
	// Done
	return cid
}

// IsColumnFinalised determines whether a given column has been finalised yet,
// or not.  Observe this will panic if the column has not at least been
// pre-registered.
func (p *Environment) IsColumnFinalised(module uint, column string) bool {
	if !p.HasColumn(module, column) {
		panic(fmt.Sprintf("column %d:%s does not exist", module, column))
	}
	//
	cref := colRef{module, column}
	// Check information is finalised.
	return p.columns[cref].IsFinalised()
}

// FinaliseColumn finalises details of a columnm, specifically its length
// multiplier and type.  After this has been called, IsColumnFinalised should
// return true for the column in question.  Obserce this will panic if the
// column has not been preregistered, or if it is already finalised.
func (p *Environment) FinaliseColumn(context tr.Context, column string, datatype sc.Type) {
	// Sanity check we are not finalising a column which has already been finalised.
	if p.IsColumnFinalised(context.Module(), column) {
		panic(fmt.Sprintf("Attempt to refinalise column %s", column))
	}
	//
	cref := colRef{context.Module(), column}
	// Extract existing (incomplete) info
	info := p.columns[cref]
	// Update incomplete info
	p.columns[cref] = ColumnInfo{info.cid, context.LengthMultiplier(), datatype}
}

// LookupModule determines the module index for a given named module, or return
// false if no such module exists.
func (p *Environment) LookupModule(module string) (uint, bool) {
	mid, ok := p.modules[module]
	return mid, ok
}

// LookupColumn determines the column index for a given named column in a given
// module, or return false if no such column exists.  Observe this will return
// information even for columns which exist by are not yet finalised.
func (p *Environment) LookupColumn(module uint, column string) (ColumnInfo, bool) {
	cref := colRef{module, column}
	cinfo, ok := p.columns[cref]

	return cinfo, ok
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
// Furthermore, this assumes that the column is finalised and, otherwise, will
// panic.
func (p *Environment) Column(module uint, column string) ColumnInfo {
	info, ok := p.LookupColumn(module, column)
	// Sanity check we found something
	if !ok {
		panic(fmt.Sprintf("unknown column %s", column))
	} else if !info.IsFinalised() {
		panic(fmt.Sprintf("column %s not yet finalised", column))
	}
	// Done
	return info
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
