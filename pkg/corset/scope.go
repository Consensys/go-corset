package corset

import (
	"fmt"

	tr "github.com/consensys/go-corset/pkg/trace"
)

// Scope represents a region of code in which an expression can be evaluated.
// The purpose of a scope is to assist with determining what, exactly, a given
// variable used within an expression refers to.  For example, a variable can
// refer to a column, or a parameter, etc.
type Scope interface {
	// HasModule checks whether a given module exists, or not.
	HasModule(string) bool
	// Attempt to bind a given symbol within this scope.  If successful, the
	// symbol is then resolved with the appropriate binding.  Return value
	// indicates whether successful or not.
	Bind(Symbol) bool
}

// =============================================================================
// Global Scope
// =============================================================================

// GlobalScope represents the top-level scope in a Corset file, and is used to
// glue the scopes for modules together.  For example, it enables one module to
// lookup columns in another.
type GlobalScope struct {
	// Top-level mapping of modules to their scopes.
	ids map[string]uint
	// List of modules in declaration order
	modules []ModuleScope
}

// NewGlobalScope constructs an empty global scope.
func NewGlobalScope() *GlobalScope {
	return &GlobalScope{make(map[string]uint), make([]ModuleScope, 0)}
}

// DeclareModule declares an initialises a new module within this global scope.
// If a module by the same name already exists, then this will panic.
func (p *GlobalScope) DeclareModule(module string) {
	// Sanity check module doesn't already exist
	if _, ok := p.ids[module]; ok {
		panic(fmt.Sprintf("duplicate module %s declared", module))
	}
	// Register module
	mid := uint(len(p.ids))
	scope := ModuleScope{module, mid, make(map[BindingId]uint), make([]Binding, 0), p}
	p.modules = append(p.modules, scope)
	p.ids[module] = mid
}

// HasModule checks whether a given module exists, or not.
func (p *GlobalScope) HasModule(module string) bool {
	// Attempt to lookup the module
	_, ok := p.ids[module]
	// Return what we found
	return ok
}

// Bind looks up a given variable being referenced within a given module.  For a
// root context, this is either a column, an alias or a function declaration.
func (p *GlobalScope) Bind(symbol Symbol) bool {
	if !symbol.IsQualified() {
		// Search for symbol in root module.
		return p.Module("").Bind(symbol)
	} else if !p.HasModule(symbol.Module()) {
		// Pontially, it might be better to report a more useful error message.
		return false
	}
	//
	return p.Module(symbol.Module()).Bind(symbol)
}

// Module returns the identifier of the module with the given name.  Observe
// that this will panic if the module in question does not exist.
func (p *GlobalScope) Module(name string) *ModuleScope {
	if mid, ok := p.ids[name]; ok {
		return &p.modules[mid]
	}
	// Problem.
	panic(fmt.Sprintf("unknown module \"%s\"", name))
}

// ToEnvironment converts this global scope into a concrete environment by
// allocating all columns within this scope.
func (p *GlobalScope) ToEnvironment() Environment {
	return NewGlobalEnvironment(p)
}

// =============================================================================
// Module Scope
// =============================================================================

// ModuleScope represents the scope characterised by a module.
type ModuleScope struct {
	// Module name
	module string
	// Module identifier
	mid uint
	// Mapping from binding identifiers to indices within the bindings array.
	ids map[BindingId]uint
	// The set of bindings in the order of declaration.
	bindings []Binding
	// Enclosing global scope
	enclosing Scope
}

// EnclosingModule returns the name of the enclosing module.  This is generally
// useful for reporting errors.
func (p *ModuleScope) EnclosingModule() string {
	return p.module
}

// HasModule checks whether a given module exists, or not.
func (p *ModuleScope) HasModule(module string) bool {
	return p.enclosing.HasModule(module)
}

// Bind looks up a given variable being referenced within a given module.  For a
// root context, this is either a column, an alias or a function declaration.
func (p *ModuleScope) Bind(symbol Symbol) bool {
	// Determine module for this lookup.
	if symbol.IsQualified() && symbol.Module() != p.module {
		// non-local lookup
		return p.enclosing.Bind(symbol)
	}
	// construct binding identifier
	id := BindingId{symbol.Name(), symbol.IsFunction()}
	// Look for it.
	if bid, ok := p.ids[id]; ok {
		// Extract binding
		binding := p.bindings[bid]
		// Resolve symbol
		return symbol.Resolve(binding)
	} else if !symbol.IsQualified() && p.module != "" {
		// Attempt to lookup in parent (unless we are the root module, in which
		// case we have no parent)
		return p.enclosing.Bind(symbol)
	} else {
		return false
	}
}

// Binding returns information about the binding of a particular symbol defined
// in this module.
func (p *ModuleScope) Binding(name string) Binding {
	// construct binding identifier
	if bid, ok := p.ids[BindingId{name, false}]; ok {
		return p.bindings[bid]
	}
	// Failure
	return nil
}

// Column returns information about a particular column declared within this
// module.
func (p *ModuleScope) Column(name string) *ColumnBinding {
	// construct binding identifier
	bid := p.ids[BindingId{name, false}]
	//
	return p.bindings[bid].(*ColumnBinding)
}

