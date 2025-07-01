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

// Packetize takes a given polynomial and breaks it into a sequence of one or
// more packets, each of which occupies at most a given number of bits.
// Furthermore, returned packets are sorted according to their starting bit.
func Packetize(p Polynomial, bitwidth uint) []Packet {
	panic("todo")
}

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
