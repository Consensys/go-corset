package trace

import (
	"io"

	"github.com/consensys/gnark-crypto/ecc/bls12-377/fr"
)

// BytesColumn represents a column of data within a trace as a raw byte array,
// such that each element occupies a fixed number of bytes.  Accessing elements
// in this column is potentially slower than for a FieldColumn, as the raw bytes
// must be converted into a field element.
type BytesColumn struct {
	module uint
	name   string
	// Determines how many bytes each field element takes.  For the BLS12-377
	// curve, this should be 32.  In the future, when other curves are
	// supported, this could be less.
	width uint8
	// The number of data elements in this column.
	length uint
	// The data stored in this column (as bytes).
	bytes []byte
	// Value to be used when padding this column
	padding *fr.Element
}

// NewBytesColumn constructs a new BytesColumn from its constituent parts.
func NewBytesColumn(module uint, name string, width uint8, length uint,
	bytes []byte, padding *fr.Element) *BytesColumn {
	return &BytesColumn{module, name, width, length, bytes, padding}
}

// Module returns the enclosing module of this column
func (p *BytesColumn) Module() uint {
	return p.module
}

// Name returns the name of this column
func (p *BytesColumn) Name() string {
	return p.name
}

// Width returns the number of bytes required for each element in this column.
func (p *BytesColumn) Width() uint {
	return uint(p.width)
}

// Height returns the number of rows in this column.
func (p *BytesColumn) Height() uint {
	return p.length
}

// Padding returns the value which will be used for padding this column.
func (p *BytesColumn) Padding() *fr.Element {
	return p.padding
}

// Get the ith row of this column as a field element.
func (p *BytesColumn) Get(i int) *fr.Element {
	// TODO: error for out-of-bounds accesses!!!!
	var elem fr.Element
	// Determine starting offset within bytes slice
	start := int(p.width) * i
	end := start + int(p.width)
	// Construct field element.
	return elem.SetBytes(p.bytes[start:end])
}

// Clone an BytesColumn
func (p *BytesColumn) Clone() Column {
	clone := new(BytesColumn)
	clone.module = p.module
	clone.name = p.name
	clone.length = p.length
	clone.width = p.width
	clone.padding = p.padding
	// NOTE: the following is as we never actually mutate the underlying bytes
	// array.
	clone.bytes = p.bytes
	// Done
	return clone
}

// SetBytes sets the raw byte array underlying this column.  Care must be taken
// when mutating a column which is already being used in a trace, as this could
// lead to unexpected behaviour.
func (p *BytesColumn) SetBytes(bytes []byte) {
	p.bytes = bytes
}

// Data constructs an array of field elements from this column.
func (p *BytesColumn) Data() []*fr.Element {
	data := make([]*fr.Element, p.length)
	offset := uint(0)

	for i := uint(0); i < p.length; i++ {
		var ith fr.Element
		// Calculate position of next element
		next := offset + uint(p.width)
		// Construct ith field element
		data[i] = ith.SetBytes(p.bytes[offset:next])
		// Move offset to next element
		offset = next
	}
	// Done
	return data
}

// Pad this column with n copies of the column's padding value.
func (p *BytesColumn) Pad(n uint) {
	// Computing padding length (in bytes)
	padding_len := n * uint(p.width)
	// Access bytes to use for padding
	padding_bytes := p.padding.Bytes()
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
	p.bytes = padded_bytes
	p.length += n
}

// Reseat updates the module index of this column (e.g. as a result of a
// realignment).
func (p *BytesColumn) Reseat(mid uint) {
	p.module = mid
}

// Write the raw bytes of this column to a given writer, returning an error
// if this failed (for some reason).
func (p *BytesColumn) Write(w io.Writer) error {
	_, err := w.Write(p.bytes)
	return err
}
