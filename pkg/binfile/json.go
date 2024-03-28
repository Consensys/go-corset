package binfile

import (
	"encoding/json"
	//	"fmt"
	"github.com/Consensys/go-corset/pkg/constraint"
	"github.com/Consensys/go-corset/pkg/table"
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
type Handle = string
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
// ConstraintSet
// =============================================================================

type ConstraintSet struct {
	Columns ColumnSet
	Constraints []table.Constraint
	// constants interface{}
	// computations interface{}
	// perspectives []string
	// transformations uint64
	// auto_constraints uint64
}

// Read a constraint set from a set of bytes representing its JSON
// encoding.  The format for this was (originally) determined by the
// Rust corset tool.
func ConstraintSetFromJson(bytes []byte) (cs ConstraintSet, err error) {
	var res ConstraintSet
	var data map[string]interface{}
	// Unmarshall
	json_err := json.Unmarshal(bytes, &data)
	//
	res.Constraints = ConstraintArrayFromJson(data["constraints"])
	// For now return directly.
	return res,json_err
}

/// Decode an array of constraints from raw JSON.
func ConstraintArrayFromJson(raw interface{}) []table.Constraint {
	arr := raw.([]interface{})
	// Construct output array
	var res = make([]table.Constraint,len(arr))
	// Parse each constraint
	for i := 0; i < len(arr); i++ {
		res[i] = ConstraintFromJson(arr[i])
	}
	return res
}

// Decode an individual constraint from raw JSON.
func ConstraintFromJson(raw interface{}) table.Constraint {
	// Expose enumeration
	enum := raw.(map[string]interface{})
	// Match on constraint kind
	if val, ok := enum["Vanishes"]; ok {
		return VanishingConstraintFromJson(val)
	} else {
		panic("unknown constraint encounted")
	}
}

// Decode an vanishing constraint from raw JSON.
func VanishingConstraintFromJson(raw interface{}) constraint.Vanishing {
	var c constraint.Vanishing
	// Expose enumeration
	data := raw.(map[string]interface{})
	//
	c.Handle = data["handle"].(string)
	c.Domain = DomainFromJson(data["domain"])
	//
	return c
}

func DomainFromJson(raw interface{}) constraint.Domain {
	return nil
}
