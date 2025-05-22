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
package air

import (
	"github.com/consensys/go-corset/pkg/schema"
	sc "github.com/consensys/go-corset/pkg/schema"
	"github.com/consensys/go-corset/pkg/schema/constraint"
)

type Constraint interface {
	sc.Constraint

	IsAir() bool
}

// LookupConstraint captures the essence of a lookup constraint at the AIR
// level.  At the AIR level, lookup constraints are only permitted between
// columns (i.e. not arbitrary expressions).
type LookupConstraint = *constraint.LookupConstraint[*ColumnAccess]

// VanishingConstraint captures the essence of a vanishing constraint at the AIR level.
type VanishingConstraint = *constraint.VanishingConstraint[Expr]

// RangeConstraint captures the essence of a range constraints at the AIR level.
type RangeConstraint = *constraint.RangeConstraint[*ColumnAccess]

// PermutationConstraint captures the essence of a permutation constraint at the AIR level.
// Specifically, it represents a constraint that one (or more) columns are a permutation of another.
type PermutationConstraint = *constraint.PermutationConstraint

// PropertyAssertion captures the notion of an arbitrary property which should
// hold for all acceptable traces.  However, such a property is not enforced by
// the prover.
type PropertyAssertion = *schema.PropertyAssertion[schema.Testable]
