package lt

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"log"

	"github.com/consensys/go-corset/pkg/trace"
)

// ToBytes writes a given trace file as an array of bytes.
func ToBytes(tr trace.Trace) ([]byte, error) {
	buf, err := ToBytesBuffer(tr)
	if err != nil {
		return nil, err
	}
	//
	return buf.Bytes(), err
}

// ToBytesBuffer writes a given trace file into a byte buffer.
func ToBytesBuffer(tr trace.Trace) (*bytes.Buffer, error) {
	var buf bytes.Buffer
	if err := WriteBytes(tr, &buf); err != nil {
		return nil, err
	}

	return &buf, nil
}

// WriteBytes a given trace file to an io.Writer.
func WriteBytes(tr trace.Trace, buf io.Writer) error {
	columns := tr.Columns()
	modules := tr.Modules()
	ncols := columns.Len()
	// Write column count
	if err := binary.Write(buf, binary.BigEndian, uint32(ncols)); err != nil {
		return err
	}
	// Write header information
	for i := uint(0); i < ncols; i++ {
		col := columns.Get(i)
		data := col.Data()
		mod := modules.Get(col.Context().Module())
		name := col.Name()
		// Prepend module name (if applicable)
		if mod.Name() != "" {
			name = fmt.Sprintf("%s.%s", mod.Name(), name)
		}
		// Write name length
		nameBytes := []byte(name)
		nameLen := uint16(len(nameBytes))

		if err := binary.Write(buf, binary.BigEndian, nameLen); err != nil {
			return err
		}
		// Write name bytes
		n, err := buf.Write(nameBytes)
		if n != int(nameLen) || err != nil {
			log.Fatal(err)
		}
		// Determine number of bytes required to hold element of this column.
		byteWidth := data.BitWidth() / 8
		if data.BitWidth()%8 != 0 {
			byteWidth++
		}
		// Write bytes per element
		if err := binary.Write(buf, binary.BigEndian, uint8(byteWidth)); err != nil {
			log.Fatal(err)
		}
		// Write Data length
		if err := binary.Write(buf, binary.BigEndian, uint32(data.Len())); err != nil {
			log.Fatal(err)
		}
	}
	// Write column data information
	for i := uint(0); i < ncols; i++ {
		col := columns.Get(i)
		if err := col.Data().Write(buf); err != nil {
			return err
		}
	}
	// Done
	return nil
}
