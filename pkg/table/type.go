package table

import (
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
	Accept(*fr.Element) bool

	// Produce a string representation of this type.
	String() string
}

// UintType represents an unsigned integer encoded using a given number of bits.
// For example, for the type "u8" then "nbits" is 8.
type UintType struct {
	// The number of bits this type represents (e.g. 8 for u8, etc).
	nbits uint
	// The numeric bound of all values in this type (e.g. 2^8 for u8, etc).
	bound *fr.Element
}

// NewUintType constructs a new integer type for a given bit width.
func NewUintType(nbits uint) *UintType {
	var maxBigInt big.Int
	// Compute 2^n
	maxBigInt.Exp(big.NewInt(2), big.NewInt(int64(nbits)), nil)
	// Construct bound
	bound := new(fr.Element)
	bound.SetBigInt(&maxBigInt)

	return &UintType{nbits, bound}
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

// Accept determines whether a given value is an element of this type.  For
// example, 123 is an element of the type u8 whilst 256 is not.
func (p *UintType) Accept(val *fr.Element) bool {
	return val.Cmp(p.bound) < 0
}

// BitWidth returns the bitwidth of this type.  For example, the
// bitwidth of the type u8 is 8.
func (p *UintType) BitWidth() uint {
	return p.nbits
}

// HasBound determines whether this type fits within a given bound.  For
// example, a u8 fits within a bound of 256 and also 65536.  However, it does
// not fit within a bound of 255.
func (p *UintType) HasBound(bound uint) bool {
	var n fr.Element = fr.NewElement(uint64(bound))
	return p.bound.Cmp(&n) <= 0
}

// Bound determines the actual bound for all values which are in this type.
func (p *UintType) Bound() *fr.Element {
	return p.bound
}

func (p *UintType) String() string {
	return fmt.Sprintf("u%d", p.nbits)
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

// Accept determines whether a given value is an element of this type.  In
// fact, all field elements are members of this type.
func (p *FieldType) Accept(val *fr.Element) bool {
	return true
}

func (p *FieldType) String() string {
	return "𝔽"
}
