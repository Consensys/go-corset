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
	"github.com/consensys/go-corset/pkg/util/math"
	"github.com/consensys/go-corset/pkg/zkc/compiler/ast/variable"
)

// Sub represents an expresion which subtracts zero or more terms from a given term.
type Sub struct {
	Exprs []Expr
}

// NewSub constructs an expression representing the subtraction of one or more
// values.
func NewSub(exprs ...Expr) Expr {
	if len(exprs) == 0 {
		panic("one or more subexpressions required")
	}
	//
	return &Sub{Exprs: exprs}
}

// Equals implementation for the Expr interface.
func (p *Sub) Equals(e Expr) bool {
	if e, ok := e.(*Sub); ok {
		return EqualsAll(p.Exprs, e.Exprs)
	}
	//
	return false
}

// Uses implementation for the Expr interface.
func (p *Sub) Uses() bit.Set {
	var reads bit.Set
	//
	for _, e := range p.Exprs {
		reads.Union(e.Uses())
	}
	//
	return reads
}

// ValueRange implementation for the Expr interface.
func (p *Sub) ValueRange(env variable.Map) math.Interval {
	var values math.Interval
	//
	for i, e := range p.Exprs {
		if i == 0 {
			values = e.ValueRange(env)
		} else {
			values.Sub(e.ValueRange(env))
		}
	}
	//
	return values
}

func (p *Sub) String(mapping variable.Map) string {
	return String(p, mapping)
}
