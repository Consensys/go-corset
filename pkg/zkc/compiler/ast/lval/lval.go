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

// LVal represents an arbitrary expression used within an instruction.
type LVal[I symbol.Symbol[I]] interface {
	// ExternUses returns the set of non-local declarations accessed by this
	// expression.  For example, external constants or memories used within.
	ExternUses() set.AnySortedSet[I]
	// RegistersRead returns the set of variables used (i.e. read) by this expression
	LocalUses() bit.Set
	// LocalDefs returns the set of local variables which assigned (either
	// fully or in part) by this expression.
	LocalDefs() bit.Set
	// String returns a string representation of this expression.
	String(mapping variable.Map) string
}

// Uses determines the (unique) set of registers read by any expression
// in the given set of expressions.
func Uses[I symbol.Symbol[I]](exprs ...LVal[I]) []variable.Id {
	var (
		reads []variable.Id
		bits  bit.Set
	)
	// extract all usages
	for _, e := range exprs {
		bits.Union(e.LocalUses())
	}
	// Collect them all up
	for iter := bits.Iter(); iter.HasNext(); {
		next := iter.Next()
		//
		reads = append(reads, next)
	}
	//
	return reads
}

// Definitions determines the (unique) set of registers written by a given set
// of lvals.
func Definitions[I symbol.Symbol[I]](lvals ...LVal[I]) []variable.Id {
	var (
		reads []variable.Id
		bits  bit.Set
	)
	// extract all usages
	for _, lv := range lvals {
		bits.Union(lv.LocalDefs())
	}
	// Collect them all up
	for iter := bits.Iter(); iter.HasNext(); {
		next := iter.Next()
		//
		reads = append(reads, next)
	}
	//
	return reads
}

// String provides a generic facility for converting an expression into a
// suitable string.
func String[I symbol.Symbol[I]](e LVal[I], mapping variable.Map) string {
	panic("todo")
}
