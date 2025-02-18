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
package ast

import (
	"math"

	tr "github.com/consensys/go-corset/pkg/trace"
	"github.com/consensys/go-corset/pkg/util"
	"github.com/consensys/go-corset/pkg/util/sexp"
)

// Binding represents an association between a name, as found in a source file,
// and concrete item (e.g. a column, function, etc).
type Binding interface {
	// Determine whether this binding is finalised or not.
	IsFinalised() bool
}

// FunctionBinding is a special kind of binding which captures the essence of
// something which can be called.  For example, this could be a user-defined
// function or an intrinsic.
type FunctionBinding interface {
	Binding
	// IsPure checks whether this function binding has side-effects or not.
	IsPure() bool
	// IsNative checks whether this function binding is native (or not).
	IsNative() bool
	// HasArity checks whether this binding supports a given number of
	// parameters.  For example, intrinsic functions are often nary --- meaning
	// they can accept any number of arguments.  In contrast, a user-defined
	// function may only accept a specific number of arguments, etc.
	HasArity(uint) bool
	// Select corresponding signature based on arity.  If no matching signature
	// exists then this will return nil.
	Select(uint) *FunctionSignature
	// Overload (a.k.a specialise) this function binding to incorporate another
	// function signature.  This can fail for a few reasons: (1) some bindings
	// (e.g. intrinsics) cannot be overloaded; (2) duplicate overloadings are
	// not permitted; (3) combinding pure and impure overloadings is also not
	// permitted.
	Overload(*DefunBinding) (FunctionBinding, bool)
}

// FunctionSignature embodies a concrete function instance.  It is necessary to
// separate bindings from signatures because, in corset, function overloading is
// supported.  That is, we can have different definitions for a function of the
// same name and arity.  The appropriate definition is then selected for the
// given parameter types.
type FunctionSignature struct {
	// Pure or not
	pure bool
	// Parameter types for this function
	parameters []Type
	// Return type for this function
	ret Type
	// Body of this function
	body Expr
}

// NewFunctionSignature creates a new function signature with a given set of
// parameter/return types and a body.  The signature can also be identified as
// that of a pure function, or not.
func NewFunctionSignature(pure bool, parameters []Type, ret Type, body Expr) *FunctionSignature {
	return &FunctionSignature{pure, parameters, ret, body}
}

// IsPure checks whether this function binding has side-effects or not.
func (p *FunctionSignature) IsPure() bool {
	return p.pure
}

// Return the (optional) return type for this signature.  If no declared return
// type is given, then the intention is that it be inferred from the body.
func (p *FunctionSignature) Return() Type {
	return p.ret
}

// Parameter returns the given parameter in this signature.
func (p *FunctionSignature) Parameter(index uint) Type {
	return p.parameters[index]
}

// Arity returns the number of parameters in this signature.
func (p *FunctionSignature) Arity() uint {
	return uint(len(p.parameters))
}

// Apply a set of concreate arguments to this function.  This substitutes
// them through the body of the function producing a single expression.
func (p *FunctionSignature) Apply(args []Expr, srcmap *sexp.SourceMaps[Node]) Expr {
	mapping := make(map[uint]Expr)
	// Setup the mapping
	for i, e := range args {
		mapping[uint(i)] = e
	}
	// Substitute through
	return Substitute(p.body, mapping, srcmap)
}

// ============================================================================
// ColumnBinding
// ============================================================================

// ColumnBinding represents something bound to a given column.
type ColumnBinding struct {
	// Context determines the real (i.e. non-virtual) enclosing module of this
	// column, and should always be a prefix of the path.   If this column was
	// declared in a perspective then it will be the perspective's enclosing
	// module.  Otherwise, it will exactly match the path's parent.
	context util.Path
	// Absolute Path of column.  This determines the name of the column, its
	// enclosing module and/or perspective.
	Path util.Path
	// Column's datatype
	DataType Type
	// Determines whether this column must be proven (or not).
	MustProve bool
	// Column's length Multiplier
	Multiplier uint
	// Determines whether this is a Computed column, or not.
	Computed bool
	// Display modifier
	Display string
}

// AbsolutePath returns the fully resolved (absolute) path of the column in question.
func (p *ColumnBinding) AbsolutePath() *util.Path {
	return &p.Path
}

