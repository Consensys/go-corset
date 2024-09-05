package lt

import (
	"bytes"
	"encoding/binary"
	"strings"

	"github.com/consensys/gnark-crypto/ecc/bls12-377/fr"
	"github.com/consensys/go-corset/pkg/trace"
	"github.com/consensys/go-corset/pkg/util"
)

// FromBytes parses a byte array representing a given LT trace file into an
// columns, or produces an error if the original file was malformed in some way.
func FromBytes(data []byte) ([]trace.RawColumn, error) {
	// Construct new bytes.Reader
	buf := bytes.NewReader(data)
	// Read Number of BytesColumns
	var ncols uint32
	if err := binary.Read(buf, binary.BigEndian, &ncols); err != nil {
		return nil, err
	}
	// Construct empty environment
	headers := make([]columnHeader, ncols)
	columns := make([]trace.RawColumn, ncols)
	// Read column headers
	for i := uint32(0); i < ncols; i++ {
		header, err := readColumnHeader(buf)
		// Read column
		if err != nil {
			// Handle error
			return nil, err
		}
		// Assign header
		headers[i] = header
	}
	// Determine byte slices
	offset := uint(len(data) - buf.Len())

	for i := uint(0); i < uint(ncols); i++ {
		ith := headers[i]
		// Split qualified column name
		mod, col := splitQualifiedColumnName(ith.name)
		// Calculate length (in bytes) of this column
		nbytes := ith.width * ith.length
		// Read column data
		elements := readColumnData(ith, data[offset:offset+nbytes])
		// Construct appropriate slice
		columns[i] = trace.RawColumn{Module: mod, Name: col, Data: elements}
		// Update byte offset
		offset += nbytes
	}
	// Done
	return columns, nil
}

type columnHeader struct {
	name   string
	length uint
	width  uint
}

// Read the meta-data for a specific column in this trace file.
func readColumnHeader(buf *bytes.Reader) (columnHeader, error) {
	var header columnHeader
	// Qualified column name length
	var nameLen uint16
	// Read column name length
	if err := binary.Read(buf, binary.BigEndian, &nameLen); err != nil {
		return header, err
	}
	// Read column name bytes
	name := make([]byte, nameLen)
	if _, err := buf.Read(name); err != nil {
		return header, err
	}

	// Read bytes per element
	var bytesPerElement uint8
	if err := binary.Read(buf, binary.BigEndian, &bytesPerElement); err != nil {
		return header, err
	}

	// Read column length
	var length uint32
	if err := binary.Read(buf, binary.BigEndian, &length); err != nil {
		return header, err
	}
	// Height is length
	header.length = uint(length)
	header.name = string(name)
	header.width = uint(bytesPerElement)
	// Add new column
	return header, nil
}

func readColumnData(header columnHeader, bytes []byte) util.FrArray {
	// Construct array
	data := util.NewFrArray(header.length, header.width*8)
	// Assign elements
	offset := uint(0)

	for i := uint(0); i < header.length; i++ {
		var ith fr.Element
		// Calculate position of next element
		next := offset + header.width
		// Initialise element
		ith.SetBytes(bytes[offset:next])
		// Construct ith field element
		data.Set(i, ith)
		// Move offset to next element
		offset = next
	}
	// Done
	return data
}

// SplitQualifiedColumnName splits a qualified column name into its module and
// column components.
func splitQualifiedColumnName(name string) (string, string) {
	i := strings.Index(name, ".")
	if i >= 0 {
		// Split on "."
		return name[0:i], name[i+1:]
	}
	// No module name given, therefore its in the prelude.
	return "", name
}
