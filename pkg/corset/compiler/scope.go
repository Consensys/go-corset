// Copyright Consensys Software Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License"); you may not use this file except in compliance with
// the License. You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
// an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
// specific language governing permissions and limitations under the License.
//
// SPDX-License-Identifier: Apache-2.0
package compiler

import (
	"fmt"

	"github.com/consensys/go-corset/pkg/corset/ast"
	sc "github.com/consensys/go-corset/pkg/schema"
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
	Bind(ast.Symbol) bool
}

// BindingId is an identifier is used to distinguish different forms of binding,
// as some forms are known from their use.  Specifically, at the current time,
// only functions are distinguished from other categories (e.g. columns,
// parameters, etc).
type BindingId struct {
	// Name of the binding
	name string
	// Indicates whether function binding (or not) and, if so, what arity the
	// function has.
	arity util.Option[uint]
}

// IsFunction checks whether or not this binding identifier refers to a function
// definition or not.
func (b BindingId) IsFunction() bool {
	return b.arity.HasValue()
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
	// Selector determining when this module is active.
	selector ast.Expr
	// Absolute path
	path util.Path
	// Map identifiers to indices within the bindings array.
	ids map[BindingId]uint
	// The set of bindings in the order of declaration.
	bindings []ast.Binding
	// Enclosing scope
	parent *ModuleScope
	// Submodules in a map (for efficient lookup)
	submodmap map[string]*ModuleScope
	// Submodules in the order of declaration (for determinism).
	submodules []*ModuleScope
}

// NewModuleScope constructs an initially empty top-level scope.
func NewModuleScope(selector ast.Expr) *ModuleScope {
	return &ModuleScope{
		selector,
		util.NewAbsolutePath(),
		make(map[BindingId]uint),
		nil,
		nil,
		make(map[string]*ModuleScope),
		nil,
	}
}

// Path returns the absolute path of this module.
func (p *ModuleScope) Path() *util.Path {
	return &p.path
}

// Name returns the name of the given module.
func (p *ModuleScope) Name() string {
	if p.path.Depth() > 0 {
		return p.path.Tail()
	}
	//
	return ""
}

// Virtual identifies whether or not this is a virtual module.
func (p *ModuleScope) Virtual() bool {
	return p.selector != nil
}

// IsRoot checks whether or not this is the root of the module tree.
func (p *ModuleScope) IsRoot() bool {
	return p.parent == nil
}

// Children returns the set of submodules defined within this module.
func (p *ModuleScope) Children() []*ModuleScope {
	return p.submodules
}

// Selector gets an HIR unit expression which evaluates to a non-zero value when
// this module is active.  This can be nil if there is no selector (i.e. this is
// a non-virtual module).
func (p *ModuleScope) Selector() ast.Expr {
	return p.selector
}

// DestructuredColumns returns the set of (destructured) columns defined within
// this module scope.  That is, source-level columns which are broken down into
// their atomic components.
func (p *ModuleScope) DestructuredColumns() []RegisterSource {
	var (
		sources []RegisterSource
		owner   util.Path = p.Owner().path
	)
	//
	for _, b := range p.bindings {
		if binding, ok := b.(*ast.ColumnBinding); ok {
			cols := p.destructureColumn(binding, owner, binding.Path, binding.DataType)
			sources = append(sources, cols...)
		}
	}
	//
	return sources
}

// DestructuredConstants returns the set of (destructured) constant definitions
// within this module scope.
func (p *ModuleScope) DestructuredConstants() []ast.ConstantBinding {
	var constants []ast.ConstantBinding

	for _, b := range p.bindings {
		if binding, ok := b.(*ast.ConstantBinding); ok {
			constants = append(constants, *binding)
		}
	}

	return constants
}

// Owner returns the enclosing non-virtual module of this module.  Observe
// that, if this is a non-virtual module, then it is returned.
func (p *ModuleScope) Owner() *ModuleScope {
	if p.selector == nil {
		return p
	} else if p.parent != nil {
		return p.parent.Owner()
	}
	// Should be unreachable
	panic("invalid module tree")
}

