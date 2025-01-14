package corset

import (
	"fmt"

	tr "github.com/consensys/go-corset/pkg/trace"
	"github.com/consensys/go-corset/pkg/util"
)

// Scope represents a region of code in which an expression can be evaluated.
// The purpose of a scope is to assist with determining what, exactly, a given
// variable used within an expression refers to.  For example, a variable can
// refer to a column, or a parameter, etc.
type Scope interface {
	// Attempt to bind a given symbol within this scope.  If successful, the
	// symbol is then resolved with the appropriate binding.  Return value
	// indicates whether successful or not.
	Bind(Symbol) bool
}

// BindingId is an identifier is used to distinguish different forms of binding,
// as some forms are known from their use.  Specifically, at the current time,
// only functions are distinguished from other categories (e.g. columns,
// parameters, etc).
type BindingId struct {
	// Name of the binding
	name string
	// Indicates whether function binding or other.
	fn bool
}

// =============================================================================
// Module Scope
// =============================================================================

// ModuleScope defines recursive tree of scopes where symbols can be resolved
// and bound.  The primary goal is to handle the various ways in which a
// symbol's qualified name (i.e. path) can be expressed.  For example, a symbol
// can be given an absolute name (which is resolved from the root of the scope
// tree), or it can be relative (in which case it is resolved relative to a
// given module).
type ModuleScope struct {
	// Absolute path
	path util.Path
	// Map identifiers to indices within the bindings array.
	ids map[BindingId]uint
	// The set of bindings in the order of declaration.
	bindings []Binding
	// Enclosing scope
	parent *ModuleScope
	// Submodules in a map (for efficient lookup)
	submodmap map[string]*ModuleScope
	// Submodules in the order of declaration (for determinism).
	submodules []*ModuleScope
	// Indicates whether or not this is a real module.
	virtual bool
}

// NewModuleScope constructs an initially empty top-level scope.
func NewModuleScope() *ModuleScope {
	return &ModuleScope{
		util.NewAbsolutePath(),
		make(map[BindingId]uint),
		nil,
		nil,
		make(map[string]*ModuleScope),
		nil,
		false,
	}
}

// IsRoot checks whether or not this is the root of the module tree.
func (p *ModuleScope) IsRoot() bool {
	return p.parent == nil
}

// Owner returns the enclosing non-virtual module of this module.  Observe
// that, if this is a non-virtual module, then it is returned.
func (p *ModuleScope) Owner() *ModuleScope {
	if !p.virtual {
		return p
	} else if p.parent != nil {
		return p.parent.Owner()
	}
	// Should be unreachable
	panic("invalid module tree")
}

// Declare a new submodule at the given (absolute) path within this tree scope.
// Submodules can be declared as "virtual" which indicates the submodule is
// simply a subset of rows of its enclosing module.  This returns true if this
// succeeds, otherwise returns false (i.e. a matching submodule already exists).
func (p *ModuleScope) Declare(submodule string, virtual bool) bool {
	if _, ok := p.submodmap[submodule]; ok {
		return false
	}
	// Construct suitable child scope
	scope := &ModuleScope{
		*p.path.Extend(submodule),
		make(map[BindingId]uint),
		nil,
		p,
		make(map[string]*ModuleScope),
		nil,
		virtual,
	}
	// Update records
	p.submodmap[submodule] = scope
	p.submodules = append(p.submodules, scope)
	// Done
	return true
}

// Binding returns information about the binding of a particular symbol defined
// in this module.
func (p *ModuleScope) Binding(name string, function bool) Binding {
	// construct binding identifier
	if bid, ok := p.ids[BindingId{name, function}]; ok {
		return p.bindings[bid]
	}
	// Failure
	return nil
}

// Bind looks up a given variable being referenced within a given module.  For a
// root context, this is either a column, an alias or a function declaration.
func (p *ModuleScope) Bind(symbol Symbol) bool {
	// Split the two cases: absolute versus relative.
	if symbol.Path().IsAbsolute() && p.parent != nil {
		// Absolute path, and this is not the root scope.  Therefore, simply
		// pass this up to the root scope for further processing.
		return p.parent.Bind(symbol)
	}
	// Relative path from this scope, or possibly an absolute path if this is
	// the root scope.
	found := p.innerBind(symbol.Path(), symbol)
	// If not found, traverse upwards.
	if !found && p.parent != nil {
		return p.parent.Bind(symbol)
	}
	//
	return found
}

