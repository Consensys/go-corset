package field

import (
	"io"
	"strings"

	"github.com/consensys/gnark-crypto/ecc/bls12-377/fr"
	"github.com/consensys/go-corset/pkg/util"
)

// ----------------------------------------------------------------------------

// FrPtrElementArray implements an array of field elements using an underlying
// byte array.  Each element occupies a fixed number of bytes, known as the
// width.  This is space efficient when a known upper bound holds for the given
// elements.  For example, when storing elements which always fit within 16bits,
// etc.
type FrPtrElementArray struct {
	// The data stored in this column (as bytes).
	elements []*fr.Element
	// Maximum number of bits required to store an element of this array.
	bitwidth uint
}

// NewFrPtrElementArray constructs a new field array with a given capacity.
func NewFrPtrElementArray(height uint, bitwidth uint) *FrPtrElementArray {
	elements := make([]*fr.Element, height)
	return &FrPtrElementArray{elements, bitwidth}
}

// Len returns the number of elements in this field array.
func (p *FrPtrElementArray) Len() uint {
	return uint(len(p.elements))
}

// BitWidth returns the width (in bits) of elements in this array.
func (p *FrPtrElementArray) BitWidth() uint {
	return p.bitwidth
}

// Get returns the field element at the given index in this array.
func (p *FrPtrElementArray) Get(index uint) fr.Element {
	return *p.elements[index]
}

// Set sets the field element at the given index in this array, overwriting the
// original value.
func (p *FrPtrElementArray) Set(index uint, element fr.Element) {
	p.elements[index] = &element
}

// Clone makes clones of this array producing an otherwise identical copy.
func (p *FrPtrElementArray) Clone() util.Array[fr.Element] {
	// Allocate sufficient memory
	ndata := make([]*fr.Element, uint(len(p.elements)))
	// Copy over the data
	copy(ndata, p.elements)
	//
	return &FrPtrElementArray{ndata, p.bitwidth}
}

// Slice out a subregion of this array.
func (p *FrPtrElementArray) Slice(start uint, end uint) util.Array[fr.Element] {
	return &FrPtrElementArray{p.elements[start:end], p.bitwidth}
}

// PadFront (i.e. insert at the beginning) this array with n copies of the given padding value.
func (p *FrPtrElementArray) PadFront(n uint, padding fr.Element) util.Array[fr.Element] {
	pad := &padding
	// Allocate sufficient memory
	ndata := make([]*fr.Element, uint(len(p.elements))+n)
	// Copy over the data
	copy(ndata[n:], p.elements)
	// Go padding!
	for i := uint(0); i < n; i++ {
		ndata[i] = pad
	}
	// Copy over
	return &FrPtrElementArray{ndata, p.bitwidth}
}

// Write the raw bytes of this column to a given writer, returning an error
// if this failed (for some reason).
func (p *FrPtrElementArray) Write(w io.Writer) error {
	for _, e := range p.elements {
		// Read exactly 32 bytes
		bytes := e.Bytes()
		// Write them out
		if _, err := w.Write(bytes[:]); err != nil {
			return err
		}
	}
	//
	return nil
}

func (p *FrPtrElementArray) String() string {
	var sb strings.Builder

	sb.WriteString("[")

	for i := 0; i < len(p.elements); i++ {
		if i != 0 {
			sb.WriteString(",")
		}

		sb.WriteString(p.elements[i].String())
	}

	sb.WriteString("]")

	return sb.String()
}
