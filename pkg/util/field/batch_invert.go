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
	"github.com/consensys/go-corset/pkg/util/collection/bit"
)

// BatchInvert efficiently inverts the list of elements s, in place.
func BatchInvert[T Element[T]](s []T) {
	if len(s) == 0 {
		return
	}
	//
	var (
		zero = Zero[T]()
		one  = One[T]()
		// identifies entries which are zero
		isZero = bit.NewSet(len(s))

		m = make([]T, len(s)) // m[i] = s[i] * s[i+1] * ...
	)
	//
	isZero.Set(len(s)-1, s[len(s)-1].IsZero())

	if isZero.Get(len(s) - 1) {
		s[len(s)-1] = one
	}

	m[len(s)-1] = s[len(s)-1]

	for i := len(s) - 2; i >= 0; i-- {
		isZero.Set(i, s[i].IsZero())

		if isZero.Get(i) {
			s[i] = one
		}

		m[i] = m[i+1].Mul(s[i])
	}

	inv := m[0].Inverse() // inv = s[0]⁻¹ * s[1]⁻¹ * ...

	for i := range len(s) - 1 {
		// inv = s[i]⁻¹ * s[i+1]⁻¹ * ...
		s[i], inv = inv.Mul(m[i+1]), inv.Mul(s[i])
		// inv = s[i+1]⁻¹ * s[i+2]⁻¹ * ...
		if isZero.Get(i) {
			s[i] = zero
		}
	}

	s[len(s)-1] = inv
	if isZero.Get(len(s) - 1) {
		s[len(s)-1] = zero
	}
}
