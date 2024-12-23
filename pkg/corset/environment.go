package corset

import (
	"fmt"

	sc "github.com/consensys/go-corset/pkg/schema"
	tr "github.com/consensys/go-corset/pkg/trace"
)

// ModuleInfo provides information about a module in the underlying HIR
// constraint set.
type ModuleInfo struct {
	// Name of this module
	Name string
	// Module identifier
	Id uint
}

// Context constructs a new context for this module assuming a given length
// multiplier.
func (p *ModuleInfo) Context(multiplier uint) tr.Context {
	return tr.NewContext(p.Id, multiplier)
}

// Environment provides an interface into the global scope which can be used for
// simply resolving column identifiers.
type Environment interface {
	// Module returns informartion about a given module, such as its module
	// identifier.
	Module(Module string) *ModuleInfo
	// RegisterOf returns the underlying (HIR) column identifier and
	// details of the register to which the column is allocated.
	RegisterOf(module string, name string) (uint, *Register)
	// Convert a context from the high-level form into the lower level form
	// suitable for HIR.
	ContextOf(from Context) tr.Context
}

// ColumnId uniquely identifiers a Corset column.  Note, however, that
// multiple Corset columns can be mapped to a single underlying register.
type ColumnId struct {
	module  string
	columnd string
}

// GlobalEnvironment is a wrapper around a global scope.  The point, really, is
// to signal the change between a global scope whose columns have yet to be
// allocated, from an environment whose columns are allocated.
type GlobalEnvironment struct {
	// Info about modules
	modules map[string]*ModuleInfo
	// Map source-level columns to registers
	columns map[ColumnId]uint
	// Registers
	registers []Register
}

// NewGlobalEnvironment constructs a new global environment from a global scope
// by allocating appropriate identifiers to all columns.
func NewGlobalEnvironment(scope *GlobalScope) GlobalEnvironment {
	env := GlobalEnvironment{nil, nil, nil}
	env.initModules(scope)
	env.initColumnsAndRegisters(scope)
	// Done
	return env
}

// Module returns informartion about a given module, such as its module
// identifier.
func (p GlobalEnvironment) Module(module string) *ModuleInfo {
	return p.modules[module]
}

// RegisterOf returns the column identifier for a given column in a given
// module, or panics if no such column exists.
func (p GlobalEnvironment) RegisterOf(module string, name string) (uint, *Register) {
	// Construct column identifier.
	cid := ColumnId{module, name}
	regId := p.columns[cid]
	// Lookup register info
	return regId, &p.registers[regId]
}

// ContextOf constructs a trace context from a given corset context.
func (p GlobalEnvironment) ContextOf(from Context) tr.Context {
	// Determine Module Identifier
	mid := p.Module(from.Module()).Id
	// Construct underlying context from this.
	return tr.NewContext(mid, from.LengthMultiplier())
}

// ===========================================================================
// Helpers
// ===========================================================================

// Module allocation is a simple process of allocating modules their specific
// identifiers.  This has to match exactly how the translator does it, otherwise
// there will be problems.
func (p *GlobalEnvironment) initModules(scope *GlobalScope) {
	p.modules = make(map[string]*ModuleInfo)
	moduleId := uint(0)
	// Allocate modules one-by-one
	for _, m := range scope.modules {
		p.modules[m.module] = &ModuleInfo{m.module, moduleId}
		moduleId++
	}
}

// Performs an initial register allocation which simply maps every column to a
// unique register.  The intention is that, subsequently, registers can be
// merged as necessary.
func (p *GlobalEnvironment) initColumnsAndRegisters(scope *GlobalScope) {
	p.columns = make(map[ColumnId]uint)
	p.registers = make([]Register, 0)
	// Allocate input columns first.
	for _, m := range scope.modules {
		for _, b := range m.bindings {
			if binding, ok := b.(*ColumnBinding); ok && !binding.computed {
				if m.module != binding.module {
					panic("unreachable?")
				}
				//
				p.allocateColumn(binding)
			}
		}
	}
	// Allocate assignments second.
	for _, m := range scope.modules {
		for _, b := range m.bindings {
			if binding, ok := b.(*ColumnBinding); ok && binding.computed {
				if m.module != binding.module {
					panic("unreachable?")
				}

				p.allocateColumn(binding)
			}
		}
	}
	// Apply aliases
	for _, m := range scope.modules {
		for id, binding_id := range m.ids {
			if binding, ok := m.bindings[binding_id].(*ColumnBinding); ok && !id.fn {
				orig := ColumnId{m.module, binding.name}
				alias := ColumnId{m.module, id.name}
				p.columns[alias] = p.columns[orig]
			}
		}
	}
}

// Allocate a source-level column into this environment.  Since a source-level
// column can correspond to multiple underling registers, this can result in the
// allocation of a number of registers (based on the columns type).  For
// example, an array of length n will allocate n registers, etc.
func (p *GlobalEnvironment) allocateColumn(column *ColumnBinding) {
	p.allocate(column.module, column.name, column.multiplier, column.dataType, column)
}

func (p *GlobalEnvironment) allocate(module string, name string, multiplier uint, datatype Type,
	binding *ColumnBinding) {
	// Check for base base
	if datatype.AsUnderlying() != nil {
		p.allocateUnit(module, name, multiplier, datatype.AsUnderlying(), binding)
	} else if arraytype, ok := datatype.(*ArrayType); ok {
		// For now, assume must be an array
		p.allocateArray(module, name, multiplier, arraytype, binding)
	} else {
		panic(fmt.Sprintf("unknown type encountered: %v", datatype))
	}
}

// Allocate an array type
func (p *GlobalEnvironment) allocateArray(module string, name string, multiplier uint, arraytype *ArrayType,
	binding *ColumnBinding) {
	// Allocate n columns
	for i := arraytype.min; i <= arraytype.max; i++ {
		ith_name := fmt.Sprintf("%s_%d", name, i)
		p.allocate(module, ith_name, multiplier, arraytype.element, binding)
	}
}

// Allocate a single register.
func (p *GlobalEnvironment) allocateUnit(module string, name string, multiplier uint, datatype sc.Type,
	binding *ColumnBinding) {
	moduleId := p.modules[module].Id
	colId := ColumnId{module, name}
	regId := uint(len(p.registers))
	// Allocate register
	p.registers = append(p.registers, Register{
		tr.NewContext(moduleId, multiplier),
		name,
		datatype,
		[]*ColumnBinding{binding},
	})
	// Map column to register
	p.columns[colId] = regId
}
