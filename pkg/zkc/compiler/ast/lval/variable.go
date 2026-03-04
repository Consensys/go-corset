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
package lval

import (
	"github.com/consensys/go-corset/pkg/util/collection/bit"
	"github.com/consensys/go-corset/pkg/util/collection/set"
	"github.com/consensys/go-corset/pkg/zkc/compiler/ast/symbol"
	"github.com/consensys/go-corset/pkg/zkc/compiler/ast/variable"
)

// Variable represents a register access within an expresion.
type Variable[I symbol.Symbol[I]] struct {
	Id variable.Id
}

// NewVariable constructs an expression representing a register access.
func NewVariable[I symbol.Symbol[I]](variable variable.Id) LVal[I] {
	return &Variable[I]{Id: variable}
}

// ExternUses implementation for the LVal interface.
func (p *Variable[I]) ExternUses() set.AnySortedSet[I] {
	return nil
}

// LocalUses implementation for the LVal interface.
func (p *Variable[I]) LocalUses() bit.Set {
	return bit.Set{}
}

// LocalDefs implementation for the LVal interface.
func (p *Variable[I]) LocalDefs() bit.Set {
	var read bit.Set
	read.Insert(p.Id)
	//
	return read
}

func (p *Variable[I]) String(mapping variable.Map) string {
	return String[I](p, mapping)
}
