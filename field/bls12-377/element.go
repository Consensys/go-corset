package bls12_377

import (
	"fmt"

	"github.com/consensys/gnark-crypto/ecc/bls12-377/fr"
)

// Element wraps fr.Element to conform
// to the field.Element interface.
type Element struct {
	*fr.Element
}

// Add x + y
func (x Element) Add(y Element) Element {
	return Element{new(fr.Element).Add(x.Element, y.Element)}
}

// Sub x - y
func (x Element) Sub(y Element) Element {
	return Element{new(fr.Element).Sub(x.Element, y.Element)}
}

// AddUint32 x + y. It's the canonical way to create new elements.
func (x Element) AddUint32(y uint32) Element {
	res := fr.NewElement(uint64(y))
	return Element{res.Add(x.Element, &res)}
}

// ToUint32 returns the numerical value of x.
func (x Element) ToUint32() uint32 {
	if !x.IsUint64() {
		panic(fmt.Errorf("cannot convert to uint64: %s", x.Element))
	}

	i := x.Uint64()
	if i >= 1<<32 {
		panic(fmt.Errorf("cannot convert to uint32: %d", i))
	}

	return uint32(i)
}

// Mul x * y
func (x Element) Mul(y Element) Element {
	return Element{new(fr.Element).Mul(x.Element, y.Element)}
}

// Cmp returns 1 if x > y, 0 if x = y, and -1 if x < y.
func (x Element) Cmp(y Element) int {
	return x.Element.Cmp(y.Element)
}

// Double 2x
func (x Element) Double() Element {
	return Element{new(fr.Element).Double(x.Element)}
}

// Half x/2
func (x Element) Half() Element {
	res := *x.Element
	res.Halve()

	return Element{&res}
}

// Inverse x⁻¹, or 0 if x = 0.
func (x Element) Inverse() Element {
	return Element{new(fr.Element).Inverse(x.Element)}
}
