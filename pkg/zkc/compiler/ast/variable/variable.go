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
package variable

import "github.com/consensys/go-corset/pkg/zkc/compiler/ast/data"

// Kind determines the type of a given variable (e.g. a parameter,
// return, etc).
type Kind uint8

// Id is a convenient alias to help clarity a given variable's intended
// purpose.
type Id = uint

var (
	// CONTEXT defines the type of "contextual" variables.
	CONTEXT = Kind(1)
	// PARAMETER defines the type of parameter variables.
	PARAMETER = Kind(2)
	// RETURN defines the type of return variables.
	RETURN = Kind(3)
	// LOCAL defines the type of local variables.
	LOCAL = Kind(4)
)

// Descriptor describes a variable used within a function (or other component),
// such as a parameter, return or local variable.  A descriptor contains all the
// key information about a variable, such as its name, type, etc.
type Descriptor struct {
	// Kind of variable (parameter, return, local, external)
	Kind Kind
	// Name of the variable
	Name string
	// Type of the variable
	DataType data.Type
}

// New constructs a new variable declaration for a given kind of
// variable.
func New(kind Kind, name string, datatype data.Type) Descriptor {
	return Descriptor{kind, name, datatype}
}

// IsParameter indicates whether or not this variable is function parameter.
func (p Descriptor) IsParameter() bool {
	return p.Kind == PARAMETER
}

// IsReturn indicates whether or not this variable is function return.
func (p Descriptor) IsReturn() bool {
	return p.Kind == RETURN
}

// IsLocal indicates whether or not this is a local variable.
func (p Descriptor) IsLocal() bool {
	return p.Kind == LOCAL
}

// BitWidth returns the required bitwidth of this data type.
func (p Descriptor) BitWidth() uint {
	return p.DataType.BitWidth()
}
