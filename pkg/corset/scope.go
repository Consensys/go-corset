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
	// Get the name of the enclosing module.  This is generally useful for
	// reporting errors.
	EnclosingModule() uint
	// HasModule checks whether a given module exists, or not.
	HasModule(string) bool
	// Lookup the identifier for a given module.  This assumes that the module
	// exists, and will panic otherwise.
	Module(string) uint
	// Lookup a given variable being referenced with an optional module
	// specifier.  This variable could correspond to a column, a function, a
	// parameter, or a local variable.  Furthermore, the returned binding will
	// be nil if this variable does not exist.
	Bind(*uint, string, bool) Binding
}

// =============================================================================
// Module Scope
// =============================================================================

// ModuleScope represents the scope characterised by a module.
type ModuleScope struct {
	// Module ID
	module uint
	// Provides access to global environment
	environment *Environment
	// Maps function names to their contents.
	functions map[string]FunctionBinding
}

// EnclosingModule returns the name of the enclosing module.  This is generally
// useful for reporting errors.
func (p *ModuleScope) EnclosingModule() uint {
	return p.module
}

// HasModule checks whether a given module exists, or not.
func (p *ModuleScope) HasModule(module string) bool {
	return p.environment.HasModule(module)
}

// Module determines the module index for a given module.  This assumes the
// module exists, and will panic otherwise.
func (p *ModuleScope) Module(module string) uint {
	return p.environment.Module(module)
}

// Bind looks up a given variable being referenced within a given module.  For a
// root context, this is either a column, an alias or a function declaration.
func (p *ModuleScope) Bind(module *uint, name string, fn bool) Binding {
	var mid uint
	// Determine module for this lookup.
	if module != nil {
		mid = *module
	} else {
		mid = p.module
	}
	// Lookup function
	if binding, ok := p.functions[name]; ok && module == nil {
		return &binding
	} else if info, ok := p.environment.LookupColumn(mid, name); ok && !fn {
		ctx := tr.NewContext(mid, info.multiplier)
		return &ColumnBinding{ctx, info.cid}
	}
	// error
	return nil
}

// DeclareFunction declares a given function within this module scope.
func (p *ModuleScope) DeclareFunction(name string, arity uint, body Expr) {
	if _, ok := p.functions[name]; ok {
		panic(fmt.Sprintf("attempt to redeclared function \"%s\"/%d", name, arity))
	}
	//
	p.functions[name] = FunctionBinding{arity, body}
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
	// Represents the enclosing scope
	enclosing Scope
	// Context for this scope
	context *tr.Context
	// Maps inputs parameters to the declaration index.
	locals map[string]uint
}

// NewLocalScope constructs a new local scope within a given enclosing scope.  A
// local scope can have local variables declared within it.  A local scope can
// also be "global" in the sense that accessing symbols from other modules is
// permitted.
func NewLocalScope(enclosing Scope, global bool) LocalScope {
	context := tr.VoidContext()
	locals := make(map[string]uint)
	//
	return LocalScope{global, enclosing, &context, locals}
}

// NestedScope creates a nested scope within this local scope.
func (p LocalScope) NestedScope() LocalScope {
	nlocals := make(map[string]uint)
	// Clone allocated variables
	for k, v := range p.locals {
		nlocals[k] = v
	}
	// Done
	return LocalScope{p.global, p.enclosing, p.context, nlocals}
}

// IsGlobal determines whether symbols can be accessed in modules other than the
// enclosing module.
func (p LocalScope) IsGlobal() bool {
	return p.global
}

// EnclosingModule returns the name of the enclosing module.  This is generally
// useful for reporting errors.
func (p LocalScope) EnclosingModule() uint {
	return p.enclosing.EnclosingModule()
}

// FixContext fixes the context for this scope.  Since every scope requires
// exactly one context, this fails if we fix it to incompatible contexts.
func (p LocalScope) FixContext(context tr.Context) bool {
	// Join contexts together
	*p.context = p.context.Join(context)
	// Check they were compatible
	return !p.context.IsConflicted()
}

// HasModule checks whether a given module exists, or not.
func (p LocalScope) HasModule(module string) bool {
	return p.enclosing.HasModule(module)
}

// Module determines the module index for a given module.  This assumes the
// module exists, and will panic otherwise.
func (p LocalScope) Module(module string) uint {
	return p.enclosing.Module(module)
}

// Bind looks up a given variable or function being referenced either within the
// enclosing scope (module==nil) or within a specified module.
func (p LocalScope) Bind(module *uint, name string, fn bool) Binding {
	// Check whether this is a local variable access.
	if id, ok := p.locals[name]; ok && !fn && module == nil {
		// Yes, this is a local variable access.
		return &ParameterBinding{id}
	}
	// No, this is not a local variable access.
	return p.enclosing.Bind(module, name, fn)
}

// DeclareLocal registers a new local variable (e.g. a parameter).
func (p LocalScope) DeclareLocal(name string) uint {
	index := uint(len(p.locals))
	p.locals[name] = index
	// Return variable index
	return index
}
