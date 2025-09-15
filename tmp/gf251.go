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
package gf251

// N defines the modulus for the GF251 prime field.
const N = 251

// R is determined by the bitwidth used for holding field elements.  In this
// case, we store our field elements in a single byte, so the bitwidth is 8.
const R = 256

// BITWIDTH identifies the bitwidth used for holding field elements.  In this
// case, we store field elements in a single byte for efficiency.
const BITWIDTH = 8

// negInvN represents -1/N mod R.
const negInvN = 205

// Element type for the GF251 prime field.  This is defined as an array of one
// element to prevent accidental use of native arithmetic operators (+,*).  An
// Element value represents an encoded form of some integer value X.
// Specifically, for some integer X, the value stored in an Element is always
// (X*R) % N.
type Element [1]uint8

// New constructs a new field element from a given unsigned integer.  This will
// panic if the supplised value is too large.
func New(val uint8) Element {
	if val >= N {
		panic("invalid GF251 element")
	}
	// Encode our integer val into the form (val*R) % N.
	element := (uint16(val) << BITWIDTH) % N
	//
	return Element{uint8(element)}
}

// Add two elements together
func (p Element) Add(q Element) Element {
	// Add to give ((p+q)*R) % 2N
	val := uint16(p[0]) + uint16(q[0])
	// Reduce to give ((p+q)*R) % N
	if val >= N {
		val -= N
	}
	// Done
	return Element{uint8(val)}
}

// Mul multiplies two elements together
func (p Element) Mul(q Element) Element {
	// Multiply to give (p*q*R*R) mod N^2
	val := uint16(p[0]) * uint16(q[0])
	//
	return Element{reduce(val)}
}

// ToByte decodes an element into the integer value it represents.
func (p Element) ToByte() uint8 {
	return reduce(uint16(p[0]))
}

// Montgomery reduction.  Value on entry has the form (x*R) mod R*N.  The goal
// is to return the value "x mod N".
func reduce(val uint16) uint8 {
	// Divide by -N
	quot := uint8(val) * negInvN
	// fmt.Print("quot", quot, "\n")
	// Determine remainder
	rem := uint32(val) + (uint32(quot) * N)
	// fmt.Print("rem", rem, "\n")
	// Divide by R
	rem = rem >> BITWIDTH
	// fmt.Print("rem", rem, "\n")
	// Reduce to (x*R) % N.
	if rem >= N {
		rem -= N
	}
	// Done
	// fmt.Print("rem-final", rem, "\n")
	return uint8(rem)
}
