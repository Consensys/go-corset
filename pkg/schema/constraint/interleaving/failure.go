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
package interleaving

import (
	"fmt"

	"github.com/consensys/go-corset/pkg/ir"
	"github.com/consensys/go-corset/pkg/schema"
	"github.com/consensys/go-corset/pkg/trace"
	"github.com/consensys/go-corset/pkg/util/collection/set"
	"github.com/consensys/go-corset/pkg/util/field/bls12_377"
)

// Failure provides structural information about a failing lookup constraint.
type Failure struct {
	// Handle of the failing constraint
	Handle string
	// Relevant context for target expressions.
	TargetContext schema.ModuleId
	// Target expression involved
	Target ir.Evaluable[bls12_377.Element]
	// Relevant context for source expressions.
	SourceContext schema.ModuleId
	// Source expression which were missing
	Source ir.Evaluable[bls12_377.Element]
	// Target row on which constraint
	Row uint
}

// Message provides a suitable error message
func (p *Failure) Message() string {
	return fmt.Sprintf("interleaving \"%s\" failed (row %d)", p.Handle, p.Row)
}

func (p *Failure) String() string {
	return p.Message()
}

// RequiredCells identifies the cells required to evaluate the failing constraint at the failing row.
func (p *Failure) RequiredCells(tr trace.Trace[bls12_377.Element]) *set.AnySortedSet[trace.CellRef] {
	var res = set.NewAnySortedSet[trace.CellRef]()
	//
	res.InsertSorted(p.Source.RequiredCells(int(p.Row), p.SourceContext))
	res.InsertSorted(p.Target.RequiredCells(int(p.Row), p.TargetContext))
	//
	return res
}