// IsFinalised checks whether this binding has been finalised yet or not.
func (p *ColumnBinding) IsFinalised() bool {
	return p.Multiplier != 0
}

// Finalise this binding by providing the necessary missing information.
func (p *ColumnBinding) Finalise(multiplier uint, datatype Type) {
	p.Multiplier = multiplier
	p.DataType = datatype
}

// Context returns the of this column.  That is, the module in which this colunm
// was declared and also the length multiplier of that module it requires.
func (p *ColumnBinding) Context() Context {
	return tr.NewContext(p.context.String(), p.Multiplier)
}

// ============================================================================
// ConstantBinding
// ============================================================================

// ConstantBinding represents a constant definition
type ConstantBinding struct {
	Path util.Path
	// Constant expression which, when evaluated, produces a constant Value.
	Value Expr
	// Determines whether or not this binding is finalised (i.e. its expression
	// has been resolved).
	finalised bool
}

// NewConstantBinding creates a new constant binding (which is initially not
// finalised).
func NewConstantBinding(path util.Path, value Expr) ConstantBinding {
	return ConstantBinding{path, value, false}
}

// IsFinalised checks whether this binding has been finalised yet or not.
func (p *ConstantBinding) IsFinalised() bool {
	return p.finalised
}

// Finalise this binding.
func (p *ConstantBinding) Finalise() {
	p.finalised = true
}

// Context returns the of this constant, noting that constants (by definition)
// do not have a context.
func (p *ConstantBinding) Context() Context {
	return tr.VoidContext[string]()
}

// ============================================================================
// ParameterBinding
// ============================================================================

// LocalVariableBinding represents something bound to a given column.
type LocalVariableBinding struct {
	// Name the local variable
	Name string
	// Type to use for this parameter.
	DataType Type
	// Identifies the variable or column Index (as appropriate).
	Index uint
}

// NewLocalVariableBinding constructs an (unitilalised) variable binding.  Being
// uninitialised means that its index identifier remains unknown.
func NewLocalVariableBinding(name string, datatype Type) LocalVariableBinding {
	return LocalVariableBinding{name, datatype, math.MaxUint}
}

// IsFinalised checks whether this binding has been finalised yet or not.
func (p *LocalVariableBinding) IsFinalised() bool {
	return p.Index != math.MaxUint
}

// Finalise this local variable binding by allocating it an identifier.
func (p *LocalVariableBinding) Finalise(index uint) {
	p.Index = index
}

// ============================================================================
// OverloadedBinding
// ============================================================================

// OverloadedBinding represents the amalgamation of two or more user-define
// function bindings.
type OverloadedBinding struct {
	pure bool
	// Available specialiases organised by arity.
	overloads []*DefunBinding
}

// IsPure checks whether this is a defpurefun or not
func (p *OverloadedBinding) IsPure() bool {
	return p.pure
}

// IsNative checks whether this function binding is native (or not).
func (p OverloadedBinding) IsNative() bool {
	// Cannot overload native
	return false
}

// IsFinalised checks whether this binding has been finalised yet or not.
func (p *OverloadedBinding) IsFinalised() bool {
	for _, binding := range p.overloads {
		if binding != nil && !binding.IsFinalised() {
			return false
		}
	}
	//
	return true
}

// HasArity checks whether this function accepts a given number of arguments (or
// not).
func (p *OverloadedBinding) HasArity(arity uint) bool {
	return arity < uint(len(p.overloads)) && p.overloads[arity] != nil
}

// Select corresponding signature based on arity.  If no matching signature
// exists then this will return nil.
func (p *OverloadedBinding) Select(arity uint) *FunctionSignature {
	if arity < uint(len(p.overloads)) && p.overloads[arity] != nil {
		signature := p.overloads[arity].Signature()
		return &signature
	}
	// failed
	return nil
}

