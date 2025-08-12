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
package field

import (
	"math/rand"
	"testing"

	"github.com/consensys/go-corset/pkg/util/assert"
	"github.com/consensys/go-corset/pkg/util/field/bls12_377"
	"github.com/consensys/go-corset/pkg/util/field/koalabear"
)

func init() {
	// make sure the interface is adhered to.
	_ = Element[koalabear.Element](koalabear.Element{})
	_ = Element[bls12_377.Element](bls12_377.Element{})
}

func TestBatchInvert(t *testing.T) {
	s := make([]koalabear.Element, 4000)
	sInv := make([]koalabear.Element, len(s))
	scratch := make([]koalabear.Element, len(s))

	for i := range s {
		s[i] = koalabear.Element{rand.Uint32()}
		if s[i][0] >= koalabear.Modulus {
			s[i][0] = 0 // getting a zero with considerable probability
		}

		sInv[i] = s[i].Inverse()

		copy(scratch[:i], s)
		BatchInvert(scratch[:i])

		for j := range i {
			assert.Equal(t, sInv[j][0], scratch[j][0], "on slice %v, at index %d", s[:i], j)
		}
	}
}
