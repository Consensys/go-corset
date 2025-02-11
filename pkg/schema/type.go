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
package schema

import (
	"encoding/gob"
	"fmt"
	"math/big"

	"github.com/consensys/gnark-crypto/ecc/bls12-377/fr"
)

// Type represents a _column type_ which restricts the set of values a column
// can take on.  For example, a column might be restricted to holding only byte
// values (i.e. in the range 0..255).
type Type interface {
	// AsUint accesses this type as an unsigned integer.  If this type is not an
	// unsigned integer, then this returns nil.
	AsUint() *UintType
	// AsField accesses this type as a field element.  If this type is not a
	// field element, then this returns nil.
	AsField() *FieldType
	// Accept checks whether a specific value is accepted by this type
	Accept(fr.Element) bool
	// Return the number of bytes required represent any element of this type.
	ByteWidth() uint
	// Return the minimum number of bits required represent any element of this type.
	BitWidth() uint
	// Compare two types, returning: a negative value if this type is "below"
	// the other; 0 if they are equal, a positive value if this type is "above"
	// the other.
	Cmp(Type) int
	// Check whether subtypes another
	SubtypeOf(Type) bool
	// Produce a string representation of this type.
	String() string
}

// UintType represents an unsigned integer encoded using a given number of bits.
// For example, for the type "u8" then "nbits" is 8.
type UintType struct {
	// The number of bits this type represents (e.g. 8 for u8, etc).
	NumOfBits uint
	// The numeric bound of all values in this type (e.g. 2^8 for u8, etc).
	ValueBound fr.Element
}

// NewUintType constructs a new integer type for a given bit width.
func NewUintType(nbits uint) *UintType {
	var maxBigInt big.Int
	// Compute 2^n
	maxBigInt.Exp(big.NewInt(2), big.NewInt(int64(nbits)), nil)
	// Construct bound
	bound := new(fr.Element)
	bound.SetBigInt(&maxBigInt)

	return &UintType{nbits, *bound}
}

// AsUint accesses this type assuming it is a Uint.  Since this is the case,
// this just returns itself.
func (p *UintType) AsUint() *UintType {
	return p
}

// AsField accesses this type assuming it is a Field.  Since this is not the
// case, this returns nil.
func (p *UintType) AsField() *FieldType {
	return nil
}

// ByteWidth returns the number of bytes required represent any element of this
// type.
func (p *UintType) ByteWidth() uint {
	m := p.NumOfBits / 8
	n := p.NumOfBits % 8
	// Check for even division
	if n == 0 {
		return m
	}
	//
	return m + 1
}

// Accept determines whether a given value is an element of this type.  For
// example, 123 is an element of the type u8 whilst 256 is not.
func (p *UintType) Accept(val fr.Element) bool {
	return val.Cmp(&p.ValueBound) < 0
}

// BitWidth returns the bitwidth of this type.  For example, the
// bitwidth of the type u8 is 8.
func (p *UintType) BitWidth() uint {
	return p.NumOfBits
}

// HasBound determines whether this type fits within a given bound.  For
// example, a u8 fits within a bound of 256 and also 65536.  However, it does
// not fit within a bound of 255.
func (p *UintType) HasBound(bound uint) bool {
	var n fr.Element = fr.NewElement(uint64(bound))
	return p.ValueBound.Cmp(&n) <= 0
}

// Bound determines the actual bound for all values which are in this type.
func (p *UintType) Bound() fr.Element {
	return p.ValueBound
}

// SubtypeOf checks whether this subtypes another
func (p *UintType) SubtypeOf(other Type) bool {
	if other.AsField() != nil {
		return true
	} else if o, ok := other.(*UintType); ok {
		return p.ValueBound == o.ValueBound
	}

	return false
}

// Cmp compares two types, returning: a negative value if this type is "below"
// the other; 0 if they are equal, a positive value if this type is "above" the
// other.
func (p *UintType) Cmp(other Type) int {
	panic("todo")
}

func (p *UintType) String() string {
	return fmt.Sprintf("u%d", p.NumOfBits)
}

// FieldType is the type of raw field elements (normally for a prime field).
type FieldType struct {
}

// AsUint accesses this type assuming it is a Uint.  Since this is not the
// case, this returns nil.
func (p *FieldType) AsUint() *UintType {
	return nil
}

// AsField accesses this type assuming it is a Field.  Since this is the case,
// this just returns itself.
func (p *FieldType) AsField() *FieldType {
	return p
}

// ByteWidth returns the number of bytes required represent any element of this
// type.
func (p *FieldType) ByteWidth() uint {
	return 32
}

// BitWidth returns the bitwidth of this type.  For example, the
// bitwidth of the type u8 is 8.
func (p *FieldType) BitWidth() uint {
	return p.ByteWidth() * 8
}

// SubtypeOf checks whether this subtypes another
func (p *FieldType) SubtypeOf(other Type) bool {
	return other.AsField() != nil
}

// Cmp compares two types, returning: a negative value if this type is "below"
// the other; 0 if they are equal, a positive value if this type is "above" the
// other.
func (p *FieldType) Cmp(other Type) int {
	panic("todo")
}

// Accept determines whether a given value is an element of this type.  In
// fact, all field elements are members of this type.
func (p *FieldType) Accept(val fr.Element) bool {
	return true
}

func (p *FieldType) String() string {
	return "𝔽"
}

// Join computes the Least Upper Bound of two types.  For example, the lub of u16
// and u128 is u128, etc.
func Join(lhs Type, rhs Type) Type {
	if lhs.AsField() != nil || rhs.AsField() != nil {
		return &FieldType{}
	}
	//
	uLhs := lhs.AsUint()
	uRhs := rhs.AsUint()
	//
	if uLhs.NumOfBits >= uRhs.NumOfBits {
		return uLhs
	}
	//
	return uRhs
}

// ============================================================================
// Encoding / Decoding
// ============================================================================

func init() {
	gob.Register(Type(&UintType{}))
	gob.Register(Type(&FieldType{}))
}
