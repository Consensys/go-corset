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
package coverage

import (
	"math"

	sc "github.com/consensys/go-corset/pkg/schema"
)

// ConstraintGroup represents a set of constraints.  The intuition is that
// metrics are reported over groups of constraints (e.g. all constraints in a
// given module).  Of course, the if group contains only a single constraint
// then you get results just for that.
type ConstraintGroup struct {
	// Identifier for module enclosing this constraint.
	ModuleId uint
	// Name of this constraint.
	Name string
	// Case number of this constraint.
	Case uint
}

// NewModuleGroup constructs a new group representing all constraints in a given
// module.
func NewModuleGroup(mid uint) ConstraintGroup {
	return ConstraintGroup{mid, "*", math.MaxUint}
}

// NewConstraintGroup constructs a new group representing an (unexpanded)
// constraint in a given module.
func NewConstraintGroup(mid uint, name string) ConstraintGroup {
	return ConstraintGroup{mid, name, math.MaxUint}
}

// NewIndividualConstraintGroup constructs a new group representing an
// individual constraint.
func NewIndividualConstraintGroup(mid uint, name string, num uint) ConstraintGroup {
	return ConstraintGroup{mid, name, num}
}

// Matches determines whether or not a given constraint should be considered
// part of this group.
func (p *ConstraintGroup) Matches(constraint sc.Constraint) bool {
	mid := constraint.Contexts()[0].Module()
	name, num := constraint.Name()
	//
	if mid != p.ModuleId {
		return false
	} else if p.Name != "*" && p.Name != name {
		return false
	} else if p.Case != math.MaxUint && p.Case != num {
		return false
	}
	// Done
	return true
}

// Select all constraints matching this group.
func (p *ConstraintGroup) Select(schema sc.Schema, filter Filter) []sc.Constraint {
	var constraints []sc.Constraint
	//
	for iter := schema.Constraints(); iter.HasNext(); {
		ith := iter.Next()
		//
		if p.Matches(ith) && filter(ith, schema) {
			constraints = append(constraints, ith)
		}
	}
	//
	return constraints
}
