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
	"github.com/consensys/go-corset/pkg/trace"
	"github.com/consensys/go-corset/pkg/util/collection/set"
)

// Failure embodies structured information about a failing constraint.
// This includes the constraint itself, along with the row
type Failure interface {
	// Provides a suitable error message
	Message() string
}

// InternalFailure is a generic mechanism for reporting failures, particularly
// as arising from evaluation of a given expression.
type InternalFailure struct {
	// Handle of the failing constraint
	Handle string
	// Row on which the constraint failed
	Row uint
	// Cells involved (if any)
	Term Contextual
	// Error message
	Error string
}

// Message provides a suitable error message
func (p *InternalFailure) Message() string {
	return p.Error
}

// RequiredCells identifies the cells required to evaluate the failing constraint at the failing row.
func (p *InternalFailure) RequiredCells(trace trace.Trace) *set.AnySortedSet[trace.CellRef] {
	// if p.Term != nil {
	// 	return p.Term.RequiredCells(int(p.Row), trace)
	// }
	// // Empty set
	// return set.NewAnySortedSet[tr.CellRef]()
	panic("todo")
}
