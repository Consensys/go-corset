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

// FrArray represents an array of field elements.
type FrArray = Array[*fr.Element]

// NewFrArray creates a new FrArray dynamically based on the given width.
func NewFrArray(height uint, width uint) FrArray {
	switch width {
	case 1, 2:
		return NewIndexArray(height, width*8)
	default:
		return NewFrElementArray(height)
	}
}

// FrArrayFromBigInts converts an array of big integers into an array of
// field elements.
func FrArrayFromBigInts(width uint, ints []*big.Int) FrArray {
	elements := NewFrArray(uint(len(ints)), width)
	// Convert each integer in turn.
	for i, v := range ints {
		element := new(fr.Element)
		element.SetBigInt(v)
		elements.Set(uint(i), element)
	}

	// Done.
	return elements
}

// ----------------------------------------------------------------------------

// FrByteArray implements an array of field elements using an underlying
// byte array.  Each element occupies a fixed number of bytes, known as the
// width.  This is space efficient when a known upper bound holds for the given
// elements.  For example, when storing elements which always fit within 16bits,
// etc.
type FrByteArray struct {
	// The data stored in this column (as bytes).
	bytes []byte
	// The number of data elements in this column.
	height uint
	// Determines how many bytes each field element takes.  For the BLS12-377
	// curve, this should be 32.  In the future, when other curves are
	// supported, this could be less.
	width uint8
}

// NewFrByteArray constructs a new field array with a given capacity.
func NewFrByteArray(height uint, width uint8) *FrByteArray {
	bytes := make([]byte, height*uint(width))
	return &FrByteArray{bytes, height, width}
}

// Len returns the number of elements in this field array.
func (p *FrByteArray) Len() uint {
	return p.height
}

// ByteWidth returns the width of elements in this array.
func (p *FrByteArray) ByteWidth() uint {
	return uint(p.width)
}

