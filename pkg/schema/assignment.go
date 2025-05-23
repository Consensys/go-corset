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
	tr "github.com/consensys/go-corset/pkg/trace"
	"github.com/consensys/go-corset/pkg/util"
)

// Assignment represents a schema element which declares one or more columns
// whose values are "assigned" from the results of a computation.  An assignment
// is a column group which, additionally, can provide information about the
// computation (e.g. which columns it depends upon, etc).
type Assignment interface {
	util.Boundable

	// ComputeColumns computes the values of columns defined by this assignment.
	// In order for this computation to makes sense, all columns on which this
	// assignment depends must exist (e.g. are either inputs or have been
	// computed already).  Computed columns do not exist in the original trace,
	// but are added during trace expansion to form the final trace.
	ComputeColumns(tr.Trace) ([]tr.ArrayColumn, error)

	// Returns the set of columns that this assignment depends upon.  That can
	// include both input columns, as well as other computed columns.
	Dependencies() []uint
}
