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

	"github.com/consensys/go-corset/pkg/trace"
	"github.com/consensys/go-corset/pkg/util/word"
)

// ToBytes writes a given trace file as an array of  bytes.
func ToBytes(heap WordHeap, columns []trace.RawColumn[word.BigEndian]) ([]byte, error) {
	var (
		buf         bytes.Buffer
		err         error
		headerBytes []byte
		heapBytes   []byte
	)
	// Construct header data
	if headerBytes, err = toHeaderBytes(columns); err != nil {
		return nil, err
	}
	// Construct heap data
	if heapBytes, err = toHeapBytes(heap); err != nil {
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
	//
	return buf.Bytes(), nil
}

func toHeaderBytes(columns []trace.RawColumn[word.BigEndian]) ([]byte, error) {
	panic("todo")
}

func toHeapBytes(heap WordHeap) ([]byte, error) {
	// FOR NOW!
	return nil, nil
}
