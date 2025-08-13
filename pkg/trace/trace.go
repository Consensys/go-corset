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
package trace

import (
	"fmt"

	"github.com/consensys/go-corset/pkg/util/collection/array"
	"github.com/consensys/go-corset/pkg/util/collection/iter"
	"github.com/consensys/go-corset/pkg/util/field"
	"github.com/consensys/go-corset/pkg/util/word"
)

// Trace describes a set of named columns.  Columns are not required to have the
// same height and can be either "data" columns or "computed" columns.
type Trace[F any] interface {
	// Access a given column difrtly via a reference.
	Column(ColumnRef) Column[F]
	// Access a given module in this trace.
	Module(ModuleId) Module[F]
	// Determine whether this trace has a module with the given name and, if so,
	// what its module index is.
	HasModule(name string) (uint, bool)
	// Returns the number of modules in this trace.
	Width() uint
	// Returns an iterator over the contained modules
	Modules() iter.Iterator[Module[F]]
	// Provides access to the internal memory pool
	Pool() word.Pool[uint, F]
}

// Module describes a module within the trace.  Every module is composed of some
// number of columns, and has a specific height.
type Module[T any] interface {
	// Module name
	Name() string
	// Access a given column in this module.
	Column(uint) Column[T]
	// Access a given column by its name.
	ColumnOf(string) Column[T]
	// Returns the number of columns in this module.
	Width() uint
	// Returns the height of this module.
	Height() uint
}

// Column describes an individual column of data within a trace table.
type Column[T any] interface {
	// Holds the name of this column
	Name() string
	// Get the value at a given row in this column.  If the row is
	// out-of-bounds, then the column's padding value is returned instead.
	// Thus, this function always succeeds.
	Get(row int) T
	// Access the underlying data array for this column.  This is useful in
	// situations where we want to clone the entire column, etc.
	Data() array.Array[T]
}

// RawColumn represents a raw column of data which has not (yet) been indexed as
// part of a trace, etc.  Raw columns are typically read difrtly from trace
// files, and subsequently indexed into a trace during the expansion process.
type RawColumn[T any] struct {
	// Name of the enclosing module
	Module string
	// Name of the column
	Name string
	// Data held in the column
	Data array.MutArray[T]
}

// QualifiedName returns the fully qualified name of this column.
func (p *RawColumn[T]) QualifiedName() string {
	return QualifiedColumnName(p.Module, p.Name)
}

// Clone this raw column producing an unaliased copy.
func (p *RawColumn[T]) Clone() RawColumn[T] {
	return RawColumn[T]{
		p.Module,
		p.Name,
		p.Data.Clone(),
	}
}

// ============================================================================
// Trace Wrapper
// ============================================================================

// Wrap provides a wrapper which makes a trace of words look like a trace of
// field elements.
func Wrap[W word.Word[W], F field.Element[F]](trace Trace[W]) Trace[F] {
	return &frTrace[W, F]{trace}
}

type frTrace[W word.Word[W], F field.Element[F]] struct {
	trTrace Trace[W]
}

// Access a given column difrtly via a reference.
func (p *frTrace[W, F]) Column(cref ColumnRef) Column[F] {
	return &frColumn[W, F]{p.trTrace.Column(cref)}
}

// Access a given module in this trace.
func (p *frTrace[W, F]) Module(mid ModuleId) Module[F] {
	return &frModule[W, F]{p.trTrace.Module(mid)}
}

// Determine whether this trace has a module with the given name and, if so,
// what its module index is.
func (p *frTrace[W, F]) HasModule(name string) (uint, bool) {
	return p.trTrace.HasModule(name)
}

// Returns the number of modules in this trace.
func (p *frTrace[W, F]) Width() uint {
	return p.trTrace.Width()
}

// Returns an iterator over the contained modules
func (p *frTrace[W, F]) Modules() iter.Iterator[Module[F]] {
	panic("unreachable")
}

// Provides access to the internal memory pool
func (p *frTrace[W, F]) Pool() word.Pool[uint, F] {
	panic("unreachable")
}

func (p *frTrace[W, F]) String() string {
	return fmt.Sprintf("%v", p.trTrace)
}

// ============================================================================
// Module Wrapper
// ============================================================================

type frModule[W word.Word[W], F field.Element[F]] struct {
	trModule Module[W]
}

// Module implementation for trace.Module interface.
func (p *frModule[W, F]) Name() string {
	return p.trModule.Name()
}

// Column implementation for trace.Module interface.
func (p *frModule[W, F]) Column(index uint) Column[F] {
	return &frColumn[W, F]{p.trModule.Column(index)}
}

// ColumnOf implementation for trace.Module interface.
func (p *frModule[W, F]) ColumnOf(name string) Column[F] {
	return &frColumn[W, F]{p.trModule.ColumnOf(name)}
}

// Width implementation for trace.Module interface.
func (p *frModule[W, F]) Width() uint {
	return p.trModule.Width()
}

// Height implementation for trace.Module interface.
func (p *frModule[W, F]) Height() uint {
	return p.trModule.Height()
}

// ============================================================================
// Column Wrapper
// ============================================================================

// frColumn is a wrapper which enables the array being computed to be accessed
// during its own computation.
type frColumn[W word.Word[W], F field.Element[F]] struct {
	trColumn Column[W]
}

// Holds the name of this column
func (p *frColumn[W, F]) Name() string {
	return p.trColumn.Name()
}

// Get implementation for trace.Column interface.
func (p *frColumn[W, F]) Get(row int) F {
	val := p.trColumn.Get(row)
	// Convert to a field element
	return field.FromBigEndianBytes[F](val.Bytes())
}

// Data implementation for trace.Column interface.
func (p *frColumn[W, F]) Data() array.Array[F] {
	return field.NewWrappedArray[W, F](p.trColumn.Data())
}
