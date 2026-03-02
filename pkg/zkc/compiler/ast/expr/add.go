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
package expr

import (
	"github.com/consensys/go-corset/pkg/util/collection/bit"
	"github.com/consensys/go-corset/pkg/util/collection/set"
	"github.com/consensys/go-corset/pkg/util/math"
	"github.com/consensys/go-corset/pkg/zkc/compiler/ast/symbol"
	"github.com/consensys/go-corset/pkg/zkc/compiler/ast/variable"
)

// Add represents an expresion which adds one or more terms together.
type Add[I symbol.Symbol[I]] struct {
	Exprs []Expr[I]
}

// NewAdd constructs an expression representing the sum of one or more values.
func NewAdd[I symbol.Symbol[I]](exprs ...Expr[I]) Expr[I] {
	if len(exprs) == 0 {
		panic("one or more subexpressions required")
	}
	//
	return &Add[I]{Exprs: exprs}
}

// NonLocalUses implementation for the Expr interface.
func (p *Add[I]) NonLocalUses() set.AnySortedSet[I] {
	return nonLocalUses(p.Exprs...)
}

// LocalUses implementation for the Expr interface.
func (p *Add[I]) LocalUses() bit.Set {
	return localUses(p.Exprs...)
}

// ValueRange implementation for the Expr interface.
func (p *Add[I]) ValueRange(env variable.Map) math.Interval {
	var values math.Interval
	//
	for i, e := range p.Exprs {
		if i == 0 {
			values = e.ValueRange(env)
		} else {
			values.Add(e.ValueRange(env))
		}
	}
	//
	return values
}

func (p *Add[I]) String(mapping variable.Map) string {
	return String[I](p, mapping)
}
