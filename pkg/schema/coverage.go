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

// Coverage provides branch coverage information for a given constraint.
// Specifically, the number of unique branches coverage along with the total
// number of branches.
type Coverage struct {
	// Number of branches Covered
	Covered bit.Set
	// Total number of branches
	Total uint
}

// NewCoverage constructs a new piece of coverage data.
func NewCoverage(covered bit.Set, total uint) Coverage {
	return Coverage{covered, total}
}

// CoverageMap is a simple datatype (for now) which associates each
// constraint with a given set of covered branches.
type CoverageMap struct {
	items map[string]Coverage
}

// NewBranchCoverage constructs an empty branch coverage set.
func NewBranchCoverage() CoverageMap {
	items := make(map[string]Coverage)
	return CoverageMap{items}
}

// IsEmpty checks whether or not this map is empty.
func (p *CoverageMap) IsEmpty() bool {
	return len(p.items) == 0
}

// CoverageOf returns, for a given constraint, the recorded coverage data.
func (p *CoverageMap) CoverageOf(name string) Coverage {
	return p.items[name]
}

// Insert some raw coverage data into this set.
func (p *CoverageMap) Insert(name string, data Coverage) {
	p.items[name] = data
}

// InsertAll entries from another (compatible) set of branch coverage data.
func (p *CoverageMap) InsertAll(other CoverageMap) {
	for k, v := range other.items {
		var res bit.Set
		// Check whether already record for this item
		if data, ok := p.items[k]; ok {
			res.InsertAll(data.Covered)
			// Sanity check
			if data.Total != v.Total {
				msg := fmt.Sprintf("inconsistent branch count for %s (%d vs %d)", k, data.Total, v.Total)
				panic(msg)
			}
		}
		//
		res.InsertAll(v.Covered)
		//
		p.items[k] = Coverage{res, v.Total}
	}
}

// Keys returns the set of constraints for which coverage data has been
// obtained.
func (p *CoverageMap) Keys() *set.SortedSet[string] {
	keys := set.NewSortedSet[string]()
	//
	for k, _ := range p.items {
		keys.Insert(k)
	}
	//
	return keys
}

// ToJson returns a representation of this coverage map suitable for being
// converted into JSON.
func (p *CoverageMap) ToJson() map[string]any {
	var json map[string]any = make(map[string]any)
	//
	keys := p.Keys()
	// Print out the data
	for iter := keys.Iter(); iter.HasNext(); {
		ith := iter.Next()
		cov := p.CoverageOf(ith)
		json[ith] = cov.Covered.Iter().Collect()
	}
	//
	return json
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
