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
package ranged

import (
	"fmt"

	"github.com/consensys/go-corset/pkg/ir"
	"github.com/consensys/go-corset/pkg/schema"
	"github.com/consensys/go-corset/pkg/trace"
	"github.com/consensys/go-corset/pkg/util/collection/set"
)

// Failure provides structural information about a failing type constraint.
type Failure struct {
	// Handle of the failing constraint
	Handle string
	// Enclosing context
	Context schema.ModuleId
	// Constraint expression
	Expr ir.Evaluable
	// Range restriction
	Bitwidth uint
	// Row on which the constraint failed
	Row uint
}

// Message provides a suitable error message
func (p *Failure) Message() string {
	// Construct useful error message
	return fmt.Sprintf("range \"%s\" is u%d does not hold (row %d)", p.Handle, p.Bitwidth, p.Row)
}

func (p *Failure) String() string {
	return p.Message()
}

// RequiredCells identifies the cells required to evaluate the failing constraint at the failing row.
func (p *Failure) RequiredCells(tr trace.Trace) *set.AnySortedSet[trace.CellRef] {
	return p.Expr.RequiredCells(int(p.Row), p.Context)
}
