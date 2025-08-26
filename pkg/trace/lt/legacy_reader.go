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
	"strings"

	"github.com/consensys/go-corset/pkg/trace"
	"github.com/consensys/go-corset/pkg/util"
	"github.com/consensys/go-corset/pkg/util/collection/array"
	"github.com/consensys/go-corset/pkg/util/word"
)

// FromBytesLegacy parses a byte array representing a given (legacy) LT trace
// file into an columns, or produces an error if the original file was malformed
// in some way.   The input represents the original legacy format of trace files
// (i.e. without any additional header information prepended, etc).
func FromBytesLegacy(data []byte) (WordHeap, []trace.RawColumn[word.BigEndian], error) {
	var (
		// Construct new bytes.Reader
		buf     = bytes.NewReader(data)
		heap    = word.NewSharedHeap[word.BigEndian]()
		builder = word.NewDynamicBuilder(heap)
	)
	// Read Number of BytesColumns
	var ncols uint32
	if err := binary.Read(buf, binary.BigEndian, &ncols); err != nil {
		return WordHeap{}, nil, err
	}
	// Construct empty environment
	headers := make([]columnHeader, ncols)
	columns := make([]trace.RawColumn[word.BigEndian], ncols)
	// Read column headers
	for i := uint32(0); i < ncols; i++ {
		header, err := readColumnHeader(buf)
		// Read column
		if err != nil {
			// Handle error
			return WordHeap{}, nil, err
		}
		// Assign header
		headers[i] = header
	}
	// Determine byte slices
	offset := uint(len(data) - buf.Len())
	c := make(chan util.Pair[uint, array.MutArray[word.BigEndian]], ncols)
	// Dispatch go-routines
	for i := uint(0); i < uint(ncols); i++ {
		ith := headers[i]
		// Calculate length (in bytes) of this column
		nbytes := ith.width * ith.length
		// Dispatch go-routine
		go func(i uint, offset uint) {
			// Read column data
			elements := readColumnData(ith, data[offset:offset+nbytes], builder)
			// Package result
			c <- util.NewPair(i, elements)
		}(i, offset)
		// Update byte offset
		offset += nbytes
	}
	// Collect results
	for i := uint(0); i < uint(ncols); i++ {
		// Read packaged result from channel
		res := <-c
		// Split qualified column name
		mod, col := splitQualifiedColumnName(headers[res.Left].name)
		// Construct appropriate slice
		columns[res.Left] = trace.RawColumn[word.BigEndian]{Module: mod, Name: col, Data: res.Right}
	}
	// Done
	return *heap.Localise(), columns, nil
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

func readColumnData(header columnHeader, bytes []byte, heap ArrayBuilder) array.MutArray[word.BigEndian] {
	// Handle special cases
	switch header.width {
	case 1:
		return readByteColumnData(header, bytes, heap)
	case 2:
		return readWordColumnData(header, bytes, heap)
	case 4:
		return readDWordColumnData(header, bytes, heap)
	case 8:
		return readQWordColumnData(header, bytes, heap)
	}
	// General case
	return readArbitraryColumnData(header, bytes, heap)
}

func readByteColumnData(header columnHeader, bytes []byte, builder ArrayBuilder) array.MutArray[word.BigEndian] {
	arr := builder.NewArray(header.length, header.width*8)
	//
	for i := uint(0); i < header.length; i++ {
		// Construct ith field element
		arr.Set(i, word.NewBigEndian(bytes[i:i+1]))
	}
	// Done
	return arr
}

func readWordColumnData(header columnHeader, bytes []byte, builder ArrayBuilder) array.MutArray[word.BigEndian] {
	var (
		arr    = builder.NewArray(header.length, header.width*8)
		offset = uint(0)
	)
	// Assign elements
	for i := uint(0); i < header.length; i++ {
		// Construct ith element
		arr.Set(i, word.NewBigEndian(bytes[offset:offset+2]))
		// Move offset to next element
		offset += 2
	}
	// Done
	return arr
}

func readDWordColumnData(header columnHeader, bytes []byte, builder ArrayBuilder) array.MutArray[word.BigEndian] {
	var (
		arr    = builder.NewArray(header.length, header.width*8)
		offset = uint(0)
	)
	// Assign elements
	for i := uint(0); i < header.length; i++ {
		// Construct ith element
		arr.Set(i, word.NewBigEndian(bytes[offset:offset+4]))
		// Move offset to next element
		offset += 4
	}
	// Done
	return arr
}

func readQWordColumnData(header columnHeader, bytes []byte, builder ArrayBuilder) array.MutArray[word.BigEndian] {
	var (
		arr    = builder.NewArray(header.length, header.width*8)
		offset = uint(0)
	)
	// Assign elements
	for i := uint(0); i < header.length; i++ {
		// Construct ith element
		arr.Set(i, word.NewBigEndian(bytes[offset:offset+8]))
		// Move offset to next element
		offset += 8
	}
	// Done
	return arr
}

// Read column data which is has arbitrary width
func readArbitraryColumnData(header columnHeader, bytes []byte, builder ArrayBuilder) array.MutArray[word.BigEndian] {
	var (
		arr    = builder.NewArray(header.length, header.width*8)
		offset = uint(0)
	)
	// Assign elements
	for i := uint(0); i < header.length; i++ {
		// Calculate position of next element
		next := offset + header.width
		// Construct ith element
		arr.Set(i, word.NewBigEndian(bytes[offset:next]))
		// Move offset to next element
		offset = next
	}
	// Done
	return arr
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
