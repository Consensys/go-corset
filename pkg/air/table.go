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
	"github.com/consensys/go-corset/pkg/trace"
	"github.com/consensys/go-corset/pkg/util/collection/iter"
)

// table columns support assignments?
type Table[C Constraint] struct {
}

// Module name
func (p *Table[C]) Name() string {
	panic("todo")
}

// Access a given column in this module.
func (p *Table[C]) Column(uint) sc.Column {
	panic("todo")
}

// Columns returns an iterator over the underlying columns of this schema.
// Specifically, the index of a column in this array is its column index.
func (p *Table[C]) Columns() iter.Iterator[sc.Column] {
	panic("todo")
}

// Constraints returns an iterator over the underlying constraints of this
// schema.
func (p *Table[C]) Constraints() iter.Iterator[C] {
	panic("todo")
}

// Returns the number of columns in this module.
func (p *Table[C]) Width() uint {
	panic("todo")
}

func (p *Table[C]) AddColumn(context trace.Context, name string, datatype sc.Type) uint {
	panic("todo")
}

func (p *Table[C]) AddAssignment(c sc.Assignment) uint {
	panic("todo")
}

func (p *Table[C]) AddConstraint(c Constraint) uint {
	panic("todo")
}
