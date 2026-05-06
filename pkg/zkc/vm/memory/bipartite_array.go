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
package memory

import (
	"github.com/consensys/go-corset/pkg/schema/register"
	"github.com/consensys/go-corset/pkg/zkc/vm/word"
)

// HALF_START is the smallest absolute word position belonging to the upper
// partition: positions in [0,HALF_START) go to the lower partition, whilst
// positions in [HALF_START,...] go to the upper partition.  This is fixed at the
// midpoint of the uint64 address space, regardless of the actual address-tuple
// width, so that the partition decision is a simple constant comparison.
const HALF_START uint64 = ^uint64(0) / 2

// TOP_POS is the largest absolute word position.  The upper partition is
// indexed from the top, so upper[j] corresponds to absolute position
// TOP_POS-j.  Like HALF_START, this is fixed at the top of the uint64 address
// space, regardless of the actual address-tuple width.
const TOP_POS uint64 = ^uint64(0)

// BiPartiteArray provides a read/write implementation of Memory optimised for
// representing the kind of split heap/stack memory found in typical compute
// architectures (e.g. RISC-V).  Here, memory is partitioned in two: the lower
// partition and the upper partition.  Here, the lower partition represents
// memory locations starting from the least addressable location (i.e. address
// 0), whilst upper represents memory locations upto the maximal addressable
// location.  We can view memory as follows:
//
// +-----------------+ ................. +-----------------+
// | lower partition |  (unallocated)    | upper partiaion |
// +-----------------+ ................. + ----------------+
//
//	0                                       n
//
// Here, n represents the largest addressable location (i.e. n==2^64-1). In
// between the two partions is a chunk of currently unallocated memory.  Thus,
// we see that as locations are read / written the two partitions move towards
// each other.  For simplicity we simply assume that any read / write to
// location l where l < n/2 is for the lower partiion, other its for the upper
// partition.
type BiPartiteArray[W word.Word[W]] struct {
	geometry Geometry[W]
	name     string
	// Lower and upper partitions
	lower, upper []W
}

// NewBiPartiteArray constructs an empty bipartite read/write memory.
func NewBiPartiteArray[W word.Word[W]](name string, registers []register.Register) *BiPartiteArray[W] {
	return &BiPartiteArray[W]{
		geometry: NewGeometry[W](registers),
		name:     name,
	}
}

// Name implementation for Memory interface.
func (p *BiPartiteArray[W]) Name() string {
	return p.name
}

// Geometry implementation for Memory interface.
func (p *BiPartiteArray[W]) Geometry() Geometry[W] {
	return p.geometry
}

// Initialise implementation for Memory interface.  The provided contents
// populate the lower partition; the upper partition is cleared.
func (p *BiPartiteArray[W]) Initialise(contents []W) {
	p.lower = contents
	p.upper = nil
}

// Read implementation for Memory interface.
func (p *BiPartiteArray[W]) Read(frame []W, address []register.Id, data []register.Id) error {
	var start, _ = p.geometry.FrameDecode(frame, address)
	//
	if start < HALF_START {
		for i := range data {
			frame[data[i].Unwrap()] = p.readLower(start + uint64(i))
		}
	} else {
		// Cap addressable cells at TOP_POS-start+1; positions beyond TOP_POS
		// are out of range and yield zero (avoids relying on uint64
		// wraparound in start+i).
		var (
			needed = TOP_POS - start + 1
			zero   W
		)
		//
		for i := range data {
			if uint64(i) < needed {
				frame[data[i].Unwrap()] = p.readUpper(start + uint64(i))
			} else {
				frame[data[i].Unwrap()] = zero
			}
		}
	}
	//
	return nil
}

// Write implementation for Memory interface.
func (p *BiPartiteArray[W]) Write(frame []W, address []register.Id, data []register.Id) error {
	var start, end = p.geometry.FrameDecode(frame, address)
	//
	if start < HALF_START {
		// extend lower partition if needed
		p.lower = expand(p.lower, end)
		// copy over values
		for i := range data {
			p.lower[start+uint64(i)] = frame[data[i].Unwrap()]
		}
	} else {
		// In upper, the largest slice index touched is TOP_POS-start (when
		// i==0) so the upper partition must have at least TOP_POS-start+1
		// elements.
		var needed = TOP_POS - start + 1
		// extend upper partition if needed
		p.upper = expand(p.upper, needed)
		// Cap iteration at `needed`: any cell whose position would exceed
		// TOP_POS lies outside the addressable range (start+i would wrap
		// uint64) and is silently dropped, mirroring the zero returned by
		// readUpper for the same positions.
		n := min(uint64(len(data)), needed)
		//
		for i := range n {
			p.upper[TOP_POS-(start+i)] = frame[data[i].Unwrap()]
		}
	}
	//
	return nil
}

// readLower returns the word at the given absolute position in the lower
// partition, returning zero for out-of-bounds accesses.
func (p *BiPartiteArray[W]) readLower(pos uint64) W {
	var zero W
	//
	if pos < uint64(len(p.lower)) {
		return p.lower[pos]
	}
	//
	return zero
}

// readUpper returns the word at the given absolute position in the upper
// partition, returning zero for out-of-bounds accesses.
func (p *BiPartiteArray[W]) readUpper(pos uint64) W {
	var (
		idx  = TOP_POS - pos
		zero W
	)

	//
	if idx < uint64(len(p.upper)) {
		return p.upper[idx]
	}
	//
	return zero
}
