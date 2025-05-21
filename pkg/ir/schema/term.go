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
	"github.com/consensys/go-corset/pkg/util"
)

// Term represents a component of an AIR expression.
type Term[T any] interface {
	Contextual
	Evaluable
	Lispifiable
	util.Boundable

	// ApplyShift applies a given shift to all variable accesses in a given term
	// by a given amount. This can be used to normalise shifting in certain
	// circumstances.
	ApplyShift(int) T

	// ShiftRange returns the minimum and maximum shift value used anywhere in
	// the given term.
	ShiftRange() (int, int)

	// Simplify constant expressions down to single values.  For example, "(+ 1
	// 2)" would be collapsed down to "3".  This is then progagated throughout
	// an expression, so that e.g. "(+ X (+ 1 2))" becomes "(+ X 3)"", etc.
	// There is also an option to retain casts, or not.
	Simplify(casts bool) T

	// ValueRange returns the interval of values that this term can evaluate to.
	// For terms accessing columns, this is determined by the declared width of
	// the column.
	ValueRange(module Module) *util.Interval
}

type LogicalTerm[T any] interface {
	Contextual
	Lispifiable
	Testable
}
