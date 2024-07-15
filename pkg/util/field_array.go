package util

import (
	"io"
	"math/big"

	"github.com/consensys/gnark-crypto/ecc/bls12-377/fr"
)

// Array provides a generice interface to an array of elements.  Typically, we
// are interested in arrays of field elements here.
type Array[T comparable] interface {
	// Returns the number of elements in this array.
	Len() uint
	// Get returns the element at the given index in this array.
	Get(uint) T
	// Set the element at the given index in this array, overwriting the
	// original value.
	Set(uint, T)
	// Clone makes clones of this array producing an otherwise identical copy.
	Clone() Array[T]
	// Return the number of bytes required to store an element of this array.
	ByteWidth() uint
	// Insert a given number of copies of T at start of array producing an
	// updated array.
	PadFront(uint, T) Array[T]
	// Write out the contents of this array, assuming a minimal unit of 1 byte
	// per element.
	Write(w io.Writer) error
}

// ----------------------------------------------------------------------------

// FieldArray implements an array of field elements using an underlying
// byte array.  Each element occupies a fixed number of bytes, known as the
// width.  This is space efficient when a known upper bound holds for the given
// elements.  For example, when storing elements which always fit within 16bits,
// etc.
type FieldArray struct {
	// The data stored in this column (as bytes).
	bytes []byte
	// The number of data elements in this column.
	height uint
	// Determines how many bytes each field element takes.  For the BLS12-377
	// curve, this should be 32.  In the future, when other curves are
	// supported, this could be less.
	width uint8
}

// NewFieldArray constructs a new field array with a given capacity.
func NewFieldArray(height uint, width uint8) *FieldArray {
	bytes := make([]byte, height*uint(width))
	return &FieldArray{bytes, height, width}
}

// FieldArrayFromBigInts converts an array of big integers into an array of
// field elements.
func FieldArrayFromBigInts(width uint8, ints []*big.Int) *FieldArray {
	elements := NewFieldArray(uint(len(ints)), width)
	// Convert each integer in turn.
	for i, v := range ints {
		element := new(fr.Element)
		element.SetBigInt(v)
		elements.Set(uint(i), element)
	}

	// Done.
	return elements
}

// Len returns the number of elements in this field array.
func (p *FieldArray) Len() uint {
	return p.height
}

// ByteWidth returns the width of elements in this array.
func (p *FieldArray) ByteWidth() uint {
	return uint(p.width)
}

// Get returns the field element at the given index in this array.
func (p *FieldArray) Get(index uint) *fr.Element {
	if index >= p.height {
		panic("out-of-bounds access")
	}
	// Element which will hold value.
	var elem fr.Element
	// Determine starting offset within bytes slice
	start := uint(p.width) * index
	end := start + uint(p.width)
	// Construct field element.
	return elem.SetBytes(p.bytes[start:end])
}

// Set sets the field element at the given index in this array, overwriting the
// original value.
func (p *FieldArray) Set(index uint, element *fr.Element) {
	bytes := element.Bytes()
	// Determine starting offset within bytes slice
	bytes_start := uint(p.width) * index
	bytes_end := bytes_start + uint(p.width)
	elem_start := 32 - p.width
	// Copy data
	copy(p.bytes[bytes_start:bytes_end], bytes[elem_start:])
}

// Clone makes clones of this array producing an otherwise identical copy.
func (p *FieldArray) Clone() Array[*fr.Element] {
	n := len(p.bytes)
	nbytes := make([]byte, n)
	copy(nbytes, p.bytes)
	// Done
	return &FieldArray{nbytes, p.height, p.width}
}

// PadFront (i.e. insert at the beginning) this array with n copies of the given padding value.
func (p *FieldArray) PadFront(n uint, padding *fr.Element) Array[*fr.Element] {
	// Computing padding length (in bytes)
	padding_len := n * uint(p.width)
	// Access bytes to use for padding
	padding_bytes := padding.Bytes()
	padded_bytes := make([]byte, padding_len+uint(len(p.bytes)))
	// Append padding
	offset := 0

	for i := uint(0); i < n; i++ {
		// Calculate starting position within the 32byte array, remembering that
		// padding_bytes is stored in _big endian_ format meaning
		// padding_bytes[0] is the _most significant_ byte.
		start := 32 - p.width
		// Copy over least significant bytes
		for j := start; j < 32; j++ {
			padded_bytes[offset] = padding_bytes[j]
			offset++
		}
	}
	// Copy over original data
	copy(padded_bytes[padding_len:], p.bytes)
	// Done
	return &FieldArray{padded_bytes, p.height + n, p.width}
}

// Write the raw bytes of this column to a given writer, returning an error
// if this failed (for some reason).
func (p *FieldArray) Write(w io.Writer) error {
	_, err := w.Write(p.bytes)
	return err
}
