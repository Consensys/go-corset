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
package binfile

import (
	"fmt"

	"github.com/consensys/go-corset/pkg/schema"
)

type jsonType struct {
	// Determines the representation of this type.  For example, a
	// 8bit unsigned integer.
	Magma any `json:"m"`
	// Determines the interpretation of this type.  Specifically,
	// for binary types, we can have an interpretation of either
	// bool (where 0 is false and anything else is true) or loob
	// (where 0 is true and anything else is false).
	Conditioning string `json:"c"`
}

// =============================================================================
// Translation
// =============================================================================

func (e *jsonType) toHir() schema.Type {
	// Check whether magma is string
	if str, ok := e.Magma.(string); ok {
		switch str {
		case "Native":
			return &schema.FieldType{}
		case "Byte":
			return schema.NewUintType(8)
		case "Binary":
			return schema.NewUintType(1)
		default:
			panic(fmt.Sprintf("Unknown JSON type encountered: %s:%s", e.Magma, e.Conditioning))
		}
	}
	// Try as integer
	if intMap, ok := e.Magma.(map[string]any); ok {
		if val, isInt := intMap["Integer"]; isInt {
			nbits := uint(val.(float64))
			return schema.NewUintType(nbits)
		}
	}
	// Fail
	panic(fmt.Sprintf("Unknown JSON type encountered: %s:%s", e.Magma, e.Conditioning))
}
