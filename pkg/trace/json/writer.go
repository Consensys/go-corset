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
	"fmt"
	"strings"

	"github.com/consensys/go-corset/pkg/trace"
	"github.com/consensys/go-corset/pkg/util/word"
)

// ToJsonString converts a trace into a JSON string.
func ToJsonString(columns []trace.RawColumn[word.BigEndian]) string {
	var builder strings.Builder
	//
	builder.WriteString("{")
	//
	for i := 0; i < len(columns); i++ {
		ith := columns[i]
		//
		if i != 0 {
			builder.WriteString(", ")
		}
		//
		builder.WriteString("\"")
		// Construct qualified column name
		name := trace.QualifiedColumnName(ith.Module, ith.Name)
		// Apply bitwidth restrictions (if applicable)
		if bitwidth := ith.Data.BitWidth(); bitwidth < 256 {
			// For now, always assume unsigned int.
			name = fmt.Sprintf("%s@u%d", name, bitwidth)
		}
		// Write out column name
		builder.WriteString(name)
		//
		builder.WriteString("\": [")

		data := ith.Data

		for j := range data.Len() {
			if j != 0 {
				builder.WriteString(", ")
			}

			jth := data.Get(j)
			builder.WriteString(jth.String())
		}

		builder.WriteString("]")
	}
	//
	builder.WriteString("}")
	// Done
	return builder.String()
}
