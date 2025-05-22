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
	sc "github.com/consensys/go-corset/pkg/schema"
	"github.com/consensys/go-corset/pkg/schema/assignment"
	"github.com/consensys/go-corset/pkg/schema/constraint"
	"github.com/consensys/go-corset/pkg/trace"
	"github.com/consensys/go-corset/pkg/util/collection/iter"
)

// DataColumn captures the essence of a data column at AIR level.
type DataColumn = *assignment.DataColumn

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
type PropertyAssertion = *sc.PropertyAssertion[sc.Testable]

type Schema = *sc.TableSchema[*Module]

type Constraint interface {
	sc.Constraint

	IsAir() bool
}

type Module struct {
	table sc.TableModule[Constraint]
}

// Module name
func (p *Module) Name() string {
	return p.table.Name()
}

// Assertions returns an iterator over the property assertions of this
// schema.  These are properties which should hold true for any valid trace
// (though, of course, may not hold true for an invalid trace).
func (p *Module) Assertions() iter.Iterator[sc.Constraint] {
	panic("todo")
}

// Access a given column in this module.
func (p *Module) Column(uint) sc.Column {
	panic("todo")
}

// Columns returns an iterator over the underlying columns of this schema.
// Specifically, the index of a column in this array is its column index.
func (p *Module) Columns() iter.Iterator[sc.Column] {
	panic("todo")
}

// Constraints returns an iterator over the underlying constraints of this
// schema.
func (p *Module) Constraints() iter.Iterator[Constraint] {
	panic("todo")
}

// Returns the number of columns in this module.
func (p *Module) Width() uint {
	panic("todo")
}

func (p *Module) AddColumn(context trace.Context, name string, datatype sc.Type) uint {
	panic("todo")
}

func (p *Module) AddConstraint(c sc.Constraint) uint {
	panic("todo")
}
