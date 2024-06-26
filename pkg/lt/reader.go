package lt

import (
	"bytes"
	"encoding/binary"
)

// FromBytes parses a byte array representing a given LT trace file into an
// columns, or produces an error if the original file was malformed in some way.
func FromBytes(data []byte) (TraceFile, error) {
	var empty TraceFile
	// Construct new bytes.Reader
	buf := bytes.NewReader(data)
	// Read Number of Columns
	var ncols uint32
	if err := binary.Read(buf, binary.BigEndian, &ncols); err != nil {
		return empty, err
	}
	// Read column headers
	columns := make([]*Column, ncols)

	for i := uint32(0); i < ncols; i++ {
		var err error
		columns[i], err = readColumnHeader(buf)
		// Sanity check whether an error occurred
		if err != nil {
			// Return what we can anyway.
			return TraceFile{columns}, err
		}
	}
	// Determine byte slices
	offset := len(data) - buf.Len()

	for i := uint32(0); i < ncols; i++ {
		// Calculate length (in bytes) of this column
		nbytes := int(columns[i].bytesPerElement) * int(columns[i].length)
		// Construct appropriate slice
		columns[i].bytes = data[offset : offset+nbytes]
	}
	// Done
	return TraceFile{columns}, nil
}

// Read the meta-data for a specific column in this trace file.
func readColumnHeader(buf *bytes.Reader) (*Column, error) {
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
	// Done
	return &Column{string(name), bytesPerElement, length, nil}, nil
}
