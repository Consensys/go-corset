package binfile

import (
	"fmt"

	"github.com/consensys/go-corset/pkg/table"
)

type jsonType struct {
	// Determines the representation of this type.  For example, a
	// 8bit unsigned integer.
	Magma any `json:"m"`
	// Determines the interpretation of this type.  Specifically,
	// for binary types, we can have an interpretation of either
	// bool (where 0 is false and anything else is true) or loob
	// (where 0 is true and anything else is false).
	Conditioning string `json:"c"`
}

// =============================================================================
// Translation
// =============================================================================

func (e *jsonType) toHir() table.Type {
	// Check whether magma is string
	if str, ok := e.Magma.(string); ok {
		switch str {
		case "Native":
			return &table.FieldType{}
		case "Byte":
			return table.NewUintType(8)
		case "Binary":
			return table.NewUintType(1)
		default:
			panic(fmt.Sprintf("Unknown JSON type encountered: %s:%s", e.Magma, e.Conditioning))
		}
	}
	// Try as integer
	if intMap, ok := e.Magma.(map[string]any); ok {
		if val, isInt := intMap["Integer"]; isInt {
			nbits := uint(val.(float64))
			return table.NewUintType(nbits)
		}
	}
	// Fail
	panic(fmt.Sprintf("Unknown JSON type encountered: %s:%s", e.Magma, e.Conditioning))
}
