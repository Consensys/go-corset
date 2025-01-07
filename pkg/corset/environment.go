package corset

import (
	"fmt"

	sc "github.com/consensys/go-corset/pkg/schema"
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
	ContextOf(from Context) tr.Context
}

// ColumnId uniquely identifiers a Corset column.  Note, however, that
// multiple Corset columns can be mapped to a single underlying register.
type ColumnId struct {
	module string
	column string
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
// by allocating appropriate identifiers to all columns.  This process is
// parameterised upon a given register allocator, thus enabling different
// allocation algorithms.
func NewGlobalEnvironment(scope *GlobalScope, allocator func(RegisterAllocation)) GlobalEnvironment {
	env := GlobalEnvironment{nil, nil, nil}
	env.initModules(scope)
	env.initColumnsAndRegisters(scope)
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
	// FIXME: this is broken
	module := column.Parent().String()
	name := column.Tail()
	// Construct column identifier.
	cid := ColumnId{module, name}
	regId := p.columns[cid]
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
func (p GlobalEnvironment) ColumnsOf(register uint) []ColumnId {
	var columns []ColumnId
	//
	for col, reg := range p.columns {
		if reg == register {
			columns = append(columns, col)
		}
	}
	//
	return columns
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
		name := m.path.String()
		p.modules[name] = &ModuleInfo{name, moduleId}
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
				p.allocateColumn(binding)
			}
		}
	}
	// Allocate assignments second.
	for _, m := range scope.modules {
		for _, b := range m.bindings {
			if binding, ok := b.(*ColumnBinding); ok && binding.computed {
				p.allocateColumn(binding)
			}
		}
	}
	// Apply aliases
	for _, m := range scope.modules {
		name := m.path.String()
		//
		for id, binding_id := range m.ids {
			if binding, ok := m.bindings[binding_id].(*ColumnBinding); ok && !id.fn {
				orig := ColumnId{name, binding.path.Tail()}
				alias := ColumnId{name, id.name}
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
	p.allocate(column, &column.path, column.dataType)
}

func (p *GlobalEnvironment) allocate(column *ColumnBinding, path *util.Path, datatype Type) {
	// Check for base base
	if datatype.AsUnderlying() != nil {
		p.allocateUnit(column, path, datatype.AsUnderlying())
	} else if arraytype, ok := datatype.(*ArrayType); ok {
		// For now, assume must be an array
		p.allocateArray(column, path, arraytype)
	} else {
		panic(fmt.Sprintf("unknown type encountered: %v", datatype))
	}
}

// Allocate an array type
func (p *GlobalEnvironment) allocateArray(column *ColumnBinding, path *util.Path, arraytype *ArrayType) {
	// Allocate n columns
	for i := arraytype.min; i <= arraytype.max; i++ {
		ith_name := fmt.Sprintf("%s_%d", path.Tail(), i)
		ith_path := path.Parent().Extend(ith_name)
		p.allocate(column, ith_path, arraytype.element)
	}
}

// Allocate a single register.
func (p *GlobalEnvironment) allocateUnit(column *ColumnBinding, path *util.Path, datatype sc.Type) {
	// FIXME: following is broken because we lose perspective information.
	module := path.Parent().String()
	name := path.Tail()
	//
	moduleId := p.modules[module].Id
	colId := ColumnId{module, name}
	regId := uint(len(p.registers))
	// Construct appropriate register source.
	source := RegisterSource{
		*path,
		column.multiplier,
		datatype,
		column.mustProve,
		column.computed}
	// Allocate register
	p.registers = append(p.registers, Register{
		tr.NewContext(moduleId, column.multiplier),
		name,
		datatype,
		[]RegisterSource{source},
	})
	// Map column to register
	p.columns[colId] = regId
}

// Apply the given register allocator to each module of this environment in turn.
func (p *GlobalEnvironment) applyRegisterAllocation(allocator func(RegisterAllocation)) {
	// Apply to each module in turn
	for m := range p.modules {
		// Determine register subset for this module
		view := p.RegistersOf(m)
		// Apply allocation to this subset
		allocator(&localRegisterAllocation{view, p})
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
	for col, reg := range p.columns {
		// Safe since as neither adding nor removing entry from map.
		p.columns[col] = mapping[reg]
	}
	// Copy over new register set, whilst slicing off inactive ones.
	p.registers = nregisters[0:j]
}

// ===========================================================================
// RegisterAllocation impl
// ===========================================================================

// LocalRegisterAllocation provides a view of the environment for the purposes
// of register allocation, such that only registers in this view will be
// considered for allocation.  This is necessary because we must not attempt to
// allocate registers across different modules (indeed, contexts) together.
// Instead, we must allocate registers on a module-by-module basis, etc.
type localRegisterAllocation struct {
	// View of registers available for register allocation.
	registers []uint
	// Parent pointer for register merging.
	env *GlobalEnvironment
}

// Len returns the number of allocated registers.
func (p *localRegisterAllocation) Len() uint {
	return uint(len(p.registers))
}

// Registers returns an iterator over the set of registers in this local
// allocation.
func (p *localRegisterAllocation) Registers() util.Iterator[uint] {
	return util.NewArrayIterator(p.registers)
}

// Access information about a specific register in this window.
func (p *localRegisterAllocation) Register(index uint) *Register {
	return &p.env.registers[index]
}

// Merge one register (src) into another (dst).  This will remove the src
// register, and automatically update all column assignments.  Therefore, any
// register identifier can be potenitally invalided by this operation.  This
// will panic if the registers are incompatible (i.e. have different contexts).
func (p *localRegisterAllocation) Merge(dst uint, src uint) {
	target := &p.env.registers[dst]
	source := &p.env.registers[src]
	// Sanity check
	if target.Context != source.Context {
		// Should be unreachable.
		panic("attempting to merge incompatible registers")
	}
	// Update column map
	for _, col := range p.env.ColumnsOf(src) {
		p.env.columns[col] = dst
	}
	//
	target.Merge(source)
}
