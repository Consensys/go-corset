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
package agnostic

import (
	"math/big"

	sc "github.com/consensys/go-corset/pkg/schema"
	"github.com/consensys/go-corset/pkg/schema/register"
)

// ApplyMapping applies a given mapping to a set of registers producing a
// corresponding set of limbs.  In essence, each register is convert to its
// limbs in turn, and these are all appended together in order of ococurence.
func ApplyMapping(mapping sc.RegisterLimbsMap, rids ...register.Id) []sc.LimbId {
	var limbs []sc.LimbId
	//
	for _, rid := range rids {
		limbs = append(limbs, mapping.LimbIds(rid)...)
	}
	//
	return limbs
}

// LimbsOf returns those limbs corresponding to a given set of identifiers.
func LimbsOf(mapping sc.RegisterLimbsMap, lids []sc.LimbId) []sc.Limb {
	var (
		limbs []sc.Limb = make([]sc.Limb, len(lids))
	)
	//
	for i, lid := range lids {
		limbs[i] = mapping.Limb(lid)
	}
	//
	return limbs
}

// IsPowerOf2 checks whether a given big integer matches 2^n for some n and, if
// so, n is returned.
func IsPowerOf2(val big.Int) (n uint, ok bool) {
	w := val.BitLen()
	//
	if w > 0 {
		m := big.NewInt(2)
		// compute 2^n-1
		m.Exp(m, big.NewInt(int64(w-1)), nil)
		// check for match
		return uint(w - 1), val.Cmp(m) == 0
	}
	//
	return 0, false
}
