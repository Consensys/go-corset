// Copyright Consensys Software Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License"); you may not use this file except in compliance with
// the License. You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
// an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
// specific language governing permissions and limitations under the License.
//
// SPDX-License-Identifier: Apache-2.0
package lt

import (
	"bytes"
	"encoding/binary"
	"io"
	"log"

	"github.com/consensys/go-corset/pkg/trace"
	"github.com/consensys/go-corset/pkg/util/collection/array"
	"github.com/consensys/go-corset/pkg/util/word"
)

// ToBytesLegacy writes a given trace file as an array of (legacy) bytes.  The
// output represents the legacy format if the bytes are used "as is" without any
// additional header information being preprended.
func ToBytesLegacy(columns []trace.RawColumn) ([]byte, error) {
	buf, err := ToBytesBuffer(columns)
	if err != nil {
		return nil, err
	}
	//
	return buf.Bytes(), err
}

// ToBytesBuffer writes a given trace file into a byte buffer.
func ToBytesBuffer(columns []trace.RawColumn) (*bytes.Buffer, error) {
	var buf bytes.Buffer
	if err := WriteBytes(columns, &buf); err != nil {
		return nil, err
	}

	return &buf, nil
}

// WriteBytes a given trace file to an io.Writer.
func WriteBytes(columns []trace.RawColumn, buf io.Writer) error {
	ncols := len(columns)
	// Write column count
	if err := binary.Write(buf, binary.BigEndian, uint32(ncols)); err != nil {
		return err
	}
	// Write header information
	for i := 0; i < ncols; i++ {
		col := columns[i]
		data := col.Data
		name := trace.QualifiedColumnName(col.Module, col.Name)
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
	for i := 0; i < ncols; i++ {
		col := columns[i]
		if err := writeArrayBytes(buf, col.Data); err != nil {
			return err
		}
	}
	// Done
	return nil
}

func writeArrayBytes(w io.Writer, data array.Array[word.BigEndian]) error {
	for i := range data.Len() {
		ith := data.Get(i)
		// Read exactly 32 bytes
		bytes := ith.Bytes()
		// Write them out
		if _, err := w.Write(bytes[:]); err != nil {
			return err
		}
	}
	//
	return nil
}
