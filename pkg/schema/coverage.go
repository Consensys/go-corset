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
package schema

import (
	"fmt"

	"github.com/consensys/go-corset/pkg/util"
	"github.com/consensys/go-corset/pkg/util/collection/bit"
	"github.com/consensys/go-corset/pkg/util/collection/set"
)

// Metric provides a unique identifier to distinguish different
// evaluations of a given term.  Roughly speaking, different identifiers
// correspond to different evaluation paths through the term.
type Metric[T any] interface {
	// Include a given evalutation identifier as part of this identifier.
	Join(T) T
	// Mark this evaluation id to indicate the ith branch out of n branches was
	// taken.
	Mark(i uint, n uint) T
	// Return empty value of this metric
	Empty() T
}

// CoverageMap is a simple datatype (for now) which associates each
// constraint with a given set of covered branches.
type CoverageMap struct {
	items map[util.Pair[uint, string]]bit.Set
}

// NewBranchCoverage constructs an empty branch coverage set.
func NewBranchCoverage() CoverageMap {
	items := make(map[util.Pair[uint, string]]bit.Set)
	return CoverageMap{items}
}

// IsEmpty checks whether or not this map is empty.
func (p *CoverageMap) IsEmpty() bool {
	return len(p.items) == 0
}

// CoverageOf returns, for a given constraint, the recorded coverage data.
func (p *CoverageMap) CoverageOf(module uint, name string) bit.Set {
	return p.items[util.NewPair(module, name)]
}

// Record some raw coverage data into this set.
func (p *CoverageMap) Record(module uint, name string, nData bit.Set) {
	var (
		entry    = util.NewPair(module, name)
		oData, _ = p.items[entry]
	)
	// Include all new items
	oData.Union(nData)
	// Assign back.
	p.items[entry] = oData
}

// Union entries from another (compatible) set of branch coverage data.
func (p *CoverageMap) Union(other CoverageMap) {
	for name, data := range other.items {
		p.Record(name.Left, name.Right, data)
	}
}

// KeysOf returns the set of constraints for which coverage data has been
// obtained.
func (p *CoverageMap) KeysOf(mid uint) *set.SortedSet[string] {
	keys := set.NewSortedSet[string]()
	//
	for k := range p.items {
		if k.Left == mid {
			keys.Insert(k.Right)
		}
	}
	//
	return keys
}

// ToJson returns a representation of this coverage map suitable for being
// converted into JSON.
func (p *CoverageMap) ToJson(schema Schema) map[string][]uint {
	var json map[string][]uint = make(map[string][]uint)
	//
	for k, cov := range p.items {
		name := jsonConstraintName(k.Left, k.Right, schema)
		json[name] = cov.Iter().Collect()
	}
	//
	return json
}

func jsonConstraintName(mid uint, name string, schema Schema) string {
	mod := schema.Modules().Nth(mid)

	if mod.Name == "" {
		return name
	}
	//
	return fmt.Sprintf("%s.%s", mod.Name, name)
}

// ============================================================================
// NoMetric
// ============================================================================

// NoMetric is simply an implementation of Metric which does nothing, and costs
// nothing.  This should be used when evaluation metrics are not required and,
// hence, there should be no associated overhead.
type NoMetric struct {
}

// Join includes a given evalutation identifier as part of this identifier.
func (p NoMetric) Join(NoMetric) NoMetric {
	// do nothing
	return p
}

// Mark this evaluation id to indicate the ith branch out of n branches was
// taken.
func (p NoMetric) Mark(i uint, n uint) NoMetric {
	// do nothing
	return p
}

// Empty returns an initial (empty) value for this metric.
func (p NoMetric) Empty() NoMetric {
	// do nothing
	return p
}

// ============================================================================
// BranchMetric
// ============================================================================

// BranchMetric identifies a particular evaluation "branch" out of a given
// number of possible evaluiation branches.  Here, an evaluation branch
// identifies a particular evaluation path through a given term.
type BranchMetric struct {
	branch      uint
	branchBound uint
}

// EmptyBranchMetric constructs a new branch metric which indicates 1 of 1 paths
// taken.
func EmptyBranchMetric() BranchMetric {
	return BranchMetric{0, 1}
}

// Empty returns an initial (empty) value for this metric.
func (p BranchMetric) Empty() BranchMetric {
	return EmptyBranchMetric()
}

// Join includes a given evalutation identifier as part of this identifier.
func (p BranchMetric) Join(other BranchMetric) BranchMetric {
	p.branchBound *= other.branchBound
	p.branch = (p.branch * other.branchBound) + other.branch

	return p
}

// Mark this evaluation id to indicate the ith branch out of n branches was
// taken.
func (p BranchMetric) Mark(i uint, n uint) BranchMetric {
	p.branchBound *= n
	p.branch = (p.branch * n) + i
	//
	return p
}

// Key returns a unique value identifying a given evaluation path through a
// constraint.
func (p BranchMetric) Key() uint {
	return p.branch
}

// Branches returns the number of potential branches encountered during this evaluation.
func (p BranchMetric) Branches() uint {
	return p.branchBound
}
