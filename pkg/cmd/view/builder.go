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
package view

import (
	"fmt"

	"github.com/consensys/go-corset/pkg/corset"
	"github.com/consensys/go-corset/pkg/schema"
	tr "github.com/consensys/go-corset/pkg/trace"
	"github.com/consensys/go-corset/pkg/util"
	"github.com/consensys/go-corset/pkg/util/collection/set"
	"github.com/consensys/go-corset/pkg/util/field"
)

// CellRefSet defines a type for sets of cell references.
type CellRefSet = set.AnySortedSet[tr.CellRef]

// Builder is responsible for building the viewing window for a trace.
type Builder[F field.Element[F]] struct {
	// Optional set of cells on which the constructed view should focus.
	// Observe that cells refer to trace cells (i.e. which are in terms of
	// limbs, not source columns).
	cells util.Option[CellRefSet]
	// Amount of additional rows to show either side of the focus.
	padding uint
	// Limbs indicates whether or not to show the raw limbs, or the combined
	// source-level register.
	limbs bool
	// Default cell width to use
	cellWidth uint
	// Limbs mapping identifies how source-level registers are mapped into
	// limbs.  This is necessary in order to reconstruct source-level column
	// data from the trace.
	mapping schema.LimbsMap
	// Optional source map information.  This is primarily used to determine
	srcmap util.Option[corset.SourceMap]
}

// NewBuilder constructs a default builder.
func NewBuilder[F field.Element[F]](mapping schema.LimbsMap) Builder[F] {
	return Builder[F]{util.None[CellRefSet](), 0, false, 16, mapping, util.None[corset.SourceMap]()}
}

// WithCellWidth sets the maximum width of any cell in the view.
func (p Builder[F]) WithCellWidth(cellWidth uint) Builder[F] {
	var builder = p
	//
	builder.cellWidth = cellWidth
	//
	return builder
}

// WithPadding sets the amount of additional rows to show either side of the viewing
// window.
func (p Builder[F]) WithPadding(padding uint) Builder[F] {
	var builder = p
	//
	builder.padding = padding
	//
	return builder
}

// WithLimbs determines whether to show columns as raw limbs, or as combined
// source-level registers.
func (p Builder[F]) WithLimbs(limbs bool) Builder[F] {
	var builder = p
	//
	builder.limbs = limbs
	//
	return builder
}

// WithSourceMap applies source-mapping information to the view.  The main
// benefit of this is that it includes display modifiers.
func (p Builder[F]) WithSourceMap(srcmap corset.SourceMap) Builder[F] {
	var builder = p
	//
	builder.srcmap = util.Some(srcmap)
	//
	return builder
}

// Build the viewing window for this trace.
func (p Builder[F]) Build(trace tr.Trace[F]) TraceView {
	var windows = make([]moduleView[F], trace.Width())
	//
	for i := range trace.Width() {
		trMod := trace.Module(i)
		scMod := p.mapping.Module(i)
		//
		display := buildDisplayModifiers(p.srcmap, trMod.Name(), len(scMod.Registers()))
		// construct initial module view
		windows[i] = moduleView[F]{
			id:        i,
			padding:   p.padding,
			limbs:     p.limbs,
			cellWidth: p.cellWidth,
			display:   display,
			trace:     trMod,
			mapping:   scMod,
			filter:    DefaultFilter().Module(i),
		}
	}
	//
	return &traceView[F]{windows}
}

func buildDisplayModifiers(srcmap util.Option[corset.SourceMap], name string, width int) []uint {
	// Check whether any
	display := make([]uint, width)
	//
	fmt.Printf("CREATING WIDTH=%d\n", width)
	//
	return display
}
