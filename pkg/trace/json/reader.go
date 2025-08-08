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
	"github.com/consensys/go-corset/pkg/util/word"
)

// WordPool provides a usefuil alias
type WordPool = word.Pool[uint, word.BigEndian]

// FromBytes parses a trace expressed in JSON notation.  For example, {"X":
// [0], "Y": [1]} is a trace containing one row of data each for two columns "X"
// and "Y".
func FromBytes(data []byte) (WordPool, []trace.BigEndianColumn, error) {
	var (
		rawData map[string]map[string][]big.Int
		cols    []trace.BigEndianColumn
	)
	// Attempt to unmarshall
	jsonErr := json.Unmarshal(data, &rawData)
	if jsonErr != nil {
		// Failed, so try and fall back on the legacy format.
		return FromBytesLegacy(data)
	}
	// Intialise pool
	pool := word.NewHeapPool[word.BigEndian]()
	//
	for mod, modData := range rawData {
		for name, rawInts := range modData {
			col, bitwidth, error := splitColumnBitwidth(name)
			// error check
			if error != nil {
				return nil, nil, error
			}
			// Validate data array
			if row := validateBigInts(bitwidth, rawInts); row != math.MaxUint {
				return nil, nil, fmt.Errorf("column %s out-of-bounds (row %d, value %s)",
					name, row, rawInts[row].String())
			}
			// Construct data array
			data := newArrayFromBigInts(bitwidth, rawInts, pool)
			// Construct column
			cols = append(cols, trace.BigEndianColumn{Module: mod, Name: col, Data: data})
		}
	}
	//
	return pool, cols, nil
}

// FromBytesLegacy parses a trace expressed in JSON notation.  For example, {"X":
// [0], "Y": [1]} is a trace containing one row of data each for two columns "X"
// and "Y".
func FromBytesLegacy(data []byte) (WordPool, []trace.BigEndianColumn, error) {
	var (
		rawData map[string][]big.Int
		pool    = word.NewHeapPool[word.BigEndian]()
	)
	// Unmarshall
	jsonErr := json.Unmarshal(data, &rawData)
	if jsonErr != nil {
		return nil, nil, jsonErr
	}
	// Construct column data
	cols := make([]trace.BigEndianColumn, len(rawData))
	index := 0
	//
	for name, rawInts := range rawData {
		// Translate raw bigints into raw field elements
		mod, col, bitwidth, error := splitQualifiedColumnName(name)
		// error check
		if error != nil {
			return nil, nil, error
		}
		// Validate data array
		if row := validateBigInts(bitwidth, rawInts); row != math.MaxUint {
			return nil, nil, fmt.Errorf("column %s out-of-bounds (row %d, value %s)",
				name, row, rawInts[row].String())
		}
		// Construct data array
		data := newArrayFromBigInts(bitwidth, rawInts, pool)
		// Construct column
		cols[index] = trace.BigEndianColumn{Module: mod, Name: col, Data: data}
		//
		index++
	}
	// Done.
	return pool, cols, nil
}

func newArrayFromBigInts[P word.Pool[uint, word.BigEndian]](bitwidth uint, data []big.Int,
	pool P) array.MutArray[word.BigEndian] {
	//
	var (
		n   = uint(len(data))
		arr = word.NewArray(n, bitwidth, pool)
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
