package json

import (
	"strings"

	"github.com/consensys/go-corset/pkg/trace"
)

// ToJsonString converts a trace into a JSON string.
func ToJsonString(columns []trace.RawColumn) string {
	var builder strings.Builder
	//
	builder.WriteString("{")
	//
	for i := 0; i < len(columns); i++ {
		ith := columns[i]
		//
		if i != 0 {
			builder.WriteString(", ")
		}
		//
		builder.WriteString("\"")
		// Construct qualified column name
		name := trace.QualifiedColumnName(ith.Module, ith.Name)
		// Write out column name
		builder.WriteString(name)
		//
		builder.WriteString("\": [")

		data := ith.Data

		for j := uint(0); j < data.Len(); j++ {
			if j != 0 {
				builder.WriteString(", ")
			}

			jth := data.Get(j)
			builder.WriteString(jth.String())
		}

		builder.WriteString("]")
	}
	//
	builder.WriteString("}")
	// Done
	return builder.String()
}
