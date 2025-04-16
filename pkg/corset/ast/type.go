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
	"math/big"

	"github.com/consensys/go-corset/pkg/schema"
	"github.com/consensys/go-corset/pkg/util"
)

// Type embodies a richer notion of type found at the Corset level, compared
// with that found at lower levels (e.g. HIR). below.
type Type interface {
	// SubtypeOf determines whether or not this type is a subtype of another.
	SubtypeOf(Type) bool

	// LeastUpperBound computes the least upper bound of this type and another. This
	// is smallest type which contains both of the arguments.  For example, i32 is
	// the least upper bound of i1 and i32, etc.  If no such type exists, then nil
	// is returned.
	LeastUpperBound(Type) Type

	// Returns the number of underlying columns represented by this column.  For
	// example, an array of size n will expand into n underlying columns.
	Width() uint

	// Determines whether or not this type has an underlying representation, or
	// not.
	HasUnderlying() bool

	// Produce a string representation of this type.
	String() string
}

// LeastUpperBound computes the Least Upper Bound of two types.  This is
// deliberately coarse-grained and does not, for example, attempt to perform any
// kind of range analysis for integer types (as this would not make sense).
func LeastUpperBound(types ...Type) Type {
	var datatype Type
	//
	for i, t := range types {
		if i == 0 {
			datatype = t
		} else {
			datatype = datatype.LeastUpperBound(t)
		}
		// sanity check
		if t == nil {
			return nil
		}
	}
	//
	return datatype
}

// ============================================================================
// AnyType
// ============================================================================

// AnyType represents the top of the type lattice.  This is the type into which
// values of any other type can flow.
type AnyType struct{}

// ANY_TYPE represents the top of the lattice.  That is, any type can flow into
// any.
var ANY_TYPE Type = nil

// HasUnderlying determines whether or not this type has an underlying
// representation, or not.
func (p *AnyType) HasUnderlying() bool {
	return false
}

// AsUnderlying converts this integer type into an underlying type.
func (p *AnyType) AsUnderlying() schema.Type {
	panic("cannot convert any type")
}

// Width returns the number of underlying columns represented by this column.
// For example, an array of size n will expand into n underlying columns.
func (p *AnyType) Width() uint {
	return 0
}

// LeastUpperBound computes the least upper bound of this type and another. This
// is smallest type which contains both of the arguments.  For example, i32 is
// the least upper bound of i1 and i32, etc.  If no such type exists, then nil
// is returned.
func (p *AnyType) LeastUpperBound(other Type) Type {
	return p
}

// SubtypeOf determines whether or not this type is a subtype of another.
func (p *AnyType) SubtypeOf(other Type) bool {
	if other == nil {
		return true
	} else if _, ok := other.(*AnyType); ok {
		return true
	}
	//
	return false
}

func (p *AnyType) String() string {
	return "any"
}

// ============================================================================
// IntType
// ============================================================================

// INT_TYPE represents the infinite integer range.  This cannot be translated
// into a concrete type at the lower level, and therefore can only be used
// internally (e.g. for type checking).
var INT_TYPE = &IntType{nil}

// IntType represents a set of signed integer values.
type IntType struct {
	values *util.Interval
}

// NewUintType constructs a native uint type of the given width which,
// initially, has no semantic specified.
func NewUintType(nbits uint) Type {
	bound := big.NewInt(2)
	bound.Exp(bound, big.NewInt(int64(nbits)), nil)
	// Subtract 1 because interval is inclusive.
	bound.Sub(bound, big.NewInt(1))
	//
	return &IntType{util.NewInterval(big.NewInt(0), bound)}
}

// NewIntType constructs a new integer type containing all values between the
// lower and upper bounds (inclusive).
func NewIntType(lower *big.Int, upper *big.Int) *IntType {
	return &IntType{util.NewInterval(lower, upper)}
}

// HasUnderlying determines whether or not this type has an underlying
// representation, or not.
func (p *IntType) HasUnderlying() bool {
	return p.values != nil
}

