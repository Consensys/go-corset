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

// DescriptorsToType construct a single type representing a given set of
// variable descriptors.
func DescriptorsToType(vars ...Descriptor) data.Type {
	var types []data.Type = make([]data.Type, len(vars))
	//
	for i, vd := range vars {
		types[i] = vd.DataType
	}
	//
	if len(types) == 1 {
		return types[0]
	}
	// construct tuple type
	return data.NewTuple(types...)
}
