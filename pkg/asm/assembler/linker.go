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
	"github.com/consensys/go-corset/pkg/asm/io/macro"
	"github.com/consensys/go-corset/pkg/schema"
	"github.com/consensys/go-corset/pkg/util/field"
	"github.com/consensys/go-corset/pkg/util/source"
)

// AssemblyItem represents an intermediate artifact producing during the
// assembly process.  The key is that, as this stage, the bus identifiers are
// not fully known.  Hence, when multiple assembly items come together we must
// "align buses appropriately between them.
type AssemblyItem[F field.Element[F]] struct {
	// Components making up this assembly item.
	Components []MacroFunction[F]
	// Mapping of instructions back to the source file.
	SourceMap source.Map[any]
}

// Link a set of one or more assembly items together to produce a complete
// program, or one or more errors.  Linking is the process of connecting buses
// which are used (e.g. by a call instruction) with their definitions (e.g. a
// function declaration).
func Link[F field.Element[F]](items ...AssemblyItem[F]) (MacroProgram[F], source.Maps[any]) {
	var (
		srcmap     source.Maps[any] = *source.NewSourceMaps[any]()
		busmap     map[string]uint  = make(map[string]uint)
		components []*MacroFunction[F]
	)
	// Constuct bus and source mappings
	for _, item := range items {
		srcmap.Join(&item.SourceMap)
		//
		for _, c := range item.Components {
			if _, ok := busmap[c.Name()]; ok {
				// Indicates component of same name already exists.  It would be
				// good to report a source error here, but the problem is that
				// our source map doesn't contain the right information.
				panic(fmt.Sprintf("duplicate component %s", c.Name()))
			}
			// Allocate bus entry
			busmap[c.Name()] = uint(len(busmap))
			//
			components = append(components, &c)
		}
	}
	// Link all assembly items
	for i := range components {
		linkComponent(uint(i), components, busmap)
	}
	//
	return io.NewProgram(components...), srcmap
}

// Link all buses used within this function to their intended targets.  This
// means, for every bus used locally, settings the global bus identifier and
// also allocated regisers for the address/data lines.
func linkComponent[F field.Element[F]](index uint, components []*MacroFunction[F], busmap map[string]uint) {
	// Mapping of bus names to allocated buses
	var (
		fn         = components[index]
		code       = fn.Code()
		localBuses = make(map[uint]io.Bus, 0)
	)
	//
	for i := range code {
		insn := code[i]
		//
		if bi, ok := insn.(macro.IoInstruction); ok {
			// Determine global bus identifier
			busId := busmap[bi.Bus().Name]
			// allocate & link bus
			bi.Link(allocateBus(busId, localBuses, index, components))
		}
	}
}

// Get the local bus declared for the given function, either by allocating a new
// bus (if was not already allocated) or returning the existing bus (if it was
// previously allocated).  Allocating a new bus requires allocating
// corresponding I/O registers within the given function.
func allocateBus[F field.Element[F]](busId uint, localBuses map[uint]io.Bus, index uint,
	components []*MacroFunction[F]) io.Bus {
	//
	var (
		fn      = components[index]
		busName = components[busId].Name()
		inputs  = components[busId].Inputs()
		outputs = components[busId].Outputs()
	)
	// Check whether previously allocated, or not.
	if bus, ok := localBuses[busId]; ok {
		// Yes, so just return previously created bus.
		return bus
	}
	// No, therefore create new bus.
	addressLines := allocateIoRegisters(busName, inputs, fn)
	dataLines := allocateIoRegisters(busName, outputs, fn)
	bus := io.NewBus(busName, busId, addressLines, dataLines)
	// Update local bus map
	localBuses[busId] = bus
	// Done
	return bus
}

func allocateIoRegisters[F field.Element[F]](busName string, registers []io.Register, fn *MacroFunction[F],
) []io.RegisterId {
	//
	var lines []io.RegisterId
	//
	for _, reg := range registers {
		var regName string
		// Determine suitable register name
		if reg.IsInput() {
			regName = fmt.Sprintf("%s>%s[%s]", fn.Name(), busName, reg.Name)
		} else if reg.IsOutput() {
			regName = fmt.Sprintf("%s<%s[%s]", fn.Name(), busName, reg.Name)
		} else {
			panic("unreachable")
		}
		// Allocate register
		lines = append(lines, fn.AllocateRegister(schema.COMPUTED_REGISTER, regName, reg.Width))
	}
	//
	return lines
}
