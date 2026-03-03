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
package typing

import "math/big"

// Uint captures the set of values expressable at a given position within an
// expression.
type Uint struct {
	MaxValue big.Int
}

// AsUint implementation for the Type interface
func (p *Uint) AsUint() *Uint {
	return p
}

// Add two integer types together
func (p *Uint) Add(o *Uint) *Uint {
	var res big.Int
	//
	return &Uint{*res.Add(&p.MaxValue, &o.MaxValue)}
}

// Mul two integer types together
func (p *Uint) Mul(o *Uint) *Uint {
	var res big.Int
	//
	return &Uint{*res.Mul(&p.MaxValue, &o.MaxValue)}
}

// BitWidth returns the bitwidth of this type, along with an indication of
// whether or not it is signed.
func (p *Uint) BitWidth() uint {
	return uint(p.MaxValue.BitLen())
}
