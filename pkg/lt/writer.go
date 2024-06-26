package lt

import (
	"bytes"
	"encoding/binary"
	"io"
	"log"
)

// ToBytes writes a given trace file as an array of bytes.
func ToBytes(tr TraceFile) ([]byte, error) {
	buf, err := ToBytesBuffer(tr)
	if err != nil {
		return nil, err
	}
	//
	return buf.Bytes(), err
}

// ToBytesBuffer writes a given trace file into a byte buffer.
func ToBytesBuffer(tr TraceFile) (*bytes.Buffer, error) {
	var buf bytes.Buffer
	if err := Write(tr, &buf); err != nil {
		return nil, err
	}

	return &buf, nil
}

// Write a given trace file to an io.Writer.
func Write(tr TraceFile, buf io.Writer) error {
	cols := tr.columns
	ncols := uint32(len(cols))
	// Write column count
	if err := binary.Write(buf, binary.BigEndian, ncols); err != nil {
		return err
	}
	// Write header information
	for _, col := range cols {
		// Write name length
		nameBytes := []byte(col.name)
		nameLen := uint16(len(nameBytes))

		if err := binary.Write(buf, binary.BigEndian, nameLen); err != nil {
			return err
		}
		// Write name bytes
		n, err := buf.Write(nameBytes)
		if n != int(nameLen) || err != nil {
			log.Fatal(err)
		}
		// Write bytes per element
		if err := binary.Write(buf, binary.BigEndian, col.bytesPerElement); err != nil {
			log.Fatal(err)
		}
		// Write Data length
		if err := binary.Write(buf, binary.BigEndian, uint32(col.Height())); err != nil {
			log.Fatal(err)
		}
	}
	// Write column data information
	for _, col := range cols {
		_, err := buf.Write(col.bytes)
		if err != nil {
			return err
		}
	}
	// Done
	return nil
}
