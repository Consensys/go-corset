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

// Map defines an abstract notion of mapping a variable identifier to a
// variable description.
type Map interface {
	// Get the descriptor for a given variable.
	Variable(Id) Descriptor
}

// ArrayMap constructs a variable map from an array of variables.
func ArrayMap(vars ...Descriptor) Map {
	return &arrayMap{vars}
}

// arrayMap constructs a variable map from an array of variable declarations.
type arrayMap struct {
	vars []Descriptor
}

// Variable implementation for Map interface
func (p arrayMap) Variable(id Id) Descriptor {
	return p.vars[id]
}
