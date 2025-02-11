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
	"math/big"
	"strings"

	"github.com/consensys/go-corset/pkg/trace"
	"github.com/consensys/go-corset/pkg/util/field"
)

// FromBytes parses a trace expressed in JSON notation.  For example, {"X":
// [0], "Y": [1]} is a trace containing one row of data each for two columns "X"
// and "Y".
func FromBytes(bytes []byte) ([]trace.RawColumn, error) {
	var rawData map[string][]*big.Int
	// Unmarshall
	jsonErr := json.Unmarshal(bytes, &rawData)
	if jsonErr != nil {
		return nil, jsonErr
	}
	// Construct column data
	cols := make([]trace.RawColumn, len(rawData))
	index := 0
	//
	for name, rawInts := range rawData {
		// Translate raw bigints into raw field elements
		mod, col := splitQualifiedColumnName(name)
		// TODO: support native field widths in column name.
		data := field.FrArrayFromBigInts(256, rawInts)
		// Construct column
		cols[index] = trace.RawColumn{Module: mod, Name: col, Data: data}
		//
		index++
	}
	// Done.
	return cols, nil
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
