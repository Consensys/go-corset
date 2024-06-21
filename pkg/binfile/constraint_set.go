package binfile

import (
	"encoding/json"
	"fmt"

	"github.com/consensys/go-corset/pkg/hir"
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
	Register int
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
	IntrinsicSizeFactor string `json:"intrinsic_size_factor"`
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
	Registers []any `json:"registers"`
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
	// Add Columns
	for _, c := range res.Columns.Cols {
		// Sanity checks
		if c.Kind == "Computed" {
			// Ignore.
		} else if c.Computed {
			fmt.Printf("COLUMN: %s\n", c.Handle)
			panic("invalid JSON column configuration")
		} else {
			t := c.Type.toHir()
			schema.AddDataColumn(c.Handle, t)
			// Check whether a type constraint required or not.
			if c.MustProve {
				schema.AddTypeConstraint(c.Handle, t)
			}
		}
	}
	// Add constraints
	for _, c := range res.Constraints {
		c.addToSchema(schema)
	}
	// Add computations
	res.Computations.addToSchema(schema)
	// For now return directly.
	return schema, jsonErr
}