// Overload (a.k.a specialise) this function binding to incorporate another
// function binding.  This can fail for a few reasons: (1) some bindings
// (e.g. intrinsics) cannot be overloaded; (2) duplicate overloadings are
// not permitted; (3) combinding pure and impure overloadings is also not
// permitted.
func (p *OverloadedBinding) Overload(overload *DefunBinding) (FunctionBinding, bool) {
	arity := len(overload.paramTypes)
	// Check matches purity
	if overload.IsPure() != p.pure {
		return nil, false
	}
	// ensure arity is defined
	for len(p.overloads) <= arity {
		p.overloads = append(p.overloads, nil)
	}
	// Check whether arity already defined
	if p.overloads[arity] != nil {
		return nil, false
	}
	// Nope, so define it
	p.overloads[arity] = overload
	// Done
	return p, true
}

// ============================================================================
// DefunBinding
// ============================================================================

// DefunBinding is a function binding arising from a user-defined function (as
// opposed, for example, to a function binding arising from an intrinsic).
type DefunBinding struct {
	// Flag whether or not is pure function
	pure bool
	// Types of parameters (optional)
	paramTypes []Type
	// Type of return (optional)
	returnType Type
	// Indicates whether this symbol is finalised (i.e. all expressions have
	// been resolved).
	finalised bool
	// body of the function in question.
	body Expr
}

var _ FunctionBinding = &DefunBinding{}

// NewDefunBinding constructs a new function binding.
func NewDefunBinding(pure bool, paramTypes []Type, returnType Type, body Expr) DefunBinding {
	return DefunBinding{pure, paramTypes, returnType, false, body}
}

// IsPure checks whether this is a defpurefun or not
func (p *DefunBinding) IsPure() bool {
	return p.pure
}

// IsNative checks whether this function binding is native (or not).
func (p *DefunBinding) IsNative() bool {
	return false
}

// IsFinalised checks whether this binding has been finalised yet or not.
func (p *DefunBinding) IsFinalised() bool {
	return p.finalised
}

// HasArity checks whether this function accepts a given number of arguments (or
// not).
func (p *DefunBinding) HasArity(arity uint) bool {
	return arity == uint(len(p.paramTypes))
}

// Signature returns the corresponding function signature for this user-defined
// function.
func (p *DefunBinding) Signature() FunctionSignature {
	return FunctionSignature{p.pure, p.paramTypes, p.returnType, p.body}
}

// Finalise this binding by providing the necessary missing information.
func (p *DefunBinding) Finalise() {
	p.finalised = true
}

// Select corresponding signature based on arity.  If no matching signature
// exists then this will return nil.
func (p *DefunBinding) Select(arity uint) *FunctionSignature {
	if arity == uint(len(p.paramTypes)) {
		return &FunctionSignature{p.pure, p.paramTypes, p.returnType, p.body}
	}
	// Ambiguous
	return nil
}

// Overload (a.k.a specialise) this function binding to incorporate another
// function binding.  This can fail for a few reasons: (1) some bindings
// (e.g. intrinsics) cannot be overloaded; (2) duplicate overloadings are
// not permitted; (3) combinding pure and impure overloadings is also not
// permitted.
func (p *DefunBinding) Overload(overload *DefunBinding) (FunctionBinding, bool) {
	var overloading = OverloadedBinding{p.IsPure(), nil}
	// Check it makes sense to do this.
	if p.IsPure() != overload.IsPure() {
		// Purity is misaligned
		return nil, false
	} else if len(p.paramTypes) == len(overload.paramTypes) {
		// Conflicting overlods
		return nil, false
	}
	// Looks good
	overloading.Overload(p)
	overloading.Overload(overload)
	//
	return &overloading, true
}

// ============================================================================
// Perspective
// ============================================================================

// PerspectiveBinding contains key information about a perspective, such as its
// selector expression.
type PerspectiveBinding struct {
	// Expression which determines when this perspective is enabled.
	Selector Expr
	// Indicates whether or not the selector has been finalised.
	resolved bool
}

var _ Binding = &PerspectiveBinding{}

// NewPerspectiveBinding constructs a new binding for a given perspective.
func NewPerspectiveBinding(selector Expr) *PerspectiveBinding {
	return &PerspectiveBinding{selector, false}
}

// IsFinalised checks whether this binding has been finalised yet or not.
func (p *PerspectiveBinding) IsFinalised() bool {
	return p.resolved
}

// Finalise this binding, which indicates the selector expression has been
// finalised.
func (p *PerspectiveBinding) Finalise() {
	p.resolved = true
}
