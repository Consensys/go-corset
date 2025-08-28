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
	"github.com/consensys/go-corset/pkg/util/word"
)

// ToBytes writes a given trace file as an array of bytes.  See FromBytes for
// more information on the layout of data in this format.
func ToBytes(heap WordHeap, rawColumns []trace.RawColumn[word.BigEndian]) ([]byte, error) {
	var (
		buf         bytes.Buffer
		err         error
		headerBytes []byte
		heapBytes   []byte
	)
	// For now we do an ugly split
	columns := splitRawColumns(rawColumns)
	// Construct header data
	if headerBytes, err = toHeaderBytes(columns); err != nil {
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
	//
	fmt.Printf("Wrote %d heap bytes with %d entries\n", len(heapBytes), heap.Size())
	//
	return buf.Bytes(), nil
}

func splitRawColumns(rawColumns []trace.RawColumn[word.BigEndian]) [][]trace.RawColumn[word.BigEndian] {
	var (
		mapping = make(map[string]uint)
		columns [][]trace.RawColumn[word.BigEndian]
	)
	//
	for _, col := range rawColumns {
		mid, ok := mapping[col.Module]
		// Check whether module seen before
		if !ok {
			// no
			mapping[col.Module] = uint(len(columns))
			columns = append(columns, nil)
		}
		//
		columns[mid] = append(columns[mid], col)
	}
	//
	return columns
}

func toHeaderBytes(modules [][]trace.RawColumn[word.BigEndian]) ([]byte, error) {
	var buf bytes.Buffer
	// Write number of modules
	if err := binary.Write(&buf, binary.BigEndian, uint32(len(modules))); err != nil {
		return nil, err
	}
	//
	for _, module := range modules {
		if err := writeModuleHeader(&buf, module); err != nil {
			return nil, err
		}
	}
	//
	return buf.Bytes(), nil
}

func writeModuleHeader(buf io.Writer, columns []trace.RawColumn[word.BigEndian]) (err error) {
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
	for _, col := range columns {
		if err = writeColumnHeader(buf, col); err != nil {
			return err
		}
	}
	//
	return nil
}

func writeColumnHeader(buf io.Writer, column trace.RawColumn[word.BigEndian]) (err error) {
	var (
		length   uint32 // in bytes
		encoding uint16
		bitwidth uint16 = uint16(column.Data.BitWidth())
	)
	// Write column name
	if err = writeName(buf, column.Name); err != nil {
		return err
	}
	// Write column data length
	if err := binary.Write(buf, binary.BigEndian, length); err != nil {
		return err
	}
	// Write column data encoding schema
	if err := binary.Write(buf, binary.BigEndian, encoding); err != nil {
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
