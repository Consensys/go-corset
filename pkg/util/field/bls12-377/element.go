package bls12_377

import (
	"fmt"
	"hash/fnv"

	"github.com/consensys/gnark-crypto/ecc/bls12-377/fr"
)

// Element wraps fr.Element to conform
// to the field.Element interface.
type Element struct {
	fr.Element
}

// Add x + y
func (x Element) Add(y Element) Element {
	var res fr.Element
	//
	res.Add(&x.Element, &y.Element)
	//
	return Element{res}
}

// Sub x - y
func (x Element) Sub(y Element) Element {
	var elem fr.Element
	//
	elem.Sub(&x.Element, &y.Element)
	//
	return Element{elem}
}

// AddUint32 x + y. It's the canonical way to create new elements.
func (x Element) AddUint32(y uint32) Element {
	res := fr.NewElement(uint64(y))
	res.Add(&x.Element, &res)
	//
	return Element{res}
}

// ToUint32 returns the numerical value of x.
func (x Element) ToUint32() uint32 {
	if !x.IsUint64() {
		panic(fmt.Errorf("cannot convert to uint64: %s", x.String()))
	}

	i := x.Uint64()
	if i >= 1<<32 {
		panic(fmt.Errorf("cannot convert to uint32: %d", i))
	}

	return uint32(i)
}

// Mul x * y
func (x Element) Mul(y Element) Element {
	var elem fr.Element
	//
	elem.Mul(&x.Element, &y.Element)
	//
	return Element{elem}
}

// Cmp returns 1 if x > y, 0 if x = y, and -1 if x < y.
func (x Element) Cmp(y Element) int {
	return x.Element.Cmp(&y.Element)
}

// Double 2x
func (x Element) Double() Element {
	elem := new(fr.Element).Double(&x.Element)
	return Element{*elem}
}

// Half x/2
func (x Element) Half() Element {
	var res fr.Element
	//
	res.Set(&x.Element)
	res.Halve()
	//
	return Element{res}
}

// Inverse x⁻¹, or 0 if x = 0.
func (x Element) Inverse() Element {
	var elem fr.Element
	//
	elem.Inverse(&x.Element)
	//
	return Element{elem}
}

// Bytes returns the big-endian encoded value of the Element, possibly with leading zeros.
func (x Element) Bytes() []byte {
	return x.Marshal()
}

// AddBytes adds the Element to the given big-endian value. It expects exactly 32 bytes.
func (x Element) AddBytes(b []byte) Element {
	var res fr.Element
	// Sanity check
	if len(b) != fr.Bytes {
		panic(fmt.Errorf("expecting exactly %d bytes", fr.Bytes))
	}
	//
	res.Unmarshal(b)
	res.Add(&x.Element, &res)
	//
	return Element{res}
}

func (x Element) String() string {
	return x.Element.String()
}

// Text implementation for the Element interface
func (x Element) Text(base int) string {
	return x.Element.Text(base)
}

// IsOne implementation for the Element interface
func (x Element) IsOne() bool {
	return x.Element.IsOne()
}

// IsZero implementation for the Element interface
func (x Element) IsZero() bool {
	return x.Element.IsZero()
}

// Bit implementation for the Word interface.
func (x Element) Bit(uint) bool {
	panic("todo")
}

// BitWidth implementation for the Word interface.
func (x Element) BitWidth() uint {
	return 252
}

// Put implementation for the Word interface.
func (x Element) Put([]byte) []byte {
	panic("todo")
}

// Set implementation for the Word interface.
func (x Element) Set([]byte) Element {
	panic("todo")
}

// Equals implementation for the Word interface.
func (x Element) Equals(other Element) bool {
	return x == other
}

// Hash implementation for the Word interface.
func (x Element) Hash() uint64 {
	hash := fnv.New64a()
	// FIXME: could do better here.
	hash.Write(x.Bytes())
	// Done
	return hash.Sum64()
}
