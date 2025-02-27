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
package corset

import (
	"encoding/gob"
	"math/big"

	"github.com/consensys/gnark-crypto/ecc/bls12-377/fr"
	"github.com/consensys/go-corset/pkg/binfile"
	"github.com/consensys/go-corset/pkg/hir"
	"github.com/consensys/go-corset/pkg/schema"
	sc "github.com/consensys/go-corset/pkg/schema"
)

// SourceMap is a binary file attribute which provides debugging
// information about the relationship between registers and source-level
// columns.  This is used, for example, within the inspector.
type SourceMap struct {
	// Root module correspond to the top-level HIR modules.  Thus, indicates into
	// this table correspond to HIR module indices, etc.
	Root SourceModule
	// Enumerations are custom types for display.  For example, we might want to
	// display opcodes as ADD, MUl, SUB, etc.
	Enumerations []Enumeration
}

// AttributeName returns the name of the binary file attribute that this will
// generate.  This is used, for example, when listing attributes contained
// within a binary file.
func (p *SourceMap) AttributeName() string {
	return "CorsetSourceMap"
}

// Flattern modules in this tree matching a given criteria
func (p *SourceMap) Flattern(predicate func(*SourceModule) bool) []SourceModule {
	return p.Root.Flattern(predicate)
}

// Enumeration is a mapping from field elements to explicitly given names.  For
// example, mapping opcode bytes to their names.
type Enumeration map[fr.Element]string

// SourceModule represents an entity at the source-level which groups together
// related columns.  Modules can be either concrete (in which case they
// correspond with HIR modules) or virtual (in which case they are encoded
// within an HIR module).
type SourceModule struct {
	// Name of this submodule.
	Name string
	// Synthetic indicates whether or not this module was automatically
	// generated or not.
	Synthetic bool
	// Virtual indicates whether or not this is a "virtual" module.  That is, a
	// module which is artificially embedded in some outer (concrete) module.
	Virtual bool
	// Selector determines when this (sub)module is active.  Specifically, when
	// it evaluates to a non-zero value the module is active.
	Selector *hir.UnitExpr
	// Submodules identifies any (virtual) submodules contained within this.
	// Currently, perspectives are the only form of submodule currently
	// supported.
	Submodules []SourceModule
	// Columns identifies any columns defined in this module.  Observe that
	// columns across modules are mapped to registers in a many-to-one fashion.
	Columns []SourceColumn
	// Constants identifiers any constants defined in this module.
	Constants []SourceConstant
}

// Flattern modules in this tree either including (or excluding) virtual
// modules.
func (p *SourceModule) Flattern(predicate func(*SourceModule) bool) []SourceModule {
	var modules []SourceModule

	if predicate(p) {
		modules = append(modules, *p)
		for _, child := range p.Submodules {
			modules = append(modules, child.Flattern(predicate)...)
		}
	}

	return modules
}

// SourceColumn represents a source-level column which is mapped to a given HIR
// register.  Observe that multiplie source-level columns can be mapped to the
// same register.
type SourceColumn struct {
	Name string
	// Length Multiplier of source-level column.
	Multiplier uint
	// Underlying DataType of the source-level column.
	DataType sc.Type
	// Provability requirement for source-level column.
	MustProve bool
	// Determines whether this is a Computed column.
	Computed bool
	// Display modifier for column. Here 0-256 are reserved, and values >256 are
	// entries in Enumerations map.  More specifically, 0=hex, 1=dec, 2=bytes.
	Display uint
	// Register at HIR level to which this column is mapped.
	Register uint
}

// DISPLAY_HEX shows values in hex
const DISPLAY_HEX = uint(0)

// DISPLAY_DEC shows values in dec
const DISPLAY_DEC = uint(1)

// DISPLAY_BYTES shows values as bytes.
const DISPLAY_BYTES = uint(2)

// DISPLAY_CUSTOM selects a custom layout
const DISPLAY_CUSTOM = uint(256)

// SourceConstant provides information about constant values which are exposed
// to the trace generator.  Such constants can, in some cases, be modified to
// reflect different environments (e.g. different chains, gas limits, etc).
type SourceConstant struct {
	Name string
	// value of the constant
	Value big.Int
	// Explicit type for this constant.  This maybe nil if no type was given
	// and, instead, the type should be inferred from context.
	DataType schema.Type
	// Indicates whether this is an "externally visible" constant.  That is, one
	// whose value can be changed after the fact.
	Extern bool
}

func init() {
	gob.Register(binfile.Attribute(&SourceMap{}))
}
