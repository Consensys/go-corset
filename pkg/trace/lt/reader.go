package lt

import (
	"bytes"
	"encoding/binary"

	"github.com/consensys/gnark-crypto/ecc/bls12-377/fr"
	"github.com/consensys/go-corset/pkg/trace"
)

// FromBytes parses a byte array representing a given LT trace file into an
// columns, or produces an error if the original file was malformed in some way.
func FromBytes(data []byte) (trace.Trace, error) {
	// Construct new bytes.Reader
	buf := bytes.NewReader(data)
	// Read Number of BytesColumns
	var ncols uint32
	if err := binary.Read(buf, binary.BigEndian, &ncols); err != nil {
		return nil, err
	}
	// Read column headers
	columns := make([]trace.Column, ncols)

	for i := uint32(0); i < ncols; i++ {
		var err error
		columns[i], err = readColumnHeader(buf)
		// Sanity check whether an error occurred
		if err != nil {
			// Return what we can anyway.
			return nil, err
		}
	}
	// Determine byte slices
	offset := uint(len(data) - buf.Len())

	for i := uint32(0); i < ncols; i++ {
		ith := columns[i].(*trace.BytesColumn)
		// Calculate length (in bytes) of this column
		nbytes := ith.Width() * ith.Height()
		// Construct appropriate slice
		ith.SetBytes(data[offset : offset+nbytes])
		// Update byte offset
		offset += nbytes
	}
	// Done
	return trace.NewArrayTrace(columns)
}

// Read the meta-data for a specific column in this trace file.
func readColumnHeader(buf *bytes.Reader) (*trace.BytesColumn, error) {
	var nameLen uint16
	// Read column name length
	if err := binary.Read(buf, binary.BigEndian, &nameLen); err != nil {
		return nil, err
	}
	// Read column name bytes
	name := make([]byte, nameLen)
	if _, err := buf.Read(name); err != nil {
		return nil, err
	}

	// Read bytes per element
	var bytesPerElement uint8
	if err := binary.Read(buf, binary.BigEndian, &bytesPerElement); err != nil {
		return nil, err
	}

	// Read column length
	var length uint32
	if err := binary.Read(buf, binary.BigEndian, &length); err != nil {
		return nil, err
	}
	// Default padding
	zero := fr.NewElement(0)
	// Done
	// FIXME: module index should not always be zero!
	return trace.NewBytesColumn(0, string(name), bytesPerElement, uint(length), nil, &zero), nil
}