// Declare declares a given binding within this module scope.
func (p *ModuleScope) Declare(symbol SymbolDefinition) bool {
	// construct binding identifier
	bid := BindingId{symbol.Name(), symbol.IsFunction()}
	// Sanity check not already declared
	if _, ok := p.ids[bid]; ok {
		// Cannot redeclare
		return false
	}
	// Done
	id := uint(len(p.bindings))
	p.bindings = append(p.bindings, symbol.Binding())
	p.ids[bid] = id
	//
	return true
}

// Alias constructs an alias for an existing symbol.  If the symbol does not
// exist, then this returns false.
func (p *ModuleScope) Alias(alias string, symbol Symbol) bool {
	// construct symbol identifier
	symbol_id := BindingId{symbol.Name(), symbol.IsFunction()}
	// construct alias identifier
	alias_id := BindingId{alias, symbol.IsFunction()}
	// Check alias does not already exist
	if _, ok := p.ids[alias_id]; !ok {
		// Check symbol being aliased exists
		if id, ok := p.ids[symbol_id]; ok {
			p.ids[alias_id] = id
			// Done
			return true
		}
	}
	// Symbol not known (yet)
	return false
}

// =============================================================================
// Local Scope
// =============================================================================

// LocalScope represents a simple implementation of scope in which local
// variables can be declared.  A local scope must have a single context
// associated with it, and this will be inferred by resolving those expressions
// which must be evaluated within.
type LocalScope struct {
	global bool
	// Determines whether or not this scope is "pure" (i.e. whether or not
	// columns can be accessed, etc).
	pure bool
	// Represents the enclosing scope
	enclosing Scope
	// Context for this scope
	context *Context
	// Maps inputs parameters to the declaration index.
	locals map[string]uint
	// Actual parameter bindings
	bindings []*LocalVariableBinding
}

// NewLocalScope constructs a new local scope within a given enclosing scope.  A
// local scope can have local variables declared within it.  A local scope can
// also be "global" in the sense that accessing symbols from other modules is
// permitted.
func NewLocalScope(enclosing Scope, global bool, pure bool) LocalScope {
	context := tr.VoidContext[string]()
	locals := make(map[string]uint)
	bindings := make([]*LocalVariableBinding, 0)
	//
	return LocalScope{global, pure, enclosing, &context, locals, bindings}
}

// NestedScope creates a nested scope within this local scope.
func (p LocalScope) NestedScope() LocalScope {
	nlocals := make(map[string]uint)
	nbindings := make([]*LocalVariableBinding, len(p.bindings))
	// Clone allocated variables
	for k, v := range p.locals {
		nlocals[k] = v
	}
	// Copy over bindings.
	copy(nbindings, p.bindings)
	// Done
	return LocalScope{p.global, p.pure, p, p.context, nlocals, nbindings}
}

// NestedPureScope creates a nested scope within this local scope which, in
// addition, is always pure.
func (p LocalScope) NestedPureScope() LocalScope {
	nlocals := make(map[string]uint)
	nbindings := make([]*LocalVariableBinding, len(p.bindings))
	// Clone allocated variables
	for k, v := range p.locals {
		nlocals[k] = v
	}
	// Copy over bindings.
	copy(nbindings, p.bindings)
	// Done
	return LocalScope{p.global, true, p, p.context, nlocals, nbindings}
}

// IsGlobal determines whether symbols can be accessed in modules other than the
// enclosing module.
func (p LocalScope) IsGlobal() bool {
	return p.global
}

// IsPure determines whether or not this scope is pure.  That is, whether or not
// expressions in this scope are permitted to access columns (either directly,
// or indirectly via impure invocations).
func (p LocalScope) IsPure() bool {
	return p.pure
}

// FixContext fixes the context for this scope.  Since every scope requires
// exactly one context, this fails if we fix it to incompatible contexts.
func (p LocalScope) FixContext(context Context) bool {
	// Join contexts together
	*p.context = p.context.Join(context)
	// Check they were compatible
	return !p.context.IsConflicted()
}

// HasModule checks whether a given module exists, or not.
func (p LocalScope) HasModule(module string) bool {
	return p.enclosing.HasModule(module)
}

// Bind looks up a given variable or function being referenced either within the
// enclosing scope (module==nil) or within a specified module.
func (p LocalScope) Bind(symbol Symbol) bool {
	// Check whether this is a local variable access.
	if id, ok := p.locals[symbol.Name()]; ok && !symbol.IsFunction() && !symbol.IsQualified() {
		// Yes, this is a local variable access.
		return symbol.Resolve(p.bindings[id])
	}
	// No, this is not a local variable access.
	return p.enclosing.Bind(symbol)
}

// DeclareLocal registers a new local variable (e.g. a parameter).
func (p *LocalScope) DeclareLocal(name string, binding *LocalVariableBinding) uint {
	index := uint(len(p.locals))
	binding.Finalise(index)
	p.locals[name] = index
	p.bindings = append(p.bindings, binding)
	// Return variable index
	return index
}
