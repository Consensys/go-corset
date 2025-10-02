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
	// Source map describes the overall structure of the schema, including
	// elements related to display.  This additionally encodes a mapping from
	// source-level columns to their limbs.
	srcmap corset.SourceMap
}

// NewBuilder constructs a default builder.
func NewBuilder[F field.Element[F]](srcmap corset.SourceMap) Builder[F] {
	return Builder[F]{util.None[CellRefSet](), 0, srcmap}
}

// Padding sets the amount of additional rows to show either side of the viewing
// window.
func (p Builder[F]) Padding(padding uint) Builder[F] {
	var builder = p
	//
	builder.padding = padding
	//
	return builder
}

// Build the viewing window for this trace.
func (p Builder[F]) Build(trace tr.Trace[F]) TraceView {
	var windows = make([]moduleView[F], trace.Width())
	//
	for i := range trace.Width() {
		srcModule := findSourceModule(trace.Module(i).Name(), p.srcmap.Root)
		// construct initial module view
		windows[i] = moduleView[F]{
			id:      i,
			padding: p.padding,
			trace:   trace.Module(i),
			srcmap:  srcModule,
		}
	}
	//
	return &traceView[F]{windows}
}

func findSourceModule(name string, mod corset.SourceModule) corset.SourceModule {
	if name == "" {
		return mod
	}
	//
	for _, m := range mod.Submodules {
		if m.Name == name {
			return m
		}
	}
	//
	panic(fmt.Sprintf("unknown module \"%s\"", name))
}
