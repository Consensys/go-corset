package field

import (
	"encoding/binary"
	"fmt"
	"io"
	"strings"

	"github.com/consensys/gnark-crypto/ecc/bls12-377/fr"
	"github.com/consensys/go-corset/pkg/util"
)

type FrUint64Array struct {
	// The data stored in this column (as bytes).
	elements []uint64
	// Maximum number of bits required to store an element of this array.
	bitwidth uint
}

// NewFrUint64Array constructs a new field array with a given capacity.
func NewFrUint64Array(height uint, bitwidth uint) *FrUint64Array {
	if bitwidth > 64 {
		panic("invalid bitwidth")
	}
	//
	elements := make([]uint64, height)
	//
	return &FrUint64Array{elements, bitwidth}
}

// Len returns the number of elements in this field array.
func (p *FrUint64Array) Len() uint {
	return uint(len(p.elements))
}

// BitWidth returns the width (in bits) of elements in this array.
func (p *FrUint64Array) BitWidth() uint {
	return p.bitwidth
}

// Get returns the field element at the given index in this array.
func (p *FrUint64Array) Get(index uint) fr.Element {
	return fr.NewElement(p.elements[index])
}

// Set sets the field element at the given index in this array, overwriting the
// original value.
func (p *FrUint64Array) Set(index uint, element fr.Element) {
	p.elements[index] = element.Uint64()
}

// Clone makes clones of this array producing an otherwise identical copy.
func (p *FrUint64Array) Clone() util.Array[fr.Element] {
	// Allocate sufficient memory
	ndata := make([]uint64, uint(len(p.elements)))
	// Copy over the data
	copy(ndata, p.elements)
	//
	return &FrUint64Array{ndata, p.bitwidth}
}

// Slice out a subregion of this array.
func (p *FrUint64Array) Slice(start uint, end uint) util.Array[fr.Element] {
	return &FrUint64Array{p.elements[start:end], p.bitwidth}
}

// PadFront (i.e. insert at the beginning) this array with n copies of the given padding value.
func (p *FrUint64Array) PadFront(n uint, padding fr.Element) util.Array[fr.Element] {
	// Allocate sufficient memory
	ndata := make([]uint64, uint(len(p.elements))+n)
	// Copy over the data
	copy(ndata[n:], p.elements)
	// Go padding!
	for i := uint(0); i < n; i++ {
		ndata[i] = padding.Uint64()
	}
	// Copy over
	return &FrUint64Array{ndata, p.bitwidth}
}

// Write the raw bytes of this column to a given writer, returning an error
// if this failed (for some reason).
func (p *FrUint64Array) Write(w io.Writer) error {
	for _, e := range p.elements {
		var bytes [8]byte
		// Read exactly 32 bytes
		binary.BigEndian.PutUint64(bytes[:], e)
		// Write them out
		if _, err := w.Write(bytes[:]); err != nil {
			return err
		}
	}
	//
	return nil
}

func (p *FrUint64Array) String() string {
	var sb strings.Builder

	sb.WriteString("[")

	for i := 0; i < len(p.elements); i++ {
		if i != 0 {
			sb.WriteString(",")
		}

		sb.WriteString(fmt.Sprintf("0x%x", p.elements[i]))
	}

	sb.WriteString("]")

	return sb.String()
}
