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
package mir

import (
	"github.com/consensys/go-corset/pkg/ir/schema"
	"github.com/consensys/go-corset/pkg/ir/schema/constraint"
)

// LookupConstraint captures the essence of a lookup constraint at the MIR
// level.
type LookupConstraint = *constraint.LookupConstraint[Expr]

// VanishingConstraint captures the essence of a vanishing constraint at the MIR
// level. A vanishing constraint is a row constraint which must evaluate to
// zero.
type VanishingConstraint = *constraint.VanishingConstraint[Logical]

// RangeConstraint captures the essence of a range constraints at the MIR level.
type RangeConstraint = *constraint.RangeConstraint[Expr]

// SortedConstraint captures the essence of a sorted constraint at the MIR
// level.
type SortedConstraint = *constraint.SortedConstraint[Expr]

// PropertyAssertion captures the notion of an arbitrary property which should
// hold for all acceptable traces.  However, such a property is not enforced by
// the prover.
type Assertion = *schema.Assertion[Logical]

type Schema = schema.Schema[*schema.Table, schema.Constraint]
