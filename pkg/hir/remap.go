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
package hir

import (
	"math"

	tr "github.com/consensys/go-corset/pkg/trace"
)

// Remap a given schema using a given criteria to determine which modules to keep.
func Remap(schema *Schema, criteria func(Module) bool) {
	var remapper Remapper
	// Remap modules
	remapper.remapModules(*schema, criteria)
	// Remap columns
	remapper.remapColumns(*schema)
	// Done
	*schema = remapper.schema
}

// Remapper is a tool for remapping modules and columns in a schema after one or
// more modules have been deleted.
type Remapper struct {
	// Mapping from module identifiers before to module identifiers after.
	modmap []uint
	// Mapping from column identifiers before to column identifiers after.
	colmap []uint
	// New schema being constructed
	schema Schema
}

func (p *Remapper) remapModules(oSchema Schema, criteria func(Module) bool) {
	// Initialise modmap
	p.modmap = make([]uint, len(oSchema.modules))
	//
	for i, m := range oSchema.modules {
		var mid uint
		// Decide whether or not to keep the module.
		if criteria(m) {
			mid = p.schema.AddModule(m.Name, m.Condition)
		} else {
			// Mark module as deleted
			mid = math.MaxUint
		}
		//
		p.modmap[i] = mid
	}
}

func (p *Remapper) remapColumns(oSchema Schema) {
	// Initialise colmap
	p.colmap = make([]uint, len(oSchema.inputs))
	//
	for iter, i := oSchema.InputColumns(), 0; iter.HasNext(); i++ {
		var (
			ith = iter.Next()
			cid uint
		)
		// Add column (if applicable)
		if p.modmap[ith.Context.ModuleId] != math.MaxUint {
			ctx := p.remapContext(ith.Context)
			cid = p.schema.AddDataColumn(ctx, ith.Name, ith.DataType)
		} else {
			// Column is part of a module which has been deleted, hence this
			// column is also deleted.
			cid = math.MaxUint
		}
		//
		p.colmap[i] = cid
	}
}

func (p *Remapper) remapContext(ctx tr.Context) tr.Context {
	mid := p.modmap[ctx.ModuleId]
	// sanity check
	if mid == math.MaxUint {
		panic("remapping deleted context")
	}
	//
	return tr.NewContext(mid, ctx.LengthMultiplier())
}
