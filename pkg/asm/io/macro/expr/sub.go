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
	"math/big"

	"github.com/consensys/go-corset/pkg/schema"
	"github.com/consensys/go-corset/pkg/schema/agnostic"
	"github.com/consensys/go-corset/pkg/util/collection/bit"
	"github.com/consensys/go-corset/pkg/util/math"
)

// Sub represents an expresion which subtracts zero or more terms from a given term.
type Sub struct {
	Exprs []Expr
}

// Eval implementation for the Expr interface.
func (p *Sub) Eval(env []big.Int) big.Int {
	var result big.Int
	//
	for i, e := range p.Exprs {
		ith := e.Eval(env)

		if i == 0 {
			result.Set(&ith)
		} else {
			result.Sub(&result, &ith)
		}
	}
	// Done
	return result
}

// Polynomial implementation for the Expr interface.
func (p *Sub) Polynomial() agnostic.Polynomial {
	var result agnostic.Polynomial
	//
	for i, e := range p.Exprs {
		ith := e.Polynomial()
		//
		if i == 0 {
			result = ith
		} else {
			result = result.Sub(ith)
		}
	}
	//
	return result
}

// RegistersRead implementation for the Expr interface.
func (p *Sub) RegistersRead() bit.Set {
	var reads bit.Set
	//
	for _, e := range p.Exprs {
		reads.Union(e.RegistersRead())
	}
	//
	return reads
}

// ValueRange implementation for the Expr interface.
func (p *Sub) ValueRange(mapping schema.RegisterMap) math.Interval {
	var values math.Interval
	//
	for i, e := range p.Exprs {
		if i == 0 {
			values = e.ValueRange(mapping)
		} else {
			values.Sub(e.ValueRange(mapping))
		}
	}
	//
	return values
}

func (p *Sub) String(mapping schema.RegisterMap) string {
	return String(p, mapping)
}