// Declare a new submodule at the given (absolute) path within this tree scope.
// Submodules can be declared as "virtual" which indicates the submodule is
// simply a subset of rows of its enclosing module.  A virtual module is
// indicated by a non-zero selector, which signals when the virtual module is
// active.  This returns true if this succeeds, otherwise returns false (i.e. a
// matching submodule already exists).
func (p *ModuleScope) Declare(submodule string, selector ast.Expr) bool {
	if _, ok := p.submodmap[submodule]; ok {
		return false
	}
	// Construct suitable child scope
	scope := &ModuleScope{
		selector,
		*p.path.Extend(submodule),
		make(map[BindingId]uint),
		nil,
		p,
		make(map[string]*ModuleScope),
		nil,
	}
	// Update records
	p.submodmap[submodule] = scope
	p.submodules = append(p.submodules, scope)
	// Done
	return true
}

// Binding returns information about the binding of a particular symbol defined
// in this module.
func (p *ModuleScope) Binding(name string, arity util.Option[uint]) ast.Binding {
	// construct binding identifier
	if bid, ok := p.ids[BindingId{name, arity}]; ok {
		return p.bindings[bid]
	}
	// Failure
	return nil
}

// Bind looks up a given variable being referenced within a given module.  For a
// root context, this is either a column, an alias or a function declaration.
func (p *ModuleScope) Bind(symbol ast.Symbol) bool {
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
func (p *ModuleScope) innerBind(path *util.Path, symbol ast.Symbol) bool {
	// Relative path.  Then, either it refers to something in this scope, or
	// something in a subscope.
	if path.Depth() == 1 {
		// Must be something in this scope,.
		id := BindingId{symbol.Path().Tail(), symbol.Arity()}
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
func (p *ModuleScope) Alias(alias string, symbol ast.Symbol) bool {
	// Sanity checks.  These are required for now, since we cannot alias
	// bindings in another scope at this time.
	if symbol.Path().IsAbsolute() || symbol.Path().Depth() != 1 {
		// This should be unreachable at the moment.
		panic(fmt.Sprintf("qualified aliases not supported %s", symbol.Path().String()))
	}
	// Extract symbol name
	name := symbol.Path().Head()
	// construct symbol identifier
	symbol_id := BindingId{name, symbol.Arity()}
	// construct alias identifier
	alias_id := BindingId{alias, symbol.Arity()}
	// Check alias does not already exist
	if _, ok := p.ids[alias_id]; !ok {
		// Check symbol being aliased exists
		if id, ok := p.ids[symbol_id]; ok {
			p.ids[alias_id] = id
			// Done
			return true
		}
	}
	// ast.Symbol not known (yet)
	return false
}

// Define a new symbol within this scope.
func (p *ModuleScope) Define(symbol ast.SymbolDefinition) bool {
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
	id := BindingId{symbol.Name(), symbol.Arity()}
	// Sanity check not already declared
	if _, ok := p.ids[id]; ok {
		// Yes, already declared.
		return false
	}
	// ast.Symbol not previously declared, so no need to consider overloadings.
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

func (p *ModuleScope) destructureColumn(column *ast.ColumnBinding, ctx util.Path, path util.Path,
	datatype ast.Type) []RegisterSource {
	// Check for base base
	if int_t, ok := datatype.(*ast.IntType); ok {
		return p.destructureAtomicColumn(column, ctx, path, int_t.AsUnderlying())
	} else if arraytype, ok := datatype.(*ast.ArrayType); ok {
		// For now, assume must be an array
		return p.destructureArrayColumn(column, ctx, path, arraytype)
	} else {
		panic(fmt.Sprintf("unknown type encountered: %v", datatype))
	}
}

// Allocate an array type
func (p *ModuleScope) destructureArrayColumn(col *ast.ColumnBinding, ctx util.Path, path util.Path,
	arrtype *ast.ArrayType) []RegisterSource {
	//
	var sources []RegisterSource
	// Allocate n columns
	for i := arrtype.MinIndex(); i <= arrtype.MaxIndex(); i++ {
		ith_name := fmt.Sprintf("%s_%d", path.Tail(), i)
		ith_path := path.Parent().Extend(ith_name)
		sources = append(sources, p.destructureColumn(col, ctx, *ith_path, arrtype.Element())...)
	}
	//
	return sources
}

// Destructure atomic column
func (p *ModuleScope) destructureAtomicColumn(column *ast.ColumnBinding, ctx util.Path, path util.Path,
	datatype sc.Type) []RegisterSource {
	// Construct register source.
	source := RegisterSource{
		ctx,
		path,
		column.Multiplier,
		datatype,
		column.MustProve,
		column.Computed,
		column.Display}
	//
	return []RegisterSource{source}
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
	// Determine whether or not this scope is defining a constant.  If so, then
	// cannot access other externalised constants.
	constant bool
	// Represents the enclosing scope
	enclosing Scope
	// Context for this scope
	context *ast.Context
	// Maps inputs parameters to the declaration index.
	locals map[string]uint
	// Actual parameter bindings
	bindings []*ast.LocalVariableBinding
}

// NewLocalScope constructs a new local scope within a given enclosing scope.  A
// local scope can have local variables declared within it.  A local scope can
// also be "global" in the sense that accessing symbols from other modules is
// permitted.
func NewLocalScope(enclosing Scope, global bool, pure bool, constant bool) LocalScope {
	context := tr.VoidContext[string]()
	locals := make(map[string]uint)
	bindings := make([]*ast.LocalVariableBinding, 0)
	//
	return LocalScope{global, pure, constant, enclosing, &context, locals, bindings}
}

// NestedScope creates a nested scope within this local scope.
func (p LocalScope) NestedScope() LocalScope {
	nlocals := make(map[string]uint)
	nbindings := make([]*ast.LocalVariableBinding, len(p.bindings))
	// Clone allocated variables
	for k, v := range p.locals {
		nlocals[k] = v
	}
	// Copy over bindings.
	copy(nbindings, p.bindings)
	// Done
	return LocalScope{p.global, p.pure, p.constant, p, p.context, nlocals, nbindings}
}

// NestedConstScope creates a nested scope within this local scope which, in
// addition, is always pure and expects a constant value.
func (p LocalScope) NestedConstScope() LocalScope {
	nlocals := make(map[string]uint)
	nbindings := make([]*ast.LocalVariableBinding, len(p.bindings))
	// Clone allocated variables
	for k, v := range p.locals {
		nlocals[k] = v
	}
	// Copy over bindings.
	copy(nbindings, p.bindings)
	// Done
	return LocalScope{p.global, true, true, p, p.context, nlocals, nbindings}
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

// IsConstant determines whether or not this scope is defining a constant.  This
// places some restrictions on what variables can be accessed, etc.
func (p LocalScope) IsConstant() bool {
	return p.constant
}

// FixContext fixes the context for this scope.  Since every scope requires
// exactly one context, this fails if we fix it to incompatible contexts.
func (p LocalScope) FixContext(context ast.Context) bool {
	// Join contexts together
	*p.context = p.context.Join(context)
	// Check they were compatible
	return !p.context.IsConflicted()
}

// Bind looks up a given variable or function being referenced either within the
// enclosing scope (module==nil) or within a specified module.
func (p LocalScope) Bind(symbol ast.Symbol) bool {
	path := symbol.Path()
	// Determine whether this symbol could be a local variable or not.
	localVar := symbol.Arity().IsEmpty() && !path.IsAbsolute() && path.Depth() == 1
	// Check whether this is a local variable access.
	if id, ok := p.locals[path.Head()]; ok && localVar {
		// Yes, this is a local variable access.
		return symbol.Resolve(p.bindings[id])
	}
	// No, this is not a local variable access.
	return p.enclosing.Bind(symbol)
}

// DeclareLocal registers a new local variable (e.g. a parameter).
func (p *LocalScope) DeclareLocal(name string, binding *ast.LocalVariableBinding) uint {
	index := uint(len(p.locals))
	binding.Finalise(index)
	p.locals[name] = index
	p.bindings = append(p.bindings, binding)
	// Return variable index
	return index
}
