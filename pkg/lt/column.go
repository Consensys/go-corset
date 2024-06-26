package lt

import (
	"github.com/consensys/gnark-crypto/ecc/bls12-377/fr"
)

// Column provides access to a specific column in the trace file.
type Column struct {
	name string
	// Determines how many bytes each field element takes.  For the BLS12-377
	// curve, this should be 32.  In the future, when other curves are
	// supported, this could be less.
	bytesPerElement uint8
	// The number of data elements in this column.
	length uint32
	// The data stored in this column (as bytes).
	bytes []byte
}

// Name returns the name of this column
func (p *Column) Name() string {
	return p.name
}

// Height returns the number of rows in this column.
func (p *Column) Height() uint {
	return uint(p.length)
}

// Get the ith row of this column as a field element.
func (p *Column) Get(i uint) *fr.Element {
	var elem fr.Element
	// Determine starting offset within bytes slice
	start := uint(p.bytesPerElement) * i
	end := start + uint(p.bytesPerElement)
	// Construct field element.
	return elem.SetBytes(p.bytes[start:end])
}

// Data constructs an array of field elements from this column.
func (p *Column) Data() []*fr.Element {
	data := make([]*fr.Element, p.length)
	offset := uint(0)

	for i := uint32(0); i < p.length; i++ {
		var ith fr.Element
		// Calculate position of next element
		next := offset + uint(p.bytesPerElement)
		// Construct ith field element
		data[i] = ith.SetBytes(p.bytes[offset:next])
		// Move offset to next element
		offset = next
	}
	// Done
	return data
}