// AsUnderlying converts this integer type into an underlying type.
func (p *IntType) AsUnderlying() schema.Type {
	width := p.values.BitWidth()
	// Sanity check (for now)
	if p.values.Contains(big.NewInt(-1)) {
		panic("cannot convert signed integer type")
	}
	//
	return schema.NewUintType(width)
}

// Width returns the number of underlying columns represented by this column.
// For example, an array of size n will expand into n underlying columns.
func (p *IntType) Width() uint {
	return 1
}

// LeastUpperBound computes the least upper bound of this type and another. This
// is smallest type which contains both of the arguments.  For example, i32 is
// the least upper bound of i1 and i32, etc.  If no such type exists, then nil
// is returned.
func (p *IntType) LeastUpperBound(other Type) Type {
	if o, ok := other.(*IntType); ok {
		var values util.Interval
		//
		switch {
		case p.values == nil && o.values == nil:
			return &IntType{nil}
		case o.values == nil:
			values.Set(p.values)
		case p.values == nil:
			values.Set(o.values)
		default:
			values.Set(p.values)
			values.Insert(o.values)
		}
		//
		return &IntType{&values}
	}

	return nil
}

// SubtypeOf determines whether or not this type is a subtype of another.
func (p *IntType) SubtypeOf(other Type) bool {
	if o, ok := other.(*IntType); ok {
		switch {
		case p.values == nil && o.values == nil:
			return true
		case o.values == nil:
			return true
		case p.values == nil:
			return false
		default:
			return p.values.Within(o.values)
		}
	}
	//
	return false
}

func (p *IntType) String() string {
	if p.values != nil {
		width := p.values.BitWidth()
		if p.values.Contains(big.NewInt(-1)) {
			return fmt.Sprintf("i%d", width)
		}
		//
		return fmt.Sprintf("u%d", width)
	}
	//
	return "int"
}

// ============================================================================
// BooleanType
// ============================================================================

// BOOL_TYPE provides a convenient singleone to use instead of creating a
// fresh boolean type, etc.
var BOOL_TYPE = &BoolType{}

// BoolType represents the type of logical conditions, such as equality,
// logical or, etc.
type BoolType struct {
}

// Width returns the number of underlying columns represented by this column.
// For example, an array of size n will expand into n underlying columns.
func (p *BoolType) Width() uint {
	return 1
}

// HasUnderlying determines whether or not this type has an underlying
// representation, or not.
func (p *BoolType) HasUnderlying() bool {
	return false
}

// LeastUpperBound computes the least upper bound of this type and another. This
// is smallest type which contains both of the arguments.  For example, i32 is
// the least upper bound of i1 and i32, etc.  If no such type exists, then nil
// is returned.
func (p *BoolType) LeastUpperBound(other Type) Type {
	if _, ok := other.(*BoolType); ok {
		return BOOL_TYPE
	}
	//
	return nil
}

// SubtypeOf determines whether or not this type is a subtype of another.
func (p *BoolType) SubtypeOf(other Type) bool {
	_, ok := other.(*BoolType)
	//
	return ok
}

func (p *BoolType) String() string {
	return "bool"
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

// HasUnderlying determines whether or not this type has an underlying
// representation, or not.
func (p *ArrayType) HasUnderlying() bool {
	return p.element.HasUnderlying()
}

// Width returns the number of underlying columns represented by this column.
// For example, an array of size n will expand into n underlying columns.
func (p *ArrayType) Width() uint {
	return p.max - p.min + 1
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

// LeastUpperBound computes the least upper bound of this type and another. This
// is smallest type which contains both of the arguments.  For example, i32 is
// the least upper bound of i1 and i32, etc.  If no such type exists, then nil
// is returned.
func (p *ArrayType) LeastUpperBound(other Type) Type {
	if o, ok := other.(*ArrayType); ok && p.min == o.min && p.max == o.max {
		if element := p.element.LeastUpperBound(o.element); element != nil {
			return NewArrayType(element, p.min, p.max)
		}
	}
	//
	return nil
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
