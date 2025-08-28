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
package json

import (
	"encoding/json"
	"fmt"
	"math"
	"math/big"
	"strconv"
	"strings"

	"github.com/consensys/go-corset/pkg/trace"
	"github.com/consensys/go-corset/pkg/util/collection/array"
	"github.com/consensys/go-corset/pkg/util/collection/pool"
	"github.com/consensys/go-corset/pkg/util/word"
)

// ArrayBuilder provides a usefuil alias
type ArrayBuilder = array.DynamicBuilder[word.BigEndian, *pool.LocalHeap[word.BigEndian]]

// WordHeap provides a usefuil alias
type WordHeap = pool.LocalHeap[word.BigEndian]

// FromBytes parses a trace expressed in JSON notation.  For example, {"X":
// [0], "Y": [1]} is a trace containing one row of data each for two columns "X"
// and "Y".
func FromBytes(data []byte) (WordHeap, []trace.RawColumn[word.BigEndian], error) {
	var (
		rawData map[string]map[string][]big.Int
		cols    []trace.RawColumn[word.BigEndian]
	)
	// Attempt to unmarshall
	jsonErr := json.Unmarshal(data, &rawData)
	if jsonErr != nil {
		// Failed, so try and fall back on the legacy format.
		return FromBytesLegacy(data)
	}
	// Intialise builder
	heap := pool.NewLocalHeap[word.BigEndian]()
	builder := array.NewDynamicBuilder(heap)
	//
	for mod, modData := range rawData {
		for name, rawInts := range modData {
			col, bitwidth, error := splitColumnBitwidth(name)
			// error check
			if error != nil {
				return WordHeap{}, nil, error
			}
			// Validate data array
			if row := validateBigInts(bitwidth, rawInts); row != math.MaxUint {
				return WordHeap{}, nil, fmt.Errorf("column %s out-of-bounds (row %d, value %s)",
					name, row, rawInts[row].String())
			}
			// Construct data array
			data := newArrayFromBigInts(bitwidth, rawInts, builder)
			// Construct column
			cols = append(cols, trace.RawColumn[word.BigEndian]{Module: mod, Name: col, Data: data})
		}
	}
	//
	return *heap, cols, nil
}

// FromBytesLegacy parses a trace expressed in JSON notation.  For example, {"X":
// [0], "Y": [1]} is a trace containing one row of data each for two columns "X"
// and "Y".
func FromBytesLegacy(data []byte) (WordHeap, []trace.RawColumn[word.BigEndian], error) {
	var (
		rawData map[string][]big.Int
		heap    = pool.NewLocalHeap[word.BigEndian]()
		builder = array.NewDynamicBuilder(heap)
	)
	// Unmarshall
	jsonErr := json.Unmarshal(data, &rawData)
	if jsonErr != nil {
		return WordHeap{}, nil, jsonErr
	}
	// Construct column data
	cols := make([]trace.RawColumn[word.BigEndian], len(rawData))
	index := 0
	//
	for name, rawInts := range rawData {
		// Translate raw bigints into raw field elements
		mod, col, bitwidth, error := splitQualifiedColumnName(name)
		// error check
		if error != nil {
			return WordHeap{}, nil, error
		}
		// Validate data array
		if row := validateBigInts(bitwidth, rawInts); row != math.MaxUint {
			return WordHeap{}, nil, fmt.Errorf("column %s out-of-bounds (row %d, value %s)",
				name, row, rawInts[row].String())
		}
		// Construct data array
		data := newArrayFromBigInts(bitwidth, rawInts, builder)
		// Construct column
		cols[index] = trace.RawColumn[word.BigEndian]{Module: mod, Name: col, Data: data}
		//
		index++
	}
	// Done.
	return *heap, cols, nil
}

func newArrayFromBigInts(bitwidth uint, data []big.Int, pool ArrayBuilder) array.MutArray[word.BigEndian] {
	//
	var (
		n   = uint(len(data))
		arr = pool.NewArray(n, bitwidth)
	)
	//
	for i := range n {
		ithBytes := data[i].Bytes()
		arr.Set(i, word.NewBigEndian(ithBytes))
	}
	//
	return arr
}

// SplitQualifiedColumnName splits a qualified column name into its module and
// column components.
func splitQualifiedColumnName(name string) (string, string, uint, error) {
	// Check whether bitwidth was provided
	name, bitwidth, error := splitColumnBitwidth(name)
	// error check
	if error != nil {
		return "", "", 0, error
	}
	// Now look for qualified name
	i := strings.Index(name, ".")
	if i >= 0 {
		// Split on "."
		return name[0:i], name[i+1:], bitwidth, nil
	}
	// No module name given, therefore its in the prelude.
	return "", name, bitwidth, nil
}

func splitColumnBitwidth(name string) (string, uint, error) {
	var (
		err      error
		bitwidth uint64
		bits     = strings.Split(name, "@")
	)
	//
	if len(bits) == 1 {
		// no bitwidth given
		return bits[0], 256, nil
	} else if len(bits) > 2 || len(bits[1]) < 2 {
		return "", 0, fmt.Errorf("malformed column name \"%s\"", name)
	} else if bits[1][0] != 'u' {
		return "", 0, fmt.Errorf("malformed column type \"%s\"", bits[1])
	}
	// Extract colwidth, whilst ignoring column type (for now)
	colwidth := bits[1][1:]
	//
	if bitwidth, err = strconv.ParseUint(colwidth, 10, 9); err != nil {
		// failure
		return "", 0, err
	}
	//
	return bits[0], uint(bitwidth), nil
}

func validateBigInts(bitwidth uint, data []big.Int) uint {
	var zero = big.NewInt(0)
	//
	for i, val := range data {
		if val.Cmp(zero) < 0 {
			return uint(i)
		} else if uint(val.BitLen()) > bitwidth {
			return uint(i)
		}
	}
	//
	return math.MaxUint
}
