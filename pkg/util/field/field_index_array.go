package field

import (
	"io"
	"strings"

	"github.com/consensys/gnark-crypto/ecc/bls12-377/fr"
	"github.com/consensys/go-corset/pkg/util"
)

// FrIndexArray provides an "indexed" array of field elements.  This applies two
// specific optimisations: (1) elements requiring only 16bits are optimised
// using a precomputed table; (2) otherwise, elements are referred to by index.
type FrIndexArray struct {
	// The data stored in this column (as indexes into the heap).
	elements []int32
	// The set of elements in this array
	heap []fr.Element
	// Maximum number of bits required to store an element of this array.
	bitwidth uint
}

// NewFrIndexArray constructs a new field array with a given capacity.
func NewFrIndexArray(height uint, bitwidth uint) *FrIndexArray {
	elements := make([]int32, height)
	// NOTE: must be one element on the heap initially, otherwise index 0 gets
	// confused with element 0.  This first element is always unused.
	heap := make([]fr.Element, 1)
	//
	return &FrIndexArray{elements, heap, bitwidth}
}

// Len returns the number of elements in this field array.
func (p *FrIndexArray) Len() uint {
	return uint(len(p.elements))
}

// BitWidth returns the width (in bits) of elements in this array.
func (p *FrIndexArray) BitWidth() uint {
	return p.bitwidth
}

// Get returns the field element at the given index in this array.
func (p *FrIndexArray) Get(index uint) fr.Element {
	// Check for pool element
	if elem := p.elements[index]; elem <= 0 {
		return pool16bit[-elem]
	} else {
		return p.heap[elem]
	}
}

// Set sets the field element at the given index in this array, overwriting the
// original value.
func (p *FrIndexArray) Set(index uint, element fr.Element) {
	//
	if e := element.Uint64(); element.IsUint64() && e < 65536 {
		p.elements[index] = -int32(e)
	} else {
		p.elements[index] = int32(len(p.heap))
		p.heap = append(p.heap, element)
	}
}

// Clone makes clones of this array producing an otherwise identical copy.
func (p *FrIndexArray) Clone() util.Array[fr.Element] {
	// Allocate sufficient memory
	elements := make([]int32, len(p.elements))
	heap := make([]fr.Element, len(p.heap))
	// Copy over the data
	copy(elements, p.elements)
	copy(heap, p.heap)
	//
	return &FrIndexArray{elements, heap, p.bitwidth}
}

// Slice out a subregion of this array.
func (p *FrIndexArray) Slice(start uint, end uint) util.Array[fr.Element] {
	panic("todo")
}

// PadFront (i.e. insert at the beginning) this array with n copies of the given padding value.
func (p *FrIndexArray) PadFront(n uint, padding fr.Element) util.Array[fr.Element] {
	// Allocate sufficient memory
	elements := make([]int32, uint(len(p.elements))+n)
	heap := make([]fr.Element, len(p.heap))
	// Copy over the data
	copy(elements[n:], p.elements)
	copy(heap, p.heap)
	//
	narr := &FrIndexArray{elements, heap, p.bitwidth}
	// Go padding!
	for i := uint(0); i < n; i++ {
		narr.Set(i, padding)
	}
	// Copy over
	return narr
}

// Write the raw bytes of this column to a given writer, returning an error
// if this failed (for some reason).
func (p *FrIndexArray) Write(w io.Writer) error {
	for _, e := range p.elements {
		var fv fr.Element
		//
		if e <= 0 {
			fv = pool16bit[-e]
		} else {
			fv = p.heap[e]
		}
		// Read exactly 32 bytes
		bytes := fv.Bytes()
		// Write them out
		if _, err := w.Write(bytes[:]); err != nil {
			return err
		}
	}
	//
	return nil
}

func (p *FrIndexArray) String() string {
	var sb strings.Builder

	sb.WriteString("[")

	for i := uint(0); i < p.Len(); i++ {
		if i != 0 {
			sb.WriteString(",")
		}

		ith := p.Get(i)
		sb.WriteString(ith.String())
	}

	sb.WriteString("]")

	return sb.String()
}
