package corset

import (
	"math"
	"reflect"

	tr "github.com/consensys/go-corset/pkg/trace"
)

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
	// HasArity checks whether this binding supports a given number of
	// parameters.  For example, intrinsic functions are often nary --- meaning
	// they can accept any number of arguments.  In contrast, a user-defined
	// function may only accept a specific number of arguments, etc.
	HasArity(uint) bool
	// Select the best fit signature based on the available parameter types.
	// Observe that, for valid arities, this always returns a signature.
	// However, that signature may not actually accept the provided parameters
	// (in which case, an error should be reported).  Furthermore, if no
	// appropriate signature exists then this will return nil.
	Select([]Type) *FunctionSignature
	// Overload (a.k.a specialise) this function binding to incorporate another
	// function binding.  This can fail for a few reasons: (1) some bindings
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

// IsPure checks whether this function binding has side-effects or not.
func (p *FunctionSignature) IsPure() bool {
	return p.pure
}

// Accepts check whether a given set of concrete argument types can be accepted
// by this signature.
func (p *FunctionSignature) Accepts(args []Type) bool {
	if len(args) != len(p.parameters) {
		return false
	}
	// Check argument at each position is accepted by parameter at that
	// position.
	for i := 0; i < len(args); i++ {
		arg_t := args[i]
		param_t := p.parameters[i]
		//
		if !arg_t.SubtypeOf(param_t) {
			return false
		}
	}
	// Done
	return true
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

// NumParameters returns the number of parameters in this signature.
func (p *FunctionSignature) NumParameters() uint {
	return uint(len(p.parameters))
}

// SubtypeOf determines whether this is a stronger specialisation than another.
func (p *FunctionSignature) SubtypeOf(other *FunctionSignature) bool {
	if len(p.parameters) != len(other.parameters) {
		return false
	}
	//
	for i := 0; i < len(p.parameters); i++ {
		pth := p.parameters[i]
		oth := other.parameters[i]
		// Check them
		if !pth.SubtypeOf(oth) {
			return false
		}
	}
	//
	return true
}

// Apply a set of concreate arguments to this function.  This substitutes
// them through the body of the function producing a single expression.
func (p *FunctionSignature) Apply(args []Expr) Expr {
	mapping := make(map[uint]Expr)
	// Setup the mapping
	for i, e := range args {
		mapping[uint(i)] = e
	}
	// Substitute through
	return p.body.Substitute(mapping)
}

// ============================================================================
// ColumnBinding
// ============================================================================

// ColumnBinding represents something bound to a given column.
type ColumnBinding struct {
	// Column's allocated identifier
	cid uint
	// Column's enclosing module
	module string
	// Determines whether this is a computed column, or not.
	computed bool
	// Determines whether this column must be proven (or not).
	mustProve bool
	// Column's length multiplier
	multiplier uint
	// Column's datatype
	dataType Type
}

// NewInputColumnBinding constructs a new column binding in a given module.
// This is for the case where all information about the column is already known,
// and will not be inferred from elsewhere.
func NewInputColumnBinding(module string, mustProve bool, multiplier uint, datatype Type) *ColumnBinding {
	return &ColumnBinding{math.MaxUint, module, false, mustProve, multiplier, datatype}
}

// NewComputedColumnBinding constructs a new column binding in a given
// module.  This is for the case where not all information is yet known about
// the column and, hence, it must be finalised later on.  For example, in a
// definterleaved constraint the target column information (e.g. its type) is
// not immediately available and must be determined from those columns from
// which it is constructed.
func NewComputedColumnBinding(module string) *ColumnBinding {
	return &ColumnBinding{math.MaxUint, module, true, false, 0, nil}
}

// IsFinalised checks whether this binding has been finalised yet or not.
func (p *ColumnBinding) IsFinalised() bool {
	return p.multiplier != 0
}

// Finalise this binding by providing the necessary missing information.
func (p *ColumnBinding) Finalise(multiplier uint, datatype Type) {
	p.multiplier = multiplier
	p.dataType = datatype
}

// Context returns the of this column.  That is, the module in which this colunm
// was declared and also the length multiplier of that module it requires.
func (p *ColumnBinding) Context() Context {
	return tr.NewContext(p.module, p.multiplier)
}

// AllocateId allocates the column identifier for this column
func (p *ColumnBinding) AllocateId(cid uint) {
	p.cid = cid
}

// ColumnId returns the allocated identifier for this column.  NOTE: this will
// panic if this column has not yet been allocated an identifier.
func (p *ColumnBinding) ColumnId() uint {
	if p.cid == math.MaxUint {
		panic("column id not yet allocated")
	}
	//
	return p.cid
}

// ============================================================================
// ConstantBinding
// ============================================================================

// ConstantBinding represents a constant definition
type ConstantBinding struct {
	// Constant expression which, when evaluated, produces a constant value.
	value Expr
	// Inferred type of the given expression
	datatype Type
}

// NewConstantBinding creates a new constant binding (which is initially not
// finalised).
func NewConstantBinding(value Expr) ConstantBinding {
	return ConstantBinding{value, nil}
}

// IsFinalised checks whether this binding has been finalised yet or not.
func (p *ConstantBinding) IsFinalised() bool {
	return p.datatype != nil
}

// Finalise this binding by providing the necessary missing information.
func (p *ConstantBinding) Finalise(datatype Type) {
	p.datatype = datatype
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
	name string
	// Type to use for this parameter.
	datatype Type
	// Identifies the variable or column index (as appropriate).
	index uint
}

// NewLocalVariableBinding constructs an (unitilalised) variable binding.  Being
// uninitialised means that its index identifier remains unknown.
func NewLocalVariableBinding(name string, datatype Type) LocalVariableBinding {
	return LocalVariableBinding{name, datatype, math.MaxUint}
}

// IsFinalised checks whether this binding has been finalised yet or not.
func (p *LocalVariableBinding) IsFinalised() bool {
	return p.index != math.MaxUint
}

// Finalise this local variable binding by allocating it an identifier.
func (p *LocalVariableBinding) Finalise(index uint) {
	p.index = index
}

// ============================================================================
// OverloadedBinding
// ============================================================================

// OverloadedBinding represents the amalgamation of two or more user-define
// function bindings.
type OverloadedBinding struct {
	// Available specialisations
	overloads []*DefunBinding
}

// IsPure checks whether this is a defpurefun or not
func (p *OverloadedBinding) IsPure() bool {
	return p.overloads[0].IsPure()
}

// IsFinalised checks whether this binding has been finalised yet or not.
func (p *OverloadedBinding) IsFinalised() bool {
	for _, binding := range p.overloads {
		if !binding.IsFinalised() {
			return false
		}
	}
	//
	return true
}

// HasArity checks whether this function accepts a given number of arguments (or
// not).
func (p *OverloadedBinding) HasArity(arity uint) bool {
	for _, binding := range p.overloads {
		if binding.HasArity(arity) {
			// match
			return true
		}
	}
	//
	return false
}

// Select the best fit signature based on the available parameter types.
// Observe that, for valid arities, this always returns a signature.
// However, that signature may not actually accept the provided parameters
// (in which case, an error should be reported).  Furthermore, if no
// appropriate signature exists then this will return nil.
func (p *OverloadedBinding) Select(args []Type) *FunctionSignature {
	var selected *FunctionSignature
	// Attempt to select the Greated Lower Bound (GLB).  This can fail if there
	// is no unique GLB.
	for _, binding := range p.overloads {
		// Extract its function signature
		sig := binding.Signature()
		// Check whether its applicable to the given argument types.
		applicable := sig.Accepts(args)
		// If it is applicable, then update the current selection as necessary.
		if applicable && selected == nil {
			selected = &sig
		} else if applicable && sig.SubtypeOf(selected) {
			// Signature is better specialisation than that currently selected.
			selected = &sig
		} else if applicable && !selected.SubtypeOf(&sig) {
			// Ambiguous, so give up.
			return nil
		}
	}
	//
	return selected
}

// Overload (a.k.a specialise) this function binding to incorporate another
// function binding.  This can fail for a few reasons: (1) some bindings
// (e.g. intrinsics) cannot be overloaded; (2) duplicate overloadings are
// not permitted; (3) combinding pure and impure overloadings is also not
// permitted.
func (p *OverloadedBinding) Overload(overload *DefunBinding) (FunctionBinding, bool) {
	// Check matches purity
	if overload.IsPure() != p.IsPure() {
		return nil, false
	}
	// Check overload does not already exist
	for _, binding := range p.overloads {
		if reflect.DeepEqual(binding.paramTypes, overload.paramTypes) {
			// Already declared
			return nil, false
		}
	}
	// Otherwise, looks good.
	p.overloads = append(p.overloads, overload)
	//
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
	// Inferred type of the body.  This is used to compare against the declared
	// type (if there is one) to check for any descrepencies.
	bodyType Type
	// body of the function in question.
	body Expr
}

var _ FunctionBinding = &DefunBinding{}

// NewDefunBinding constructs a new function binding.
func NewDefunBinding(pure bool, paramTypes []Type, returnType Type, body Expr) DefunBinding {
	return DefunBinding{pure, paramTypes, returnType, nil, body}
}

// IsPure checks whether this is a defpurefun or not
func (p *DefunBinding) IsPure() bool {
	return p.pure
}

// IsFinalised checks whether this binding has been finalised yet or not.
func (p *DefunBinding) IsFinalised() bool {
	return p.bodyType != nil
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
func (p *DefunBinding) Finalise(bodyType Type) {
	p.bodyType = bodyType
}

// Select the best fit signature based on the available parameter types.
// Observe that, for valid arities, this always returns a signature.
// However, that signature may not actually accept the provided parameters
// (in which case, an error should be reported).  Furthermore, if no
// appropriate signature exists then this will return nil.
func (p *DefunBinding) Select(args []Type) *FunctionSignature {
	if len(args) == len(p.paramTypes) {
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
	if p.IsPure() != overload.IsPure() {
		// Purity is misaligned
		return nil, false
	} else if reflect.DeepEqual(p.paramTypes, overload.paramTypes) {
		// Specialisation already exists!
		return nil, false
	}
	//
	return &OverloadedBinding{[]*DefunBinding{p, overload}}, true
}
