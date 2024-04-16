package binfile

import (
	"encoding/json"
	// "fmt"
	// "github.com/Consensys/go-corset/pkg/ast"
)

// This is very much a Work-In-Progress :)

// =============================================================================
// ColumnSet
// =============================================================================

type RegisterID = any
type Value struct{}
type Magma = any
type Kind struct {
	m string
	c string
}
type Base = string
type Handle = string
type Register = any
type FieldRegister = any

// Column .
type Column struct {
	Register            int    `json:"register"`
	Shift               int    `json:"shift"`
	PaddingValue        Value  `json:"padding_value"`
	Used                bool   `json:"used"`
	MustProve           bool   `json:"must_prove"`
	Kind                string `json:"kind"`
	T                   Magma  `json:"t"`
	IntrinsicSizeFactor string `json:"intrinsic_size_factor"`
	Base                Base   `json:"base"`
	Gandle              Handle `json:"gandle"`
	Computed            bool   `json:"computed"`
}

// ColumnSet .
type ColumnSet struct {
	// Raw array of column data, including virtial those which are
	// virtual and/or overlapping with others.
	Cols []Column `json:"_cols"`
	// Maps column handles to their index in the Cols array.
	ColsMap map[Handle]uint `json:"cols"`
	// Maps column handles? to their length
	EffectiveLen map[string]int
	// ?
	MinLen map[string]uint
	// ?
	FieldRegisters []FieldRegister
	// ?
	Registers []Register
	// ?
	Spilling map[string]int
}

// JsonDomain domain type.
type JsonDomain = string

// ConstraintSet .
type ConstraintSet struct {
	//Columns ColumnSet `json:"columns"`
	Constraints []JsonConstraint `json:"constraints"`
	//
	// constants any
	// computations any
	// perspectives []string
	// transformations uint64
	// auto_constraints uint64
}

// ConstraintSetFromJson reads a constraint set from a set of bytes representing its JSON
// encoding.  The format for this was (originally) determined by the
// Rust corset tool.
func ConstraintSetFromJson(bytes []byte) (cs ConstraintSet, err error) {
	var res ConstraintSet
	//var data map[string]any
	// Unmarshall
	jsonErr := json.Unmarshal(bytes, &res)
	//
	//res.Constraints = ConstraintArrayFromJson(data["constraints"])
	// For now return directly.
	return res, jsonErr
}
