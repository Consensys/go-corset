package corset

import (
	tr "github.com/consensys/go-corset/pkg/trace"
)

// Environment provides an interface into the global scope which can be used for
// simply resolving column identifiers.
type Environment interface {
	// Module returns the module identifier for a given module, or panics if no
	// such module exists.
	Module(name string) *ModuleScope
	// Column returns the column identifier for a given column in a given
	// module, or panics if no such column exists.
	Column(module string, name string) *ColumnBinding
	// Convert a context from the high-level form into the lower level form
	// suitable for HIR.
	ToContext(from Context) tr.Context
	// Construct a trace context from a given module and multiplier.
	ContextFrom(module string, multiplier uint) tr.Context
}

// GlobalEnvironment is a wrapper around a global scope.  The point, really, is
// to signal the change between a global scope whose columns have yet to be
// allocated, from an environment whose columns are allocated.
type GlobalEnvironment struct {
	scope *GlobalScope
}

// NewGlobalEnvironment constructs a new global environment from a global scope
// by allocating appropriate identifiers to all columns.
func NewGlobalEnvironment(scope *GlobalScope) GlobalEnvironment {
	columnId := uint(0)
	// Allocate input columns first.
	for _, m := range scope.modules {
		for _, b := range m.bindings {
			if binding, ok := b.(*ColumnBinding); ok && !binding.computed {
				binding.AllocateId(columnId)
				// Increase the column id
				columnId++
			}
		}
	}
	// Allocate assignments second.
	for _, m := range scope.modules {
		for _, b := range m.bindings {
			if binding, ok := b.(*ColumnBinding); ok && binding.computed {
				binding.AllocateId(columnId)
				// Increase the column id
				columnId++
			}
		}
	}
	// Done
	return GlobalEnvironment{scope}
}

// Module returns the identifier of the module with the given name.
func (p GlobalEnvironment) Module(name string) *ModuleScope {
	return p.scope.Module(name)
}

// Column returns the column identifier for a given column in a given
// module, or panics if no such column exists.
func (p GlobalEnvironment) Column(module string, name string) *ColumnBinding {
	// Lookup the given binding, expecting that it is a column binding.  If not,
	// then this will fail.
	return p.Module(module).Bind(nil, name, false).(*ColumnBinding)
}

// ContextFrom constructs a trace context for a given module and length
// multiplier.
func (p GlobalEnvironment) ContextFrom(module string, multiplier uint) tr.Context {
	return tr.NewContext(p.Module(module).mid, multiplier)
}

// ToContext constructs a trace context from a given corset context.
func (p GlobalEnvironment) ToContext(from Context) tr.Context {
	return p.ContextFrom(from.Module(), from.LengthMultiplier())
}
