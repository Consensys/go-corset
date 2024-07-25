package util

import (
	"sync"

	"github.com/consensys/gnark-crypto/ecc/bls12-377/fr"
)

// FrPool captures a pool of field elements which are used to reduce unnecessary
// duplication of elements.
type FrPool[K any] interface {
	// Allocate an item into the pool, returning its index.
	Put(*fr.Element) K

	// Lookup a given item in the pool using an index.
	Get(K) *fr.Element
}

// ----------------------------------------------------------------------------

// FrMapPool is a pool implementation indexed using a map and backed using a
// dynamically sized array.
type FrMapPool struct {
	bitwidth uint
}

// NewFrMapPool constructs a new pool which uses map to index items.
func NewFrMapPool(bitwidth uint) FrMapPool {
	initMapPool()
	return FrMapPool{bitwidth}
}

// Get looks up the given item in the pool.
func (p FrMapPool) Get(index uint32) *fr.Element {
	poolMapLock.RLock()
	item := &poolMapArray[index]
	poolMapLock.RUnlock()

	return item
}

// Put allocates an item into the pool, returning its index.
func (p FrMapPool) Put(element *fr.Element) uint32 {
	// Lock items
	poolMapLock.Lock()
	index, ok := poolMapIndex[*element]
	//
	if !ok {
		len := uint32(len(poolMapArray))
		index = poolMapSize
		// Update index
		poolMapIndex[*element] = index
		//
		if index == len {
			// capacity reached, so double it.
			tmp := make([]fr.Element, len*5)
			copy(tmp, poolMapArray)
			poolMapArray = tmp
		}
		//
		poolMapArray[index] = *element
		poolMapSize++
	}
	// Done
	poolMapLock.Unlock()
	//
	return index
}

var poolMapLock sync.RWMutex
var poolMapIndex map[[4]uint64]uint32
var poolMapArray []fr.Element
var poolMapSize uint32

func initMapPool() {
	poolMapLock.Lock()
	if poolMapIndex == nil {
		// Initial capacity for 500 elements
		poolMapIndex = make(map[[4]uint64]uint32, 1000)
		poolMapArray = make([]fr.Element, 1000)
	}
	poolMapLock.Unlock()
}

// ----------------------------------------------------------------------------

// FrBitPool is a pool implementation indexed using a single bit which is backed
// by an array of a fixed size.  This is ideally suited for representing bit
// columns.
type FrBitPool struct{}

// NewFrBitPool constructs a new pool which uses a single bit for indexing.
func NewFrBitPool() FrBitPool {
	initPool16()
	//
	return FrBitPool{}
}

// Get looks up the given item in the pool.
func (p FrBitPool) Get(index bool) *fr.Element {
	if index {
		return &pool16bit[1]
	}
	//
	return &pool16bit[0]
}

// Put allocates an item into the pool, returning its index.  Since the pool is
// fixed, then so is the index.
func (p FrBitPool) Put(element *fr.Element) bool {
	val := element.Uint64()
	// Sanity checks
	if !element.IsUint64() || val >= 2 {
		panic("invalid field element for bit pool")
	} else if val == 1 {
		return true
	}
	// Done
	return false
}

// ----------------------------------------------------------------------------

// FrIndexPool is a pool implementation which is backed by an array of a fixed
// size.
type FrIndexPool[K uint8 | uint16] struct{}

// NewFrIndexPool constructs a new pool which uses a given key type for
// indexing.
func NewFrIndexPool[K uint8 | uint16]() FrIndexPool[K] {
	initPool16()
	//
	return FrIndexPool[K]{}
}

// Get looks up the given item in the pool.
func (p FrIndexPool[K]) Get(index K) *fr.Element {
	return &pool16bit[index]
}

// Put allocates an item into the pool, returning its index.  Since the pool is
// fixed, then so is the index.
func (p FrIndexPool[K]) Put(element *fr.Element) K {
	val := element.Uint64()
	// Sanity checks
	if !element.IsUint64() || val >= 65536 {
		panic("invalid field element for array pool")
	}

	return K(val)
}

// -------------------------------------------------------------------------------

var pool16init sync.Once
var pool16bit []fr.Element

// Initialise the index pool.
func initPool16() {
	// Singleton pattern for initialisation.
	pool16init.Do(func() {
		// Construct empty array
		tmp := make([]fr.Element, 65536)
		// Initialise array
		for i := uint(0); i < 65536; i++ {
			tmp[i] = fr.NewElement(uint64(i))
		}
		// Should not race
		pool16bit = tmp
	})
}
