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
package agnostic

// GF_251 is teany tiny prime field used exclusively for testing.
var GF_251 = FieldConfig{"GF_251", 7, 4}

// GF_8209 is small prime field used exclusively for testing.
var GF_8209 = FieldConfig{"GF_8209", 13, 8}

// BLS12_377 is the defacto default field at this time.
var BLS12_377 = FieldConfig{"BLS12_377", 252, 160}

// FIELD_CONFIGS determines the set of supported fields.
var FIELD_CONFIGS = []FieldConfig{
	GF_251,
	GF_8209,
	BLS12_377,
}

// FieldConfig provides a simple mechanism for configuring the field agnosticity
// pipeline.
type FieldConfig struct {
	// Name suitable for identifying the config.  This is only really used for
	// improving error reporting, etc.
	Name string
	// Maximum field bandwidth available in the field.
	FieldBandWidth uint
	// Maximum register width to use with this field.
	RegisterWidth uint
}

// GetFieldConfig returns the field configuration corresponding with the given
// name, or nil no such config exists.
func GetFieldConfig(name string) *FieldConfig {
	for i := range FIELD_CONFIGS {
		if FIELD_CONFIGS[i].Name == name {
			return &FIELD_CONFIGS[i]
		}
	}
	//
	return nil
}
