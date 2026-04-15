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
package compiler

import (
	"strconv"

	"github.com/consensys/go-corset/pkg/util/source"
	"github.com/consensys/go-corset/pkg/zkc/compiler/ast"
	"github.com/consensys/go-corset/pkg/zkc/compiler/ast/data"
	"github.com/consensys/go-corset/pkg/zkc/compiler/ast/decl"
	"github.com/consensys/go-corset/pkg/zkc/compiler/ast/symbol"
	"github.com/consensys/go-corset/pkg/zkc/compiler/ast/variable"
)

// Lowering is the configuration options to lower the ast program.
type Lowering struct {
	expandFixedArrays bool
}

func (p Lowering) ExpandFixedArrays(flag bool) Lowering { return Lowering{expandFixedArrays: flag} }

func (p Lowering) Lower(program ast.Program) (ast.Program, []source.SyntaxError) {
	for i, d := range program.Components() {
		switch d_type := d.(type) {
		case *decl.ResolvedConstant:
			// nothing to do
		case *decl.ResolvedMemory:
			// nothing to do
		case *decl.ResolvedFunction:
			var expandedVariables []variable.Descriptor[symbol.Resolved]
			for _, v := range d_type.Variables {
				switch v_type := v.DataType.(type) {
				case *data.ResolvedFixedArray:
					for j := range v_type.Size {
						name := v.Name + "$" + strconv.FormatUint(uint64(j), 10)
						expandedVar := variable.New(v.Kind, name, data.NewUnsignedInt[symbol.Resolved](data.BitWidthOf(v_type, program.Environment()), false))
						expandedVariables = append(expandedVariables, expandedVar)
					}
				default:
					expandedVariables = append(expandedVariables, v)

				}
			}
			d_type.Variables = expandedVariables
			program.Components()[i] = d_type
		case *decl.ResolvedTypeAlias:
			// nothing to do
		}

	}
	return program, nil
}
