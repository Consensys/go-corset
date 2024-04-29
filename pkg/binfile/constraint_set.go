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

type RegisterID = interface{}
type Value struct {}
type Magma = interface{}
type Kind struct {
	m string
	c string
}
type Base = string
type Handle struct {
	H string  `json:"h"`
	ID int  `json:"id"`
}
type Register = interface{}
type FieldRegister = interface{}

type Column struct {
    Register int
    Shift int
    Padding_value Value
    Used bool
    Must_prove bool
    Kind string
    T Magma
    Intrinsic_size_factor string
    Base Base
    Gandle Handle
    Computed bool
}

type ColumnSet struct {
	// Raw array of column data, including virtial those which are
	// virtual and/or overlapping with others.
	Cols []Column `json:"_cols"`
	// Maps column handles to their index in the Cols array.
	ColsMap map[Handle]uint  `json:"cols"`
	// Maps column handles? to their length
	Effective_len map[string]int
	// ?
	min_len map[string]uint
	// ?
	field_registers []FieldRegister
	// ?
	registers []Register
	// ?
	spilling map[string]int
}

// =============================================================================
// Domain
// =============================================================================
type JsonDomain = string

// =============================================================================
// ConstraintSet
// =============================================================================

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

// Read a constraint set from a set of bytes representing its JSON
// encoding.  The format for this was (originally) determined by the
// Rust corset tool.
func ConstraintSetFromJson(bytes []byte) (cs ConstraintSet, err error) {
	var res ConstraintSet
	//var data map[string]interface{}
	// Unmarshall
	json_err := json.Unmarshal(bytes, &res)
	//
	//res.Constraints = ConstraintArrayFromJson(data["constraints"])
	// For now return directly.
	return res,json_err
}
