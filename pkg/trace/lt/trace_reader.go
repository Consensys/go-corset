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

	"github.com/consensys/go-corset/pkg/util/collection/array"
	"github.com/consensys/go-corset/pkg/util/collection/pool"
	"github.com/consensys/go-corset/pkg/util/word"
)

// FromBytes parses a byte array representing a given LTv2 trace file into a set
// of columns, or produces an error if the original file was malformed in some
// way.  The file is organised into three sections:
//
//	+-----------------+
//	| HEADER          |
//	+-----------------+
//	| HEAP            |
//	+-----------------+
//	| COLUMNS         |
//	+-----------------+
//
// Here, the HEADER consists primarily of one or more module definitions, each
// of which identifies a module and the columns it contains.  Likewise, the HEAP
// represents a collection of indexed word data.  Finally, COLUMNS contains
// concrete column data which is either represented explicitly (for narrow
// words, like u8) or using indexes into the heap (for wide words, like u256).
func FromBytes(data []byte) (WordHeap, []RawColumn, error) {
	var (
		err                    error
		buf                    = bytes.NewReader(data)
		heap                   pool.LocalHeap[word.BigEndian]
		columns                []RawColumn
		headerBytes, heapBytes uint32
		headers                []moduleHeader
		offset                 uint
	)
	// Determine sizes of all three sections
	if headerBytes, heapBytes, err = readSectionSizes(buf); err != nil {
		return WordHeap{}, nil, err
	}
	// Read header
	if headers, err = readModuleHeaders(buf, headerBytes+heapBytes); err != nil {
		return WordHeap{}, nil, err
	}
	// Read heap
	if err := heap.UnmarshalBinary(data[8+headerBytes : 8+headerBytes+heapBytes]); err != nil {
		return WordHeap{}, nil, err
	}
	//
	offset = uint(8 + headerBytes + heapBytes)
	// Read column data (sequentially)
	for _, module := range headers {
		for _, column := range module.columns {
			var encoding = array.Encoding{
				Encoding: column.encoding,
				Bytes:    data[offset : offset+uint(column.length)],
			}
			// Decode array data
			data := array.Decode(encoding, &heap)
			// Include it
			columns = append(columns, RawColumn{
				Module: module.name,
				Name:   column.name,
				Data:   data,
			})
		}
	}
	// Done
	return heap, columns, nil
}

func readSectionSizes(buf *bytes.Reader) (headerBytes, heapBytes uint32, err error) {
	// Read size of header (in bytes)
	if err := binary.Read(buf, binary.BigEndian, &headerBytes); err != nil {
		return headerBytes, heapBytes, err
	}
	// Read size of heap (in bytes)
	if err := binary.Read(buf, binary.BigEndian, &heapBytes); err != nil {
		return headerBytes, heapBytes, err
	}
	// done
	return headerBytes, heapBytes, nil
}

func readModuleHeaders(buf *bytes.Reader, offset uint32) (headers []moduleHeader, err error) {
	var nModules uint32
	// Read number of modules
	if err := binary.Read(buf, binary.BigEndian, &nModules); err != nil {
		return nil, err
	}
	//
	headers = make([]moduleHeader, nModules)
	//
	for i := range headers {
		if headers[i], offset, err = readModuleHeader(buf, offset); err != nil {
			return headers, err
		}
	}
	// Done
	return headers, nil
}

func readModuleHeader(buf *bytes.Reader, columnOffset uint32) (header moduleHeader, offset uint32, err error) {
	var nColumns uint32
	// Read module name
	if header.name, err = readName(buf); err != nil {
		return header, columnOffset, err
	}
	// Read module height
	if err := binary.Read(buf, binary.BigEndian, &header.height); err != nil {
		return header, columnOffset, err
	}
	// Read number of columns
	if err := binary.Read(buf, binary.BigEndian, &nColumns); err != nil {
		return header, columnOffset, err
	}
	// Read columns
	header.columns = make([]columnHeader, nColumns)
	//
	//
	for i := range header.columns {
		if header.columns[i], err = readColumnHeader(buf, columnOffset); err != nil {
			return header, columnOffset, err
		}
		//
		columnOffset += header.columns[i].length
	}
	//
	return header, columnOffset, nil
}

func readColumnHeader(buf *bytes.Reader, columnOffset uint32) (header columnHeader, err error) {
	// Read column name
	if header.name, err = readName(buf); err != nil {
		return header, err
	}
	// Assign column data offset
	header.offset = columnOffset
	// Read column data length (in bytes)
	if err := binary.Read(buf, binary.BigEndian, &header.length); err != nil {
		return header, err
	}
	// read column data encoding
	if err := binary.Read(buf, binary.BigEndian, &header.encoding); err != nil {
		return header, err
	}
	// read column bitwidth
	if err := binary.Read(buf, binary.BigEndian, &header.bitwidth); err != nil {
		return header, err
	}
	// Done
	return header, nil
}

func readName(buf *bytes.Reader) (string, error) {
	var (
		// Qualified column name length
		nameLen uint16
	)
	// Read name length
	if err := binary.Read(buf, binary.BigEndian, &nameLen); err != nil {
		return "", err
	}
	// Read name bytes
	name := make([]byte, nameLen)
	if _, err := buf.Read(name); err != nil {
		return "", err
	}
	// Done
	return string(name), nil
}

type moduleHeader struct {
	// Name of the module
	name string
	// Height (in rows) of the module
	height uint32
	// Column details
	columns []columnHeader
}

type columnHeader struct {
	// Name of the column
	name string
	// Starting offset of column data
	offset uint32
	// Length (in bytes) of column data
	length uint32
	// encoding scheme used for column data
	encoding uint32
	// Bitwidth of column
	bitwidth uint16
}
