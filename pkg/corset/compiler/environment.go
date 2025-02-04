package compiler

import (
	"github.com/consensys/go-corset/pkg/corset/ast"
	tr "github.com/consensys/go-corset/pkg/trace"
	"github.com/consensys/go-corset/pkg/util"
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
	// Register returns information about a given register, based on its index
	// (i.e. underlying HIR column identifier).
	Register(index uint) *Register
	// RegisterOf identifiers the register (i.e. underlying (HIR) column) to
	// which a given source-level (i.e. corset) column is allocated.  This
	// expects an absolute path.
	RegisterOf(path *util.Path) uint
	// RegistersOf identifies the set of registers (i.e. underlying (HIR)
	// columns) associated with a given module.
	RegistersOf(module string) []uint
	// Convert a context from the high-level form into the lower level form
	// suitable for HIR.
	ContextOf(from ast.Context) tr.Context
}

// GlobalEnvironment is a wrapper around a global scope.  The point, really, is
// to signal the change between a global scope whose columns have yet to be
// allocated, from an environment whose columns are allocated.
type GlobalEnvironment struct {
	// Info about modules
	modules map[string]*ModuleInfo
	// Registers (i.e. HIR-level columns)
	registers []Register
	// Map source-level columnMap to registers
	columnMap map[string]uint
}

// NewGlobalEnvironment constructs a new global environment from a global scope
// by allocating appropriate identifiers to all columns.  This process is
// parameterised upon a given register allocator, thus enabling different
// allocation algorithms.
func NewGlobalEnvironment(root *ModuleScope, allocator func(RegisterAllocation)) GlobalEnvironment {
	// Sanity Check
	if !root.IsRoot() {
		// Definitely should be unreachable.
		panic("root scope required")
	}
	// Construct top-level module list.
	modules := root.Flattern()
	// Initialise the environment
	env := GlobalEnvironment{nil, nil, nil}
	env.initModules(modules)
	env.initColumnsAndRegisters(modules)
	// Apply register allocation.
	env.applyRegisterAllocation(allocator)
	// Done
	return env
}

// Module returns informartion about a given module, such as its module
// identifier.
func (p GlobalEnvironment) Module(module string) *ModuleInfo {
	return p.modules[module]
}

// Register returns information about a given register, based on its index
// (i.e. underlying HIR column identifier).
func (p GlobalEnvironment) Register(index uint) *Register {
	return &p.registers[index]
}

// RegisterOf identifies the register (i.e. underlying (HIR) column) to
// which a given source-level (i.e. corset) column is allocated.
func (p GlobalEnvironment) RegisterOf(column *util.Path) uint {
	regId := p.columnMap[column.String()]
	// Lookup register info
	return regId
}

// RegistersOf identifies the set of registers (i.e. underlying (HIR)
// columns) associated with a given module.
func (p GlobalEnvironment) RegistersOf(module string) []uint {
	mid := p.modules[module].Id
	regs := make([]uint, 0)
	// Iterate all registers looking for those in the given module.
	for i, reg := range p.registers {
		if reg.Context.Module() == mid {
			// match
			regs = append(regs, uint(i))
		}
	}
	// Done
	return regs
}

// ColumnsOf returns the set of registers allocated to a given column.
func (p GlobalEnvironment) ColumnsOf(register uint) []string {
	var columns []string
	//
	for col, reg := range p.columnMap {
		if reg == register {
			columns = append(columns, col)
		}
	}
	//
	return columns
}

// ContextOf constructs a trace context from a given corset context.
func (p GlobalEnvironment) ContextOf(from ast.Context) tr.Context {
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
func (p *GlobalEnvironment) initModules(modules []*ModuleScope) {
	p.modules = make(map[string]*ModuleInfo)
	moduleId := uint(0)
	// Allocate submodules one-by-one
	for _, m := range modules {
		if !m.Virtual() {
			name := m.path.String()
			p.modules[name] = &ModuleInfo{name, moduleId}
			moduleId++
		}
	}
}

// Performs an initial register allocation which simply maps every column to a
// unique register.  The intention is that, subsequently, registers can be
// merged as necessary.
func (p *GlobalEnvironment) initColumnsAndRegisters(modules []*ModuleScope) {
	p.columnMap = make(map[string]uint)
	p.registers = make([]Register, 0)
	// Allocate input columns first.
	for _, m := range modules {
		for _, col := range m.DestructuredColumns() {
			if !col.Computed {
				p.allocateRegister(col)
			}
		}
	}
	// Allocate assignments second.
	for _, m := range modules {
		for _, col := range m.DestructuredColumns() {
			if col.Computed {
				p.allocateRegister(col)
			}
		}
	}
	// Apply aliases
	for _, m := range modules {
		for id, binding_id := range m.ids {
			if binding, ok := m.bindings[binding_id].(*ast.ColumnBinding); ok && !id.fn {
				orig := binding.Path.String()
				alias := m.path.Extend(id.name).String()
				p.columnMap[alias] = p.columnMap[orig]
			}
		}
	}
}

// Allocate a source-level column into this environment.  Since a source-level
// column can correspond to multiple underling registers, this can result in the
// allocation of a number of registers (based on the columns type).  For
// example, an array of length n will allocate n registers, etc.
func (p *GlobalEnvironment) allocateRegister(source RegisterSource) {
	module := source.Context.String()
	//
	moduleId := p.modules[module].Id
	regId := uint(len(p.registers))
	// Allocate register
	p.registers = append(p.registers, Register{
		tr.NewContext(moduleId, source.Multiplier),
		source.DataType,
		[]RegisterSource{source},
		nil,
	})
	// Map column to register
	p.columnMap[source.Name.String()] = regId
}

// Apply the given register allocator to each module of this environment in turn.
func (p *GlobalEnvironment) applyRegisterAllocation(allocator func(RegisterAllocation)) {
	// Apply to each module in turn
	for m := range p.modules {
		// Determine register subset for this module
		view := p.RegistersOf(m)
		// Apply allocation to this subset
		allocator(&RegisterAllocationView{view, p})
	}
	// Remove inactive registers.  This is necessary because register allocation
	// marks a register as inactive when they its merged into another, but does
	// not actually delete the register.
	mapping := make([]uint, len(p.registers))
	// Overallocate set of new registers
	nregisters := make([]Register, len(p.registers))
	// Index into nregisters
	j := uint(0)
	// Build mapping and remove registers
	for i := 0; i < len(p.registers); i++ {
		ith := p.registers[i]
		//
		if ith.IsActive() {
			mapping[i] = j
			nregisters[j] = ith
			j++
		}
	}
	// Update the columns maps, etc.
	for col, reg := range p.columnMap {
		// Safe since as neither adding nor removing entry from map.
		p.columnMap[col] = mapping[reg]
	}
	// Copy over new register set, whilst slicing off inactive ones.
	p.registers = nregisters[0:j]
}
