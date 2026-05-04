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

import "github.com/consensys/go-corset/pkg/zkc/vm/word"

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
