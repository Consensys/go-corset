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

// CoverageKey unique identifiers a constraint within the system.
type CoverageKey struct {
	// Identifier for the enclosing module.
	Module uint
	// Name of the constraint
	Name string
}

// CoverageMap is a simple datatype (for now) which associates each
// constraint with a given set of covered branches.
type CoverageMap struct {
	items map[CoverageKey][]bit.Set
}

// NewBranchCoverage constructs an empty branch coverage set.
func NewBranchCoverage() CoverageMap {
	items := make(map[CoverageKey][]bit.Set)
	return CoverageMap{items}
}

// IsEmpty checks whether or not this map is empty.
func (p *CoverageMap) IsEmpty() bool {
	return len(p.items) == 0
}

// CoverageOf returns, for a given constraint, the recorded coverage data.
func (p *CoverageMap) CoverageOf(module uint, name string) []bit.Set {
	return p.items[CoverageKey{module, name}]
}

// Record some raw coverage data into this set.
func (p *CoverageMap) Record(module uint, name string, casenum uint, nData bit.Set) {
	var (
		entry    = CoverageKey{module, name}
		items, _ = p.items[entry]
	)
	// Ensure enuough items
	for uint(len(items)) <= casenum {
		items = append(items, bit.Set{})
	}
	// Include specifc case coverage
	items[casenum].Union(nData)
	// Assign back.
	p.items[entry] = items
}

// Union entries from another (compatible) set of branch coverage data.
func (p *CoverageMap) Union(other CoverageMap) {
	for key, bitsets := range other.items {
		for i, entry := range bitsets {
			p.Record(key.Module, key.Name, uint(i), entry)
		}
	}
}

// KeysOf returns the set of constraints for which coverage data has been
// obtained.
func (p *CoverageMap) KeysOf(mid uint) *set.SortedSet[string] {
	keys := set.NewSortedSet[string]()
	//
	for k := range p.items {
		if k.Module == mid {
			keys.Insert(k.Name)
		}
	}
	//
	return keys
}

// ToJson returns a representation of this coverage map suitable for being
// converted into JSON.
func (p *CoverageMap) ToJson(schema Schema) map[string][]uint {
	var json = make(map[string][]uint)
	//
	for k, bitsets := range p.items {
		for i, cov := range bitsets {
			name := jsonConstraintName(k.Module, k.Name, uint(i), schema)
			json[name] = cov.Iter().Collect()
		}
	}
	//
	return json
}

func jsonConstraintName(mid uint, name string, casenum uint, schema Schema) string {
	mod := schema.Modules().Nth(mid)
	name = fmt.Sprintf("%s#%d", name, casenum)
	//
	if mod.Name == "" {
		return name
	}
	//
	return fmt.Sprintf("%s.%s", mod.Name, name)
}
