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

	"github.com/consensys/go-corset/pkg/util/collection/array"
	"github.com/consensys/go-corset/pkg/util/collection/pool"
	"github.com/consensys/go-corset/pkg/util/word"
)

// FromBytesLegacy parses a byte array representing a given (legacy) LT trace
// file into an columns, or produces an error if the original file was malformed
// in some way.   The input represents the original legacy format of trace files
// (i.e. without any additional header information prepended, etc).
func FromBytesLegacy(data []byte) (WordHeap, []Module[word.BigEndian], error) {
	var (
		buf     = bytes.NewReader(data)
		heap    = pool.NewSharedHeap[word.BigEndian]()
		builder = array.NewDynamicBuilder(heap)
		names   []string
		headers [][]legacyColumnHeader
		modmap  map[string]uint = make(map[string]uint)
	)
	// Read Number of BytesColumns
	var ncols uint32
	if err := binary.Read(buf, binary.BigEndian, &ncols); err != nil {
		return WordHeap{}, nil, err
	}
	// Read column headers
	for i := uint32(0); i < ncols; i++ {
		var header, err = readLegacyColumnHeader(buf)
		// Read column
		if err != nil {
			// Handle error
			return WordHeap{}, nil, err
		}
		// Split out module and column name
		mod, col := splitQualifiedColumnName(header.name)
		// Lookup module index
		index, ok := modmap[mod]
		// Check whether exists
		if !ok {
			index = uint(len(headers))
			modmap[mod] = index
			//
			headers = append(headers, nil)
			names = append(names, mod)
		}
		// Reset header name
		header.name = col
		// Assign header
		headers[index] = append(headers[index], header)
	}
	// Determine byte slices
	offset := uint(len(data) - buf.Len())
	c := make(chan legacyResult, ncols)
	modules := make([]Module[word.BigEndian], len(headers))
	// Dispatch go-routines
	for i, ith := range headers {
		modules[i] = Module[word.BigEndian]{
			Name:    names[i],
			Columns: make([]Column[word.BigEndian], len(ith)),
		}
		//
		for j, jth := range ith {
			// Calculate length (in bytes) of this column
			nbytes := jth.width * jth.length
			// Dispatch go-routine
			go func(mid, cid int, offset uint) {
				// Read column data
				elements := readColumnData(jth, data[offset:offset+nbytes], builder)
				// Package result
				c <- legacyResult{mid, cid, elements}
			}(i, j, offset)
			// Update byte offset
			offset += nbytes
		}
	}
	// Collect results
	for range ncols {
		// Read packaged result from channel
		res := <-c
		// Determine column name
		name := headers[res.module][res.column].name
		// Construct appropriate slice
		modules[res.module].Columns[res.column] = Column[word.BigEndian]{Name: name, Data: res.data}
	}
	// Done
	return *heap.Localise(), modules, nil
}

type legacyResult struct {
	module int
	column int
	data   array.MutArray[word.BigEndian]
}

type legacyColumnHeader struct {
	name   string
	length uint
	width  uint
}

// Read the meta-data for a specific column in this trace file.
func readLegacyColumnHeader(buf *bytes.Reader) (legacyColumnHeader, error) {
	var header legacyColumnHeader
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

func readColumnData(header legacyColumnHeader, bytes []byte, heap ArrayBuilder) array.MutArray[word.BigEndian] {
	// Handle special cases
	switch header.width {
	case 1:
		// Check whether can optimise this case
		if areAllBits(bytes) {
			return readBitColumnData(header, bytes)
		}
		//
		return readByteColumnData(header.length, bytes, 0, 1)
	case 2:
		return readWordColumnData(header, bytes)
	case 4:
		return readDWordColumnData(header, bytes)
	case 8:
		return readQWordColumnData(header, bytes, heap)
	}
	// General case
	return readArbitraryColumnData(header, bytes, heap)
}

func areAllBits(bytes []byte) bool {
	for _, b := range bytes {
		if b > 1 {
			return false
		}
	}
	//
	return true
}

func readBitColumnData(header legacyColumnHeader, bytes []byte) array.MutArray[word.BigEndian] {
	arr := array.NewBitArray[word.BigEndian](header.length)
	//
	for i := uint(0); i < header.length; i++ {
		ith := bytes[i]
		arr.SetRaw(i, ith > 0)
	}
	// Done
	return &arr
}

func readByteColumnData(length uint, bytes []byte, start, stride uint) array.MutArray[word.BigEndian] {
	//
	var (
		arr    = array.NewSmallArray[uint8, word.BigEndian](length, 8)
		offset = start
	)
	//
	for i := uint(0); i < length; i++ {
		ith := bytes[offset]
		arr.SetRaw(i, ith)
		//
		offset += stride
	}
	// Done
	return &arr
}

func readWordColumnData(header legacyColumnHeader, bytes []byte) array.MutArray[word.BigEndian] {
	var (
		arr    = array.NewSmallArray[uint16, word.BigEndian](header.length, header.width*8)
		offset = uint(0)
		mx     uint16
	)
	// Assign elements
	for i := uint(0); i < header.length; i++ {
		var (
			b1  = uint16(bytes[offset])
			b0  = uint16(bytes[offset+1])
			ith = (b1 << 8) | b0
		)
		// Construct ith element
		arr.SetRaw(i, ith)
		// Move offset to next element
		offset += 2
		mx = max(mx, ith)
	}
	//
	switch {
	case mx == 0:
		return array.NewZeroArray[word.BigEndian](header.length)
	case mx < 256:
		return readByteColumnData(header.length, bytes, 1, 2)
	}
	// Done
	return &arr
}

func readDWordColumnData(header legacyColumnHeader, bytes []byte) array.MutArray[word.BigEndian] {
	var (
		arr    = array.NewSmallArray[uint32, word.BigEndian](header.length, header.width*8)
		offset = uint(0)
	)
	// Assign elements
	for i := uint(0); i < header.length; i++ {
		var (
			b3 = uint32(bytes[offset])
			b2 = uint32(bytes[offset+1])
			b1 = uint32(bytes[offset+2])
			b0 = uint32(bytes[offset+3])
		)
		// Construct ith element
		arr.SetRaw(i, (b3<<24)|(b2<<16)|(b1<<8)|b0)
		// Move offset to next element
		offset += 4
	}
	// Done
	return &arr
}

func readQWordColumnData(header legacyColumnHeader, bytes []byte, builder ArrayBuilder) array.MutArray[word.BigEndian] {
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
func readArbitraryColumnData(header legacyColumnHeader, bytes []byte, builder ArrayBuilder,
) array.MutArray[word.BigEndian] {
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
