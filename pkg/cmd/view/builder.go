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
	"github.com/consensys/go-corset/pkg/corset"
	"github.com/consensys/go-corset/pkg/schema"
	sc "github.com/consensys/go-corset/pkg/schema"
	tr "github.com/consensys/go-corset/pkg/trace"
	"github.com/consensys/go-corset/pkg/util"
	"github.com/consensys/go-corset/pkg/util/collection/set"
	"github.com/consensys/go-corset/pkg/util/field"
	"github.com/consensys/go-corset/pkg/util/file"
)

// CellRefSet defines a type for sets of cell references.
type CellRefSet = set.AnySortedSet[tr.CellRef]

// Builder is responsible for building the viewing window for a trace.
type Builder[F field.Element[F]] struct {
	// Optional set of cells on which the constructed view should focus.
	// Observe that cells refer to trace cells (i.e. which are in terms of
	// limbs, not source columns).
	cells util.Option[CellRefSet]
	// Limbs indicates whether or not to show the raw limbs, or the combined
	// source-level register.
	limbs bool
	// Default cell width to use
	cellWidth uint
	// Default title width to use
	titleWidth uint
	// Limbs mapping identifies how source-level registers are mapped into
	// limbs.  This is necessary in order to reconstruct source-level column
	// data from the trace.
	mapping schema.LimbsMap
	// Formatting to use
	formatting TraceFormatting
	// Optional source map information.  This is primarily used to determine
	srcmap util.Option[corset.SourceMap]
}

// NewBuilder constructs a default builder.
func NewBuilder[F field.Element[F]](mapping schema.LimbsMap) Builder[F] {
	return Builder[F]{util.None[CellRefSet](), false, 16, 16, mapping,
		DefaultFormatter(), util.None[corset.SourceMap]()}
}

// WithCellWidth sets the maximum width of any cell in the view.
func (p Builder[F]) WithCellWidth(cellWidth uint) Builder[F] {
	var builder = p
	//
	builder.cellWidth = cellWidth
	//
	return builder
}

// WithTitleWidth sets the maximum width of any row title in the view.
func (p Builder[F]) WithTitleWidth(titleWidth uint) Builder[F] {
	var builder = p
	//
	builder.titleWidth = titleWidth
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

// WithFormatting determines what formatting to use when rendering this view.
func (p Builder[F]) WithFormatting(formatting TraceFormatting) Builder[F] {
	var builder = p
	//
	builder.formatting = formatting
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
	var windows []ModuleView
	//
	srcmap, enums := extractSourceMap(p.srcmap)
	//
	for i := range p.mapping.Width() {
		trMod := trace.Module(i)
		scMod := p.mapping.Module(i)
		//
		public, columns := extractSourceMapData(trMod.Name(), srcmap, scMod)
		//
		data := newModuleData(i, scMod, trMod, public, enums, columns)
		// construct initial module view
		windows = append(windows, &moduleView[F]{
			window:     data.Window(),
			limbs:      p.limbs,
			cellWidth:  p.cellWidth,
			titleWidth: p.titleWidth,
			formatting: p.formatting.Module(data),
			data:       data,
		})
	}
	//
	return &traceView{windows}
}

func extractSourceMap(optSrcmap util.Option[corset.SourceMap]) (map[string]corset.SourceModule, []corset.Enumeration) {
	var (
		mapping = make(map[string]corset.SourceModule)
		enums   []corset.Enumeration
	)

	if optSrcmap.HasValue() {
		var srcmap = optSrcmap.Unwrap()
		//
		for _, module := range srcmap.Flattern(concreteModules) {
			mapping[module.Name] = module
		}
		//
		enums = srcmap.Enumerations
	}
	//
	return mapping, enums
}

func extractSourceMapData(name string, srcmap map[string]corset.SourceModule,
	mapping sc.RegisterLimbsMap) (bool, []SourceColumn) {
	// Check whether any
	var (
		public  = true
		columns []SourceColumn
	)
	//
	if m, ok := srcmap[name]; ok {
		public = m.Public
		columns = extractSourceColumns(file.NewAbsolutePath(""), m.Selector, m.Columns, m.Submodules, mapping)
	}
	//
	return public, columns
}

// ExtractSourceColumns extracts source column descriptions for a given module
// based on the corset source mapping.  This is particularly useful when you
// want to show the original name for a column (e.g. when its in a perspective),
// rather than the raw register name.
func extractSourceColumns(path file.Path, selector util.Option[string], columns []corset.SourceColumn,
	submodules []corset.SourceModule, mapping sc.RegisterLimbsMap) []SourceColumn {
	//
	var srcColumns []SourceColumn
	//
	for _, col := range columns {
		name := path.Extend(col.Name).String()[1:]
		//
		srcCol := SourceColumn{
			Name:     name,
			Display:  col.Display,
			Computed: col.Computed,
			Selector: selector,
			Register: col.Register.Register(),
			Limbs:    mapping.LimbIds(col.Register.Register()),
		}
		srcColumns = append(srcColumns, srcCol)
	}
	//
	for _, submod := range submodules {
		// Curiously, it only makes sense to recurse on virtual modules here.
		if submod.Virtual {
			subpath := path.Extend(submod.Name)
			subSrcColumns := extractSourceColumns(*subpath, submod.Selector, submod.Columns, submod.Submodules, mapping)
			srcColumns = append(srcColumns, subSrcColumns...)
		}
	}
	//
	return srcColumns
}

func concreteModules(m *corset.SourceModule) bool {
	return !m.Virtual
}
