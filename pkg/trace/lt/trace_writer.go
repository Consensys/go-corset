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
	"fmt"
	"io"

	"github.com/consensys/go-corset/pkg/trace"
	"github.com/consensys/go-corset/pkg/util/collection/array"
	"github.com/consensys/go-corset/pkg/util/word"
)

// RawColumn provides a convenient alias
type RawColumn = trace.RawColumn[word.BigEndian]

// ToBytes writes a given trace file as an array of bytes.  See FromBytes for
// more information on the layout of data in this format.
func ToBytes(heap WordHeap, rawColumns []RawColumn) ([]byte, error) {
	var (
		buf         bytes.Buffer
		err         error
		headerBytes []byte
		heapBytes   []byte
	)
	// For now we do an ugly split
	columns, moduleEncodings := splitRawColumns(rawColumns)
	// Construct header data
	if headerBytes, err = toHeaderBytes(columns, moduleEncodings); err != nil {
		return nil, err
	}
	// Construct heap data
	if heapBytes, err = heap.MarshalBinary(); err != nil {
		return nil, err
	}
	// Write header size
	if err := binary.Write(&buf, binary.BigEndian, uint32(len(headerBytes))); err != nil {
		return nil, err
	}
	// Write heap size
	if err := binary.Write(&buf, binary.BigEndian, uint32(len(heapBytes))); err != nil {
		return nil, err
	}
	// Write header bytes
	if n, err := buf.Write(headerBytes); err != nil {
		return nil, err
	} else if n != len(headerBytes) {
		return nil, fmt.Errorf("wrote insufficient header bytes (%d v %d)", n, len(headerBytes))
	}
	// Write heap bytes
	if n, err := buf.Write(heapBytes); err != nil {
		return nil, err
	} else if n != len(heapBytes) {
		return nil, fmt.Errorf("wrote insufficient heap bytes (%d v %d)", n, len(heapBytes))
	}
	// Write column data
	for _, columnEncodings := range moduleEncodings {
		for _, encoding := range columnEncodings {
			if n, err := buf.Write(encoding.Bytes); err != nil {
				return nil, err
			} else if n != len(heapBytes) {
				return nil, fmt.Errorf("wrote insufficient encoded column bytes (%d v %d)", n, len(encoding.Bytes))
			}
		}
	}
	//
	return buf.Bytes(), nil
}

func splitRawColumns(rawColumns []RawColumn) ([][]RawColumn, [][]array.Encoding) {
	var (
		mapping   = make(map[string]uint)
		columns   [][]RawColumn
		encodings [][]array.Encoding
	)
	//
	for _, col := range rawColumns {
		mid, ok := mapping[col.Module]
		// Check whether module seen before
		if !ok {
			// no
			mapping[col.Module] = uint(len(columns))
			columns = append(columns, nil)
			encodings = append(encodings, nil)
		}
		//
		columns[mid] = append(columns[mid], col)
		encodings[mid] = append(encodings[mid], col.Data.Encode())
	}
	//
	return columns, encodings
}

func toHeaderBytes(modules [][]RawColumn, encodings [][]array.Encoding) ([]byte, error) {
	var buf bytes.Buffer
	// Write number of modules
	if err := binary.Write(&buf, binary.BigEndian, uint32(len(modules))); err != nil {
		return nil, err
	}
	//
	for i, module := range modules {
		if err := writeModuleHeader(&buf, module, encodings[i]); err != nil {
			return nil, err
		}
	}
	//
	return buf.Bytes(), nil
}

func writeModuleHeader(buf io.Writer, columns []RawColumn, encodings []array.Encoding) (err error) {
	var (
		name   string
		height uint32
	)
	//
	if len(columns) > 0 {
		name = columns[0].Module
		height = uint32(columns[0].Data.Len())
	}
	// Write module name
	if err = writeName(buf, name); err != nil {
		return err
	}
	// Write module height
	if err := binary.Write(buf, binary.BigEndian, height); err != nil {
		return err
	}
	// Write column count
	if err := binary.Write(buf, binary.BigEndian, uint32(len(columns))); err != nil {
		return err
	}
	// Write column info
	for i, col := range columns {
		if err = writeColumnHeader(buf, col, encodings[i]); err != nil {
			return err
		}
	}
	//
	return nil
}

func writeColumnHeader(buf io.Writer, column RawColumn, encoding array.Encoding) (err error) {
	var (
		bitwidth uint16 = uint16(column.Data.BitWidth())
		len      uint32 = uint32(len(encoding.Bytes))
	)
	// Write column name
	if err = writeName(buf, column.Name); err != nil {
		return err
	}
	// Write column data length
	if err := binary.Write(buf, binary.BigEndian, len); err != nil {
		return err
	}
	// Write column data encoding schema
	if err := binary.Write(buf, binary.BigEndian, encoding.Encoding); err != nil {
		return err
	}
	// Write column bitwidth
	if err := binary.Write(buf, binary.BigEndian, bitwidth); err != nil {
		return err
	}
	// Done
	return nil
}

func writeName(buf io.Writer, name string) (err error) {
	nameBytes := []byte(name)
	nameLen := uint16(len(nameBytes))

	if err := binary.Write(buf, binary.BigEndian, nameLen); err != nil {
		return err
	}
	// Write name bytes
	n, err := buf.Write(nameBytes)
	if n != int(nameLen) {
		return fmt.Errorf("incorrect name bytes written (%d v %d)", nameLen, n)
	}
	//
	return err
}
