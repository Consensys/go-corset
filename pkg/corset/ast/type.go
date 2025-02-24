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
	"fmt"

	sc "github.com/consensys/go-corset/pkg/schema"
)

// Type embodies a richer notion of type found at the Corset level, compared
// with that found at lower levels (e.g. HIR). below.
type Type interface {
	// Determines whether or not this type supports "loobean" semantics.  If so,
	// this means that 0 is treated as true, with anything else being false.
	HasLoobeanSemantics() bool

	// Determines whether or not this type supports "boolean" semantics.  If so,
	// this means that 0 is treated as false, with anything else being true.
	HasBooleanSemantics() bool

	// Construct a variant of this type which employs loobean semantics.  This
	// will panic if the type has already been given boolean semantics.
	WithLoobeanSemantics() Type

	// Construct a variant of this type which employs boolean semantics.  This
	// will panic if the type has already been given loobean semantics.
	WithBooleanSemantics() Type

	// SubtypeOf determines whether or not this type is a subtype of another.
	SubtypeOf(Type) bool

	// ContainsFieldType determines whether this type contains a native field type.
	ContainsFieldType() bool

	// Access an underlying representation of this type (should one exist).  If
	// this doesn't exist, then nil is returned.
	AsUnderlying() sc.Type

	// Returns the number of underlying columns represented by this column.  For
	// example, an array of size n will expand into n underlying columns.
	Width() uint

	// Produce a string representation of this type.
	String() string
}

// NewFieldType constructs a native field type which, initially, has no semantic
// specified.
func NewFieldType() Type {
	return &NativeType{&sc.FieldType{}, false, false}
}

// NewUintType constructs a native uint type of the given width which,
// initially, has no semantic specified.
func NewUintType(nbits uint) Type {
	return &NativeType{sc.NewUintType(nbits), false, false}
}

// GreatestLowerBoundAll joins zero or more types together using the GLB
// operator.
func GreatestLowerBoundAll(types []Type) Type {
	var datatype Type
	//
	for _, t := range types {
		if datatype == nil {
			datatype = t
		} else if t != nil {
			datatype = GreatestLowerBound(datatype, t)
		}
	}
	//
	return datatype
}

// GreatestLowerBound computes the Greatest Lower Bound of two types.  For
// example, the lub of u16 and u128 is u128, etc.  This means that, when joining
// the bottom type with a type that has semantics, you get the former.
func GreatestLowerBound(lhs Type, rhs Type) Type {
	// Sanity checks
	if lhs == nil || rhs == nil {
		return nil
	}
	// Proceed
	var (
		l_loobean bool = lhs.HasLoobeanSemantics()
		r_loobean bool = rhs.HasLoobeanSemantics()
		l_boolean bool = lhs.HasBooleanSemantics()
		r_boolean bool = rhs.HasBooleanSemantics()
	)
	// Determine join of underlying types
	underlying := sc.Join(lhs.AsUnderlying(), rhs.AsUnderlying())
	//
	return &NativeType{underlying, l_loobean && r_loobean, l_boolean && r_boolean}
}

// LeastUpperBoundAll joins zero or more types together using the LUB operator.
func LeastUpperBoundAll(types []Type) Type {
	var datatype Type
	//
	for _, t := range types {
		if datatype == nil {
			datatype = t
		} else if t != nil {
			datatype = LeastUpperBound(datatype, t)
		}
	}
	//
	return datatype
}

// LeastUpperBound computes the Least Upper Bound of two types.  For example,
// the lub of u16 and u128 is u128, etc.    This means that, when joining the
// bottom type with a type that has semantics, you get the latter.
func LeastUpperBound(lhs Type, rhs Type) Type {
	// Sanity checks
	if lhs == nil || rhs == nil {
		return nil
	}
	// Proceed
	var (
		l_loobean bool = lhs.HasLoobeanSemantics()
		r_loobean bool = rhs.HasLoobeanSemantics()
		l_boolean bool = lhs.HasBooleanSemantics()
		r_boolean bool = rhs.HasBooleanSemantics()
	)
	// Determine join of underlying types
	underlying := sc.Join(lhs.AsUnderlying(), rhs.AsUnderlying())
	//
	return &NativeType{underlying, l_loobean || r_loobean, l_boolean || r_boolean}
}

// ============================================================================
// NativeType
// ============================================================================

// NativeType simply wraps one of the types available at the HIR level (and below).
type NativeType struct {
	// The underlying type
	datatype sc.Type
	// Determines whether or not this type supports "loobean" semantics.  If so,
	// this means that 0 is treated as true, with anything else being false.
	loobean bool
	// Determines whether or not this type supports "boolean" semantics.  If so,
	// this means that 0 is treated as false, with anything else being true.
	boolean bool
}

// HasLoobeanSemantics indicates whether or not this type supports "loobean"
// semantics or not. If so, this means that 0 is treated as true, with anything
// else being false.
func (p *NativeType) HasLoobeanSemantics() bool {
	return p.loobean && !p.boolean
}

