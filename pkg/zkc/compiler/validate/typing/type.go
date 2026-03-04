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
package typing

import (
	"math/big"

	"github.com/consensys/go-corset/pkg/zkc/compiler/ast/data"
	"github.com/consensys/go-corset/pkg/zkc/compiler/ast/variable"
)

// Type provides an abstraction over the different kinds of type which are
// possible (e.g. integer types versus struct or array types, etc).
type Type interface {
	// AsUint() attempts to convert this type into an unsigned integer type,
	// whilst returning nil if this fails.
	AsUint() *Uint
}

// From constructs a suitable type from a given data type.
func From(datatype data.Type) Type {
	//
	var (
		bound    = big.NewInt(2)
		bitwidth = datatype.BitWidth()
	)
	// compute 2^bitwidth
	bound.Exp(bound, big.NewInt(int64(bitwidth)), nil)
	// Subtract 1 because interval is inclusive.
	bound.Sub(bound, big.NewInt(1))
	//
	return &Uint{MaxValue: *bound}
}

// FromVariables constructs a type from a given set of variable descriptors.
func FromVariables(variables ...variable.Descriptor) Type {
	var (
		types = make([]Type, len(variables))
	)
	//
	for i, v := range variables {
		types[i] = From(v.DataType)
	}
	//
	if len(types) == 1 {
		return types[0]
	}
	//
	panic("todo")
}
