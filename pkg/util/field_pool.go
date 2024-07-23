package util

import "github.com/consensys/gnark-crypto/ecc/bls12-377/fr"

// FrPool captures a pool of field elements which are used to reduce unnecessary
// duplication of elements.
type FrPool[K any] interface {
	// Allocate an item into the pool, returning its index.
	Put(*fr.Element) K

	// Lookup a given item in the pool using an index.
	Get(K) *fr.Element
}

type FrIndexPool[K PoolIndex] struct {
	pool []fr.Element
}

type PoolIndex interface {
	uint8 | uint16
}

func NewFrIndexPool[K PoolIndex](bitwidth uint) FrIndexPool[K] {
	len := uint(1) << bitwidth
	// Construct empty array
	pool := make([]fr.Element, len)
	// Initialise array
	for i := uint(0); i < len; i++ {
		pool[i] = fr.NewElement(uint64(i))
	}
	//
	return FrIndexPool[K]{pool}
}

func (p *FrIndexPool[K]) Get(index K) *fr.Element {
	return &p.pool[index]
}

func (p *FrIndexPool[K]) Put(element *fr.Element) K {
	val := element.Uint64()
	// Sanity checks
	if !element.IsUint64() || val >= uint64(len(p.pool)) {
		panic("invalid field element for array pool")
	}
	// Done
	return K(val)
}