// HasBooleanSemantics indicates whether or not this type supports "boolean"
// semantics. If so, this means that 0 is treated as false, with anything else
// being true.
func (p *NativeType) HasBooleanSemantics() bool {
	return p.boolean && !p.loobean
}

// WithLoobeanSemantics constructs a variant of this type which employs loobean
// semantics.  This will panic if the type has already been given boolean
// semantics.
func (p *NativeType) WithLoobeanSemantics() Type {
	if p.HasBooleanSemantics() {
		panic("type already given boolean semantics")
	}
	// Done
	return &NativeType{p.datatype, true, false}
}

// WithBooleanSemantics constructs a variant of this type which employs boolean
// semantics.  This will panic if the type has already been given boolean
// semantics.
func (p *NativeType) WithBooleanSemantics() Type {
	if p.HasLoobeanSemantics() {
		panic("type already given loobean semantics")
	}
	// Done
	return &NativeType{p.datatype, false, true}
}

// ContainsFieldType indicates indicates whether or not this type contains(is=) a field type
func (p *NativeType) ContainsFieldType() bool {
	_, ok := p.AsUnderlying().(*sc.FieldType)
	return ok
}

// Width returns the number of underlying columns represented by this column.
// For example, an array of size n will expand into n underlying columns.
func (p *NativeType) Width() uint {
	return 1
}

// AsUnderlying attempts to convert this type into an underlying type.  If this
// is not possible, then nil is returned.
func (p *NativeType) AsUnderlying() sc.Type {
	return p.datatype
}

// SubtypeOf determines whether or not this type is a subtype of another.
func (p *NativeType) SubtypeOf(other Type) bool {
	if o, ok := other.(*NativeType); ok && p.datatype.SubtypeOf(o.datatype) {
		// An interpreted type can flow into an uninterpreted type.
		return (!o.loobean && !o.boolean) || p == o
	}
	//
	return false
}

func (p *NativeType) String() string {
	if p.loobean {
		return fmt.Sprintf("%s@loob", p.datatype.String())
	} else if p.boolean {
		return fmt.Sprintf("%s@bool", p.datatype.String())
	}
	//
	return p.datatype.String()
}

// ============================================================================
// ArrayType
// ============================================================================

// ArrayType represents a statically-sized array of types.
type ArrayType struct {
	// element type
	element Type
	// min index
	min uint
	// max index
	max uint
}

// NewArrayType constructs a new array type of a given (fixed) size.
func NewArrayType(element Type, min uint, max uint) *ArrayType {
	return &ArrayType{element, min, max}
}

// HasLoobeanSemantics indicates whether or not this type supports "loobean"
// semantics or not. If so, this means that 0 is treated as true, with anything
// else being false.
func (p *ArrayType) HasLoobeanSemantics() bool {
	return false
}

// HasBooleanSemantics indicates whether or not this type supports "boolean"
// semantics. If so, this means that 0 is treated as false, with anything else
// being true.
func (p *ArrayType) HasBooleanSemantics() bool {
	return false
}

// WithLoobeanSemantics constructs a variant of this type which employs loobean
// semantics.  This will panic if the type has already been given boolean
// semantics.
func (p *ArrayType) WithLoobeanSemantics() Type {
	panic("unreachable")
}

// WithBooleanSemantics constructs a variant of this type which employs boolean
// semantics.  This will panic if the type has already been given boolean
// semantics.
func (p *ArrayType) WithBooleanSemantics() Type {
	panic("unreachable")
}

// ContainsFieldType indicates indicates whether or not this array type contains a field type
func (p *ArrayType) ContainsFieldType() bool {
	for i := p.min; i <= p.max; i++ {
		if p.element.ContainsFieldType() {
			return true
		}
	}
	return false
}

// Width returns the number of underlying columns represented by this column.
// For example, an array of size n will expand into n underlying columns.
func (p *ArrayType) Width() uint {
	return p.max - p.min + 1
}

// AsUnderlying attempts to convert this type into an underlying type.  If this
// is not possible, then nil is returned.
func (p *ArrayType) AsUnderlying() sc.Type {
	return nil
}

// Element returns the element of this array type.
func (p *ArrayType) Element() Type {
	return p.element
}

// MinIndex returns the smallest index of elements in this array type.
func (p *ArrayType) MinIndex() uint {
	return p.min
}

// MaxIndex returns the largest index of elements in this array type.
func (p *ArrayType) MaxIndex() uint {
	return p.max
}

// SubtypeOf determines whether or not this type is a subtype of another.
func (p *ArrayType) SubtypeOf(other Type) bool {
	if o, ok := other.(*ArrayType); ok {
		return p.element.SubtypeOf(o.element)
	}
	//
	return false
}

func (p *ArrayType) String() string {
	return fmt.Sprintf("(%s)[%d:%d]", p.element.String(), p.min, p.max)
}
