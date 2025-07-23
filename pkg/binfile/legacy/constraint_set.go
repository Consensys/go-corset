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
package binfile

import (
	"encoding/json"
	"fmt"

	"github.com/consensys/go-corset/pkg/hir"
	sc "github.com/consensys/go-corset/pkg/schema"
	"github.com/consensys/go-corset/pkg/trace"
)

// This is very much a Work-In-Progress :)

// =============================================================================
// ColumnSet
// =============================================================================

type column struct {
	// The name of this column in the format "module:name".
	Handle string
	// The numerical column to which this column is assigned.
	// Specifically, as a result of perspectives, multiple columns
	// can be assigned to the same "register".
	Register uint
	// Indicates the padding value (if given) to use when padding
	// out a trace for this column.
	PaddingValue any `json:"padding_value"`
	// Determines whether the type was marked with "@prove" or
	// not.  Such types must be established by corset in some way
	// (e.g. by adding an appropriate constraint).
	MustProve bool `json:"must_prove"`
	// Specifies the type that all values of this column are
	// intended to adhere to.  Observe, however, this is only
	// guaranteed when MustProve holds.  Otherwise, they are
	// really just a suggestion for debugging purposes.
	Type *jsonType `json:"t"`
	// Indicates some kind of "length multiplier" which is gives
	// information about the length of this column (e.g. its a
	// multiple of two).  This seems only relevant for computed
	// columns.
	IntrinsicSizeFactor uint `json:"intrinsic_size_factor"`
	// Indicates this is a computed column.  For binfiles being
	// compiled without expansion, this should always be false.
	Computed bool
	// Provides additional information about whether this column
	// is computed or not.  A "Commitment" kind indicates a
	// user-defined columns (i.e is directly filled from trace
	// files); a "Computed" column is filled by a given function;
	// an "Expression" kind indicates a column whose values are
	// computed from an expresion known at compile time.  As for
	// the Computed field, for binfiles compiled without expansion
	// the only value should be "Commitment".
	Kind string
	// Determines how values of this column should be displayed
	// (e.g. using hexadecimal notation, etc).  This only for
	// debugging purposes.
	Base string
	// Indicates whether or not this column is used by any
	// constraints.  Presumably, this is intended to enable the
	// corset tool to report a warning.
	Used bool
}

type register struct {
	// The name of this register in the format "module:name".
	Handle string `json:"handle"`
	// Indicates this is a computed column.  For binfiles being
	// compiled without expansion, this identifies columns defined by sorted
	// permutations.
	Computed bool
	// Specifies the type that all values of this column are
	// intended to adhere to.  Observe, however, this is only
	// guaranteed when MustProve holds.  Otherwise, they are
	// really just a suggestion for debugging purposes.
	Type *jsonType `json:"magma"`
	// Width specifies (I believe) the number of field elements (exo-columns)
	// required for this register.  For our purposes here, this should always be
	// 1.
	Width uint `json:"width"`
	// MustProve indicates whether or not this register should have its type
	// enforced using a range constraint.  Observe this field is not present in
	// the original binfile format.  Instead, this field is determined from
	// parsing the binfile format.
	MustProve bool
	// LengthMultiplier indicates the length multiplier for this column.  This
	// must be a factor of the number of rows in the column.  For example, a
	// column with length multiplier of 2 must have an even number of rows, etc.
	LengthMultiplier uint `json:"length_multiplier"`
}

type columnSet struct {
	// Raw array of column data, including virtual those which are
	// virtual and/or overlapping with others.
	Cols []column `json:"_cols"`
	// Maps column handles to their index in the Cols array.
	ColsMap map[string]uint `json:"cols"`
	// Maps column handles? to their length
	EffectiveLen map[string]int `json:"effective_len"`
	// ?
	MinLen map[string]uint `json:"min_len"`
	// ?
	FieldRegisters []any `json:"field_registers"`
	// ?
	Registers []register `json:"registers"`
	// ?
	Spilling map[string]int `json:"spilling"`
}

// ConstraintSet .
type constraintSet struct {
	Columns     columnSet        `json:"columns"`
	Constraints []jsonConstraint `json:"constraints"`
	//
	// constants any
	Computations jsonComputationSet `json:"computations"`
	// perspectives []string
	// transformations uint64
	// auto_constraints uint64
}