// InnerBind is really a helper which allows us to split out the symbol path
// from the symbol itself.  This then lets us "traverse" the path as we go
// looking through submodules, etc.
func (p *ModuleScope) innerBind(path *util.Path, symbol Symbol) bool {
	// Relative path.  Then, either it refers to something in this scope, or
	// something in a subscope.
	if path.Depth() == 1 {
		// Must be something in this scope,.
		id := BindingId{symbol.Path().Tail(), symbol.IsFunction()}
		// Look for it.
		if bid, ok := p.ids[id]; ok {
			// Extract binding
			binding := p.bindings[bid]
			// Resolve symbol
			return symbol.Resolve(binding)
		}
	} else if submod, ok := p.submodmap[path.Head()]; ok {
		// Looks like this could be in the child scope, so continue searching there.
		return submod.innerBind(path.Dehead(), symbol)
	}
	// Otherwise, try traversing upwards.
	return false
}

// Enter returns a given submodule within this module.
func (p *ModuleScope) Enter(submodule string) *ModuleScope {
	if child, ok := p.submodmap[submodule]; ok {
		// Looks like this is in the child scope, so continue searching there.
		return child
	}
	// Should be unreachable.
	panic("unknown submodule")
}

// Alias constructs an alias for an existing symbol.  If the symbol does not
// exist, then this returns false.
func (p *ModuleScope) Alias(alias string, symbol Symbol) bool {
	// Sanity checks.  These are required for now, since we cannot alias
	// bindings in another scope at this time.
	if symbol.Path().IsAbsolute() || symbol.Path().Depth() != 1 {
		// This should be unreachable at the moment.
		panic(fmt.Sprintf("qualified aliases not supported %s", symbol.Path().String()))
	}
	// Extract symbol name
	name := symbol.Path().Head()
	// construct symbol identifier
	symbol_id := BindingId{name, symbol.IsFunction()}
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

// Define a new symbol within this scope.
func (p *ModuleScope) Define(symbol SymbolDefinition) bool {
	// Sanity checks
	if !symbol.Path().IsAbsolute() {
		// Definitely should be unreachable.
		panic("symbole definition cannot have relative path!")
	} else if !p.path.PrefixOf(*symbol.Path()) {
		// Should be unreachable.
		err := fmt.Sprintf("invalid symbol definition (%s not prefix of %s)", p.path.String(), symbol.Path().String())
		panic(err)
	} else if !symbol.Path().Parent().Equals(p.path) {
		name := symbol.Path().Get(p.path.Depth())
		// Looks like this definition is for a submodule.  Therefore, attempt to
		// find it and then define it there.
		if mod, ok := p.submodmap[name]; ok {
			// Found it, so attempt definition.
			return mod.Define(symbol)
		}
		// Failed
		return false
	}
	// construct binding identifier
	id := BindingId{symbol.Name(), symbol.IsFunction()}
	// Sanity check not already declared
	if bid, ok := p.ids[id]; ok && !symbol.IsFunction() {
		// Symbol already declared, and not a function.
		return false
	} else if ok {
		// Following must be true because we internally never attempt to
		// redeclare an intrinsic.
		def_binding := p.bindings[bid].(FunctionBinding)
		sym_binding := symbol.Binding().(*DefunBinding)
		// Attempt to overload the existing definition.
		if overloaded_binding, ok := def_binding.Overload(sym_binding); ok {
			// Success
			p.bindings[bid] = overloaded_binding
			return true
		}
		// Failed
		return false
	}
	// Symbol not previously declared, so no need to consider overloadings.
	bid := uint(len(p.bindings))
	p.bindings = append(p.bindings, symbol.Binding())
	p.ids[id] = bid
	//
	return true
}

// Flattern flatterns the tree into a flat array of modules, such that a module
// always comes before its own submodules.
func (p *ModuleScope) Flattern() []*ModuleScope {
	modules := []*ModuleScope{p}
	//
	for _, m := range p.submodules {
		modules = append(modules, m.Flattern()...)
	}
	//
	return modules
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

// Bind looks up a given variable or function being referenced either within the
// enclosing scope (module==nil) or within a specified module.
func (p LocalScope) Bind(symbol Symbol) bool {
	path := symbol.Path()
	// Determine whether this symbol could be a local variable or not.
	localVar := !symbol.IsFunction() && !path.IsAbsolute() && path.Depth() == 1
	// Check whether this is a local variable access.
	if id, ok := p.locals[path.Head()]; ok && localVar {
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
