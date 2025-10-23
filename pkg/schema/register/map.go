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
package register

import (
	"fmt"
	"strings"
)

// Map provides a generic interface for entities which hold information
// about registers.
type Map interface {
	fmt.Stringer
	// Name returns the name given to the enclosing entity (i.e. module or
	// function).
	Name() string
	// HasRegister checks whether a register with the given name exists and, if
	// so, returns its register identifier.  Otherwise, it returns false.
	HasRegister(name string) (Id, bool)
	// Access a given register in this module.
	Register(Id) Register
	// Registers providers access to the underlying registers of this map.
	Registers() []Register
}

// MapToString provides a default method for converting a register map
// into a simple string representation.
func MapToString(p Map) string {
	var builder strings.Builder
	//
	builder.WriteString("{")
	builder.WriteString(p.Name())
	builder.WriteString(":")
	//
	for i, r := range p.Registers() {
		if i != 0 {
			builder.WriteString(",")
		}
		//
		builder.WriteString(r.Name)
	}
	//
	builder.WriteString("}")
	//
	return builder.String()
}