// HirSchemaFromJson constructs an HIR schema from a set of bytes representing
// the JSON encoding for a set of constraints / columns.
func HirSchemaFromJson(bytes []byte) (schema *hir.Schema, err error) {
	var res constraintSet
	// Unmarshall
	jsonErr := json.Unmarshal(bytes, &res)
	// Construct schema
	schema = hir.EmptySchema()
	// Transfer column info
	transferColumnInfo(&res.Columns)
	// Allocate registers
	colmap := allocateRegisters(&res, schema)
	// Double check allocation is correct
	checkAllocation(&res.Columns, colmap, schema)
	// Finally, add constraints
	for _, c := range res.Constraints {
		c.addToSchema(colmap, schema)
	}

	// For now return directly.
	return schema, jsonErr
}

// This transfers over some information from columns to registers.  It may seem
// a slightly odd thing to do, but it simply allows us to separate processing of
// columns from processing of registers.
func transferColumnInfo(cs *columnSet) {
	// Move key data from columns to registers
	for _, c := range cs.Cols {
		// Sanity checks
		if c.Kind == "Computed" {
			cs.Registers[c.Register].Computed = true
		} else if c.Computed {
			fmt.Printf("COLUMN: %s\n", c.Handle)
			panic("invalid JSON column configuration")
		} else if c.MustProve {
			// Copy over must-prove info.
			cs.Registers[c.Register].MustProve = true
		}
	}
}

// Allocate all registers as columns in the given schema, whilst producing a
// "column mapping".  The mapping goes from binfile column indices to schema
// column indices.
func allocateRegisters(cs *constraintSet, schema *hir.Schema) map[uint]uint {
	colmap := make(map[uint]uint)
	//
	for _, c := range cs.Columns.Registers {
		// Computed columns are ignored because they are added separately from
		// computations (see below).
		if !c.Computed {
			handle := asHandle(c.Handle)
			mid := registerModule(schema, handle.module)
			ctx := trace.NewContext(mid, c.LengthMultiplier)
			col_type := c.Type.toHir()
			// Add column for this
			cid := schema.AddDataColumn(ctx, handle.column, col_type)
			// Check whether a type constraint required or not.
			if c.MustProve && col_type.AsUint() != nil {
				bound := col_type.AsUint().BitWidth()
				schema.AddRangeConstraint(c.Handle, ctx, hir.NewColumnAccess(cid, 0), bound)
			}
		}
	}
	// Build preliminary column map
	for i, col := range cs.Columns.Cols {
		// Determine register ID
		reg := cs.Columns.Registers[col.Register]
		//
		if !reg.Computed {
			// Extract register handle
			handle := asHandle(reg.Handle)
			// Determine enclosing module
			mid := registerModule(schema, handle.module)
			// Lookup register in schema
			cid, ok := sc.ColumnIndexOf(schema, mid, handle.column)
			// Handle error case
			if !ok {
				panic(fmt.Sprintf("unknown column %s.%s", handle.module, handle.column))
			}

			colmap[uint(i)] = cid
		}
	}
	// Add computations (and finalise column map)
	cs.Computations.addToSchema(cs.Columns.Cols, colmap, schema)
	//
	return colmap
}

// Double check the allocation was made correctly.  This step is strictly
// unnecessary, but provides a useful safety net given the complexity and
// significance of getting the allocation right.
func checkAllocation(cs *columnSet, colmap map[uint]uint, schema *hir.Schema) {
	for i, col := range cs.Cols {
		// Determine register ID
		reg := cs.Registers[col.Register]
		// Extract register handle
		handle := asHandle(reg.Handle)
		// Check it all lines up.
		cid, ok := colmap[uint(i)]
		// Sanity check
		if !ok {
			panic(fmt.Sprintf("unallocated column %s.%s", handle.module, handle.column))
		}

		sc_col := schema.Columns().Nth(cid)
		sc_mod := schema.Modules().Nth(sc_col.Context.Module())
		// Perform the check
		if sc_mod.Name != handle.module || sc_col.Name != handle.column {
			panic(fmt.Sprintf("invalid allocation %s.%s != %s.%s", handle.module, handle.column, sc_mod.Name, sc_col.Name))
		}
	}
}

// Register a module within the schema.  If the module already exists, then
// simply return its existing index.
func registerModule(schema *hir.Schema, module string) uint {
	// Attempt to find existing module with same name
	mid, ok := schema.Modules().Find(func(m sc.Module) bool {
		return m.Name == module
	})
	// Check whether search successful, or not.
	if ok {
		return mid
	}
	// Not successful, so create new one.
	return schema.AddModule(module)
}
