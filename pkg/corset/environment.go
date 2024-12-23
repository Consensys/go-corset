package corset

import (
	sc "github.com/consensys/go-corset/pkg/schema"
	tr "github.com/consensys/go-corset/pkg/trace"
)

// Environment provides an interface into the global scope which can be used for
// simply resolving column identifiers.
type Environment interface {
	// Module returns informartion about a given module, such as its module
	// identifier.
	Module(Module string) *ModuleInfo
	// RegisterOf returns the column identifier for a given column in a given
	// module, or panics if no such column exists.
	RegisterOf(module string, name string) *RegisterInfo
	// Convert a context from the high-level form into the lower level form
	// suitable for HIR.
	ContextOf(from Context) tr.Context
}

// GlobalEnvironment is a wrapper around a global scope.  The point, really, is
// to signal the change between a global scope whose columns have yet to be
// allocated, from an environment whose columns are allocated.
type GlobalEnvironment struct {
	// Map module names to records
	modules   map[string]*ModuleInfo
	registers map[columnId]*RegisterInfo
}

// columnId uniquely identifiers a Corset column.  Note, however, that
// multiplier Corset columns can be mapped to a single underlying register.
type columnId struct {
	module  string
	columnd string
}

// NewGlobalEnvironment constructs a new global environment from a global scope
// by allocating appropriate identifiers to all columns.
func NewGlobalEnvironment(scope *GlobalScope) GlobalEnvironment {
	modules := moduleAllocation(scope)
	registers := registerAllocation(modules, scope)
	// Allocate module identifiers
	columnId := uint(0)
	// Allocate input columns first.
	for _, m := range scope.modules {
		for _, b := range m.bindings {
			if binding, ok := b.(*ColumnBinding); ok && !binding.computed {
				//binding.AllocateId(columnId)
				// Increase the column id
				columnId += binding.dataType.Width()
			}
		}
	}
	// Allocate assignments second.
	for _, m := range scope.modules {
		for _, b := range m.bindings {
			if binding, ok := b.(*ColumnBinding); ok && binding.computed {
				//binding.AllocateId(columnId)
				// Increase the column id
				columnId += binding.dataType.Width()
			}
		}
	}
	// Done
	return GlobalEnvironment{modules, registers}
}

// Module returns informartion about a given module, such as its module
// identifier.
func (p GlobalEnvironment) Module(module string) *ModuleInfo {
	return p.modules[module]
}

// RegisterOf returns the column identifier for a given column in a given
// module, or panics if no such column exists.
func (p GlobalEnvironment) RegisterOf(module string, name string) *RegisterInfo {
	// Construct column identifier.
	cid := columnId{module, name}
	// Lookup column.
	return p.registers[cid]
}

// ContextOf constructs a trace context from a given corset context.
func (p GlobalEnvironment) ContextOf(from Context) tr.Context {
	// Determine Module Identifier
	mid := p.Module(from.Module()).Id
	// Construct underlying context from this.
	return tr.NewContext(mid, from.LengthMultiplier())
}

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

// RegisterInfo encapsulates information about a "register" in the underlying
// constraint system.  The rough analogy is that "register allocation" is
// applied to map Corset columns down to HIR columns (a.k.a. registers).  The
// distinction between columns at the Corset level, and registers at the HIR
// level is necessary for two reasons: firstly, one corset column can expand to
// several HIR registers; secondly, register allocation is applied to columns in
// different perspectives of the same module.
type RegisterInfo struct {
	// Context (i.e. module + multiplier) of this register.
	Context tr.Context
	// Name of this register
	Name string
	// Column Identifier for this register.
	Id uint
	// Underlying datatype of this register.
	DataType sc.Type
}

// Module allocation is a simple process of allocating modules their specific
// identifiers.  This has to match exactly how the translator does it, otherwise
// there will be problems.
func moduleAllocation(scope *GlobalScope) map[string]*ModuleInfo {
	modules := make(map[string]*ModuleInfo)
	moduleId := uint(0)
	// Allocate modules one-by-one
	for _, m := range scope.modules {
		modules[m.module] = &ModuleInfo{m.module, moduleId}
		moduleId++
	}
	// Done
	return modules
}

// Register allocation is the process of allocating columns to their underlying
// HIR columns (a.k.a registers).  This is straightforward when there is a 1-1
// mapping from a Corset column to an HIR column.  However, this is not always
// the case.  For example, array columns at the Corset level map to multiple
// columns at the HIR level.  Likewise, perspectives allow columns to be reused,
// meaning that multiple columns at the Corset level can be mapped down to just
// a single column at the HIR level.
func registerAllocation(modules map[string]*ModuleInfo, scope *GlobalScope) map[columnId]*RegisterInfo {
	registers := make(map[columnId]*RegisterInfo)
	registerId := uint(0)
	// Allocate input columns first.
	for _, m := range scope.modules {
		for _, b := range m.bindings {
			if binding, ok := b.(*ColumnBinding); ok && !binding.computed {
				registerAllocate(modules, registers, m.module, binding, registerId)
				// Increase the column id
				registerId += binding.dataType.Width()
			}
		}
	}
	// Allocate assignments second.
	for _, m := range scope.modules {
		for _, b := range m.bindings {
			if binding, ok := b.(*ColumnBinding); ok && binding.computed {
				registerAllocate(modules, registers, m.module, binding, registerId)
				// Increase the column id
				registerId += binding.dataType.Width()
			}
		}
	}
	// Apply aliases
	for _, m := range scope.modules {
		for id, binding_id := range m.ids {
			if binding, ok := m.bindings[binding_id].(*ColumnBinding); ok && !id.fn {
				orig := columnId{m.module, binding.name}
				alias := columnId{m.module, id.name}
				registers[alias] = registers[orig]
			}
		}
	}
	// Done
	return registers
}

// Allocate a single register.  This is just boilerplate really.
func registerAllocate(modules map[string]*ModuleInfo, registers map[columnId]*RegisterInfo, module string,
	binding *ColumnBinding, id uint) {
	//
	moduleId := modules[module].Id
	colId := columnId{module, binding.name}
	//
	registers[colId] = &RegisterInfo{
		tr.NewContext(moduleId, binding.multiplier),
		binding.name,
		id,
		binding.dataType.AsUnderlying(),
	}
}
