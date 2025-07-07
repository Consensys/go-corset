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
	"slices"

	sc "github.com/consensys/go-corset/pkg/schema"
)

// Packet represents a section of a source polynomial which fits within a given
// bitrange.  For example, the polynomial 256.X'1 + X'0 + 1 can be broken into
// two packets: 256.X'1 which occupies bits 8..15 (inclusive); and X'0 + 1 which
// occupies bits 0..8 (inclusive).  Thus, these two packets overlap on bit 8.
//
// Packets are normalised based on their bit range.  Thus, in our example above,
// the contents of the packet representing 256.X'1 is, in fact, simple X'1 with
// the associated bit range being 8..15.  Thus, the original polynomial
// component can always be reconstructed by multiply it by 2^8 (in this case).
type Packet struct {
	// Starting bit occupied by this packet
	Start uint
	// Final bit occupied by this packet (inclusive).
	End uint
	// Contents of this packet as a polynomial.
	Contents Polynomial
}

// NewPacket constructs a new package from a given set of monomials which are
// assumed to have the same coefficient.
func NewPacket(monomials []Monomial) Packet {
	var poly Polynomial
	// FIXME: sort out start and end.
	return Packet{0, 0, poly.Set(monomials...)}
}

// Packetize takes a given polynomial and breaks it into a sequence of one or
// more packets, each of which occupies at most a given number of bits.
// Furthermore, returned packets are sorted according to their starting bit.
func Packetize(p Polynomial) []Packet {
	var (
		first     = 0
		coeff     big.Int
		packets   []Packet
		monomials = SortMonomialsByCoefficient(p)
	)
	//
	for i, m := range monomials {
		ith := m.Coefficient()
		//
		if i != 0 && coeff.Cmp(&ith) != 0 {
			// end of packet
			packets = append(packets, NewPacket(monomials[first:i]))
			// start next one
			first = i
		}
		//
		coeff = m.Coefficient()
	}
	// Last packet
	packets = append(packets, NewPacket(monomials[first:]))
	//
	return packets
}

// Extract all monomials from the given polynomial which contribute to its first
// n bits of information.  For example, consider extracting the first 8bits of
// information from the polynomial 256*X + Y + 1, where X and Y are u8
// registers.  This results in the polynomial Y+1, since 256*X cannot contribute
// to the first 8bits of information.  Observe, however, that Y+1 contributes to
// more than just the first 8bits!
func Extract(nbits uint, p Polynomial, env sc.RegisterMapping) Polynomial {
	var (
		pivot = big.NewInt(2)
		terms []Monomial
		poly  Polynomial
	)
	// determine 2^nbits
	pivot.Exp(pivot, big.NewInt(int64(nbits)), nil)
	//
	for i := range p.Len() {
		var (
			ith   = p.Term(i)
			coeff = ith.Coefficient()
		)
		// Check whether this contributes towards the first nbits.
		if pivot.Cmp(&coeff) >= 0 {
			// yes it does
			terms = append(terms, ith)
		}
	}
	// Done
	return poly.Set(terms...)
}

// SortMonomialsByCoefficient returns the mononimals of a given polynomial
// sorted into increasing order of their coefficient.
func SortMonomialsByCoefficient(p Polynomial) []Monomial {
	var monomials = make([]Monomial, p.Len())
	// Copy over monomials
	for i := range p.Len() {
		monomials[i] = p.Term(i)
	}
	// Sort them
	slices.SortFunc(monomials, func(l Monomial, r Monomial) int {
		var (
			lc = l.Coefficient()
			rc = r.Coefficient()
		)
		// Smallest first
		return lc.Cmp(&rc)
	})
	//
	return monomials
}
