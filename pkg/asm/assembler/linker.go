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
package assembler

import (
	"fmt"

	"github.com/consensys/go-corset/pkg/asm/io"
	"github.com/consensys/go-corset/pkg/util/source"
)

// AssemblyItem represents an intermediate artifact producing during the
// assembly process.  The key is that, as this stage, the bus identifiers are
// not fully known.  Hence, when multiple assembly items come together we must
// "align buses appropriately between them.
type AssemblyItem struct {
	// Buses maps bus names to the concrete identifiers used within this item.
	Buses []string
	// Components making up this assembly item.
	Components []MacroFunction
	// Mapping of instructions back to the source file.
	SourceMap source.Map[any]
}

// Link a set of one or more assembly items together to produce a complete
// program, or one or more errors.  Linking is the process of connecting buses
// which are used (e.g. by a call instruction) with their definitions (e.g. a
// function declaration).
func Link(items ...AssemblyItem) (MacroProgram, source.Maps[any]) {
	var (
		srcmap     source.Maps[any] = *source.NewSourceMaps[any]()
		busmap     map[string]uint  = make(map[string]uint)
		components []MacroFunction
	)
	// Constuct bus and source mappings
	for _, item := range items {
		srcmap.Join(&item.SourceMap)
		//
		for _, c := range item.Components {
			if _, ok := busmap[c.Name]; ok {
				// Indicates component of same name already exists.  It would be
				// good to report a source error here, but the problem is that
				// our source map doesn't contain the right information.
				panic(fmt.Sprintf("duplicate component %s", c.Name))
			}
			// Allocate bus entry
			busmap[c.Name] = uint(len(busmap))
			//
			components = append(components, c)
		}
	}
	// Link all assembly items
	for _, item := range items {
		linkAssemblyItem(item, busmap)
	}
	//
	return io.NewProgram(components...), srcmap
}

func linkAssemblyItem(item AssemblyItem, busmap map[string]uint) {
	//
	var buses = make([]uint, len(item.Buses))
	//
	for i, name := range item.Buses {
		bid, ok := busmap[name]
		//
		if !ok {
			bid = io.UNKNOWN_BUS
		}
		//
		buses[i] = bid
	}
	// Link each component
	for _, fn := range item.Components {
		for i := range fn.Code {
			fn.Code[i].Link(buses)
		}
	}
}