// Get returns the field element at the given index in this array.
func (p *FrByteArray) Get(index uint) *fr.Element {
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
func (p *FrByteArray) Set(index uint, element *fr.Element) {
	bytes := element.Bytes()
	// Determine starting offset within bytes slice
	bytes_start := uint(p.width) * index
	bytes_end := bytes_start + uint(p.width)
	elem_start := 32 - p.width
	// Copy data
	copy(p.bytes[bytes_start:bytes_end], bytes[elem_start:])
}

// Clone makes clones of this array producing an otherwise identical copy.
func (p *FrByteArray) Clone() Array[*fr.Element] {
	n := len(p.bytes)
	nbytes := make([]byte, n)
	copy(nbytes, p.bytes)
	// Done
	return &FrByteArray{nbytes, p.height, p.width}
}

// PadFront (i.e. insert at the beginning) this array with n copies of the given padding value.
func (p *FrByteArray) PadFront(n uint, padding *fr.Element) Array[*fr.Element] {
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
	return &FrByteArray{padded_bytes, p.height + n, p.width}
}

// Write the raw bytes of this column to a given writer, returning an error
// if this failed (for some reason).
func (p *FrByteArray) Write(w io.Writer) error {
	_, err := w.Write(p.bytes)
	return err
}

// ----------------------------------------------------------------------------

// FrElementArray implements an array of field elements using an underlying
// byte array.  Each element occupies a fixed number of bytes, known as the
// width.  This is space efficient when a known upper bound holds for the given
// elements.  For example, when storing elements which always fit within 16bits,
// etc.
type FrElementArray struct {
	// The data stored in this column (as bytes).
	elements []*fr.Element
}

// NewFrElementArray constructs a new field array with a given capacity.
func NewFrElementArray(height uint) *FrElementArray {
	elements := make([]*fr.Element, height)
	return &FrElementArray{elements}
}

// Len returns the number of elements in this field array.
func (p *FrElementArray) Len() uint {
	return uint(len(p.elements))
}

// ByteWidth returns the width of elements in this array.
func (p *FrElementArray) ByteWidth() uint {
	return 32
}

// Get returns the field element at the given index in this array.
func (p *FrElementArray) Get(index uint) *fr.Element {
	return p.elements[index]
}

// Set sets the field element at the given index in this array, overwriting the
// original value.
func (p *FrElementArray) Set(index uint, element *fr.Element) {
	p.elements[index] = element
}

// Clone makes clones of this array producing an otherwise identical copy.
func (p *FrElementArray) Clone() Array[*fr.Element] {
	// Allocate sufficient memory
	ndata := make([]*fr.Element, uint(len(p.elements)))
	// Copy over the data
	copy(ndata, p.elements)
	//
	return &FrElementArray{ndata}
}

// PadFront (i.e. insert at the beginning) this array with n copies of the given padding value.
func (p *FrElementArray) PadFront(n uint, padding *fr.Element) Array[*fr.Element] {
	// Allocate sufficient memory
	ndata := make([]*fr.Element, uint(len(p.elements))+n)
	// Copy over the data
	copy(ndata[n:], p.elements)
	// Go padding!
	for i := uint(0); i < n; i++ {
		ndata[i] = padding
	}
	// Copy over
	return &FrElementArray{ndata}
}

// Write the raw bytes of this column to a given writer, returning an error
// if this failed (for some reason).
func (p *FrElementArray) Write(w io.Writer) error {
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

// ----------------------------------------------------------------------------

// FrIndexArray implements an array of field elements using an index to pool
// identical elements.  Specificially, an array of all elements upto a certain
// bound is used as the index.
type FrIndexArray struct {
	// Elements in this array, where each is an index into the pool.
	elements []uint16
	// Determines how many bits are required to hold an element of this array.
	bitwidth uint
}

var pool16bit []fr.Element

// Initialise the index pool.
func initPool() {
	if pool16bit == nil {
		// Construct empty array
		pool16bit = make([]fr.Element, 65536)
		// Initialise array
		for i := uint(0); i < 65536; i++ {
			pool16bit[i] = fr.NewElement(uint64(i))
		}
	}
}

// NewIndexArray constructs a new field array with a given capacity.
func NewIndexArray(height uint, bitwidth uint) *FrIndexArray {
	// Ensure pool initialised
	initPool()
	// Sanity check pool makes sense
	if bitwidth > 16 {
		panic("invalid bitwidth for index array")
	}
	// Create empty array
	elements := make([]uint16, height)
	// Done
	return &FrIndexArray{elements, bitwidth}
}

// Len returns the number of elements in this field array.
func (p *FrIndexArray) Len() uint {
	return uint(len(p.elements))
}

// ByteWidth returns the width of elements in this array.
func (p *FrIndexArray) ByteWidth() uint {
	bw := p.bitwidth / 8
	if p.bitwidth%8 != 0 {
		bw++
	}

	return bw
}

// Get returns the field element at the given index in this array.
func (p *FrIndexArray) Get(index uint) *fr.Element {
	i := p.elements[index]
	return &pool16bit[i]
}

// Set sets the field element at the given index in this array, overwriting the
// original value.
func (p *FrIndexArray) Set(index uint, element *fr.Element) {
	val := element.Uint64()

	if !element.IsUint64() || val >= 65536 {
		panic("field element out-of-bounds")
	}

	p.elements[index] = uint16(val)
}

// Clone makes clones of this array producing an otherwise identical copy.
func (p *FrIndexArray) Clone() Array[*fr.Element] {
	// Allocate sufficient memory
	ndata := make([]uint16, len(p.elements))
	// Copy over the data
	copy(ndata, p.elements)
	//
	return &FrIndexArray{ndata, p.bitwidth}
}

// PadFront (i.e. insert at the beginning) this array with n copies of the given padding value.
func (p *FrIndexArray) PadFront(n uint, padding *fr.Element) Array[*fr.Element] {
	// Extract padding index
	val := padding.Uint64()
	if !padding.IsUint64() || val >= 65536 {
		panic("field element out-of-bounds")
	}
	// Allocate sufficient memory
	nelements := make([]uint16, uint(len(p.elements))+n)
	// Copy over the data
	copy(nelements[n:], p.elements)
	// Go padding!
	for i := uint(0); i < n; i++ {
		nelements[i] = uint16(val)
	}
	// Copy over
	return &FrIndexArray{nelements, p.bitwidth}
}

// Write the raw bytes of this column to a given writer, returning an error
// if this failed (for some reason).
func (p *FrIndexArray) Write(w io.Writer) error {
	for _, i := range p.elements {
		// Read exactly 32 bytes
		bytes := pool16bit[i].Bytes()
		// Write them out
		if _, err := w.Write(bytes[:]); err != nil {
			return err
		}
	}
	//
	return nil
}

// ----------------------------------------------------------------------------

// FrPoolArray implements an array of field elements using an index to pool
// identical elements.  Specificially, an array of all elements upto a certain
// bound is used as the index.
type FrPoolArray[K any] struct {
	pool FrPool[K]
	// Elements in this array, where each is an index into the pool.
	elements []K
	// Determines how many bits are required to hold an element of this array.
	bitwidth uint
}

// NewMapArray constructs a new field array with a given capacity.
func NewMapArray[K any](height uint, bitwidth uint, pool FrPool[K]) *FrPoolArray[K] {
	// Create empty array
	elements := make([]K, height)
	// Done
	return &FrPoolArray[K]{pool, elements, bitwidth}
}

// Len returns the number of elements in this field array.
func (p *FrPoolArray[K]) Len() uint {
	return uint(len(p.elements))
}

// ByteWidth returns the width of elements in this array.
func (p *FrPoolArray[K]) ByteWidth() uint {
	bw := p.bitwidth / 8
	if p.bitwidth%8 != 0 {
		bw++
	}

	return bw
}

// Get returns the field element at the given index in this array.
func (p *FrPoolArray[K]) Get(index uint) *fr.Element {
	key := p.elements[index]
	return p.pool.Get(key)
}

// Set sets the field element at the given index in this array, overwriting the
// original value.
func (p *FrPoolArray[K]) Set(index uint, element *fr.Element) {
	p.elements[index] = p.pool.Put(element)
}

// Clone makes clones of this array producing an otherwise identical copy.
func (p *FrPoolArray[K]) Clone() Array[*fr.Element] {
	// Allocate sufficient memory
	ndata := make([]K, len(p.elements))
	// Copy over the data
	copy(ndata, p.elements)
	//
	return &FrPoolArray[K]{p.pool, ndata, p.bitwidth}
}

// PadFront (i.e. insert at the beginning) this array with n copies of the given padding value.
func (p *FrPoolArray[K]) PadFront(n uint, padding *fr.Element) Array[*fr.Element] {
	key := p.pool.Put(padding)
	// Allocate sufficient memory
	nelements := make([]K, uint(len(p.elements))+n)
	// Copy over the data
	copy(nelements[n:], p.elements)
	// Go padding!
	for i := uint(0); i < n; i++ {
		nelements[i] = key
	}
	// Copy over
	return &FrPoolArray[K]{p.pool, nelements, p.bitwidth}
}

// Write the raw bytes of this column to a given writer, returning an error
// if this failed (for some reason).
func (p *FrPoolArray[K]) Write(w io.Writer) error {
	for _, i := range p.elements {
		// Read exactly 32 bytes
		bytes := p.pool.Get(i).Bytes()
		// Write them out
		if _, err := w.Write(bytes[:]); err != nil {
			return err
		}
	}
	//
	return nil
}
