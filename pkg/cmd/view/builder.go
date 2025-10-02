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
	// Limbs mapping identifies how source-level registers are mapped into
	// limbs.  This is necessary in order to reconstruct source-level column
	// data from the trace.
	mapping schema.LimbsMap
}

// NewBuilder constructs a default builder.
func NewBuilder[F field.Element[F]](mapping schema.LimbsMap) Builder[F] {
	return Builder[F]{util.None[CellRefSet](), 0, false, mapping}
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

// Build the viewing window for this trace.
func (p Builder[F]) Build(trace tr.Trace[F]) TraceView {
	var windows = make([]moduleView[F], trace.Width())
	//
	for i := range trace.Width() {
		// construct initial module view
		windows[i] = moduleView[F]{
			id:      i,
			padding: p.padding,
			limbs:   p.limbs,
			trace:   trace.Module(i),
			mapping: p.mapping.Module(i),
		}
	}
	//
	return &traceView[F]{windows}
}
