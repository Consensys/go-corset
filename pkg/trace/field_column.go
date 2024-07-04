package trace

import (
	"io"

	"github.com/consensys/gnark-crypto/ecc/bls12-377/fr"
)

// FieldColumn represents a column of data within a trace where each row is
// stored directly as a field element.  This is the simplest form of column,
// which provides the fastest Get operation (i.e. because it just reads the
// field element out directly).  However, at the same time, it can potentially
// use quite a lot of memory.  In particular, when there are many different
// field elements which have smallish values then this requires excess data.
type FieldColumn struct {
	module uint
	// Holds the name of this column
	name string
	// Length multiplier (i.e. of the data array)
	multiplier uint
	// Holds the raw data making up this column
	data []*fr.Element
	// Value to be used when padding this column
	padding *fr.Element
}

// NewFieldColumn constructs a FieldColumn with the give name, data and padding.
func NewFieldColumn(module uint, name string, multiplier uint, data []*fr.Element, padding *fr.Element) *FieldColumn {
	// Sanity check data length
	if uint(len(data))%multiplier != 0 {
		panic("data length has incorrect multiplier")
	}
	// Done
	return &FieldColumn{module, name, multiplier, data, padding}
}

// Module returns the enclosing module of this column
func (p *FieldColumn) Module() uint {
	return p.module
}

// Name returns the name of the given column.
func (p *FieldColumn) Name() string {
	return p.name
}

// Width determines the number of bytes per element for this column (which, in
// this case, is always 32).
func (p *FieldColumn) Width() uint {
	return 32
}

// Height determines the height of this column.
func (p *FieldColumn) Height() uint {
	return uint(len(p.data))
}

// LengthMultiplier is a multiplier which must be a factor of the height.  For
// example, if the factor is 2 then the height must always be a multiple of 2,
// etc.  This affects padding also, as we must pad to this factor.
func (p *FieldColumn) LengthMultiplier() uint {
	return p.multiplier
}

// Padding returns the value which will be used for padding this column.
func (p *FieldColumn) Padding() *fr.Element {
	return p.padding
}

// Data returns the data for the given column.
func (p *FieldColumn) Data() []*fr.Element {
	return p.data
}

// Get the value at a given row in this column.  If the row is
// out-of-bounds, then the column's padding value is returned instead.
// Thus, this function always succeeds.
func (p *FieldColumn) Get(row int) *fr.Element {
	if row < 0 || row >= len(p.data) {
		// out-of-bounds access
		return p.padding
	}
	// in-bounds access
	return p.data[row]
}

// Clone an FieldColumn
func (p *FieldColumn) Clone() Column {
	clone := new(FieldColumn)
	clone.module = p.module
	clone.name = p.name
	clone.multiplier = p.multiplier
	clone.padding = p.padding
	// NOTE: the following is as we never actually mutate the underlying bytes
	// array.
	clone.data = p.data

	return clone
}

// Pad this column with n copies of the column's padding value.
func (p *FieldColumn) Pad(n uint) {
	// Apply the length multiplier
	n = n * p.multiplier
	// Allocate sufficient memory
	ndata := make([]*fr.Element, uint(len(p.data))+n)
	// Copy over the data
	copy(ndata[n:], p.data)
	// Go padding!
	for i := uint(0); i < n; i++ {
		ndata[i] = p.padding
	}
	// Copy over
	p.data = ndata
}

// Reseat updates the module index of this column (e.g. as a result of a
// realignment).
func (p *FieldColumn) Reseat(mid uint) {
	p.module = mid
}

// Write the raw bytes of this column to a given writer, returning an error
// if this failed (for some reason).  Observe that this always writes data in
// 32byte chunks.
func (p *FieldColumn) Write(w io.Writer) error {
	for _, e := range p.data {
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
