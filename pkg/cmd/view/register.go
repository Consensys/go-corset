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
package view

import (
	"math/big"

	"github.com/consensys/go-corset/pkg/schema/register"
	tr "github.com/consensys/go-corset/pkg/trace"
	"github.com/consensys/go-corset/pkg/util/field"
	"github.com/consensys/go-corset/pkg/util/math"
)

// RegisterView provides an abstract view of a given column.
type RegisterView interface {
	//IsComputed() bool
	Get(uint) big.Int
}

type registerView[F field.Element[F]] struct {
	trace    tr.Module[F]
	register register.Id
	mapping  register.LimbsMap
}

func (p *registerView[F]) Get(row uint) big.Int {
	var (
		bits  = uint(0)
		value big.Int
		limbs = p.mapping.LimbIds(p.register)
	)
	//
	for _, lid := range limbs {
		var (
			data    = p.trace.Column(lid.Unwrap()).Data()
			element = data.Get(row)
			limb    = p.mapping.Limb(lid)
			val     big.Int
		)
		// Construct value from field element
		val.SetBytes(element.Bytes())
		// Shift and add
		value.Add(&value, val.Mul(&val, math.Pow2(bits)))
		//
		bits += limb.Width
	}
	//
	return value
}
